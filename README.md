# cyta-golang

Scheduling domain package for Ichizen OS. Provides reusable views, templates, route configurations, and label structs for events, calendar, and recurrence patterns.

**Module path:** `github.com/erniealice/cyta-golang`

## Dependencies

- `github.com/erniealice/pyeza-golang` -- UI framework (view system, template engine, types)
- `github.com/erniealice/esqyma` -- Proto schemas (event, event_attendee, event_occurrence, event_product, event_resource)
- `github.com/erniealice/lyngua` -- Translation/i18n
- `github.com/teambition/rrule-go` -- RFC 5545 RRULE parsing and expansion

## Package Structure

```
cyta-golang/
  go.mod
  routes.go             # Route path constants (EventListURL, CalendarURL, RecurrenceListURL, etc.)
  routes_config.go      # EventRoutes, RecurrenceRoutes structs + Default*Routes() constructors
  labels.go             # EventLabels, RecurrenceLabels, CalendarLabels structs
  block/
    block.go            # Block() constructor — wires espyna use cases into views
    wiring.go           # Reflection-based use-case extraction from opaque *usecases.Aggregate
  services/
    recurrence/
      expander.go       # Expander — expands RRULE strings into Occurrence timestamps (pure, no DB)
      expander_test.go
      occurrence_writer.go
    availability/
      engine.go         # Availability conflict detection engine
      engine_test.go
      conflicts.go
  views/
    event/
      module.go         # ModuleDeps struct (use-case function fields), RenderModule()
      list/page.go      # Event list view
      detail/page.go    # Event detail view with tabs
      action/actions.go # CRUD + status + bulk action handlers
      calendar/page.go  # Calendar view (month/week/day)
      calendar/templates/
```

## Key Exports

- `EventRoutes` / `DefaultEventRoutes()` — configurable URL paths for event CRUD and calendar
- `RecurrenceRoutes` / `DefaultRecurrenceRoutes()` — configurable URL paths for recurrence patterns
- `EventLabels`, `RecurrenceLabels`, `CalendarLabels` — translatable string structs loaded from lyngua
- `services/recurrence.Expander` — stateless RFC 5545 RRULE expander; produces `[]Occurrence` from a rule string and time window
- `services/availability.Engine` — detects scheduling conflicts between events

## Role in the Monorepo

cyta-golang sits in the domain layer above pyeza (UI framework) and espyna (backend framework). Consumer apps (e.g., `apps/service-admin`) call `block.Block()` to mount the scheduling module into their router. The `block/` sub-package uses reflection to extract execute methods from espyna's `usecases.Aggregate` so that cyta does not take a direct compile-time dependency on espyna.

## Centavo Convention

This package does not handle financial amounts directly. Scheduling-related costs (if any) are delegated to centymo or fycha.
