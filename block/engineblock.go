package block

import (
	"context"
	"sort"
	"strings"
	"time"

	eventdashboardview "github.com/erniealice/cyta-golang/domain/event/event/dashboard"
	eventform "github.com/erniealice/cyta-golang/domain/event/event/form"

	"github.com/erniealice/espyna-golang/consumer"
	composehelper "github.com/erniealice/espyna-golang/consumer/compose"
	"github.com/erniealice/espyna-golang/reference"
	commonpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/common"
	attachmentpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/document/attachment"
	eventpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/event/event"
	eventattendeepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/event/event_attendee"
	eventtagpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/event/event_tag"
	eventtagassignmentpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/event/event_tag_assignment"
	scheduledashpb "github.com/erniealice/esqyma/pkg/schema/v1/service/dashboard/schedule"
	"github.com/erniealice/pyeza-golang"
)

// cytaEngineBlock returns a pyeza.AppOption that registers the cyta event
// domain modules via the compose engine (replaces legacy cytaBlock).
func EngineBlock() pyeza.AppOption {
	return func(ctx *pyeza.AppContext) error {
		uc, err := composehelper.RequireUseCases(ctx, "cytaEngineBlock")
		if err != nil {
			return err
		}
		adapted := buildCytaUseCases(uc)

		infra := &Infra{}
		infra.UploadFile, _ = ctx.UploadFile.(func(context.Context, string, string, []byte, string) error)
		infra.ListAttachments, _ = ctx.ListAttachments.(func(context.Context, string, string) (*attachmentpb.ListAttachmentsResponse, error))
		infra.CreateAttachment, _ = ctx.CreateAttachment.(func(context.Context, *attachmentpb.CreateAttachmentRequest) (*attachmentpb.CreateAttachmentResponse, error))
		infra.DeleteAttachment, _ = ctx.DeleteAttachment.(func(context.Context, *attachmentpb.DeleteAttachmentRequest) (*attachmentpb.DeleteAttachmentResponse, error))
		infra.NewAttachmentID, _ = ctx.NewAttachmentID.(func() string)
		if ctx.RefChecker != nil {
			if rc, ok := ctx.RefChecker.(reference.Checker); ok {
				infra.RefChecker = rc
			}
		}

		units := AllUnits(adapted, infra)
		return composehelper.AssembleEngineBlock("cyta", units, ctx)
	}
}

// ---------------------------------------------------------------------------
// cyta adapter
// ---------------------------------------------------------------------------

// buildCytaUseCases maps espyna's *consumer.UseCases to cyta block's typed
// shape (packages/cyta-golang/block/usecases.go). All sub-group wiring is
// nil-safe.
//
// Phase 2 Round 2 (Q-WIRE-1): besides the proto-typed leaf CRUD, this builds
// the derived join/picker closures in their final VIEW shape (the prior cyta
// reflection in wiring.go is gone). The tag/attendee joins compose multiple
// espyna use cases and would otherwise reference espyna-internal request types
// (SetEventTagAssignmentsRequest) — here they are built reflection-free from
// the public proto-typed Create/Delete/ListByEvent closures (diff-and-replace),
// and the schedule dashboard is translated proto→view.
func buildCytaUseCases(uc *consumer.UseCases) *UseCases {
	result := &UseCases{
		GetWorkspaceIDFromCtx: consumer.GetWorkspaceIDFromContext,
	}

	if uc.Event == nil {
		return result
	}
	ev := uc.Event

	// -- Event leaf CRUD ---------------------------------------------------------
	if ev.Event != nil {
		result.Event.Create = ev.Event.CreateEvent.Execute
		result.Event.Read = ev.Event.ReadEvent.Execute
		result.Event.Update = ev.Event.UpdateEvent.Execute
		result.Event.Delete = ev.Event.DeleteEvent.Execute
		result.Event.List = ev.Event.ListEvents.Execute
	}

	// -- Event nested-entity lists (detail tabs; optional) -----------------------
	if ev.EventAttendee != nil {
		result.Event.ListAttendees = ev.EventAttendee.ListEventAttendees.Execute
	}
	if ev.EventResource != nil {
		result.Event.ListResources = ev.EventResource.ListEventResources.Execute
	}
	if ev.EventProduct != nil {
		result.Event.ListProducts = ev.EventProduct.ListEventProducts.Execute
	}
	if ev.EventOccurrence != nil {
		result.Event.ListOccurrences = ev.EventOccurrence.ListEventOccurrences.Execute
	}

	// -- EventTag master CRUD ----------------------------------------------------
	if ev.EventTag != nil {
		result.EventTag.Create = ev.EventTag.CreateEventTag.Execute
		result.EventTag.Read = ev.EventTag.ReadEventTag.Execute
		result.EventTag.Update = ev.EventTag.UpdateEventTag.Execute
		result.EventTag.Delete = ev.EventTag.DeleteEventTag.Execute
		result.EventTag.List = ev.EventTag.ListEventTags.Execute
		result.EventTag.GetListPageData = ev.EventTag.GetEventTagListPageData.Execute

		// ListTagOptions — workspace-scoped master tag list for the multi-picker.
		listTags := ev.EventTag.ListEventTags
		result.Event.ListTagOptions = func(ctx context.Context) ([]eventform.Option, error) {
			resp, err := listTags.Execute(ctx, &eventtagpb.ListEventTagsRequest{})
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

	// -- ListTagsForEvent + SetEventTagAssignments — per-event tag join ----------
	//
	// ListEventTagAssignmentsByEvent returns the full assignment rows; the picker
	// wants the ordered tag IDs. SetEventTagAssignments is composed reflection-free
	// from the public proto Create/Delete closures (diff-and-replace) — the espyna
	// SetEventTagAssignments use case takes an internal Go request type cyta /
	// service-admin cannot import, so we replicate its delete-then-create semantics
	// here over the proto-typed leaf closures.
	if ev.EventTagAssignment != nil {
		asg := ev.EventTagAssignment

		if asg.ListEventTagAssignmentsByEvent != nil {
			byEvent := asg.ListEventTagAssignmentsByEvent
			result.Event.ListTagsForEvent = func(ctx context.Context, eventID string) ([]string, error) {
				resp, err := byEvent.Execute(ctx, eventID)
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

		// SetEventTagAssignments — diff-and-replace via proto Create/Delete +
		// ListByEvent. Requires ReadEvent for the workspace_id of new rows.
		if asg.CreateEventTagAssignment != nil && asg.DeleteEventTagAssignment != nil &&
			asg.ListEventTagAssignmentsByEvent != nil && ev.Event != nil {
			createAsg := asg.CreateEventTagAssignment
			deleteAsg := asg.DeleteEventTagAssignment
			byEvent := asg.ListEventTagAssignmentsByEvent
			readEvent := ev.Event.ReadEvent
			result.Event.SetEventTagAssignments = func(ctx context.Context, eventID string, tagIDs []string) error {
				// Desired set (deduped, trimmed).
				desired := make(map[string]struct{}, len(tagIDs))
				order := make([]string, 0, len(tagIDs))
				for _, id := range tagIDs {
					id = strings.TrimSpace(id)
					if id == "" {
						continue
					}
					if _, seen := desired[id]; seen {
						continue
					}
					desired[id] = struct{}{}
					order = append(order, id)
				}

				// Resolve workspace_id by reading the event.
				workspaceID := ""
				readResp, err := readEvent.Execute(ctx, &eventpb.ReadEventRequest{
					Data: &eventpb.Event{Id: eventID},
				})
				if err != nil {
					return err
				}
				if readResp != nil && len(readResp.GetData()) > 0 {
					workspaceID = readResp.GetData()[0].GetWorkspaceId()
				}

				// Current assignments for diffing.
				curResp, err := byEvent.Execute(ctx, eventID)
				if err != nil {
					return err
				}
				existing := make(map[string]*eventtagassignmentpb.EventTagAssignment)
				for _, r := range curResp.GetData() {
					if r == nil || !r.GetActive() {
						continue
					}
					if id := r.GetEventTagId(); id != "" {
						existing[id] = r
					}
				}

				// Delete rows no longer desired.
				for tagID, row := range existing {
					if _, keep := desired[tagID]; keep {
						continue
					}
					if _, derr := deleteAsg.Execute(ctx, &eventtagassignmentpb.DeleteEventTagAssignmentRequest{
						Data: &eventtagassignmentpb.EventTagAssignment{Id: row.GetId()},
					}); derr != nil {
						return derr
					}
				}

				// Create rows newly desired (preserving caller order as position).
				for pos, tagID := range order {
					if _, have := existing[tagID]; have {
						continue
					}
					if _, cerr := createAsg.Execute(ctx, &eventtagassignmentpb.CreateEventTagAssignmentRequest{
						Data: &eventtagassignmentpb.EventTagAssignment{
							EventId:     eventID,
							EventTagId:  tagID,
							WorkspaceId: workspaceID,
							Position:    int32(pos),
							Active:      true,
						},
					}); cerr != nil {
						return cerr
					}
				}
				return nil
			}
		}
	}

	// -- ListAttendeesForEvent + SyncEventAttendees — attendee join --------------
	if ev.EventAttendee != nil {
		att := ev.EventAttendee
		listAttendees := att.ListEventAttendees

		// Filter-by-event_id request builder.
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
		// "client:<id>" ref scheme; "" when neither FK is populated.
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

		if listAttendees != nil {
			result.Event.ListAttendeesForEvent = func(ctx context.Context, eventID string) ([]eventform.SelectedOption, error) {
				resp, err := listAttendees.Execute(ctx, buildListReq(eventID))
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
		if listAttendees != nil && att.CreateEventAttendee != nil && att.DeleteEventAttendee != nil && ev.Event != nil {
			createAtt := att.CreateEventAttendee
			deleteAtt := att.DeleteEventAttendee
			readEvent := ev.Event.ReadEvent
			result.Event.SyncEventAttendees = func(ctx context.Context, eventID string, attendeeRefs []string) error {
				desired := make(map[string]struct{}, len(attendeeRefs))
				for _, r := range attendeeRefs {
					r = strings.TrimSpace(r)
					if r == "" {
						continue
					}
					desired[r] = struct{}{}
				}

				// Resolve workspace_id (needed for new attendee rows).
				workspaceID := ""
				if readEvent != nil {
					readResp, err := readEvent.Execute(ctx, &eventpb.ReadEventRequest{
						Data: &eventpb.Event{Id: eventID},
					})
					if err != nil {
						return err
					}
					if readResp != nil && len(readResp.GetData()) > 0 {
						workspaceID = readResp.GetData()[0].GetWorkspaceId()
					}
				}

				curResp, err := listAttendees.Execute(ctx, buildListReq(eventID))
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

				// Delete rows no longer desired.
				for ref, row := range existing {
					if _, keep := desired[ref]; keep {
						continue
					}
					if _, derr := deleteAtt.Execute(ctx, &eventattendeepb.DeleteEventAttendeeRequest{
						Data: &eventattendeepb.EventAttendee{Id: row.GetId()},
					}); derr != nil {
						return derr
					}
				}

				// Create rows newly desired.
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
					if _, cerr := createAtt.Execute(ctx, &eventattendeepb.CreateEventAttendeeRequest{Data: row}); cerr != nil {
						return cerr
					}
				}
				return nil
			}
		}
	}

	// SearchAttendees: not wired — workspace_user search backing TBD; the
	// drawer multi-picker degrades to pre-selected-only when nil (matches prior).

	// -- Service.Dashboard.Schedule — proto→view translation ---------------------
	//
	// Proto Response carries *ScheduleStats (Today/ThisWeek/ByTag/UtilizationPct),
	// ByDay{Labels,Values}, Upcoming ([]*eventpb.Event), and ByTag []*ScheduleTagSlice.
	// The view-layer Response wants ByTag as map[string]int64 and the stat fields
	// flat. Same empty-workspace fallback as the fayna dashboards.
	if uc.Service != nil && uc.Service.Dashboard != nil &&
		uc.Service.Dashboard.Schedule != nil && uc.Service.Dashboard.Schedule.GetScheduleDashboard != nil {
		schedDash := uc.Service.Dashboard.Schedule.GetScheduleDashboard
		result.GetScheduleDashboardData = func(ctx context.Context, req *eventdashboardview.Request) (*eventdashboardview.Response, error) {
			workspaceID := ""
			if req != nil {
				workspaceID = req.WorkspaceID
			}
			if workspaceID == "" {
				workspaceID = consumer.GetWorkspaceIDFromContext(ctx)
			}
			nowMillis := time.Now().UnixMilli()
			resp, err := schedDash.Execute(ctx, &scheduledashpb.GetScheduleDashboardRequest{
				WorkspaceId: workspaceID,
				NowMillis:   &nowMillis,
			})
			if err != nil {
				return nil, err
			}
			if resp == nil {
				return nil, nil
			}
			byTag := make(map[string]int64)
			for _, s := range resp.GetByTag() {
				if s == nil {
					continue
				}
				byTag[s.GetTag()] = s.GetCount()
			}
			return &eventdashboardview.Response{
				Today:          resp.GetStats().GetToday(),
				ThisWeek:       resp.GetStats().GetThisWeek(),
				ByTagCount:     resp.GetStats().GetByTag(),
				UtilizationPct: resp.GetStats().GetUtilizationPct(),
				ByDayLabels:    resp.GetByDayLabels(),
				ByDayValues:    resp.GetByDayValues(),
				ByTag:          byTag,
				Upcoming:       resp.GetUpcoming(),
			}, nil
		}
	}

	return result
}
