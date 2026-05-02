// Package cyta provides translatable label structs for the scheduling/events domain.
//
// This file contains all label types for events, calendar, and recurrence patterns.
// Labels are loaded from lyngua translation files and injected into views at startup.
package cyta

// ---------------------------------------------------------------------------
// Event labels
// ---------------------------------------------------------------------------

// EventLabels holds all translatable strings for the event (schedule) module.
type EventLabels struct {
	Page    EventPageLabels    `json:"page"`
	Buttons EventButtonLabels  `json:"buttons"`
	Columns EventColumnLabels  `json:"columns"`
	Empty   EventEmptyLabels   `json:"empty"`
	Form    EventFormLabels    `json:"form"`
	Actions EventActionLabels  `json:"actions"`
	Detail  EventDetailLabels  `json:"detail"`
	Tabs    EventTabLabels     `json:"tabs"`
	Confirm EventConfirmLabels `json:"confirm"`
	Errors    EventErrorLabels           `json:"errors"`
	Status    EventStatusLabels          `json:"status"`
	Calendar  EventCalendarPopoverLabels `json:"calendar"`
	Dashboard ScheduleDashboardLabels    `json:"dashboard"`
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

// EventCalendarPopoverLabels holds strings rendered inside the month/week/day
// cell popover (Phase 3 of event-management epic). One popover per cell;
// "View day" preserves existing day-view nav, "New event at HH:MM" opens
// the event drawer pre-seeded with that time.
type EventCalendarPopoverLabels struct {
	ViewDay     string `json:"viewDay"`
	NewEventAt  string `json:"newEventAt"`  // template: "New event at {{time}}"
	NewEvent    string `json:"newEvent"`    // fallback when no time available
	MoreEvents  string `json:"moreEvents"`  // "+N more events"
}

type EventPageLabels struct {
	Heading          string `json:"heading"`
	HeadingUpcoming  string `json:"headingUpcoming"`
	HeadingConfirmed string `json:"headingConfirmed"`
	HeadingCompleted string `json:"headingCompleted"`
	HeadingCancelled string `json:"headingCancelled"`
	Caption          string `json:"caption"`
}

type EventButtonLabels struct {
	AddEvent string `json:"addEvent"`
}

type EventColumnLabels struct {
	Name      string `json:"name"`
	StartDate string `json:"startDate"`
	EndDate   string `json:"endDate"`
	Organizer string `json:"organizer"`
	Location  string `json:"location"`
	Status    string `json:"status"`
	Recurs    string `json:"recurs"`
}

type EventEmptyLabels struct {
	Heading    string `json:"heading"`
	Subheading string `json:"subheading"`
}

type EventFormLabels struct {
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

type EventActionLabels struct {
	Edit      string `json:"edit"`
	Delete    string `json:"delete"`
	Cancel    string `json:"cancel"`
	Confirm   string `json:"confirm"`
	Duplicate string `json:"duplicate"`
}

type EventDetailLabels struct {
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

type EventTabLabels struct {
	Overview    string `json:"overview"`
	Attendees   string `json:"attendees"`
	Resources   string `json:"resources"`
	Products    string `json:"products"`
	Occurrences string `json:"occurrences"`
}

type EventConfirmLabels struct {
	DeleteTitle   string `json:"deleteTitle"`
	DeleteMessage string `json:"deleteMessage"`
	CancelTitle   string `json:"cancelTitle"`
	CancelMessage string `json:"cancelMessage"`
}

type EventErrorLabels struct {
	NameRequired      string `json:"nameRequired"`
	StartDateRequired string `json:"startDateRequired"`
	EndDateRequired   string `json:"endDateRequired"`
	InvalidDateRange  string `json:"invalidDateRange"`
}

type EventStatusLabels struct {
	Tentative string `json:"tentative"`
	Confirmed string `json:"confirmed"`
	Cancelled string `json:"cancelled"`
}

// ---------------------------------------------------------------------------
// EventTag labels
// ---------------------------------------------------------------------------

// EventTagLabels holds all translatable strings for the event-tag module.
// Mirrors RoleLabels (entydad-golang) but omits the Detail sub-struct —
// tags only have a list + drawer form, no dedicated detail page.
type EventTagLabels struct {
	Page    EventTagPageLabels   `json:"page"`
	Buttons EventTagButtonLabels `json:"buttons"`
	Columns EventTagColumnLabels `json:"columns"`
	Empty   EventTagEmptyLabels  `json:"empty"`
	Form    EventTagFormLabels   `json:"form"`
	Actions EventTagActionLabels `json:"actions"`
}

type EventTagPageLabels struct {
	Heading         string `json:"heading"`
	HeadingActive   string `json:"headingActive"`
	HeadingInactive string `json:"headingInactive"`
	Caption         string `json:"caption"`
	CaptionActive   string `json:"captionActive"`
	CaptionInactive string `json:"captionInactive"`
}

type EventTagButtonLabels struct {
	AddTag string `json:"addTag"`
}

type EventTagColumnLabels struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Color       string `json:"color"`
	Status      string `json:"status"`
	DateCreated string `json:"dateCreated"`
}

type EventTagEmptyLabels struct {
	ActiveTitle     string `json:"activeTitle"`
	ActiveMessage   string `json:"activeMessage"`
	InactiveTitle   string `json:"inactiveTitle"`
	InactiveMessage string `json:"inactiveMessage"`
}

type EventTagFormLabels struct {
	Name                   string `json:"name"`
	NamePlaceholder        string `json:"namePlaceholder"`
	Description            string `json:"description"`
	DescriptionPlaceholder string `json:"descriptionPlaceholder"`
	Color                  string `json:"color"`
	ColorPlaceholder       string `json:"colorPlaceholder"`
	Active                 string `json:"active"`
}

type EventTagActionLabels struct {
	View       string `json:"view"`
	Edit       string `json:"edit"`
	Delete     string `json:"delete"`
	Activate   string `json:"activate"`
	Deactivate string `json:"deactivate"`
}

// ---------------------------------------------------------------------------
// Recurrence labels
// ---------------------------------------------------------------------------

// RecurrenceLabels holds all translatable strings for the recurrence pattern module.
type RecurrenceLabels struct {
	Page    RecurrencePageLabels    `json:"page"`
	Buttons RecurrenceButtonLabels  `json:"buttons"`
	Columns RecurrenceColumnLabels  `json:"columns"`
	Empty   RecurrenceEmptyLabels   `json:"empty"`
	Form    RecurrenceFormLabels    `json:"form"`
	Confirm RecurrenceConfirmLabels `json:"confirm"`
}

type RecurrencePageLabels struct {
	Heading string `json:"heading"`
	Caption string `json:"caption"`
}

type RecurrenceButtonLabels struct {
	AddPattern string `json:"addPattern"`
}

type RecurrenceColumnLabels struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Frequency   string `json:"frequency"`
	Interval    string `json:"interval"`
	Rule        string `json:"rule"`
}

type RecurrenceEmptyLabels struct {
	Heading    string `json:"heading"`
	Subheading string `json:"subheading"`
}

type RecurrenceFormLabels struct {
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

type RecurrenceConfirmLabels struct {
	DeleteTitle   string `json:"deleteTitle"`
	DeleteMessage string `json:"deleteMessage"`
}

// ---------------------------------------------------------------------------
// Calendar labels
// ---------------------------------------------------------------------------

// CalendarLabels holds all translatable strings for the calendar view.
type CalendarLabels struct {
	Page CalendarPageLabels `json:"page"`
}

type CalendarPageLabels struct {
	Heading string `json:"heading"`
	Caption string `json:"caption"`
}
