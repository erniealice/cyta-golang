package calendar

import (
	"context"
	"fmt"
	"time"

	cyta "github.com/erniealice/cyta-golang"
	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"
)

// CalendarDate holds a year/month/day triple for template rendering.
type CalendarDate struct {
	Year  int
	Month int
	Day   int
}

// CalendarEvent is a single event used in both month and week views.
type CalendarEvent struct {
	ID        string
	Name      string
	StartTime string
	EndTime   string
	Status    string // "confirmed" | "tentative" | "cancelled"
	DetailURL string
	HXTarget  string
	// Week-view positioning (server-computed)
	TopPct    float64
	HeightPct float64
	IsCompact bool
}

// CalendarDay is one cell in the month grid.
type CalendarDay struct {
	Date         CalendarDate
	IsToday      bool
	IsOtherMonth bool
	Events       []CalendarEvent
	MoreCount    int
	DayURL       string
}

// CalendarWeek is a row of 7 days in the month grid.
type CalendarWeek struct {
	Days []CalendarDay
}

// CalendarMonthData is the data passed to the calendar-month template.
type CalendarMonthData struct {
	types.PageData
	ContentTemplate string

	Year      int
	Month     int
	MonthName string
	Weeks     []CalendarWeek
	Today     CalendarDate

	PrevMonthURL string
	NextMonthURL string
	TodayURL     string

	ViewMode     string
	MonthViewURL string
	WeekViewURL  string
	DayViewURL   string
	HXTarget     string
}

// CalendarHour is a single hour label in the week-view time column.
type CalendarHour struct {
	Hour  int
	Label string
}

// CalendarWeekDay is one column in the week time-grid.
type CalendarWeekDay struct {
	Date    CalendarDate
	DayName string
	IsToday bool
	Events  []CalendarEvent
	DayURL  string
}

// CalendarWeekData is the data passed to the calendar-week template.
type CalendarWeekData struct {
	types.PageData
	ContentTemplate string

	RangeLabel string
	Days       []CalendarWeekDay
	HourStart  int
	HourEnd    int
	Hours      []CalendarHour

	PrevWeekURL  string
	NextWeekURL  string
	TodayURL     string

	ViewMode     string
	MonthViewURL string
	WeekViewURL  string
	DayViewURL   string
	HXTarget     string
}

// ViewDeps holds dependencies for the calendar view.
type ViewDeps struct {
	Routes       cyta.EventRoutes
	Labels       cyta.EventLabels
	CommonLabels pyeza.CommonLabels
}

// NewView creates the calendar view handler.
func NewView(deps *ViewDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		r := viewCtx.Request

		viewMode := r.URL.Query().Get("view")
		if viewMode == "" {
			viewMode = "month"
		}

		dateParam := r.URL.Query().Get("date")
		target := r.URL.Query().Get("target")
		if target == "" {
			target = "#calendar"
		}

		now := time.Now()
		var focusDate time.Time
		if dateParam != "" {
			if d, err := time.Parse("2006-01-02", dateParam); err == nil {
				focusDate = d
			} else {
				focusDate = now
			}
		} else {
			focusDate = now
		}

		baseURL := deps.Routes.CalendarURL
		todayStr := now.Format("2006-01-02")

		todayURL := fmt.Sprintf("%s?view=%s&date=%s", baseURL, viewMode, todayStr)
		monthViewURL := fmt.Sprintf("%s?view=month&date=%s", baseURL, focusDate.Format("2006-01-02"))
		weekViewURL := fmt.Sprintf("%s?view=week&date=%s", baseURL, focusDate.Format("2006-01-02"))
		dayViewURL := fmt.Sprintf("%s?view=day&date=%s", baseURL, focusDate.Format("2006-01-02"))

		// Sample events for visual demo
		todayDate := now
		tomorrowDate := now.AddDate(0, 0, 1)

		sampleEvents := []CalendarEvent{
			{
				ID:        "ev-001",
				Name:      "Team Meeting",
				StartTime: "10:00 AM",
				EndTime:   "11:00 AM",
				Status:    "confirmed",
				DetailURL: fmt.Sprintf("/app/schedule/detail/ev-001"),
			},
			{
				ID:        "ev-002",
				Name:      "Client Consultation",
				StartTime: "2:00 PM",
				EndTime:   "3:30 PM",
				Status:    "tentative",
				DetailURL: fmt.Sprintf("/app/schedule/detail/ev-002"),
			},
			{
				ID:        "ev-003",
				Name:      "Staff Training",
				StartTime: "9:00 AM",
				EndTime:   "12:00 PM",
				Status:    "confirmed",
				DetailURL: fmt.Sprintf("/app/schedule/detail/ev-003"),
			},
			{
				ID:        "ev-004",
				Name:      "Product Launch Review",
				StartTime: "3:00 PM",
				EndTime:   "4:00 PM",
				Status:    "cancelled",
				DetailURL: fmt.Sprintf("/app/schedule/detail/ev-004"),
			},
		}

		switch viewMode {
		case "week":
			data := buildWeekData(viewCtx, deps, focusDate, now, todayDate, tomorrowDate, sampleEvents,
				todayURL, monthViewURL, weekViewURL, dayViewURL, baseURL, target)
			isHTMX := r.Header.Get("HX-Request") == "true"
			if isHTMX {
				return view.OK("event-calendar-week-content", data)
			}
			return view.OK("event-calendar", data)

		default: // "month"
			data := buildMonthData(viewCtx, deps, focusDate, now, todayDate, tomorrowDate, sampleEvents,
				todayURL, monthViewURL, weekViewURL, dayViewURL, baseURL, target)
			isHTMX := r.Header.Get("HX-Request") == "true"
			if isHTMX {
				return view.OK("event-calendar-month-content", data)
			}
			return view.OK("event-calendar", data)
		}
	})
}

// buildMonthData constructs CalendarMonthData for the current month containing focusDate.
func buildMonthData(
	viewCtx *view.ViewContext,
	deps *ViewDeps,
	focusDate, now, todayDate, tomorrowDate time.Time,
	sampleEvents []CalendarEvent,
	todayURL, monthViewURL, weekViewURL, dayViewURL, baseURL, target string,
) *CalendarMonthData {
	year, month, _ := focusDate.Date()

	// First day of the month
	firstDay := time.Date(year, month, 1, 0, 0, 0, 0, time.Local)

	// Navigation URLs
	prevMonth := firstDay.AddDate(0, -1, 0)
	nextMonth := firstDay.AddDate(0, 1, 0)
	prevMonthURL := fmt.Sprintf("%s?view=month&date=%s", baseURL, prevMonth.Format("2006-01-02"))
	nextMonthURL := fmt.Sprintf("%s?view=month&date=%s", baseURL, nextMonth.Format("2006-01-02"))

	// Build weeks
	weeks := buildMonthWeeks(year, month, now, todayDate, tomorrowDate, sampleEvents, baseURL)

	todayCalDate := CalendarDate{Year: now.Year(), Month: int(now.Month()), Day: now.Day()}

	return &CalendarMonthData{
		PageData: types.PageData{
			CacheVersion: viewCtx.CacheVersion,
			Title:        "Calendar",
			CurrentPath:  viewCtx.CurrentPath,
			ActiveNav:    deps.Routes.ActiveNav,
			ActiveSubNav: deps.Routes.ActiveSubNav,
			CommonLabels: deps.CommonLabels,
		},
		ContentTemplate: "event-calendar-month-content",
		Year:            year,
		Month:           int(month),
		MonthName:       fmt.Sprintf("%s %d", month.String(), year),
		Weeks:           weeks,
		Today:           todayCalDate,
		PrevMonthURL:    prevMonthURL,
		NextMonthURL:    nextMonthURL,
		TodayURL:        todayURL,
		ViewMode:        "month",
		MonthViewURL:    monthViewURL,
		WeekViewURL:     weekViewURL,
		DayViewURL:      dayViewURL,
		HXTarget:        target,
	}
}

// buildMonthWeeks returns the grid rows for the month calendar.
func buildMonthWeeks(
	year int, month time.Month,
	now, todayDate, tomorrowDate time.Time,
	sampleEvents []CalendarEvent,
	baseURL string,
) []CalendarWeek {
	firstDay := time.Date(year, month, 1, 0, 0, 0, 0, time.Local)
	lastDay := firstDay.AddDate(0, 1, -1)

	// Start of the grid (Sunday of the week containing firstDay)
	startOffset := int(firstDay.Weekday()) // 0=Sun
	gridStart := firstDay.AddDate(0, 0, -startOffset)

	// End of the grid (Saturday of the week containing lastDay)
	endOffset := 6 - int(lastDay.Weekday())
	gridEnd := lastDay.AddDate(0, 0, endOffset)

	var weeks []CalendarWeek
	current := gridStart

	for !current.After(gridEnd) {
		var week CalendarWeek
		for d := 0; d < 7; d++ {
			day := current.AddDate(0, 0, d)
			dayStr := day.Format("2006-01-02")

			isToday := sameDay(day, now)
			isOther := day.Month() != month

			var dayEvents []CalendarEvent
			// Assign sample events to today and tomorrow
			if sameDay(day, todayDate) {
				dayEvents = append(dayEvents, sampleEvents[0], sampleEvents[1])
			}
			if sameDay(day, tomorrowDate) {
				dayEvents = append(dayEvents, sampleEvents[2], sampleEvents[3])
			}

			moreCount := 0
			if len(dayEvents) > 3 {
				moreCount = len(dayEvents) - 3
			}

			week.Days = append(week.Days, CalendarDay{
				Date:         CalendarDate{Year: day.Year(), Month: int(day.Month()), Day: day.Day()},
				IsToday:      isToday,
				IsOtherMonth: isOther,
				Events:       dayEvents,
				MoreCount:    moreCount,
				DayURL:       fmt.Sprintf("%s?view=day&date=%s", baseURL, dayStr),
			})
		}
		weeks = append(weeks, week)
		current = current.AddDate(0, 0, 7)
	}

	return weeks
}

// buildWeekData constructs CalendarWeekData for the week containing focusDate.
func buildWeekData(
	viewCtx *view.ViewContext,
	deps *ViewDeps,
	focusDate, now, todayDate, tomorrowDate time.Time,
	sampleEvents []CalendarEvent,
	todayURL, monthViewURL, weekViewURL, dayViewURL, baseURL, target string,
) *CalendarWeekData {
	hourStart := 7
	hourEnd := 21
	totalHours := float64(hourEnd - hourStart)

	// Start of week (Sunday)
	weekdayOffset := int(focusDate.Weekday())
	weekStart := focusDate.AddDate(0, 0, -weekdayOffset)
	weekEnd := weekStart.AddDate(0, 0, 6)

	prevWeek := weekStart.AddDate(0, 0, -7)
	nextWeek := weekStart.AddDate(0, 0, 7)
	prevWeekURL := fmt.Sprintf("%s?view=week&date=%s", baseURL, prevWeek.Format("2006-01-02"))
	nextWeekURL := fmt.Sprintf("%s?view=week&date=%s", baseURL, nextWeek.Format("2006-01-02"))

	rangeLabel := fmt.Sprintf("%s %d–%d, %d",
		weekStart.Format("Jan"),
		weekStart.Day(),
		weekEnd.Day(),
		weekEnd.Year(),
	)

	// Build days
	days := make([]CalendarWeekDay, 7)
	for i := 0; i < 7; i++ {
		d := weekStart.AddDate(0, 0, i)
		dayStr := d.Format("2006-01-02")

		var evs []CalendarEvent
		if sameDay(d, todayDate) {
			// Team Meeting: 10:00–11:00
			evs = append(evs, positionWeekEvent(sampleEvents[0], 10, 0, 11, 0, hourStart, totalHours))
			// Client Consultation: 14:00–15:30
			evs = append(evs, positionWeekEvent(sampleEvents[1], 14, 0, 15, 30, hourStart, totalHours))
		}
		if sameDay(d, tomorrowDate) {
			// Staff Training: 9:00–12:00
			evs = append(evs, positionWeekEvent(sampleEvents[2], 9, 0, 12, 0, hourStart, totalHours))
			// Cancelled: 15:00–16:00
			evs = append(evs, positionWeekEvent(sampleEvents[3], 15, 0, 16, 0, hourStart, totalHours))
		}

		dayNames := []string{"Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"}
		days[i] = CalendarWeekDay{
			Date:    CalendarDate{Year: d.Year(), Month: int(d.Month()), Day: d.Day()},
			DayName: dayNames[i],
			IsToday: sameDay(d, now),
			Events:  evs,
			DayURL:  fmt.Sprintf("%s?view=day&date=%s", baseURL, dayStr),
		}
	}

	// Build hour labels
	hours := make([]CalendarHour, hourEnd-hourStart)
	for i := hourStart; i < hourEnd; i++ {
		label := fmt.Sprintf("%d AM", i)
		if i == 0 {
			label = "12 AM"
		} else if i == 12 {
			label = "12 PM"
		} else if i > 12 {
			label = fmt.Sprintf("%d PM", i-12)
		}
		hours[i-hourStart] = CalendarHour{Hour: i, Label: label}
	}

	return &CalendarWeekData{
		PageData: types.PageData{
			CacheVersion: viewCtx.CacheVersion,
			Title:        "Calendar",
			CurrentPath:  viewCtx.CurrentPath,
			ActiveNav:    deps.Routes.ActiveNav,
			ActiveSubNav: deps.Routes.ActiveSubNav,
			CommonLabels: deps.CommonLabels,
		},
		ContentTemplate: "event-calendar-week-content",
		RangeLabel:      rangeLabel,
		Days:            days,
		HourStart:       hourStart,
		HourEnd:         hourEnd,
		Hours:           hours,
		PrevWeekURL:     prevWeekURL,
		NextWeekURL:     nextWeekURL,
		TodayURL:        todayURL,
		ViewMode:        "week",
		MonthViewURL:    monthViewURL,
		WeekViewURL:     weekViewURL,
		DayViewURL:      dayViewURL,
		HXTarget:        target,
	}
}

// positionWeekEvent clones an event and computes TopPct/HeightPct for the week time grid.
func positionWeekEvent(ev CalendarEvent, startHour, startMin, endHour, endMin, hourStart int, totalHours float64) CalendarEvent {
	startOffset := float64(startHour-hourStart) + float64(startMin)/60.0
	endOffset := float64(endHour-hourStart) + float64(endMin)/60.0
	durationHours := endOffset - startOffset

	ev.TopPct = (startOffset / totalHours) * 100
	ev.HeightPct = (durationHours / totalHours) * 100
	ev.IsCompact = durationHours <= 0.5
	return ev
}

// sameDay returns true if two times share the same calendar date.
func sameDay(a, b time.Time) bool {
	ay, am, ad := a.Date()
	by, bm, bd := b.Date()
	return ay == by && am == bm && ad == bd
}
