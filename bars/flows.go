package bars

import "context"

// Flow is a lecture flow.
type Flow struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// FlowPlanLink binds a plan to a flow.
type FlowPlanLink struct {
	Flow Flow `json:"flow"`
	// CheckpointPlan is camelCase on the wire.
	CheckpointPlan FlowLinkPlan `json:"checkpointPlan"`
}

// FlowLinkPlan is the plan of a FlowPlanLink; the discipline has the
// label/value shape here.
type FlowLinkPlan struct {
	ID         int64        `json:"id"`
	Discipline LabeledValue `json:"discipline"`
	Terms      []int        `json:"terms"`
}

// FlowPlanLinks returns the flow to plan links, optionally filtered by flow
// name. The server does not page the result.
//
// GET /flows/plans
//
// Audience: Admin (DOD).
func (c *Client) FlowPlanLinks(ctx context.Context, name string) ([]FlowPlanLink, error) {
	return fetch[[]FlowPlanLink](ctx, c, get("flows/plans", q().str("name", name).values()))
}

// Flows searches flows by name; all is sent only when true.
//
// GET /flows
//
// Audience: Admin (DOD).
func (c *Client) Flows(ctx context.Context, name string, all bool) ([]Flow, error) {
	return fetch[[]Flow](ctx, c, get("flows", q().str("name", name).flag("all", all).values()))
}

// LinkFlowPlan makes the journals of a flow use a plan.
//
// POST /flows/{flowId}/plans/{planId}
//
// Audience: Admin (DOD).
func (c *Client) LinkFlowPlan(ctx context.Context, flowID, planID int64) error {
	return c.exec(ctx, post(route("flows", id(flowID), "plans", id(planID)), nil))
}

// UnlinkFlowPlan removes a flow to plan link.
//
// DELETE /flows/{flowId}/plans/{planId}
//
// Audience: Admin (DOD).
func (c *Client) UnlinkFlowPlan(ctx context.Context, flowID, planID int64) error {
	return c.exec(ctx, del(route("flows", id(flowID), "plans", id(planID)), nil))
}
