package myitmo

import (
	"context"
	"net/http"
	"time"
)

// ElectiveCourseService is an elective course enrolment service. MyITMO runs
// two services with the same API: introductory courses (/api/intro) and
// optional courses (/api/facultative).
type ElectiveCourseService struct {
	c      *Client
	prefix string
}

func (s *ElectiveCourseService) path(p string) string { return "api/" + s.prefix + "/" + p }

// ElectiveNode is a node of the election tree: module group, module,
// discipline, parent flow, middle flow or child flow, nested through Variants.
// Fields not relevant to a level are absent.
type ElectiveNode struct {
	// ID is the flow ID on flow levels (what Book and Commit take) and the discipline ID on discipline level.
	ID   int64  `json:"id"`
	Name string `json:"name"`
	// Required means the node's children must satisfy Selections.
	Required bool `json:"required"`
	// Selections are the allowed counts of selected children, e.g. [1] or [2,3].
	Selections []int          `json:"selections"`
	Variants   []ElectiveNode `json:"variants"`
	// ColorIndex is the module colour.
	ColorIndex *int `json:"color_index"`
	// WorkType is the flow work type: 1 lecture, 2 lab, 3 practice, 5 exam,
	// 6 credit, 9 graded credit, 10 and 12 consultation.
	WorkType *int     `json:"work_type"`
	Teachers []string `json:"teachers"`
	// Language is the discipline language code, e.g. "ru" or "en".
	Language string `json:"language"`
	// FlowInfo is a discipline problem code; 1 means no flow is free of clashes with the main timetable.
	FlowInfo *int `json:"flow_info"`
	// Selectable is a discipline-level flag (checker view).
	Selectable bool `json:"selectable"`
	// AvailableSelections are the allowed exact sets of child IDs (checker view).
	AvailableSelections [][]int64 `json:"available_selections"`
	// Hours may be absent; its presence is not confirmed.
	Hours *float64 `json:"hours"`
}

// Election state values of ElectiveCurrent.Status.
const (
	ElectiveStatusNone      = 0 // nothing saved
	ElectiveStatusBooked    = 1 // booked, not confirmed; reset automatically when time runs out
	ElectiveStatusConfirmed = 2
)

// ElectiveCurrent is the user's election state.
type ElectiveCurrent struct {
	// Status is one of the ElectiveStatus constants.
	Status int `json:"status"`
	// FlowID lists the selected flow IDs.
	FlowID        []int64                `json:"flow_id"`
	Intersections []ElectiveIntersection `json:"intersections"`
}

// Clash kinds of ElectiveIntersection.IntersectionType.
const (
	ElectiveClashSoft = 1 // partial overlap
	ElectiveClashHard = 2 // conflicting classes
)

// ElectiveIntersection is a timetable clash between two flows.
type ElectiveIntersection struct {
	Flow1 ElectiveIntersectionFlow `json:"flow1"`
	Flow2 ElectiveIntersectionFlow `json:"flow2"`
	// IntersectionType is ElectiveClashSoft or ElectiveClashHard.
	IntersectionType int `json:"intersection_type"`
}

// ElectiveIntersectionFlow is one side of a clash.
type ElectiveIntersectionFlow struct {
	FlowID    int64     `json:"flow_id"`
	Date      time.Time `json:"date"`
	DateStart time.Time `json:"date_start"`
}

// ElectiveCodeClashes is the error_code of Book when the flows were booked
// but clash with the timetable.
const ElectiveCodeClashes = 109

// ElectiveBooking is the outcome of Book.
type ElectiveBooking struct {
	// Clashes are the new timetable clashes; set only when the server answered ElectiveCodeClashes.
	Clashes []ElectiveIntersection
}

// ElectiveCommitAction is the status sent to Commit.
type ElectiveCommitAction int

// Commit actions.
const (
	ElectiveCommitReset     ElectiveCommitAction = 0 // drop the choice
	ElectiveCommitUnconfirm ElectiveCommitAction = 1 // back to booked, not confirmed
	ElectiveCommitConfirm   ElectiveCommitAction = 2
)

type electiveBookingBody struct {
	SelectedFlows []int64 `json:"selected_flows"`
	CanceledFlows []int64 `json:"canceled_flows"`
}

type electiveCommitBody struct {
	Status ElectiveCommitAction `json:"status"`
	FlowID []int64              `json:"flow_id"`
}

// ElectiveDescription is the short description of a discipline.
type ElectiveDescription struct {
	// ID is the discipline ID.
	ID          int64  `json:"id"`
	Description string `json:"description"`
}

// ElectiveCatalogGroup is a module group of the catalogue shown before the
// election opens.
type ElectiveCatalogGroup struct {
	ID       int64                   `json:"id"`
	Name     string                  `json:"name"`
	Variants []ElectiveCatalogModule `json:"variants"`
}

// ElectiveCatalogModule is a module of the catalogue.
type ElectiveCatalogModule struct {
	Name     string                      `json:"name"`
	Variants []ElectiveCatalogDiscipline `json:"variants"`
}

// ElectiveCatalogDiscipline is a discipline of the catalogue.
type ElectiveCatalogDiscipline struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	// URL is the "more" link.
	URL string `json:"url"`
}

// ElectiveLimits is the seat limits and the booking window.
type ElectiveLimits struct {
	// Limits is keyed by flow ID.
	Limits    map[int64]ElectiveFlowLimit `json:"limits"`
	StartTime time.Time                   `json:"start_time"`
	EndTime   time.Time                   `json:"end_time"`
}

// ElectiveFlowLimit is the seats of a flow.
type ElectiveFlowLimit struct {
	// Limit is the remaining seats; a flow cannot be booked when it is <= 0.
	Limit    *int `json:"limit"`
	LimitMax int  `json:"limit_max"`
}

// ElectiveFlowLimitEntry is the seats of a flow in the checker view.
type ElectiveFlowLimitEntry struct {
	FlowID   int64 `json:"flow_id"`
	Limit    *int  `json:"limit"`
	LimitMax int   `json:"limit_max"`
}

// Choice states of ElectiveSemester.ChoiceStatus.
const (
	ElectiveChoiceOpen         = 0
	ElectiveChoiceNotOpened    = 97  // not open yet; the catalogue is shown
	ElectiveChoiceNotOpenedAlt = 98  // not open yet; the catalogue is shown
	ElectiveChoiceNotAvailable = 104 // no election for this user
)

// Values of ElectiveSemester.Semester.
const (
	ElectiveSemesterSpring = 0
	ElectiveSemesterAutumn = 1
)

// ElectiveSemester is the election period.
type ElectiveSemester struct {
	DateStart Date `json:"date_start"`
	DateEnd   Date `json:"date_end"`
	// ChoiceStatus is one of the ElectiveChoice constants.
	ChoiceStatus int `json:"choice_status"`
	// Semester is ElectiveSemesterSpring or ElectiveSemesterAutumn.
	Semester  int    `json:"semester"`
	StudyYear string `json:"study_year"`
}

type electiveTree struct {
	JSON []ElectiveNode `json:"json"`
}

type electiveSchedule struct {
	JSON []ScheduleDay `json:"json"`
}

// Tree returns the user's election tree; the top-level nodes are module groups.
// GET /api/{intro|facultative}/json/
func (s *ElectiveCourseService) Tree(ctx context.Context) ([]ElectiveNode, error) {
	r, err := call[electiveTree](ctx, s.c, get(s.path("json/"), nil))
	return r.JSON, err
}

// TreeOf returns another student's election tree (checker view).
// GET /api/{intro|facultative}/json/{isu}
// Staff only.
func (s *ElectiveCourseService) TreeOf(ctx context.Context, isu int64) ([]ElectiveNode, error) {
	r, err := call[electiveTree](ctx, s.c, get(s.path("json/"+id(isu)), nil))
	return r.JSON, err
}

// Current returns the user's election state: selected flows, status and clashes.
// GET /api/{intro|facultative}/current/
func (s *ElectiveCourseService) Current(ctx context.Context) (*ElectiveCurrent, error) {
	return call[*ElectiveCurrent](ctx, s.c, get(s.path("current/"), nil))
}

// CurrentOf returns another student's election state (checker view).
// GET /api/{intro|facultative}/current/{isu}
// Staff only.
func (s *ElectiveCourseService) CurrentOf(ctx context.Context, isu int64) (*ElectiveCurrent, error) {
	return call[*ElectiveCurrent](ctx, s.c, get(s.path("current/"+id(isu)), nil))
}

// Book reserves the selected flows and releases the canceled ones; the
// election becomes ElectiveStatusBooked. When the flows were booked but clash
// with the timetable, it returns both the booking with Clashes filled and an
// *Error whose Code is ElectiveCodeClashes.
// POST /api/{intro|facultative}/booking/
func (s *ElectiveCourseService) Book(ctx context.Context, selected, canceled []int64) (*ElectiveBooking, error) {
	if selected == nil {
		selected = []int64{}
	}
	if canceled == nil {
		canceled = []int64{}
	}
	r := post(s.path("booking/"), electiveBookingBody{SelectedFlows: selected, CanceledFlows: canceled})
	env, err := callRaw[envelope[jsonValue]](ctx, s.c, r)
	if err != nil {
		return nil, err
	}
	switch env.ErrorCode {
	case 0:
		return &ElectiveBooking{}, nil
	case ElectiveCodeClashes:
		b := &ElectiveBooking{}
		if err := decodeInto(env.Result, &b.Clashes); err != nil {
			return nil, &DecodeError{Method: r.Method, Path: r.Path, Err: err}
		}
		return b, &Error{Method: r.Method, Path: r.Path, StatusCode: http.StatusOK, Code: env.ErrorCode, Message: env.ErrorMessage}
	default:
		return nil, &Error{Method: r.Method, Path: r.Path, StatusCode: http.StatusOK, Code: env.ErrorCode, Message: env.ErrorMessage, Details: details(env.Result)}
	}
}

// Commit confirms, unconfirms or resets the election. flowIDs are all
// currently selected flows. The server refuses to confirm while required
// groups are unsatisfied or clashes exist.
// POST /api/{intro|facultative}/commit/
func (s *ElectiveCourseService) Commit(ctx context.Context, action ElectiveCommitAction, flowIDs []int64) error {
	if flowIDs == nil {
		flowIDs = []int64{}
	}
	return exec(ctx, s.c, post(s.path("commit/"), electiveCommitBody{Status: action, FlowID: flowIDs}))
}

// FlowSchedule returns the classes of one flow, for previewing it in the calendar.
// GET /api/{intro|facultative}/schedule/flow/{flowId}/
func (s *ElectiveCourseService) FlowSchedule(ctx context.Context, flowID int64) ([]ScheduleDay, error) {
	r, err := call[electiveSchedule](ctx, s.c, get(s.path("schedule/flow/"+id(flowID)+"/"), nil))
	return r.JSON, err
}

// UserSchedule returns the user's main timetable, shown next to candidate flows so clashes are visible.
// GET /api/{intro|facultative}/schedule/user/
func (s *ElectiveCourseService) UserSchedule(ctx context.Context) ([]ScheduleDay, error) {
	r, err := call[electiveSchedule](ctx, s.c, get(s.path("schedule/user/"), nil))
	return r.JSON, err
}

// Descriptions returns the short descriptions of the disciplines.
// GET /api/{intro|facultative}/description/
func (s *ElectiveCourseService) Descriptions(ctx context.Context) ([]ElectiveDescription, error) {
	r, err := call[struct {
		Description []ElectiveDescription `json:"description"`
	}](ctx, s.c, get(s.path("description/"), nil))
	return r.Description, err
}

// Catalog returns the structured discipline catalogue shown before the election opens.
// GET /api/{intro|facultative}/description/start/
func (s *ElectiveCourseService) Catalog(ctx context.Context) ([]ElectiveCatalogGroup, error) {
	r, err := call[struct {
		Description []ElectiveCatalogGroup `json:"description"`
	}](ctx, s.c, get(s.path("description/start/"), nil))
	return r.Description, err
}

// Limits returns the remaining seats per flow and the booking window.
// GET /api/{intro|facultative}/limits/
func (s *ElectiveCourseService) Limits(ctx context.Context) (*ElectiveLimits, error) {
	return call[*ElectiveLimits](ctx, s.c, get(s.path("limits/"), nil))
}

// LimitsOf returns the flow seats for the checker view of a student.
// GET /api/{intro|facultative}/limits/{isu}
// Staff only.
func (s *ElectiveCourseService) LimitsOf(ctx context.Context, isu int64) ([]ElectiveFlowLimitEntry, error) {
	r, err := call[struct {
		Limits []ElectiveFlowLimitEntry `json:"limits"`
	}](ctx, s.c, get(s.path("limits/"+id(isu)), nil))
	return r.Limits, err
}

// Status returns the election period and whether the choice is open.
// GET /api/{intro|facultative}/status/
func (s *ElectiveCourseService) Status(ctx context.Context) (*ElectiveSemester, error) {
	r, err := call[struct {
		Semester *ElectiveSemester `json:"semester"`
	}](ctx, s.c, get(s.path("status/"), nil))
	return r.Semester, err
}
