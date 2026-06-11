package event

// Three-level routing system for cyta views:
//
// Level 1: Generic defaults from Go consts (this file).
// Level 2: Industry-specific overrides via JSON (loaded by consumer apps).
// Level 3: App-specific overrides via Go field assignment (optional).

// Routes holds all route paths for event (scheduling) views and actions.
type Routes struct {
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
	CalendarDataURL     string `json:"calendar_data_url"`
	DashboardURL        string `json:"dashboard_url"`
}

// DefaultRoutes returns a Routes populated from package-level constants.
func DefaultRoutes() Routes {
	return Routes{
		ActiveNav:           "schedule",
		ActiveSubNav:        "schedule",
		ListURL:             ListURL,
		DetailURL:           DetailURL,
		AddURL:              AddURL,
		EditURL:             EditURL,
		DeleteURL:           DeleteURL,
		BulkDeleteURL:       BulkDeleteURL,
		SetStatusURL:        SetStatusURL,
		BulkSetStatusURL:    BulkSetStatusURL,
		TabActionURL:        TabActionURL,
		CalendarURL:         CalendarURL,
		AttachmentUploadURL: AttachmentUploadURL,
		AttachmentDeleteURL: AttachmentDeleteURL,
		CalendarDataURL:     CalendarDataURL,
		DashboardURL:        DashboardURL,
	}
}

// RouteMap returns a map[string]string for template URL resolution.
func (r Routes) RouteMap() map[string]string {
	return map[string]string{
		"event.list":              r.ListURL,
		"event.detail":            r.DetailURL,
		"event.add":               r.AddURL,
		"event.edit":              r.EditURL,
		"event.delete":            r.DeleteURL,
		"event.bulk_delete":       r.BulkDeleteURL,
		"event.set_status":        r.SetStatusURL,
		"event.bulk_set_status":   r.BulkSetStatusURL,
		"event.tab_action":        r.TabActionURL,
		"event.attachment_upload": r.AttachmentUploadURL,
		"event.attachment_delete": r.AttachmentDeleteURL,
		"calendar.view":           r.CalendarURL,
		"calendar.data":           r.CalendarDataURL,
		"event.dashboard":         r.DashboardURL,
	}
}

// Default route constants for cyta event views.
// Consumer apps can use these or define their own.
const (
	// Event (scheduling) routes
	ListURL             = "/schedule/list/{status}"
	DetailURL           = "/schedule/detail/{id}"
	AddURL              = "/action/schedule/add"
	EditURL             = "/action/schedule/edit/{id}"
	DeleteURL           = "/action/schedule/delete"
	BulkDeleteURL       = "/action/schedule/bulk-delete"
	SetStatusURL        = "/action/schedule/set-status"
	BulkSetStatusURL    = "/action/schedule/bulk-set-status"
	TabActionURL        = "/action/schedule/detail/{id}/tab/{tab}"
	AttachmentUploadURL = "/action/schedule/detail/{id}/attachments/upload"
	AttachmentDeleteURL = "/action/schedule/detail/{id}/attachments/delete"

	// Calendar view routes
	CalendarURL     = "/schedule/calendar"
	CalendarDataURL = "/action/schedule/calendar/data"

	// Schedule (event) dashboard route
	DashboardURL = "/schedule/dashboard"
)
