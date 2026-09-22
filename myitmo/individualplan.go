package myitmo

import (
	"context"
	"time"

	"github.com/kewldan/go-itmo/internal/jsonx"
)

// IndividualPlanService is the staff tool for individual study plans
// (/api/individual-plan). Every route is staff only; the service answers 403
// otherwise, and [IndividualPlanRole.Forbidden] tells whether the tool may be used.
//
// Editing works on a draft: [IndividualPlanService.CreateDraft] before the
// first change, module and discipline edits, then [IndividualPlanService.PublishDraft].
type IndividualPlanService struct{ c *Client }

// Plan statuses (IndividualPlanStatus.ID of a plan); 1..5 are used, the
// meaning of the others is not known.
const (
	// IndividualPlanStatusActive is a plan being implemented; choice editing is enabled.
	IndividualPlanStatusActive = 2
	// IndividualPlanStatusArchived is an archived plan.
	IndividualPlanStatusArchived = 4
)

// Work types of [IndividualPlanWorkType].
const (
	IndividualPlanWorkTypeSelfStudy = 22
	IndividualPlanWorkTypeContact   = 128
)

// IndividualPlanFlagPastChoice is the flag that enables editing choices of past semesters.
const IndividualPlanFlagPastChoice = "past_choice"

// Validation error filters of [IndividualPlanListParams].ErrorFilter.
const (
	IndividualPlanPlanContentErrors   = "plan_content_errors"
	IndividualPlanChoicePastErrors    = "choice_past_errors"
	IndividualPlanChoiceCurrentErrors = "choice_current_errors"
)

// IndividualPlanValidationStatusError is the IndividualPlanValidation.Status of a failed check.
const IndividualPlanValidationStatusError = "error"

// IndividualPlanRole is what the current employee may do in the tool.
type IndividualPlanRole struct {
	Forbidden bool `json:"forbidden"`
	CanCreate bool `json:"can_create"`
	IsAdmin   bool `json:"is_admin"`
	// EduLevels are the education level IDs the employee may work with.
	EduLevels []int64 `json:"edu_levels"`
}

// IndividualPlanProgram is an educational programme.
type IndividualPlanProgram struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	// EnrollmentYear is a year, sent as a number or a string.
	EnrollmentYear RawJSON `json:"enrollment_year"`
}

// IndividualPlanDirection is a field of study.
type IndividualPlanDirection struct {
	ID int64 `json:"id"`
	// Code is like "09.03.04".
	Code string `json:"code"`
	Name string `json:"name"`
}

// IndividualPlanImplementer is a department implementing a plan or a discipline.
type IndividualPlanImplementer struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	ShortName string `json:"short_name"`
}

// IndividualPlanStatus is the status of a plan or a discipline; which name is
// filled depends on the route.
type IndividualPlanStatus struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	ShortName string `json:"short_name"`
}

// IndividualPlanFlag is a feature flag of the service.
type IndividualPlanFlag struct {
	// Key is like [IndividualPlanFlagPastChoice].
	Key     string `json:"key"`
	Enabled bool   `json:"enabled"`
}

// IndividualPlanListParams filters [IndividualPlanService.List]. Zero values are omitted.
type IndividualPlanListParams struct {
	Status        *int64
	EPID          *int64
	DirectionID   *int64
	ImplementerID *int64
	Archive       bool
	Query         string
	// ErrorFilter is one of the IndividualPlan*Errors constants.
	ErrorFilter string
	// Limit is the page size (typically 15).
	Limit  int
	Offset int
}

// IndividualPlanList is a page of plans.
type IndividualPlanList struct {
	Count int                      `json:"count"`
	Items []IndividualPlanListItem `json:"items"`
}

// IndividualPlanListItem is a plan in the list.
type IndividualPlanListItem struct {
	ID          int64                     `json:"id"`
	ISU         *int64                    `json:"isu"`
	Fio         string                    `json:"fio"`
	EPDirection IndividualPlanDirection   `json:"ep_direction"`
	EP          IndividualPlanProgram     `json:"ep"`
	Implementer IndividualPlanImplementer `json:"implementer"`
	DateStart   *Date                     `json:"date_start"`
	DateEnd     *Date                     `json:"date_end"`
	// Status is nil for a plan without a student.
	Status   *IndividualPlanStatus `json:"status"`
	HasDraft bool                  `json:"has_draft"`
}

// IndividualPlanCreate is the body of [IndividualPlanService.Create].
type IndividualPlanCreate struct {
	StartYear     int64   `json:"start_year"`
	EPID          int64   `json:"ep_id"`
	DirectionID   int64   `json:"direction_id"`
	ImplementerID int64   `json:"implementer_id"`
	ISU           []int64 `json:"isu"`
	Manual        bool    `json:"manual"`
}

// IndividualPlanCreated identifies a created plan.
type IndividualPlanCreated struct {
	ID int64 `json:"id"`
}

// IndividualPlanInfo is the card of a plan.
type IndividualPlanInfo struct {
	ID          int64                      `json:"id"`
	ISU         *int64                     `json:"isu"`
	Fio         string                     `json:"fio"`
	Status      *IndividualPlanStatus      `json:"status"`
	DateStart   *Date                      `json:"date_start"`
	DateEnd     *Date                      `json:"date_end"`
	Implementer IndividualPlanImplementer  `json:"implementer"`
	Group       string                     `json:"group"`
	EP          IndividualPlanProgram      `json:"ep"`
	EPDirection IndividualPlanDirection    `json:"ep_direction"`
	EduStandard *IndividualPlanEduStandard `json:"edu_standard"`
	CanEdit     bool                       `json:"can_edit"`
}

// IndividualPlanEduStandard is the educational standard of a plan.
type IndividualPlanEduStandard struct {
	Name string `json:"name"`
}

// IndividualPlanContent is the published content or the draft of a plan.
type IndividualPlanContent struct {
	Plan              []IndividualPlanNode `json:"plan"`
	SemestersCount    int                  `json:"semesters_count"`
	UpdatedAt         *time.Time           `json:"updated_at"`
	FilledMinCapacity float64              `json:"filled_min_capacity"`
	FilledMaxCapacity float64              `json:"filled_max_capacity"`
	FullCapacity      *float64             `json:"full_capacity"`
	HasDraft          bool                 `json:"has_draft"`
	// ChangedChoice is sent for drafts only: publishing it changes the student's choice.
	ChangedChoice bool `json:"changed_choice"`
}

// IndividualPlanNode is a module or a discipline of the plan tree; fields not
// relevant to Type ("module" or "discipline") are absent.
type IndividualPlanNode struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Type      string `json:"type"`
	IsRoot    bool   `json:"is_root"`
	CanDelete bool   `json:"can_delete"`

	// Module fields.
	NameRu         string               `json:"name_ru"`
	NameEn         string               `json:"name_en"`
	BlockID        int64                `json:"block_id"`
	ParentID       *int64               `json:"parent_id"`
	Rule           *IndividualPlanRule  `json:"rule"`
	Children       []IndividualPlanNode `json:"children"`
	MinCapacity    float64              `json:"min_capacity"`
	MaxCapacity    float64              `json:"max_capacity"`
	TargetCapacity *float64             `json:"target_capacity"`
	// ChildrenTypes (root only) are the program_type_id values for
	// [IndividualPlanService.AvailableDisciplines].
	ChildrenTypes []int64 `json:"children_types"`
	// RuleTypes (root only) are the rule types allowed in the plan.
	RuleTypes []int64 `json:"rule_types"`

	// Discipline fields.
	ModuleID int64 `json:"module_id"`
	// Status.ID is 1..7.
	Status      *IndividualPlanStatus      `json:"status"`
	Implementer *IndividualPlanImplementer `json:"implementer"`
	// Capacity is in credit points.
	Capacity       float64 `json:"capacity"`
	SemestersCount int     `json:"semesters_count"`
	// Semesters are the 1-based semesters the discipline is placed in.
	Semesters       []int                          `json:"semesters"`
	NoFlowChoice    bool                           `json:"no_flow_choice"`
	Grades          []IndividualPlanGrade          `json:"grades"`
	Selection       *IndividualPlanSelection       `json:"selection"`
	ProgramContents []IndividualPlanProgramContent `json:"program_contents"`
}

// IndividualPlanRule is the selection rule of a module.
type IndividualPlanRule struct {
	// TypeID uses the StudyPlanRule constants (1 choose N, 2 all, 21 free, 41, 61 credit based).
	TypeID    int64  `json:"type_id"`
	TypeName  string `json:"type_name"`
	RuleValue []int  `json:"rule_value"`
}

// IndividualPlanGrade is a grade received for a discipline.
type IndividualPlanGrade struct {
	SemesterID int64 `json:"semester_id"`
	// Grade is a string or a number.
	Grade        RawJSON `json:"grade"`
	Letter       string  `json:"letter"`
	WorkTypeName string  `json:"work_type_name"`
}

// IndividualPlanSelection is the student's choice of a discipline.
type IndividualPlanSelection struct {
	// Semesters are 9-digit semester IDs (YYYYYYYYS).
	Semesters         []int64 `json:"semesters"`
	Protocol          bool    `json:"protocol"`
	ProtocolSemesters []int64 `json:"protocol_semesters"`
}

// IndividualPlanProgramContent is one part of a discipline.
type IndividualPlanProgramContent struct {
	ID               int64                    `json:"id"`
	ContentOrder     int                      `json:"content_order"`
	Capacity         float64                  `json:"capacity"`
	ProgramWorkTypes []IndividualPlanWorkType `json:"program_work_types"`
}

// IndividualPlanWorkType is the workload of one kind in a discipline part.
type IndividualPlanWorkType struct {
	ID           int64   `json:"id"`
	WorkTypeID   int64   `json:"work_type_id"`
	WorkTypeName string  `json:"work_type_name"`
	Hours        float64 `json:"hours"`
	IsControl    bool    `json:"is_control"`
}

// IndividualPlanChangesParams filters [IndividualPlanService.Changes]. Zero values are omitted.
type IndividualPlanChangesParams struct {
	// Category is "choice" or "plan".
	Category string
	Query    string
	Limit    int
	Offset   int
}

// IndividualPlanChanges is a page of the change history.
type IndividualPlanChanges struct {
	Count int                    `json:"count"`
	Items []IndividualPlanChange `json:"items"`
}

// IndividualPlanChange is one entry of the change history.
type IndividualPlanChange struct {
	ID   int64     `json:"id"`
	Date time.Time `json:"date"`
	// Category is "choice" or "plan".
	Category string `json:"category"`
	// EntityType is "module", "discipline", "choice" or "draft".
	EntityType     string                          `json:"entity_type"`
	Summary        string                          `json:"summary"`
	Author         *IndividualPlanAuthor           `json:"author"`
	DisciplineID   *int64                          `json:"discipline_id"`
	DisciplineName string                          `json:"discipline_name"`
	ModuleID       *int64                          `json:"module_id"`
	ModuleName     string                          `json:"module_name"`
	Description    IndividualPlanChangeDescription `json:"description"`
}

// IndividualPlanAuthor is the author of a change.
type IndividualPlanAuthor struct {
	ID  int64  `json:"id"`
	Fio string `json:"fio"`
}

// IndividualPlanChangeDescription is the state before and after a change; nil for creation or deletion.
type IndividualPlanChangeDescription struct {
	Before *IndividualPlanChangeState `json:"before"`
	After  *IndividualPlanChangeState `json:"after"`
}

// IndividualPlanChangeState is an entity snapshot in the change history; the
// fields present depend on the entity type.
type IndividualPlanChangeState struct {
	// Module: Name, NameEn, Rule, ParentModule.
	Name         string                      `json:"name"`
	NameEn       string                      `json:"name_en"`
	Rule         *IndividualPlanChangeRule   `json:"rule"`
	ParentModule *IndividualPlanChangeModule `json:"parent_module"`
	// Discipline: Module, Block, Semesters (1-based), FlowChoice.
	Module     *IndividualPlanChangeModule `json:"module"`
	Block      *IndividualPlanChangeBlock  `json:"block"`
	FlowChoice *bool                       `json:"flow_choice"`
	// Choice: Semesters (9-digit IDs), BlockName.
	Semesters []int64 `json:"semesters"`
	BlockName string  `json:"block_name"`
}

// IndividualPlanChangeRule is a module rule in the change history.
type IndividualPlanChangeRule struct {
	Name  string `json:"name"`
	Value []int  `json:"value"`
}

// IndividualPlanChangeModule is a module reference in the change history.
type IndividualPlanChangeModule struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	RuleName  string `json:"rule_name"`
	RuleValue []int  `json:"rule_value"`
}

// IndividualPlanChangeBlock is a block reference in the change history.
type IndividualPlanChangeBlock struct {
	Name string `json:"name"`
}

// IndividualPlanDisciplineSearch is the result of [IndividualPlanService.AvailableDisciplines].
type IndividualPlanDisciplineSearch struct {
	Count       int                                 `json:"count"`
	Disciplines []IndividualPlanAvailableDiscipline `json:"disciplines"`
}

// IndividualPlanAvailableDiscipline is a discipline that can be added to a module.
type IndividualPlanAvailableDiscipline struct {
	ID                   int64                          `json:"id"`
	Name                 string                         `json:"name"`
	Capacity             float64                        `json:"capacity"`
	ImplementerShortName string                         `json:"implementer_short_name"`
	SemestersCount       int                            `json:"semesters_count"`
	StatusName           string                         `json:"status_name"`
	AnnotationRu         string                         `json:"annotation_ru"`
	ProgramContents      []IndividualPlanProgramContent `json:"program_contents"`
}

// IndividualPlanModuleUse is a module where a discipline is used.
type IndividualPlanModuleUse struct {
	ID           int64  `json:"id"`
	NameRu       string `json:"name_ru"`
	RuleTypeName string `json:"rule_type_name"`
	RuleValue    []int  `json:"rule_value"`
}

// IndividualPlanModuleRef is a module that can receive a module or a discipline.
type IndividualPlanModuleRef struct {
	ID           int64  `json:"id"`
	BlockID      int64  `json:"block_id"`
	NameRu       string `json:"name_ru"`
	NameEn       string `json:"name_en"`
	ParentID     *int64 `json:"parent_id"`
	ParentName   string `json:"parent_name"`
	RuleTypeID   int64  `json:"rule_type_id"`
	RuleTypeName string `json:"rule_type_name"`
	RuleValue    []int  `json:"rule_value"`
}

// IndividualPlanModuleInput is the body of [IndividualPlanService.CreateModule]
// and [IndividualPlanService.UpdateModule].
type IndividualPlanModuleInput struct {
	NameRu string `json:"name_ru"`
	// NameEn must not contain Cyrillic.
	NameEn     string `json:"name_en"`
	RuleTypeID int64  `json:"rule_type_id"`
	// RuleValue is sent as null when nil (rules without values).
	RuleValue    IndividualPlanRuleValue `json:"rule_value"`
	ParentID     int64                   `json:"parent_id"`
	ParentName   string                  `json:"parent_name"`
	BlockID      int64                   `json:"block_id"`
	RuleTypeName string                  `json:"rule_type_name"`
}

// IndividualPlanRuleValue is the rule parameter list of a module input; nil is sent as null.
type IndividualPlanRuleValue []int

// MarshalJSON implements json.Marshaler.
func (v IndividualPlanRuleValue) MarshalJSON() ([]byte, error) {
	if v == nil {
		return []byte("null"), nil
	}
	return jsonx.Marshal([]int(v))
}

// IndividualPlanValidation is the result of a plan check.
type IndividualPlanValidation struct {
	// Status is [IndividualPlanValidationStatusError] when the check failed.
	Status string                         `json:"status"`
	Errors []IndividualPlanValidationItem `json:"errors"`
	// ImpossibleModules is sent by the selection check only.
	ImpossibleModules []int64 `json:"impossible_modules"`
}

// IndividualPlanValidationItem lists the errors of one module or discipline.
type IndividualPlanValidationItem struct {
	ModuleID       *int64                  `json:"module_id"`
	ModuleName     string                  `json:"module_name"`
	DisciplineID   *int64                  `json:"discipline_id"`
	DisciplineName string                  `json:"discipline_name"`
	Errors         []IndividualPlanMessage `json:"errors"`
}

// IndividualPlanMessage is a localised message.
type IndividualPlanMessage struct {
	Ru string `json:"ru"`
}

func individualPlanPath(planID int64) string {
	return "api/individual-plan/plan/" + id(planID)
}

func individualPlanDisciplinePath(planID, moduleID, disciplineID int64) string {
	return individualPlanPath(planID) + "/module/" + id(moduleID) + "/discipline/" + id(disciplineID)
}

// Role returns the permissions of the current employee. Staff only.
// GET /api/individual-plan/plan/role
func (s *IndividualPlanService) Role(ctx context.Context) (*IndividualPlanRole, error) {
	return call[*IndividualPlanRole](ctx, s.c, get("api/individual-plan/plan/role", nil))
}

// Programs returns the educational programmes for the list filter. Staff only.
// GET /api/individual-plan/plan/programs
func (s *IndividualPlanService) Programs(ctx context.Context) ([]IndividualPlanProgram, error) {
	return call[[]IndividualPlanProgram](ctx, s.c, get("api/individual-plan/plan/programs", nil))
}

// Directions returns the fields of study for the list filter. Staff only.
// GET /api/individual-plan/plan/directions
func (s *IndividualPlanService) Directions(ctx context.Context) ([]IndividualPlanDirection, error) {
	return call[[]IndividualPlanDirection](ctx, s.c, get("api/individual-plan/plan/directions", nil))
}

// Implementers returns the implementing departments. Staff only.
// GET /api/individual-plan/plan/implementers
func (s *IndividualPlanService) Implementers(ctx context.Context) ([]IndividualPlanImplementer, error) {
	return call[[]IndividualPlanImplementer](ctx, s.c, get("api/individual-plan/plan/implementers", nil))
}

// Statuses returns the plan statuses. Staff only.
// GET /api/individual-plan/plan/statuses
func (s *IndividualPlanService) Statuses(ctx context.Context) ([]IndividualPlanStatus, error) {
	return call[[]IndividualPlanStatus](ctx, s.c, get("api/individual-plan/plan/statuses", nil))
}

// CurrentSemester returns the 9-digit ID (YYYYYYYYS) of the current semester. Staff only.
// GET /api/individual-plan/semesters/current
func (s *IndividualPlanService) CurrentSemester(ctx context.Context) (int64, error) {
	r, err := call[struct {
		SemID int64 `json:"sem_id"`
	}](ctx, s.c, get("api/individual-plan/semesters/current", nil))
	return r.SemID, err
}

// Flags returns the feature flags of the service. Staff only.
// GET /api/individual-plan/flags
func (s *IndividualPlanService) Flags(ctx context.Context) ([]IndividualPlanFlag, error) {
	return call[[]IndividualPlanFlag](ctx, s.c, get("api/individual-plan/flags", nil))
}

// List returns a page of plans. Staff only.
// GET /api/individual-plan/plan
func (s *IndividualPlanService) List(ctx context.Context, p IndividualPlanListParams) (*IndividualPlanList, error) {
	qv := q().set("status", p.Status).set("ep_id", p.EPID).set("direction_id", p.DirectionID).
		set("implementer_id", p.ImplementerID).set("query", p.Query)
	if p.Archive {
		qv.set("archive", 1)
	} else {
		qv.set("archive", 0)
	}
	if p.Limit > 0 {
		qv.set("limit", p.Limit)
	}
	if p.Offset > 0 {
		qv.set("offset", p.Offset)
	}
	if p.ErrorFilter != "" {
		qv.set(p.ErrorFilter, 1)
	}
	return call[*IndividualPlanList](ctx, s.c, get("api/individual-plan/plan", qv))
}

// Create creates a plan and returns the created plans. Staff only (IndividualPlanRole.CanCreate).
// POST /api/individual-plan/plan
func (s *IndividualPlanService) Create(ctx context.Context, body IndividualPlanCreate) ([]IndividualPlanCreated, error) {
	if body.ISU == nil {
		body.ISU = []int64{}
	}
	return call[[]IndividualPlanCreated](ctx, s.c, post("api/individual-plan/plan", body))
}

// Get returns the card of a plan. Staff only.
// GET /api/individual-plan/plan/{plan_id}
func (s *IndividualPlanService) Get(ctx context.Context, planID int64) (*IndividualPlanInfo, error) {
	return call[*IndividualPlanInfo](ctx, s.c, get(individualPlanPath(planID), nil))
}

// Content returns the published content of a plan. Staff only.
// GET /api/individual-plan/plan/{plan_id}/plan
func (s *IndividualPlanService) Content(ctx context.Context, planID int64) (*IndividualPlanContent, error) {
	return call[*IndividualPlanContent](ctx, s.c, get(individualPlanPath(planID)+"/plan", nil))
}

// Draft returns the draft of a plan; the service answers an error when there is none. Staff only.
// GET /api/individual-plan/plan/{plan_id}/draft
func (s *IndividualPlanService) Draft(ctx context.Context, planID int64) (*IndividualPlanContent, error) {
	return call[*IndividualPlanContent](ctx, s.c, get(individualPlanPath(planID)+"/draft", nil))
}

// CreateDraft creates a draft before the first change. Staff only.
// POST /api/individual-plan/plan/{plan_id}/draft
func (s *IndividualPlanService) CreateDraft(ctx context.Context, planID int64) error {
	return exec(ctx, s.c, post(individualPlanPath(planID)+"/draft", nil))
}

// DeleteDraft discards the draft. Staff only.
// DELETE /api/individual-plan/plan/{plan_id}/draft
func (s *IndividualPlanService) DeleteDraft(ctx context.Context, planID int64) error {
	return exec(ctx, s.c, del(individualPlanPath(planID)+"/draft", nil))
}

// PublishDraft publishes the draft. Staff only.
// POST /api/individual-plan/plan/{plan_id}/draft/save
func (s *IndividualPlanService) PublishDraft(ctx context.Context, planID int64) error {
	return exec(ctx, s.c, post(individualPlanPath(planID)+"/draft/save", nil))
}

// Changes returns a page of the change history. Staff only.
// GET /api/individual-plan/plan/{plan_id}/changes
func (s *IndividualPlanService) Changes(ctx context.Context, planID int64, p IndividualPlanChangesParams) (*IndividualPlanChanges, error) {
	qv := q().set("category", p.Category).set("query", p.Query)
	if p.Limit > 0 {
		qv.set("limit", p.Limit)
	}
	if p.Offset > 0 {
		qv.set("offset", p.Offset)
	}
	return call[*IndividualPlanChanges](ctx, s.c, get(individualPlanPath(planID)+"/changes", qv))
}

// Semesters returns the 9-digit semester IDs of a plan; index+1 is the semester number. Staff only.
// GET /api/individual-plan/plan/{plan_id}/semesters
func (s *IndividualPlanService) Semesters(ctx context.Context, planID int64) ([]int64, error) {
	return call[[]int64](ctx, s.c, get(individualPlanPath(planID)+"/semesters", nil))
}

// AvailableDisciplines searches disciplines to add to a plan; programTypeIDs
// are the root module's ChildrenTypes. Staff only.
// GET /api/individual-plan/plan/{plan_id}/availableDisciplines
func (s *IndividualPlanService) AvailableDisciplines(ctx context.Context, planID int64, query string, programTypeIDs []int64) (*IndividualPlanDisciplineSearch, error) {
	return call[*IndividualPlanDisciplineSearch](ctx, s.c, get(individualPlanPath(planID)+"/availableDisciplines",
		q().set("query", query).set("program_type_id", programTypeIDs)))
}

// DisciplineUse returns the modules of the plan (or of its draft) that use a discipline. Staff only.
// GET /api/individual-plan/plan/{plan_id}/discipline/{discipline_id}/use
func (s *IndividualPlanService) DisciplineUse(ctx context.Context, planID, disciplineID int64, draft bool) ([]IndividualPlanModuleUse, error) {
	return call[[]IndividualPlanModuleUse](ctx, s.c, get(individualPlanPath(planID)+"/discipline/"+id(disciplineID)+"/use",
		q().set("draft", draft)))
}

// UpperModules returns the modules a module can be moved into. Staff only.
// GET /api/individual-plan/plan/{plan_id}/upperModules
func (s *IndividualPlanService) UpperModules(ctx context.Context, planID int64) ([]IndividualPlanModuleRef, error) {
	return call[[]IndividualPlanModuleRef](ctx, s.c, get(individualPlanPath(planID)+"/upperModules", nil))
}

// LowerModules returns the modules a discipline can be moved into. Staff only.
// GET /api/individual-plan/plan/{plan_id}/lowerModules
func (s *IndividualPlanService) LowerModules(ctx context.Context, planID int64) ([]IndividualPlanModuleRef, error) {
	return call[[]IndividualPlanModuleRef](ctx, s.c, get(individualPlanPath(planID)+"/lowerModules", nil))
}

// AddChoice records the student's choice of a discipline for the given
// 9-digit semester IDs. Staff only.
// POST /api/individual-plan/plan/{plan_id}/module/{module_id}/discipline/{discipline_id}/choice
func (s *IndividualPlanService) AddChoice(ctx context.Context, planID, moduleID, disciplineID int64, semesterIDs []int64) error {
	return exec(ctx, s.c, post(individualPlanDisciplinePath(planID, moduleID, disciplineID)+"/choice", semesterIDs))
}

// UpdateChoice changes the semesters of a choice. Staff only.
// PATCH /api/individual-plan/plan/{plan_id}/module/{module_id}/discipline/{discipline_id}/choice
func (s *IndividualPlanService) UpdateChoice(ctx context.Context, planID, moduleID, disciplineID int64, semesterIDs []int64) error {
	return exec(ctx, s.c, patch(individualPlanDisciplinePath(planID, moduleID, disciplineID)+"/choice", semesterIDs))
}

// DeleteChoice removes the choice of a discipline. Staff only.
// DELETE /api/individual-plan/plan/{plan_id}/module/{module_id}/discipline/{discipline_id}/choice
func (s *IndividualPlanService) DeleteChoice(ctx context.Context, planID, moduleID, disciplineID int64) error {
	return exec(ctx, s.c, del(individualPlanDisciplinePath(planID, moduleID, disciplineID)+"/choice", nil))
}

// CreateModule creates a module in the draft and returns its ID. Staff only.
// POST /api/individual-plan/plan/{plan_id}/module
func (s *IndividualPlanService) CreateModule(ctx context.Context, planID int64, m IndividualPlanModuleInput) (int64, error) {
	return call[int64](ctx, s.c, post(individualPlanPath(planID)+"/module", m))
}

// UpdateModule changes a module (including moving it under another parent) and returns its ID. Staff only.
// PATCH /api/individual-plan/plan/{plan_id}/module/{module_id}
func (s *IndividualPlanService) UpdateModule(ctx context.Context, planID, moduleID int64, m IndividualPlanModuleInput) (int64, error) {
	return call[int64](ctx, s.c, patch(individualPlanPath(planID)+"/module/"+id(moduleID), m))
}

// DeleteModule deletes a module with all its content. Staff only.
// DELETE /api/individual-plan/plan/{plan_id}/module/{module_id}
func (s *IndividualPlanService) DeleteModule(ctx context.Context, planID, moduleID int64) error {
	return exec(ctx, s.c, del(individualPlanPath(planID)+"/module/"+id(moduleID), nil))
}

// AddDisciplines adds disciplines to a module. Staff only.
// POST /api/individual-plan/plan/{plan_id}/module/{module_id}/discipline
func (s *IndividualPlanService) AddDisciplines(ctx context.Context, planID, moduleID int64, disciplineIDs []int64) error {
	return exec(ctx, s.c, post(individualPlanPath(planID)+"/module/"+id(moduleID)+"/discipline", disciplineIDs))
}

// DeleteDiscipline removes a discipline from a module. Staff only.
// DELETE /api/individual-plan/plan/{plan_id}/module/{module_id}/discipline/{discipline_id}
func (s *IndividualPlanService) DeleteDiscipline(ctx context.Context, planID, moduleID, disciplineID int64) error {
	return exec(ctx, s.c, del(individualPlanDisciplinePath(planID, moduleID, disciplineID), nil))
}

// SetDisciplineSemesters places a discipline in the given 1-based semesters
// (for multi-semester disciplines the start semester, or none). Staff only.
// PATCH /api/individual-plan/plan/{plan_id}/module/{module_id}/discipline/{discipline_id}
func (s *IndividualPlanService) SetDisciplineSemesters(ctx context.Context, planID, moduleID, disciplineID int64, semesters []int) error {
	return exec(ctx, s.c, patch(individualPlanDisciplinePath(planID, moduleID, disciplineID), semesters))
}

// MoveDiscipline moves a discipline to targetModuleID. Staff only.
// PATCH /api/individual-plan/plan/{plan_id}/module/{module_id}/discipline/{discipline_id}/move
func (s *IndividualPlanService) MoveDiscipline(ctx context.Context, planID, moduleID, disciplineID, targetModuleID int64) error {
	body := struct {
		ModuleID int64 `json:"module_id"`
	}{targetModuleID}
	return exec(ctx, s.c, patch(individualPlanDisciplinePath(planID, moduleID, disciplineID)+"/move", body))
}

// ToggleFlowChoice toggles the "no flow choice" flag (IndividualPlanNode.NoFlowChoice)
// of a discipline; allowed in rule-2 modules only. Staff only.
// PATCH /api/individual-plan/plan/{plan_id}/module/{module_id}/discipline/{discipline_id}/flowChoice
func (s *IndividualPlanService) ToggleFlowChoice(ctx context.Context, planID, moduleID, disciplineID int64) error {
	return exec(ctx, s.c, patch(individualPlanDisciplinePath(planID, moduleID, disciplineID)+"/flowChoice", nil))
}

// ValidatePlanContent checks the content of the plan or its draft; a failed
// check blocks publishing. Staff only.
// POST /api/individual-plan/plan/{plan_id}/plan/validate/plan-content
func (s *IndividualPlanService) ValidatePlanContent(ctx context.Context, planID int64, draft bool) (*IndividualPlanValidation, error) {
	return call[*IndividualPlanValidation](ctx, s.c, withQuery(post(individualPlanPath(planID)+"/plan/validate/plan-content", nil),
		q().set("draft", draft)))
}

// ValidatePersonalPlanContent checks the personal content of the plan or its draft. Staff only.
// POST /api/individual-plan/plan/{plan_id}/plan/validate/personal-plan-content
func (s *IndividualPlanService) ValidatePersonalPlanContent(ctx context.Context, planID int64, draft bool) (*IndividualPlanValidation, error) {
	return call[*IndividualPlanValidation](ctx, s.c, withQuery(post(individualPlanPath(planID)+"/plan/validate/personal-plan-content", nil),
		q().set("draft", draft)))
}

// ValidateSelection checks the student's choice against the module rules;
// considerStatedSemester false also checks the choice of past semesters. Staff only.
// POST /api/individual-plan/plan/{plan_id}/plan/validate/selection
func (s *IndividualPlanService) ValidateSelection(ctx context.Context, planID int64, draft, considerStatedSemester bool) (*IndividualPlanValidation, error) {
	return call[*IndividualPlanValidation](ctx, s.c, withQuery(post(individualPlanPath(planID)+"/plan/validate/selection", nil),
		q().set("draft", draft).set("consider_stated_semester", considerStatedSemester)))
}
