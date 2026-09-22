package myitmo

import (
	"context"
	"time"
)

// ElectionService is elective discipline and flow selection (/api/election).
//
// A campaign has two stages: the student first picks disciplines (their
// groupFlow values, [ElectionService.SelectDisciplines]), the server then
// builds flow chains, and the student picks flows ([ElectionService.ChangeFlows]).
type ElectionService struct{ c *Client }

// Error codes of [ElectionService.SelectDisciplines], read with [ErrorCode].
const (
	// ElectionErrNoFlows means no flows are available for the chosen disciplines.
	ElectionErrNoFlows = 105
	// ElectionErrNoDisjointFlows means no set of non-overlapping flows exists.
	ElectionErrNoDisjointFlows = 106
	// ElectionErrNoPlaces means some of the chosen disciplines ran out of places.
	ElectionErrNoPlaces = 109
)

// ElectionAvailability is the state and schedule of the current campaign.
type ElectionAvailability struct {
	// ID is the campaign state; [ElectionOpen] means selection is open.
	ID            int       `json:"id"`
	Status        string    `json:"status"`
	SemesterStart time.Time `json:"semesterStart"`
	SemesterEnd   time.Time `json:"semesterEnd"`
	DateStart     time.Time `json:"dateStart"`
	DateEnd       time.Time `json:"dateEnd"`
	// TimeStart and TimeEnd are wall-clock times of the campaign bounds.
	TimeStart  string `json:"timeStart"`
	TimeEnd    string `json:"timeEnd"`
	StudyYear  string `json:"studyYear"`
	SemesterID int64  `json:"semesterId"`
	// Semester is the parity: 0 spring, 1 autumn.
	Semester int `json:"semester"`
}

// ElectionOpen is the ElectionAvailability.ID of an open campaign.
const ElectionOpen = 1

// ElectionAvailableDiscipline is a discipline offered in the current campaign.
type ElectionAvailableDiscipline struct {
	DcID         int64  `json:"dcId"`
	DiscID       int64  `json:"discId"`
	LangID       int64  `json:"langId"`
	DepName      string `json:"depName"`
	DepNameShort string `json:"depNameShort"`
	DiscName     string `json:"discName"`
	// LangCode is like "RU" or "EN".
	LangCode    string                                `json:"langCode"`
	Required    bool                                  `json:"required"`
	Description string                                `json:"description"`
	Semesters   []ElectionAvailableDisciplineSemester `json:"semesters"`
	// NotCompatibleWith lists DiscID values that cannot be chosen together with this one.
	NotCompatibleWith []int64 `json:"notCompatibleWith"`
}

// ElectionAvailableDisciplineSemester is the availability of a discipline in one semester.
type ElectionAvailableDisciplineSemester struct {
	Semester int `json:"semester"`
	// StatusID is [ElectionDisciplineAvailable] when the discipline can be chosen;
	// otherwise StatusName says why not.
	StatusID   int    `json:"statusId"`
	StatusName string `json:"statusName"`
	// GroupFlow is the opaque value sent to [ElectionService.SelectDisciplines].
	GroupFlow string `json:"groupFlow"`
	// Themes is set for disciplines split into themes; Themes[0] is the relevant one.
	Themes []ElectionThemeRoot `json:"themes"`
}

// ElectionDisciplineAvailable is the StatusID of a discipline that can be chosen.
const ElectionDisciplineAvailable = 1

// ElectionThemeRoot groups the theme blocks of a discipline.
type ElectionThemeRoot struct {
	Children []ElectionThemeBlock `json:"children"`
}

// ElectionThemeBlock is a block of themes with a selection rule.
type ElectionThemeBlock struct {
	ID int64 `json:"id"`
	// RuleID is one of the ElectionRule constants.
	RuleID int `json:"ruleId"`
	// RuleParams lists the allowed numbers of themes for [ElectionRuleChooseN].
	RuleParams []int           `json:"ruleParams"`
	Themes     []ElectionTheme `json:"themes"`
}

// Theme selection rules of [ElectionThemeBlock].
const (
	// ElectionRuleChooseN means choose N themes, N one of RuleParams.
	ElectionRuleChooseN = 1
	// ElectionRuleAll means every theme must be chosen.
	ElectionRuleAll = 2
	// ElectionRuleFree means any number of themes.
	ElectionRuleFree = 21
)

// ElectionTheme is one selectable theme.
type ElectionTheme struct {
	// GroupFlow is sent to [ElectionService.SelectDisciplines] like a discipline's.
	GroupFlow string `json:"groupFlow"`
	Name      string `json:"name"`
}

// ElectionSelectionValidation is the server check of a set of chosen disciplines.
type ElectionSelectionValidation struct {
	// Disciplines are the DiscID values the selection resolves to.
	Disciplines            []int64 `json:"disciplines"`
	ValidVariantSelection  bool    `json:"validVariantSelection"`
	ValidRequiredSelection bool    `json:"validRequiredSelection"`
	NeedSelectRequired     int     `json:"needSelectRequired"`
	NeedSelectVariants     int     `json:"needSelectVariants"`
	NeedSelectVariantsMax  int     `json:"needSelectVariantsMax"`
	ScheduleAvailable      bool    `json:"scheduleAvailable"`
}

// ElectionFlowLimit is the capacity of a flow or a group flow.
type ElectionFlowLimit struct {
	LimitMax int64 `json:"limitMax"`
	// Occupied is not sent for group flows.
	Occupied int64 `json:"occupied"`
	// Free may be negative; treat it as 0.
	Free int64 `json:"free"`
}

// ElectionChangeResult is the result of changing the selection.
type ElectionChangeResult struct {
	// Status is an undocumented server status.
	Status *int64 `json:"status"`
	Name   string `json:"name"`
	// Flows is the resulting list of chosen flows.
	Flows []int64 `json:"flows"`
}

// ElectionFlowChain is a chosen discipline with the tree of its flows.
type ElectionFlowChain struct {
	GroupFlow      string         `json:"groupFlow"`
	DisciplineID   int64          `json:"disciplineId"`
	DisciplineName string         `json:"disciplineName"`
	Flows          []ElectionFlow `json:"flows"`
}

// ElectionFlow is a node of the recursive flow tree.
type ElectionFlow struct {
	ID       int64    `json:"id"`
	Name     string   `json:"name"`
	LimitMax *int64   `json:"limitMax"`
	Teachers []string `json:"teachers"`
	// Variants are the alternatives below this node.
	Variants []ElectionFlow `json:"variants"`
	// WorkType is one of the ElectionWorkType constants; 0 for grouping nodes.
	WorkType   int   `json:"workType"`
	Available  bool  `json:"available"`
	Selections []int `json:"selections"`
}

// Work types of [ElectionFlow] and [ElectionLesson].
const (
	ElectionWorkTypeLecture       = 1
	ElectionWorkTypeLab           = 2
	ElectionWorkTypePractice      = 3
	ElectionWorkTypeExam          = 5
	ElectionWorkTypeCredit        = 6
	ElectionWorkTypeDiffCredit    = 9
	ElectionWorkTypeConsultation  = 10
	ElectionWorkTypeConsultation2 = 12
)

// ElectionSelectedFlowChains is the fixed selection shown when the campaign is closed.
type ElectionSelectedFlowChains struct {
	// SelectedBy is who made the choice; [ElectionSelectedByIndividualPlan]
	// means it comes from the individual study plan.
	SelectedBy *int64                      `json:"selectedBy"`
	FlowChains []ElectionSelectedFlowChain `json:"flowChains"`
}

// ElectionSelectedByIndividualPlan is the SelectedBy value of a choice made through the individual plan.
const ElectionSelectedByIndividualPlan = 2

// ElectionSelectedFlowChain is a discipline and its chosen flows.
type ElectionSelectedFlowChain struct {
	DisciplineID int64                  `json:"disciplineId"`
	DiscName     string                 `json:"discName"`
	Flows        []ElectionSelectedFlow `json:"flows"`
}

// ElectionSelectedFlow is a short flow reference.
type ElectionSelectedFlow struct {
	FlowID   int64  `json:"flowId"`
	FlowName string `json:"flowName"`
}

// ElectionFlowIntersection is a schedule conflict between two flows.
type ElectionFlowIntersection struct {
	Flow1 ElectionIntersectingFlow `json:"flow1"`
	Flow2 ElectionIntersectingFlow `json:"flow2"`
	// IntersectionType is 1, 2 or 3: Flow1 conflicts for 1, Flow2 for 3.
	IntersectionType int `json:"intersectionType"`
}

// ElectionIntersectingFlow is one side of an [ElectionFlowIntersection].
type ElectionIntersectingFlow struct {
	ID        int64     `json:"id"`
	DateStart time.Time `json:"dateStart"`
}

// ElectionLesson is a lesson of the combined schedule (base schedule plus chosen flows).
type ElectionLesson struct {
	TimeStart      time.Time `json:"timeStart"`
	TimeEnd        time.Time `json:"timeEnd"`
	DateStart      time.Time `json:"dateStart"`
	DisciplineID   int64     `json:"disciplineId"`
	DisciplineName string    `json:"disciplineName"`
	// WorkTypeID is one of the ElectionWorkType constants.
	WorkTypeID   int    `json:"workTypeId"`
	TeacherFio   string `json:"teacherFio"`
	RoomName     string `json:"roomName"`
	BuildingName string `json:"buildingName"`
	FlowID       int64  `json:"flowId"`
}

// Availability returns the state and schedule of the current campaign.
// GET /api/election/students/availability
func (s *ElectionService) Availability(ctx context.Context) (*ElectionAvailability, error) {
	return call[*ElectionAvailability](ctx, s.c, get("api/election/students/availability", nil))
}

// SelectedFlowChains returns the fixed selection of a closed campaign; nil when there is none.
// GET /api/election/students/selected_flow_chains
//
// Deprecated: the endpoint is meant for closed campaigns only;
// use [ElectionService.OrderedFlowChains] and [ElectionService.ChosenFlows].
func (s *ElectionService) SelectedFlowChains(ctx context.Context) (*ElectionSelectedFlowChains, error) {
	return call[*ElectionSelectedFlowChains](ctx, s.c, get("api/election/students/selected_flow_chains", nil))
}

// AvailableDisciplines returns the disciplines that can be chosen.
// GET /api/election/students/available_disciplines
func (s *ElectionService) AvailableDisciplines(ctx context.Context) ([]ElectionAvailableDiscipline, error) {
	return call[[]ElectionAvailableDiscipline](ctx, s.c, get("api/election/students/available_disciplines", nil))
}

// ValidateDisciplines checks a set of groupFlow values without saving it.
// POST /api/election/students/group_flow_available_disciplines
func (s *ElectionService) ValidateDisciplines(ctx context.Context, groupFlows []string) (*ElectionSelectionValidation, error) {
	return call[*ElectionSelectionValidation](ctx, s.c, post("api/election/students/group_flow_available_disciplines", groupFlows))
}

// FlowLimits returns the capacity of every flow, keyed by flow ID.
// GET /api/election/students/limits/flows
func (s *ElectionService) FlowLimits(ctx context.Context) (map[int64]ElectionFlowLimit, error) {
	return call[map[int64]ElectionFlowLimit](ctx, s.c, get("api/election/students/limits/flows", nil))
}

// FlowGroupLimits returns the capacity of every group flow, keyed by groupFlow.
// GET /api/election/students/limits/flow_groups
func (s *ElectionService) FlowGroupLimits(ctx context.Context) (map[string]ElectionFlowLimit, error) {
	return call[map[string]ElectionFlowLimit](ctx, s.c, get("api/election/students/limits/flow_groups", nil))
}

// SelectDisciplines saves the chosen disciplines (groupFlow values) and builds the flow chains.
// Failures carry the ElectionErr codes.
// POST /api/election/students/order/
func (s *ElectionService) SelectDisciplines(ctx context.Context, groupFlows []string) (*ElectionChangeResult, error) {
	return call[*ElectionChangeResult](ctx, s.c, post("api/election/students/order/", groupFlows))
}

// ClearSelection resets the selection back to the discipline stage.
// POST /api/election/students/order/clear
func (s *ElectionService) ClearSelection(ctx context.Context) error {
	return exec(ctx, s.c, post("api/election/students/order/clear", nil))
}

// ChangeFlows replaces the chosen flows with flowIDs (the full selection).
// POST /api/election/students/order/change
func (s *ElectionService) ChangeFlows(ctx context.Context, flowIDs []int64) (*ElectionChangeResult, error) {
	return call[*ElectionChangeResult](ctx, s.c, post("api/election/students/order/change", flowIDs))
}

// ChosenFlows returns the IDs of the flows already chosen.
// GET /api/election/students/chosen_flows
func (s *ElectionService) ChosenFlows(ctx context.Context) ([]int64, error) {
	return call[[]int64](ctx, s.c, get("api/election/students/chosen_flows", nil))
}

// OrderedFlowChains returns the chosen disciplines with their flow trees.
// GET /api/election/students/ordered_flow_chains
func (s *ElectionService) OrderedFlowChains(ctx context.Context) ([]ElectionFlowChain, error) {
	return call[[]ElectionFlowChain](ctx, s.c, get("api/election/students/ordered_flow_chains", nil))
}

// FlowIntersections returns the schedule conflicts of a set of flows; a
// selection with conflicts should not be saved.
// POST /api/election/students/schedule/flows/intersections
func (s *ElectionService) FlowIntersections(ctx context.Context, flowIDs []int64) ([]ElectionFlowIntersection, error) {
	return call[[]ElectionFlowIntersection](ctx, s.c, post("api/election/students/schedule/flows/intersections", flowIDs))
}

// BaseTimeline returns the weeks with lessons of the base schedule. The
// elements are week keys ("YYYY-MM-DD", not confirmed);
// zero dates are omitted.
// GET /api/election/students/schedule/base/timeline
func (s *ElectionService) BaseTimeline(ctx context.Context, dateStart, dateEnd Date) ([]string, error) {
	return call[[]string](ctx, s.c, get("api/election/students/schedule/base/timeline",
		q().set("date_start", dateStart).set("date_end", dateEnd)))
}

// FlowTimeline returns the weeks with lessons of a flow, like [ElectionService.BaseTimeline].
// GET /api/election/students/schedule/flows/{flow_id}/timeline
func (s *ElectionService) FlowTimeline(ctx context.Context, flowID int64, dateStart, dateEnd Date) ([]string, error) {
	return call[[]string](ctx, s.c, get("api/election/students/schedule/flows/"+id(flowID)+"/timeline",
		q().set("date_start", dateStart).set("date_end", dateEnd)))
}

// CombinedSchedule returns the base schedule merged with the lessons of flowIDs for a period.
// POST /api/election/students/schedule/combined
func (s *ElectionService) CombinedSchedule(ctx context.Context, dateStart, dateEnd Date, flowIDs []int64) ([]ElectionLesson, error) {
	return call[[]ElectionLesson](ctx, s.c, withQuery(post("api/election/students/schedule/combined", flowIDs),
		q().set("date_start", dateStart).set("date_end", dateEnd)))
}
