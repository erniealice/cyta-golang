// Package event_tag wires the cyta event-tag views (list page + drawer form)
// into a pyeza Module. The caller (cyta/block/block.go) supplies use-case
// function pointers from espyna-golang's event.UseCases.EventTag; that
// indirection keeps this package free of any espyna import.
package event_tag

import (
	"context"

	cyta "github.com/erniealice/cyta-golang"
	cytaeventtagaction "github.com/erniealice/cyta-golang/views/event_tag/action"
	cytaeventtaglist "github.com/erniealice/cyta-golang/views/event_tag/list"

	eventtagpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/event/event_tag"

	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"
)

// ModuleDeps holds all dependencies for the event-tag module.
type ModuleDeps struct {
	Routes       cyta.EventTagRoutes
	Labels       cyta.EventTagLabels
	CommonLabels pyeza.CommonLabels
	TableLabels  types.TableLabels

	// Use-case function pointers — satisfied in block.go from
	// espyna-golang's event.UseCases.EventTag.*.Execute.
	CreateEventTag          func(ctx context.Context, req *eventtagpb.CreateEventTagRequest) (*eventtagpb.CreateEventTagResponse, error)
	ReadEventTag            func(ctx context.Context, req *eventtagpb.ReadEventTagRequest) (*eventtagpb.ReadEventTagResponse, error)
	UpdateEventTag          func(ctx context.Context, req *eventtagpb.UpdateEventTagRequest) (*eventtagpb.UpdateEventTagResponse, error)
	DeleteEventTag          func(ctx context.Context, req *eventtagpb.DeleteEventTagRequest) (*eventtagpb.DeleteEventTagResponse, error)
	ListEventTags           func(ctx context.Context, req *eventtagpb.ListEventTagsRequest) (*eventtagpb.ListEventTagsResponse, error)
	GetEventTagListPageData func(ctx context.Context, req *eventtagpb.GetEventTagListPageDataRequest) (*eventtagpb.GetEventTagListPageDataResponse, error)

	// Reference-check for the delete-guard — satisfied by
	// reference.Checker.GetEventTagInUseIDs.
	GetEventTagInUseIDs func(ctx context.Context, ids []string) (map[string]bool, error)
}

// Module holds all constructed event-tag views.
type Module struct {
	routes cyta.EventTagRoutes
	List   view.View
	Add    view.View
	Edit   view.View
	Delete view.View
}

// NewModule creates a new event-tag module with all views wired.
func NewModule(deps *ModuleDeps) *Module {
	actionDeps := &cytaeventtagaction.Deps{
		Routes:              deps.Routes,
		Labels:              deps.Labels,
		CreateEventTag:      deps.CreateEventTag,
		ReadEventTag:        deps.ReadEventTag,
		UpdateEventTag:      deps.UpdateEventTag,
		DeleteEventTag:      deps.DeleteEventTag,
		GetEventTagInUseIDs: deps.GetEventTagInUseIDs,
	}
	listDeps := &cytaeventtaglist.Deps{
		Routes:                  deps.Routes,
		Labels:                  deps.Labels,
		CommonLabels:            deps.CommonLabels,
		TableLabels:             deps.TableLabels,
		GetEventTagListPageData: deps.GetEventTagListPageData,
		GetEventTagInUseIDs:     deps.GetEventTagInUseIDs,
	}

	return &Module{
		routes: deps.Routes,
		List:   cytaeventtaglist.NewView(listDeps),
		Add:    cytaeventtagaction.NewAddAction(actionDeps),
		Edit:   cytaeventtagaction.NewEditAction(actionDeps),
		Delete: cytaeventtagaction.NewDeleteAction(actionDeps),
	}
}

// RegisterRoutes registers all event-tag routes against the pyeza registrar.
func (m *Module) RegisterRoutes(r view.RouteRegistrar) {
	r.GET(m.routes.ListURL, m.List)
	r.GET(m.routes.AddURL, m.Add)
	r.POST(m.routes.AddURL, m.Add)
	r.GET(m.routes.EditURL, m.Edit)
	r.POST(m.routes.EditURL, m.Edit)
	r.POST(m.routes.DeleteURL, m.Delete)
}
