// Package recurrence provides translatable label structs for the recurrence pattern entity.
package recurrence

// Labels holds all translatable strings for the recurrence pattern module.
type Labels struct {
	Page    PageLabels    `json:"page"`
	Buttons ButtonLabels  `json:"buttons"`
	Columns ColumnLabels  `json:"columns"`
	Empty   EmptyLabels   `json:"empty"`
	Form    FormLabels    `json:"form"`
	Confirm ConfirmLabels `json:"confirm"`
}

type PageLabels struct {
	Heading string `json:"heading"`
	Caption string `json:"caption"`
}

type ButtonLabels struct {
	AddPattern string `json:"addPattern"`
}

type ColumnLabels struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Frequency   string `json:"frequency"`
	Interval    string `json:"interval"`
	Rule        string `json:"rule"`
}

type EmptyLabels struct {
	Heading    string `json:"heading"`
	Subheading string `json:"subheading"`
}

type FormLabels struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Frequency   string `json:"frequency"`
	Interval    string `json:"interval"`
	Count       string `json:"count"`
	Until       string `json:"until"`
	ByDay       string `json:"byDay"`
	ByMonthDay  string `json:"byMonthDay"`
	RuleString  string `json:"ruleString"`
}

type ConfirmLabels struct {
	DeleteTitle   string `json:"deleteTitle"`
	DeleteMessage string `json:"deleteMessage"`
}
