package block

import (
	"context"
	"strings"
	"testing"

	eventpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/event/event"
	eventtagpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/event/event_tag"
)

// ---------------------------------------------------------------------------
// MustValidate — FAIL-CLOSED wiring guard (architecture-roast burn #1).
//
// RequireFor returns an error; MustValidate adds the posture: in dev/test
// (testing.Testing() is true here) a missing REQUIRED closure PANICS — loud,
// stack-traced, uncatchable-by-accident — so a nil-closure wiring gap can never
// be silently dropped into an empty-state render. OPTIONAL nils never trip it.
// ---------------------------------------------------------------------------

// wireEventRequired sets every closure RequireFor checks for the Event module.
func wireEventRequired(uc *UseCases) {
	ev := &uc.Event
	ev.Create = func(context.Context, *eventpb.CreateEventRequest) (*eventpb.CreateEventResponse, error) {
		return nil, nil
	}
	ev.Read = func(context.Context, *eventpb.ReadEventRequest) (*eventpb.ReadEventResponse, error) {
		return nil, nil
	}
	ev.Update = func(context.Context, *eventpb.UpdateEventRequest) (*eventpb.UpdateEventResponse, error) {
		return nil, nil
	}
	ev.Delete = func(context.Context, *eventpb.DeleteEventRequest) (*eventpb.DeleteEventResponse, error) {
		return nil, nil
	}
	ev.List = func(context.Context, *eventpb.ListEventsRequest) (*eventpb.ListEventsResponse, error) {
		return nil, nil
	}
}

// wireEventTagRequired sets every closure RequireFor checks for the EventTag module.
func wireEventTagRequired(uc *UseCases) {
	tag := &uc.EventTag
	tag.Create = func(context.Context, *eventtagpb.CreateEventTagRequest) (*eventtagpb.CreateEventTagResponse, error) {
		return nil, nil
	}
	tag.Read = func(context.Context, *eventtagpb.ReadEventTagRequest) (*eventtagpb.ReadEventTagResponse, error) {
		return nil, nil
	}
	tag.Update = func(context.Context, *eventtagpb.UpdateEventTagRequest) (*eventtagpb.UpdateEventTagResponse, error) {
		return nil, nil
	}
	tag.Delete = func(context.Context, *eventtagpb.DeleteEventTagRequest) (*eventtagpb.DeleteEventTagResponse, error) {
		return nil, nil
	}
	tag.List = func(context.Context, *eventtagpb.ListEventTagsRequest) (*eventtagpb.ListEventTagsResponse, error) {
		return nil, nil
	}
	tag.GetListPageData = func(context.Context, *eventtagpb.GetEventTagListPageDataRequest) (*eventtagpb.GetEventTagListPageDataResponse, error) {
		return nil, nil
	}
}

// TestMustValidate_NilRequiredClosure_Panics is the core burn-#1 proof: with
// the Event module enabled but one REQUIRED closure (List) left nil,
// MustValidate must PANIC under test — not return an empty render, not silently
// degrade. This is the loud failure the bare-return path lacked.
func TestMustValidate_NilRequiredClosure_Panics(t *testing.T) {
	t.Parallel()

	uc := &UseCases{}
	wireEventRequired(uc)
	uc.Event.List = nil // drop exactly one REQUIRED closure

	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("MustValidate(Event enabled, List nil) should PANIC in dev/test, but did not")
		}
		msg, _ := r.(string)
		if !strings.Contains(msg, "List") {
			t.Fatalf("panic message should name the missing field; got %q", msg)
		}
	}()

	// Should not reach the next line — MustValidate panics first.
	_ = uc.MustValidate(&blockConfig{event: true})
	t.Fatal("MustValidate returned instead of panicking on a nil REQUIRED closure")
}

// TestMustValidate_EmptyUseCases_EnableAll_Panics: a fully empty UseCases with
// every module enabled (the "permanently nil dashboard" trap) must panic loudly
// in dev/test rather than register a wall of empty views.
func TestMustValidate_EmptyUseCases_EnableAll_Panics(t *testing.T) {
	t.Parallel()

	uc := &UseCases{}
	defer func() {
		if recover() == nil {
			t.Fatal("MustValidate(empty UseCases, enableAll) should PANIC in dev/test")
		}
	}()
	_ = uc.MustValidate(&blockConfig{enableAll: true})
	t.Fatal("MustValidate returned instead of panicking on an empty enableAll wiring")
}

// TestMustValidate_NilOptionalClosure_OK proves the required-vs-optional
// discrimination survives the fail-closed wrapper: the OPTIONAL derived picker
// and dashboard closures with nil values must pass MustValidate with NO panic
// and NO error — disabled/optional features stay legitimately nil.
func TestMustValidate_NilOptionalClosure_OK(t *testing.T) {
	t.Parallel()

	uc := &UseCases{}
	// Wire only the REQUIRED event closures; leave all OPTIONAL closures nil
	// (nested-entity lists, derived pickers, schedule dashboard).
	wireEventRequired(uc)

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("MustValidate(optional nil closures) must NOT panic; panicked with %v", r)
		}
	}()
	if err := uc.MustValidate(&blockConfig{event: true}); err != nil {
		t.Fatalf("MustValidate(optional nil closures) should be nil, got %v", err)
	}
}

// TestMustValidate_FullyWired_OK: a completely wired REQUIRED set passes with no
// panic and no error (happy path — guard is silent when wiring is complete).
func TestMustValidate_FullyWired_OK(t *testing.T) {
	t.Parallel()

	uc := &UseCases{}
	wireEventRequired(uc)
	wireEventTagRequired(uc)

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("MustValidate(fully wired Event+EventTag) must NOT panic; panicked with %v", r)
		}
	}()
	if err := uc.MustValidate(&blockConfig{enableAll: true}); err != nil {
		t.Fatalf("MustValidate(fully wired Event+EventTag) should be nil, got %v", err)
	}
}
