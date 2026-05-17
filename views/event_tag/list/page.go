// Package list implements the event-tag list page with active/inactive tabs.
//
// Mirrors packages/entydad-golang/views/role/list/page.go but without the
// permissions matrix column and with a color-swatch column. Because the
// EventTagListURL does not include a {status} path param, the active/inactive
// tab is selected via the ?status= query string.
package list

import (
	"context"
	"fmt"
	"log"
	"math"
	"strconv"

	espynahttp "github.com/erniealice/espyna-golang/contrib/http"
	"github.com/erniealice/espyna-golang/tableparams"
	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/pyeza-golang/route"
	"github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"

	commonpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/common"
	eventtagpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/event/event_tag"

	cyta "github.com/erniealice/cyta-golang"
	lynguaV1 "github.com/erniealice/lyngua/golang/v1"
)

// Deps holds view dependencies.
type Deps struct {
	Routes                  cyta.EventTagRoutes
	Labels                  cyta.EventTagLabels
	CommonLabels            pyeza.CommonLabels
	TableLabels             types.TableLabels
	GetEventTagListPageData func(ctx context.Context, req *eventtagpb.GetEventTagListPageDataRequest) (*eventtagpb.GetEventTagListPageDataResponse, error)
	GetEventTagInUseIDs     func(ctx context.Context, ids []string) (map[string]bool, error)
}

// PageData holds the data for the event-tag list page.
type PageData struct {
	types.PageData
	ContentTemplate string
	Table           *types.TableConfig
}

var eventTagSearchFields = []string{"name", "description"}

// NewView creates the event-tag list view (full page).
func NewView(deps *Deps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		if !view.GetUserPermissions(ctx).Can("event_tag", "list") {
			return view.Forbidden("event_tag:list")
		}

		status := parseStatus(viewCtx)

		columns := eventTagColumns(deps.Labels)
		p, err := espynahttp.ParseTableParamsWithFilters(viewCtx.Request, types.SortableKeys(columns), types.FilterableKeys(columns), "name", "asc")
		if err != nil {
			return view.Error(err)
		}

		tableConfig, err := buildTableConfig(ctx, deps, status, p, columns)
		if err != nil {
			return view.Error(err)
		}

		pageData := &PageData{
			PageData: types.PageData{
				CacheVersion:   viewCtx.CacheVersion,
				Title:          statusPageTitle(deps.Labels, status),
				CurrentPath:    viewCtx.CurrentPath,
				ActiveNav:      "schedule",
				ActiveSubNav:   "event-tags-" + status,
				HeaderTitle:    statusPageTitle(deps.Labels, status),
				HeaderSubtitle: statusPageCaption(deps.Labels, status),
				HeaderIcon:     "icon-tag",
				CommonLabels:   deps.CommonLabels,
			},
			ContentTemplate: "event-tag-list-content",
			Table:           tableConfig,
		}

		// KB help content
		if viewCtx.Translations != nil {
			if provider, ok := viewCtx.Translations.(*lynguaV1.TranslationProvider); ok {
				if kb, _ := provider.LoadKBIfExists(viewCtx.Lang, viewCtx.BusinessType, "event_tag"); kb != nil {
					pageData.HasHelp = true
					pageData.HelpContent = kb.Body
				}
			}
		}

		return view.OK("event-tag-list", pageData)
	})
}

// parseStatus returns "active" by default; "inactive" when ?status=inactive.
func parseStatus(viewCtx *view.ViewContext) string {
	s := viewCtx.Request.URL.Query().Get("status")
	if s != "inactive" {
		return "active"
	}
	return "inactive"
}

// buildTableConfig fetches event-tag data and builds the table configuration.
func buildTableConfig(ctx context.Context, deps *Deps, status string, p tableparams.TableQueryParams, columns []types.TableColumn) (*types.TableConfig, error) {
	perms := view.GetUserPermissions(ctx)

	listParams := espynahttp.ToListParams(p, eventTagSearchFields)

	// Inject active/inactive filter server-side so pagination counts are correct.
	// Table alias in the event_tag adapter is "et" (see
	// packages/espyna-golang/contrib/postgres/internal/adapter/event/event_tag.go).
	activeValue := status != "inactive"
	if listParams.Filters == nil {
		listParams.Filters = &commonpb.FilterRequest{}
	}
	listParams.Filters.Filters = append(listParams.Filters.Filters, &commonpb.TypedFilter{
		Field: "et.active",
		FilterType: &commonpb.TypedFilter_BooleanFilter{
			BooleanFilter: &commonpb.BooleanFilter{Value: activeValue},
		},
	})

	resp, err := deps.GetEventTagListPageData(ctx, &eventtagpb.GetEventTagListPageDataRequest{
		Search:     listParams.Search,
		Filters:    listParams.Filters,
		Sort:       listParams.Sort,
		Pagination: listParams.Pagination,
	})
	if err != nil {
		log.Printf("Failed to list event_tags: %v", err)
		return nil, fmt.Errorf("failed to load event_tags: %w", err)
	}

	// Check which items are in use (for delete-guard).
	var inUseIDs map[string]bool
	if deps.GetEventTagInUseIDs != nil {
		var itemIDs []string
		for _, item := range resp.GetEventTagList() {
			itemIDs = append(itemIDs, item.GetId())
		}
		inUseIDs, _ = deps.GetEventTagInUseIDs(ctx, itemIDs)
	}

	l := deps.Labels
	rows := buildTableRows(resp.GetEventTagList(), l, deps.CommonLabels, deps.Routes, inUseIDs, perms)
	types.ApplyColumnStyles(columns, rows)

	// Refresh URL preserves the tab by including ?status= in the query.
	refreshURL := deps.Routes.ListURL + "?status=" + status

	totalRows := int(resp.GetPagination().GetTotalItems())
	sp := &types.ServerPagination{
		Enabled:           true,
		Mode:              "offset",
		CurrentPage:       p.Page,
		PageSize:          p.PageSize,
		TotalRows:         totalRows,
		TotalPages:        int(math.Ceil(float64(totalRows) / float64(p.PageSize))),
		SearchQuery:       p.Search,
		SortColumn:        p.SortColumn,
		SortDirection:     p.SortDir,
		FiltersJSON:       p.FiltersRaw,
		PaginationURL:     refreshURL,
		PaginationBodyURL: refreshURL,
	}
	sp.BuildDisplay()

	// Primary "Add Tag" button only shown on the active tab (creating a new
	// record always lands in active state).
	var primaryAction *types.PrimaryAction
	if status == "active" {
		primaryAction = &types.PrimaryAction{
			Label:           l.Buttons.AddTag,
			ActionURL:       deps.Routes.AddURL,
			Icon:            "icon-plus",
			Disabled:        !perms.Can("event_tag", "create"),
			DisabledTooltip: fmt.Sprintf(deps.CommonLabels.Errors.MissingPermission, "event_tag:create"),
		}
	}

	tableConfig := &types.TableConfig{
		ID:                   "event-tags-table",
		RefreshURL:           refreshURL,
		Columns:              columns,
		Rows:                 rows,
		ShowSearch:           true,
		ShowActions:          true,
		ShowFilters:          true,
		ShowSort:             true,
		ShowColumns:          true,
		ShowExport:           true,
		ShowDensity:          true,
		ShowEntries:          true,
		DefaultSortColumn:    "name",
		DefaultSortDirection: "asc",
		Labels:               deps.TableLabels,
		EmptyState: types.TableEmptyState{
			Title:   statusEmptyTitle(l, status),
			Message: statusEmptyMessage(l, status),
		},
		PrimaryAction:    primaryAction,
		ServerPagination: sp,
	}
	types.ApplyTableSettings(tableConfig)

	return tableConfig, nil
}

func eventTagColumns(l cyta.EventTagLabels) []types.TableColumn {
	return []types.TableColumn{
		{Key: "name", Label: l.Columns.Name, MinWidth: "9.375rem"},
		{Key: "description", Label: l.Columns.Description, NoSort: true, MinWidth: "9.375rem"},
		{Key: "color", Label: l.Columns.Color, NoSort: true, NoFilter: true, WidthClass: "col-3xl"},
		{Key: "status", Label: l.Columns.Status, NoSort: true, NoFilter: true, WidthClass: "col-2xl"},
		{Key: "date_created", Label: l.Columns.DateCreated, WidthClass: "col-6xl"},
	}
}

// buildTableRows maps EventTag records into pyeza table rows.
// The color column renders a colored dot + the hex value in monospace; we
// emit the markup as a "raw" HTML cell (pyeza table renders {Type: "html"}
// as-is). If "html" type is unavailable in this pyeza version, falling back
// to a "text" cell is safe — the dot simply won't appear.
func buildTableRows(tags []*eventtagpb.EventTag, l cyta.EventTagLabels, common pyeza.CommonLabels, routes cyta.EventTagRoutes, inUseIDs map[string]bool, perms *types.UserPermissions) []types.TableRow {
	rows := []types.TableRow{}
	for _, t := range tags {
		active := t.GetActive()
		recordStatus := "active"
		if !active {
			recordStatus = "inactive"
		}

		id := t.GetId()
		name := t.GetName()
		description := t.GetDescription()
		color := t.GetColor()

		// Color cell: swatch + monospaced hex. Rendered as HTML so the dot is
		// visible at a glance. The class name `table-color-dot` mirrors the
		// pyeza convention; if absent, the inline `background-color` still
		// paints the dot.
		colorCellHTML := fmt.Sprintf(
			`<span class="table-color-dot" style="display:inline-block;width:0.75rem;height:0.75rem;border-radius:9999px;background-color:%s;vertical-align:middle;margin-right:0.5rem;"></span><code style="font-family:var(--font-mono,monospace);font-size:0.8125rem;">%s</code>`,
			color, color,
		)

		canUpdate := perms.Can("event_tag", "update")
		actions := []types.TableAction{
			{
				Type:            "edit",
				Label:           l.Actions.Edit,
				Action:          "edit",
				URL:             route.ResolveURL(routes.EditURL, "id", id),
				Disabled:        !canUpdate,
				DisabledTooltip: fmt.Sprintf(common.Errors.MissingPermission, "event_tag:update"),
			},
		}
		deleteAction := types.TableAction{
			Type:     "delete",
			Label:    l.Actions.Delete,
			Action:   "delete",
			URL:      route.ResolveURL(routes.DeleteURL, "id", id),
			ItemName: name,
		}
		if inUseIDs[id] {
			deleteAction.Disabled = true
			// TODO: swap the hardcoded English for a translated string once
			// EventTagLabels gains an Errors sub-struct.
			deleteAction.DisabledTooltip = "Cannot delete: in use by one or more events"
		} else if !perms.Can("event_tag", "delete") {
			deleteAction.Disabled = true
			deleteAction.DisabledTooltip = fmt.Sprintf(common.Errors.MissingPermission, "event_tag:delete")
		}
		actions = append(actions, deleteAction)

		rows = append(rows, types.TableRow{
			ID: id,
			Cells: []types.TableCell{
				{Type: "text", Value: name},
				{Type: "text", Value: description},
				{Type: "html", Value: colorCellHTML},
				{Type: "badge", Value: recordStatus, Variant: statusVariant(recordStatus)},
				types.DateTimeCell(t.GetDateCreatedString(), types.DateReadable),
			},
			DataAttrs: map[string]string{
				"name":        name,
				"description": description,
				"color":       color,
				"status":      recordStatus,
				"deletable":   strconv.FormatBool(!inUseIDs[id]),
			},
			Actions: actions,
		})
	}
	return rows
}

func statusVariant(status string) string {
	switch status {
	case "active":
		return "success"
	case "inactive":
		return "warning"
	default:
		return "default"
	}
}

func statusPageTitle(l cyta.EventTagLabels, status string) string {
	switch status {
	case "active":
		return l.Page.HeadingActive
	case "inactive":
		return l.Page.HeadingInactive
	default:
		return l.Page.Heading
	}
}

func statusPageCaption(l cyta.EventTagLabels, status string) string {
	switch status {
	case "active":
		return l.Page.CaptionActive
	case "inactive":
		return l.Page.CaptionInactive
	default:
		return l.Page.Caption
	}
}

func statusEmptyTitle(l cyta.EventTagLabels, status string) string {
	switch status {
	case "inactive":
		return l.Empty.InactiveTitle
	default:
		return l.Empty.ActiveTitle
	}
}

func statusEmptyMessage(l cyta.EventTagLabels, status string) string {
	switch status {
	case "inactive":
		return l.Empty.InactiveMessage
	default:
		return l.Empty.ActiveMessage
	}
}
