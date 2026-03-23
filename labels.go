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
	Errors  EventErrorLabels   `json:"errors"`
	Status  EventStatusLabels  `json:"status"`
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
	Name        string `json:"name"`
	Description string `json:"description"`
	StartDate   string `json:"startDate"`
	EndDate     string `json:"endDate"`
	Timezone    string `json:"timezone"`
	AllDay      string `json:"allDay"`
	Organizer   string `json:"organizer"`
	Location    string `json:"location"`
	Status      string `json:"status"`
	Recurrence  string `json:"recurrence"`
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
