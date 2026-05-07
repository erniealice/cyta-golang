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
	EventTabActionURL            = "/action/schedule/detail/{id}/tab/{tab}"
	EventAttachmentUploadURL     = "/action/schedule/detail/{id}/attachments/upload"
	EventAttachmentDeleteURL     = "/action/schedule/detail/{id}/attachments/delete"

	// Calendar view routes
	CalendarURL     = "/app/schedule/calendar"
	CalendarDataURL = "/action/schedule/calendar/data"

	// Schedule (event) dashboard route
	ScheduleDashboardURL = "/app/schedule/dashboard"

	// Recurrence pattern routes
	RecurrenceListURL   = "/app/schedule/recurrence/list/{status}"
	RecurrenceDetailURL = "/app/schedule/recurrence/detail/{id}"
	RecurrenceAddURL    = "/action/schedule/recurrence/add"
	RecurrenceEditURL   = "/action/schedule/recurrence/edit/{id}"
	RecurrenceDeleteURL = "/action/schedule/recurrence/delete"

	// Event tag routes
	EventTagListURL   = "/app/schedule/tags"
	EventTagDetailURL = "/app/schedule/tags/detail/{id}"
	EventTagAddURL    = "/action/schedule/tag/add"
	EventTagEditURL   = "/action/schedule/tag/edit/{id}"
	EventTagDeleteURL = "/action/schedule/tag/delete/{id}"
)
