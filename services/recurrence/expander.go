// Package recurrence expands RFC 5545 RRULE strings into concrete occurrence
// timestamps.
//
// Design:
//   - All times are UTC internally.
//   - Expansion is pure domain logic; no database calls are made here.
//   - EXDATE filtering is done by date (not full timestamp) to match RFC 5545
//     semantics: an EXDATE removes the whole occurrence regardless of time-of-day.
package recurrence

import (
	"fmt"
	"strings"
	"time"

	"github.com/teambition/rrule-go"
)

// Occurrence represents a single materialized occurrence of a recurring event.
type Occurrence struct {
	EventID          string
	StartDateTimeUTC time.Time
	EndDateTimeUTC   time.Time
	IsException      bool
	IsCancelled      bool
	WorkspaceID      string
}

// Expander expands RRULE strings into concrete occurrences.
// It is stateless and safe for concurrent use.
type Expander struct{}

// NewExpander creates a new Expander.
func NewExpander() *Expander {
	return &Expander{}
}

// ExpandRRule parses an RRULE string and returns occurrences within the given
// horizon window starting from dtStart.
//
// Parameters:
//   - rruleString: RFC 5545 RRULE (e.g. "FREQ=WEEKLY;BYDAY=MO,WE,FR;COUNT=10").
//     Must NOT include the "RRULE:" prefix.
//   - dtStart: the event's original start time (UTC).
//   - duration: how long each occurrence lasts (end = start + duration).
//   - horizon: how far into the future to expand (e.g. 2*365*24*time.Hour).
//   - exdates: excluded dates — any occurrence whose start date (UTC calendar
//     date) matches an exdate date is omitted.
//
// Returns a slice of Occurrence with computed start/end times. The returned
// Occurrence structs have empty EventID and WorkspaceID fields; the caller
// must fill those in.
func (e *Expander) ExpandRRule(
	rruleString string,
	dtStart time.Time,
	duration time.Duration,
	horizon time.Duration,
	exdates []time.Time,
) ([]Occurrence, error) {
	if rruleString == "" {
		return nil, fmt.Errorf("recurrence: rrule string is empty")
	}
	if horizon <= 0 {
		return nil, fmt.Errorf("recurrence: horizon must be positive")
	}

	// Normalise dtStart to UTC.
	dtStart = dtStart.UTC()

	// Parse the RRULE string into an ROption struct.
	// StrToRRule expects the raw rule body (no "RRULE:" prefix).
	ropt, err := rrule.StrToROption(rruleString)
	if err != nil {
		return nil, fmt.Errorf("recurrence: invalid rrule %q: %w", rruleString, err)
	}

	// Override dtStart from the rule options with the caller-supplied value so
	// that the event's actual start time is authoritative.
	ropt.Dtstart = dtStart

	r, err := rrule.NewRRule(*ropt)
	if err != nil {
		return nil, fmt.Errorf("recurrence: could not build rrule: %w", err)
	}

	// Build an exdate set keyed by UTC calendar date for O(1) lookup.
	exdateSet := make(map[string]struct{}, len(exdates))
	for _, xd := range exdates {
		exdateSet[calendarDate(xd.UTC())] = struct{}{}
	}

	horizonEnd := dtStart.Add(horizon)

	// Between returns occurrences in [after, before] inclusive when inc=true.
	// We want [dtStart, horizonEnd] inclusive.
	rawTimes := r.Between(dtStart, horizonEnd, true)

	occurrences := make([]Occurrence, 0, len(rawTimes))
	for _, t := range rawTimes {
		t = t.UTC()
		if _, excluded := exdateSet[calendarDate(t)]; excluded {
			continue
		}
		occurrences = append(occurrences, Occurrence{
			StartDateTimeUTC: t,
			EndDateTimeUTC:   t.Add(duration),
		})
	}

	return occurrences, nil
}

// ParseExdates parses a comma-separated EXDATE string into a []time.Time slice.
// Each element must be a valid RFC 3339 / ISO 8601 timestamp or date string.
// Accepted formats: "2006-01-02T15:04:05Z", "2006-01-02T15:04:05-07:00",
// "2006-01-02".
// Empty string returns nil, nil (not an error).
func ParseExdates(exdateString string) ([]time.Time, error) {
	exdateString = strings.TrimSpace(exdateString)
	if exdateString == "" {
		return nil, nil
	}

	parts := strings.Split(exdateString, ",")
	result := make([]time.Time, 0, len(parts))

	for _, raw := range parts {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			continue
		}

		t, err := parseFlexibleTime(raw)
		if err != nil {
			return nil, fmt.Errorf("recurrence: invalid exdate %q: %w", raw, err)
		}
		result = append(result, t.UTC())
	}

	return result, nil
}

// ---------------------------------------------------------------------------
// internal helpers
// ---------------------------------------------------------------------------

// calendarDate returns a string key "YYYY-MM-DD" for the given UTC time.
func calendarDate(t time.Time) string {
	y, m, d := t.Date()
	return fmt.Sprintf("%04d-%02d-%02d", y, int(m), d)
}

// parseFlexibleTime attempts to parse t as RFC 3339 (with or without offset),
// or as a plain date "2006-01-02".
func parseFlexibleTime(s string) (time.Time, error) {
	formats := []string{
		time.RFC3339,
		"2006-01-02T15:04:05Z07:00",
		"2006-01-02T15:04:05",
		"2006-01-02",
	}
	for _, f := range formats {
		if t, err := time.Parse(f, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("unrecognised time format")
}
