package bars

import "context"

// DisciplineQuery filters the discipline catalogue of the selected period.
// Empty and false values are not sent.
type DisciplineQuery struct {
	// Name is a substring search.
	Name string
	// Type and Identifier restrict the result to a group or flow.
	Type       string
	Identifier string
	// WithCheckpointPlansOnly keeps disciplines that have a plan.
	WithCheckpointPlansOnly bool
	// DistinctByCheckpointPlans returns one entry per plan.
	DistinctByCheckpointPlans bool
	// Size limits the result (not sent when 0); its effect is not confirmed.
	Size int
}

func (p DisciplineQuery) values() query {
	return q().flag("withCheckpointPlansOnly", p.WithCheckpointPlansOnly).
		flag("distinctByCheckpointPlans", p.DistinctByCheckpointPlans).
		str("name", p.Name).
		str("identifier", p.Identifier).
		str("type", p.Type).
		num("size", int64(p.Size))
}

// DisciplineFilter filters admin discipline lists. Empty values are not sent.
type DisciplineFilter struct {
	Name       string
	Type       string
	Identifier string
	// ExcludedIDs are discipline IDs to leave out.
	ExcludedIDs []int64
}

// SearchDisciplines returns the catalogue of the selected period with all
// filters; the server scopes it to the selected role.
// [Client.Disciplines] is the short form.
//
// GET /journal/disciplines
//
// Audience: Student, Teacher, Admin, Parent.
func (c *Client) SearchDisciplines(ctx context.Context, p DisciplineQuery) ([]Discipline, error) {
	return fetch[[]Discipline](ctx, c, get("journal/disciplines", p.values().values()))
}

// LegacyDisciplines is the old discipline list. The response shape is
// unknown.
//
// GET /disciplines
//
// Audience: Teacher, Admin.
func (c *Client) LegacyDisciplines(ctx context.Context, p DisciplineQuery) (RawJSON, error) {
	return fetch[RawJSON](ctx, c, get("disciplines", p.values().values()))
}

// RealizerDisciplines returns the disciplines a teacher implements;
// realizerLogin is the teacher's login (UserSummary.Login).
//
// GET /disciplines/realizer
//
// Audience: Admin (DOD, superuser).
func (c *Client) RealizerDisciplines(ctx context.Context, realizerLogin string, f DisciplineFilter) ([]Discipline, error) {
	qv := q().str("name", f.Name).str("type", f.Type).str("identifier", f.Identifier)
	if len(f.ExcludedIDs) > 0 {
		qv.list("excludedIds", f.ExcludedIDs)
	}
	qv.str("realizerPeopleId", realizerLogin)
	return fetch[[]Discipline](ctx, c, get("disciplines/realizer", qv.values()))
}

// AllDisciplines returns all disciplines regardless of the user; Name is
// sent as the filter parameter. The element shape is assumed to be Discipline.
//
// GET /disciplines/all
//
// Audience: Admin (DOD, superuser).
func (c *Client) AllDisciplines(ctx context.Context, f DisciplineFilter) ([]Discipline, error) {
	qv := q().str("filter", f.Name).str("type", f.Type).str("identifier", f.Identifier)
	if len(f.ExcludedIDs) > 0 {
		qv.list("excludedIds", f.ExcludedIDs)
	}
	return fetch[[]Discipline](ctx, c, get("disciplines/all", qv.values()))
}

// DisciplineByID returns one discipline. The response shape is unknown.
//
// GET /disciplines/{disciplineId}
//
// Audience: unknown.
func (c *Client) DisciplineByID(ctx context.Context, disciplineID int64) (RawJSON, error) {
	return fetch[RawJSON](ctx, c, get(route("disciplines", id(disciplineID)), nil))
}

// PlanDisciplines returns the disciplines a plan can be created for, each
// with its single semester in Term; flowID 0 means all, otherwise the
// disciplines of that flow.
//
// GET /disciplines_for_plan/
//
// Audience: Teacher (teacher-admin), Admin (DOD).
func (c *Client) PlanDisciplines(ctx context.Context, flowID int64) ([]PlanDiscipline, error) {
	return fetch[[]PlanDiscipline](ctx, c, get("disciplines_for_plan/", q().num("flowId", flowID).values()))
}

// DisciplineAccess returns the access settings of a discipline. The response
// shape is unknown.
//
// GET /discipline/{id}/access
//
// Audience: Admin.
func (c *Client) DisciplineAccess(ctx context.Context, disciplineID int64) (RawJSON, error) {
	return fetch[RawJSON](ctx, c, get(route("discipline", id(disciplineID), "access"), nil))
}

// SetDisciplineAccess saves the access settings of a discipline. Body and
// response are unknown.
//
// POST /discipline/{id}/access
//
// Audience: Admin.
func (c *Client) SetDisciplineAccess(ctx context.Context, disciplineID int64, body RawJSON) (RawJSON, error) {
	return fetch[RawJSON](ctx, c, post(route("discipline", id(disciplineID), "access"), raw(body)))
}

// GroupsAndFlowsQuery filters [Client.SearchGroupsAndFlows]; zero values are not sent.
type GroupsAndFlowsQuery struct {
	DisciplineID     int64
	CheckpointPlanID int64
	Name             string
}

// SearchGroupsAndFlows returns the groups and flows of the selected period
// with all filters. [Client.GroupsAndFlows] is the short form.
//
// GET /journal/groups-and-flows
//
// Audience: Student, Teacher, Admin, Parent.
func (c *Client) SearchGroupsAndFlows(ctx context.Context, p GroupsAndFlowsQuery) ([]GroupOrFlow, error) {
	qv := q().num("disciplineId", p.DisciplineID).num("checkpointPlanId", p.CheckpointPlanID).str("name", p.Name)
	return fetch[[]GroupOrFlow](ctx, c, get("journal/groups-and-flows", qv.values()))
}

// Groups lists groups. The response shape is unknown.
//
// GET /groups
//
// Audience: unknown.
func (c *Client) Groups(ctx context.Context) (RawJSON, error) {
	return fetch[RawJSON](ctx, c, get("groups", nil))
}

// DisciplineGroups lists the groups of a discipline. The response shape is
// unknown.
//
// GET /disciplines/{disciplineId}/group
//
// Audience: unknown.
func (c *Client) DisciplineGroups(ctx context.Context, disciplineID int64) (RawJSON, error) {
	return fetch[RawJSON](ctx, c, get(route("disciplines", id(disciplineID), "group"), nil))
}

// PlanGroupAndFlowNames returns the names of the groups and flows bound to a
// plan. The response shape is unknown.
//
// GET /checkpoint_plans/{id}/group_and_flow_names
//
// Audience: unknown.
func (c *Client) PlanGroupAndFlowNames(ctx context.Context, planID int64) (RawJSON, error) {
	return fetch[RawJSON](ctx, c, get(route("checkpoint_plans", id(planID), "group_and_flow_names"), nil))
}
