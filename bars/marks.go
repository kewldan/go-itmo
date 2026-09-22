package bars

import (
	"context"
	"errors"
	"strconv"
	"time"
)

// TeacherJournal is the full journal of a plan for a group or flow: the same
// shape as StudentJournal with every student of the group or flow.
type TeacherJournal = StudentJournal

// Mark types: the {markType} path segment and MarkUpdate.Type.
const (
	MarkTypeCurrent    = "current"
	MarkTypeFinal      = "final"
	MarkTypeAdditional = "additional"
	MarkTypeCourse     = "course"
)

// MarkUpdate is the body that sets or clears a mark.:
//
//   - current: Mark (nil clears) and CheckpointID of the checkpoint;
//   - final: Mark (0 when absent), CheckpointID of the final checkpoint and IsAbsent;
//   - additional: Mark (nil clears); CheckpointID and CheckpointPlanID are both the plan ID;
//   - course: Mark, CheckpointID of CheckpointPlan.CourseProjectCheckpoint, optionally IsAbsent.
type MarkUpdate struct {
	Mark         *float64 `json:"mark"`
	CheckpointID int64    `json:"checkpoint_id"`
	// Type is one of the MarkType* constants.
	Type             string `json:"type"`
	IsAbsent         *bool  `json:"is_absent,omitzero"`
	CheckpointPlanID int64  `json:"checkpoint_plan_id,omitzero"`
}

// MarkHistoryEntry is one change of a mark or of an approval. Mark is set
// for mark history, MarksSum and MarkString for approval history.
type MarkHistoryEntry struct {
	ID           int64     `json:"id"`
	SetAt        time.Time `json:"set_at,omitzero"`
	SetByName    string    `json:"set_by_name"`
	Mark         *float64  `json:"mark"`
	MarksSum     *float64  `json:"marks_sum"`
	MarkString   string    `json:"mark_string"`
	CheckpointID *int64    `json:"checkpoint_id"`
}

// JournalPlan is a plan available for the Google Sheets export.
type JournalPlan struct {
	ID       int64                `json:"id"`
	Name     string               `json:"name"`
	Terms    []int                `json:"terms"`
	Programs []EducationalProgram `json:"programs"`
}

// TeacherJournal returns the full journal of a plan for a group or flow. The
// server answers 423 when the user has no such journal in the selected period.
//
// GET /marks/{checkpointPlanId}/{type}/{identifier}
//
// Audience: Teacher.
func (c *Client) TeacherJournal(ctx context.Context, planID int64, flowType, identifier string) (*TeacherJournal, error) {
	return fetch[*TeacherJournal](ctx, c, get(route("marks", id(planID), flowType, identifier), nil))
}

// TeacherJournalAuto returns the journal of a discipline for a group or flow
// when the plan is not known; the server picks the plan.
//
// GET /marks/{disciplineId}/{type}/{identifier}/auto
//
// Audience: Teacher.
func (c *Client) TeacherJournalAuto(ctx context.Context, disciplineID int64, flowType, identifier string) (*TeacherJournal, error) {
	return fetch[*TeacherJournal](ctx, c, get(route("marks", id(disciplineID), flowType, identifier, "auto"), nil))
}

// JournalStudent returns one student's row of a journal, for example to
// reload it after a mark change. studentID is StudentRecord.StudentID.
//
// GET /marks/{checkpointPlanId}/{type}/{identifier}/student/{studentId}
//
// Audience: Teacher.
func (c *Client) JournalStudent(ctx context.Context, planID int64, flowType, identifier string, studentID int64) (*StudentRecord, error) {
	return fetch[*StudentRecord](ctx, c, get(route("marks", id(planID), flowType, identifier, "student", id(studentID)), nil))
}

// MarkHistory returns the changes of a student's mark for a checkpoint;
// checkpointID 0 selects the additional points.
//
// GET /marks/{checkpointPlanId}/{type}/{identifier}/student/{studentId}/history
//
// Audience: Teacher.
func (c *Client) MarkHistory(ctx context.Context, planID int64, flowType, identifier string, studentID, checkpointID int64) ([]MarkHistoryEntry, error) {
	path := route("marks", id(planID), flowType, identifier, "student", id(studentID), "history")
	return fetch[[]MarkHistoryEntry](ctx, c, get(path, q().num("checkpointId", checkpointID).values()))
}

// SetMark sets, clears or marks absent one mark of a student in a group or
// flow; the path takes m.Type as {markType}. The plan is not part of the path.
//
// POST /marks/{type}/{identifier}/{markType}/{studentId}
//
// Audience: Teacher.
func (c *Client) SetMark(ctx context.Context, flowType, identifier string, studentID int64, m MarkUpdate) error {
	if m.Type == "" {
		return errors.New("bars: MarkUpdate.Type is required")
	}
	return c.exec(ctx, post(route("marks", flowType, identifier, m.Type, id(studentID)), m))
}

// SetAdditionalMarkLegacy sets additional points by personal plan
// checkpoint. Prefer [Client.SetMark]; the body is assumed to be a
// MarkUpdate.
//
// POST /marks/additional/{persPlanCheckpointId}/{studentId}
//
// Audience: Teacher.
func (c *Client) SetAdditionalMarkLegacy(ctx context.Context, personalPlanCheckpointID, studentID int64, m MarkUpdate) error {
	return c.exec(ctx, post(route("marks", "additional", id(personalPlanCheckpointID), id(studentID)), m))
}

// SetFinalMarkLegacy sets the final mark by personal plan checkpoint. The
// body is assumed to be a MarkUpdate; prefer [Client.SetMark].
//
// POST /marks/final/{persPlanCheckpointId}/{studentId}
//
// Audience: Teacher.
func (c *Client) SetFinalMarkLegacy(ctx context.Context, personalPlanCheckpointID, studentID int64, m MarkUpdate) error {
	return c.exec(ctx, post(route("marks", "final", id(personalPlanCheckpointID), id(studentID)), m))
}

// DisciplineMarks returns the marks of a discipline. The route has a
// trailing slash; the response shape is unknown.
//
// GET /marks/{disciplineId}/
//
// Audience: unknown.
func (c *Client) DisciplineMarks(ctx context.Context, disciplineID int64) (RawJSON, error) {
	return fetch[RawJSON](ctx, c, get(route("marks", id(disciplineID))+"/", nil))
}

// ApproveStudent approves (locks) a student's result for an attempt
// (AttemptFirst, AttemptRetry, AttemptSecondRetry); course approves the
// course project instead of the discipline and is sent as "?course=true".
//
// POST /marks/{checkpointPlanId}/{type}/{identifier}/{studentId}/approval/{attempt}
//
// Audience: Teacher.
func (c *Client) ApproveStudent(ctx context.Context, planID int64, flowType, identifier string, studentID int64, attempt int, course bool) error {
	path := route("marks", id(planID), flowType, identifier, id(studentID), "approval", strconv.Itoa(attempt))
	return c.exec(ctx, withQuery(post(path, nil), q().flag("course", course).values()))
}

// ApproveStudents approves the results of several students for an attempt
// and returns the created approvals (shape assumed). studentIDs nil omits
// the studentIds parameter; a non-nil empty slice sends it empty. course
// adds course=true.
//
// POST /marks/{checkpointPlanId}/{type}/{identifier}/approval/{attempt}
//
// Audience: Teacher.
func (c *Client) ApproveStudents(ctx context.Context, planID int64, flowType, identifier string, attempt int, studentIDs []int64, course bool) ([]Approval, error) {
	path := route("marks", id(planID), flowType, identifier, "approval", strconv.Itoa(attempt))
	qv := q().list("studentIds", studentIDs).flag("course", course).values()
	return fetch[[]Approval](ctx, c, withQuery(post(path, nil), qv))
}

// ApprovalHistory returns the changes of an approval (Approval.ID).
//
// GET /marks/{type}/{identifier}/approval/{approvalId}/history
//
// Audience: Teacher.
func (c *Client) ApprovalHistory(ctx context.Context, flowType, identifier string, approvalID int64) ([]MarkHistoryEntry, error) {
	return fetch[[]MarkHistoryEntry](ctx, c, get(route("marks", flowType, identifier, "approval", id(approvalID), "history"), nil))
}

// AcceptJournalAgreement records the student's consent to the journal terms
// of a plan.
//
// POST /marks/{checkpointPlanId}/student/agreement
//
// Audience: Student.
func (c *Client) AcceptJournalAgreement(ctx context.Context, planID int64) error {
	return c.exec(ctx, post(route("marks", id(planID), "student", "agreement"), nil))
}

// JournalPlans returns the plans available for the Google Sheets export,
// optionally filtered by name. The filter is sent as name={name}, a
// parameter name that is not confirmed.
//
// GET /journal/checkpoint-plans
//
// Audience: Teacher.
func (c *Client) JournalPlans(ctx context.Context, name string) ([]JournalPlan, error) {
	return fetch[[]JournalPlan](ctx, c, get("journal/checkpoint-plans", q().str("name", name).values()))
}

// ClearJournalCache drops the server journal cache of a period
// (year "2025/2026").
//
// DELETE /journal/cache
//
// Audience: Admin (DOD).
func (c *Client) ClearJournalCache(ctx context.Context, year string, term Term) error {
	qv := q().str("year", year)
	qv["term"] = []string{strconv.Itoa(int(term))}
	return c.exec(ctx, del("journal/cache", qv.values()))
}
