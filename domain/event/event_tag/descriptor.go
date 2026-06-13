package event_tag

import "github.com/erniealice/pyeza-golang/compose"

func Describe() compose.Unit {
	r := DefaultRoutes()
	l := DefaultLabels()
	return compose.Unit{
		Key:       "event.event_tag",
		Routes:    &r,
		RouteJSON: compose.JSONBinding{File: "route.json", Key: "event_tag"},
		Labels:    &l,
		LabelJSON: compose.JSONBinding{File: "event_tag.json", Key: ""},
		LabelName: "EventTagLabels",
		Templates: TemplatesFS,
	}
}
