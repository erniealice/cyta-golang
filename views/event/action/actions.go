package action

import (
	"context"
	"fmt"
	"log"
	"net/http"

	cyta "github.com/erniealice/cyta-golang"

	"github.com/erniealice/pyeza-golang/route"
	"github.com/erniealice/pyeza-golang/view"

	eventpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/event/event"
)

// FormData is the template data for the event drawer form.
type FormData struct {
	FormAction   string
	IsEdit       bool
	ID           string
	Name         string
	Description  string
	StartDate    string
	EndDate      string
	Timezone     string
	AllDay       bool
	OrganizerID  string
	LocationID   string
	Status       eventpb.EventStatus
	Labels       cyta.EventLabels
	CommonLabels any
}

// Deps holds dependencies for event action handlers.
type Deps struct {
	Routes      cyta.EventRoutes
	Labels      cyta.EventLabels
	CreateEvent func(ctx context.Context, req *eventpb.CreateEventRequest) (*eventpb.CreateEventResponse, error)
	ReadEvent   func(ctx context.Context, req *eventpb.ReadEventRequest) (*eventpb.ReadEventResponse, error)
	UpdateEvent func(ctx context.Context, req *eventpb.UpdateEventRequest) (*eventpb.UpdateEventResponse, error)
	DeleteEvent func(ctx context.Context, req *eventpb.DeleteEventRequest) (*eventpb.DeleteEventResponse, error)
	ListEvents  func(ctx context.Context, req *eventpb.ListEventsRequest) (*eventpb.ListEventsResponse, error)
}

// NewAddAction creates the event add action (GET = form, POST = create).
func NewAddAction(deps *Deps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("event", "create") {
			return htmxError(deps.Labels.Errors.NameRequired)
		}

		if viewCtx.Request.Method == http.MethodGet {
			return view.OK("event-drawer-form", &FormData{
				FormAction:   deps.Routes.AddURL,
				Labels:       deps.Labels,
				CommonLabels: nil, // injected by ViewAdapter
			})
		}

		// POST — create event
		if err := viewCtx.Request.ParseForm(); err != nil {
			return htmxError("Invalid form data")
		}

		r := viewCtx.Request

		resp, err := deps.CreateEvent(ctx, &eventpb.CreateEventRequest{
			Data: &eventpb.Event{
				Name:        r.FormValue("name"),
				Description: strPtr(r.FormValue("description")),
				Timezone:    r.FormValue("timezone"),
				OrganizerId: strPtr(r.FormValue("organizer_id")),
				LocationId:  strPtr(r.FormValue("location_id")),
				Status:      eventpb.EventStatus_EVENT_STATUS_TENTATIVE,
			},
		})
		if err != nil {
			log.Printf("Failed to create event: %v", err)
			return htmxError(err.Error())
		}

		newID := ""
		if respData := resp.GetData(); len(respData) > 0 {
			newID = respData[0].GetId()
		}
		if newID != "" {
			return view.ViewResult{
				StatusCode: http.StatusOK,
				Headers: map[string]string{
					"HX-Trigger":  `{"formSuccess":true}`,
					"HX-Redirect": route.ResolveURL(deps.Routes.DetailURL, "id", newID),
				},
			}
		}

		return htmxSuccess("events-table")
	})
}

// NewEditAction creates the event edit action (GET = form, POST = update).
func NewEditAction(deps *Deps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("event", "update") {
			return htmxError(deps.Labels.Errors.NameRequired)
		}

		id := viewCtx.Request.PathValue("id")

		if viewCtx.Request.Method == http.MethodGet {
			readResp, err := deps.ReadEvent(ctx, &eventpb.ReadEventRequest{
				Data: &eventpb.Event{Id: id},
			})
			if err != nil {
				log.Printf("Failed to read event %s: %v", id, err)
				return htmxError(deps.Labels.Errors.NameRequired)
			}
			readData := readResp.GetData()
			if len(readData) == 0 {
				return htmxError(deps.Labels.Errors.NameRequired)
			}
			record := readData[0]

			return view.OK("event-drawer-form", &FormData{
				FormAction:   route.ResolveURL(deps.Routes.EditURL, "id", id),
				IsEdit:       true,
				ID:           id,
				Name:         record.GetName(),
				Description:  record.GetDescription(),
				Timezone:     record.GetTimezone(),
				AllDay:       record.GetAllDay(),
				OrganizerID:  record.GetOrganizerId(),
				LocationID:   record.GetLocationId(),
				Status:       record.GetStatus(),
				Labels:       deps.Labels,
				CommonLabels: nil, // injected by ViewAdapter
			})
		}

		// POST — update event
		if err := viewCtx.Request.ParseForm(); err != nil {
			return htmxError("Invalid form data")
		}

		r := viewCtx.Request

		_, err := deps.UpdateEvent(ctx, &eventpb.UpdateEventRequest{
			Data: &eventpb.Event{
				Id:          id,
				Name:        r.FormValue("name"),
				Description: strPtr(r.FormValue("description")),
				Timezone:    r.FormValue("timezone"),
				OrganizerId: strPtr(r.FormValue("organizer_id")),
				LocationId:  strPtr(r.FormValue("location_id")),
			},
		})
		if err != nil {
			log.Printf("Failed to update event %s: %v", id, err)
			return htmxError(err.Error())
		}

		return view.ViewResult{
			StatusCode: http.StatusOK,
			Headers: map[string]string{
				"HX-Trigger":  `{"formSuccess":true}`,
				"HX-Redirect": route.ResolveURL(deps.Routes.DetailURL, "id", id),
			},
		}
	})
}

// NewDeleteAction creates the event delete action (POST only).
func NewDeleteAction(deps *Deps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("event", "delete") {
			return htmxError(deps.Labels.Errors.NameRequired)
		}

		id := viewCtx.Request.URL.Query().Get("id")
		if id == "" {
			_ = viewCtx.Request.ParseForm()
			id = viewCtx.Request.FormValue("id")
		}
		if id == "" {
			return htmxError("ID is required")
		}

		_, err := deps.DeleteEvent(ctx, &eventpb.DeleteEventRequest{
			Data: &eventpb.Event{Id: id},
		})
		if err != nil {
			log.Printf("Failed to delete event %s: %v", id, err)
			return htmxError(err.Error())
		}

		return htmxSuccess("events-table")
	})
}

// NewBulkDeleteAction creates the event bulk delete action (POST only).
func NewBulkDeleteAction(deps *Deps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("event", "delete") {
			return htmxError(deps.Labels.Errors.NameRequired)
		}

		_ = viewCtx.Request.ParseMultipartForm(32 << 20)

		ids := viewCtx.Request.Form["id"]
		if len(ids) == 0 {
			return htmxError("No IDs provided")
		}

		for _, id := range ids {
			_, err := deps.DeleteEvent(ctx, &eventpb.DeleteEventRequest{
				Data: &eventpb.Event{Id: id},
			})
			if err != nil {
				log.Printf("Failed to delete event %s: %v", id, err)
			}
		}

		return htmxSuccess("events-table")
	})
}

// NewSetStatusAction creates the event status update action (POST only).
func NewSetStatusAction(deps *Deps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("event", "update") {
			return htmxError(deps.Labels.Errors.NameRequired)
		}

		id := viewCtx.Request.URL.Query().Get("id")
		targetStatus := viewCtx.Request.URL.Query().Get("status")

		if id == "" {
			_ = viewCtx.Request.ParseForm()
			id = viewCtx.Request.FormValue("id")
			targetStatus = viewCtx.Request.FormValue("status")
		}
		if id == "" {
			return htmxError("ID is required")
		}
		if targetStatus == "" {
			return htmxError("Status is required")
		}

		statusEnum := eventStatusToEnum(targetStatus)

		_, err := deps.UpdateEvent(ctx, &eventpb.UpdateEventRequest{
			Data: &eventpb.Event{Id: id, Status: statusEnum},
		})
		if err != nil {
			log.Printf("Failed to update event status %s: %v", id, err)
			return htmxError(err.Error())
		}

		return htmxSuccess("events-table")
	})
}

// NewBulkSetStatusAction creates the event bulk status update action (POST only).
func NewBulkSetStatusAction(deps *Deps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("event", "update") {
			return htmxError(deps.Labels.Errors.NameRequired)
		}

		_ = viewCtx.Request.ParseMultipartForm(32 << 20)

		ids := viewCtx.Request.Form["id"]
		targetStatus := viewCtx.Request.FormValue("target_status")

		if len(ids) == 0 {
			return htmxError("No IDs provided")
		}
		if targetStatus == "" {
			return htmxError("Target status is required")
		}

		statusEnum := eventStatusToEnum(targetStatus)

		for _, id := range ids {
			if _, err := deps.UpdateEvent(ctx, &eventpb.UpdateEventRequest{
				Data: &eventpb.Event{Id: id, Status: statusEnum},
			}); err != nil {
				log.Printf("Failed to update event status %s: %v", id, err)
			}
		}

		return htmxSuccess("events-table")
	})
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// strPtr returns a pointer to a string.
func strPtr(s string) *string {
	return &s
}

// eventStatusToEnum converts a status string to the protobuf EventStatus enum.
func eventStatusToEnum(status string) eventpb.EventStatus {
	switch status {
	case "tentative":
		return eventpb.EventStatus_EVENT_STATUS_TENTATIVE
	case "confirmed":
		return eventpb.EventStatus_EVENT_STATUS_CONFIRMED
	case "cancelled":
		return eventpb.EventStatus_EVENT_STATUS_CANCELLED
	default:
		return eventpb.EventStatus_EVENT_STATUS_UNSPECIFIED
	}
}

// htmxSuccess returns a header-only response that signals the sheet to close and table to refresh.
func htmxSuccess(tableID string) view.ViewResult {
	return view.ViewResult{
		StatusCode: http.StatusOK,
		Headers: map[string]string{
			"HX-Trigger": fmt.Sprintf(`{"formSuccess":true,"refreshTable":"%s"}`, tableID),
		},
	}
}

// htmxError returns a header-only response that signals a form error.
func htmxError(message string) view.ViewResult {
	return view.ViewResult{
		StatusCode: http.StatusUnprocessableEntity,
		Headers: map[string]string{
			"HX-Error-Message": message,
		},
	}
}
