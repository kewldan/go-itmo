package myitmo

import (
	"context"
)

// NavigatorService is campus buildings map (/api/navigator). Its responses
// have no envelope. Vector tiles are served by
// https://mbtiles.itmo.su/services/{building_id}/tiles/{z}/{x}/{y}.pbf without auth.
type NavigatorService struct{ c *Client }

// NavigatorBuilding is a building with an indoor map.
type NavigatorBuilding struct {
	// BuildingID is also the building id of [BookingRoom.BuildingID].
	BuildingID int64  `json:"building_id"`
	Name       string `json:"name"`
	// Bounds is [west, south, east, north] in degrees of longitude and latitude.
	Bounds [4]float64 `json:"bounds"`
	// Bearing is the map rotation in degrees.
	Bearing float64 `json:"bearing"`
}

// NavigatorTransitions is the routing graph of a building as a GeoJSON FeatureCollection.
type NavigatorTransitions struct {
	Type     string                       `json:"type"`
	Features []NavigatorTransitionFeature `json:"features"`
}

// NavigatorTransitionFeature is one node of the routing graph.
type NavigatorTransitionFeature struct {
	Type       string                  `json:"type"`
	Properties NavigatorTransitionNode `json:"properties"`
	// Geometry is a standard GeoJSON geometry.
	Geometry RawJSON `json:"geometry"`
}

// NavigatorTransitionNode is a graph node and its outgoing edges.
type NavigatorTransitionNode struct {
	ID FlexID `json:"id"`
	// AreaID is the id of the map area (room feature) the node belongs to.
	AreaID int64 `json:"area_id"`
	// Level is the zero-based floor index.
	Level   int  `json:"level"`
	Outdoor bool `json:"outdoor"`
	// Transitions are the edges; the cost of an edge is Weight*Distance.
	Transitions []NavigatorTransitionEdge `json:"transitions"`
}

// NavigatorTransitionEdge is an edge of the routing graph.
type NavigatorTransitionEdge struct {
	// ID is the id of the neighbour node.
	ID       FlexID  `json:"id"`
	Weight   float64 `json:"weight"`
	Distance float64 `json:"distance"`
}

// Buildings returns the buildings with an indoor map.
// GET /api/navigator/buildings
func (s *NavigatorService) Buildings(ctx context.Context) ([]NavigatorBuilding, error) {
	return callRaw[[]NavigatorBuilding](ctx, s.c, get("api/navigator/buildings", nil))
}

// Style returns the MapLibre GL style document of the indoor map. Its
// sources are meant to be replaced with the tile URLs of a building.
// GET /api/navigator/style
func (s *NavigatorService) Style(ctx context.Context) (RawJSON, error) {
	return callRaw[RawJSON](ctx, s.c, get("api/navigator/style", nil))
}

// Transitions returns the routing graph of a building.
// GET /api/navigator/buildings/{building_id}/transition
func (s *NavigatorService) Transitions(ctx context.Context, buildingID int64) (*NavigatorTransitions, error) {
	return callRaw[*NavigatorTransitions](ctx, s.c, get("api/navigator/buildings/"+id(buildingID)+"/transition", nil))
}
