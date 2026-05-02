// Package form holds the data + label types for the event-tag drawer form.
package form

// Labels holds i18n labels for the drawer form template.
// Flat shape consumed directly by the template (no shared LabelsFromX mapper —
// this form has a single source struct, cyta.EventTagFormLabels).
type Labels struct {
	Name                   string
	NamePlaceholder        string
	Description            string
	DescriptionPlaceholder string
	Color                  string
	ColorPlaceholder       string
	Active                 string
}

// Data is the template data for the event-tag drawer form.
type Data struct {
	FormAction   string
	IsEdit       bool
	ID           string
	Name         string
	Description  string
	Color        string
	Active       bool
	Labels       Labels
	CommonLabels any
}
