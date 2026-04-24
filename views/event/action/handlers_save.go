package action

import (
	"context"
	"log"
	"net/http"
	"strings"
	"time"

	cytaeventform "github.com/erniealice/cyta-golang/views/event/form"
	"github.com/erniealice/pyeza-golang/route"
	"github.com/erniealice/pyeza-golang/view"

	eventpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/event/event"
)

// handleSave is the unified POST handler for both Add and Edit. It:
//   1. Parses the multipart form body
//   2. Creates or updates the event
//   3. Sets tag assignments via SetEventTagAssignments (atomic replace)
//   4. Syncs attendees via SyncEventAttendees
//   5. Returns an HX-Redirect to the event detail page (or HX-Trigger for in-place refresh)
//
// existingID = "" for Add (new event), or the event ID for Edit.
//
// Multipart attachment files (Phase 5) are accepted in the same body; they
// are uploaded by the caller after the event ID is known. For now this handler
// reads the file headers but defers actual upload to a Phase 5 follow-up that
// will plug into the document.Attachment use case.
func handleSave(ctx context.Context, viewCtx *view.ViewContext, deps *Deps, existingID string) view.ViewResult {
	r := viewCtx.Request

	// Multipart accepts up to 32 MB of form envelope (matches hybra attachment handler).
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		// Fallback to plain form parse for non-multipart requests.
		if err := r.ParseForm(); err != nil {
			return htmxError("Invalid form data")
		}
	}

	name := strings.TrimSpace(r.FormValue("name"))
	if name == "" {
		return htmxError(deps.Labels.Errors.NameRequired)
	}

	allDay := r.FormValue("all_day") == "true"
	startMillis, endMillis, err := parseDateRange(r.FormValue("start_date"), r.FormValue("start_time"), r.FormValue("end_date"), r.FormValue("end_time"), allDay)
	if err != nil {
		return htmxError(err.Error())
	}

	notes := strings.TrimSpace(r.FormValue("notes"))
	timezone := strings.TrimSpace(r.FormValue("timezone"))
	if timezone == "" {
		timezone = "UTC"
	}
	statusEnum := cytaeventform.StatusFromString(r.FormValue("status"))

	eventID := existingID

	if existingID == "" {
		resp, err := deps.CreateEvent(ctx, &eventpb.CreateEventRequest{
			Data: &eventpb.Event{
				Name:             name,
				Description:      strPtr(notes),
				StartDateTimeUtc: startMillis,
				EndDateTimeUtc:   endMillis,
				Timezone:         timezone,
				AllDay:           allDay,
				Status:           statusEnum,
			},
		})
		if err != nil {
			log.Printf("Failed to create event: %v", err)
			return htmxError(err.Error())
		}
		respData := resp.GetData()
		if len(respData) > 0 {
			eventID = respData[0].GetId()
		}
	} else {
		_, err := deps.UpdateEvent(ctx, &eventpb.UpdateEventRequest{
			Data: &eventpb.Event{
				Id:               existingID,
				Name:             name,
				Description:      strPtr(notes),
				StartDateTimeUtc: startMillis,
				EndDateTimeUtc:   endMillis,
				Timezone:         timezone,
				AllDay:           allDay,
				Status:           statusEnum,
			},
		})
		if err != nil {
			log.Printf("Failed to update event %s: %v", existingID, err)
			return htmxError(err.Error())
		}
	}

	if eventID == "" {
		// Without an ID we can't wire dependents; return a success-only header so the
		// drawer closes and the calendar can refresh.
		return htmxRefresh()
	}

	// Tags — atomic replace
	if deps.SetEventTagAssignments != nil {
		tagIDs := splitCSV(r.FormValue("tag_ids"))
		if err := deps.SetEventTagAssignments(ctx, eventID, tagIDs); err != nil {
			log.Printf("SetEventTagAssignments(%s) failed: %v", eventID, err)
			// Soft-fail: event saved; tags didn't. Surface via header but don't 422.
		}
	}

	// Attendees — sync (ref scheme: "user:<id>" or "client:<id>")
	if deps.SyncEventAttendees != nil {
		attendees := splitCSV(r.FormValue("invitees"))
		if err := deps.SyncEventAttendees(ctx, eventID, attendees); err != nil {
			log.Printf("SyncEventAttendees(%s) failed: %v", eventID, err)
		}
	}

	// Phase 5 attachments — uploaded after event ID is known.
	// Multipart files are at r.MultipartForm.File["attachment_files"]; the
	// hybra attachment handler will be wired through a callback in a future
	// patch. For now, log the count so we can confirm files reach the server.
	if r.MultipartForm != nil {
		if files := r.MultipartForm.File["attachment_files"]; len(files) > 0 {
			log.Printf("Event %s: %d attachment file(s) staged (upload wiring pending Phase 5)", eventID, len(files))
		}
	}

	// On Add, redirect to the new detail page; on Edit, refresh in place.
	if existingID == "" {
		return view.ViewResult{
			StatusCode: http.StatusOK,
			Headers: map[string]string{
				"HX-Trigger":  `{"formSuccess":true,"refreshCalendar":true}`,
				"HX-Redirect": route.ResolveURL(deps.Routes.DetailURL, "id", eventID),
			},
		}
	}
	return htmxRefresh()
}

// htmxRefresh closes the sheet and signals the calendar/list to re-fetch.
func htmxRefresh() view.ViewResult {
	return view.ViewResult{
		StatusCode: http.StatusOK,
		Headers: map[string]string{
			"HX-Trigger": `{"formSuccess":true,"refreshCalendar":true,"refreshTable":"events-table"}`,
		},
	}
}

// parseDateRange merges date+time strings into UTC unix-millis.
// allDay collapses both endpoints to date-only at 00:00 / 23:59.
func parseDateRange(startDate, startTime, endDate, endTime string, allDay bool) (int64, int64, error) {
	if startDate == "" {
		return 0, 0, errFormStartRequired
	}
	if endDate == "" {
		endDate = startDate
	}
	loc := time.UTC
	if allDay {
		s, err := time.ParseInLocation("2006-01-02", startDate, loc)
		if err != nil {
			return 0, 0, err
		}
		e, err := time.ParseInLocation("2006-01-02", endDate, loc)
		if err != nil {
			return 0, 0, err
		}
		// All-day: span the full UTC days.
		eEnd := e.Add(24*time.Hour - 1*time.Minute)
		return s.UnixMilli(), eEnd.UnixMilli(), nil
	}
	if startTime == "" {
		startTime = "00:00"
	}
	if endTime == "" {
		endTime = startTime
	}
	s, err := time.ParseInLocation("2006-01-02 15:04", startDate+" "+startTime, loc)
	if err != nil {
		return 0, 0, err
	}
	e, err := time.ParseInLocation("2006-01-02 15:04", endDate+" "+endTime, loc)
	if err != nil {
		return 0, 0, err
	}
	if e.Before(s) {
		return 0, 0, errFormDateRangeInvalid
	}
	return s.UnixMilli(), e.UnixMilli(), nil
}

func splitCSV(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

// Sentinel errors — could be turned into label lookups later.
var (
	errFormStartRequired    = newFormErr("Start date is required")
	errFormDateRangeInvalid = newFormErr("End must be after start")
)

type formErr struct{ msg string }

func newFormErr(s string) error           { return &formErr{msg: s} }
func (e *formErr) Error() string          { return e.msg }
