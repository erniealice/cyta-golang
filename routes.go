package cyta

// Default route constants for cyta views.
// Consumer apps can use these or define their own.
const (
	// Event (scheduling) routes
	EventListURL          = "/app/schedule/list/{status}"
	EventDetailURL        = "/app/schedule/detail/{id}"
	EventAddURL           = "/action/schedule/add"
	EventEditURL          = "/action/schedule/edit/{id}"
	EventDeleteURL        = "/action/schedule/delete"
	EventBulkDeleteURL    = "/action/schedule/bulk-delete"
	EventSetStatusURL     = "/action/schedule/set-status"
	EventBulkSetStatusURL = "/action/schedule/bulk-set-status"
	EventTabActionURL     = "/action/schedule/detail/{id}/tab/{tab}"

	// Calendar view routes
	CalendarURL     = "/app/schedule/calendar"
	CalendarDataURL = "/action/schedule/calendar/data"

	// Recurrence pattern routes
	RecurrenceListURL   = "/app/schedule/recurrence/list/{status}"
	RecurrenceDetailURL = "/app/schedule/recurrence/detail/{id}"
	RecurrenceAddURL    = "/action/schedule/recurrence/add"
	RecurrenceEditURL   = "/action/schedule/recurrence/edit/{id}"
	RecurrenceDeleteURL = "/action/schedule/recurrence/delete"
)
