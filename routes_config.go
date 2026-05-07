package cyta

// Three-level routing system for cyta views:
//
// Level 1: Generic defaults from Go consts (this file).
// Level 2: Industry-specific overrides via JSON (loaded by consumer apps).
// Level 3: App-specific overrides via Go field assignment (optional).

// EventRoutes holds all route paths for event (scheduling) views and actions.
type EventRoutes struct {
	ActiveNav    string `json:"active_nav"`
	ActiveSubNav string `json:"active_sub_nav"`

	ListURL          string `json:"list_url"`
	DetailURL        string `json:"detail_url"`
	AddURL           string `json:"add_url"`
	EditURL          string `json:"edit_url"`
	DeleteURL        string `json:"delete_url"`
	BulkDeleteURL    string `json:"bulk_delete_url"`
	SetStatusURL     string `json:"set_status_url"`
	BulkSetStatusURL string `json:"bulk_set_status_url"`
	TabActionURL     string `json:"tab_action_url"`
	CalendarURL      string `json:"calendar_url"`

	// Attachment routes
	AttachmentUploadURL string `json:"attachment_upload_url"`
	AttachmentDeleteURL string `json:"attachment_delete_url"`
	CalendarDataURL  string `json:"calendar_data_url"`
	DashboardURL     string `json:"dashboard_url"`
}

// DefaultEventRoutes returns an EventRoutes populated from package-level constants.
func DefaultEventRoutes() EventRoutes {
	return EventRoutes{
		ActiveNav:        "schedule",
		ActiveSubNav:     "schedule",
		ListURL:          EventListURL,
		DetailURL:        EventDetailURL,
		AddURL:           EventAddURL,
		EditURL:          EventEditURL,
		DeleteURL:        EventDeleteURL,
		BulkDeleteURL:    EventBulkDeleteURL,
		SetStatusURL:     EventSetStatusURL,
		BulkSetStatusURL: EventBulkSetStatusURL,
		TabActionURL:        EventTabActionURL,
		CalendarURL:         CalendarURL,
		AttachmentUploadURL: EventAttachmentUploadURL,
		AttachmentDeleteURL: EventAttachmentDeleteURL,
		CalendarDataURL:  CalendarDataURL,
		DashboardURL:     ScheduleDashboardURL,
	}
}

// RouteMap returns a map[string]string for template URL resolution.
func (r EventRoutes) RouteMap() map[string]string {
	return map[string]string{
		"event.list":          r.ListURL,
		"event.detail":        r.DetailURL,
		"event.add":           r.AddURL,
		"event.edit":          r.EditURL,
		"event.delete":        r.DeleteURL,
		"event.bulk_delete":     r.BulkDeleteURL,
		"event.set_status":      r.SetStatusURL,
		"event.bulk_set_status": r.BulkSetStatusURL,
		"event.tab_action":           r.TabActionURL,
		"event.attachment_upload":    r.AttachmentUploadURL,
		"event.attachment_delete":    r.AttachmentDeleteURL,
		"calendar.view":              r.CalendarURL,
		"calendar.data":       r.CalendarDataURL,
		"event.dashboard":     r.DashboardURL,
	}
}

// RecurrenceRoutes holds all route paths for recurrence pattern views.
type RecurrenceRoutes struct {
	ActiveNav    string `json:"active_nav"`
	ActiveSubNav string `json:"active_sub_nav"`

	ListURL   string `json:"list_url"`
	DetailURL string `json:"detail_url"`
	AddURL    string `json:"add_url"`
	EditURL   string `json:"edit_url"`
	DeleteURL string `json:"delete_url"`
}

// DefaultRecurrenceRoutes returns a RecurrenceRoutes populated from package-level constants.
func DefaultRecurrenceRoutes() RecurrenceRoutes {
	return RecurrenceRoutes{
		ActiveNav:    "schedule",
		ActiveSubNav: "recurrence-patterns",
		ListURL:      RecurrenceListURL,
		DetailURL:    RecurrenceDetailURL,
		AddURL:       RecurrenceAddURL,
		EditURL:      RecurrenceEditURL,
		DeleteURL:    RecurrenceDeleteURL,
	}
}

// RouteMap returns a map[string]string for template URL resolution.
func (r RecurrenceRoutes) RouteMap() map[string]string {
	return map[string]string{
		"recurrence.list":   r.ListURL,
		"recurrence.detail": r.DetailURL,
		"recurrence.add":    r.AddURL,
		"recurrence.edit":   r.EditURL,
		"recurrence.delete": r.DeleteURL,
	}
}

// EventTagRoutes holds all route paths for event-tag views and actions.
type EventTagRoutes struct {
	ActiveNav    string `json:"active_nav"`
	ActiveSubNav string `json:"active_sub_nav"`

	ListURL   string `json:"list_url"`
	DetailURL string `json:"detail_url"`
	AddURL    string `json:"add_url"`
	EditURL   string `json:"edit_url"`
	DeleteURL string `json:"delete_url"`
}

// DefaultEventTagRoutes returns an EventTagRoutes populated from package-level constants.
func DefaultEventTagRoutes() EventTagRoutes {
	return EventTagRoutes{
		ActiveNav:    "schedule",
		ActiveSubNav: "event-tags-active",
		ListURL:      EventTagListURL,
		DetailURL:    EventTagDetailURL,
		AddURL:       EventTagAddURL,
		EditURL:      EventTagEditURL,
		DeleteURL:    EventTagDeleteURL,
	}
}

// RouteMap returns a map[string]string for template URL resolution.
func (r EventTagRoutes) RouteMap() map[string]string {
	return map[string]string{
		"event_tag.list":   r.ListURL,
		"event_tag.detail": r.DetailURL,
		"event_tag.add":    r.AddURL,
		"event_tag.edit":   r.EditURL,
		"event_tag.delete": r.DeleteURL,
	}
}
