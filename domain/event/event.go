// Package event is the facade for the cyta event domain.
//
// Consumers (block/, service-admin) use the E-prefixed names exported here
// (e.g. EventLabels, DefaultEventRoutes) which resolve to the entity-local
// types in the sub-packages below. This keeps consumer call sites unchanged
// while the internals follow the entity-local naming convention.
//
// Import cycle rule: entity packages MUST NEVER import this facade.
// Cross-entity references go DIRECT to the sibling package.
package event

import (
	eventpkg "github.com/erniealice/cyta-golang/domain/event/event"
	eventtagpkg "github.com/erniealice/cyta-golang/domain/event/event_tag"
	recurrencepkg "github.com/erniealice/cyta-golang/domain/event/recurrence"
	calendarpkg "github.com/erniealice/cyta-golang/domain/event/calendar"
)

// ---------------------------------------------------------------------------
// Event entity re-exports
// ---------------------------------------------------------------------------

// EventLabels re-exports Labels from the event entity package.
type EventLabels = eventpkg.Labels

// ScheduleDashboardLabels re-exports ScheduleDashboardLabels from the event entity package.
type ScheduleDashboardLabels = eventpkg.ScheduleDashboardLabels

// EventCalendarPopoverLabels re-exports CalendarPopoverLabels from the event entity package.
type EventCalendarPopoverLabels = eventpkg.CalendarPopoverLabels

// EventPageLabels re-exports PageLabels from the event entity package.
type EventPageLabels = eventpkg.PageLabels

// EventButtonLabels re-exports ButtonLabels from the event entity package.
type EventButtonLabels = eventpkg.ButtonLabels

// EventColumnLabels re-exports ColumnLabels from the event entity package.
type EventColumnLabels = eventpkg.ColumnLabels

// EventEmptyLabels re-exports EmptyLabels from the event entity package.
type EventEmptyLabels = eventpkg.EmptyLabels

// EventFormLabels re-exports FormLabels from the event entity package.
type EventFormLabels = eventpkg.FormLabels

// EventActionLabels re-exports ActionLabels from the event entity package.
type EventActionLabels = eventpkg.ActionLabels

// EventDetailLabels re-exports DetailLabels from the event entity package.
type EventDetailLabels = eventpkg.DetailLabels

// EventTabLabels re-exports TabLabels from the event entity package.
type EventTabLabels = eventpkg.TabLabels

// EventConfirmLabels re-exports ConfirmLabels from the event entity package.
type EventConfirmLabels = eventpkg.ConfirmLabels

// EventErrorLabels re-exports ErrorLabels from the event entity package.
type EventErrorLabels = eventpkg.ErrorLabels

// EventStatusLabels re-exports StatusLabels from the event entity package.
type EventStatusLabels = eventpkg.StatusLabels

// EventRoutes re-exports Routes from the event entity package.
type EventRoutes = eventpkg.Routes

// DefaultEventRoutes re-exports DefaultRoutes from the event entity package.
var DefaultEventRoutes = eventpkg.DefaultRoutes

// Event URL constants re-exported from the event entity package.
const (
	EventListURL             = eventpkg.ListURL
	EventDetailURL           = eventpkg.DetailURL
	EventAddURL              = eventpkg.AddURL
	EventEditURL             = eventpkg.EditURL
	EventDeleteURL           = eventpkg.DeleteURL
	EventBulkDeleteURL       = eventpkg.BulkDeleteURL
	EventSetStatusURL        = eventpkg.SetStatusURL
	EventBulkSetStatusURL    = eventpkg.BulkSetStatusURL
	EventTabActionURL        = eventpkg.TabActionURL
	EventAttachmentUploadURL = eventpkg.AttachmentUploadURL
	EventAttachmentDeleteURL = eventpkg.AttachmentDeleteURL
	CalendarURL              = eventpkg.CalendarURL
	CalendarDataURL          = eventpkg.CalendarDataURL
	ScheduleDashboardURL     = eventpkg.DashboardURL
)

// ---------------------------------------------------------------------------
// EventTag entity re-exports
// ---------------------------------------------------------------------------

// EventTagLabels re-exports Labels from the event_tag entity package.
type EventTagLabels = eventtagpkg.Labels

// EventTagPageLabels re-exports PageLabels from the event_tag entity package.
type EventTagPageLabels = eventtagpkg.PageLabels

// EventTagButtonLabels re-exports ButtonLabels from the event_tag entity package.
type EventTagButtonLabels = eventtagpkg.ButtonLabels

// EventTagColumnLabels re-exports ColumnLabels from the event_tag entity package.
type EventTagColumnLabels = eventtagpkg.ColumnLabels

// EventTagEmptyLabels re-exports EmptyLabels from the event_tag entity package.
type EventTagEmptyLabels = eventtagpkg.EmptyLabels

// EventTagFormLabels re-exports FormLabels from the event_tag entity package.
type EventTagFormLabels = eventtagpkg.FormLabels

// EventTagActionLabels re-exports ActionLabels from the event_tag entity package.
type EventTagActionLabels = eventtagpkg.ActionLabels

// EventTagRoutes re-exports Routes from the event_tag entity package.
type EventTagRoutes = eventtagpkg.Routes

// DefaultEventTagRoutes re-exports DefaultRoutes from the event_tag entity package.
var DefaultEventTagRoutes = eventtagpkg.DefaultRoutes

// EventTag URL constants re-exported from the event_tag entity package.
const (
	EventTagListURL   = eventtagpkg.ListURL
	EventTagDetailURL = eventtagpkg.DetailURL
	EventTagAddURL    = eventtagpkg.AddURL
	EventTagEditURL   = eventtagpkg.EditURL
	EventTagDeleteURL = eventtagpkg.DeleteURL
)

// ---------------------------------------------------------------------------
// Recurrence entity re-exports
// ---------------------------------------------------------------------------

// RecurrenceLabels re-exports Labels from the recurrence entity package.
type RecurrenceLabels = recurrencepkg.Labels

// RecurrencePageLabels re-exports PageLabels from the recurrence entity package.
type RecurrencePageLabels = recurrencepkg.PageLabels

// RecurrenceButtonLabels re-exports ButtonLabels from the recurrence entity package.
type RecurrenceButtonLabels = recurrencepkg.ButtonLabels

// RecurrenceColumnLabels re-exports ColumnLabels from the recurrence entity package.
type RecurrenceColumnLabels = recurrencepkg.ColumnLabels

// RecurrenceEmptyLabels re-exports EmptyLabels from the recurrence entity package.
type RecurrenceEmptyLabels = recurrencepkg.EmptyLabels

// RecurrenceFormLabels re-exports FormLabels from the recurrence entity package.
type RecurrenceFormLabels = recurrencepkg.FormLabels

// RecurrenceConfirmLabels re-exports ConfirmLabels from the recurrence entity package.
type RecurrenceConfirmLabels = recurrencepkg.ConfirmLabels

// RecurrenceRoutes re-exports Routes from the recurrence entity package.
type RecurrenceRoutes = recurrencepkg.Routes

// DefaultRecurrenceRoutes re-exports DefaultRoutes from the recurrence entity package.
var DefaultRecurrenceRoutes = recurrencepkg.DefaultRoutes

// Recurrence URL constants re-exported from the recurrence entity package.
const (
	RecurrenceListURL   = recurrencepkg.ListURL
	RecurrenceDetailURL = recurrencepkg.DetailURL
	RecurrenceAddURL    = recurrencepkg.AddURL
	RecurrenceEditURL   = recurrencepkg.EditURL
	RecurrenceDeleteURL = recurrencepkg.DeleteURL
)

// ---------------------------------------------------------------------------
// Calendar re-exports
// ---------------------------------------------------------------------------

// CalendarLabels re-exports Labels from the calendar package.
type CalendarLabels = calendarpkg.Labels

// CalendarPageLabels re-exports PageLabels from the calendar package.
type CalendarPageLabels = calendarpkg.PageLabels
