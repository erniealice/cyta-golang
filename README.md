# cyta-golang

Scheduling domain package for Ichizen OS. Owns exactly **one esqyma domain — `event`** — making it the textbook single-domain exemplar (alongside centymo).

**Module path:** `github.com/erniealice/cyta-golang`

## Domain ownership

cyta maps 1:1 to the esqyma proto domain `event` (`proto/v1/domain/event/`). It renders views for the event entity and its sub-entities (event_attendee, event_attribute, event_client, event_occurrence, event_product, event_recurrence, event_resource, event_tag). No other proto domain lives here.

## Package structure

```
cyta-golang/
  placement_test.go        # STRICT placement gate (see below) — the ONLY file at root
  go.mod
  go.sum
  domain/
    event/                 # package event — all code for the event domain
      labels.go            # EventLabels struct
      routes.go            # route path constants
      routes_const.go      # EventRoutes struct + DefaultEventRoutes()
      routes_test.go
      views/
        event/             # views for the event entity
          module.go        # ModuleDeps + wiring
          embed.go
          list/page.go
          detail/page.go
          detail/tabs.go
          detail/deps.go
          detail/attachment.go
          action/actions.go
          action/handlers_save.go
          calendar/page.go
          calendar/embed.go
          calendar/templates/calendar.html
          dashboard/page.go
          form/data.go
          form/options.go
          templates/        # dashboard.html, detail.html, event-drawer-form.html, …
        event_tag/         # views for the event_tag sub-entity
          module.go
          embed.go
          list/page.go
          action/action.go
          form/form.go
          templates/        # event-tag-drawer-form.html, list.html
  block/
    block.go               # Block() constructor — wires espyna use cases into views
    usecases.go
    wiring.go              # reflection-based use-case extraction from *usecases.Aggregate
  services/
    recurrence/
      expander.go          # stateless RFC 5545 RRULE expander → []Occurrence
      expander_test.go
      occurrence_writer.go
      occurrence_writer_test.go
    availability/
      engine.go            # conflict-detection engine
      engine_test.go
      conflicts.go
```

## Placement gate (`placement_test.go`)

cyta carries a **STRICT** placement gate (`legacyAllow` is empty — zero migration debt). The gate runs as `go test ./...` and enforces four rules:

| Rule | What it checks |
|------|---------------|
| **R1** Empty root | No package `.go` files at module root — only `_test.go` permitted |
| **R2** Canonical esqyma domains | Every `domain/<x>/` must match a real esqyma proto domain; no invented namespaces |
| **R3** Entity placement | `Labels`/`Routes` types resolve to the esqyma entity→domain map; an `event` type may not live inside a non-`event` domain directory |
| **R4** No god-files | No `.go` file (excl. `_test.go`) may exceed 1,200 lines — split per entity |

`crossCutting = false` — the domain variant applies. esqyma's `proto/v1/domain/` is located at test time so the rules never drift from the live proto tree.

## Private services

`services/recurrence` and `services/availability` are chartered private helpers under `services/` (an allowed first-level directory). They are not exported as a separate module. `services/recurrence` imports `github.com/teambition/rrule-go` for RFC 5545 RRULE parsing and expansion.

## Dependencies

- `github.com/erniealice/pyeza-golang` — UI framework (view system, template engine, types)
- `github.com/erniealice/esqyma` — proto schemas (event domain)
- `github.com/erniealice/lyngua` — translation/i18n
- `github.com/teambition/rrule-go` — RFC 5545 RRULE parsing and expansion

## Role in the monorepo

cyta sits in the domain layer above pyeza and espyna. Consumer apps (e.g., `apps/service-admin`) call `block.Block()` to mount the scheduling module. `block/wiring.go` uses reflection to extract use-case methods from espyna's opaque `*usecases.Aggregate`, so cyta does not take a direct compile-time dependency on espyna.

See `docs/wiki/articles/vertical-slices.md` for the full entity trace and `docs/wiki/articles/package-map.md` for the monorepo dependency graph.
