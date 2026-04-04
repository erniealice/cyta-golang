package recurrence

import (
	"testing"
	"time"
)

// ---------------------------------------------------------------------------
// DefaultConfig tests
// ---------------------------------------------------------------------------

func TestDefaultConfig_Values(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Horizon != DefaultHorizon {
		t.Errorf("Horizon = %v, want %v", cfg.Horizon, DefaultHorizon)
	}
	if cfg.BatchSize != DefaultBatchSize {
		t.Errorf("BatchSize = %d, want %d", cfg.BatchSize, DefaultBatchSize)
	}
}

// ---------------------------------------------------------------------------
// NewOccurrenceWriter tests
// ---------------------------------------------------------------------------

func TestNewOccurrenceWriter_ZeroConfig_AppliesDefaults(t *testing.T) {
	w := NewOccurrenceWriter(OccurrenceWriterConfig{})

	if w.config.Horizon != DefaultHorizon {
		t.Errorf("Horizon = %v, want %v (default)", w.config.Horizon, DefaultHorizon)
	}
	if w.config.BatchSize != DefaultBatchSize {
		t.Errorf("BatchSize = %d, want %d (default)", w.config.BatchSize, DefaultBatchSize)
	}
	if w.expander == nil {
		t.Error("expander is nil, want non-nil")
	}
}

func TestNewOccurrenceWriter_CustomConfig_Preserved(t *testing.T) {
	cfg := OccurrenceWriterConfig{
		Horizon:   30 * 24 * time.Hour,
		BatchSize: 100,
	}
	w := NewOccurrenceWriter(cfg)

	if w.config.Horizon != cfg.Horizon {
		t.Errorf("Horizon = %v, want %v", w.config.Horizon, cfg.Horizon)
	}
	if w.config.BatchSize != cfg.BatchSize {
		t.Errorf("BatchSize = %d, want %d", w.config.BatchSize, cfg.BatchSize)
	}
}

func TestNewOccurrenceWriter_NegativeValues_AppliesDefaults(t *testing.T) {
	w := NewOccurrenceWriter(OccurrenceWriterConfig{
		Horizon:   -1 * time.Hour,
		BatchSize: -10,
	})

	if w.config.Horizon != DefaultHorizon {
		t.Errorf("Horizon = %v, want %v (default for negative input)", w.config.Horizon, DefaultHorizon)
	}
	if w.config.BatchSize != DefaultBatchSize {
		t.Errorf("BatchSize = %d, want %d (default for negative input)", w.config.BatchSize, DefaultBatchSize)
	}
}

// ---------------------------------------------------------------------------
// ExpandEvent tests
// ---------------------------------------------------------------------------

func TestExpandEvent_ValidInputs(t *testing.T) {
	w := NewOccurrenceWriter(DefaultConfig())

	tests := []struct {
		name        string
		eventID     string
		workspaceID string
		rrule       string
		exdates     string
		dtStart     time.Time
		duration    time.Duration
		wantCount   int
	}{
		{
			name:        "daily for 5 days",
			eventID:     "evt-001",
			workspaceID: "ws-100",
			rrule:       "FREQ=DAILY;COUNT=5",
			exdates:     "",
			dtStart:     baseStart,
			duration:    time.Hour,
			wantCount:   5,
		},
		{
			name:        "weekly 3 occurrences",
			eventID:     "evt-002",
			workspaceID: "ws-200",
			rrule:       "FREQ=WEEKLY;COUNT=3",
			exdates:     "",
			dtStart:     baseStart,
			duration:    2 * time.Hour,
			wantCount:   3,
		},
		{
			name:        "daily 5 with 1 exdate",
			eventID:     "evt-003",
			workspaceID: "ws-300",
			rrule:       "FREQ=DAILY;COUNT=5",
			exdates:     "2026-01-07T09:00:00Z",
			dtStart:     baseStart,
			duration:    time.Hour,
			wantCount:   4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			occs, err := w.ExpandEvent(tt.eventID, tt.workspaceID, tt.rrule, tt.exdates, tt.dtStart, tt.duration)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(occs) != tt.wantCount {
				t.Fatalf("got %d occurrences, want %d", len(occs), tt.wantCount)
			}

			for i, occ := range occs {
				if occ.EventID != tt.eventID {
					t.Errorf("occurrence %d: EventID = %q, want %q", i, occ.EventID, tt.eventID)
				}
				if occ.WorkspaceID != tt.workspaceID {
					t.Errorf("occurrence %d: WorkspaceID = %q, want %q", i, occ.WorkspaceID, tt.workspaceID)
				}
			}
		})
	}
}

func TestExpandEvent_EmptyEventID_ReturnsError(t *testing.T) {
	w := NewOccurrenceWriter(DefaultConfig())

	_, err := w.ExpandEvent("", "ws-100", "FREQ=DAILY;COUNT=3", "", baseStart, time.Hour)
	if err == nil {
		t.Fatal("expected error for empty eventID, got nil")
	}
}

func TestExpandEvent_EmptyWorkspaceID_ReturnsError(t *testing.T) {
	w := NewOccurrenceWriter(DefaultConfig())

	_, err := w.ExpandEvent("evt-001", "", "FREQ=DAILY;COUNT=3", "", baseStart, time.Hour)
	if err == nil {
		t.Fatal("expected error for empty workspaceID, got nil")
	}
}

// ---------------------------------------------------------------------------
// Negative / defensive ExpandEvent tests
// ---------------------------------------------------------------------------

func TestExpandEvent_InvalidRRULE_ReturnsError(t *testing.T) {
	w := NewOccurrenceWriter(DefaultConfig())

	_, err := w.ExpandEvent("evt-001", "ws-100", "FREQ=NOTAFREQUENCY;BADKEY=YES", "", baseStart, time.Hour)
	if err == nil {
		t.Fatal("expected error for invalid RRULE string, got nil")
	}
}

func TestExpandEvent_InvalidExdateFormat_ReturnsError(t *testing.T) {
	w := NewOccurrenceWriter(DefaultConfig())

	_, err := w.ExpandEvent("evt-001", "ws-100", "FREQ=DAILY;COUNT=3", "not-a-date,also-bad", baseStart, time.Hour)
	if err == nil {
		t.Fatal("expected error for invalid exdate format, got nil")
	}
}

func TestExpandEvent_VeryLongRRULE_ManyOccurrences(t *testing.T) {
	w := NewOccurrenceWriter(DefaultConfig())

	// Daily with COUNT=1500 — the 2-year default horizon caps the actual count.
	// Verify the expansion completes without panic or error and produces a
	// large number of occurrences (capped by the horizon window, ~730 days).
	occs, err := w.ExpandEvent("evt-001", "ws-100", "FREQ=DAILY;COUNT=1500", "", baseStart, time.Hour)
	if err != nil {
		t.Fatalf("unexpected error for large occurrence count: %v", err)
	}
	if len(occs) < 700 {
		t.Errorf("expected at least 700 occurrences within the 2-year horizon, got %d", len(occs))
	}
}

func TestExpandEvent_PastDtStart(t *testing.T) {
	w := NewOccurrenceWriter(DefaultConfig())

	pastStart := time.Date(2020, time.January, 1, 9, 0, 0, 0, time.UTC)
	occs, err := w.ExpandEvent("evt-001", "ws-100", "FREQ=DAILY;COUNT=3", "", pastStart, time.Hour)
	if err != nil {
		t.Fatalf("unexpected error for past dtStart: %v", err)
	}
	if len(occs) != 3 {
		t.Errorf("expected 3 occurrences with past dtStart, got %d", len(occs))
	}
	// Verify occurrences are stamped correctly even when in the past.
	for i, occ := range occs {
		if occ.EventID != "evt-001" {
			t.Errorf("occurrence %d: EventID = %q, want %q", i, occ.EventID, "evt-001")
		}
		if occ.WorkspaceID != "ws-100" {
			t.Errorf("occurrence %d: WorkspaceID = %q, want %q", i, occ.WorkspaceID, "ws-100")
		}
	}
}
