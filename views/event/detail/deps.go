package detail

import (
	"context"

	cyta "github.com/erniealice/cyta-golang"
	eventform "github.com/erniealice/cyta-golang/views/event/form"

	"github.com/erniealice/hybra-golang/views/attachment"
	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/pyeza-golang/types"

	eventpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/event/event"
	eventattendeepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/event/event_attendee"
	eventoccurrencepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/event/event_occurrence"
	eventproductpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/event/event_product"
	eventresourcepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/event/event_resource"
)

// DetailViewDeps holds view dependencies for the event detail views.
type DetailViewDeps struct {
	Routes       cyta.EventRoutes
	Labels       cyta.EventLabels
	CommonLabels pyeza.CommonLabels
	TableLabels  types.TableLabels

	// Event read
	ReadEvent func(ctx context.Context, req *eventpb.ReadEventRequest) (*eventpb.ReadEventResponse, error)

	// Sub-entity list operations (for tabs)
	ListEventAttendees   func(ctx context.Context, req *eventattendeepb.ListEventAttendeesRequest) (*eventattendeepb.ListEventAttendeesResponse, error)
	ListEventResources   func(ctx context.Context, req *eventresourcepb.ListEventResourcesRequest) (*eventresourcepb.ListEventResourcesResponse, error)
	ListEventProducts    func(ctx context.Context, req *eventproductpb.ListEventProductsRequest) (*eventproductpb.ListEventProductsResponse, error)
	ListEventOccurrences func(ctx context.Context, req *eventoccurrencepb.ListEventOccurrencesRequest) (*eventoccurrencepb.ListEventOccurrencesResponse, error)

	// Phase 5 — attachments tab (legacy flat-list path, kept for action.Deps compatibility).
	// When the hybra attachment ops below are wired, the detail page renders the
	// standard hybra attachment table instead.
	ListEventAttachments func(ctx context.Context, eventID string) ([]eventform.Attachment, error)

	// Hybra attachment ops (upload/delete/list via hybra attachment.BuildTable).
	attachment.AttachmentOps
}
