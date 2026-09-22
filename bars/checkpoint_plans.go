package bars

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"
)

// Plan statuses (CheckpointPlanWrite.Status).
const (
	PlanStatusSaved = "SAVED" // Сохранено
	PlanStatusSent  = "SENT"  // Отправлено
)

// CourseProjectRange is the grade range of a course project, usually 0..100.
type CourseProjectRange struct {
	MinGrade float64 `json:"min_grade"`
	MaxGrade float64 `json:"max_grade"`
}

// CheckpointPlanWrite is the body to create or update a plan. New checkpoints carry a client GID ("chk_1", "sub_1",
// "f_1") and no ID; saved ones keep their ID.
type CheckpointPlanWrite struct {
	// ID is set for updates only.
	ID  int64  `json:"id,omitzero"`
	GID string `json:"gid,omitzero"`
	// Discipline is an entry of [Client.PlanDisciplines] as is.
	Discipline         PlanDiscipline       `json:"discipline"`
	Terms              []int                `json:"terms"`
	Programs           []EducationalProgram `json:"programs"`
	RegularCheckpoints []Checkpoint         `json:"regular_checkpoints"`
	// FinalCheckpoint needs GID, Key, TypeID, Type, MinGrade and MaxGrade.
	FinalCheckpoint *Checkpoint `json:"final_checkpoint"`
	// PointDistribution is the maximum for regular checkpoints (80 by
	// default); the final checkpoint gets the rest of 100.
	PointDistribution int  `json:"point_distribution"`
	AdditionalPoints  bool `json:"additional_points"`
	AlternateMethods  bool `json:"alternate_methods"`
	// HasCourseProject mirrors Discipline.CourseProject.
	HasCourseProject bool `json:"has_course_project"`
	// CourseProjectCheckpoint is nil (sent as null) without a course project.
	CourseProjectCheckpoint *CourseProjectRange `json:"course_project_checkpoint"`
	// Status is PlanStatusSaved or PlanStatusSent.
	Status string `json:"status"`
	// UpdatedAt is the client time of the change.
	UpdatedAt Millis `json:"updated_at,omitzero"`
}

// CourseProjectPlanWrite is the short body of a course-project-only plan.
// Note the key spelling course_project_checkPoint.
type CourseProjectPlanWrite struct {
	// ID is set for updates only.
	ID            int64 `json:"id,omitzero"`
	CourseProject bool  `json:"course_project"`
	// CourseProjectCheckPoint must have MinGrade < MaxGrade.
	CourseProjectCheckPoint CourseProjectRange `json:"course_project_checkPoint"`
	// Status is usually PlanStatusSent.
	Status     string         `json:"status"`
	Discipline PlanDiscipline `json:"discipline"`
}

// CheckpointPlanSummary is an entry of the paged plan list; the discipline
// has the label/value shape here.
type CheckpointPlanSummary struct {
	ID         int64                `json:"id"`
	Discipline LabeledValue         `json:"discipline"`
	Programs   []EducationalProgram `json:"programs"`
	Terms      []int                `json:"terms"`
	CreatedAt  time.Time            `json:"created_at,omitzero"`
	UpdatedAt  time.Time            `json:"updated_at,omitzero"`
}

// CheckpointPlanQuery pages and filters [Client.CheckpointPlans].
type CheckpointPlanQuery struct {
	// Name searches by discipline name.
	Name string
	// Page is 0-based and always sent.
	Page int
	// Size is the page size; 0 leaves the server default.
	Size int
	// Sort is "<field>,<asc|desc>" with field discipline.name, program,
	// term, createdAt or updatedAt.
	Sort string
	// StartDate and EndDate filter by date; zero values are not sent.
	StartDate time.Time
	EndDate   time.Time
}

// CheckpointPlan returns a plan with all checkpoints.
//
// GET /checkpoint_plans/{id}
//
// Audience: Student (preview), Teacher, Admin.
func (c *Client) CheckpointPlan(ctx context.Context, planID int64) (*CheckpointPlan, error) {
	return fetch[*CheckpointPlan](ctx, c, get(route("checkpoint_plans", id(planID)), nil))
}

// DisciplineCheckpointPlans returns the plans of a discipline, optionally
// restricted to programmes. programIDs nil sends no filter; an empty non-nil
// slice sends "programIds=", which selects university-wide disciplines
// without a programme.
//
// GET /checkpoint_plans/discipline/{disciplineId}
//
// Audience: Admin (DOD).
func (c *Client) DisciplineCheckpointPlans(ctx context.Context, disciplineID int64, programIDs []int64) ([]CheckpointPlan, error) {
	return fetch[[]CheckpointPlan](ctx, c, get(route("checkpoint_plans", "discipline", id(disciplineID)), q().list("programIds", programIDs).values()))
}

// CreateCheckpointPlan creates a plan. The answer is decoded as a full plan;
// only its id is confirmed. The server answers 409
// when the discipline already has a plan. Editing needs the global setting
// SettingEditBars.
//
// POST /checkpoint_plans
//
// Audience: Teacher (teacher-admin).
func (c *Client) CreateCheckpointPlan(ctx context.Context, plan CheckpointPlanWrite) (*CheckpointPlan, error) {
	return fetch[*CheckpointPlan](ctx, c, post("checkpoint_plans", plan))
}

// UpdateCheckpointPlan replaces a plan; plan.ID selects it. The server
// answers 204.
//
// PUT /checkpoint_plans/{id}
//
// Audience: Teacher (teacher-admin).
func (c *Client) UpdateCheckpointPlan(ctx context.Context, plan CheckpointPlanWrite) error {
	if plan.ID == 0 {
		return errors.New("bars: UpdateCheckpointPlan needs plan.ID")
	}
	return c.exec(ctx, put(route("checkpoint_plans", id(plan.ID)), plan))
}

// CreateCourseProjectPlan creates a course-project-only plan.
//
// POST /checkpoint_plans
//
// Audience: Teacher (teacher-admin).
func (c *Client) CreateCourseProjectPlan(ctx context.Context, plan CourseProjectPlanWrite) (*CheckpointPlan, error) {
	return fetch[*CheckpointPlan](ctx, c, post("checkpoint_plans", plan))
}

// UpdateCourseProjectPlan saves the course project part of a plan; plan.ID
// selects it.
//
// PUT /checkpoint_plans/{id}
//
// Audience: Teacher (teacher-admin).
func (c *Client) UpdateCourseProjectPlan(ctx context.Context, plan CourseProjectPlanWrite) error {
	if plan.ID == 0 {
		return errors.New("bars: UpdateCourseProjectPlan needs plan.ID")
	}
	return c.exec(ctx, put(route("checkpoint_plans", id(plan.ID)), plan))
}

// DeleteCheckpointPlan deletes a plan.
//
// DELETE /checkpoint_plans/{id}
//
// Audience: Teacher (teacher-admin).
func (c *Client) DeleteCheckpointPlan(ctx context.Context, planID int64) error {
	return c.exec(ctx, del(route("checkpoint_plans", id(planID)), nil))
}

// CheckpointPlans searches plans page by page.
//
// GET /checkpoint_plans
//
// Audience: Teacher (teacher-admin).
func (c *Client) CheckpointPlans(ctx context.Context, p CheckpointPlanQuery) (*Page[CheckpointPlanSummary], error) {
	qv := q().str("name", p.Name)
	qv["page"] = []string{strconv.Itoa(p.Page)}
	qv.num("size", int64(p.Size)).str("sort", p.Sort)
	if !p.StartDate.IsZero() {
		qv.num("start_date", p.StartDate.UnixMilli())
	}
	if !p.EndDate.IsZero() {
		qv.num("end_date", p.EndDate.UnixMilli())
	}
	return fetch[*Page[CheckpointPlanSummary]](ctx, c, get("checkpoint_plans", qv.values()))
}

// ImportRPD creates plans from work programmes (РПД) and returns the IDs of
// the disciplines imported. Several IDs are joined with commas, which the
// parameter name suggests but is not confirmed.
//
// Warning: this GET creates plans on the server.
//
// GET /checkpoint_plans/rpd/constructor/import
//
// Audience: Teacher (teacher-admin).
func (c *Client) ImportRPD(ctx context.Context, rpdIDs ...string) ([]int64, error) {
	if len(rpdIDs) == 0 {
		return nil, errors.New("bars: ImportRPD needs at least one ID")
	}
	return fetch[[]int64](ctx, c, get("checkpoint_plans/rpd/constructor/import", q().str("ids", strings.Join(rpdIDs, ",")).values()))
}

// PersonalPlan returns a personal plan. The response shape is unknown.
//
// GET /personal_plans/{id}
//
// Audience: unknown.
func (c *Client) PersonalPlan(ctx context.Context, planID int64) (RawJSON, error) {
	return fetch[RawJSON](ctx, c, get(route("personal_plans", id(planID)), nil))
}

// GroupPersonalPlan returns the personal plan of a discipline for a group.
// The response shape is unknown.
//
// GET /personal_plans/{disciplineId}/{groupName}
//
// Audience: unknown.
func (c *Client) GroupPersonalPlan(ctx context.Context, disciplineID int64, groupName string) (RawJSON, error) {
	return fetch[RawJSON](ctx, c, get(route("personal_plans", id(disciplineID), groupName), nil))
}

// SavePersonalPlan saves a personal plan. Body and response are unknown.
//
// POST /personal_plans
//
// Audience: unknown.
func (c *Client) SavePersonalPlan(ctx context.Context, body RawJSON) (RawJSON, error) {
	return fetch[RawJSON](ctx, c, post("personal_plans", raw(body)))
}
