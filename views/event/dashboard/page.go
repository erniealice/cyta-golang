// Package dashboard implements the read-only Schedule live dashboard view
// (Phase 6 — Pyeza dashboard block + per-app live dashboards plan).
package dashboard

import (
	"context"
	"fmt"

	cyta "github.com/erniealice/cyta-golang"

	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"

	eventpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/event/event"
)

// Request mirrors espyna's request.
type Request struct {
	WorkspaceID string
}

// Response is the view-local projection. The view layer never imports the
// espyna use case package; the orchestrator adapts use case responses into
// this shape via the GetDashboardData callback.
type Response struct {
	Today          int64
	ThisWeek       int64
	ByTagCount     int64
	UtilizationPct int64
	ByDayLabels    []string
	ByDayValues    []float64
	ByTag          map[string]int64
	Upcoming       []*eventpb.Event
}

// Deps holds view dependencies.
type Deps struct {
	Routes           cyta.EventRoutes
	EventTagRoutes   cyta.EventTagRoutes
	RecurrenceRoutes cyta.RecurrenceRoutes
	Labels           cyta.EventLabels
	CommonLabels     pyeza.CommonLabels
	GetDashboardData func(ctx context.Context, req *Request) (*Response, error)
}

// PageData is the dashboard template payload.
type PageData struct {
	types.PageData
	ContentTemplate string
	Dashboard       types.DashboardData
}

// NewView creates the schedule dashboard view.
func NewView(deps *Deps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		l := deps.Labels.Dashboard

		var resp *Response
		if deps.GetDashboardData != nil {
			r, err := deps.GetDashboardData(ctx, &Request{WorkspaceID: ""})
			if err == nil && r != nil {
				resp = r
			}
		}
		if resp == nil {
			resp = &Response{ByTag: map[string]int64{}}
		}

		// 14-day events-per-day chart.
		labels := resp.ByDayLabels
		values := resp.ByDayValues
		if len(labels) == 0 {
			labels = make([]string, 14)
			values = make([]float64, 14)
		}
		byDay := &types.ChartData{
			Labels: labels,
			Series: []types.ChartSeries{{
				Name:   l.WidgetByDay,
				Values: values,
				Color:  "terracotta",
			}},
			YAxis: l.WidgetByDay,
		}
		byDay.AutoScale()

		// By-tag pie.
		tagLabels := make([]string, 0, len(resp.ByTag))
		tagValues := make([]float64, 0, len(resp.ByTag))
		for k, v := range resp.ByTag {
			tagLabels = append(tagLabels, k)
			tagValues = append(tagValues, float64(v))
		}
		byTag := &types.ChartData{
			Labels: tagLabels,
			Series: []types.ChartSeries{{
				Name:   l.WidgetByTag,
				Values: tagValues,
			}},
		}
		byTag.AutoScale()

		// Upcoming events list.
		upcoming := buildUpcomingList(resp.Upcoming, l)

		// Manage-tags route falls back to "#" if not wired.
		tagsHref := deps.EventTagRoutes.ListURL
		if tagsHref == "" {
			tagsHref = "#"
		}
		recurrenceHref := deps.RecurrenceRoutes.ListURL
		if recurrenceHref == "" {
			recurrenceHref = "#"
		}

		dash := types.DashboardData{
			Title:    l.Title,
			Icon:     "icon-calendar",
			Subtitle: l.Subtitle,
			QuickActions: []types.QuickAction{
				{Icon: "icon-plus", Label: l.QuickNew, Href: deps.Routes.AddURL, Variant: "primary", TestID: "schedule-action-new"},
				{Icon: "icon-calendar", Label: l.QuickCalendar, Href: deps.Routes.CalendarURL, TestID: "schedule-action-calendar"},
				{Icon: "icon-tag", Label: l.QuickTags, Href: tagsHref, TestID: "schedule-action-tags"},
				{Icon: "icon-repeat", Label: l.QuickRecurrence, Href: recurrenceHref, TestID: "schedule-action-recurrence"},
			},
			Stats: []types.StatCardData{
				{Icon: "icon-calendar", Value: fmt.Sprintf("%d", resp.Today), Label: l.StatToday, Color: "terracotta", TestID: "schedule-stat-today"},
				{Icon: "icon-calendar", Value: fmt.Sprintf("%d", resp.ThisWeek), Label: l.StatThisWeek, Color: "sage", TestID: "schedule-stat-this-week"},
				{Icon: "icon-tag", Value: fmt.Sprintf("%d", resp.ByTagCount), Label: l.StatByTag, Color: "amber", TestID: "schedule-stat-by-tag"},
				{Icon: "icon-activity", Value: fmt.Sprintf("%d", resp.UtilizationPct), Label: l.StatUtilization, Color: "navy", TestID: "schedule-stat-utilization"},
			},
			Widgets: []types.DashboardWidget{
				{
					ID: "events-by-day", Title: l.WidgetByDay,
					Type: "chart", ChartKind: "bar",
					ChartData: byDay, Span: 2,
				},
				{
					ID: "events-by-tag", Title: l.WidgetByTag,
					Type: "chart", ChartKind: "pie",
					ChartData: byTag, Span: 2,
					EmptyState: &types.EmptyStateData{
						Icon: "icon-tag", Title: l.WidgetByTag, Desc: l.EmptyByTag,
					},
				},
				{
					ID: "upcoming-events", Title: l.WidgetUpcoming, Type: "list", Span: 1,
					HeaderActions: []types.QuickAction{
						{Label: l.ViewAll, Href: deps.Routes.ListURL},
					},
					ListItems: upcoming,
					EmptyState: &types.EmptyStateData{
						Icon: "icon-calendar", Title: l.WidgetUpcoming, Desc: l.EmptyUpcoming,
					},
				},
			},
		}

		pageData := &PageData{
			PageData: types.PageData{
				CacheVersion:   viewCtx.CacheVersion,
				Title:          l.Title,
				CurrentPath:    viewCtx.CurrentPath,
				ActiveNav:      "schedule",
				ActiveSubNav:   "dashboard",
				HeaderTitle:    l.Title,
				HeaderSubtitle: l.Subtitle,
				HeaderIcon:     "icon-calendar",
				CommonLabels:   deps.CommonLabels,
			},
			ContentTemplate: "schedule-dashboard-content",
			Dashboard:       dash,
		}
		return view.OK("schedule-dashboard", pageData)
	})
}

func buildUpcomingList(events []*eventpb.Event, l cyta.ScheduleDashboardLabels) []types.ActivityItem {
	if len(events) == 0 {
		return nil
	}
	items := make([]types.ActivityItem, 0, len(events))
	for i, e := range events {
		desc := ""
		if d := e.GetDescription(); d != "" {
			desc = d
		}
		items = append(items, types.ActivityItem{
			IconName:    "icon-calendar",
			IconVariant: "client",
			Title:       e.GetName(),
			Description: desc,
			Time:        e.GetStartDateTimeUtcString(),
			TestID:      fmt.Sprintf("schedule-list-item-%d", i),
		})
	}
	return items
}
