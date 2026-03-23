// Package availability implements a FHIR-inspired availability engine for
// computing free time slots for resources (staff, rooms, equipment).
//
// Design:
//   - Schedule (FHIR Schedule equivalent): defines a resource's working hours.
//   - Slot (FHIR Slot equivalent): a computed free or busy window — NOT persisted.
//   - All computation is pure domain logic; no database calls are made here.
//   - All times are UTC internally; the Timezone field is for display only.
package availability

import "time"

// Schedule defines a resource's working hours (FHIR Schedule equivalent).
// It does not persist state — it is a value object passed into the engine.
type Schedule struct {
	ResourceID   string
	ResourceType string       // "staff", "room", "equipment"
	WorkingHours []TimeWindow // e.g., Mon-Fri 9am-5pm
	Timezone     string       // IANA timezone, for display only
}

// TimeWindow describes the working hours for a given day of the week.
type TimeWindow struct {
	DayOfWeek time.Weekday
	StartHour int // 0-23
	StartMin  int // 0-59
	EndHour   int // 0-23
	EndMin    int // 0-59
}

// Slot is a computed time window (FHIR Slot equivalent — NOT persisted).
// Status is "free" when the window is unbooked, "busy" otherwise.
type Slot struct {
	Start    time.Time
	End      time.Time
	Duration time.Duration
	Status   string // "free" or "busy"
}

// Booking represents an existing time commitment for a resource.
// The engine treats any Booking as an opaque busy block.
type Booking struct {
	Start   time.Time
	End     time.Time
	EventID string // optional, for tracing back to the source event
}

// ConflictInfo describes a detected overlap between an existing Booking and
// a proposed time window.
type ConflictInfo struct {
	BookingStart time.Time
	BookingEnd   time.Time
	EventID      string
}

// AvailabilityEngine computes free slots for resources by subtracting existing
// bookings from working hours. It is stateless and safe for concurrent use.
type AvailabilityEngine struct{}

// NewAvailabilityEngine creates a new engine.
func NewAvailabilityEngine() *AvailabilityEngine {
	return &AvailabilityEngine{}
}

// ComputeSlots returns available ("free") slots for a resource within
// [windowStart, windowEnd).
//
// Algorithm:
//  1. Iterate each calendar day that falls within the query window.
//  2. For each day, find the matching TimeWindow entries (by DayOfWeek).
//  3. Generate candidate slots of exactly slotDuration within the working
//     hours, stepping by slotDuration from the working-hours start.
//  4. Discard any candidate slot that overlaps with an existing booking.
//
// All input times must be UTC. slotDuration must be positive; if it is zero
// or negative ComputeSlots returns nil.
func (e *AvailabilityEngine) ComputeSlots(
	schedule Schedule,
	existingBookings []Booking,
	windowStart, windowEnd time.Time,
	slotDuration time.Duration,
) []Slot {
	if slotDuration <= 0 {
		return nil
	}
	if !windowEnd.After(windowStart) {
		return nil
	}

	// Normalise the window to UTC so day-boundary arithmetic is consistent.
	windowStart = windowStart.UTC()
	windowEnd = windowEnd.UTC()

	// Build a lookup: weekday -> []TimeWindow for fast per-day access.
	byDay := make(map[time.Weekday][]TimeWindow, len(schedule.WorkingHours))
	for _, tw := range schedule.WorkingHours {
		byDay[tw.DayOfWeek] = append(byDay[tw.DayOfWeek], tw)
	}

	var slots []Slot

	// Walk day by day from windowStart to windowEnd.
	// dayStart always points to midnight UTC of the current day.
	dayStart := truncateToDay(windowStart)
	for !dayStart.After(windowEnd) {
		windows, ok := byDay[dayStart.Weekday()]
		if ok {
			for _, tw := range windows {
				// Build the working-hours interval for this specific date.
				whStart := time.Date(
					dayStart.Year(), dayStart.Month(), dayStart.Day(),
					tw.StartHour, tw.StartMin, 0, 0, time.UTC,
				)
				whEnd := time.Date(
					dayStart.Year(), dayStart.Month(), dayStart.Day(),
					tw.EndHour, tw.EndMin, 0, 0, time.UTC,
				)

				// Clip to the query window.
				if whStart.Before(windowStart) {
					whStart = windowStart
				}
				if whEnd.After(windowEnd) {
					whEnd = windowEnd
				}
				if !whEnd.After(whStart) {
					continue
				}

				// Generate candidate slots within the (clipped) working hours.
				// Loop condition: candidateStart.Before(whEnd) — iterate while
				// the candidate window start is still inside working hours.
				for candidateStart := whStart; candidateStart.Before(whEnd); candidateStart = candidateStart.Add(slotDuration) {
					candidateEnd := candidateStart.Add(slotDuration)
					if candidateEnd.After(whEnd) {
						// Candidate would extend past the end of working hours;
						// no more full-length slots are possible for this window.
						break
					}

					// Skip if any booking overlaps this candidate.
					if overlapsAnyBooking(existingBookings, candidateStart, candidateEnd) {
						continue
					}

					slots = append(slots, Slot{
						Start:    candidateStart,
						End:      candidateEnd,
						Duration: slotDuration,
						Status:   "free",
					})
				}
			}
		}

		dayStart = dayStart.Add(24 * time.Hour)
	}

	return slots
}

// CheckConflicts detects double-booking for a resource by comparing
// [proposedStart, proposedEnd) against all existing bookings.
// It returns every booking that overlaps the proposed window.
// Two intervals [a,b) and [c,d) overlap when a < d && c < b.
func (e *AvailabilityEngine) CheckConflicts(
	existingBookings []Booking,
	proposedStart, proposedEnd time.Time,
) []ConflictInfo {
	var conflicts []ConflictInfo
	for _, b := range existingBookings {
		if intervalsOverlap(proposedStart, proposedEnd, b.Start, b.End) {
			conflicts = append(conflicts, ConflictInfo{
				BookingStart: b.Start,
				BookingEnd:   b.End,
				EventID:      b.EventID,
			})
		}
	}
	return conflicts
}

// ---------------------------------------------------------------------------
// internal helpers
// ---------------------------------------------------------------------------

// truncateToDay returns t truncated to midnight UTC.
func truncateToDay(t time.Time) time.Time {
	t = t.UTC()
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}

// intervalsOverlap reports whether [aStart, aEnd) and [bStart, bEnd) overlap.
// Uses the half-open interval convention: aStart < bEnd && bStart < aEnd.
func intervalsOverlap(aStart, aEnd, bStart, bEnd time.Time) bool {
	return aStart.Before(bEnd) && bStart.Before(aEnd)
}

// overlapsAnyBooking reports whether [start, end) overlaps any booking in bs.
func overlapsAnyBooking(bs []Booking, start, end time.Time) bool {
	for _, b := range bs {
		if intervalsOverlap(start, end, b.Start, b.End) {
			return true
		}
	}
	return false
}
