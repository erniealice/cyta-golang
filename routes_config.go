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
	CalendarDataURL  string `json:"calendar_data_url"`
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
		TabActionURL:     EventTabActionURL,
		CalendarURL:      CalendarURL,
		CalendarDataURL:  CalendarDataURL,
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
		"event.bulkDelete":    r.BulkDeleteURL,
		"event.setStatus":     r.SetStatusURL,
		"event.bulkSetStatus": r.BulkSetStatusURL,
		"event.tabAction":     r.TabActionURL,
		"calendar.view":       r.CalendarURL,
		"calendar.data":       r.CalendarDataURL,
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
