// Package event_tag provides translatable label structs for the event-tag entity.
package event_tag

// Labels holds all translatable strings for the event-tag module.
// Mirrors RoleLabels (entydad-golang) but omits the Detail sub-struct —
// tags only have a list + drawer form, no dedicated detail page.
type Labels struct {
	Page    PageLabels   `json:"page"`
	Buttons ButtonLabels `json:"buttons"`
	Columns ColumnLabels `json:"columns"`
	Empty   EmptyLabels  `json:"empty"`
	Form    FormLabels   `json:"form"`
	Actions ActionLabels `json:"actions"`
}

type PageLabels struct {
	Heading         string `json:"heading"`
	HeadingActive   string `json:"headingActive"`
	HeadingInactive string `json:"headingInactive"`
	Caption         string `json:"caption"`
	CaptionActive   string `json:"captionActive"`
	CaptionInactive string `json:"captionInactive"`
}

type ButtonLabels struct {
	AddTag string `json:"addTag"`
}

type ColumnLabels struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Color       string `json:"color"`
	Status      string `json:"status"`
	DateCreated string `json:"dateCreated"`
}

type EmptyLabels struct {
	ActiveTitle     string `json:"activeTitle"`
	ActiveMessage   string `json:"activeMessage"`
	InactiveTitle   string `json:"inactiveTitle"`
	InactiveMessage string `json:"inactiveMessage"`
}

type FormLabels struct {
	Name                   string `json:"name"`
	NamePlaceholder        string `json:"namePlaceholder"`
	Description            string `json:"description"`
	DescriptionPlaceholder string `json:"descriptionPlaceholder"`
	Color                  string `json:"color"`
	ColorPlaceholder       string `json:"colorPlaceholder"`
	Active                 string `json:"active"`
}

type ActionLabels struct {
	View       string `json:"view"`
	Edit       string `json:"edit"`
	Delete     string `json:"delete"`
	Activate   string `json:"activate"`
	Deactivate string `json:"deactivate"`
}
