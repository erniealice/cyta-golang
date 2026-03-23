package block

// wiring.go provides reflection-based use-case wiring helpers for Block().
//
// The cyta/block sub-package does not import espyna-golang to avoid dependency
// cycles. Instead, reflect is used to navigate the opaque *usecases.Aggregate
// and extract .Execute methods for the event use cases.
//
// Struct field path (mirrors espyna's usecases.Aggregate):
//
//	UseCases (Aggregate)
//	  └─ Event (*EventUseCases)
//	       ├─ Event (*event.UseCases)
//	       │    ├─ CreateEvent (*CreateEventUseCase) → .Execute(ctx, req) (resp, error)
//	       │    ├─ ListEvents  (*ListEventsUseCase)  → .Execute(ctx, req) (resp, error)
//	       │    └─ ...
//	       ├─ EventAttendee, EventOccurrence, EventProduct, EventResource
//	       └─ ...

import (
	"context"
	"reflect"

	eventpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/event/event"
	eventattendeepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/event/event_attendee"
	eventoccurrencepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/event/event_occurrence"
	eventproductpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/event/event_product"
	eventresourcepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/event/event_resource"

	eventmod "github.com/erniealice/cyta-golang/views/event"
)

// ---------------------------------------------------------------------------
// Reflection helpers
// ---------------------------------------------------------------------------

// ucAggregate wraps the opaque ctx.UseCases value for safe field navigation.
type ucAggregate struct {
	v reflect.Value // dereferenced *usecases.Aggregate struct
}

// assertUseCases wraps ctx.UseCases in a reflection accessor.
// Returns nil if ctx.UseCases is nil or not a pointer-to-struct.
func assertUseCases(raw any) *ucAggregate {
	if raw == nil {
		return nil
	}
	v := reflect.ValueOf(raw)
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return nil
		}
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return nil
	}
	return &ucAggregate{v: v}
}

// ptrField safely dereferences a pointer-typed struct field by name.
// Returns zero Value if the field is not found or is nil.
func ptrField(v reflect.Value, name string) reflect.Value {
	if !v.IsValid() {
		return reflect.Value{}
	}
	f := v.FieldByName(name)
	if !f.IsValid() {
		return reflect.Value{}
	}
	if f.Kind() == reflect.Ptr {
		if f.IsNil() {
			return reflect.Value{}
		}
		return f.Elem()
	}
	return f
}

// execFn extracts the Execute method from a pointer-typed use-case leaf field
// and returns it as interface{} for type-assertion.
func execFn(parent reflect.Value, fieldName string) any {
	if !parent.IsValid() {
		return nil
	}
	f := parent.FieldByName(fieldName)
	if !f.IsValid() {
		return nil
	}
	if f.Kind() == reflect.Ptr {
		if f.IsNil() {
			return nil
		}
		m := f.MethodByName("Execute")
		if m.IsValid() {
			return m.Interface()
		}
		m = f.Elem().MethodByName("Execute")
		if m.IsValid() {
			return m.Interface()
		}
		return nil
	}
	m := f.MethodByName("Execute")
	if !m.IsValid() {
		return nil
	}
	return m.Interface()
}

// ---------------------------------------------------------------------------
// Event module wiring
// ---------------------------------------------------------------------------

// wireEventDeps overlays real use-case functions onto deps where available.
// Stub functions already set in block.go are kept if the real ones are not found.
func wireEventDeps(deps *eventmod.ModuleDeps, uc *ucAggregate) {
	// usecases.Aggregate.Event is *event.EventUseCases
	ev := ptrField(uc.v, "Event")
	if !ev.IsValid() {
		return
	}

	// ev is the dereferenced EventUseCases struct.
	// Its "Event" field is *event.UseCases (the leaf use-case group).
	evLeaf := ptrField(ev, "Event")
	if evLeaf.IsValid() {
		if fn, ok := execFn(evLeaf, "CreateEvent").(func(context.Context, *eventpb.CreateEventRequest) (*eventpb.CreateEventResponse, error)); ok {
			deps.CreateEvent = fn
		}
		if fn, ok := execFn(evLeaf, "ReadEvent").(func(context.Context, *eventpb.ReadEventRequest) (*eventpb.ReadEventResponse, error)); ok {
			deps.ReadEvent = fn
		}
		if fn, ok := execFn(evLeaf, "UpdateEvent").(func(context.Context, *eventpb.UpdateEventRequest) (*eventpb.UpdateEventResponse, error)); ok {
			deps.UpdateEvent = fn
		}
		if fn, ok := execFn(evLeaf, "DeleteEvent").(func(context.Context, *eventpb.DeleteEventRequest) (*eventpb.DeleteEventResponse, error)); ok {
			deps.DeleteEvent = fn
		}
		if fn, ok := execFn(evLeaf, "ListEvents").(func(context.Context, *eventpb.ListEventsRequest) (*eventpb.ListEventsResponse, error)); ok {
			deps.ListEvents = fn
		}
	}

	// Attendee operations
	attendee := ptrField(ev, "EventAttendee")
	if attendee.IsValid() {
		if fn, ok := execFn(attendee, "ListEventAttendees").(func(context.Context, *eventattendeepb.ListEventAttendeesRequest) (*eventattendeepb.ListEventAttendeesResponse, error)); ok {
			deps.ListEventAttendees = fn
		}
	}

	// Resource operations
	resource := ptrField(ev, "EventResource")
	if resource.IsValid() {
		if fn, ok := execFn(resource, "ListEventResources").(func(context.Context, *eventresourcepb.ListEventResourcesRequest) (*eventresourcepb.ListEventResourcesResponse, error)); ok {
			deps.ListEventResources = fn
		}
	}

	// Product operations
	product := ptrField(ev, "EventProduct")
	if product.IsValid() {
		if fn, ok := execFn(product, "ListEventProducts").(func(context.Context, *eventproductpb.ListEventProductsRequest) (*eventproductpb.ListEventProductsResponse, error)); ok {
			deps.ListEventProducts = fn
		}
	}

	// Occurrence operations
	occurrence := ptrField(ev, "EventOccurrence")
	if occurrence.IsValid() {
		if fn, ok := execFn(occurrence, "ListEventOccurrences").(func(context.Context, *eventoccurrencepb.ListEventOccurrencesRequest) (*eventoccurrencepb.ListEventOccurrencesResponse, error)); ok {
			deps.ListEventOccurrences = fn
		}
	}
}
