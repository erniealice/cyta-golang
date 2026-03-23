package detail

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

// PageData holds the data for the event detail page.
type PageData struct {
	types.PageData
	ContentTemplate   string
	Event             map[string]any
	Labels            cyta.EventLabels
	ActiveTab         string
	TabItems          []pyeza.TabItem
	AttendeesTable    *types.TableConfig
	ResourcesTable    *types.TableConfig
	ProductsTable     *types.TableConfig
	OccurrencesTable  *types.TableConfig
}

// eventToMap converts an Event protobuf to a map[string]any for template use.
func eventToMap(e *eventpb.Event) map[string]any {
	return map[string]any{
		"id":                       e.GetId(),
		"name":                     e.GetName(),
		"description":              e.GetDescription(),
		"start_date_time_utc":      e.GetStartDateTimeUtc(),
		"end_date_time_utc":        e.GetEndDateTimeUtc(),
		"start_date_time_string":   e.GetStartDateTimeUtcString(),
		"end_date_time_string":     e.GetEndDateTimeUtcString(),
		"timezone":                 e.GetTimezone(),
		"all_day":                  e.GetAllDay(),
		"organizer_id":             e.GetOrganizerId(),
		"location_id":              e.GetLocationId(),
		"event_recurrence_id":      e.GetEventRecurrenceId(),
		"parent_event_id":          e.GetParentEventId(),
		"workspace_id":             e.GetWorkspaceId(),
		"status":                   eventStatusString(e.GetStatus()),
		"status_variant":           eventStatusVariant(e.GetStatus()),
		"active":                   e.GetActive(),
		"date_created_string":      e.GetDateCreatedString(),
		"date_modified_string":     e.GetDateModifiedString(),
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

func eventStatusVariant(s eventpb.EventStatus) string {
	switch s {
	case eventpb.EventStatus_EVENT_STATUS_TENTATIVE:
		return "warning"
	case eventpb.EventStatus_EVENT_STATUS_CONFIRMED:
		return "success"
	case eventpb.EventStatus_EVENT_STATUS_CANCELLED:
		return "default"
	default:
		return "warning"
	}
}

// NewView creates the event detail view.
func NewView(deps *DetailViewDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		id := viewCtx.Request.PathValue("id")

		resp, err := deps.ReadEvent(ctx, &eventpb.ReadEventRequest{
			Data: &eventpb.Event{Id: id},
		})
		if err != nil {
			log.Printf("Failed to read event %s: %v", id, err)
			return view.Error(fmt.Errorf("failed to load event: %w", err))
		}
		data := resp.GetData()
		if len(data) == 0 {
			log.Printf("Event %s not found", id)
			return view.Error(fmt.Errorf("event not found"))
		}
		event := eventToMap(data[0])

		eventName, _ := event["name"].(string)
		headerTitle := eventName

		l := deps.Labels

		activeTab := viewCtx.QueryParams["tab"]
		if activeTab == "" {
			activeTab = "overview"
		}
		tabItems := buildTabItems(l, id, deps.Routes)

		pageData := &PageData{
			PageData: types.PageData{
				CacheVersion:   viewCtx.CacheVersion,
				Title:          headerTitle,
				CurrentPath:    viewCtx.CurrentPath,
				ActiveNav:      deps.Routes.ActiveNav,
				HeaderTitle:    headerTitle,
				HeaderSubtitle: l.Detail.Heading,
				HeaderIcon:     "icon-calendar",
				CommonLabels:   deps.CommonLabels,
			},
			ContentTemplate: "event-detail-content",
			Event:           event,
			Labels:          l,
			ActiveTab:       activeTab,
			TabItems:        tabItems,
		}

		// KB help content
		if viewCtx.Translations != nil {
			if provider, ok := viewCtx.Translations.(*lynguaV1.TranslationProvider); ok {
				if kb, _ := provider.LoadKBIfExists(viewCtx.Lang, viewCtx.BusinessType, "events-detail"); kb != nil {
					pageData.HasHelp = true
					pageData.HelpContent = kb.Body
				}
			}
		}

		// Load tab-specific data
		loadTabData(ctx, deps, pageData, id, activeTab)

		return view.OK("event-detail", pageData)
	})
}

func buildTabItems(l cyta.EventLabels, id string, routes cyta.EventRoutes) []pyeza.TabItem {
	base := route.ResolveURL(routes.DetailURL, "id", id)
	action := route.ResolveURL(routes.TabActionURL, "id", id, "tab", "")
	return []pyeza.TabItem{
		{Key: "overview", Label: l.Tabs.Overview, Href: base + "?tab=overview", HxGet: action + "overview", Icon: "icon-info"},
		{Key: "attendees", Label: l.Tabs.Attendees, Href: base + "?tab=attendees", HxGet: action + "attendees", Icon: "icon-users"},
		{Key: "resources", Label: l.Tabs.Resources, Href: base + "?tab=resources", HxGet: action + "resources", Icon: "icon-box"},
		{Key: "products", Label: l.Tabs.Products, Href: base + "?tab=products", HxGet: action + "products", Icon: "icon-tag"},
		{Key: "occurrences", Label: l.Tabs.Occurrences, Href: base + "?tab=occurrences", HxGet: action + "occurrences", Icon: "icon-repeat"},
	}
}

// loadTabData populates the PageData with tab-specific data.
func loadTabData(ctx context.Context, deps *DetailViewDeps, pageData *PageData, id string, activeTab string) {
	switch activeTab {
	case "overview":
		// event map has all the overview fields
	case "attendees":
		loadAttendeesTab(ctx, deps, pageData, id)
	case "resources":
		loadResourcesTab(ctx, deps, pageData, id)
	case "products":
		loadProductsTab(ctx, deps, pageData, id)
	case "occurrences":
		loadOccurrencesTab(ctx, deps, pageData, id)
	}
}

// NewTabAction creates the tab action view (partial — returns only the tab content).
func NewTabAction(deps *DetailViewDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		id := viewCtx.Request.PathValue("id")
		tab := viewCtx.Request.PathValue("tab")
		if tab == "" {
			tab = "overview"
		}

		resp, err := deps.ReadEvent(ctx, &eventpb.ReadEventRequest{
			Data: &eventpb.Event{Id: id},
		})
		if err != nil {
			log.Printf("Failed to read event %s: %v", id, err)
			return view.Error(fmt.Errorf("failed to load event: %w", err))
		}
		data := resp.GetData()
		if len(data) == 0 {
			log.Printf("Event %s not found", id)
			return view.Error(fmt.Errorf("event not found"))
		}
		event := eventToMap(data[0])

		l := deps.Labels
		pageData := &PageData{
			PageData: types.PageData{
				CacheVersion: viewCtx.CacheVersion,
				CommonLabels: deps.CommonLabels,
			},
			Event:     event,
			Labels:    l,
			ActiveTab: tab,
			TabItems:  buildTabItems(l, id, deps.Routes),
		}

		// Load tab-specific data
		loadTabData(ctx, deps, pageData, id, tab)

		templateName := "event-tab-" + tab
		return view.OK(templateName, pageData)
	})
}
