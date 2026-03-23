package availability

import "time"

// ConflictChecker wraps conflict-detection logic bound to a fixed set of
// existing bookings. It is a convenience layer for call-sites that need to
// validate multiple proposed windows against the same booking slice without
// repeatedly passing the slice on every call.
//
// Example usage:
//
//	checker := NewConflictChecker(engine, existingBookings)
//	if checker.HasConflict(proposedStart, proposedEnd) {
//	    // reject the proposed booking
//	}
type ConflictChecker struct {
	engine   *AvailabilityEngine
	bookings []Booking
}

// NewConflictChecker creates a ConflictChecker bound to the supplied booking
// slice. The slice is not copied; callers must not mutate it after passing.
func NewConflictChecker(engine *AvailabilityEngine, bookings []Booking) *ConflictChecker {
	return &ConflictChecker{engine: engine, bookings: bookings}
}

// HasConflict reports whether [proposedStart, proposedEnd) overlaps any
// existing booking. It is a zero-allocation convenience wrapper when the
// caller only needs a boolean answer.
func (c *ConflictChecker) HasConflict(proposedStart, proposedEnd time.Time) bool {
	for _, b := range c.bookings {
		if intervalsOverlap(proposedStart, proposedEnd, b.Start, b.End) {
			return true
		}
	}
	return false
}

// ConflictsForWindow returns every booking that overlaps [proposedStart,
// proposedEnd). Delegates to AvailabilityEngine.CheckConflicts with the
// bound booking slice.
func (c *ConflictChecker) ConflictsForWindow(proposedStart, proposedEnd time.Time) []ConflictInfo {
	return c.engine.CheckConflicts(c.bookings, proposedStart, proposedEnd)
}

// ConflictsForSlots checks each slot in the provided slice and returns a map
// of slot-index → conflicts for any slot that has at least one conflict.
// Slots with no conflicts are omitted from the map.
// This is useful for batch-validating a set of proposed slots in one pass.
func (c *ConflictChecker) ConflictsForSlots(slots []Slot) map[int][]ConflictInfo {
	result := make(map[int][]ConflictInfo)
	for i, s := range slots {
		if conflicts := c.engine.CheckConflicts(c.bookings, s.Start, s.End); len(conflicts) > 0 {
			result[i] = conflicts
		}
	}
	if len(result) == 0 {
		return nil
	}
	return result
}
