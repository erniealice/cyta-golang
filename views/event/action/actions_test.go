package action

import (
	"net/http"
	"testing"

	eventpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/event/event"
)

// ---------------------------------------------------------------------------
// eventStatusToEnum tests
// ---------------------------------------------------------------------------

func TestEventStatusToEnum(t *testing.T) {
	tests := []struct {
		input string
		want  eventpb.EventStatus
	}{
		{"tentative", eventpb.EventStatus_EVENT_STATUS_TENTATIVE},
		{"confirmed", eventpb.EventStatus_EVENT_STATUS_CONFIRMED},
		{"cancelled", eventpb.EventStatus_EVENT_STATUS_CANCELLED},
		{"", eventpb.EventStatus_EVENT_STATUS_UNSPECIFIED},
		{"unknown", eventpb.EventStatus_EVENT_STATUS_UNSPECIFIED},
		{"CONFIRMED", eventpb.EventStatus_EVENT_STATUS_UNSPECIFIED}, // case-sensitive
	}

	for _, tt := range tests {
		t.Run("status_"+tt.input, func(t *testing.T) {
			got := eventStatusToEnum(tt.input)
			if got != tt.want {
				t.Errorf("eventStatusToEnum(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// strPtr tests
// ---------------------------------------------------------------------------

func TestStrPtr(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"non-empty", "hello"},
		{"empty", ""},
		{"with spaces", "hello world"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ptr := strPtr(tt.input)
			if ptr == nil {
				t.Fatal("strPtr returned nil")
			}
			if *ptr != tt.input {
				t.Errorf("*strPtr(%q) = %q, want %q", tt.input, *ptr, tt.input)
			}
			// Verify it is a distinct allocation (pointer is not to the same variable).
			original := tt.input
			*ptr = "mutated"
			if original == "mutated" {
				t.Error("strPtr did not create an independent copy")
			}
		})
	}
}

// ---------------------------------------------------------------------------
// htmxSuccess tests
// ---------------------------------------------------------------------------

func TestHtmxSuccess(t *testing.T) {
	tests := []struct {
		name    string
		tableID string
	}{
		{"events-table", "events-table"},
		{"custom-table", "my-custom-table"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := htmxSuccess(tt.tableID)

			if result.StatusCode != http.StatusOK {
				t.Errorf("StatusCode = %d, want %d", result.StatusCode, http.StatusOK)
			}

			trigger, ok := result.Headers["HX-Trigger"]
			if !ok {
				t.Fatal("missing HX-Trigger header")
			}

			wantTrigger := `{"formSuccess":true,"refreshTable":"` + tt.tableID + `"}`
			if trigger != wantTrigger {
				t.Errorf("HX-Trigger = %q, want %q", trigger, wantTrigger)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// htmxError tests
// ---------------------------------------------------------------------------

func TestHtmxError(t *testing.T) {
	tests := []struct {
		name    string
		message string
	}{
		{"simple message", "Something went wrong"},
		{"empty message", ""},
		{"special chars", `error: "bad input" & <script>`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := htmxError(tt.message)

			if result.StatusCode != http.StatusUnprocessableEntity {
				t.Errorf("StatusCode = %d, want %d", result.StatusCode, http.StatusUnprocessableEntity)
			}

			errMsg, ok := result.Headers["HX-Error-Message"]
			if !ok {
				t.Fatal("missing HX-Error-Message header")
			}
			if errMsg != tt.message {
				t.Errorf("HX-Error-Message = %q, want %q", errMsg, tt.message)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Negative / defensive helper tests
// ---------------------------------------------------------------------------

func TestEventStatusToEnum_MissingAndInvalid(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  eventpb.EventStatus
	}{
		{"whitespace only", "   ", eventpb.EventStatus_EVENT_STATUS_UNSPECIFIED},
		{"numeric", "123", eventpb.EventStatus_EVENT_STATUS_UNSPECIFIED},
		{"partial match", "confirm", eventpb.EventStatus_EVENT_STATUS_UNSPECIFIED},
		{"with trailing space", "confirmed ", eventpb.EventStatus_EVENT_STATUS_UNSPECIFIED},
		{"null string", "null", eventpb.EventStatus_EVENT_STATUS_UNSPECIFIED},
		{"special chars", "confirmed;DROP TABLE", eventpb.EventStatus_EVENT_STATUS_UNSPECIFIED},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := eventStatusToEnum(tt.input)
			if got != tt.want {
				t.Errorf("eventStatusToEnum(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestHtmxSuccess_EmptyTableID(t *testing.T) {
	result := htmxSuccess("")

	if result.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %d, want %d", result.StatusCode, http.StatusOK)
	}

	trigger, ok := result.Headers["HX-Trigger"]
	if !ok {
		t.Fatal("missing HX-Trigger header")
	}

	wantTrigger := `{"formSuccess":true,"refreshTable":""}`
	if trigger != wantTrigger {
		t.Errorf("HX-Trigger = %q, want %q", trigger, wantTrigger)
	}
}

func TestHtmxError_VeryLongMessage(t *testing.T) {
	longMsg := ""
	for i := 0; i < 500; i++ {
		longMsg += "error "
	}

	result := htmxError(longMsg)
	if result.StatusCode != http.StatusUnprocessableEntity {
		t.Errorf("StatusCode = %d, want %d", result.StatusCode, http.StatusUnprocessableEntity)
	}

	errMsg := result.Headers["HX-Error-Message"]
	if errMsg != longMsg {
		t.Error("expected long error message to be preserved in full")
	}
}

func TestStrPtr_SpecialStrings(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"null literal", "null"},
		{"unicode", "日本語テスト"},
		{"newlines", "line1\nline2\n"},
		{"sql injection attempt", "'; DROP TABLE events; --"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ptr := strPtr(tt.input)
			if ptr == nil {
				t.Fatal("strPtr returned nil")
			}
			if *ptr != tt.input {
				t.Errorf("*strPtr(%q) = %q, want %q", tt.input, *ptr, tt.input)
			}
		})
	}
}
