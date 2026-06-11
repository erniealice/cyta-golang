package recurrence

// Routes holds all route paths for recurrence pattern views.
type Routes struct {
	ActiveNav    string `json:"active_nav"`
	ActiveSubNav string `json:"active_sub_nav"`

	ListURL   string `json:"list_url"`
	DetailURL string `json:"detail_url"`
	AddURL    string `json:"add_url"`
	EditURL   string `json:"edit_url"`
	DeleteURL string `json:"delete_url"`
}

// DefaultRoutes returns a Routes populated from package-level constants.
func DefaultRoutes() Routes {
	return Routes{
		ActiveNav:    "schedule",
		ActiveSubNav: "recurrence-patterns",
		ListURL:      ListURL,
		DetailURL:    DetailURL,
		AddURL:       AddURL,
		EditURL:      EditURL,
		DeleteURL:    DeleteURL,
	}
}

// RouteMap returns a map[string]string for template URL resolution.
func (r Routes) RouteMap() map[string]string {
	return map[string]string{
		"recurrence.list":   r.ListURL,
		"recurrence.detail": r.DetailURL,
		"recurrence.add":    r.AddURL,
		"recurrence.edit":   r.EditURL,
		"recurrence.delete": r.DeleteURL,
	}
}

// Default route constants for cyta recurrence pattern views.
// Consumer apps can use these or define their own.
const (
	ListURL   = "/schedule/recurrence/list/{status}"
	DetailURL = "/schedule/recurrence/detail/{id}"
	AddURL    = "/action/schedule/recurrence/add"
	EditURL   = "/action/schedule/recurrence/edit/{id}"
	DeleteURL = "/action/schedule/recurrence/delete"
)
