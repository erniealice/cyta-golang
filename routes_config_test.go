package cyta

import (
	"reflect"
	"testing"
)

// ---------------------------------------------------------------------------
// DefaultEventRoutes tests
// ---------------------------------------------------------------------------

func TestDefaultEventRoutes_RouteMapKeys(t *testing.T) {
	expectedKeys := []string{
		"event.list",
		"event.detail",
		"event.add",
		"event.edit",
		"event.delete",
		"event.bulk_delete",
		"event.set_status",
		"event.bulk_set_status",
		"event.tab_action",
		"calendar.view",
		"calendar.data",
	}

	rm := DefaultEventRoutes().RouteMap()

	for _, key := range expectedKeys {
		if _, ok := rm[key]; !ok {
			t.Errorf("RouteMap missing key %q", key)
		}
	}

	if len(rm) != len(expectedKeys) {
		t.Errorf("RouteMap has %d keys, want %d", len(rm), len(expectedKeys))
	}
}

func TestDefaultEventRoutes_NoEmptyFields(t *testing.T) {
	routes := DefaultEventRoutes()
	rv := reflect.ValueOf(routes)
	rt := rv.Type()

	for i := 0; i < rv.NumField(); i++ {
		field := rv.Field(i)
		if field.Kind() == reflect.String && field.String() == "" {
			t.Errorf("EventRoutes.%s is empty", rt.Field(i).Name)
		}
	}
}

func TestDefaultEventRoutes_MatchesConstants(t *testing.T) {
	tests := []struct {
		name string
		got  string
		want string
	}{
		{"ListURL", DefaultEventRoutes().ListURL, EventListURL},
		{"DetailURL", DefaultEventRoutes().DetailURL, EventDetailURL},
		{"AddURL", DefaultEventRoutes().AddURL, EventAddURL},
		{"EditURL", DefaultEventRoutes().EditURL, EventEditURL},
		{"DeleteURL", DefaultEventRoutes().DeleteURL, EventDeleteURL},
		{"BulkDeleteURL", DefaultEventRoutes().BulkDeleteURL, EventBulkDeleteURL},
		{"SetStatusURL", DefaultEventRoutes().SetStatusURL, EventSetStatusURL},
		{"BulkSetStatusURL", DefaultEventRoutes().BulkSetStatusURL, EventBulkSetStatusURL},
		{"TabActionURL", DefaultEventRoutes().TabActionURL, EventTabActionURL},
		{"CalendarURL", DefaultEventRoutes().CalendarURL, CalendarURL},
		{"CalendarDataURL", DefaultEventRoutes().CalendarDataURL, CalendarDataURL},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("%s = %q, want %q", tt.name, tt.got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// DefaultRecurrenceRoutes tests
// ---------------------------------------------------------------------------

func TestDefaultRecurrenceRoutes_RouteMapKeys(t *testing.T) {
	expectedKeys := []string{
		"recurrence.list",
		"recurrence.detail",
		"recurrence.add",
		"recurrence.edit",
		"recurrence.delete",
	}

	rm := DefaultRecurrenceRoutes().RouteMap()

	for _, key := range expectedKeys {
		if _, ok := rm[key]; !ok {
			t.Errorf("RouteMap missing key %q", key)
		}
	}

	if len(rm) != len(expectedKeys) {
		t.Errorf("RouteMap has %d keys, want %d", len(rm), len(expectedKeys))
	}
}

func TestDefaultRecurrenceRoutes_NoEmptyFields(t *testing.T) {
	routes := DefaultRecurrenceRoutes()
	rv := reflect.ValueOf(routes)
	rt := rv.Type()

	for i := 0; i < rv.NumField(); i++ {
		field := rv.Field(i)
		if field.Kind() == reflect.String && field.String() == "" {
			t.Errorf("RecurrenceRoutes.%s is empty", rt.Field(i).Name)
		}
	}
}

func TestDefaultRecurrenceRoutes_MatchesConstants(t *testing.T) {
	tests := []struct {
		name string
		got  string
		want string
	}{
		{"ListURL", DefaultRecurrenceRoutes().ListURL, RecurrenceListURL},
		{"DetailURL", DefaultRecurrenceRoutes().DetailURL, RecurrenceDetailURL},
		{"AddURL", DefaultRecurrenceRoutes().AddURL, RecurrenceAddURL},
		{"EditURL", DefaultRecurrenceRoutes().EditURL, RecurrenceEditURL},
		{"DeleteURL", DefaultRecurrenceRoutes().DeleteURL, RecurrenceDeleteURL},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("%s = %q, want %q", tt.name, tt.got, tt.want)
			}
		})
	}
}
