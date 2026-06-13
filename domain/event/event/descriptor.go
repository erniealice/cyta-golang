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
	}
}
