package myitmo_test

import (
	"encoding/json/v2"
	"net/http"
	"testing"

	"github.com/kewldan/go-itmo/myitmo"
)

func TestNavigatorBuildings(t *testing.T) {
	f, c := newFake(t)
	f.reply(http.StatusOK, `[{"building_id":2,"name":"Main building","bounds":[30.30,59.95,30.32,59.96],"bearing":-17.5}]`)
	b := must[[]myitmo.NavigatorBuilding](t)(c.Navigator.Buildings(ctx))
	f.expect(http.MethodGet, "/api/navigator/buildings")
	if len(b) != 1 || b[0].BuildingID != 2 || b[0].Bounds[2] != 30.32 || b[0].Bearing != -17.5 {
		t.Errorf("buildings = %+v", b)
	}
}

func TestNavigatorStyle(t *testing.T) {
	f, c := newFake(t)
	f.reply(http.StatusOK, `{"version":8,"sources":{},"layers":[{"id":"room","type":"fill"}]}`)
	style := must[myitmo.RawJSON](t)(c.Navigator.Style(ctx))
	f.expect(http.MethodGet, "/api/navigator/style")
	var doc struct {
		Version int `json:"version"`
		Layers  []struct {
			ID string `json:"id"`
		} `json:"layers"`
	}
	if err := json.Unmarshal(style, &doc); err != nil || doc.Version != 8 || doc.Layers[0].ID != "room" {
		t.Errorf("style = %s (%v)", style, err)
	}
}

func TestNavigatorTransitions(t *testing.T) {
	f, c := newFake(t)
	f.reply(http.StatusOK, `{"type":"FeatureCollection","features":[
		{"type":"Feature","geometry":{"type":"Point","coordinates":[30.31,59.955]},
		 "properties":{"id":17,"area_id":301,"level":0,"outdoor":false,
		  "transitions":[{"id":"n18","weight":1.5,"distance":12.25},{"id":19,"weight":1,"distance":3}]}}]}`)
	g := must[*myitmo.NavigatorTransitions](t)(c.Navigator.Transitions(ctx, 2))
	f.expect(http.MethodGet, "/api/navigator/buildings/2/transition")
	if g.Type != "FeatureCollection" || len(g.Features) != 1 {
		t.Fatalf("graph = %+v", g)
	}
	n := g.Features[0].Properties
	if n.ID != "17" || n.AreaID != 301 || n.Transitions[0].ID != "n18" || n.Transitions[1].ID != "19" ||
		n.Transitions[0].Weight*n.Transitions[0].Distance != 18.375 {
		t.Errorf("node = %+v", n)
	}
	if len(g.Features[0].Geometry) == 0 {
		t.Error("geometry is empty")
	}
}
