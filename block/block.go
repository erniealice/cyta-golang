// Package block implements the Lego pattern for the cyta domain.
//
// Block() returns a pyeza.AppOption that registers the cyta event/calendar module
// using AppContext as the shared infrastructure carrier.
//
// Usage:
//
//	// Register all cyta modules (currently only event)
//	app.Apply(cytablock.Block())
//
//	// Register only specific modules
//	app.Apply(cytablock.Block(cytablock.WithEvent()))
package block

import (
	"context"
	"fmt"
	"log"

	event "github.com/erniealice/cyta-golang/domain/event"
	eventmod "github.com/erniealice/cyta-golang/domain/event/views/event"
	eventtagmod "github.com/erniealice/cyta-golang/domain/event/views/event_tag"
	"github.com/erniealice/espyna-golang/reference"
	attachmentpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/document/attachment"
	eventpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/event/event"
	lynguaV1 "github.com/erniealice/lyngua/golang/v1"
	pyeza "github.com/erniealice/pyeza-golang"
)

// ---------------------------------------------------------------------------
// BlockOption — per-module granular selection
// ---------------------------------------------------------------------------

// BlockOption enables specific cyta sub-modules within Block().
type BlockOption func(*blockConfig)

type blockConfig struct {
	enableAll bool
	event     bool
	eventTag  bool
	useCases  *UseCases
}

// WithUseCases supplies the typed use-case closures to Block().
// Required: Block() returns an error if this option is not provided.
// Service-admin constructs the *UseCases via an adapter function that bridges
// espyna's consumer container to cyta's typed shape (see buildCytaUseCases in
// apps/service-admin/internal/composition/adapters.go).
func WithUseCases(uc *UseCases) BlockOption {
	return func(c *blockConfig) { c.useCases = uc }
}

// WithEvent registers the Event module (list, detail, CRUD, calendar).
func WithEvent() BlockOption { return func(c *blockConfig) { c.event = true } }

// WithEventTag registers the EventTag module (list + drawer form).
func WithEventTag() BlockOption { return func(c *blockConfig) { c.eventTag = true } }

func (c *blockConfig) wantEvent() bool    { return c.enableAll || c.event }
func (c *blockConfig) wantEventTag() bool { return c.enableAll || c.eventTag }

// ---------------------------------------------------------------------------
// Block — the main Lego entry point
// ---------------------------------------------------------------------------

// Block registers cyta domain modules (schedule: events, calendar).
// Call with no options to register ALL modules. Call with specific With*() options
// to register a subset.
func Block(opts ...BlockOption) pyeza.AppOption {
	cfg := &blockConfig{}
	for _, opt := range opts {
		opt(cfg)
	}
	// "Enable all modules" is derived — true when no module-toggling option was
	// passed. Non-module options (WithUseCases) must NOT flip this off, else
	// `Block(WithUseCases(...))` would silently register zero modules.
	moduleSelected := cfg.event || cfg.eventTag
	cfg.enableAll = !moduleSelected

	return func(ctx *pyeza.AppContext) error {
		// --- typed UseCases supplied via WithUseCases() ---
		if cfg.useCases == nil {
			return fmt.Errorf("event.Block: WithUseCases(...) is required")
		}
		if err := cfg.useCases.RequireFor(cfg); err != nil {
			return err
		}
		uc := cfg.useCases // local alias for brevity

		// --- Type-assert translations ---
		translations, ok := ctx.Translations.(*lynguaV1.TranslationProvider)
		if !ok || translations == nil {
			return fmt.Errorf("event.Block: ctx.Translations must be *lynguaV1.TranslationProvider")
		}

		// --- Register Event module ---
		if cfg.wantEvent() {
			// Load routes (defaults + optional lyngua overrides)
			eventRoutes := event.DefaultEventRoutes()
			_ = translations.LoadPathIfExists("en", ctx.BusinessType, "route.json", "event", &eventRoutes)

			// Load labels — event.json is required in service translations but we
			// fall back to defaults if not present so Block() is self-contained.
			eventLabels := defaultEventLabels()
			_ = translations.LoadPathIfExists("en", ctx.BusinessType, "event.json", "event", &eventLabels)

			// --- Type-assert attachment operations ---
			uploadFile, _ := ctx.UploadFile.(func(context.Context, string, string, []byte, string) error)
			listAttachments, _ := ctx.ListAttachments.(func(context.Context, string, string) (*attachmentpb.ListAttachmentsResponse, error))
			createAttachment, _ := ctx.CreateAttachment.(func(context.Context, *attachmentpb.CreateAttachmentRequest) (*attachmentpb.CreateAttachmentResponse, error))
			deleteAttachment, _ := ctx.DeleteAttachment.(func(context.Context, *attachmentpb.DeleteAttachmentRequest) (*attachmentpb.DeleteAttachmentResponse, error))
			newAttachmentID, _ := ctx.NewAttachmentID.(func() string)

			// Wire use cases — scheduling engine is not yet fully wired in espyna
			// so we always provide stub fallbacks (matching domain_schedule.go).
			deps := &eventmod.ModuleDeps{
				Routes:       eventRoutes,
				Labels:       eventLabels,
				CommonLabels: ctx.Common,
				TableLabels:  ctx.Table,

				// Stub fallbacks — always provided so the calendar view works
				// without a real scheduling backend.
				CreateEvent: func(_ context.Context, _ *eventpb.CreateEventRequest) (*eventpb.CreateEventResponse, error) {
					return nil, nil
				},
				ListEvents: func(_ context.Context, _ *eventpb.ListEventsRequest) (*eventpb.ListEventsResponse, error) {
					return nil, nil
				},
			}

			// Overlay with the typed use cases (compile-checked; nil leaf
			// closures keep the stub fallbacks above).
			wireEventDeps(deps, uc)
			// Schedule dashboard (nil-safe: only wires if the typed slot is set).
			wireScheduleDashboard(deps, uc)

			// Wire attachment ops (nil-safe — degrade gracefully when not provided).
			deps.UploadFile = uploadFile
			deps.ListAttachments = listAttachments
			deps.CreateAttachment = createAttachment
			deps.DeleteAttachment = deleteAttachment
			deps.NewID = newAttachmentID

			eventmod.NewModule(deps).RegisterRoutes(ctx.Routes)
		}

		// --- Register EventTag module ---
		if cfg.wantEventTag() {
			// Load routes (defaults + optional lyngua overrides).
			eventTagRoutes := event.DefaultEventTagRoutes()
			_ = translations.LoadPathIfExists("en", ctx.BusinessType, "route.json", "event_tag", &eventTagRoutes)

			// Load labels — event_tag.json has no root wrap (flat keys), so we
			// pass "" as the dotPath. Falls back silently to zero values if
			// the file is absent (e.g. for a tier that hasn't localized yet).
			var eventTagLabels event.EventTagLabels
			if err := translations.LoadPathIfExists("en", ctx.BusinessType, "event_tag.json", "", &eventTagLabels); err != nil {
				log.Printf("Warning: Failed to load event_tag labels: %v", err)
			}

			eventTagDeps := &eventtagmod.ModuleDeps{
				Routes:       eventTagRoutes,
				Labels:       eventTagLabels,
				CommonLabels: ctx.Common,
				TableLabels:  ctx.Table,
			}

			// Overlay with the typed use cases (compile-checked).
			wireEventTagDeps(eventTagDeps, uc)

			// Reference-checker for the delete-guard. Optional — if not
			// wired, the list page simply renders without the in-use tooltip.
			if ctx.RefChecker != nil {
				if refChecker, ok := ctx.RefChecker.(reference.Checker); ok && refChecker != nil {
					eventTagDeps.GetEventTagInUseIDs = refChecker.GetEventTagInUseIDs
				}
			}

			eventtagmod.NewModule(eventTagDeps).RegisterRoutes(ctx.Routes)
		}

		log.Println("  ✓ Schedule domain initialized (cyta)")
		return nil
	}
}

// defaultEventLabels returns EventLabels with sensible English defaults.
// Cyta has no Default*Labels() function in its root package so we define
// the defaults here to make Block() self-contained.
func defaultEventLabels() event.EventLabels {
	return event.EventLabels{
		Page: event.EventPageLabels{
			Heading:          "Events",
			HeadingUpcoming:  "Upcoming Events",
			HeadingConfirmed: "Confirmed Events",
			HeadingCompleted: "Completed Events",
			HeadingCancelled: "Cancelled Events",
			Caption:          "Manage scheduled events and appointments",
		},
		Buttons: event.EventButtonLabels{
			AddEvent: "Add Event",
		},
		Columns: event.EventColumnLabels{
			Name:      "Name",
			StartDate: "Start",
			EndDate:   "End",
			Organizer: "Organizer",
			Location:  "Location",
			Status:    "Status",
			Recurs:    "Recurs",
		},
		Empty: event.EventEmptyLabels{
			Heading:    "No events found",
			Subheading: "No events to display.",
		},
		Form: event.EventFormLabels{
			Name:        "Name",
			Description: "Description",
			StartDate:   "Start Date",
			EndDate:     "End Date",
			Timezone:    "Timezone",
			AllDay:      "All Day",
			Organizer:   "Organizer",
			Location:    "Location",
			Status:      "Status",
			Recurrence:  "Recurrence",
		},
		Actions: event.EventActionLabels{
			Edit:      "Edit",
			Delete:    "Delete",
			Cancel:    "Cancel",
			Confirm:   "Confirm",
			Duplicate: "Duplicate",
		},
		Detail: event.EventDetailLabels{
			Heading:     "Event Details",
			Overview:    "Overview",
			Organizer:   "Organizer",
			Location:    "Location",
			Duration:    "Duration",
			TimeRange:   "Time",
			Status:      "Status",
			AllDay:      "All Day",
			Recurrence:  "Recurrence",
			Description: "Description",
		},
		Tabs: event.EventTabLabels{
			Overview:    "Overview",
			Attendees:   "Attendees",
			Resources:   "Resources",
			Products:    "Products",
			Occurrences: "Occurrences",
			Attachments: "Attachments",
		},
		Confirm: event.EventConfirmLabels{
			DeleteTitle:   "Delete Event",
			DeleteMessage: "Are you sure you want to delete this event? This action cannot be undone.",
			CancelTitle:   "Cancel Event",
			CancelMessage: "Are you sure you want to cancel this event?",
		},
		Errors: event.EventErrorLabels{
			NameRequired:      "Event name is required",
			StartDateRequired: "Start date is required",
			EndDateRequired:   "End date is required",
			InvalidDateRange:  "End date must be after start date",
		},
		Status: event.EventStatusLabels{
			Tentative: "Tentative",
			Confirmed: "Confirmed",
			Cancelled: "Cancelled",
		},
	}
}
