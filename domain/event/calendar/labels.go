// Package eventcalendar provides translatable label structs for the calendar view.
// Package name is disambiguated (eventcalendar not calendar) because this is
// a domain-level view, not an esqyma entity.
package eventcalendar

// Labels holds all translatable strings for the calendar view.
type Labels struct {
	Page PageLabels `json:"page"`
}

type PageLabels struct {
	Heading string `json:"heading"`
	Caption string `json:"caption"`
}
