package block

import (
	event "github.com/erniealice/cyta-golang/domain/event"
	eventpkg "github.com/erniealice/cyta-golang/domain/event/event"
	eventtagpkg "github.com/erniealice/cyta-golang/domain/event/event_tag"
	"github.com/erniealice/pyeza-golang/compose"
)

// EventUnit returns a compose.Unit for the event (schedule) module.
// The Mount closure mirrors block.go's event registration block.
func EventUnit(uc *UseCases, infra *Infra) compose.Unit {
	u := eventpkg.Describe()
	u.Mount = func(mc *compose.MountContext) error {
		r := u.Routes.(*eventpkg.Routes)
		l := u.Labels.(*eventpkg.Labels)

		deps := &event.ModuleDeps{
			Routes:       *r,
			Labels:       *l,
			CommonLabels: mc.Common,
			TableLabels:  mc.Table,
		}

		// Wire cross-entity routes if sibling units are present.
		if etRoutes, ok := compose.RoutesOf[*eventtagpkg.Routes](mc, "event.event_tag"); ok {
			deps.EventTagRoutes = *etRoutes
		}

		// Attachment ops (nil-safe — degrade gracefully when not provided).
		deps.UploadFile = infra.UploadFile
		deps.ListAttachments = infra.ListAttachments
		deps.CreateAttachment = infra.CreateAttachment
		deps.DeleteAttachment = infra.DeleteAttachment
		deps.NewID = infra.NewAttachmentID

		wireEventDeps(deps, uc)
		wireScheduleDashboard(deps, uc)
		event.NewModule(deps).RegisterRoutes(mc.Routes)
		return nil
	}
	return u
}

// EventTagUnit returns a compose.Unit for the event-tag module.
// The Mount closure mirrors block.go's event-tag registration block.
func EventTagUnit(uc *UseCases, infra *Infra) compose.Unit {
	u := eventtagpkg.Describe()
	u.Mount = func(mc *compose.MountContext) error {
		r := u.Routes.(*eventtagpkg.Routes)
		l := u.Labels.(*eventtagpkg.Labels)

		deps := &event.EventTagModuleDeps{
			Routes:       *r,
			Labels:       *l,
			CommonLabels: mc.Common,
			TableLabels:  mc.Table,
		}

		wireEventTagDeps(deps, uc)

		if infra.RefChecker != nil {
			deps.GetEventTagInUseIDs = infra.RefChecker.GetEventTagInUseIDs
		}

		event.NewEventTagModule(deps).RegisterRoutes(mc.Routes)
		return nil
	}
	return u
}

// AllUnits returns the complete unit list for the cyta/event domain,
// in the same registration order as Block().
func AllUnits(uc *UseCases, infra *Infra) []compose.Unit {
	return []compose.Unit{
		EventUnit(uc, infra),
		EventTagUnit(uc, infra),
	}
}
