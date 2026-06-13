package event

import "github.com/erniealice/pyeza-golang/compose"

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
				{Key: "calendar", Route: "calendar.view", Label: "Calendar", Icon: "icon-calendar", Permission: "event:list"},
				{Key: "events-upcoming", Route: "event.list", Params: map[string]string{"status": "upcoming"}, Label: "Upcoming", Icon: "icon-clock", Permission: "event:list"},
				{Key: "events-confirmed", Route: "event.list", Params: map[string]string{"status": "confirmed"}, Label: "Confirmed", Icon: "icon-check-circle", Permission: "event:list"},
				{Key: "events-completed", Route: "event.list", Params: map[string]string{"status": "completed"}, Label: "Completed", Icon: "icon-check-square", Permission: "event:list"},
				{Key: "events-cancelled", Route: "event.list", Params: map[string]string{"status": "cancelled"}, Label: "Cancelled", Icon: "icon-x-circle", Permission: "event:list"},
				{Key: "recurrence-patterns", Route: "recurrence.list", Params: map[string]string{"status": "active"}, Label: "Recurrence", Icon: "icon-repeat", Permission: "event:list"},
			},
		},
	}
}
