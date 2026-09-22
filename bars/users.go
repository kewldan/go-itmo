package bars

import "context"

// UserSummary is a user in search results and role bindings, usually shown
// as "{id} / {last} {first} {middle}".
type UserSummary struct {
	// ID is the BARS user ID, not ISU.
	ID         int64  `json:"id"`
	Login      string `json:"login,omitzero"`
	FirstName  string `json:"first_name"`
	MiddleName string `json:"middle_name"`
	LastName   string `json:"last_name"`
}

// WhiteListEntry restricts a role to a group, a flow or a discipline.
type WhiteListEntry struct {
	// ID is a server ID for saved entries; new entries may carry
	// client-side string IDs.
	ID RawJSON `json:"id,omitzero"`
	// Type is FlowTypeGroup or FlowTypeFlow; discipline entries use another value or none.
	Type       string `json:"type,omitzero"`
	Name       string `json:"name"`
	Identifier string `json:"identifier"`
	Term       *int   `json:"term,omitzero"`
	Year       string `json:"year,omitzero"`
	// DisciplineIdentifier and DisciplineName are set for discipline entries.
	DisciplineIdentifier *int64 `json:"discipline_identifier,omitzero"`
	DisciplineName       string `json:"discipline_name,omitzero"`
	// Departments is copied from Discipline.Departments; shape unknown.
	Departments RawJSON `json:"departments,omitzero"`
}

// Users searches all users by name or number; "" lists without a filter.
//
// GET /users
//
// Audience: Admin (DOD).
func (c *Client) Users(ctx context.Context, filter string) ([]UserSummary, error) {
	return fetch[[]UserSummary](ctx, c, get("users", q().str("filter", filter).values()))
}

// Students searches students. The response is assumed to be a list of
// users.
//
// GET /users/students
//
// Audience: Teacher, Admin.
func (c *Client) Students(ctx context.Context, filter string) ([]UserSummary, error) {
	return fetch[[]UserSummary](ctx, c, get("users/students", q().str("filter", filter).values()))
}

// UserRoles returns every role with its assignment state for a user: Selected
// marks assigned roles, WhiteList and MainUser hold the restrictions.
//
// GET /users/{userId}/roles
//
// Audience: Admin (DOD).
func (c *Client) UserRoles(ctx context.Context, userID int64) ([]UserRole, error) {
	return fetch[[]UserRole](ctx, c, get(route("users", id(userID), "roles"), nil))
}

// SetUserRoles saves the role assignments of a user: send the full list from
// [Client.UserRoles] with Selected, WhiteList and MainUser edited. Roles with
// AllowsMultiple may appear several times, once per main user. The server
// answers 409 when the user has never signed in.
//
// POST /users/{userId}/roles
//
// Audience: Admin (DOD).
func (c *Client) SetUserRoles(ctx context.Context, userID int64, roles []UserRole) error {
	if roles == nil {
		roles = []UserRole{}
	}
	return c.exec(ctx, post(route("users", id(userID), "roles"), roles))
}

// SelectRole switches the active role of the current user; pass an element
// of User.UserRoles as is. Read [Client.CurrentUser] afterwards.
//
// POST /users/set_selected_role
//
// Audience: Student, Teacher, Admin, Parent.
func (c *Client) SelectRole(ctx context.Context, role UserRole) error {
	return c.exec(ctx, post("users/set_selected_role", role))
}

// UsersByRole searches users that have a role, to bind them as the main
// user of another role (UserRole.RequiresMainUserUserRole).
//
// GET /users/by_role/{roleId}
//
// Audience: Admin (DOD).
func (c *Client) UsersByRole(ctx context.Context, roleID int64, filter string) ([]UserSummary, error) {
	return fetch[[]UserSummary](ctx, c, get(route("users", "by_role", id(roleID)), q().str("filter", filter).values()))
}

// WhiteListCandidates searches groups (FlowTypeGroup) or flows
// (FlowTypeFlow) for role white lists and report filters. realizerLogin
// restricts the search to a teacher (the login of the bound main user); ""
// means no restriction; no query is sent then.
//
// GET /users/white_list/{type}
//
// Audience: Teacher, Admin.
func (c *Client) WhiteListCandidates(ctx context.Context, flowType, filter, realizerLogin string) ([]WhiteListEntry, error) {
	qv := q().str("filter", filter).str("realizerPeopleId", realizerLogin).values()
	return fetch[[]WhiteListEntry](ctx, c, get(route("users", "white_list", flowType), qv))
}
