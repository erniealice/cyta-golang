package event_tag

// Routes holds all route paths for event-tag views and actions.
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
		ActiveSubNav: "event-tags-active",
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
		"event_tag.list":   r.ListURL,
		"event_tag.detail": r.DetailURL,
		"event_tag.add":    r.AddURL,
		"event_tag.edit":   r.EditURL,
		"event_tag.delete": r.DeleteURL,
	}
}

// Default route constants for cyta event-tag views.
// Consumer apps can use these or define their own.
const (
	ListURL   = "/schedule/tags"
	DetailURL = "/schedule/tags/detail/{id}"
	AddURL    = "/action/schedule/tag/add"
	EditURL   = "/action/schedule/tag/edit/{id}"
	DeleteURL = "/action/schedule/tag/delete/{id}"
)
