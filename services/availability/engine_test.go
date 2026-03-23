package availability

import (
	"testing"
	"time"
)

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

func mustUTC(s string) time.Time {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		panic("bad time literal: " + s)
	}
	return t.UTC()
}

// weekdaySchedule builds a Schedule with a single daily window covering the
// supplied weekdays, from startH:00 to endH:00 UTC.
func weekdaySchedule(days []time.Weekday, startH, endH int) Schedule {
	windows := make([]TimeWindow, len(days))
	for i, d := range days {
		windows[i] = TimeWindow{
			DayOfWeek: d,
			StartHour: startH,
			StartMin:  0,
			EndHour:   endH,
			EndMin:    0,
		}
	}
	return Schedule{
		ResourceID:   "res-1",
		ResourceType: "staff",
		WorkingHours: windows,
		Timezone:     "UTC",
	}
}

// allWeekdays returns Monday–Friday.
func allWeekdays() []time.Weekday {
	return []time.Weekday{
		time.Monday,
		time.Tuesday,
		time.Wednesday,
		time.Thursday,
		time.Friday,
	}
}

// ---------------------------------------------------------------------------
// ComputeSlots tests
// ---------------------------------------------------------------------------

// TestComputeSlots_NoBookings verifies that with no existing bookings the
// engine returns the full set of candidate slots covering the working hours.
func TestComputeSlots_NoBookings(t *testing.T) {
	engine := NewAvailabilityEngine()

	// Monday 2026-03-23, 9am–5pm, 1-hour slots → expect 8 slots.
	sched := weekdaySchedule([]time.Weekday{time.Monday}, 9, 17)
	windowStart := mustUTC("2026-03-23T09:00:00Z")
	windowEnd := mustUTC("2026-03-23T17:00:00Z")
	slotDur := time.Hour

	slots := engine.ComputeSlots(sched, nil, windowStart, windowEnd, slotDur)

	const want = 8
	if len(slots) != want {
		t.Fatalf("expected %d slots, got %d", want, len(slots))
	}
	for _, s := range slots {
		if s.Status != "free" {
			t.Errorf("expected status 'free', got %q", s.Status)
		}
		if s.Duration != slotDur {
			t.Errorf("expected duration %v, got %v", slotDur, s.Duration)
		}
	}
	// Verify first and last slot boundaries.
	if !slots[0].Start.Equal(windowStart) {
		t.Errorf("first slot start = %v, want %v", slots[0].Start, windowStart)
	}
	if !slots[want-1].End.Equal(windowEnd) {
		t.Errorf("last slot end = %v, want %v", slots[want-1].End, windowEnd)
	}
}

// TestComputeSlots_OneBookingInMiddle verifies that a single booking splits
// the working hours around it, removing the covered slots.
func TestComputeSlots_OneBookingInMiddle(t *testing.T) {
	engine := NewAvailabilityEngine()

	// Mon 9am-5pm, 1-hour slots → 8 candidate slots.
	// Booking 11am-1pm blocks 2 slots → expect 6 free slots.
	sched := weekdaySchedule([]time.Weekday{time.Monday}, 9, 17)
	windowStart := mustUTC("2026-03-23T09:00:00Z")
	windowEnd := mustUTC("2026-03-23T17:00:00Z")

	bookings := []Booking{
		{
			Start:   mustUTC("2026-03-23T11:00:00Z"),
			End:     mustUTC("2026-03-23T13:00:00Z"),
			EventID: "ev-1",
		},
	}

	slots := engine.ComputeSlots(sched, bookings, windowStart, windowEnd, time.Hour)

	const want = 6
	if len(slots) != want {
		t.Fatalf("expected %d slots, got %d", want, len(slots))
	}

	// None of the returned slots should overlap 11am–1pm.
	bookedStart := mustUTC("2026-03-23T11:00:00Z")
	bookedEnd := mustUTC("2026-03-23T13:00:00Z")
	for _, s := range slots {
		if intervalsOverlap(s.Start, s.End, bookedStart, bookedEnd) {
			t.Errorf("slot [%v, %v] overlaps the booking", s.Start, s.End)
		}
	}
}

// TestComputeSlots_BackToBackBookings verifies that when bookings fill the
// entire working period no free slots are returned.
func TestComputeSlots_BackToBackBookings(t *testing.T) {
	engine := NewAvailabilityEngine()

	sched := weekdaySchedule([]time.Weekday{time.Monday}, 9, 11)
	windowStart := mustUTC("2026-03-23T09:00:00Z")
	windowEnd := mustUTC("2026-03-23T11:00:00Z")

	// Two back-to-back 1-hour bookings cover the entire 2-hour window.
	bookings := []Booking{
		{Start: mustUTC("2026-03-23T09:00:00Z"), End: mustUTC("2026-03-23T10:00:00Z"), EventID: "ev-1"},
		{Start: mustUTC("2026-03-23T10:00:00Z"), End: mustUTC("2026-03-23T11:00:00Z"), EventID: "ev-2"},
	}

	slots := engine.ComputeSlots(sched, bookings, windowStart, windowEnd, time.Hour)

	if len(slots) != 0 {
		t.Fatalf("expected 0 free slots, got %d", len(slots))
	}
}

// TestComputeSlots_MultipleDay verifies that the engine spans multiple days
// and returns slots for each working day in the window.
func TestComputeSlots_MultipleDay(t *testing.T) {
	engine := NewAvailabilityEngine()

	// Mon–Fri, 9am–10am (1 slot per day), no bookings.
	// Window: Mon–Fri of one week → expect exactly 5 slots.
	sched := weekdaySchedule(allWeekdays(), 9, 10)
	windowStart := mustUTC("2026-03-23T00:00:00Z") // Monday
	windowEnd := mustUTC("2026-03-28T00:00:00Z")   // Saturday (exclusive)

	slots := engine.ComputeSlots(sched, nil, windowStart, windowEnd, time.Hour)

	const want = 5
	if len(slots) != want {
		t.Fatalf("expected %d slots (one per weekday), got %d", want, len(slots))
	}

	// Each slot should start at 09:00 on its respective day.
	for _, s := range slots {
		if s.Start.Hour() != 9 || s.Start.Minute() != 0 {
			t.Errorf("unexpected slot start time %v", s.Start)
		}
	}
}

// TestComputeSlots_ZeroDuration verifies that a zero slotDuration returns nil.
func TestComputeSlots_ZeroDuration(t *testing.T) {
	engine := NewAvailabilityEngine()
	sched := weekdaySchedule([]time.Weekday{time.Monday}, 9, 17)
	slots := engine.ComputeSlots(sched, nil,
		mustUTC("2026-03-23T09:00:00Z"),
		mustUTC("2026-03-23T17:00:00Z"),
		0,
	)
	if slots != nil {
		t.Fatalf("expected nil for zero duration, got %v", slots)
	}
}

// TestComputeSlots_WindowOutsideWorkingHours verifies that no slots are
// produced when the query window does not intersect any working hours.
func TestComputeSlots_WindowOutsideWorkingHours(t *testing.T) {
	engine := NewAvailabilityEngine()

	// Working hours 9am–5pm, but window is 6pm–8pm.
	sched := weekdaySchedule([]time.Weekday{time.Monday}, 9, 17)
	slots := engine.ComputeSlots(sched, nil,
		mustUTC("2026-03-23T18:00:00Z"),
		mustUTC("2026-03-23T20:00:00Z"),
		time.Hour,
	)
	if len(slots) != 0 {
		t.Fatalf("expected 0 slots, got %d", len(slots))
	}
}

// ---------------------------------------------------------------------------
// CheckConflicts tests
// ---------------------------------------------------------------------------

// TestCheckConflicts_WithOverlap verifies that an overlapping booking is
// detected and returned.
func TestCheckConflicts_WithOverlap(t *testing.T) {
	engine := NewAvailabilityEngine()

	bookings := []Booking{
		{Start: mustUTC("2026-03-23T10:00:00Z"), End: mustUTC("2026-03-23T11:00:00Z"), EventID: "ev-1"},
	}

	// Proposed window overlaps: 10:30–11:30 intersects 10:00–11:00.
	conflicts := engine.CheckConflicts(
		bookings,
		mustUTC("2026-03-23T10:30:00Z"),
		mustUTC("2026-03-23T11:30:00Z"),
	)

	if len(conflicts) != 1 {
		t.Fatalf("expected 1 conflict, got %d", len(conflicts))
	}
	if conflicts[0].EventID != "ev-1" {
		t.Errorf("expected EventID 'ev-1', got %q", conflicts[0].EventID)
	}
}

// TestCheckConflicts_NoOverlap verifies that a non-overlapping proposed window
// produces no conflicts.
func TestCheckConflicts_NoOverlap(t *testing.T) {
	engine := NewAvailabilityEngine()

	bookings := []Booking{
		{Start: mustUTC("2026-03-23T10:00:00Z"), End: mustUTC("2026-03-23T11:00:00Z"), EventID: "ev-1"},
	}

	// Proposed window is adjacent but does not overlap: 11:00–12:00.
	conflicts := engine.CheckConflicts(
		bookings,
		mustUTC("2026-03-23T11:00:00Z"),
		mustUTC("2026-03-23T12:00:00Z"),
	)

	if len(conflicts) != 0 {
		t.Fatalf("expected 0 conflicts, got %d", len(conflicts))
	}
}

// TestCheckConflicts_MultipleOverlaps verifies that all overlapping bookings
// are reported when more than one conflict exists.
func TestCheckConflicts_MultipleOverlaps(t *testing.T) {
	engine := NewAvailabilityEngine()

	bookings := []Booking{
		{Start: mustUTC("2026-03-23T09:00:00Z"), End: mustUTC("2026-03-23T10:00:00Z"), EventID: "ev-1"},
		{Start: mustUTC("2026-03-23T09:30:00Z"), End: mustUTC("2026-03-23T10:30:00Z"), EventID: "ev-2"},
		{Start: mustUTC("2026-03-23T11:00:00Z"), End: mustUTC("2026-03-23T12:00:00Z"), EventID: "ev-3"},
	}

	// Proposed 09:15–10:15 overlaps ev-1 and ev-2 but not ev-3.
	conflicts := engine.CheckConflicts(
		bookings,
		mustUTC("2026-03-23T09:15:00Z"),
		mustUTC("2026-03-23T10:15:00Z"),
	)

	if len(conflicts) != 2 {
		t.Fatalf("expected 2 conflicts, got %d", len(conflicts))
	}

	ids := map[string]bool{}
	for _, c := range conflicts {
		ids[c.EventID] = true
	}
	if !ids["ev-1"] || !ids["ev-2"] {
		t.Errorf("expected conflicts for ev-1 and ev-2, got %v", ids)
	}
}

// TestCheckConflicts_EmptyBookings verifies that no conflicts are reported
// when there are no existing bookings.
func TestCheckConflicts_EmptyBookings(t *testing.T) {
	engine := NewAvailabilityEngine()

	conflicts := engine.CheckConflicts(
		nil,
		mustUTC("2026-03-23T10:00:00Z"),
		mustUTC("2026-03-23T11:00:00Z"),
	)

	if len(conflicts) != 0 {
		t.Fatalf("expected 0 conflicts, got %d", len(conflicts))
	}
}

// ---------------------------------------------------------------------------
// ConflictChecker tests
// ---------------------------------------------------------------------------

// TestConflictChecker_HasConflict verifies the convenience boolean checker.
func TestConflictChecker_HasConflict(t *testing.T) {
	engine := NewAvailabilityEngine()
	bookings := []Booking{
		{Start: mustUTC("2026-03-23T10:00:00Z"), End: mustUTC("2026-03-23T11:00:00Z"), EventID: "ev-1"},
	}
	checker := NewConflictChecker(engine, bookings)

	if !checker.HasConflict(mustUTC("2026-03-23T10:30:00Z"), mustUTC("2026-03-23T11:30:00Z")) {
		t.Error("expected HasConflict to return true for overlapping window")
	}
	if checker.HasConflict(mustUTC("2026-03-23T11:00:00Z"), mustUTC("2026-03-23T12:00:00Z")) {
		t.Error("expected HasConflict to return false for adjacent (non-overlapping) window")
	}
}

// TestConflictChecker_ConflictsForSlots verifies batch slot conflict checking.
func TestConflictChecker_ConflictsForSlots(t *testing.T) {
	engine := NewAvailabilityEngine()
	bookings := []Booking{
		{Start: mustUTC("2026-03-23T10:00:00Z"), End: mustUTC("2026-03-23T11:00:00Z"), EventID: "ev-1"},
	}
	checker := NewConflictChecker(engine, bookings)

	slots := []Slot{
		{Start: mustUTC("2026-03-23T09:00:00Z"), End: mustUTC("2026-03-23T10:00:00Z"), Duration: time.Hour, Status: "free"}, // index 0 — no conflict
		{Start: mustUTC("2026-03-23T10:00:00Z"), End: mustUTC("2026-03-23T11:00:00Z"), Duration: time.Hour, Status: "free"}, // index 1 — exact overlap
		{Start: mustUTC("2026-03-23T11:00:00Z"), End: mustUTC("2026-03-23T12:00:00Z"), Duration: time.Hour, Status: "free"}, // index 2 — no conflict
	}

	result := checker.ConflictsForSlots(slots)

	if len(result) != 1 {
		t.Fatalf("expected 1 conflicting slot, got %d", len(result))
	}
	if _, ok := result[1]; !ok {
		t.Error("expected slot at index 1 to have conflicts")
	}
}
