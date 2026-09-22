package bars

import (
	"context"
	"time"
)

// UserDeadline is an entry of the user's deadline calendar.
type UserDeadline struct {
	DisciplineName string `json:"discipline_name"`
	// PlanIdentifier is the group or flow identifier (GroupOrFlow.Identifier).
	PlanIdentifier string      `json:"plan_identifier"`
	Checkpoint     *Checkpoint `json:"checkpoint"`
	DeadlineDate   time.Time   `json:"deadline_date,omitzero"`
}

// Deadline is the deadline of one checkpoint of a plan for a group or flow.
// It is read and written back, so members this client does not know are kept
// in Unknown.
type Deadline struct {
	CheckpointID int64 `json:"checkpoint_id"`
	// ParentCheckpointID is set for sub-checkpoints.
	ParentCheckpointID *int64 `json:"parent_checkpoint_id,omitzero"`
	// DeadlineDate is null when no deadline is set.
	DeadlineDate Millis `json:"deadline_date"`
	// CreatedBy and UpdatedBy have an unknown shape.
	CreatedBy RawJSON `json:"created_by,omitzero"`
	UpdatedBy RawJSON `json:"updated_by,omitzero"`
	CreatedAt Millis  `json:"created_at,omitzero"`
	UpdatedAt Millis  `json:"updated_at,omitzero"`
	// Unknown holds the other members.
	Unknown RawJSON `json:",embed"`
}

type singleDeadline struct {
	Checkpoint         Checkpoint `json:"checkpoint"`
	CheckpointID       int64      `json:"checkpoint_id"`
	ParentCheckpointID *int64     `json:"parent_checkpoint_id"`
	DeadlineDate       Millis     `json:"deadline_date"`
}

// Deadlines returns all deadlines of the current user (the calendar page).
// The server answers 423 when the user is not a student in the selected period.
//
// GET /deadline
//
// Audience: Student, Parent, Teacher.
func (c *Client) Deadlines(ctx context.Context) ([]UserDeadline, error) {
	return fetch[[]UserDeadline](ctx, c, get("deadline", nil))
}

// PlanDeadlines returns the deadlines of a plan for a group or flow.
//
// GET /deadline/{type}/{identifier}/{checkpointPlanId}
//
// Audience: Teacher; Student (read only).
func (c *Client) PlanDeadlines(ctx context.Context, flowType, identifier string, planID int64) ([]Deadline, error) {
	return fetch[[]Deadline](ctx, c, get(route("deadline", flowType, identifier, id(planID)), nil))
}

// SaveDeadlines replaces the deadlines of a group or flow. It drops the
// created/updated audit fields before sending. Every checkpoint should have a
// deadline, including an entry for every group checkpoint whose date is the
// latest date of its sub-checkpoints; include such entries in deadlines
// yourself.
//
// POST /deadline/{type}/{identifier}
//
// Audience: Teacher, Admin.
func (c *Client) SaveDeadlines(ctx context.Context, flowType, identifier string, deadlines []Deadline) error {
	body := make([]Deadline, len(deadlines))
	for i, d := range deadlines {
		d.CreatedBy, d.UpdatedBy, d.CreatedAt, d.UpdatedAt = nil, nil, Millis{}, Millis{}
		body[i] = d
	}
	return c.exec(ctx, post(route("deadline", flowType, identifier), body))
}

// SetDeadline sets the deadline of one checkpoint for a group or flow; the
// zero date clears it. checkpoint is the full checkpoint from the plan.
//
// POST /deadline/single/{type}/{identifier}
//
// Audience: Teacher.
func (c *Client) SetDeadline(ctx context.Context, flowType, identifier string, checkpoint Checkpoint, date time.Time) error {
	body := singleDeadline{
		Checkpoint:         checkpoint,
		CheckpointID:       checkpoint.ID,
		ParentCheckpointID: checkpoint.ParentCheckpointID,
		DeadlineDate:       MillisOf(date),
	}
	return c.exec(ctx, post(route("deadline", "single", flowType, identifier), body))
}
