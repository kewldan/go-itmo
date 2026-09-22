package bars

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/kewldan/go-itmo/internal/rest"
)

// Term is the season stored in User.SelectedTerm. It is not a semester number.
type Term int

// Terms.
const (
	Spring Term = 0
	Autumn Term = 1
)

func (t Term) String() string {
	switch t {
	case Spring:
		return "spring"
	case Autumn:
		return "autumn"
	}
	return "term(" + strconv.Itoa(int(t)) + ")"
}

// User is the current BARS user with the selected academic period.
type User struct {
	// ID is the BARS user ID, not ISU.
	ID int64 `json:"id"`
	// Login is the ITMO.ID login; for students it is the ISU number.
	Login        string     `json:"login"`
	FirstName    string     `json:"first_name"`
	MiddleName   string     `json:"middle_name"`
	LastName     string     `json:"last_name"`
	UserRoles    []UserRole `json:"user_roles"`
	SelectedRole *UserRole  `json:"selected_role"`
	// SelectedYear is "2025/2026".
	SelectedYear   string    `json:"selected_year"`
	SelectedTerm   Term      `json:"selected_term"`
	PersonalConfig []Setting `json:"personal_config"`
	// CanChangeUser allows [Client.Impersonate].
	CanChangeUser                  bool `json:"can_change_user"`
	RestrictedToHaveReadOnlyAccess bool `json:"restricted_to_have_read_only_access"`
	// SuperUserPersonalNumber is the superuser's own number while it
	// impersonates someone; pass it to [Client.Impersonate] to switch back.
	// Not confirmed; seen only as null.
	SuperUserPersonalNumber string `json:"super_user_personal_number"`
	// CreatedBy, UpdatedBy, CreatedAt and UpdatedAt have been seen only as null.
	CreatedBy RawJSON `json:"created_by,omitzero"`
	UpdatedBy RawJSON `json:"updated_by,omitzero"`
	CreatedAt RawJSON `json:"created_at,omitzero"`
	UpdatedAt RawJSON `json:"updated_at,omitzero"`
}

// UserRole is a BARS role such as «Обучающийся». The flags are sent for
// selected_role only; GET /users/{id}/roles returns all of them. See the
// Role* constants for the IDs.
type UserRole struct {
	ID              int64  `json:"id"`
	Name            string `json:"name"`
	Locked          *bool  `json:"locked"`
	Selected        *bool  `json:"selected"`
	AllowsWhiteList *bool  `json:"allows_white_list"`
	AllowsMultiple  *bool  `json:"allows_multiple"`
	// WhiteList restricts the role to groups, flows or disciplines.
	WhiteList []WhiteListEntry `json:"white_list,omitzero"`
	// MainUser is the principal a parent or mentor role is bound to.
	MainUser *UserSummary `json:"main_user,omitzero"`
	// RequiresMainUserUserRole is the role ID the main user must have.
	RequiresMainUserUserRole *int64 `json:"requires_main_user_user_role,omitzero"`
}

// Setting is a name/value pair; values are always strings.
type Setting struct {
	ID    *int64 `json:"id,omitzero"`
	Name  string `json:"name"`
	Value string `json:"value"`
}

// Discipline is an entry of the journal catalogue for the selected period.
type Discipline struct {
	// ID is the BARS discipline ID, unrelated to MyITMO discipline_id.
	ID   int64  `json:"id"`
	Name string `json:"name"`
	// Terms are semester numbers the plan applies to, e.g. [2, 4, 6].
	Terms             []int   `json:"terms"`
	CheckpointPlanIDs []int64 `json:"checkpoint_plan_ids"`
	// Departments is copied into white lists; shape unknown.
	Departments RawJSON `json:"departments,omitzero"`
}

// GroupOrFlow addresses a journal together with a checkpoint plan.
type GroupOrFlow struct {
	// Type is "flow" or "group" (only "flow" confirmed).
	Type string `json:"type"`
	Name string `json:"name"`
	// Identifier is opaque even when numeric; pass it back as is.
	Identifier        string  `json:"identifier"`
	CheckpointPlanIDs []int64 `json:"checkpoint_plan_ids"`
	// Term and Year are optional; not confirmed.
	Term *int   `json:"term,omitzero"`
	Year string `json:"year,omitzero"`
}

// StudentJournal is the user's own journal for a plan and flow.
type StudentJournal struct {
	// Students contains the current student only; check the login anyway.
	Students []StudentRecord `json:"students"`
	Headers  JournalHeaders  `json:"headers"`
}

// JournalHeaders is the plan and the flow the journal was read for.
type JournalHeaders struct {
	Plan       CheckpointPlan `json:"plan"`
	Type       string         `json:"type"`
	Identifier string         `json:"identifier"`
	Name       string         `json:"name"`
	Deadlines  RawJSON        `json:"deadlines,omitzero"`
}

// CheckpointPlan is the list of checkpoints of a discipline (a "BARS table").
// Fields after Status are not confirmed.
type CheckpointPlan struct {
	// ID is unrelated to MyITMO est_id.
	ID                 int64          `json:"id"`
	GID                string         `json:"gid"`
	Year               string         `json:"year"`
	Terms              []int          `json:"terms"`
	Discipline         PlanDiscipline `json:"discipline"`
	RegularCheckpoints []Checkpoint   `json:"regular_checkpoints"`
	FinalCheckpoint    *Checkpoint    `json:"final_checkpoint"`
	PointDistribution  int            `json:"point_distribution"`
	AdditionalPoints   bool           `json:"additional_points"`
	HasCourseProject   bool           `json:"has_course_project"`
	// Status is either null or a string (PlanStatusSaved, PlanStatusSent);
	// not a pass/fail status.
	Status RawJSON `json:"status,omitzero"`

	// Term is used by personal plans; presence unknown.
	Term *int `json:"term,omitzero"`
	// AlternateMethods is "alternative grading"; seen only as null.
	AlternateMethods *bool `json:"alternate_methods"`
	// CourseProjectCheckpoint carries the course project grade range; seen only as null.
	CourseProjectCheckpoint *Checkpoint          `json:"course_project_checkpoint"`
	Programs                []EducationalProgram `json:"programs,omitzero"`
	// Components has been seen only empty.
	Components RawJSON   `json:"components,omitzero"`
	CreatedAt  time.Time `json:"created_at,omitzero"`
	UpdatedAt  time.Time `json:"updated_at,omitzero"`
}

// PlanDiscipline is the discipline of a plan.
type PlanDiscipline struct {
	ID            int64  `json:"id"`
	Name          string `json:"name"`
	CourseProject bool   `json:"course_project"`
	// Term is the single semester of an entry of GET /disciplines_for_plan/;
	// null inside plans.
	Term *int `json:"term,omitzero"`
}

// Checkpoint is one assessment. Scores live in Mark and link by ID.
type Checkpoint struct {
	// ID is zero for checkpoints that are not saved yet; it is omitted then.
	ID  int64  `json:"id,omitzero"`
	GID string `json:"gid"`
	// Name is empty for the final checkpoint; Type describes it then.
	Name   string `json:"name"`
	Type   string `json:"type"`
	TypeID int64  `json:"type_id"`
	// Week is nil for the final checkpoint.
	Week               *int         `json:"week"`
	Group              bool         `json:"group"`
	Key                bool         `json:"key"`
	MinGrade           float64      `json:"min_grade"`
	MaxGrade           float64      `json:"max_grade"`
	SubCheckpoints     []Checkpoint `json:"sub_checkpoints"`
	ParentCheckpointID *int64       `json:"parent_checkpoint_id"`
	// MaxSubCheckpointsFillable makes a group "best N of M".
	MaxSubCheckpointsFillable *int `json:"max_sub_checkpoints_fillable,omitzero"`
	// TestID is the CDO test ([Client.Tests]) of a checkpoint of type
	// "Электронное тестирование в ЦДО".
	TestID *int64 `json:"test_id,omitzero"`
	// TestName is the CDO test name, preferred over Name when set; seen only as null.
	TestName string `json:"test_name,omitzero"`
}

// StudentRecord is the student's row in a journal.
type StudentRecord struct {
	// StudentID is the BARS ID, not ISU.
	StudentID            int64                 `json:"student_id"`
	StudentLogin         string                `json:"student_login"`
	StudentName          string                `json:"student_name"`
	Marks                StudentMarks          `json:"marks"`
	Accessibility        *StudentAccessibility `json:"accessibility"`
	WantsToIncreaseMarks bool                  `json:"wants_to_increase_marks"`
}

// StudentMarks are the scores and approvals of a student. Sums are computed by the server.
type StudentMarks struct {
	Regular    []Mark `json:"regular"`
	Final      *Mark  `json:"final"`
	Additional *Mark  `json:"additional"`
	// Course is the course project mark (teacher journals).
	Course     *Mark    `json:"course,omitzero"`
	RegularSum *float64 `json:"regularSum"`
	// Total is 0 for an empty journal: that means "nothing yet", not a grade; see HasAnyMark.
	Total           *float64   `json:"total"`
	ActiveApprovals []Approval `json:"active_approvals"`
}

// HasAnyMark reports whether any score has been entered.
func (m *StudentMarks) HasAnyMark() bool {
	return len(m.Regular) > 0 || m.Final != nil || m.Additional != nil
}

// Mark is the score of one checkpoint. A missing mark means "not entered", not 0.
type Mark struct {
	ID int64 `json:"id"`
	// CheckpointID is nil for additional points.
	CheckpointID     *int64   `json:"checkpoint_id"`
	CheckpointPlanID int64    `json:"checkpoint_plan_id"`
	Mark             *float64 `json:"mark"`
	// Type is "current", "final", "additional" or "course" (see the
	// MarkType* constants).
	Type             string    `json:"type"`
	Absent           bool      `json:"is_absent"`
	NotBiggerThanMax bool      `json:"is_not_bigger_than_max"`
	CreatedAt        time.Time `json:"created_at,omitzero"`
	UpdatedAt        time.Time `json:"updated_at,omitzero"`
	CreatedByName    string    `json:"created_by_name"`
	UpdatedByName    string    `json:"updated_by_name"`
}

// StudentAccessibility are journal permission flags.
type StudentAccessibility struct {
	CanEditCurrentMarks           bool `json:"can_edit_current_marks"`
	CanEditAdditionalMarks        bool `json:"can_edit_additional_marks"`
	CanEditCourseMarks            bool `json:"can_edit_course_marks"`
	CanEditFinalMarks             bool `json:"can_edit_final_marks"`
	CanApproveMarks               bool `json:"can_approve_marks"`
	CanApproveRetryMarks          bool `json:"can_approve_retry_marks"`
	HasUnfilledKeyCheckpoints     bool `json:"has_unfilled_key_checkpoints"`
	CourseProjectThemeNotApproved bool `json:"course_project_theme_not_approved"`
	// The course project approval flags are sent in teacher journals only.
	CanApproveCourseProject      bool `json:"can_approve_course_project"`
	CanApproveRetryCourseProject bool `json:"can_approve_retry_course_project"`
}

// Approval is the approved result of one attempt.
type Approval struct {
	ID               int64  `json:"id"`
	StudentID        int64  `json:"student_id"`
	StudentLogin     string `json:"student_login"`
	CheckpointPlanID int64  `json:"checkpoint_plan_id"`
	// Attempt starts at 1; a retake adds a record with a higher number.
	Attempt  int      `json:"attempt"`
	MarksSum *float64 `json:"marks_sum"`
	// MarkString is the grade in words: "Отл., A", "Удвл., E", "Зачет"; see GradeCode.
	MarkString   string    `json:"mark_string"`
	Active       bool      `json:"is_active"`
	Invalid      bool      `json:"is_invalid"`
	Absent       bool      `json:"is_absent"`
	Recalculated bool      `json:"was_recalculated"`
	Course       bool      `json:"course"`
	CreatedAt    time.Time `json:"created_at,omitzero"`
	UpdatedAt    time.Time `json:"updated_at,omitzero"`
	// UpdatedByName is the teacher who approved.
	UpdatedByName string `json:"updated_by_name"`
}

var gradeLetter = regexp.MustCompile(`^[A-FX]{1,2}$`)

// GradeCode converts MarkString to the MyITMO format: "5/A", "4/C", "3/E",
// "2/FX". Pass/fail and unknown strings are returned trimmed.
func (a *Approval) GradeCode() string {
	text := strings.TrimSpace(a.MarkString)
	word, letter, ok := strings.Cut(text, ",")
	if !ok {
		return text
	}
	word = strings.TrimRight(strings.ReplaceAll(strings.ToLower(strings.TrimSpace(word)), "ё", "е"), ".")
	letter = strings.ToUpper(strings.TrimSpace(letter))
	if !gradeLetter.MatchString(letter) {
		return text
	}
	switch word {
	case "отл":
		return "5/" + letter
	case "хор":
		return "4/" + letter
	case "удвл", "удовл":
		return "3/" + letter
	case "неуд":
		return "2/" + letter
	}
	return text
}

// CurrentUser returns the user and the selected period.
//
// GET /users/current_user/
//
// Audience: Student, Teacher, Admin, Parent.
func (c *Client) CurrentUser(ctx context.Context) (*User, error) {
	return fetch[*User](ctx, c, get("users/current_user/", nil))
}

// Config returns the global settings, including current_year and current_term.
//
// GET /config/
//
// Audience: Student, Teacher, Admin, Parent.
func (c *Client) Config(ctx context.Context) ([]Setting, error) {
	return fetch[[]Setting](ctx, c, get("config/", nil))
}

// SetPersonalSetting stores a personal setting. It is shared with other BARS sessions.
//
// POST /config/personal
//
// Audience: Student, Teacher, Admin, Parent.
func (c *Client) SetPersonalSetting(ctx context.Context, name, value string) (*Setting, error) {
	return fetch[*Setting](ctx, c, post("config/personal", Setting{Name: name, Value: value}))
}

var yearPattern = regexp.MustCompile(`^\d{4}/\d{4}$`)

// ErrPeriodNotApplied means the server did not switch to the requested period.
var ErrPeriodNotApplied = errors.New("bars: server did not apply the selected period")

// SelectPeriod makes year ("2025/2026") and term the user's current period,
// writing only what differs and verifying the result.
func (c *Client) SelectPeriod(ctx context.Context, year string, term Term) (*User, error) {
	if !yearPattern.MatchString(year) {
		return nil, fmt.Errorf("bars: year must look like 2025/2026, got %q", year)
	}
	u, err := c.CurrentUser(ctx)
	if err != nil {
		return nil, err
	}
	if u.SelectedYear == year && u.SelectedTerm == term {
		return u, nil
	}
	if u.SelectedYear != year {
		if _, err := c.SetPersonalSetting(ctx, "current_year", year); err != nil {
			return nil, err
		}
	}
	if u.SelectedTerm != term {
		if _, err := c.SetPersonalSetting(ctx, "current_term", strconv.Itoa(int(term))); err != nil {
			return nil, err
		}
	}
	if u, err = c.CurrentUser(ctx); err != nil {
		return nil, err
	}
	if u.SelectedYear != year || u.SelectedTerm != term {
		return nil, ErrPeriodNotApplied
	}
	return u, nil
}

// WithPeriod selects the period and runs fn while no other WithPeriod call on
// this client can switch it. Requests inside fn may run concurrently.
func (c *Client) WithPeriod(ctx context.Context, year string, term Term, fn func(ctx context.Context) error) error {
	c.periodMu.Lock()
	defer c.periodMu.Unlock()
	if _, err := c.SelectPeriod(ctx, year, term); err != nil {
		return err
	}
	return fn(ctx)
}

// Disciplines returns the catalogue of the selected period. withPlansOnly
// keeps only disciplines that have checkpoint plans.
//
// GET /journal/disciplines
//
// Audience: Student, Teacher, Admin, Parent.
func (c *Client) Disciplines(ctx context.Context, withPlansOnly bool) ([]Discipline, error) {
	return fetch[[]Discipline](ctx, c, get("journal/disciplines", url.Values{"withCheckpointPlansOnly": {strconv.FormatBool(withPlansOnly)}}))
}

// GroupsAndFlows returns flows and groups of the selected period; disciplineID 0 means all.
//
// GET /journal/groups-and-flows
//
// Audience: Student, Teacher, Admin, Parent.
func (c *Client) GroupsAndFlows(ctx context.Context, disciplineID int64) ([]GroupOrFlow, error) {
	var qv url.Values
	if disciplineID != 0 {
		qv = url.Values{"disciplineId": {strconv.FormatInt(disciplineID, 10)}}
	}
	return fetch[[]GroupOrFlow](ctx, c, get("journal/groups-and-flows", qv))
}

// StudentJournal returns the user's own journal (a parent gets the child's).
// The plan must belong to the selected period, otherwise the server answers
// 401; 423 means the user is not a student in the selected period.
//
// GET /marks/{checkpointPlanId}/{type}/{identifier}/student
//
// Audience: Student, Parent.
func (c *Client) StudentJournal(ctx context.Context, planID int64, flowType, identifier string) (*StudentJournal, error) {
	path := "marks/" + strconv.FormatInt(planID, 10) + "/" + url.PathEscape(flowType) + "/" + url.PathEscape(identifier) + "/student"
	return fetch[*StudentJournal](ctx, c, get(path, nil))
}

// fetch performs r and decodes the body into a fresh T.
func fetch[T any](ctx context.Context, c *Client, r *rest.Request) (T, error) {
	var out T
	if err := c.do(ctx, r, &out); err != nil {
		var zero T
		return zero, err
	}
	return out, nil
}

func get(path string, qv url.Values) *rest.Request {
	return &rest.Request{Method: http.MethodGet, Path: path, Query: qv}
}

func post(path string, body any) *rest.Request {
	return &rest.Request{Method: http.MethodPost, Path: path, JSON: body}
}
