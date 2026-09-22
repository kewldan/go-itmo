package myitmo_test

import (
	"net/http"
	"testing"

	"github.com/kewldan/go-itmo/myitmo"
)

func TestSystemDashboard(t *testing.T) {
	f, c := newFake(t)
	f.result(`[{"i":"1","x":0,"y":0,"w":4,"h":2,"widget":"ScholarshipWidget","static":false,"min_w":2,"max_w":null,"min_h":null,"max_h":4,
		"is_draggable":true,"is_resizable":false,"preserve_aspect_ratio":null,"drag_allow_from":".handle","drag_ignore_from":"","resize_ignore_from":""}]`)
	items := must[[]myitmo.DashboardItem](t)(c.System.Dashboard(ctx))
	f.expect(http.MethodGet, "/api/system/dashboard")
	if len(items) != 1 {
		t.Fatalf("items = %+v", items)
	}
	it := items[0]
	if it.Widget != "ScholarshipWidget" || it.W != 4 || *it.MinW != 2 || it.MaxW != nil || *it.MaxH != 4 ||
		!*it.IsDraggable || *it.IsResizable || it.PreserveAspectRatio != nil || it.DragAllowFrom != ".handle" {
		t.Errorf("item = %+v", it)
	}
}

func TestSystemMenu(t *testing.T) {
	f, c := newFake(t)
	f.result(`{"menu":[{"title":"menu.finances","icon":"icon-legacy wallet","to":"/finances","exact":false,
		"children":[{"title":"menu.finances.scholarship","to":"/finances/scholarship","exact":true}]}]}`)
	menu := must[*myitmo.MenuResponse](t)(c.System.Menu(ctx))
	f.expect(http.MethodGet, "/api/system/v1/menu/items")
	if len(menu.Menu) != 1 || menu.Menu[0].Title != "menu.finances" || len(menu.Menu[0].Children) != 1 || !menu.Menu[0].Children[0].Exact {
		t.Errorf("menu = %+v", menu)
	}

	f.result(`{"menu":[{"title":"services.sport","description":"services.sport.desc","to":"/sport","icon":"run","multicolor_icon":"run-color","color":"#1946ba","exact":false}]}`)
	svc := must[*myitmo.MenuResponse](t)(c.System.Services(ctx))
	f.expect(http.MethodGet, "/api/system/v1/menu/services")
	if len(svc.Menu) != 1 || svc.Menu[0].MulticolorIcon != "run-color" || svc.Menu[0].Color != "#1946ba" || svc.Menu[0].Description == "" {
		t.Errorf("services = %+v", svc)
	}
}
