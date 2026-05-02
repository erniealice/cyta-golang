// Package form holds the data + label types and mappers for the event drawer
// form. Modelled on the centymo price-plan form package — the Data type is
// what the drawer template consumes, and the Labels type is the flat shape
// every template field reads from.
package form

import (
	pyeza "github.com/erniealice/pyeza-golang"
)

// Option is a generic select-option pair (status, tag, attendee).
// Selected is set when rendering an existing event in edit mode.
type Option struct {
	Value    string
	Label    string
	Selected bool
}

// SelectedOption is the {Value, Label} pair the multi-select component
// uses to seed pre-selected chips in edit mode.
type SelectedOption struct {
	Value string
	Label string
}

// Data is the template data passed into event-drawer-form.html. Field names
// align to the form-control names used by handlers_save.go:
//   name, all_day, start_date, start_time, end_date, end_time,
//   notes, status, invitees (csv), tag_ids (csv).
type Data struct {
	// Form metadata
	FormAction string
	IsEdit     bool
	ID         string

	// Field values (always strings in template land — Go zero-value renders empty)
	Name      string
	Notes     string
	StartDate string // YYYY-MM-DD
	StartTime string // HH:MM (24h)
	EndDate   string // YYYY-MM-DD
	EndTime   string // HH:MM (24h)
	Timezone  string
	AllDay    bool

	// Status select options
	StatusOptions []Option

	// Tag multi-picker (workspace-scoped tag list with selected state)
	TagOptions      []Option
	SelectedTags    []SelectedOption

	// Attendee multi-picker (workspace_user + client union)
	AttendeeOptions     []Option
	SelectedAttendees   []SelectedOption

	// Labels (flat — template reads .Labels.* directly)
	Labels       Labels
	CommonLabels pyeza.CommonLabels

	// Phase 5 attachments — list rendered for edit mode; empty for add.
	// Each entry uses pyeza's existing attachment list shape.
	Attachments []Attachment
}

// Attachment mirrors what packages/hybra-golang renders today; kept local
// to avoid an import cycle. Populated by handlers_save.go in edit mode
// from document.Attachment with module_key="event", foreign_key=event.id.
type Attachment struct {
	ID       string
	Filename string
	SizeKB   int64
	MimeType string
	URL      string // download/preview URL
}

// Labels is the flat template-facing label shape. Every drawer label the
// template reads goes through here. Keep the field set as a strict superset
// of EventFormLabels so LabelsFromEvent fills it cleanly.
type Labels struct {
	// Identity
	NameLabel       string
	NamePlaceholder string

	// Time
	AllDayLabel    string
	StartDateLabel string
	StartTimeLabel string
	EndDateLabel   string
	EndTimeLabel   string
	TimezoneLabel  string

	// Status
	StatusLabel string

	// Multi-pickers
	InviteesLabel       string
	InviteesPlaceholder string
	TagsLabel           string
	TagsPlaceholder     string

	// Notes (maps to event.description)
	NotesLabel       string
	NotesPlaceholder string

	// Attachments (Phase 5)
	AttachmentsLabel string
	AttachmentsHint  string
}

