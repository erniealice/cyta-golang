package event

import (
	"context"

	cyta "github.com/erniealice/cyta-golang"

	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"

	attachmentpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/document/attachment"
	eventpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/event/event"
	eventattendeepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/event/event_attendee"
	eventoccurrencepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/event/event_occurrence"
	eventproductpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/event/event_product"
	eventresourcepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/event/event_resource"

	eventaction "github.com/erniealice/cyta-golang/views/event/action"
	eventcalendar "github.com/erniealice/cyta-golang/views/event/calendar"
	eventdashboard "github.com/erniealice/cyta-golang/views/event/dashboard"
	eventdetail "github.com/erniealice/cyta-golang/views/event/detail"
	eventform "github.com/erniealice/cyta-golang/views/event/form"
	eventlist "github.com/erniealice/cyta-golang/views/event/list"
)

// ModuleDeps holds all dependencies for the event module.
type ModuleDeps struct {
	Routes           cyta.EventRoutes
	EventTagRoutes   cyta.EventTagRoutes
	RecurrenceRoutes cyta.RecurrenceRoutes
	Labels           cyta.EventLabels
	CommonLabels     pyeza.CommonLabels
	TableLabels      types.TableLabels

	// Event CRUD
	CreateEvent func(ctx context.Context, req *eventpb.CreateEventRequest) (*eventpb.CreateEventResponse, error)
	ReadEvent   func(ctx context.Context, req *eventpb.ReadEventRequest) (*eventpb.ReadEventResponse, error)
	UpdateEvent func(ctx context.Context, req *eventpb.UpdateEventRequest) (*eventpb.UpdateEventResponse, error)
	DeleteEvent func(ctx context.Context, req *eventpb.DeleteEventRequest) (*eventpb.DeleteEventResponse, error)
	ListEvents  func(ctx context.Context, req *eventpb.ListEventsRequest) (*eventpb.ListEventsResponse, error)

	// Attendee operations
	ListEventAttendees func(ctx context.Context, req *eventattendeepb.ListEventAttendeesRequest) (*eventattendeepb.ListEventAttendeesResponse, error)

	// Resource operations
	ListEventResources func(ctx context.Context, req *eventresourcepb.ListEventResourcesRequest) (*eventresourcepb.ListEventResourcesResponse, error)

	// Product operations
	ListEventProducts func(ctx context.Context, req *eventproductpb.ListEventProductsRequest) (*eventproductpb.ListEventProductsResponse, error)

	// Occurrence operations
	ListEventOccurrences func(ctx context.Context, req *eventoccurrencepb.ListEventOccurrencesRequest) (*eventoccurrencepb.ListEventOccurrencesResponse, error)

	// Phase 4 — drawer pickers (all nillable; degrade gracefully when nil)
	ListEventTags          func(ctx context.Context) ([]eventform.Option, error)
	ListEventTagsForEvent  func(ctx context.Context, eventID string) ([]string, error)
	SearchAttendees        func(ctx context.Context, query string) ([]eventform.Option, error)
	ListAttendeesForEvent  func(ctx context.Context, eventID string) ([]eventform.SelectedOption, error)
	SetEventTagAssignments func(ctx context.Context, eventID string, tagIDs []string) error
	SyncEventAttendees     func(ctx context.Context, eventID string, attendeeRefs []string) error
	ListEventAttachments   func(ctx context.Context, eventID string) ([]eventform.Attachment, error)

	// Phase 6 — schedule dashboard read-only projection callback.
	// Nillable; when nil the dashboard renders empty stats / widgets.
	GetScheduleDashboardData func(ctx context.Context, req *eventdashboard.Request) (*eventdashboard.Response, error)

	// Hybra attachment ops (nillable; degrade gracefully when nil).
	UploadFile       func(ctx context.Context, bucket, key string, content []byte, contentType string) error
	ListAttachments  func(ctx context.Context, moduleKey, foreignKey string) (*attachmentpb.ListAttachmentsResponse, error)
	CreateAttachment func(ctx context.Context, req *attachmentpb.CreateAttachmentRequest) (*attachmentpb.CreateAttachmentResponse, error)
	DeleteAttachment func(ctx context.Context, req *attachmentpb.DeleteAttachmentRequest) (*attachmentpb.DeleteAttachmentResponse, error)
	NewID            func() string
}

// Module holds all constructed event views.
type Module struct {
	routes           cyta.EventRoutes
	List             view.View
	Detail           view.View
	TabAction        view.View
	Add              view.View
	Edit             view.View
	Delete           view.View
	BulkDelete       view.View
	SetStatus        view.View
	BulkSetStatus    view.View
	Calendar         view.View
	Dashboard        view.View
	AttachmentUpload view.View
	AttachmentDelete view.View
}

// NewModule creates a new event module with all views wired.
func NewModule(deps *ModuleDeps) *Module {
	detailDeps := &eventdetail.DetailViewDeps{
		Routes:               deps.Routes,
		Labels:               deps.Labels,
		CommonLabels:         deps.CommonLabels,
		TableLabels:          deps.TableLabels,
		ReadEvent:            deps.ReadEvent,
		ListEventAttendees:   deps.ListEventAttendees,
		ListEventResources:   deps.ListEventResources,
		ListEventProducts:    deps.ListEventProducts,
		ListEventOccurrences: deps.ListEventOccurrences,
		ListEventAttachments: deps.ListEventAttachments,
	}
	detailDeps.UploadFile = deps.UploadFile
	detailDeps.ListAttachments = deps.ListAttachments
	detailDeps.CreateAttachment = deps.CreateAttachment
	detailDeps.DeleteAttachment = deps.DeleteAttachment
	detailDeps.NewAttachmentID = deps.NewID

	actionDeps := &eventaction.Deps{
		Routes:                 deps.Routes,
		Labels:                 deps.Labels,
		CommonLabels:           deps.CommonLabels,
		CreateEvent:            deps.CreateEvent,
		ReadEvent:              deps.ReadEvent,
		UpdateEvent:            deps.UpdateEvent,
		DeleteEvent:            deps.DeleteEvent,
		ListEvents:             deps.ListEvents,
		ListEventTags:          deps.ListEventTags,
		ListEventTagsForEvent:  deps.ListEventTagsForEvent,
		SearchAttendees:        deps.SearchAttendees,
		ListAttendeesForEvent:  deps.ListAttendeesForEvent,
		SetEventTagAssignments: deps.SetEventTagAssignments,
		SyncEventAttendees:     deps.SyncEventAttendees,
		ListEventAttachments:   deps.ListEventAttachments,
	}

	return &Module{
		routes: deps.Routes,
		List: eventlist.NewView(&eventlist.ListViewDeps{
			Routes:       deps.Routes,
			ListEvents:   deps.ListEvents,
			Labels:       deps.Labels,
			CommonLabels: deps.CommonLabels,
			TableLabels:  deps.TableLabels,
		}),
		Detail:        eventdetail.NewView(detailDeps),
		TabAction:     eventdetail.NewTabAction(detailDeps),
		Add:           eventaction.NewAddAction(actionDeps),
		Edit:          eventaction.NewEditAction(actionDeps),
		Delete:        eventaction.NewDeleteAction(actionDeps),
		BulkDelete:    eventaction.NewBulkDeleteAction(actionDeps),
		SetStatus:     eventaction.NewSetStatusAction(actionDeps),
		BulkSetStatus: eventaction.NewBulkSetStatusAction(actionDeps),
		Calendar: eventcalendar.NewView(&eventcalendar.ViewDeps{
			Routes:       deps.Routes,
			Labels:       deps.Labels,
			CommonLabels: deps.CommonLabels,
		}),
		Dashboard: eventdashboard.NewView(&eventdashboard.Deps{
			Routes:           deps.Routes,
			EventTagRoutes:   deps.EventTagRoutes,
			RecurrenceRoutes: deps.RecurrenceRoutes,
			Labels:           deps.Labels,
			CommonLabels:     deps.CommonLabels,
			GetDashboardData: deps.GetScheduleDashboardData,
		}),
		AttachmentUpload: eventdetail.NewAttachmentUploadAction(detailDeps),
		AttachmentDelete: eventdetail.NewAttachmentDeleteAction(detailDeps),
	}
}

// RegisterRoutes registers all event routes.
func (m *Module) RegisterRoutes(r view.RouteRegistrar) {
	r.GET(m.routes.ListURL, m.List)
	r.GET(m.routes.DetailURL, m.Detail)
	r.GET(m.routes.TabActionURL, m.TabAction)
	r.GET(m.routes.AddURL, m.Add)
	r.POST(m.routes.AddURL, m.Add)
	r.GET(m.routes.EditURL, m.Edit)
	r.POST(m.routes.EditURL, m.Edit)
	r.POST(m.routes.DeleteURL, m.Delete)
	r.POST(m.routes.BulkDeleteURL, m.BulkDelete)
	r.POST(m.routes.SetStatusURL, m.SetStatus)
	r.POST(m.routes.BulkSetStatusURL, m.BulkSetStatus)
	r.GET(m.routes.CalendarURL, m.Calendar)
	r.GET(m.routes.DashboardURL, m.Dashboard)
	// Attachments
	if m.AttachmentUpload != nil {
		r.GET(m.routes.AttachmentUploadURL, m.AttachmentUpload)
		r.POST(m.routes.AttachmentUploadURL, m.AttachmentUpload)
		r.POST(m.routes.AttachmentDeleteURL, m.AttachmentDelete)
	}
}
