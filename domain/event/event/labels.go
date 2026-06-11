// Package event provides translatable label structs for the event (scheduling) entity.
package event

// Labels holds all translatable strings for the event (schedule) module.
type Labels struct {
	Page      PageLabels              `json:"page"`
	Buttons   ButtonLabels            `json:"buttons"`
	Columns   ColumnLabels            `json:"columns"`
	Empty     EmptyLabels             `json:"empty"`
	Form      FormLabels              `json:"form"`
	Actions   ActionLabels            `json:"actions"`
	Detail    DetailLabels            `json:"detail"`
	Tabs      TabLabels               `json:"tabs"`
	Confirm   ConfirmLabels           `json:"confirm"`
	Errors    ErrorLabels             `json:"errors"`
	Status    StatusLabels            `json:"status"`
	Calendar  CalendarPopoverLabels   `json:"calendar"`
	Dashboard ScheduleDashboardLabels `json:"dashboard"`
}

// ScheduleDashboardLabels holds translatable strings for the Schedule live
// dashboard (Phase 6 — Pyeza dashboard block + per-app live dashboards plan).
// Sidebar/active-nav key is "schedule" so the struct name reflects that even
// though the underlying domain is "event".
type ScheduleDashboardLabels struct {
	Title    string `json:"title"`
	Subtitle string `json:"subtitle"`
	// Stats
	StatToday       string `json:"statToday"`
	StatThisWeek    string `json:"statThisWeek"`
	StatByTag       string `json:"statByTag"`
	StatUtilization string `json:"statUtilization"`
	// Widgets
	WidgetByDay    string `json:"widgetByDay"`
	WidgetByTag    string `json:"widgetByTag"`
	WidgetUpcoming string `json:"widgetUpcoming"`
	// Quick actions
	QuickNew        string `json:"quickNew"`
	QuickCalendar   string `json:"quickCalendar"`
	QuickTags       string `json:"quickTags"`
	QuickRecurrence string `json:"quickRecurrence"`
	// Common
	ViewAll       string `json:"viewAll"`
	EmptyUpcoming string `json:"emptyUpcoming"`
	EmptyByTag    string `json:"emptyByTag"`
}

// CalendarPopoverLabels holds strings rendered inside the month/week/day
// cell popover (Phase 3 of event-management epic). One popover per cell;
// "View day" preserves existing day-view nav, "New event at HH:MM" opens
// the event drawer pre-seeded with that time.
type CalendarPopoverLabels struct {
	ViewDay    string `json:"viewDay"`
	NewEventAt string `json:"newEventAt"` // template: "New event at {{time}}"
	NewEvent   string `json:"newEvent"`   // fallback when no time available
	MoreEvents string `json:"moreEvents"` // "+N more events"
}

type PageLabels struct {
	Heading          string `json:"heading"`
	HeadingUpcoming  string `json:"headingUpcoming"`
	HeadingConfirmed string `json:"headingConfirmed"`
	HeadingCompleted string `json:"headingCompleted"`
	HeadingCancelled string `json:"headingCancelled"`
	Caption          string `json:"caption"`
}

type ButtonLabels struct {
	AddEvent string `json:"addEvent"`
}

type ColumnLabels struct {
	Name      string `json:"name"`
	StartDate string `json:"startDate"`
	EndDate   string `json:"endDate"`
	Organizer string `json:"organizer"`
	Location  string `json:"location"`
	Status    string `json:"status"`
	Recurs    string `json:"recurs"`
}

type EmptyLabels struct {
	Heading    string `json:"heading"`
	Subheading string `json:"subheading"`
}

type FormLabels struct {
	Name                string `json:"name"`
	NamePlaceholder     string `json:"namePlaceholder"`
	Description         string `json:"description"`
	StartDate           string `json:"startDate"`
	EndDate             string `json:"endDate"`
	StartTime           string `json:"startTime"`
	EndTime             string `json:"endTime"`
	Timezone            string `json:"timezone"`
	AllDay              string `json:"allDay"`
	Organizer           string `json:"organizer"`
	Location            string `json:"location"`
	Status              string `json:"status"`
	Recurrence          string `json:"recurrence"`
	Notes               string `json:"notes"`
	NotesPlaceholder    string `json:"notesPlaceholder"`
	Invitees            string `json:"invitees"`
	InviteesPlaceholder string `json:"inviteesPlaceholder"`
	Tags                string `json:"tags"`
	TagsPlaceholder     string `json:"tagsPlaceholder"`
	Attachments         string `json:"attachments"`
	AttachmentsHint     string `json:"attachmentsHint"`
}

type ActionLabels struct {
	Edit      string `json:"edit"`
	Delete    string `json:"delete"`
	Cancel    string `json:"cancel"`
	Confirm   string `json:"confirm"`
	Duplicate string `json:"duplicate"`
}

type DetailLabels struct {
	Heading     string `json:"heading"`
	Overview    string `json:"overview"`
	Organizer   string `json:"organizer"`
	Location    string `json:"location"`
	Duration    string `json:"duration"`
	TimeRange   string `json:"timeRange"`
	Status      string `json:"status"`
	AllDay      string `json:"allDay"`
	Recurrence  string `json:"recurrence"`
	Description string `json:"description"`
}

type TabLabels struct {
	Overview    string `json:"overview"`
	Attendees   string `json:"attendees"`
	Resources   string `json:"resources"`
	Products    string `json:"products"`
	Occurrences string `json:"occurrences"`
	Attachments string `json:"tabAttachments"`
}

type ConfirmLabels struct {
	DeleteTitle   string `json:"deleteTitle"`
	DeleteMessage string `json:"deleteMessage"`
	CancelTitle   string `json:"cancelTitle"`
	CancelMessage string `json:"cancelMessage"`
}

type ErrorLabels struct {
	NameRequired      string `json:"nameRequired"`
	StartDateRequired string `json:"startDateRequired"`
	EndDateRequired   string `json:"endDateRequired"`
	InvalidDateRange  string `json:"invalidDateRange"`
}

type StatusLabels struct {
	Tentative string `json:"tentative"`
	Confirmed string `json:"confirmed"`
	Cancelled string `json:"cancelled"`
}
