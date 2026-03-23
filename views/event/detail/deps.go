package detail

import (
	"context"

	cyta "github.com/erniealice/cyta-golang"

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
}
