package bars

import (
	"context"
	"errors"
)

// StudentSession is the intermediate certification schedule of one student.
type StudentSession struct {
	Student  SessionStudent   `json:"student"`
	Schedule []SessionAttempt `json:"schedule"`
}

// SessionStudent is the student of a StudentSession. It is sent back as
// received, so members this client does not know are kept in Unknown.
// (camelCase on the wire).
type SessionStudent struct {
	LastName   string `json:"lastName"`
	FirstName  string `json:"firstName"`
	MiddleName string `json:"middleName"`
	GroupName  string `json:"groupName"`
	// Unknown holds the other members, such as IDs.
	Unknown RawJSON `json:",embed"`
}

// SessionAttempt is the date window of one attempt (AttemptFirst or
// AttemptRetry).
type SessionAttempt struct {
	Attempt int    `json:"attempt"`
	From    Millis `json:"from"`
	To      Millis `json:"to"`
}

type studentSessionUpdate struct {
	Schedule []SessionAttempt `json:"schedule"`
	Student  SessionStudent   `json:"student"`
	Course   *bool            `json:"course"`
}

type groupSessionUpdate struct {
	Course   *bool            `json:"course"`
	Schedule []SessionAttempt `json:"schedule"`
}

// planSegment joins plan IDs with commas into one path segment.
func planSegment(planIDs []int64) (string, error) {
	if len(planIDs) == 0 {
		return "", errors.New("bars: at least one checkpoint plan ID is required")
	}
	return ids(planIDs), nil
}

func attempts(s []SessionAttempt) []SessionAttempt {
	if s == nil {
		return []SessionAttempt{}
	}
	return s
}

func courseFlag(course bool) *bool {
	if course {
		return &course
	}
	return nil
}

// StudentSessions returns the per-student certification date windows of a
// plan for a group or flow; course selects the course project schedule.
// Several plan IDs are comma-joined into the path segment.
//
// GET /session/{checkpointPlanId}/{type}/{identifier}
//
// Audience: Admin (DOD).
func (c *Client) StudentSessions(ctx context.Context, planIDs []int64, flowType, identifier string, course bool) ([]StudentSession, error) {
	plans, err := planSegment(planIDs)
	if err != nil {
		return nil, err
	}
	// The plan segment holds digits and commas only and is not escaped.
	path := plans + "/" + route(flowType, identifier)
	return fetch[[]StudentSession](ctx, c, get("session/"+path, q().flag("course", course).values()))
}

// SetStudentSession saves custom certification dates for one student;
// student is StudentSession.Student as received.
//
// POST /session/{checkpointPlanId}
//
// Audience: Admin (DOD).
func (c *Client) SetStudentSession(ctx context.Context, planIDs []int64, student SessionStudent, schedule []SessionAttempt, course bool) error {
	plans, err := planSegment(planIDs)
	if err != nil {
		return err
	}
	return c.exec(ctx, post("session/"+plans, studentSessionUpdate{Schedule: attempts(schedule), Student: student, Course: courseFlag(course)}))
}

// SetGroupSession sets the same certification dates for every student of a
// group or flow. Send only attempts with both dates set.
//
// POST /session/{checkpointPlanId}/{type}/{identifier}
//
// Audience: Admin (DOD).
func (c *Client) SetGroupSession(ctx context.Context, planIDs []int64, flowType, identifier string, schedule []SessionAttempt, course bool) error {
	plans, err := planSegment(planIDs)
	if err != nil {
		return err
	}
	return c.exec(ctx, post("session/"+plans+"/"+route(flowType, identifier), groupSessionUpdate{Course: courseFlag(course), Schedule: attempts(schedule)}))
}
