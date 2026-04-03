package list

import (
	"context"
	"fmt"
	"log"

	cyta "github.com/erniealice/cyta-golang"
	lynguaV1 "github.com/erniealice/lyngua/golang/v1"

	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/pyeza-golang/route"
	"github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"

	eventpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/event/event"
)

// ListViewDeps holds view dependencies.
type ListViewDeps struct {
	Routes       cyta.EventRoutes
	ListEvents   func(ctx context.Context, req *eventpb.ListEventsRequest) (*eventpb.ListEventsResponse, error)
	Labels       cyta.EventLabels
	CommonLabels pyeza.CommonLabels
	TableLabels  types.TableLabels
}

// PageData holds the data for the event list page.
type PageData struct {
	types.PageData
	ContentTemplate string
	Table           *types.TableConfig
}

// NewView creates the event list view.
func NewView(deps *ListViewDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)

		status := viewCtx.Request.PathValue("status")
		if status == "" {
			status = "upcoming"
		}

		resp, err := deps.ListEvents(ctx, &eventpb.ListEventsRequest{})
		if err != nil {
			log.Printf("Failed to list events: %v", err)
			return view.Error(fmt.Errorf("failed to load events: %w", err))
		}

		l := deps.Labels
		columns := eventColumns(l)
		rows := buildTableRows(resp.GetData(), status, l, deps.Routes, perms)
		types.ApplyColumnStyles(columns, rows)

		tableConfig := &types.TableConfig{
			ID:                   "events-table",
			Columns:              columns,
			Rows:                 rows,
			ShowSearch:           true,
			ShowActions:          true,
			ShowSort:             true,
			ShowColumns:          true,
			ShowDensity:          true,
			ShowEntries:          true,
			DefaultSortColumn:    "start_date",
			DefaultSortDirection: "asc",
			Labels:               deps.TableLabels,
			EmptyState: types.TableEmptyState{
				Title:   l.Empty.Heading,
				Message: l.Empty.Subheading,
			},
			PrimaryAction: &types.PrimaryAction{
				Label:           l.Buttons.AddEvent,
				ActionURL:       deps.Routes.AddURL,
				Icon:            "icon-plus",
				Disabled:        !perms.Can("event", "create"),
				DisabledTooltip: l.Errors.NameRequired,
			},
		}
		types.ApplyTableSettings(tableConfig)

		pageData := &PageData{
			PageData: types.PageData{
				CacheVersion:   viewCtx.CacheVersion,
				Title:          statusPageTitle(l, status),
				CurrentPath:    viewCtx.CurrentPath,
				ActiveNav:      deps.Routes.ActiveNav,
				ActiveSubNav:   deps.Routes.ActiveSubNav,
				HeaderTitle:    statusPageTitle(l, status),
				HeaderSubtitle: l.Page.Caption,
				HeaderIcon:     "icon-calendar",
				CommonLabels:   deps.CommonLabels,
			},
			ContentTemplate: "event-list-content",
			Table:           tableConfig,
		}

		// KB help content
		if viewCtx.Translations != nil {
			if provider, ok := viewCtx.Translations.(*lynguaV1.TranslationProvider); ok {
				if kb, _ := provider.LoadKBIfExists(viewCtx.Lang, viewCtx.BusinessType, "event"); kb != nil {
					pageData.HasHelp = true
					pageData.HelpContent = kb.Body
				}
			}
		}

		return view.OK("event-list", pageData)
	})
}

func eventColumns(l cyta.EventLabels) []types.TableColumn {
	return []types.TableColumn{
		{Key: "name", Label: l.Columns.Name, Sortable: true},
		{Key: "start_date", Label: l.Columns.StartDate, Sortable: true, Width: "150px"},
		{Key: "end_date", Label: l.Columns.EndDate, Sortable: true, Width: "150px"},
		{Key: "organizer", Label: l.Columns.Organizer, Sortable: true},
		{Key: "location", Label: l.Columns.Location, Sortable: true},
		{Key: "status", Label: l.Columns.Status, Sortable: true, Width: "130px"},
		{Key: "recurs", Label: l.Columns.Recurs, Sortable: false, Width: "80px"},
	}
}

func buildTableRows(
	events []*eventpb.Event,
	status string,
	l cyta.EventLabels,
	routes cyta.EventRoutes,
	perms *types.UserPermissions,
) []types.TableRow {
	rows := []types.TableRow{}
	for _, e := range events {
		eventStatus := eventStatusString(e.GetStatus())
		if !matchesStatusFilter(eventStatus, status) {
			continue
		}

		id := e.GetId()
		name := e.GetName()
		startDate := e.GetStartDateTimeUtcString()
		endDate := e.GetEndDateTimeUtcString()
		recurs := ""
		if e.GetEventRecurrenceId() != "" {
			recurs = "Yes"
		}
		detailURL := route.ResolveURL(routes.DetailURL, "id", id)

		rows = append(rows, types.TableRow{
			ID:   id,
			Href: detailURL,
			Cells: []types.TableCell{
				{Type: "text", Value: name},
				types.DateTimeCell(startDate, types.DateReadable),
				types.DateTimeCell(endDate, types.DateReadable),
				{Type: "text", Value: e.GetOrganizerId()},
				{Type: "text", Value: e.GetLocationId()},
				{Type: "badge", Value: eventStatus, Variant: eventStatusVariant(eventStatus)},
				{Type: "text", Value: recurs},
			},
			DataAttrs: map[string]string{
				"name":       name,
				"start_date": startDate,
				"end_date":   endDate,
				"status":     eventStatus,
				"recurs":     recurs,
			},
			Actions: []types.TableAction{
				{Type: "view", Label: l.Actions.Edit, Action: "view", Href: detailURL},
				{Type: "edit", Label: l.Actions.Edit, Action: "edit", URL: route.ResolveURL(routes.EditURL, "id", id), DrawerTitle: l.Actions.Edit, Disabled: !perms.Can("event", "update"), DisabledTooltip: l.Errors.NameRequired},
				{Type: "delete", Label: l.Actions.Delete, Action: "delete", URL: routes.DeleteURL, ItemName: name, Disabled: !perms.Can("event", "delete"), DisabledTooltip: l.Errors.NameRequired},
			},
		})
	}
	return rows
}

// matchesStatusFilter returns true if the event status matches the requested filter.
func matchesStatusFilter(eventStatus, filterStatus string) bool {
	switch filterStatus {
	case "upcoming":
		return eventStatus == "tentative" || eventStatus == "confirmed"
	case "confirmed":
		return eventStatus == "confirmed"
	case "completed":
		return eventStatus == "completed"
	case "cancelled":
		return eventStatus == "cancelled"
	default:
		return true
	}
}

func eventStatusString(s eventpb.EventStatus) string {
	switch s {
	case eventpb.EventStatus_EVENT_STATUS_TENTATIVE:
		return "tentative"
	case eventpb.EventStatus_EVENT_STATUS_CONFIRMED:
		return "confirmed"
	case eventpb.EventStatus_EVENT_STATUS_CANCELLED:
		return "cancelled"
	default:
		return "tentative"
	}
}

func eventStatusVariant(status string) string {
	switch status {
	case "tentative":
		return "warning"
	case "confirmed":
		return "success"
	case "cancelled":
		return "default"
	default:
		return "default"
	}
}

func statusPageTitle(l cyta.EventLabels, status string) string {
	switch status {
	case "upcoming":
		return l.Page.HeadingUpcoming
	case "confirmed":
		return l.Page.HeadingConfirmed
	case "completed":
		return l.Page.HeadingCompleted
	case "cancelled":
		return l.Page.HeadingCancelled
	default:
		return l.Page.Heading
	}
}
