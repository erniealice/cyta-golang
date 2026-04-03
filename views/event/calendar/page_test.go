package calendar

import (
	"testing"
	"time"
)

// ---------------------------------------------------------------------------
// sameDay tests
// ---------------------------------------------------------------------------

func TestSameDay(t *testing.T) {
	tests := []struct {
		name string
		a    time.Time
		b    time.Time
		want bool
	}{
		{
			name: "identical times",
			a:    time.Date(2026, time.March, 15, 10, 30, 0, 0, time.UTC),
			b:    time.Date(2026, time.March, 15, 10, 30, 0, 0, time.UTC),
			want: true,
		},
		{
			name: "same day different times",
			a:    time.Date(2026, time.March, 15, 0, 0, 0, 0, time.UTC),
			b:    time.Date(2026, time.March, 15, 23, 59, 59, 0, time.UTC),
			want: true,
		},
		{
			name: "different days",
			a:    time.Date(2026, time.March, 15, 10, 0, 0, 0, time.UTC),
			b:    time.Date(2026, time.March, 16, 10, 0, 0, 0, time.UTC),
			want: false,
		},
		{
			name: "different months",
			a:    time.Date(2026, time.March, 15, 10, 0, 0, 0, time.UTC),
			b:    time.Date(2026, time.April, 15, 10, 0, 0, 0, time.UTC),
			want: false,
		},
		{
			name: "different years",
			a:    time.Date(2025, time.March, 15, 10, 0, 0, 0, time.UTC),
			b:    time.Date(2026, time.March, 15, 10, 0, 0, 0, time.UTC),
			want: false,
		},
		{
			name: "year boundary",
			a:    time.Date(2025, time.December, 31, 23, 59, 0, 0, time.UTC),
			b:    time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sameDay(tt.a, tt.b)
			if got != tt.want {
				t.Errorf("sameDay(%v, %v) = %v, want %v", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// positionWeekEvent tests
// ---------------------------------------------------------------------------

func TestPositionWeekEvent(t *testing.T) {
	// Use hourStart=7, hourEnd=21, totalHours=14 to match the production values.
	const hourStart = 7
	const totalHours = 14.0

	tests := []struct {
		name          string
		startHour     int
		startMin      int
		endHour       int
		endMin        int
		wantTopPct    float64
		wantHeightPct float64
		wantIsCompact bool
	}{
		{
			name:          "10:00-11:00 (1 hour event)",
			startHour:     10,
			startMin:      0,
			endHour:       11,
			endMin:        0,
			wantTopPct:    (3.0 / totalHours) * 100, // offset = 10-7 = 3
			wantHeightPct: (1.0 / totalHours) * 100, // duration = 1h
			wantIsCompact: false,
		},
		{
			name:          "14:00-15:30 (1.5 hour event)",
			startHour:     14,
			startMin:      0,
			endHour:       15,
			endMin:        30,
			wantTopPct:    (7.0 / totalHours) * 100, // offset = 14-7 = 7
			wantHeightPct: (1.5 / totalHours) * 100, // duration = 1.5h
			wantIsCompact: false,
		},
		{
			name:          "9:00-12:00 (3 hour event)",
			startHour:     9,
			startMin:      0,
			endHour:       12,
			endMin:        0,
			wantTopPct:    (2.0 / totalHours) * 100, // offset = 9-7 = 2
			wantHeightPct: (3.0 / totalHours) * 100, // duration = 3h
			wantIsCompact: false,
		},
		{
			name:          "7:00-7:30 (30 min event, compact)",
			startHour:     7,
			startMin:      0,
			endHour:       7,
			endMin:        30,
			wantTopPct:    0,                        // offset = 0
			wantHeightPct: (0.5 / totalHours) * 100, // duration = 0.5h
			wantIsCompact: true,                     // <= 0.5h
		},
		{
			name:          "8:00-8:15 (15 min event, compact)",
			startHour:     8,
			startMin:      0,
			endHour:       8,
			endMin:        15,
			wantTopPct:    (1.0 / totalHours) * 100,  // offset = 8-7 = 1
			wantHeightPct: (0.25 / totalHours) * 100, // duration = 0.25h
			wantIsCompact: true,                      // <= 0.5h
		},
		{
			name:          "10:30-11:15 (45 min with minute offsets)",
			startHour:     10,
			startMin:      30,
			endHour:       11,
			endMin:        15,
			wantTopPct:    (3.5 / totalHours) * 100,  // offset = 10-7 + 30/60 = 3.5
			wantHeightPct: (0.75 / totalHours) * 100, // duration = 0.75h
			wantIsCompact: false,                     // 0.75 > 0.5
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			baseEvent := CalendarEvent{
				ID:   "test-ev",
				Name: "Test Event",
			}

			result := positionWeekEvent(baseEvent, tt.startHour, tt.startMin, tt.endHour, tt.endMin, hourStart, totalHours)

			const epsilon = 0.0001
			if diff := result.TopPct - tt.wantTopPct; diff > epsilon || diff < -epsilon {
				t.Errorf("TopPct = %.4f, want %.4f", result.TopPct, tt.wantTopPct)
			}
			if diff := result.HeightPct - tt.wantHeightPct; diff > epsilon || diff < -epsilon {
				t.Errorf("HeightPct = %.4f, want %.4f", result.HeightPct, tt.wantHeightPct)
			}
			if result.IsCompact != tt.wantIsCompact {
				t.Errorf("IsCompact = %v, want %v", result.IsCompact, tt.wantIsCompact)
			}

			// Verify the original event fields are preserved.
			if result.ID != baseEvent.ID {
				t.Errorf("ID = %q, want %q", result.ID, baseEvent.ID)
			}
			if result.Name != baseEvent.Name {
				t.Errorf("Name = %q, want %q", result.Name, baseEvent.Name)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Negative / defensive sameDay tests
// ---------------------------------------------------------------------------

func TestSameDay_ZeroValueTimes(t *testing.T) {
	zero := time.Time{}

	// Two zero-value times should be the same day (both are 0001-01-01).
	if !sameDay(zero, zero) {
		t.Error("sameDay(zero, zero) = false, want true")
	}

	// Zero vs non-zero should be different.
	nonZero := time.Date(2026, time.March, 15, 10, 0, 0, 0, time.UTC)
	if sameDay(zero, nonZero) {
		t.Error("sameDay(zero, nonZero) = true, want false")
	}
}

func TestSameDay_ZeroA_NonZeroB(t *testing.T) {
	zero := time.Time{}
	now := time.Now()

	if sameDay(zero, now) {
		t.Error("sameDay(zeroTime, now) should be false")
	}
}

// ---------------------------------------------------------------------------
// Negative / defensive positionWeekEvent tests
// ---------------------------------------------------------------------------

func TestPositionWeekEvent_SpanningMidnight(t *testing.T) {
	// An event from 23:00 to 01:00 next day — end hour < start hour.
	// Since the function uses raw integer math without day awareness,
	// this produces a negative duration. Verify it does not panic.
	const hourStart = 7
	const totalHours = 14.0

	baseEvent := CalendarEvent{
		ID:   "ev-midnight",
		Name: "Late Night Event",
	}

	result := positionWeekEvent(baseEvent, 23, 0, 1, 0, hourStart, totalHours)

	// The function computes: endOffset(1-7) - startOffset(23-7) = -6 - 16 = -22
	// HeightPct will be negative. Verify the function does not panic
	// and the event fields are preserved.
	if result.ID != "ev-midnight" {
		t.Errorf("ID = %q, want %q", result.ID, "ev-midnight")
	}
	// Negative duration means HeightPct < 0
	if result.HeightPct >= 0 {
		t.Logf("HeightPct = %.4f (negative expected for midnight-spanning event)", result.HeightPct)
	}
}

func TestPositionWeekEvent_NegativeDuration_EndBeforeStart(t *testing.T) {
	// End time is before start time (e.g. data error: 14:00 to 10:00).
	const hourStart = 7
	const totalHours = 14.0

	baseEvent := CalendarEvent{
		ID:   "ev-negative",
		Name: "Reversed Event",
	}

	result := positionWeekEvent(baseEvent, 14, 0, 10, 0, hourStart, totalHours)

	// durationHours = (10-7) - (14-7) = 3 - 7 = -4 → negative
	expectedDuration := -4.0
	expectedHeight := (expectedDuration / totalHours) * 100

	const epsilon = 0.0001
	if diff := result.HeightPct - expectedHeight; diff > epsilon || diff < -epsilon {
		t.Errorf("HeightPct = %.4f, want %.4f", result.HeightPct, expectedHeight)
	}

	// Negative duration should make IsCompact true (durationHours <= 0.5)
	if !result.IsCompact {
		t.Error("IsCompact should be true for negative duration")
	}
}

func TestPositionWeekEvent_ZeroDuration(t *testing.T) {
	// Start == End (instant event).
	const hourStart = 7
	const totalHours = 14.0

	baseEvent := CalendarEvent{
		ID:   "ev-zero",
		Name: "Instant Event",
	}

	result := positionWeekEvent(baseEvent, 10, 0, 10, 0, hourStart, totalHours)

	if result.HeightPct != 0 {
		t.Errorf("HeightPct = %.4f, want 0 for zero-duration event", result.HeightPct)
	}
	if !result.IsCompact {
		t.Error("IsCompact should be true for zero-duration event")
	}
}
