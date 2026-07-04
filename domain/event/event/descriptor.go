package event

import "github.com/erniealice/espyna-golang/consumer/compose"

func Describe() compose.Unit {
	r := DefaultRoutes()
	l := DefaultLabels()
	return compose.Unit{
		Key:       "event.event",
		Routes:    &r,
		RouteJSON: compose.JSONBinding{File: "route.json", Key: "event"},
		Labels:    &l,
		LabelJSON: compose.JSONBinding{File: "event.json", Key: "event"},
		LabelName: "EventLabels",
		Templates: TemplatesFS,
		Nav: compose.NavContrib{
			Permission: "event:list",
			AppEntry: &compose.AppEntry{
				Key:        "schedule",
				Route:      "event.calendar",
				Label:      "Schedule",
				Icon:       "icon-calendar",
				Permission: "event:list",
			},
			Items: []compose.NavItem{
				{Key: "calendar", Route: "calendar.view", Label: "Calendar", Icon: "icon-calendar", Permission: "event:list", LabelKey: "schedule_calendar_label", IconKey: "schedule_calendar_icon"},
				{Key: "events-upcoming", Route: "event.list", Params: map[string]string{"status": "upcoming"}, Label: "Upcoming", Icon: "icon-clock", Permission: "event:list", LabelKey: "schedule_upcoming_label", IconKey: "schedule_upcoming_icon"},
				{Key: "events-confirmed", Route: "event.list", Params: map[string]string{"status": "confirmed"}, Label: "Confirmed", Icon: "icon-check-circle", Permission: "event:list", LabelKey: "schedule_confirmed_label", IconKey: "schedule_confirmed_icon"},
				{Key: "events-completed", Route: "event.list", Params: map[string]string{"status": "completed"}, Label: "Completed", Icon: "icon-check-square", Permission: "event:list", LabelKey: "schedule_completed_label", IconKey: "schedule_completed_icon"},
				{Key: "events-cancelled", Route: "event.list", Params: map[string]string{"status": "cancelled"}, Label: "Cancelled", Icon: "icon-x-circle", Permission: "event:list", LabelKey: "schedule_cancelled_label", IconKey: "schedule_cancelled_icon"},
				// NOTE: "recurrence-patterns" nav item removed — the recurrence
				// entity (domain/event/recurrence) is scaffold-only (labels.go +
				// routes.go, no view/module/descriptor), so it is not mounted in
				// block.AllUnits and "recurrence.list" is never in the route table.
				// A dangling nav ref fail-closes the entire cyta engine assembly
				// (compose phase-3). Restore this item once recurrence ships a view
				// module and is added to block.AllUnits.
			},
		},
	}
}
