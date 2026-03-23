package detail

import (
	"context"
	"fmt"
	"log"
	"time"

	cyta "github.com/erniealice/cyta-golang"
	"github.com/erniealice/pyeza-golang/types"

	eventattendeepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/event/event_attendee"
	eventoccurrencepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/event/event_occurrence"
	eventproductpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/event/event_product"
	eventresourcepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/event/event_resource"
)

// ---------------------------------------------------------------------------
// Attendees tab
// ---------------------------------------------------------------------------

func loadAttendeesTab(ctx context.Context, deps *DetailViewDeps, pageData *PageData, eventID string) {
	if deps.ListEventAttendees == nil {
		return
	}

	resp, err := deps.ListEventAttendees(ctx, &eventattendeepb.ListEventAttendeesRequest{})
	if err != nil {
		log.Printf("Failed to list event attendees for event %s: %v", eventID, err)
		return
	}

	var attendees []*eventattendeepb.EventAttendee
	for _, a := range resp.GetData() {
		if a.GetEventId() == eventID {
			attendees = append(attendees, a)
		}
	}

	l := deps.Labels
	pageData.AttendeesTable = buildAttendeesTable(attendees, l, deps.TableLabels)
}

func buildAttendeesTable(
	attendees []*eventattendeepb.EventAttendee,
	l cyta.EventLabels,
	tableLabels types.TableLabels,
) *types.TableConfig {
	columns := []types.TableColumn{
		{Key: "name", Label: l.Columns.Name, Sortable: true},
		{Key: "role", Label: l.Detail.Overview, Sortable: true, Width: "120px"},
		{Key: "status", Label: l.Columns.Status, Sortable: true, Width: "120px"},
	}

	rows := []types.TableRow{}
	for _, a := range attendees {
		id := a.GetId()
		name := a.GetDisplayName()
		role := attendeeRoleString(a.GetRole())
		status := attendeeStatusString(a.GetStatus())

		rows = append(rows, types.TableRow{
			ID: id,
			Cells: []types.TableCell{
				{Type: "text", Value: name},
				{Type: "badge", Value: role, Variant: "info"},
				{Type: "badge", Value: status, Variant: attendeeStatusVariant(status)},
			},
			DataAttrs: map[string]string{
				"name":   name,
				"role":   role,
				"status": status,
			},
		})
	}

	types.ApplyColumnStyles(columns, rows)

	return &types.TableConfig{
		ID:         "attendees-table",
		Columns:    columns,
		Rows:       rows,
		ShowSearch: true,
		ShowSort:   true,
		Labels:     tableLabels,
		EmptyState: types.TableEmptyState{
			Title:   "No attendees",
			Message: "No attendees have been added to this event yet.",
		},
	}
}

func attendeeRoleString(r eventattendeepb.AttendeeRole) string {
	switch r {
	case eventattendeepb.AttendeeRole_ATTENDEE_ROLE_CHAIR:
		return "chair"
	case eventattendeepb.AttendeeRole_ATTENDEE_ROLE_REQUIRED:
		return "required"
	case eventattendeepb.AttendeeRole_ATTENDEE_ROLE_OPTIONAL:
		return "optional"
	case eventattendeepb.AttendeeRole_ATTENDEE_ROLE_RESOURCE:
		return "resource"
	default:
		return "optional"
	}
}

func attendeeStatusString(s eventattendeepb.AttendeeStatus) string {
	switch s {
	case eventattendeepb.AttendeeStatus_ATTENDEE_STATUS_NEEDS_ACTION:
		return "needs_action"
	case eventattendeepb.AttendeeStatus_ATTENDEE_STATUS_ACCEPTED:
		return "accepted"
	case eventattendeepb.AttendeeStatus_ATTENDEE_STATUS_DECLINED:
		return "declined"
	case eventattendeepb.AttendeeStatus_ATTENDEE_STATUS_TENTATIVE:
		return "tentative"
	default:
		return "needs_action"
	}
}

func attendeeStatusVariant(status string) string {
	switch status {
	case "needs_action":
		return "warning"
	case "accepted":
		return "success"
	case "declined":
		return "danger"
	case "tentative":
		return "default"
	default:
		return "default"
	}
}

// ---------------------------------------------------------------------------
// Resources tab
// ---------------------------------------------------------------------------

func loadResourcesTab(ctx context.Context, deps *DetailViewDeps, pageData *PageData, eventID string) {
	if deps.ListEventResources == nil {
		return
	}

	resp, err := deps.ListEventResources(ctx, &eventresourcepb.ListEventResourcesRequest{})
	if err != nil {
		log.Printf("Failed to list event resources for event %s: %v", eventID, err)
		return
	}

	var resources []*eventresourcepb.EventResource
	for _, r := range resp.GetData() {
		if r.GetEventId() == eventID {
			resources = append(resources, r)
		}
	}

	l := deps.Labels
	pageData.ResourcesTable = buildResourcesTable(resources, l, deps.TableLabels)
}

func buildResourcesTable(
	resources []*eventresourcepb.EventResource,
	l cyta.EventLabels,
	tableLabels types.TableLabels,
) *types.TableConfig {
	columns := []types.TableColumn{
		{Key: "name", Label: l.Columns.Name, Sortable: true},
		{Key: "type", Label: l.Detail.Overview, Sortable: true, Width: "120px"},
		{Key: "status", Label: l.Columns.Status, Sortable: true, Width: "120px"},
	}

	rows := []types.TableRow{}
	for _, r := range resources {
		id := r.GetId()
		name := r.GetName()
		resourceType := resourceTypeString(r.GetResourceType())
		status := resourceStatusString(r.GetStatus())

		rows = append(rows, types.TableRow{
			ID: id,
			Cells: []types.TableCell{
				{Type: "text", Value: name},
				{Type: "text", Value: resourceType},
				{Type: "badge", Value: status, Variant: resourceStatusVariant(status)},
			},
			DataAttrs: map[string]string{
				"name":   name,
				"type":   resourceType,
				"status": status,
			},
		})
	}

	types.ApplyColumnStyles(columns, rows)

	return &types.TableConfig{
		ID:         "resources-table",
		Columns:    columns,
		Rows:       rows,
		ShowSearch: true,
		ShowSort:   true,
		Labels:     tableLabels,
		EmptyState: types.TableEmptyState{
			Title:   "No resources",
			Message: "No resources have been assigned to this event yet.",
		},
	}
}

func resourceTypeString(t eventresourcepb.ResourceType) string {
	switch t {
	case eventresourcepb.ResourceType_RESOURCE_TYPE_ROOM:
		return "room"
	case eventresourcepb.ResourceType_RESOURCE_TYPE_EQUIPMENT:
		return "equipment"
	case eventresourcepb.ResourceType_RESOURCE_TYPE_STAFF:
		return "staff"
	default:
		return "equipment"
	}
}

func resourceStatusString(s eventresourcepb.ResourceStatus) string {
	switch s {
	case eventresourcepb.ResourceStatus_RESOURCE_STATUS_ASSIGNED:
		return "assigned"
	case eventresourcepb.ResourceStatus_RESOURCE_STATUS_CONFIRMED:
		return "confirmed"
	case eventresourcepb.ResourceStatus_RESOURCE_STATUS_RELEASED:
		return "released"
	default:
		return "assigned"
	}
}

func resourceStatusVariant(status string) string {
	switch status {
	case "assigned":
		return "warning"
	case "confirmed":
		return "success"
	case "released":
		return "default"
	default:
		return "default"
	}
}

// ---------------------------------------------------------------------------
// Products tab
// ---------------------------------------------------------------------------

func loadProductsTab(ctx context.Context, deps *DetailViewDeps, pageData *PageData, eventID string) {
	if deps.ListEventProducts == nil {
		return
	}

	resp, err := deps.ListEventProducts(ctx, &eventproductpb.ListEventProductsRequest{})
	if err != nil {
		log.Printf("Failed to list event products for event %s: %v", eventID, err)
		return
	}

	var products []*eventproductpb.EventProduct
	for _, p := range resp.GetData() {
		if p.GetEventId() == eventID {
			products = append(products, p)
		}
	}

	l := deps.Labels
	pageData.ProductsTable = buildProductsTable(products, l, deps.TableLabels)
}

func buildProductsTable(
	products []*eventproductpb.EventProduct,
	l cyta.EventLabels,
	tableLabels types.TableLabels,
) *types.TableConfig {
	columns := []types.TableColumn{
		{Key: "name", Label: l.Columns.Name, Sortable: true},
		{Key: "quantity", Label: l.Detail.Overview, Sortable: false, Width: "80px"},
		{Key: "unit_price", Label: l.Detail.Overview, Sortable: false, Width: "120px"},
		{Key: "total_price", Label: l.Detail.Overview, Sortable: false, Width: "120px"},
	}

	rows := []types.TableRow{}
	for _, p := range products {
		id := p.GetId()
		productName := ""
		if prod := p.GetProduct(); prod != nil {
			productName = prod.GetName()
		}
		if productName == "" {
			productName = p.GetProductId()
		}
		quantity := fmt.Sprintf("%d", p.GetQuantity())
		unitPrice := fmt.Sprintf("%.2f %s", p.GetUnitPrice(), p.GetCurrency())
		totalPrice := fmt.Sprintf("%.2f %s", p.GetTotalPrice(), p.GetCurrency())

		rows = append(rows, types.TableRow{
			ID: id,
			Cells: []types.TableCell{
				{Type: "text", Value: productName},
				{Type: "text", Value: quantity},
				{Type: "text", Value: unitPrice},
				{Type: "text", Value: totalPrice},
			},
			DataAttrs: map[string]string{
				"name":        productName,
				"quantity":    quantity,
				"unit_price":  unitPrice,
				"total_price": totalPrice,
			},
		})
	}

	types.ApplyColumnStyles(columns, rows)

	return &types.TableConfig{
		ID:         "products-table",
		Columns:    columns,
		Rows:       rows,
		ShowSearch: true,
		ShowSort:   true,
		Labels:     tableLabels,
		EmptyState: types.TableEmptyState{
			Title:   "No products",
			Message: "No products have been associated with this event yet.",
		},
	}
}

// ---------------------------------------------------------------------------
// Occurrences tab
// ---------------------------------------------------------------------------

func loadOccurrencesTab(ctx context.Context, deps *DetailViewDeps, pageData *PageData, eventID string) {
	if deps.ListEventOccurrences == nil {
		return
	}

	resp, err := deps.ListEventOccurrences(ctx, &eventoccurrencepb.ListEventOccurrencesRequest{})
	if err != nil {
		log.Printf("Failed to list event occurrences for event %s: %v", eventID, err)
		return
	}

	var occurrences []*eventoccurrencepb.EventOccurrence
	for _, o := range resp.GetData() {
		if o.GetEventId() == eventID {
			occurrences = append(occurrences, o)
		}
	}

	l := deps.Labels
	pageData.OccurrencesTable = buildOccurrencesTable(occurrences, l, deps.TableLabels)
}

func buildOccurrencesTable(
	occurrences []*eventoccurrencepb.EventOccurrence,
	l cyta.EventLabels,
	tableLabels types.TableLabels,
) *types.TableConfig {
	columns := []types.TableColumn{
		{Key: "start_date", Label: l.Columns.StartDate, Sortable: true, Width: "150px"},
		{Key: "end_date", Label: l.Columns.EndDate, Sortable: true, Width: "150px"},
		{Key: "exception", Label: l.Detail.Overview, Sortable: false, Width: "100px"},
		{Key: "cancelled", Label: l.Columns.Status, Sortable: false, Width: "100px"},
	}

	rows := []types.TableRow{}
	for _, o := range occurrences {
		id := o.GetId()
		startDate := epochToString(o.GetStartDateTimeUtc())
		endDate := epochToString(o.GetEndDateTimeUtc())
		isException := "No"
		if o.GetIsException() {
			isException = "Yes"
		}
		isCancelled := "No"
		cancelledVariant := "success"
		if o.GetIsCancelled() {
			isCancelled = "Yes"
			cancelledVariant = "default"
		}

		rows = append(rows, types.TableRow{
			ID: id,
			Cells: []types.TableCell{
				types.DateTimeCell(startDate, types.DateReadable),
				types.DateTimeCell(endDate, types.DateReadable),
				{Type: "text", Value: isException},
				{Type: "badge", Value: isCancelled, Variant: cancelledVariant},
			},
			DataAttrs: map[string]string{
				"start_date":   startDate,
				"end_date":     endDate,
				"is_exception": isException,
				"is_cancelled": isCancelled,
			},
		})
	}

	types.ApplyColumnStyles(columns, rows)

	return &types.TableConfig{
		ID:         "occurrences-table",
		Columns:    columns,
		Rows:       rows,
		ShowSearch: false,
		ShowSort:   true,
		Labels:     tableLabels,
		EmptyState: types.TableEmptyState{
			Title:   "No occurrences",
			Message: "This event has no recurring occurrences.",
		},
	}
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// epochToString converts a Unix epoch (seconds) to a human-readable date string.
func epochToString(epoch int64) string {
	if epoch == 0 {
		return ""
	}
	return time.Unix(epoch, 0).UTC().Format("2006-01-02 15:04:05")
}
