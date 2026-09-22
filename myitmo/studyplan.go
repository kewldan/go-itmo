package myitmo

import "context"

// StudyPlanService is the curriculum (/api/eduPlanNew and the legacy /api/eduPlan).
type StudyPlanService struct{ c *Client }

// StudyPlanPrograms lists the plans available to the user.
type StudyPlanPrograms struct {
	ISU      int64              `json:"isu"`
	Programs []StudyPlanProgram `json:"programs"`
}

// StudyPlanProgram is a programme and its plan.
type StudyPlanProgram struct {
	PlanID int64 `json:"planId"`
	// SpecializationID is nil for programmes without a specialisation.
	SpecializationID *int64 `json:"specializationId"`
	Name             string `json:"name"`
	IsActive         bool   `json:"isActive"`
}

// StudyPlan is the full recursive structure of a plan.
type StudyPlan struct {
	ID                int64               `json:"id"`
	CurrentSemester   int                 `json:"currentSemester"`
	CurrentSemesterID int64               `json:"currentSemesterId"`
	SemestersCount    int                 `json:"semestersCount"`
	PlanInfo          StudyPlanInfo       `json:"planInfo"`
	Semesters         []StudyPlanSemester `json:"semesters"`
	Structure         []StudyPlanNode     `json:"structure"`
}

// StudyPlanInfo describes the programme of a plan.
type StudyPlanInfo struct {
	// DirectionCode is like "09.03.04".
	DirectionCode      string `json:"directionCode"`
	DirectionName      string `json:"directionName"`
	LevelQualification string `json:"levelQualification"`
	PlanType           string `json:"planType"`
	ProgramName        string `json:"programName"`
	StartYear          int    `json:"startYear"`
}

// StudyPlanSemester is a semester of a plan.
type StudyPlanSemester struct {
	Semester   int   `json:"semester"`
	SemesterID int64 `json:"semesterId"`
	// SemesterParity is 0 or 1.
	SemesterParity int    `json:"semesterParity"`
	StudyYear      string `json:"studyYear"`
}

// StudyPlanNode is a block, module or discipline; fields not relevant to Type are absent.
type StudyPlanNode struct {
	ID                      int64                `json:"id"`
	Name                    string               `json:"name"`
	Type                    string               `json:"type"`
	ModuleID                *int64               `json:"moduleId"`
	BlockID                 *int64               `json:"blockId"`
	BlockName               string               `json:"blockName"`
	ChoiceParameterID       *int64               `json:"choiceParameterId"`
	ChoiceParameterName     string               `json:"choiceParameterName"`
	ChoiceAvailable         *bool                `json:"choiceAvailable"`
	FlowSelectable          *bool                `json:"flowSelectable"`
	Replaceable             *bool                `json:"replaceable"`
	StartSemesterSelectable *bool                `json:"startSemesterSelectable"`
	CreditPoints            *int                 `json:"creditPoints"`
	DisciplineDuration      *int                 `json:"disciplineDuration"`
	Description             string               `json:"description"`
	LangCode                string               `json:"langCode"`
	LangName                string               `json:"langName"`
	RPDURL                  string               `json:"rpdUrl"`
	Department              *StudyPlanDepartment `json:"department"`
	Rules                   []int64              `json:"rules"`
	Children                []StudyPlanNode      `json:"children"`
	// Contents maps a semester number (as a string) to the workload in it.
	Contents map[string][]StudyPlanContent `json:"contents"`
}

// StudyPlanDepartment is the department responsible for a node.
type StudyPlanDepartment struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	ShortName string `json:"shortName"`
}

// StudyPlanContent is the workload of a discipline in one semester.
type StudyPlanContent struct {
	ID           int64               `json:"id"`
	ModuleID     int64               `json:"moduleId"`
	DisciplineID int64               `json:"disciplineId"`
	Order        int                 `json:"order"`
	Semester     int                 `json:"semester"`
	CreditPoints int                 `json:"creditPoints"`
	Activities   []StudyPlanActivity `json:"activities"`
}

// StudyPlanActivity is one kind of work in a workload.
type StudyPlanActivity struct {
	ID        int64  `json:"id"`
	ContentID int64  `json:"contentId"`
	Name      string `json:"name"`
	// Volume is in academic hours; nil for assessment forms.
	Volume *float64 `json:"volume"`
	// WorkTypeID: 1 lectures, 3 practice, 4 self-study, 5 exam, 6 credit.
	WorkTypeID int64 `json:"workTypeId"`
}

// Programs returns the user's plans and the active programme.
// GET /api/eduPlanNew/programs
func (s *StudyPlanService) Programs(ctx context.Context) (*StudyPlanPrograms, error) {
	return call[*StudyPlanPrograms](ctx, s.c, get("api/eduPlanNew/programs", nil))
}

// Plan returns the full structure of a plan. specializationID may be nil.
// GET /api/eduPlanNew/study_plan/{plan_id}
func (s *StudyPlanService) Plan(ctx context.Context, planID int64, specializationID *int64) (*StudyPlan, error) {
	return call[*StudyPlan](ctx, s.c, get("api/eduPlanNew/study_plan/"+id(planID), q().set("spec_id", specializationID)))
}

// Module selection rules of StudyPlanNode.ChoiceParameterID.
const (
	// StudyPlanRuleChooseN means choose N disciplines, N one of Rules.
	StudyPlanRuleChooseN = 1
	// StudyPlanRuleAll means every discipline of the module is studied.
	StudyPlanRuleAll = 2
	// StudyPlanRuleFree means any number of disciplines.
	StudyPlanRuleFree = 21
	// StudyPlanRuleCreditsIn means the total credit points must be one of Rules.
	StudyPlanRuleCreditsIn = 41
	// StudyPlanRuleCreditsMax means the total credit points must not exceed max(Rules).
	StudyPlanRuleCreditsMax = 61
)

// StudyPlanChoice is the student's current choice in a plan.
type StudyPlanChoice struct {
	ChosenContents []StudyPlanChosenContent `json:"chosenContents"`
}

// StudyPlanChosenContent is one chosen discipline variant.
type StudyPlanChosenContent struct {
	ModuleID     int64 `json:"moduleId"`
	DisciplineID int64 `json:"disciplineId"`
	// Semester is the start semester of the chosen variant.
	Semester int `json:"semester"`
	// Locked means the choice can no longer be changed.
	Locked bool `json:"locked"`
}

// StudyPlanSignContent is one part of the chosen start-semester variant of a discipline.
type StudyPlanSignContent struct {
	// DcID is StudyPlanContent.ID.
	DcID     int64 `json:"dc_id"`
	Semester int   `json:"semester"`
}

// StudyPlanReplaceable is a discipline that can replace another one.
type StudyPlanReplaceable struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	Abstract string `json:"abstract"`
}

// StudyPlanReplacement is a pending request to replace a discipline.
type StudyPlanReplacement struct {
	// ID is the request ID for [StudyPlanService.CancelReplacement] (not confirmed).
	ID               int64 `json:"id"`
	ModuleID         int64 `json:"moduleId"`
	DisciplineIDFrom int64 `json:"disciplineIdFrom"`
	// Name is the name of the replacement discipline.
	Name string `json:"name"`
}

func studyPlanDisciplinePath(planID, moduleID, disciplineID int64) string {
	return "api/eduPlan/" + id(planID) + "/modules/" + id(moduleID) + "/disciplines/" + id(disciplineID)
}

// Choice returns the disciplines the student has chosen in a plan.
// GET /api/eduPlan/choice/{plan_id}
func (s *StudyPlanService) Choice(ctx context.Context, planID int64) (*StudyPlanChoice, error) {
	return call[*StudyPlanChoice](ctx, s.c, get("api/eduPlan/choice/"+id(planID), nil))
}

// SignDiscipline chooses a discipline of an elective module; contents are the
// parts of the chosen start-semester variant.
// POST /api/eduPlan/{plan_id}/modules/{module_id}/disciplines/{discipline_id}/sign
func (s *StudyPlanService) SignDiscipline(ctx context.Context, planID, moduleID, disciplineID int64, contents []StudyPlanSignContent) error {
	return exec(ctx, s.c, post(studyPlanDisciplinePath(planID, moduleID, disciplineID)+"/sign", contents))
}

// UnsignDiscipline cancels the choice of a discipline; contents are the parts of the chosen variant.
// DELETE /api/eduPlan/{plan_id}/modules/{module_id}/disciplines/{discipline_id}/sign
func (s *StudyPlanService) UnsignDiscipline(ctx context.Context, planID, moduleID, disciplineID int64, contents []StudyPlanSignContent) error {
	return exec(ctx, s.c, del(studyPlanDisciplinePath(planID, moduleID, disciplineID)+"/sign", contents))
}

// Replaceable returns the disciplines that can replace a discipline.
// GET /api/eduPlan/{plan_id}/modules/{module_id}/disciplines/{discipline_id}/replaceable
func (s *StudyPlanService) Replaceable(ctx context.Context, planID, moduleID, disciplineID int64) ([]StudyPlanReplaceable, error) {
	return call[[]StudyPlanReplaceable](ctx, s.c, get(studyPlanDisciplinePath(planID, moduleID, disciplineID)+"/replaceable", nil))
}

// Replacements returns the pending replacement requests of a plan.
// GET /api/eduPlan/{plan_id}/replaceable_discs
func (s *StudyPlanService) Replacements(ctx context.Context, planID int64) ([]StudyPlanReplacement, error) {
	return call[[]StudyPlanReplacement](ctx, s.c, get("api/eduPlan/"+id(planID)+"/replaceable_discs", nil))
}

// RequestReplacement asks to replace a discipline with newDisciplineID
// (a [StudyPlanReplaceable].ID) and returns the request ID.
// POST /api/eduPlan/{plan_id}/modules/{module_id}/disciplines/{discipline_id}/replace/{new_discipline_id}
func (s *StudyPlanService) RequestReplacement(ctx context.Context, planID, moduleID, disciplineID, newDisciplineID int64) (int64, error) {
	return call[int64](ctx, s.c, post(studyPlanDisciplinePath(planID, moduleID, disciplineID)+"/replace/"+id(newDisciplineID), nil))
}

// CancelReplacement withdraws a replacement request.
// DELETE /api/eduPlan/{plan_id}/replace/{replacement_id}
func (s *StudyPlanService) CancelReplacement(ctx context.Context, planID, replacementID int64) error {
	return exec(ctx, s.c, del("api/eduPlan/"+id(planID)+"/replace/"+id(replacementID), nil))
}
