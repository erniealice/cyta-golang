package form

import (
	cyta "github.com/erniealice/cyta-golang"
	eventpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/event/event"
)

// BuildStatusOptions returns the EventStatus enum as a select-option list,
// labels drawn from EventStatusLabels.
func BuildStatusOptions(labels cyta.EventStatusLabels, current eventpb.EventStatus) []Option {
	rows := []struct {
		val  string
		enum eventpb.EventStatus
		txt  string
	}{
		{"tentative", eventpb.EventStatus_EVENT_STATUS_TENTATIVE, labels.Tentative},
		{"confirmed", eventpb.EventStatus_EVENT_STATUS_CONFIRMED, labels.Confirmed},
		{"cancelled", eventpb.EventStatus_EVENT_STATUS_CANCELLED, labels.Cancelled},
	}
	out := make([]Option, len(rows))
	for i, r := range rows {
		out[i] = Option{Value: r.val, Label: r.txt, Selected: current == r.enum}
	}
	return out
}

// StatusFromString parses a form-posted status value into the proto enum.
// Defaults to TENTATIVE for empty / unknown input.
func StatusFromString(s string) eventpb.EventStatus {
	switch s {
	case "confirmed":
		return eventpb.EventStatus_EVENT_STATUS_CONFIRMED
	case "cancelled":
		return eventpb.EventStatus_EVENT_STATUS_CANCELLED
	default:
		return eventpb.EventStatus_EVENT_STATUS_TENTATIVE
	}
}
