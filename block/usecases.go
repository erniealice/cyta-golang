// Package block — typed wiring contract for cyta.Block.
//
// This file declares what cyta's Block() needs from outside. Service-admin's
// composition layer constructs a *UseCases value from espyna's consumer
// container; cyta's Block() consumes only this typed shape.
//
// Shape this struct by what CYTA needs, NOT by mirroring espyna's
// *usecases.Aggregate. Service-admin's adapter is the only place that knows
// both vocabularies. If espyna restructures its container, only that adapter
// changes — cyta sees a compile error on the typed field, never a silent nil.
//
// This replaces the prior reflection-based wiring (wiring.go's ucAggregate /
// assertUseCases / ptrField / execFn navigators), per Q-WIRE-1: drift between
// espyna and cyta is now a compile error, not a silent no-op.
package block

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"

	eventpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/event/event"
	eventattendeepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/event/event_attendee"
	eventoccurrencepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/event/event_occurrence"
	eventproductpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/event/event_product"
	eventresourcepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/event/event_resource"
	eventtagpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/event/event_tag"

	eventdashboard "github.com/erniealice/cyta-golang/domain/event/event/dashboard"
	eventform "github.com/erniealice/cyta-golang/domain/event/event/form"
)

// UseCases declares everything cyta's Block() needs from outside.
// Construction is service-admin's job; cyta only declares the shape.
//
// Naming conventions (mirrors entydad/block/usecases.go):
//  1. Field names are SINGULAR matching the proto folder name.
//  2. Group struct types use the `<Entity>UseCases` suffix.
//  3. Leaf closure signatures use proto request/response types (no
//     block-local transport types) — these match the cyta view ModuleDeps
//     signatures exactly so block.go can assign them straight through.
//  4. Derived join closures (tag/attendee pickers, set/sync) and the
//     schedule dashboard are exposed in their final VIEW shape. They compose
//     multiple espyna use cases and/or reference espyna-internal request
//     types (e.g. SetEventTagAssignmentsRequest) that cyta cannot import, so
//     service-admin's adapter constructs them. This is the same boundary
//     fayna uses for its dashboard translations.
type UseCases struct {
	// GetWorkspaceIDFromCtx extracts the workspace ID from a request context.
	// Wired by service-admin as consumer.GetWorkspaceIDFromContext. Used as
	// the empty-workspace fallback for the schedule dashboard.
	GetWorkspaceIDFromCtx func(ctx context.Context) string

	// Event — leaf CRUD + nested-entity list closures + derived join closures.
	Event EventUseCases

	// EventTag — master tag CRUD + list (the EventTag module).
	EventTag EventTagUseCases

	// GetScheduleDashboardData is the schedule dashboard's typed view-returning
	// slot. Service-admin's adapter calls espyna's GetScheduleDashboard use case
	// and maps the proto response (ScheduleStats / ScheduleTagSlice) into the
	// cyta view shape. Nillable — when nil the dashboard renders empty stats.
	GetScheduleDashboardData func(ctx context.Context, req *eventdashboard.Request) (*eventdashboard.Response, error)
}

// EventUseCases — the Event module's use-case surface.
//
// The first group (Create..ListEventOccurrences) are proto-typed leaf closures
// that match eventmod.ModuleDeps field-for-field. The second group are derived
// join/picker closures already in their final view shape — built by the
// service-admin adapter because they compose multiple espyna use cases and
// touch espyna-internal request types cyta cannot import.
type EventUseCases struct {
	// --- Proto-typed leaf CRUD (required by RequireFor) ---
	Create func(ctx context.Context, req *eventpb.CreateEventRequest) (*eventpb.CreateEventResponse, error)
	Read   func(ctx context.Context, req *eventpb.ReadEventRequest) (*eventpb.ReadEventResponse, error)
	Update func(ctx context.Context, req *eventpb.UpdateEventRequest) (*eventpb.UpdateEventResponse, error)
	Delete func(ctx context.Context, req *eventpb.DeleteEventRequest) (*eventpb.DeleteEventResponse, error)
	List   func(ctx context.Context, req *eventpb.ListEventsRequest) (*eventpb.ListEventsResponse, error)

	// --- Proto-typed nested-entity lists (optional; nil → tab renders empty) ---
	ListAttendees   func(ctx context.Context, req *eventattendeepb.ListEventAttendeesRequest) (*eventattendeepb.ListEventAttendeesResponse, error)
	ListResources   func(ctx context.Context, req *eventresourcepb.ListEventResourcesRequest) (*eventresourcepb.ListEventResourcesResponse, error)
	ListProducts    func(ctx context.Context, req *eventproductpb.ListEventProductsRequest) (*eventproductpb.ListEventProductsResponse, error)
	ListOccurrences func(ctx context.Context, req *eventoccurrencepb.ListEventOccurrencesRequest) (*eventoccurrencepb.ListEventOccurrencesResponse, error)

	// --- Derived drawer-picker closures, in final view shape (all optional) ---
	// Built by service-admin's adapter. Nil → the drawer picker degrades to
	// pre-selected-only (matching the prior reflection-nil behaviour).
	ListTagOptions         func(ctx context.Context) ([]eventform.Option, error)
	ListTagsForEvent       func(ctx context.Context, eventID string) ([]string, error)
	SearchAttendees        func(ctx context.Context, query string) ([]eventform.Option, error)
	ListAttendeesForEvent  func(ctx context.Context, eventID string) ([]eventform.SelectedOption, error)
	SetEventTagAssignments func(ctx context.Context, eventID string, tagIDs []string) error
	SyncEventAttendees     func(ctx context.Context, eventID string, attendeeRefs []string) error
}

// EventTagUseCases — master EventTag CRUD + list. All proto-typed leaf closures
// matching eventtagmod.ModuleDeps field-for-field.
type EventTagUseCases struct {
	Create          func(ctx context.Context, req *eventtagpb.CreateEventTagRequest) (*eventtagpb.CreateEventTagResponse, error)
	Read            func(ctx context.Context, req *eventtagpb.ReadEventTagRequest) (*eventtagpb.ReadEventTagResponse, error)
	Update          func(ctx context.Context, req *eventtagpb.UpdateEventTagRequest) (*eventtagpb.UpdateEventTagResponse, error)
	Delete          func(ctx context.Context, req *eventtagpb.DeleteEventTagRequest) (*eventtagpb.DeleteEventTagResponse, error)
	List            func(ctx context.Context, req *eventtagpb.ListEventTagsRequest) (*eventtagpb.ListEventTagsResponse, error)
	GetListPageData func(ctx context.Context, req *eventtagpb.GetEventTagListPageDataRequest) (*eventtagpb.GetEventTagListPageDataResponse, error)
}

// RequireFor returns an error listing every needed-but-nil field for cfg's
// enabled modules. Called at Block() entry; a missing field → startup error.
//
// CRITICAL: this is the deterministic completeness check that replaces the
// prior silent-nil reflection wiring. Partial wiring is a startup error, not a
// runtime nil panic. Only the always-present leaf operations are required;
// derived picker closures + the dashboard are intentionally optional (they
// degrade gracefully), matching the prior reflection-nil behaviour.
func (u *UseCases) RequireFor(cfg *blockConfig) error {
	if u == nil {
		return fmt.Errorf("cyta.Block: WithUseCases(...) was not supplied")
	}

	var missing []string
	check := func(ok bool, name string) {
		if !ok {
			missing = append(missing, name)
		}
	}

	if cfg.wantEvent() {
		check(u.Event.Create != nil, "UseCases.Event.Create")
		check(u.Event.Read != nil, "UseCases.Event.Read")
		check(u.Event.Update != nil, "UseCases.Event.Update")
		check(u.Event.Delete != nil, "UseCases.Event.Delete")
		check(u.Event.List != nil, "UseCases.Event.List")
		// Nested-entity lists + derived pickers + dashboard are optional
		// (nil-safe wiring; the tab/picker simply renders empty).
	}

	if cfg.wantEventTag() {
		check(u.EventTag.Create != nil, "UseCases.EventTag.Create")
		check(u.EventTag.Read != nil, "UseCases.EventTag.Read")
		check(u.EventTag.Update != nil, "UseCases.EventTag.Update")
		check(u.EventTag.Delete != nil, "UseCases.EventTag.Delete")
		check(u.EventTag.List != nil, "UseCases.EventTag.List")
		// Required: the event-tag list page calls GetEventTagListPageData
		// UNGUARDED (views/event_tag/list/page.go), so a missing closure is a
		// boot error, not a silent runtime nil panic.
		check(u.EventTag.GetListPageData != nil, "UseCases.EventTag.GetListPageData")
	}

	if len(missing) > 0 {
		return fmt.Errorf("cyta.Block: incomplete UseCases — missing %v", missing)
	}
	return nil
}

// MustValidate is the FAIL-CLOSED enforcement wrapper around RequireFor. It is
// the seam-level guard that makes a missing REQUIRED closure impossible to
// ignore — mirroring the AUTHZ_ENFORCE boot-guard in service-admin's
// container.go (a missing security precondition is a boot REFUSAL, never a
// silent degrade).
//
// Why a wrapper and not just `return RequireFor(...)`: a bare returned error is
// fail-OPEN by convention. A caller can drop it (`_ =`, an ignored value, a
// future app that doesn't check) and the block silently registers an empty
// feature — the exact nil-closure trap the architecture roast (burn #1) named.
// MustValidate removes that escape hatch:
//
//   - In dev/test (running under `go test`, OR CYTA_BLOCK_STRICT truthy) a
//     missing REQUIRED closure PANICS with the full field list. A panic cannot
//     be silently dropped, prints a stack trace at the offending wiring site,
//     and fails the test/CI loudly. This is where a developer wiring a new
//     entity discovers a gap — at their desk, not in prod.
//   - In prod a missing REQUIRED closure logs a screaming FATAL line at the
//     seam (so even a caller that drops the returned error leaves an
//     unmissable log record) AND returns the error so Block() propagates it and
//     NewServiceAdmin halts boot with a clear "domain block failed" message.
//
// OPTIONAL ports (Event nested-entity lists, derived picker closures, the
// schedule dashboard) are NEVER flagged — that required-vs-optional
// discrimination lives entirely in RequireFor, which only asserts a field when
// its enabling cfg.wantXxx() module is on. MustValidate adds posture, not
// policy: it changes HOW a gap fails, not WHICH fields gate.
func (u *UseCases) MustValidate(cfg *blockConfig) error {
	err := u.RequireFor(cfg)
	if err == nil {
		return nil
	}
	if blockStrictMode() {
		// Dev/test: loud, uncatchable-by-accident, stack-traced.
		panic("FATAL: " + err.Error() + " — REQUIRED block wiring is nil. " +
			"Fix the closure assignment in service-admin's buildCytaUseCases " +
			"(adapters.go) before this reaches prod.")
	}
	// Prod: scream at the seam, then return so boot halts. The log line is the
	// belt to the returned-error's suspenders (a dropped error still screams).
	log.Printf("FATAL: %v — refusing to register cyta modules with a nil "+
		"REQUIRED closure (fail-closed wiring).", err)
	return err
}

// blockStrictMode reports whether the fail-closed wiring guard should PANIC
// (dev/test) rather than return-and-log (prod) on a missing REQUIRED closure.
//
// True when running under `go test` (testing.Testing(), Go 1.21+ — zero env
// coupling, auto-on in every test + CI run) OR when CYTA_BLOCK_STRICT is set to
// an explicit truthy value (the dev escape hatch for `go run` smoke tests).
// The env matching mirrors container.go's authzEnforceEnabled — anything else
// (unset, "", "0", "false") is prod posture.
func blockStrictMode() bool {
	if testing.Testing() {
		return true
	}
	switch os.Getenv("CYTA_BLOCK_STRICT") {
	case "1", "true", "TRUE", "True", "yes", "on":
		return true
	default:
		return false
	}
}
