package action

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	cyta "github.com/erniealice/cyta-golang"
	cytaeventform "github.com/erniealice/cyta-golang/views/event/form"

	"github.com/erniealice/pyeza-golang/route"
	"github.com/erniealice/pyeza-golang/view"

	eventpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/event/event"
	pyeza "github.com/erniealice/pyeza-golang"
)

// Deps holds dependencies for event action handlers.
//
// The classic CRUD funcs are unchanged. The new fields (ListEventTags,
// SearchAttendees, SetEventTagAssignments, SyncEventAttendees, ListAttachments)
// are nillable — when nil the corresponding feature degrades gracefully so the
// drawer still renders for environments where the espyna wiring isn't complete.
type Deps struct {
	Routes       cyta.EventRoutes
	Labels       cyta.EventLabels
	CommonLabels pyeza.CommonLabels

	// Core event CRUD
	CreateEvent func(ctx context.Context, req *eventpb.CreateEventRequest) (*eventpb.CreateEventResponse, error)
	ReadEvent   func(ctx context.Context, req *eventpb.ReadEventRequest) (*eventpb.ReadEventResponse, error)
	UpdateEvent func(ctx context.Context, req *eventpb.UpdateEventRequest) (*eventpb.UpdateEventResponse, error)
	DeleteEvent func(ctx context.Context, req *eventpb.DeleteEventRequest) (*eventpb.DeleteEventResponse, error)
	ListEvents  func(ctx context.Context, req *eventpb.ListEventsRequest) (*eventpb.ListEventsResponse, error)

	// Phase 4 additions — pickers + sync
	// All return ([]form.Option, ...) shaped lists so the templates don't
	// need to know about proto types. Wired by block.go.
	ListEventTags          func(ctx context.Context) ([]cytaeventform.Option, error)
	ListEventTagsForEvent  func(ctx context.Context, eventID string) ([]string, error) // assigned tag IDs
	SearchAttendees        func(ctx context.Context, query string) ([]cytaeventform.Option, error)
	ListAttendeesForEvent  func(ctx context.Context, eventID string) ([]cytaeventform.SelectedOption, error)
	SetEventTagAssignments func(ctx context.Context, eventID string, tagIDs []string) error
	SyncEventAttendees     func(ctx context.Context, eventID string, attendeeRefs []string) error
	ListEventAttachments   func(ctx context.Context, eventID string) ([]cytaeventform.Attachment, error)
}

// NewAddAction creates the event add action (GET = form, POST = create).
//
// Query params on GET: ?date=YYYY-MM-DD&at=HH:MM (set by the calendar
// popover; pre-seeds the start fields). When absent, start defaults to the
// next-half-hour-from-now logic in the JS.
//
// POST is delegated to the unified save handler in handlers_save.go.
func NewAddAction(deps *Deps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("event", "create") {
			return htmxError(deps.Labels.Errors.NameRequired)
		}

		if viewCtx.Request.Method == http.MethodGet {
			date := viewCtx.Request.URL.Query().Get("date")
			at := viewCtx.Request.URL.Query().Get("at")

			startDate, startTime := seedStartFromQuery(date, at, time.Now())
			endDate, endTime := defaultEnd(startDate, startTime)

			tagOptions := safeListTags(ctx, deps)

			return view.OK("event-drawer-form", &cytaeventform.Data{
				FormAction:    deps.Routes.AddURL,
				StartDate:     startDate,
				StartTime:     startTime,
				EndDate:       endDate,
				EndTime:       endTime,
				StatusOptions: cytaeventform.BuildStatusOptions(deps.Labels.Status, eventpb.EventStatus_EVENT_STATUS_TENTATIVE),
				TagOptions:    tagOptions,
				// AttendeeOptions empty — populated client-side by the multi-select search hook.
				Labels: cytaeventform.Labels{
					NameLabel:           deps.Labels.Form.Name,
					NamePlaceholder:     deps.Labels.Form.NamePlaceholder,
					AllDayLabel:         deps.Labels.Form.AllDay,
					StartDateLabel:      deps.Labels.Form.StartDate,
					StartTimeLabel:      deps.Labels.Form.StartTime,
					EndDateLabel:        deps.Labels.Form.EndDate,
					EndTimeLabel:        deps.Labels.Form.EndTime,
					TimezoneLabel:       deps.Labels.Form.Timezone,
					StatusLabel:         deps.Labels.Form.Status,
					InviteesLabel:       deps.Labels.Form.Invitees,
					InviteesPlaceholder: deps.Labels.Form.InviteesPlaceholder,
					TagsLabel:           deps.Labels.Form.Tags,
					TagsPlaceholder:     deps.Labels.Form.TagsPlaceholder,
					NotesLabel:          deps.Labels.Form.Notes,
					NotesPlaceholder:    deps.Labels.Form.NotesPlaceholder,
					AttachmentsLabel:    deps.Labels.Form.Attachments,
					AttachmentsHint:     deps.Labels.Form.AttachmentsHint,
				},
				CommonLabels: deps.CommonLabels,
			})
		}

		// POST — delegate to save handler
		return handleSave(ctx, viewCtx, deps, "" /* no existing id */)
	})
}

// NewEditAction creates the event edit action (GET = form, POST = update).
func NewEditAction(deps *Deps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("event", "update") {
			return htmxError(deps.Labels.Errors.NameRequired)
		}

		id := viewCtx.Request.PathValue("id")

		if viewCtx.Request.Method == http.MethodGet {
			readResp, err := deps.ReadEvent(ctx, &eventpb.ReadEventRequest{
				Data: &eventpb.Event{Id: id},
			})
			if err != nil {
				log.Printf("Failed to read event %s: %v", id, err)
				return htmxError(deps.Labels.Errors.NameRequired)
			}
			readData := readResp.GetData()
			if len(readData) == 0 {
				return htmxError(deps.Labels.Errors.NameRequired)
			}
			record := readData[0]

			startDate, startTime := splitTimestamp(record.GetStartDateTimeUtc())
			endDate, endTime := splitTimestamp(record.GetEndDateTimeUtc())

			// The four auxiliary list calls are independent — fetch them in
			// parallel to reduce drawer-open latency.
			var (
				tagOptions        []cytaeventform.Option
				selectedTagIDs    []string
				selectedAttendees []cytaeventform.SelectedOption
				attachments       []cytaeventform.Attachment
				wg                sync.WaitGroup
			)
			wg.Add(4)
			go func() {
				defer wg.Done()
				tagOptions = safeListTags(ctx, deps)
			}()
			go func() {
				defer wg.Done()
				selectedTagIDs = safeListTagIDsForEvent(ctx, deps, id)
			}()
			go func() {
				defer wg.Done()
				selectedAttendees = safeListAttendees(ctx, deps, id)
			}()
			go func() {
				defer wg.Done()
				attachments = safeListAttachments(ctx, deps, id)
			}()
			wg.Wait()
			markSelectedTags(tagOptions, selectedTagIDs)

			return view.OK("event-drawer-form", &cytaeventform.Data{
				FormAction:        route.ResolveURL(deps.Routes.EditURL, "id", id),
				IsEdit:            true,
				ID:                id,
				Name:              record.GetName(),
				Notes:             record.GetDescription(),
				StartDate:         startDate,
				StartTime:         startTime,
				EndDate:           endDate,
				EndTime:           endTime,
				Timezone:          record.GetTimezone(),
				AllDay:            record.GetAllDay(),
				StatusOptions:     cytaeventform.BuildStatusOptions(deps.Labels.Status, record.GetStatus()),
				TagOptions:        tagOptions,
				SelectedTags:      buildSelectedFromOptions(tagOptions, selectedTagIDs),
				AttendeeOptions:   nil, // edit mode shows existing as chips; new searches via client-side
				SelectedAttendees: selectedAttendees,
				Attachments:       attachments,
				Labels: cytaeventform.Labels{
					NameLabel:           deps.Labels.Form.Name,
					NamePlaceholder:     deps.Labels.Form.NamePlaceholder,
					AllDayLabel:         deps.Labels.Form.AllDay,
					StartDateLabel:      deps.Labels.Form.StartDate,
					StartTimeLabel:      deps.Labels.Form.StartTime,
					EndDateLabel:        deps.Labels.Form.EndDate,
					EndTimeLabel:        deps.Labels.Form.EndTime,
					TimezoneLabel:       deps.Labels.Form.Timezone,
					StatusLabel:         deps.Labels.Form.Status,
					InviteesLabel:       deps.Labels.Form.Invitees,
					InviteesPlaceholder: deps.Labels.Form.InviteesPlaceholder,
					TagsLabel:           deps.Labels.Form.Tags,
					TagsPlaceholder:     deps.Labels.Form.TagsPlaceholder,
					NotesLabel:          deps.Labels.Form.Notes,
					NotesPlaceholder:    deps.Labels.Form.NotesPlaceholder,
					AttachmentsLabel:    deps.Labels.Form.Attachments,
					AttachmentsHint:     deps.Labels.Form.AttachmentsHint,
				},
				CommonLabels: deps.CommonLabels,
			})
		}

		// POST — delegate to save handler
		return handleSave(ctx, viewCtx, deps, id)
	})
}

// NewDeleteAction — unchanged from the previous implementation.
func NewDeleteAction(deps *Deps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("event", "delete") {
			return htmxError(deps.Labels.Errors.NameRequired)
		}

		id := viewCtx.Request.URL.Query().Get("id")
		if id == "" {
			_ = viewCtx.Request.ParseForm()
			id = viewCtx.Request.FormValue("id")
		}
		if id == "" {
			return htmxError("ID is required")
		}

		_, err := deps.DeleteEvent(ctx, &eventpb.DeleteEventRequest{
			Data: &eventpb.Event{Id: id},
		})
		if err != nil {
			log.Printf("Failed to delete event %s: %v", id, err)
			return htmxError(err.Error())
		}

		return htmxSuccess("events-table")
	})
}

// NewBulkDeleteAction — unchanged.
func NewBulkDeleteAction(deps *Deps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("event", "delete") {
			return htmxError(deps.Labels.Errors.NameRequired)
		}

		_ = viewCtx.Request.ParseMultipartForm(32 << 20)

		ids := viewCtx.Request.Form["id"]
		if len(ids) == 0 {
			return htmxError("No IDs provided")
		}

		for _, id := range ids {
			_, err := deps.DeleteEvent(ctx, &eventpb.DeleteEventRequest{
				Data: &eventpb.Event{Id: id},
			})
			if err != nil {
				log.Printf("Failed to delete event %s: %v", id, err)
			}
		}

		return htmxSuccess("events-table")
	})
}

// NewSetStatusAction — unchanged.
func NewSetStatusAction(deps *Deps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("event", "update") {
			return htmxError(deps.Labels.Errors.NameRequired)
		}

		id := viewCtx.Request.URL.Query().Get("id")
		targetStatus := viewCtx.Request.URL.Query().Get("status")

		if id == "" {
			_ = viewCtx.Request.ParseForm()
			id = viewCtx.Request.FormValue("id")
			targetStatus = viewCtx.Request.FormValue("status")
		}
		if id == "" {
			return htmxError("ID is required")
		}
		if targetStatus == "" {
			return htmxError("Status is required")
		}

		statusEnum := cytaeventform.StatusFromString(targetStatus)

		_, err := deps.UpdateEvent(ctx, &eventpb.UpdateEventRequest{
			Data: &eventpb.Event{Id: id, Status: statusEnum},
		})
		if err != nil {
			log.Printf("Failed to update event status %s: %v", id, err)
			return htmxError(err.Error())
		}

		return htmxSuccess("events-table")
	})
}

// NewBulkSetStatusAction — unchanged.
func NewBulkSetStatusAction(deps *Deps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("event", "update") {
			return htmxError(deps.Labels.Errors.NameRequired)
		}

		_ = viewCtx.Request.ParseMultipartForm(32 << 20)

		ids := viewCtx.Request.Form["id"]
		targetStatus := viewCtx.Request.FormValue("target_status")

		if len(ids) == 0 {
			return htmxError("No IDs provided")
		}
		if targetStatus == "" {
			return htmxError("Target status is required")
		}

		statusEnum := cytaeventform.StatusFromString(targetStatus)

		for _, id := range ids {
			if _, err := deps.UpdateEvent(ctx, &eventpb.UpdateEventRequest{
				Data: &eventpb.Event{Id: id, Status: statusEnum},
			}); err != nil {
				log.Printf("Failed to update event status %s: %v", id, err)
			}
		}

		return htmxSuccess("events-table")
	})
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// strPtr returns a pointer to a string.
func strPtr(s string) *string {
	return &s
}

// htmxSuccess returns a header-only response that signals the sheet to close and table to refresh.
func htmxSuccess(tableID string) view.ViewResult {
	return view.ViewResult{
		StatusCode: http.StatusOK,
		Headers: map[string]string{
			"HX-Trigger": fmt.Sprintf(`{"formSuccess":true,"refreshTable":"%s"}`, tableID),
		},
	}
}

// htmxError returns a header-only response that signals a form error.
func htmxError(message string) view.ViewResult {
	return view.ViewResult{
		StatusCode: http.StatusUnprocessableEntity,
		Headers: map[string]string{
			"HX-Error-Message": message,
		},
	}
}

// safeListTags wraps deps.ListEventTags with a nil check.
func safeListTags(ctx context.Context, deps *Deps) []cytaeventform.Option {
	if deps.ListEventTags == nil {
		return nil
	}
	tags, err := deps.ListEventTags(ctx)
	if err != nil {
		log.Printf("ListEventTags failed: %v", err)
		return nil
	}
	return tags
}

func safeListTagIDsForEvent(ctx context.Context, deps *Deps, eventID string) []string {
	if deps.ListEventTagsForEvent == nil {
		return nil
	}
	ids, err := deps.ListEventTagsForEvent(ctx, eventID)
	if err != nil {
		log.Printf("ListEventTagsForEvent failed: %v", err)
		return nil
	}
	return ids
}

func safeListAttendees(ctx context.Context, deps *Deps, eventID string) []cytaeventform.SelectedOption {
	if deps.ListAttendeesForEvent == nil {
		return nil
	}
	out, err := deps.ListAttendeesForEvent(ctx, eventID)
	if err != nil {
		log.Printf("ListAttendeesForEvent failed: %v", err)
		return nil
	}
	return out
}

func safeListAttachments(ctx context.Context, deps *Deps, eventID string) []cytaeventform.Attachment {
	if deps.ListEventAttachments == nil {
		return nil
	}
	out, err := deps.ListEventAttachments(ctx, eventID)
	if err != nil {
		log.Printf("ListEventAttachments failed: %v", err)
		return nil
	}
	return out
}

// markSelectedTags flips Option.Selected for any option whose Value is in selectedIDs.
func markSelectedTags(opts []cytaeventform.Option, selectedIDs []string) {
	if len(opts) == 0 || len(selectedIDs) == 0 {
		return
	}
	set := make(map[string]struct{}, len(selectedIDs))
	for _, id := range selectedIDs {
		set[id] = struct{}{}
	}
	for i := range opts {
		if _, ok := set[opts[i].Value]; ok {
			opts[i].Selected = true
		}
	}
}

// buildSelectedFromOptions returns the {Value, Label} pairs for the supplied
// IDs in the same order as the master option list (so chip ordering matches
// the dropdown's natural ordering).
func buildSelectedFromOptions(opts []cytaeventform.Option, selectedIDs []string) []cytaeventform.SelectedOption {
	if len(opts) == 0 || len(selectedIDs) == 0 {
		return nil
	}
	wanted := make(map[string]struct{}, len(selectedIDs))
	for _, id := range selectedIDs {
		wanted[id] = struct{}{}
	}
	var out []cytaeventform.SelectedOption
	for _, o := range opts {
		if _, ok := wanted[o.Value]; ok {
			out = append(out, cytaeventform.SelectedOption{Value: o.Value, Label: o.Label})
		}
	}
	return out
}

// seedStartFromQuery parses ?date=YYYY-MM-DD&at=HH:MM into form-friendly strings.
// Falls back to today + next half-hour-from-now if either is missing.
func seedStartFromQuery(date, at string, now time.Time) (string, string) {
	if date == "" {
		date = now.Format("2006-01-02")
	}
	if at == "" {
		at = nextHalfHour(now)
	}
	return date, at
}

// defaultEnd = start + 60 min, mirrored across midnight if needed.
func defaultEnd(startDate, startTime string) (string, string) {
	if startDate == "" || startTime == "" {
		return "", ""
	}
	t, err := time.Parse("2006-01-02 15:04", startDate+" "+startTime)
	if err != nil {
		return startDate, ""
	}
	end := t.Add(60 * time.Minute)
	return end.Format("2006-01-02"), end.Format("15:04")
}

// nextHalfHour returns the next half-hour boundary ≥ now ("HH:MM" 24h).
// Caps at 18:00; past business hours falls back to "09:00".
func nextHalfHour(now time.Time) string {
	h, m := now.Hour(), now.Minute()
	if m == 0 || m == 30 {
		// already on a boundary — push one half-hour to ensure ≥ now
		m += 30
	} else if m < 30 {
		m = 30
	} else {
		h++
		m = 0
	}
	if h >= 18 {
		return "09:00"
	}
	if h < 9 {
		return "09:00"
	}
	if m == 60 {
		h++
		m = 0
	}
	return fmt.Sprintf("%02d:%02d", h, m)
}

// splitTimestamp converts a UTC unix-millis timestamp into ("YYYY-MM-DD", "HH:MM").
// Returns empty strings for zero/invalid input.
func splitTimestamp(ms int64) (string, string) {
	if ms == 0 {
		return "", ""
	}
	t := time.UnixMilli(ms)
	return t.Format("2006-01-02"), t.Format("15:04")
}
