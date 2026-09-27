package dashboard

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"
)

// U-04 (plan 20260927-tenant-boundary-hardening Wave 0): the schedule
// dashboard lists event names and aggregates, so it needs event:list and must
// not call the data source when the permission is missing.
func TestDashboard_RequiresEventList(t *testing.T) {
	calls := 0
	deps := &Deps{GetDashboardData: func(context.Context, *Request) (*Response, error) {
		calls++
		return &Response{ByTag: map[string]int64{}}, nil
	}}
	vc := &view.ViewContext{Request: httptest.NewRequest("GET", "/schedule/dashboard", nil)}

	for name, perms := range map[string]*types.UserPermissions{
		"nil permissions":      nil,
		"unrelated permission": types.NewUserPermissions([]string{"client:list"}),
	} {
		ctx := context.Background()
		if perms != nil {
			ctx = view.WithUserPermissions(ctx, perms)
		}
		if got := NewView(deps).Handle(ctx, vc); got.Template != "forbidden" {
			t.Fatalf("%s: want forbidden, got template %q", name, got.Template)
		}
	}
	if calls != 0 {
		t.Fatalf("data source must not be called without event:list, got %d calls", calls)
	}

	ctx := view.WithUserPermissions(context.Background(), types.NewUserPermissions([]string{"event:list"}))
	if got := NewView(deps).Handle(ctx, vc); got.Template == "forbidden" {
		t.Fatal("event:list holder must see the dashboard")
	}
	if calls != 1 {
		t.Fatalf("want 1 data-source call for the permitted request, got %d", calls)
	}
}
