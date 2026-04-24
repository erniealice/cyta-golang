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
	"sort"
	"strings"

	commonpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/common"
	eventpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/event/event"
	eventattendeepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/event/event_attendee"
	eventoccurrencepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/event/event_occurrence"
	eventproductpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/event/event_product"
	eventresourcepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/event/event_resource"
	eventtagpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/event/event_tag"
	eventtagassignmentpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/event/event_tag_assignment"

	eventmod "github.com/erniealice/cyta-golang/views/event"
	eventform "github.com/erniealice/cyta-golang/views/event/form"
	eventtagmod "github.com/erniealice/cyta-golang/views/event_tag"
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

	// -----------------------------------------------------------------
	// Phase 4 — drawer picker wiring (tags + attendees)
	// -----------------------------------------------------------------

	// ListEventTags — workspace-scoped master tag list for the multi-picker.
	tag := ptrField(ev, "EventTag")
	if tag.IsValid() {
		if listTagsFn, ok := execFn(tag, "ListEventTags").(func(context.Context, *eventtagpb.ListEventTagsRequest) (*eventtagpb.ListEventTagsResponse, error)); ok {
			deps.ListEventTags = func(ctx context.Context) ([]eventform.Option, error) {
				resp, err := listTagsFn(ctx, &eventtagpb.ListEventTagsRequest{})
				if err != nil {
					return nil, err
				}
				out := make([]eventform.Option, 0, len(resp.GetData()))
				for _, t := range resp.GetData() {
					if t == nil || !t.GetActive() {
						continue
					}
					out = append(out, eventform.Option{Value: t.GetId(), Label: t.GetName()})
				}
				return out, nil
			}
		}
	}

	// ListEventTagsForEvent + SetEventTagAssignments — per-event tag join.
	asg := ptrField(ev, "EventTagAssignment")
	if asg.IsValid() {
		// ListEventTagAssignmentsByEvent returns the full response struct, not []string.
		// Wrap it: extract event_tag_ids sorted by position.
		byEventFn, _ := execFn(asg, "ListEventTagAssignmentsByEvent").(func(context.Context, string) (*eventtagassignmentpb.ListEventTagAssignmentsResponse, error))
		if byEventFn != nil {
			deps.ListEventTagsForEvent = func(ctx context.Context, eventID string) ([]string, error) {
				resp, err := byEventFn(ctx, eventID)
				if err != nil {
					return nil, err
				}
				rows := resp.GetData()
				active := make([]*eventtagassignmentpb.EventTagAssignment, 0, len(rows))
				for _, r := range rows {
					if r == nil || !r.GetActive() {
						continue
					}
					active = append(active, r)
				}
				sort.SliceStable(active, func(i, j int) bool {
					return active[i].GetPosition() < active[j].GetPosition()
				})
				out := make([]string, 0, len(active))
				for _, r := range active {
					if id := r.GetEventTagId(); id != "" {
						out = append(out, id)
					}
				}
				return out, nil
			}
		}

		// SetEventTagAssignments — espyna use case takes a local struct
		// (*SetEventTagAssignmentsRequest with EventID / WorkspaceID / TagIDs),
		// not a proto. We go via reflection so we don't import espyna-golang.
		// Workspace ID is resolved by reading the event first — the block layer
		// does not have access to espyna's session-context keys.
		setField := asg.FieldByName("SetEventTagAssignments")
		if setField.IsValid() && setField.Kind() == reflect.Ptr && !setField.IsNil() {
			setExec := setField.MethodByName("Execute")
			readEventFn := deps.ReadEvent
			if setExec.IsValid() && setExec.Type().NumIn() == 2 && readEventFn != nil {
				reqType := setExec.Type().In(1)
				if reqType.Kind() == reflect.Ptr && reqType.Elem().Kind() == reflect.Struct {
					deps.SetEventTagAssignments = func(ctx context.Context, eventID string, tagIDs []string) error {
						// Resolve workspace_id by reading the event.
						readResp, err := readEventFn(ctx, &eventpb.ReadEventRequest{
							Data: &eventpb.Event{Id: eventID},
						})
						if err != nil {
							return err
						}
						workspaceID := ""
						if readResp != nil && len(readResp.GetData()) > 0 {
							workspaceID = readResp.GetData()[0].GetWorkspaceId()
						}

						// Build *SetEventTagAssignmentsRequest via reflection.
						reqPtr := reflect.New(reqType.Elem())
						req := reqPtr.Elem()
						if f := req.FieldByName("EventID"); f.IsValid() && f.CanSet() {
							f.SetString(eventID)
						}
						if f := req.FieldByName("WorkspaceID"); f.IsValid() && f.CanSet() {
							f.SetString(workspaceID)
						}
						if f := req.FieldByName("TagIDs"); f.IsValid() && f.CanSet() && f.Type().Kind() == reflect.Slice {
							sliced := reflect.MakeSlice(f.Type(), 0, len(tagIDs))
							for _, id := range tagIDs {
								id = strings.TrimSpace(id)
								if id == "" {
									continue
								}
								sliced = reflect.Append(sliced, reflect.ValueOf(id))
							}
							f.Set(sliced)
						}

						results := setExec.Call([]reflect.Value{reflect.ValueOf(ctx), reqPtr})
						if len(results) == 2 && !results[1].IsNil() {
							if errVal, ok := results[1].Interface().(error); ok {
								return errVal
							}
						}
						return nil
					}
				}
			}
		}
	}

	// ListAttendeesForEvent + SyncEventAttendees — attendee join.
	if attendee.IsValid() {
		listAttendeesFn := deps.ListEventAttendees
		createField := attendee.FieldByName("CreateEventAttendee")
		deleteField := attendee.FieldByName("DeleteEventAttendee")

		createFn, _ := execFn(attendee, "CreateEventAttendee").(func(context.Context, *eventattendeepb.CreateEventAttendeeRequest) (*eventattendeepb.CreateEventAttendeeResponse, error))
		deleteFn, _ := execFn(attendee, "DeleteEventAttendee").(func(context.Context, *eventattendeepb.DeleteEventAttendeeRequest) (*eventattendeepb.DeleteEventAttendeeResponse, error))

		// Helper: filter-by-event_id request for ListEventAttendees.
		buildListReq := func(eventID string) *eventattendeepb.ListEventAttendeesRequest {
			return &eventattendeepb.ListEventAttendeesRequest{
				Filters: &commonpb.FilterRequest{
					Filters: []*commonpb.TypedFilter{
						{
							Field: "event_id",
							FilterType: &commonpb.TypedFilter_StringFilter{
								StringFilter: &commonpb.StringFilter{
									Value:    eventID,
									Operator: commonpb.StringOperator_STRING_EQUALS,
								},
							},
						},
					},
				},
			}
		}

		// attendeeRef formats an attendee row into the picker's "user:<id>" /
		// "client:<id>" ref scheme; returns "" when neither FK is populated.
		attendeeRef := func(a *eventattendeepb.EventAttendee) string {
			if a == nil {
				return ""
			}
			if id := a.GetWorkspaceUserId(); id != "" {
				return "user:" + id
			}
			if id := a.GetClientId(); id != "" {
				return "client:" + id
			}
			return ""
		}

		if listAttendeesFn != nil {
			deps.ListAttendeesForEvent = func(ctx context.Context, eventID string) ([]eventform.SelectedOption, error) {
				resp, err := listAttendeesFn(ctx, buildListReq(eventID))
				if err != nil {
					return nil, err
				}
				out := make([]eventform.SelectedOption, 0, len(resp.GetData()))
				for _, a := range resp.GetData() {
					if a == nil || !a.GetActive() {
						continue
					}
					ref := attendeeRef(a)
					if ref == "" {
						continue
					}
					label := a.GetDisplayName()
					if label == "" {
						// Fallback placeholder — workspace_user / client lookup
						// backing is deferred; see SearchAttendees note below.
						if strings.HasPrefix(ref, "user:") {
							label = "User " + a.GetWorkspaceUserId()
						} else {
							label = "Client " + a.GetClientId()
						}
					}
					out = append(out, eventform.SelectedOption{Value: ref, Label: label})
				}
				return out, nil
			}
		}

		// SyncEventAttendees — diff-and-replace using current list + Create/Delete.
		if listAttendeesFn != nil && createField.IsValid() && !createField.IsNil() && deleteField.IsValid() && !deleteField.IsNil() && createFn != nil && deleteFn != nil {
			deps.SyncEventAttendees = func(ctx context.Context, eventID string, attendeeRefs []string) error {
				// Normalize desired set.
				desired := make(map[string]struct{}, len(attendeeRefs))
				for _, r := range attendeeRefs {
					r = strings.TrimSpace(r)
					if r == "" {
						continue
					}
					desired[r] = struct{}{}
				}

				// Resolve workspace_id (needed for new attendee rows).
				var workspaceID string
				if deps.ReadEvent != nil {
					readResp, err := deps.ReadEvent(ctx, &eventpb.ReadEventRequest{
						Data: &eventpb.Event{Id: eventID},
					})
					if err != nil {
						return err
					}
					if readResp != nil && len(readResp.GetData()) > 0 {
						workspaceID = readResp.GetData()[0].GetWorkspaceId()
					}
				}

				// Fetch current attendees for diffing.
				curResp, err := listAttendeesFn(ctx, buildListReq(eventID))
				if err != nil {
					return err
				}
				existing := make(map[string]*eventattendeepb.EventAttendee, len(curResp.GetData()))
				for _, a := range curResp.GetData() {
					if a == nil || !a.GetActive() {
						continue
					}
					if ref := attendeeRef(a); ref != "" {
						existing[ref] = a
					}
				}

				// Delete rows present in existing but missing from desired.
				for ref, row := range existing {
					if _, keep := desired[ref]; keep {
						continue
					}
					if _, derr := deleteFn(ctx, &eventattendeepb.DeleteEventAttendeeRequest{
						Data: &eventattendeepb.EventAttendee{Id: row.GetId()},
					}); derr != nil {
						return derr
					}
				}

				// Create rows present in desired but missing from existing.
				for ref := range desired {
					if _, have := existing[ref]; have {
						continue
					}
					parts := strings.SplitN(ref, ":", 2)
					if len(parts) != 2 || parts[1] == "" {
						continue
					}
					row := &eventattendeepb.EventAttendee{
						EventId:     eventID,
						WorkspaceId: workspaceID,
						Active:      true,
					}
					switch parts[0] {
					case "user":
						id := parts[1]
						row.WorkspaceUserId = &id
					case "client":
						id := parts[1]
						row.ClientId = &id
					default:
						continue
					}
					if _, cerr := createFn(ctx, &eventattendeepb.CreateEventAttendeeRequest{Data: row}); cerr != nil {
						return cerr
					}
				}
				return nil
			}
		}
	}

	// SearchAttendees: not wired — workspace_user search backing TBD; the
	// drawer's multi-picker degrades to pre-selected-only when this is nil.
	//
	// ListEventAttachments: wired in Phase 5 when the hybra attachment use
	// case is bridged into the cyta block.
}

// ---------------------------------------------------------------------------
// EventTag module wiring
// ---------------------------------------------------------------------------

// wireEventTagDeps overlays real use-case functions onto deps where available.
// Field path:  Aggregate.Event (*EventUseCases).EventTag (*eventtag.UseCases)
func wireEventTagDeps(deps *eventtagmod.ModuleDeps, uc *ucAggregate) {
	ev := ptrField(uc.v, "Event")
	if !ev.IsValid() {
		return
	}
	tag := ptrField(ev, "EventTag")
	if !tag.IsValid() {
		return
	}

	if fn, ok := execFn(tag, "CreateEventTag").(func(context.Context, *eventtagpb.CreateEventTagRequest) (*eventtagpb.CreateEventTagResponse, error)); ok {
		deps.CreateEventTag = fn
	}
	if fn, ok := execFn(tag, "ReadEventTag").(func(context.Context, *eventtagpb.ReadEventTagRequest) (*eventtagpb.ReadEventTagResponse, error)); ok {
		deps.ReadEventTag = fn
	}
	if fn, ok := execFn(tag, "UpdateEventTag").(func(context.Context, *eventtagpb.UpdateEventTagRequest) (*eventtagpb.UpdateEventTagResponse, error)); ok {
		deps.UpdateEventTag = fn
	}
	if fn, ok := execFn(tag, "DeleteEventTag").(func(context.Context, *eventtagpb.DeleteEventTagRequest) (*eventtagpb.DeleteEventTagResponse, error)); ok {
		deps.DeleteEventTag = fn
	}
	if fn, ok := execFn(tag, "ListEventTags").(func(context.Context, *eventtagpb.ListEventTagsRequest) (*eventtagpb.ListEventTagsResponse, error)); ok {
		deps.ListEventTags = fn
	}
	if fn, ok := execFn(tag, "GetEventTagListPageData").(func(context.Context, *eventtagpb.GetEventTagListPageDataRequest) (*eventtagpb.GetEventTagListPageDataResponse, error)); ok {
		deps.GetEventTagListPageData = fn
	}
}
