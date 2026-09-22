package myitmo

import "context"

// SystemService is cabinet layout: dashboard, menu, services (/api/system).
type SystemService struct{ c *Client }

// DashboardItem is the position and limits of one home-page widget (grid units).
type DashboardItem struct {
	// I is the layout item id.
	I string `json:"i"`
	X int    `json:"x"`
	Y int    `json:"y"`
	W int    `json:"w"`
	H int    `json:"h"`
	// Widget is the widget name, e.g. "ScholarshipWidget".
	Widget              string `json:"widget"`
	Static              bool   `json:"static"`
	MinW                *int   `json:"min_w"`
	MaxW                *int   `json:"max_w"`
	MinH                *int   `json:"min_h"`
	MaxH                *int   `json:"max_h"`
	IsDraggable         *bool  `json:"is_draggable"`
	IsResizable         *bool  `json:"is_resizable"`
	PreserveAspectRatio *bool  `json:"preserve_aspect_ratio"`
	DragAllowFrom       string `json:"drag_allow_from"`
	DragIgnoreFrom      string `json:"drag_ignore_from"`
	ResizeIgnoreFrom    string `json:"resize_ignore_from"`
}

// MenuItem is a navigation section or a service card.
type MenuItem struct {
	// Title and Description are translation keys, not display text.
	Title       string `json:"title"`
	Description string `json:"description"`
	// To is an internal route or an external link.
	To string `json:"to"`
	// Icon is a monochrome icon name; "icon-legacy" marks an old icon class.
	Icon           string `json:"icon"`
	MulticolorIcon string `json:"multicolor_icon"`
	// Exact requires an exact route match to mark the item active.
	Exact bool `json:"exact"`
	// Color is the service card colour; usually empty for menu items.
	Color    string     `json:"color"`
	Children []MenuItem `json:"children"`
}

// MenuResponse wraps a menu.
type MenuResponse struct {
	Menu []MenuItem `json:"menu"`
}

// Dashboard returns the home-page widget layout.
// GET /api/system/dashboard
func (s *SystemService) Dashboard(ctx context.Context) ([]DashboardItem, error) {
	return call[[]DashboardItem](ctx, s.c, get("api/system/dashboard", nil))
}

// Menu returns the sidebar navigation menu.
// GET /api/system/v1/menu/items
func (s *SystemService) Menu(ctx context.Context) (*MenuResponse, error) {
	return call[*MenuResponse](ctx, s.c, get("api/system/v1/menu/items", nil))
}

// Services returns the services catalogue.
// GET /api/system/v1/menu/services
func (s *SystemService) Services(ctx context.Context) (*MenuResponse, error) {
	return call[*MenuResponse](ctx, s.c, get("api/system/v1/menu/services", nil))
}
