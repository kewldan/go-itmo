package bars

import (
	"context"
	"strconv"
	"strings"
)

// EducationalProgram is a programme (направление) a plan applies to.
// Lists of plans may carry only Code and Name.
type EducationalProgram struct {
	ID   int64  `json:"id,omitzero"`
	Code string `json:"code"`
	Name string `json:"name"`
}

// Test is a CDO (ЦДО) test a checkpoint can be bound to through
// Checkpoint.TestID.
type Test struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	// Year is "2025/2026".
	Year string `json:"year"`
}

// CheckpointType is a kind of assessment (оценочное средство) matched by
// Checkpoint.TypeID.
type CheckpointType struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// CheckpointKind groups checkpoint types.
type CheckpointKind string

// Checkpoint kinds.
const (
	CheckpointKindRegular CheckpointKind = "regular" // regular checkpoints
	CheckpointKindFinal   CheckpointKind = "final"   // intermediate certification
	CheckpointKindCourse  CheckpointKind = "course"  // course project
)

// EducationalPrograms returns the programmes of a discipline for the given
// semesters. Without semesters the literal term=null is sent.
//
// GET /educational_programs/{disciplineId}
//
// Audience: Teacher (teacher-admin), Admin (DOD).
func (c *Client) EducationalPrograms(ctx context.Context, disciplineID int64, terms ...int) ([]EducationalProgram, error) {
	term := "null"
	if len(terms) > 0 {
		parts := make([]string, len(terms))
		for i, t := range terms {
			parts[i] = strconv.Itoa(t)
		}
		term = strings.Join(parts, ",")
	}
	return fetch[[]EducationalProgram](ctx, c, get(route("educational_programs", id(disciplineID)), q().str("term", term).values()))
}

// Tests returns the CDO tests of a year ("2025/2026"); "" means all years.
//
// GET /tests
//
// Audience: Teacher (teacher-admin), Admin (DOD).
func (c *Client) Tests(ctx context.Context, year string) ([]Test, error) {
	return fetch[[]Test](ctx, c, get("tests", q().str("year", year).values()))
}

type newTest struct {
	Name string `json:"name"`
	Year string `json:"year"`
}

// CreateTest registers a CDO test for a year ("2025/2026").
//
// POST /tests
//
// Audience: Admin (DOD).
func (c *Client) CreateTest(ctx context.Context, name, year string) error {
	return c.exec(ctx, post("tests", newTest{Name: name, Year: year}))
}

// CheckpointTypes returns the checkpoint types, optionally filtered by name
// and kind ("" for either means no filter).
//
// GET /checkpoint_types
//
// Audience: Teacher, Admin.
func (c *Client) CheckpointTypes(ctx context.Context, name string, kind CheckpointKind) ([]CheckpointType, error) {
	return fetch[[]CheckpointType](ctx, c, get("checkpoint_types", q().str("name", name).str("type", string(kind)).values()))
}

type namedEntity struct {
	Name string `json:"name"`
}

// CreateCheckpointType adds a checkpoint type of a kind. The server answers
// 201 on success and 409 for a duplicate name.
//
// POST /checkpoint_types/{type}
//
// Audience: Admin (DOD).
func (c *Client) CreateCheckpointType(ctx context.Context, kind CheckpointKind, name string) error {
	return c.exec(ctx, post(route("checkpoint_types", string(kind)), namedEntity{Name: name}))
}
