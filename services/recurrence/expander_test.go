package recurrence

import (
	"testing"
	"time"
)

// baseStart is a fixed Monday at 09:00 UTC used as a stable dtStart across tests.
var baseStart = time.Date(2026, time.January, 5, 9, 0, 0, 0, time.UTC) // Monday

func newExpander() *Expander { return NewExpander() }

// ---------------------------------------------------------------------------
// ExpandRRule tests
// ---------------------------------------------------------------------------

func TestExpandRRule_DailyForSevenDays(t *testing.T) {
	e := newExpander()

	occs, err := e.ExpandRRule(
		"FREQ=DAILY;COUNT=7",
		baseStart,
		time.Hour, // 1-hour events
		365*24*time.Hour,
		nil,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(occs) != 7 {
		t.Errorf("expected 7 occurrences, got %d", len(occs))
	}

	// Verify first occurrence start matches dtStart.
	if !occs[0].StartDateTimeUTC.Equal(baseStart) {
		t.Errorf("first occurrence start: got %v, want %v", occs[0].StartDateTimeUTC, baseStart)
	}

	// Verify end = start + duration for each occurrence.
	for i, occ := range occs {
		want := occ.StartDateTimeUTC.Add(time.Hour)
		if !occ.EndDateTimeUTC.Equal(want) {
			t.Errorf("occurrence %d: end time = %v, want %v", i, occ.EndDateTimeUTC, want)
		}
	}
}

func TestExpandRRule_WeeklyMoWeFrForTwoWeeks(t *testing.T) {
	e := newExpander()

	// baseStart is a Monday; MO,WE,FR over 2 weeks = 6 occurrences.
	occs, err := e.ExpandRRule(
		"FREQ=WEEKLY;BYDAY=MO,WE,FR;COUNT=6",
		baseStart,
		30*time.Minute,
		365*24*time.Hour,
		nil,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(occs) != 6 {
		t.Errorf("expected 6 occurrences, got %d", len(occs))
	}

	// Week 1: Mon Jan 5, Wed Jan 7, Fri Jan 9
	// Week 2: Mon Jan 12, Wed Jan 14, Fri Jan 16
	expectedDays := []int{5, 7, 9, 12, 14, 16}
	for i, occ := range occs {
		if occ.StartDateTimeUTC.Day() != expectedDays[i] {
			t.Errorf("occurrence %d: day = %d, want %d",
				i, occ.StartDateTimeUTC.Day(), expectedDays[i])
		}
	}
}

func TestExpandRRule_MonthlyWithCount3(t *testing.T) {
	e := newExpander()

	occs, err := e.ExpandRRule(
		"FREQ=MONTHLY;COUNT=3",
		baseStart,
		2*time.Hour,
		365*24*time.Hour,
		nil,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(occs) != 3 {
		t.Errorf("expected 3 occurrences, got %d", len(occs))
	}

	// Jan 5, Feb 5, Mar 5
	expectedMonths := []time.Month{time.January, time.February, time.March}
	for i, occ := range occs {
		if occ.StartDateTimeUTC.Month() != expectedMonths[i] {
			t.Errorf("occurrence %d: month = %v, want %v",
				i, occ.StartDateTimeUTC.Month(), expectedMonths[i])
		}
	}
}

func TestExpandRRule_StopsAtUntilDate(t *testing.T) {
	e := newExpander()

	// UNTIL is inclusive. Daily from Jan 5 until Jan 10 = 6 occurrences.
	occs, err := e.ExpandRRule(
		"FREQ=DAILY;UNTIL=20260110T090000Z",
		baseStart,
		time.Hour,
		365*24*time.Hour,
		nil,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(occs) != 6 {
		t.Errorf("expected 6 occurrences (Jan 5–10), got %d", len(occs))
	}

	last := occs[len(occs)-1]
	if last.StartDateTimeUTC.Day() != 10 {
		t.Errorf("last occurrence day = %d, want 10", last.StartDateTimeUTC.Day())
	}
}

func TestExpandRRule_ExdatesAreFiltered(t *testing.T) {
	e := newExpander()

	// Daily for 5 days; exclude Jan 7 (day index 2).
	exdates := []time.Time{
		time.Date(2026, time.January, 7, 0, 0, 0, 0, time.UTC),
	}

	occs, err := e.ExpandRRule(
		"FREQ=DAILY;COUNT=5",
		baseStart,
		time.Hour,
		365*24*time.Hour,
		exdates,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 5 total − 1 excluded = 4
	if len(occs) != 4 {
		t.Errorf("expected 4 occurrences after exdate, got %d", len(occs))
	}

	// Confirm Jan 7 is absent.
	for _, occ := range occs {
		if occ.StartDateTimeUTC.Day() == 7 {
			t.Errorf("occurrence on Jan 7 should have been excluded")
		}
	}
}

func TestExpandRRule_InvalidRRule_ReturnsError(t *testing.T) {
	e := newExpander()

	_, err := e.ExpandRRule(
		"FREQ=NOTAFREQUENCY",
		baseStart,
		time.Hour,
		365*24*time.Hour,
		nil,
	)
	if err == nil {
		t.Fatal("expected error for invalid RRULE, got nil")
	}
}

// ---------------------------------------------------------------------------
// ParseExdates tests
// ---------------------------------------------------------------------------

func TestParseExdates_EmptyString(t *testing.T) {
	result, err := ParseExdates("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Errorf("expected nil for empty string, got %v", result)
	}
}

func TestParseExdates_RFC3339(t *testing.T) {
	result, err := ParseExdates("2026-01-07T09:00:00Z,2026-01-14T09:00:00Z")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 exdates, got %d", len(result))
	}
	if result[0].Day() != 7 {
		t.Errorf("first exdate day = %d, want 7", result[0].Day())
	}
	if result[1].Day() != 14 {
		t.Errorf("second exdate day = %d, want 14", result[1].Day())
	}
}

func TestParseExdates_DateOnly(t *testing.T) {
	result, err := ParseExdates("2026-03-01,2026-04-15")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 exdates, got %d", len(result))
	}
}

func TestParseExdates_InvalidEntry_ReturnsError(t *testing.T) {
	_, err := ParseExdates("not-a-date")
	if err == nil {
		t.Fatal("expected error for invalid exdate, got nil")
	}
}

// ---------------------------------------------------------------------------
// Edge-case tests
// ---------------------------------------------------------------------------

func TestExpandRRule_EmptyRRuleString_ReturnsError(t *testing.T) {
	e := newExpander()
	_, err := e.ExpandRRule("", baseStart, time.Hour, 365*24*time.Hour, nil)
	if err == nil {
		t.Fatal("expected error for empty rrule string, got nil")
	}
}

func TestExpandRRule_NegativeHorizon_ReturnsError(t *testing.T) {
	e := newExpander()
	_, err := e.ExpandRRule("FREQ=DAILY;COUNT=5", baseStart, time.Hour, -time.Hour, nil)
	if err == nil {
		t.Fatal("expected error for negative horizon, got nil")
	}
}

func TestExpandRRule_AllTimesUTC(t *testing.T) {
	e := newExpander()

	// Supply a non-UTC dtStart to confirm it is normalised.
	eastern, _ := time.LoadLocation("America/New_York")
	localStart := time.Date(2026, time.January, 5, 4, 0, 0, 0, eastern) // = 09:00 UTC

	occs, err := e.ExpandRRule(
		"FREQ=DAILY;COUNT=3",
		localStart,
		time.Hour,
		365*24*time.Hour,
		nil,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for i, occ := range occs {
		if occ.StartDateTimeUTC.Location() != time.UTC {
			t.Errorf("occurrence %d StartDateTimeUTC is not UTC: %v", i, occ.StartDateTimeUTC.Location())
		}
		if occ.EndDateTimeUTC.Location() != time.UTC {
			t.Errorf("occurrence %d EndDateTimeUTC is not UTC: %v", i, occ.EndDateTimeUTC.Location())
		}
	}
}
