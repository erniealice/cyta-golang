// Package action implements the event-tag drawer actions (Add, Edit, Delete).
//
// Mirrors packages/entydad-golang/views/role/action/action.go but without the
// permissions matrix (event_tag has no child rows). The drawer template is
// "event-tag-drawer-form" (see ../templates/event-tag-drawer-form.html).
package action

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/erniealice/pyeza-golang/route"
	"github.com/erniealice/pyeza-golang/view"

	eventtagpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/event/event_tag"

	cyta "github.com/erniealice/cyta-golang"
)

// defaultColor is used when a drawer POST arrives with an empty color field.
const defaultColor = "#6B7280"

// FormLabels holds i18n labels for the drawer form template.
// Flat shape consumed directly by the template (no shared LabelsFromX mapper —
// this form has a single source struct, cyta.EventTagFormLabels).
type FormLabels struct {
	Name                   string
	NamePlaceholder        string
	Description            string
	DescriptionPlaceholder string
	Color                  string
	ColorPlaceholder       string
	Active                 string
}

// FormData is the template data for the event-tag drawer form.
type FormData struct {
	FormAction   string
	IsEdit       bool
	ID           string
	Name         string
	Description  string
	Color        string
	Active       bool
	Labels       FormLabels
	CommonLabels any
}

// Deps holds dependencies for event-tag action handlers.
type Deps struct {
	Routes              cyta.EventTagRoutes
	Labels              cyta.EventTagLabels
	CreateEventTag      func(ctx context.Context, req *eventtagpb.CreateEventTagRequest) (*eventtagpb.CreateEventTagResponse, error)
	ReadEventTag        func(ctx context.Context, req *eventtagpb.ReadEventTagRequest) (*eventtagpb.ReadEventTagResponse, error)
	UpdateEventTag      func(ctx context.Context, req *eventtagpb.UpdateEventTagRequest) (*eventtagpb.UpdateEventTagResponse, error)
	DeleteEventTag      func(ctx context.Context, req *eventtagpb.DeleteEventTagRequest) (*eventtagpb.DeleteEventTagResponse, error)
	GetEventTagInUseIDs func(ctx context.Context, ids []string) (map[string]bool, error)
}

// formLabels maps cyta.EventTagFormLabels onto the flat template struct.
// Keeps the template-facing shape independent of cyta's label-struct layout.
func formLabels(src cyta.EventTagFormLabels) FormLabels {
	return FormLabels{
		Name:                   src.Name,
		NamePlaceholder:        src.NamePlaceholder,
		Description:            src.Description,
		DescriptionPlaceholder: src.DescriptionPlaceholder,
		Color:                  src.Color,
		ColorPlaceholder:       src.ColorPlaceholder,
		Active:                 src.Active,
	}
}

// NewAddAction creates the event-tag add action (GET = form, POST = create).
func NewAddAction(deps *Deps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("event_tag", "create") {
			return htmxError(viewCtx.T("shared.errors.permissionDenied"))
		}

		if viewCtx.Request.Method == http.MethodGet {
			return view.OK("event-tag-drawer-form", &FormData{
				FormAction:   deps.Routes.AddURL,
				Active:       true,
				Color:        defaultColor,
				Labels:       formLabels(deps.Labels.Form),
				CommonLabels: nil, // injected by ViewAdapter
			})
		}

		// POST — create event_tag
		if err := viewCtx.Request.ParseForm(); err != nil {
			return htmxError(viewCtx.T("shared.errors.invalidFormData"))
		}

		r := viewCtx.Request
		color := r.FormValue("color")
		if color == "" {
			color = defaultColor
		}
		active := r.FormValue("active") == "true"

		_, err := deps.CreateEventTag(ctx, &eventtagpb.CreateEventTagRequest{
			Data: &eventtagpb.EventTag{
				Name:        r.FormValue("name"),
				Description: r.FormValue("description"),
				Color:       color,
				Active:      active,
			},
		})
		if err != nil {
			log.Printf("Failed to create event_tag: %v", err)
			return htmxError(err.Error())
		}

		return htmxSuccess("event-tags-table")
	})
}

// NewEditAction creates the event-tag edit action (GET = form, POST = update).
func NewEditAction(deps *Deps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("event_tag", "update") {
			return htmxError(viewCtx.T("shared.errors.permissionDenied"))
		}
		id := viewCtx.Request.PathValue("id")

		if viewCtx.Request.Method == http.MethodGet {
			resp, err := deps.ReadEventTag(ctx, &eventtagpb.ReadEventTagRequest{
				Data: &eventtagpb.EventTag{Id: id},
			})
			if err != nil {
				log.Printf("Failed to read event_tag %s: %v", id, err)
				return htmxError(viewCtx.T("shared.errors.notFound"))
			}
			data := resp.GetData()
			if len(data) == 0 {
				return htmxError(viewCtx.T("shared.errors.notFound"))
			}
			tag := data[0]

			return view.OK("event-tag-drawer-form", &FormData{
				FormAction:   route.ResolveURL(deps.Routes.EditURL, "id", id),
				IsEdit:       true,
				ID:           id,
				Name:         tag.GetName(),
				Description:  tag.GetDescription(),
				Color:        tag.GetColor(),
				Active:       tag.GetActive(),
				Labels:       formLabels(deps.Labels.Form),
				CommonLabels: nil, // injected by ViewAdapter
			})
		}

		// POST — update event_tag
		if err := viewCtx.Request.ParseForm(); err != nil {
			return htmxError(viewCtx.T("shared.errors.invalidFormData"))
		}

		r := viewCtx.Request
		color := r.FormValue("color")
		if color == "" {
			color = defaultColor
		}
		active := r.FormValue("active") == "true"

		_, err := deps.UpdateEventTag(ctx, &eventtagpb.UpdateEventTagRequest{
			Data: &eventtagpb.EventTag{
				Id:          id,
				Name:        r.FormValue("name"),
				Description: r.FormValue("description"),
				Color:       color,
				Active:      active,
			},
		})
		if err != nil {
			log.Printf("Failed to update event_tag %s: %v", id, err)
			return htmxError(err.Error())
		}

		return htmxSuccess("event-tags-table")
	})
}

// NewDeleteAction creates the event-tag delete action (POST only).
// The row ID comes via the URL path (EventTagDeleteURL = "/action/schedule/tag/delete/{id}").
//
// Before deletion the reference checker is consulted: if the tag is attached
// to any active events (via event_tag_assignment), deletion is blocked with
// an in-use error.
func NewDeleteAction(deps *Deps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("event_tag", "delete") {
			return htmxError(viewCtx.T("shared.errors.permissionDenied"))
		}
		id := viewCtx.Request.PathValue("id")
		if id == "" {
			id = viewCtx.Request.URL.Query().Get("id")
		}
		if id == "" {
			_ = viewCtx.Request.ParseForm()
			id = viewCtx.Request.FormValue("id")
		}
		if id == "" {
			return htmxError(viewCtx.T("shared.errors.idRequired"))
		}

		// Delete-guard: block when the tag is in use by an active assignment.
		// TODO: once EventTagLabels gains an Errors sub-struct, replace the
		// hardcoded English fallback with the translated string.
		if deps.GetEventTagInUseIDs != nil {
			inUse, err := deps.GetEventTagInUseIDs(ctx, []string{id})
			if err == nil && inUse[id] {
				return htmxError("Cannot delete tag: it is in use by one or more active events.")
			}
		}

		_, err := deps.DeleteEventTag(ctx, &eventtagpb.DeleteEventTagRequest{
			Data: &eventtagpb.EventTag{Id: id},
		})
		if err != nil {
			log.Printf("Failed to delete event_tag %s: %v", id, err)
			return htmxError(err.Error())
		}

		return htmxSuccess("event-tags-table")
	})
}

// ---------------------------------------------------------------------------
// HTMX response helpers
// ---------------------------------------------------------------------------

// htmxSuccess returns a header-only response that signals the sheet to close
// and the named table to refresh.
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
