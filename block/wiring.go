package block

// wiring.go maps the typed *UseCases contract (usecases.go) onto the cyta view
// ModuleDeps structs.
//
// Previously this file navigated the opaque *usecases.Aggregate via reflection
// (ucAggregate / assertUseCases / ptrField / execFn) and translated proto types
// by field-name. That is gone: service-admin's composition adapter now hands
// cyta a fully-typed *UseCases, so wiring is plain struct-field assignment and
// every espyna/cyta drift is a compile error instead of a silent nil.
//
// These helpers only ASSIGN already-built closures; they never construct them.
// Construction (including the proto→view dashboard translation and the
// tag/attendee join closures that touch espyna-internal request types) lives in
// service-admin's adapter, the only place that knows both vocabularies.

import (
	event "github.com/erniealice/cyta-golang/domain/event"
)

// ---------------------------------------------------------------------------
// Event module wiring
// ---------------------------------------------------------------------------

// wireEventDeps overlays the typed Event use cases onto deps. Nil closures are
// left as-is so block.go's stub fallbacks (for the not-yet-wired scheduling
// backend) survive — matching the prior overlay-if-present behaviour.
func wireEventDeps(deps *event.ModuleDeps, uc *UseCases) {
	if uc == nil {
		return
	}
	ev := uc.Event

	// Leaf CRUD.
	if ev.Create != nil {
		deps.CreateEvent = ev.Create
	}
	if ev.Read != nil {
		deps.ReadEvent = ev.Read
	}
	if ev.Update != nil {
		deps.UpdateEvent = ev.Update
	}
	if ev.Delete != nil {
		deps.DeleteEvent = ev.Delete
	}
	if ev.List != nil {
		deps.ListEvents = ev.List
	}

	// Nested-entity lists (detail tabs).
	if ev.ListAttendees != nil {
		deps.ListEventAttendees = ev.ListAttendees
	}
	if ev.ListResources != nil {
		deps.ListEventResources = ev.ListResources
	}
	if ev.ListProducts != nil {
		deps.ListEventProducts = ev.ListProducts
	}
	if ev.ListOccurrences != nil {
		deps.ListEventOccurrences = ev.ListOccurrences
	}

	// Derived drawer-picker closures (already in view shape; built by the
	// service-admin adapter). Nil → the picker degrades to pre-selected-only.
	if ev.ListTagOptions != nil {
		deps.ListEventTags = ev.ListTagOptions
	}
	if ev.ListTagsForEvent != nil {
		deps.ListEventTagsForEvent = ev.ListTagsForEvent
	}
	if ev.SearchAttendees != nil {
		deps.SearchAttendees = ev.SearchAttendees
	}
	if ev.ListAttendeesForEvent != nil {
		deps.ListAttendeesForEvent = ev.ListAttendeesForEvent
	}
	if ev.SetEventTagAssignments != nil {
		deps.SetEventTagAssignments = ev.SetEventTagAssignments
	}
	if ev.SyncEventAttendees != nil {
		deps.SyncEventAttendees = ev.SyncEventAttendees
	}
}

// wireScheduleDashboard sets deps.GetScheduleDashboardData from the typed slot.
// The proto→view translation that this function used to perform via reflection
// now lives in service-admin's adapter, which supplies the closure ready-made.
// Nil → the dashboard view renders empty stats (its own nil-safe fallback).
func wireScheduleDashboard(deps *event.ModuleDeps, uc *UseCases) {
	if uc == nil || uc.GetScheduleDashboardData == nil {
		return
	}
	deps.GetScheduleDashboardData = uc.GetScheduleDashboardData
}

// ---------------------------------------------------------------------------
// EventTag module wiring
// ---------------------------------------------------------------------------

// wireEventTagDeps overlays the typed EventTag use cases onto deps.
func wireEventTagDeps(deps *event.EventTagModuleDeps, uc *UseCases) {
	if uc == nil {
		return
	}
	tag := uc.EventTag
	if tag.Create != nil {
		deps.CreateEventTag = tag.Create
	}
	if tag.Read != nil {
		deps.ReadEventTag = tag.Read
	}
	if tag.Update != nil {
		deps.UpdateEventTag = tag.Update
	}
	if tag.Delete != nil {
		deps.DeleteEventTag = tag.Delete
	}
	if tag.List != nil {
		deps.ListEventTags = tag.List
	}
	if tag.GetListPageData != nil {
		deps.GetEventTagListPageData = tag.GetListPageData
	}
}
