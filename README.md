# cyta-golang

Scheduling domain package for Ichizen OS. Owns exactly **one esqyma domain — `event`** — making it the textbook single-domain exemplar.

**Module path:** `github.com/erniealice/cyta-golang`

## Domain ownership

cyta maps 1:1 to the esqyma proto domain `event` (`proto/v1/domain/event/`). It renders views for the event entity and its sub-entities (event_attendee, event_attribute, event_client, event_occurrence, event_product, event_recurrence, event_resource, event_tag, event_tag_assignment). No other proto domain lives here.

## Package structure (Option B)

Under Option B the ENTITY is the contract package. Each `domain/<d>/<e>/` directory is one esqyma entity. The domain facade (`domain/<d>/<d>.go`) re-exports entity-local types as Go type aliases so consumers never change their import paths.

```
cyta-golang/
  placement_test.go            # B-STRICT placement gate — the ONLY file at root
  go.mod / go.sum
  domain/
    event/                     # package event — facade for the event domain
      event.go                 # facade: type EventLabels = event.Labels, etc. (aliases only)
      event_module.go          # NewModule() assembler for the event entity
      event_tag_module.go      # NewEventTagModule() assembler for the event_tag entity
      event/                   # entity: esqyma event/event
        labels.go              # Labels struct (EventLabels in the facade)
        routes.go              # Routes struct + DefaultRoutes()
        embed.go               # template embed.FS
        list/page.go           # list page handler
        detail/
          page.go              # detail page handler
          tabs.go              # tab partial handlers
          deps.go              # ModuleDeps
          attachment.go        # attachment tab handler
        action/
          actions.go           # CRUD action handlers (add / edit / delete / set-status)
          handlers_save.go     # save handler for the event drawer form
        calendar/
          page.go              # calendar page handler
          embed.go             # calendar template embed.FS
          templates/calendar.html
        dashboard/page.go      # schedule dashboard handler
        form/
          data.go              # form.Data + form.FormData structs
          options.go           # BuildStatusOptions and other form helpers
        templates/             # dashboard.html, detail.html, event-drawer-form.html, …
      event_tag/               # entity: esqyma event/event_tag
        labels.go              # Labels struct
        routes.go              # Routes struct + DefaultRoutes()
        embed.go               # template embed.FS
        action/action.go       # CRUD action handlers
        form/form.go           # form.Data struct
        list/page.go           # list page handler
        templates/             # event-tag-drawer-form.html, list.html
      calendar/                # domain-level calendar view (legacyAllow — dir rename pending)
        labels.go              # Labels struct for the calendar view surface
      recurrence/              # esqyma entity is event_recurrence (legacyAllow — dir rename pending)
        labels.go              # Labels struct for the recurrence pattern view
        routes.go              # Routes struct + DefaultRoutes()
  block/
    block.go                   # Block() constructor — pyeza.AppOption entry point
    usecases.go                # *UseCases typed wiring contract + RequireFor + MustValidate
    wiring.go                  # assigns *UseCases closures onto view ModuleDeps
    block_test.go              # MustValidate fail-closed wiring tests
  services/
    recurrence/
      expander.go              # stateless RFC 5545 RRULE expander → []Occurrence
      occurrence_writer.go     # writes expanded occurrences to the DB
    availability/
      engine.go                # conflict-detection engine
      conflicts.go             # conflict predicate helpers
```

## Placement gate (`placement_test.go`)

cyta carries a **B-STRICT** placement gate (v2, Option B). `legacyAllow` holds two dated residuals pending dir renames; the target state is empty (STRICT).

| Rule | What it checks |
|------|----------------|
| **R1** Empty root | No package `.go` files at module root — only `_test.go` permitted |
| **R2** Canonical dirs | Every first-level dir is an allowed infra surface; every `domain/<d>` is an esqyma proto domain |
| **R2′** Entity dirs | Every `domain/<d>/<child>/` DIR is an esqyma entity of domain `<d>`, `shared`, or a domain-view (name starts with `<d>`) |
| **R3′** Entity contract | No real `*Labels`/`*Routes` type declaration at the domain root — only alias re-exports (`type X = pkg.Y`) are allowed |
| **R4** No god-files | No `.go` file (excl. `_test.go`) may exceed 1,200 lines |
| **R5** Facade exists | A facade `domain/<d>/<d>.go` must exist for every domain dir with ≥1 entity subdir |
| **R6** No cycles | Enforced by `lint-no-domain-cycles.sh` (external, go-list based) |

`crossCutting = false` — the domain variant applies. esqyma's `proto/v1/domain/` is located at test time so the rules never drift from the live proto tree.

Current `legacyAllow` residuals (both EXPIRES 2026-07-15):
- `domain/event/calendar` — dir name does not start with `event`; pending rename to `eventcalendar`
- `domain/event/recurrence` — esqyma entity is `event_recurrence`; pending rename

## Fail-closed wiring (`block/usecases.go`)

`*UseCases` is the typed wiring contract between service-admin's composition layer and cyta's view modules. `RequireFor(cfg)` lists every missing REQUIRED closure for the enabled modules. `MustValidate(cfg)` adds fail-closed posture:

- **dev/test** (`testing.Testing()` true or `CYTA_BLOCK_STRICT` truthy): PANIC with the full field list — uncatchable-by-accident, stack-traced, fails CI loudly.
- **prod**: `log.Printf("FATAL: ...")` at the seam AND returns the error → `Block()` propagates → `NewServiceAdmin` halts boot.

OPTIONAL closures (nested-entity lists, derived picker closures, the schedule dashboard) are never flagged — they degrade gracefully to empty-state.

## Private services

`services/recurrence` and `services/availability` are chartered private helpers under `services/` (an allowed first-level directory). They are not exported as a separate module. `services/recurrence` imports `github.com/teambition/rrule-go` for RFC 5545 RRULE parsing and expansion.

## Dependencies

- `github.com/erniealice/pyeza-golang` — UI framework (view system, template engine, types)
- `github.com/erniealice/esqyma` — proto schemas (event domain)
- `github.com/erniealice/lyngua` — translation/i18n
- `github.com/teambition/rrule-go` — RFC 5545 RRULE parsing and expansion

## Role in the monorepo

cyta sits in the domain layer above pyeza and espyna. Consumer apps (e.g., `apps/service-admin`) call `block.Block()` to mount the scheduling module, supplying a `*UseCases` via `block.WithUseCases(...)`. The typed contract ensures any drift between espyna and cyta is a compile error, not a silent nil.

See `docs/wiki/articles/vertical-slices.md` for the full entity trace and `docs/wiki/articles/package-map.md` for the monorepo dependency graph.
