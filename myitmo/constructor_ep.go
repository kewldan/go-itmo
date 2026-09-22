package myitmo

import (
	"context"
	"encoding/json/v2"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// ConstructorEPService is the educational programme constructor (/api/constructor-ep):
// programmes and their curricula, the module bank and the calendar study
// schedule (КУГ). Every route is staff only.
type ConstructorEPService struct{ c *Client }

// Status IDs shared by the plan, KUG, characteristics, discipline and bank
// module statuses. Names come from [ConstructorEPService.Statuses].
const (
	EPStatusDraft     = 1
	EPStatusInWork    = 2
	EPStatusRevision  = 3
	EPStatusArchived  = 4
	EPStatusExpertise = 5
	EPStatusSigned    = 6
	EPStatusSigning   = 7
)

// Module rule types (EPRule.TypeID). Rule types 1, 41 and 61 need a rule value.
const (
	// EPRuleChooseN means choose RuleValue[0] of the children.
	EPRuleChooseN = 1
	// EPRuleByCredits means choose children worth RuleValue[0] credits.
	EPRuleByCredits = 41
)

// EPPartPlan is the part_id of comments on the plan content.
const EPPartPlan = 2

// Curriculum document variants and formats for [ConstructorEPService.PlanDocument].
const (
	EPPlanVariantAbit    = "abit"
	EPPlanVariantDefault = "default"
	EPFormatXLS          = "xls"
	EPFormatXLSX         = "xlsx"
	EPFormatPDF          = "pdf"
)

// Sign task types for EPTasksParams.TaskType.
const (
	EPTaskPlan = "plan"
	EPTaskKUG  = "kug"
)

// EPOpenEnd is the date_end of an open-ended manager position.
var EPOpenEnd = time.Date(9999, time.September, 9, 0, 0, 0, 0, MSK)

// EPRef is a generic reference item (languages, implementers, levels, forms,
// attributes, positions, statuses, countries, rule types). Some embedded
// references carry only one of the fields.
type EPRef struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// EPPerson is a person reference. Some embeddings carry only the name.
type EPPerson struct {
	ISU int64  `json:"isu"`
	FIO string `json:"fio"`
}

// EPRuleValue is the parameter list of a module rule; nil is sent as null.
type EPRuleValue []float64

// MarshalJSON encodes nil as null.
func (r EPRuleValue) MarshalJSON() ([]byte, error) {
	if r == nil {
		return []byte("null"), nil
	}
	return json.Marshal([]float64(r))
}

// EPDirection is a training direction (направление подготовки).
type EPDirection struct {
	ID   int64  `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
}

// EPEnrollmentYear is an enrollment year reference.
type EPEnrollmentYear struct {
	ID   int64 `json:"id"`
	Year int   `json:"year"`
}

// EPStandard is an educational standard. Embedded copies (a programme's or a
// bank module's standard) carry a subset of the fields.
type EPStandard struct {
	ID       int64    `json:"id"`
	Name     string   `json:"name"`
	IsActive FlexBool `json:"is_active"`
	// EducationLevel and EducationForm are the level and form of the standard.
	EducationLevel            EPRef              `json:"education_level"`
	EducationForm             EPRef              `json:"education_form"`
	StandardEducationDuration EPStandardDuration `json:"standard_education_duration"`
	SemestersCount            int                `json:"semesters_count"`
}

// EPStandardDuration is the standard study duration.
type EPStandardDuration struct {
	// Duration is in years.
	Duration float64 `json:"duration"`
}

// EPStandardWithBlocks is a standard with its plan blocks.
type EPStandardWithBlocks struct {
	ID     int64   `json:"id"`
	Name   string  `json:"name"`
	Blocks []EPRef `json:"blocks"`
}

// EPPartnerUniversity is a partner university for academic mobility.
type EPPartnerUniversity struct {
	ID      int64  `json:"id"`
	Name    string `json:"name"`
	Country EPRef  `json:"country"`
}

// EPActivityType is a calendar study schedule activity type.
type EPActivityType struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Color string `json:"color"`
	// Visible is false for hidden types.
	Visible bool `json:"visible"`
}

// EPProgramListRow is a row of the programme list.
type EPProgramListRow struct {
	ID                    int64         `json:"id"`
	Name                  string        `json:"name"`
	AdmissionYear         FlexInt       `json:"admission_year"`
	Directions            []EPDirection `json:"directions"`
	EducationLevel        EPRef         `json:"education_level"`
	Implementer           EPRef         `json:"implementer"`
	Manager               *EPPerson     `json:"manager"`
	EducationPlanStatus   EPRef         `json:"education_plan_status"`
	CharacteristicsStatus EPRef         `json:"characteristics_status"`
	KUGStatus             EPRef         `json:"kug_status"`
}

// EPProgramListPage is a page of the programme list.
type EPProgramListPage struct {
	Rows      []EPProgramListRow `json:"rows"`
	Count     int                `json:"count"`
	CanCreate bool               `json:"can_create"`
	CanSign   bool               `json:"can_sign"`
}

// EPProgramHeader is a programme header: name, permissions and statuses.
type EPProgramHeader struct {
	ID         int64      `json:"id"`
	Name       string     `json:"name"`
	IsAdmin    bool       `json:"is_admin"`
	IsManager  bool       `json:"is_manager"`
	CanEdit    bool       `json:"can_edit"`
	ArchivedAt *time.Time `json:"archived_at"`
	// Status IDs are the EPStatus* constants.
	PlanStatus            EPRef `json:"plan_status"`
	KUGStatus             EPRef `json:"kug_status"`
	CharacteristicsStatus EPRef `json:"characteristics_status"`
}

// EPProgramInfo is the full programme card.
type EPProgramInfo struct {
	ID                    int64            `json:"id"`
	NameRU                string           `json:"name_ru"`
	NameEN                string           `json:"name_en"`
	EnrollmentYear        EPEnrollmentYear `json:"enrollment_year"`
	EducationalStandard   EPStandard       `json:"educational_standard"`
	Implementer           *EPRef           `json:"implementer"`
	Languages             []EPRef          `json:"languages"`
	Attributes            []EPRef          `json:"attributes"`
	EducationalDirections []EPDirection    `json:"educational_directions"`
	// ActualEducationDuration is in years.
	ActualEducationDuration float64 `json:"actual_education_duration"`
	Note                    string  `json:"note"`
	// AbitExport publishes the programme on the admissions site.
	AbitExport bool `json:"abit_export"`
	// ExistsInISU is true once the programme was sent to ISU.
	ExistsInISU    bool               `json:"exists_in_isu"`
	AnnotationRU   string             `json:"annotation_ru"`
	AnnotationEN   string             `json:"annotation_en"`
	KCP            []EPKCP            `json:"kcp"`
	Managers       []EPManager        `json:"managers"`
	Companies      []EPProgramCompany `json:"companies"`
	AdditionalInfo *EPMobility        `json:"additional_info"`
	// PartnerUniversity is read by another code path with the same shape as
	// AdditionalInfo; one of the two names is legacy.
	PartnerUniversity *EPMobility `json:"partner_university,omitzero"`
}

// EPMobility is the academic mobility of a programme.
type EPMobility struct {
	PartnerUniversity     *EPPartnerUniversity `json:"partner_university"`
	AcademicMobilityStart *Date                `json:"academic_mobility_start"`
	AcademicMobilityEnd   *Date                `json:"academic_mobility_end"`
}

// EPKCP is the admission control figures (КЦП) of one direction. It is both
// returned in EPProgramInfo and sent to [ConstructorEPService.SetKCP].
type EPKCP struct {
	KCPID       int64  `json:"kcp_id"`
	DirectionID int64  `json:"direction_id"`
	Code        string `json:"code"`
	Name        string `json:"name"`
	// BudgetForm, ContractForm and OtherForm are seat counts; nil clears them.
	BudgetForm         *int `json:"budget_form"`
	ContractForm       *int `json:"contract_form"`
	OtherForm          *int `json:"other_form"`
	AbitExport         bool `json:"abit_export"`
	MilitaryDepartment bool `json:"military_department"`
}

// EPManager is a programme administrator.
type EPManager struct {
	ISU       int64               `json:"isu"`
	FIO       string              `json:"fio"`
	Positions []EPManagerPosition `json:"positions"`
}

// EPManagerPosition is a position held by a manager. A DateEnd of 9999-09-09 means open-ended.
type EPManagerPosition struct {
	PositionID   int64  `json:"position_id"`
	PositionName string `json:"position_name"`
	DateStart    Date   `json:"date_start"`
	DateEnd      Date   `json:"date_end"`
}

// EPManagerForm adds a manager in [ConstructorEPService.AddManagers].
type EPManagerForm struct {
	ISU       int64                   `json:"isu"`
	Positions []EPManagerPositionForm `json:"positions"`
}

// EPManagerPositionForm is a position of a manager. Dates are full
// timestamps; use [EPOpenEnd] as DateEnd for an open-ended position.
// Position 3 is exempt from the one-holder-per-position rule.
type EPManagerPositionForm struct {
	PositionID int64     `json:"position_id"`
	DateStart  time.Time `json:"date_start"`
	DateEnd    time.Time `json:"date_end"`
}

// EPCompany is a company from the registry.
type EPCompany struct {
	INN              string `json:"inn"`
	NameShortWithOPF string `json:"name_short_with_opf"`
	NameFullWithOPF  string `json:"name_full_with_opf"`
}

// EPProgramCompany is a company attached to a programme under an IT-benefit type.
type EPProgramCompany struct {
	// ID is the link ID used by the company file routes.
	ID               int64  `json:"id"`
	TypeID           int64  `json:"type_id"`
	INN              string `json:"inn"`
	NameShortWithOPF string `json:"name_short_with_opf"`
	NameFullWithOPF  string `json:"name_full_with_opf"`
	FileAttached     bool   `json:"file_attached"`
	FileName         string `json:"file_name"`
	// FileType is the file extension.
	FileType string `json:"file_type"`
}

// EPITType is an IT-benefit (ИТ-льгота) company category.
type EPITType struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Order int    `json:"order"`
	// FileAllowed allows a supporting file for companies of this type.
	FileAllowed bool `json:"file_allowed"`
}

// EPComment is an expert comment on a programme part.
type EPComment struct {
	ID int64 `json:"id"`
	// PartID is [EPPartPlan] for the plan content.
	PartID  int64  `json:"part_id"`
	Comment string `json:"comment"`
	FIO     string `json:"fio"`
	UserISU int64  `json:"user_isu"`
}

// EPPlan is the curriculum of a programme.
type EPPlan struct {
	Name           string  `json:"name"`
	EnrollmentYear FlexInt `json:"enrollment_year"`
	// FullCapacity is the number of credits required.
	FullCapacity      *float64     `json:"full_capacity"`
	FilledMinCapacity float64      `json:"filled_min_capacity"`
	FilledMaxCapacity float64      `json:"filled_max_capacity"`
	Plan              []EPPlanNode `json:"plan"`
}

// Plan tree node types.
const (
	EPNodeModule     = "module"
	EPNodeDiscipline = "discipline"
)

// EPPlanNode is a node of a curriculum or bank module tree: a module or a discipline.
type EPPlanNode struct {
	ID int64 `json:"id"`
	// Type is [EPNodeModule] or [EPNodeDiscipline].
	Type     string `json:"type"`
	Name     string `json:"name"`
	NameRU   string `json:"name_ru"`
	NameEN   string `json:"name_en"`
	ParentID *int64 `json:"parent_id"`
	// ModuleID is the owning module of a discipline.
	ModuleID int64 `json:"module_id"`
	// IsRoot marks a block (root module).
	IsRoot bool `json:"is_root"`
	// IsBank marks a module from the module bank.
	IsBank  bool    `json:"is_bank"`
	IsShop  bool    `json:"is_shop"`
	BlockID int64   `json:"block_id"`
	Rule    *EPRule `json:"rule"`
	// RuleTypes are the rule types allowed in a root module.
	RuleTypes []int64 `json:"rule_types"`
	// ChildrenTypes are the discipline program types allowed in a root module.
	ChildrenTypes        []int64      `json:"children_types"`
	SameCapacityChildren bool         `json:"same_capacity_children"`
	Capacity             float64      `json:"capacity"`
	MinCapacity          float64      `json:"min_capacity"`
	MaxCapacity          float64      `json:"max_capacity"`
	TargetCapacity       *float64     `json:"target_capacity"`
	Children             []EPPlanNode `json:"children"`
	// Discipline fields.
	Status         *EPRef         `json:"status"`
	Implementer    *EPImplementer `json:"implementer"`
	TypeID         int64          `json:"type_id"`
	Semesters      []int          `json:"semesters"`
	SemestersCount int            `json:"semesters_count"`
	LanguageID     int64          `json:"language_id"`
	LanguageCode   string         `json:"language_code"`
}

// EPRule is a module rule.
type EPRule struct {
	// TypeID is one of the EPRule* constants or another rule type.
	TypeID    int64     `json:"type_id"`
	TypeName  string    `json:"type_name"`
	RuleValue []float64 `json:"rule_value"`
}

// EPImplementer is the implementing department of a discipline.
type EPImplementer struct {
	Name      string `json:"name"`
	ShortName string `json:"short_name"`
}

// EPModuleForm creates or edits a module of a plan or bank module tree.
type EPModuleForm struct {
	NameRU string `json:"name_ru"`
	// NameEN must be Latin.
	NameEN     string      `json:"name_en"`
	RuleTypeID int64       `json:"rule_type_id"`
	RuleValue  EPRuleValue `json:"rule_value"`
	ParentID   int64       `json:"parent_id"`
	BlockID    int64       `json:"block_id"`
	StandardID int64       `json:"standard_id"`
}

// EPDisciplineSearchItem is a discipline found by [ConstructorEPService.Disciplines].
type EPDisciplineSearchItem struct {
	ID              int64   `json:"id"`
	Name            string  `json:"name"`
	AnnotationRU    string  `json:"annotation_ru"`
	Capacity        float64 `json:"capacity"`
	ImplementerName string  `json:"implementer_name"`
	SemestersCount  int     `json:"semesters_count"`
	StatusID        int64   `json:"status_id"`
	StatusName      string  `json:"status_name"`
	EducationLevels []EPRef `json:"education_levels"`
}

// EPSignTask is a signing task for a plan or KUG document. The files come
// from the sign service (/api/sign/tasks/{task_id}/files/...).
type EPSignTask struct {
	TaskID int64  `json:"task_id"`
	Name   string `json:"name"`
	// Type is "pdf".
	Type string `json:"type"`
	// Status is "SIGNED" and others.
	Status     string        `json:"status"`
	Signatures []EPSignature `json:"signatures"`
}

// EPSignature is a signature slot of a sign task.
type EPSignature struct {
	SignatureID int64 `json:"signature_id"`
}

// EPTaskList is the result of [ConstructorEPService.Tasks]; other keys are unknown.
type EPTaskList struct {
	Rows []EPSignTask `json:"rows"`
}

// EPSimpleProgram is a programme of the lightweight list.
type EPSimpleProgram struct {
	ID             int64  `json:"id"`
	Name           string `json:"name"`
	EnrollmentYear int    `json:"enrollment_year"`
}

// EPProgramForm creates or edits a programme's main information.
type EPProgramForm struct {
	NameRU string `json:"name_ru"`
	// NameEN must be Latin.
	NameEN                string  `json:"name_en"`
	EnrollmentYearID      int64   `json:"enrollment_year_id"`
	EducationalStandardID int64   `json:"educational_standard_id"`
	ImplementerID         int64   `json:"implementer_id"`
	Languages             []int64 `json:"languages"`
	// ActualEducationDuration is in years.
	ActualEducationDuration float64 `json:"actual_education_duration"`
	DirectionsID            []int64 `json:"directions_id"`
	Attributes              []int64 `json:"attributes"`
	Note                    *string `json:"note"`
	AbitExport              bool    `json:"abit_export"`
}

// EPMobilityForm sets the academic mobility; all nil clears it.
type EPMobilityForm struct {
	PartnerUniversityID   *int64 `json:"partner_university_id"`
	AcademicMobilityStart *Date  `json:"academic_mobility_start"`
	AcademicMobilityEnd   *Date  `json:"academic_mobility_end"`
}

// EPProgramListParams filters [ConstructorEPService.List]. Slices repeat the query key.
type EPProgramListParams struct {
	Query                    string
	PlanStatusIDs            []int64
	CharacteristicsStatusIDs []int64
	KUGStatusIDs             []int64
	EducationLevelIDs        []int64
	ImplementerIDs           []int64
	DirectionIDs             []int64
	// EnrollmentYears are enrollment year IDs (EPEnrollmentYear.ID).
	EnrollmentYears  []int64
	EducationFormIDs []int64
	// StandardID and KUGFree are used when picking programmes for a KUG template.
	StandardID int64
	KUGFree    bool
	// Archived lists archived programmes.
	Archived bool
	// Limit is the page size (typically 15); 0 omits it.
	Limit  int
	Offset int
}

// EPSimpleListParams filters [ConstructorEPService.SimpleList].
type EPSimpleListParams struct {
	Query            string
	EnrollmentYears  []int64
	ImplementerID    int64
	EducationLevelID int64
	Archived         bool
}

// EPTasksParams filters [ConstructorEPService.Tasks]. Zero values are omitted.
type EPTasksParams struct {
	Query            string
	EducationLevelID int64
	ImplementerID    int64
	EnrollmentYear   int64
	// TaskType is [EPTaskPlan] or [EPTaskKUG].
	TaskType string
}

// EPDisciplineSearchParams filters [ConstructorEPService.Disciplines].
type EPDisciplineSearchParams struct {
	Query            string
	EducationLevelID int64
	// StatusIDs are sent comma-separated (e.g. 2,3,5,6,7).
	StatusIDs []int64
	// ProgramTypeIDs are the allowed discipline program types (repeated key).
	ProgramTypeIDs []int64
}

func epSetID(v query, key string, n int64) query {
	if n != 0 {
		v.set(key, n)
	}
	return v
}

func epFlag(b bool) int {
	if b {
		return 1
	}
	return 0
}

func epNullable(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func epPath(parts ...string) string {
	return "api/constructor-ep/" + strings.Join(parts, "/")
}

func epProgram(programID int64, parts ...string) string {
	return epPath(append([]string{"programs", id(programID)}, parts...)...)
}

// List returns a page of educational programmes.
// Staff only.
// GET /api/constructor-ep/programs/list
func (s *ConstructorEPService) List(ctx context.Context, p EPProgramListParams) (*EPProgramListPage, error) {
	v := q().set("query", p.Query).
		set("plan_status_id", p.PlanStatusIDs).
		set("characteristics_status_id", p.CharacteristicsStatusIDs).
		set("kug_status_id", p.KUGStatusIDs).
		set("education_level_id", p.EducationLevelIDs).
		set("implementer_id", p.ImplementerIDs).
		set("direction_id", p.DirectionIDs).
		set("enrollment_year", p.EnrollmentYears).
		set("education_form_id", p.EducationFormIDs).
		set("archive", epFlag(p.Archived)).
		set("offset", p.Offset)
	epSetID(v, "standard_id", p.StandardID)
	if p.KUGFree {
		v.set("kug_free", true)
	}
	if p.Limit > 0 {
		v.set("limit", p.Limit)
	}
	return call[*EPProgramListPage](ctx, s.c, get(epPath("programs", "list"), v))
}

// SimpleList returns a lightweight programme list (for bank module access).
// Staff only.
// GET /api/constructor-ep/programs/simple_list
func (s *ConstructorEPService) SimpleList(ctx context.Context, p EPSimpleListParams) ([]EPSimpleProgram, error) {
	v := q().set("enrollment_year", p.EnrollmentYears).set("query", p.Query)
	epSetID(v, "implementer_id", p.ImplementerID)
	epSetID(v, "education_level_id", p.EducationLevelID)
	v.set("archive", epFlag(p.Archived))
	return call[[]EPSimpleProgram](ctx, s.c, get(epPath("programs", "simple_list"), v))
}

// Tasks returns the plan or KUG signing tasks of the current signer.
// Staff only.
// GET /api/constructor-ep/programs/tasks_list
func (s *ConstructorEPService) Tasks(ctx context.Context, p EPTasksParams) (*EPTaskList, error) {
	v := q().set("query", p.Query)
	epSetID(v, "education_level_id", p.EducationLevelID)
	epSetID(v, "implementer_id", p.ImplementerID)
	epSetID(v, "enrollment_year", p.EnrollmentYear)
	v.set("task_type", p.TaskType)
	return call[*EPTaskList](ctx, s.c, get(epPath("programs", "tasks_list"), v))
}

// Create creates a programme and returns its ID.
// Staff only.
// POST /api/constructor-ep/programs/create
func (s *ConstructorEPService) Create(ctx context.Context, form EPProgramForm) (int64, error) {
	return call[int64](ctx, s.c, post(epPath("programs", "create"), form))
}

// Update edits a programme's main information.
// Staff only.
// PATCH /api/constructor-ep/programs/{program_id}
func (s *ConstructorEPService) Update(ctx context.Context, programID int64, form EPProgramForm) error {
	return exec(ctx, s.c, patch(epProgram(programID), form))
}

// Info returns the full programme card.
// Staff only.
// GET /api/constructor-ep/programs/{program_id}/info
func (s *ConstructorEPService) Info(ctx context.Context, programID int64) (*EPProgramInfo, error) {
	return call[*EPProgramInfo](ctx, s.c, get(epProgram(programID, "info"), nil))
}

// Status returns the programme header: name, permissions and statuses.
// Staff only.
// GET /api/constructor-ep/programs/{program_id}/status
func (s *ConstructorEPService) Status(ctx context.Context, programID int64) (*EPProgramHeader, error) {
	return call[*EPProgramHeader](ctx, s.c, get(epProgram(programID, "status"), nil))
}

// SetAnnotation sets the programme annotation; an empty string is sent as null.
// Staff only.
// PATCH /api/constructor-ep/programs/{program_id}/annotation
func (s *ConstructorEPService) SetAnnotation(ctx context.Context, programID int64, ru, en string) error {
	body := struct {
		AnnotationRU *string `json:"annotation_ru"`
		AnnotationEN *string `json:"annotation_en"`
	}{epNullable(ru), epNullable(en)}
	return exec(ctx, s.c, patch(epProgram(programID, "annotation"), body))
}

// SetNote sets the programme note; an empty string deletes it.
// Staff only.
// PATCH /api/constructor-ep/programs/{program_id}/note
func (s *ConstructorEPService) SetNote(ctx context.Context, programID int64, note string) error {
	body := struct {
		Note *string `json:"note"`
	}{epNullable(note)}
	return exec(ctx, s.c, patch(epProgram(programID, "note"), body))
}

// SetMobility sets or clears the academic mobility partner and period.
// Staff only.
// PATCH /api/constructor-ep/programs/{program_id}/mobility
func (s *ConstructorEPService) SetMobility(ctx context.Context, programID int64, form EPMobilityForm) error {
	return exec(ctx, s.c, patch(epProgram(programID, "mobility"), form))
}

// SetKCP saves the admission control figures per direction.
// Staff only.
// PATCH /api/constructor-ep/programs/{program_id}/kcp
func (s *ConstructorEPService) SetKCP(ctx context.Context, programID int64, rows []EPKCP) error {
	return exec(ctx, s.c, patch(epProgram(programID, "kcp"), rows))
}

// AddManagers adds programme administrators.
// Staff only.
// POST /api/constructor-ep/programs/{program_id}/managers
func (s *ConstructorEPService) AddManagers(ctx context.Context, programID int64, managers []EPManagerForm) error {
	return exec(ctx, s.c, post(epProgram(programID, "managers"), managers))
}

// UpdateManager replaces the positions of a manager.
// Staff only.
// PATCH /api/constructor-ep/programs/{program_id}/managers/{isu}
func (s *ConstructorEPService) UpdateManager(ctx context.Context, programID, isu int64, positions []EPManagerPositionForm) error {
	return exec(ctx, s.c, patch(epProgram(programID, "managers", id(isu)), positions))
}

// DeleteManager removes a manager from the programme.
// Staff only.
// DELETE /api/constructor-ep/programs/{program_id}/managers/{isu}
func (s *ConstructorEPService) DeleteManager(ctx context.Context, programID, isu int64) error {
	return exec(ctx, s.c, del(epProgram(programID, "managers", id(isu)), nil))
}

// Archive archives the programme.
// Staff only.
// POST /api/constructor-ep/programs/{program_id}/archive
func (s *ConstructorEPService) Archive(ctx context.Context, programID int64) error {
	return exec(ctx, s.c, post(epProgram(programID, "archive"), nil))
}

// SendToISU sends the programme to ISU; later main-info changes sync there.
// Staff only.
// POST /api/constructor-ep/programs/{program_id}/to_isu
func (s *ConstructorEPService) SendToISU(ctx context.Context, programID int64) error {
	return exec(ctx, s.c, post(epProgram(programID, "to_isu"), nil))
}

// AddCompany attaches a registry company to the programme under an IT-benefit type (EPITType.ID).
// Staff only.
// POST /api/constructor-ep/programs/{program_id}/companies
func (s *ConstructorEPService) AddCompany(ctx context.Context, programID int64, company EPCompany, typeID int64) error {
	body := struct {
		EPCompany
		TypeID int64 `json:"type_id"`
	}{company, typeID}
	return exec(ctx, s.c, post(epProgram(programID, "companies"), body))
}

// DeleteCompany detaches a company (EPProgramCompany.ID) from the programme.
// Staff only.
// DELETE /api/constructor-ep/programs/{program_id}/companies/{company_id}
func (s *ConstructorEPService) DeleteCompany(ctx context.Context, programID, companyID int64) error {
	return exec(ctx, s.c, del(epProgram(programID, "companies", id(companyID)), nil))
}

// Comments returns the expert comments on the programme parts.
// Staff only.
// GET /api/constructor-ep/programs/{program_id}/comments
func (s *ConstructorEPService) Comments(ctx context.Context, programID int64) ([]EPComment, error) {
	return call[[]EPComment](ctx, s.c, get(epProgram(programID, "comments"), nil))
}

// AddComment adds an expert comment to a programme part ([EPPartPlan] for the plan).
// Staff only.
// POST /api/constructor-ep/programs/{program_id}/comments
func (s *ConstructorEPService) AddComment(ctx context.Context, programID int64, comment string, partID int64) error {
	body := struct {
		Comment string `json:"comment"`
		PartID  int64  `json:"part_id"`
	}{comment, partID}
	return exec(ctx, s.c, post(epProgram(programID, "comments"), body))
}

// UpdateComment edits an expert comment.
// Staff only.
// PATCH /api/constructor-ep/programs/{program_id}/comments/{comment_id}
func (s *ConstructorEPService) UpdateComment(ctx context.Context, programID, commentID int64, comment string) error {
	body := struct {
		Comment string `json:"comment"`
	}{comment}
	return exec(ctx, s.c, patch(epProgram(programID, "comments", id(commentID)), body))
}

// DeleteComment deletes an expert comment.
// Staff only.
// DELETE /api/constructor-ep/programs/{program_id}/comments/{comment_id}
func (s *ConstructorEPService) DeleteComment(ctx context.Context, programID, commentID int64) error {
	return exec(ctx, s.c, del(epProgram(programID, "comments", id(commentID)), nil))
}

// Plan returns the curriculum tree with credit totals.
// Staff only.
// GET /api/constructor-ep/programs/{program_id}/plan
func (s *ConstructorEPService) Plan(ctx context.Context, programID int64) (*EPPlan, error) {
	return call[*EPPlan](ctx, s.c, get(epProgram(programID, "plan"), nil))
}

// PlanPracticeCredits returns the practice credits per semester, indexed by
// the 1-based semester number. Whether it is an object or an array is unknown.
// Staff only.
// GET /api/constructor-ep/programs/{program_id}/plan/capacity_sum/by_semesters
func (s *ConstructorEPService) PlanPracticeCredits(ctx context.Context, programID int64) (RawJSON, error) {
	return call[RawJSON](ctx, s.c, get(epProgram(programID, "plan", "capacity_sum", "by_semesters"), nil))
}

// CreatePlanModule creates a module in the plan tree and returns its ID.
// Staff only.
// POST /api/constructor-ep/programs/{program_id}/plan/module
func (s *ConstructorEPService) CreatePlanModule(ctx context.Context, programID int64, form EPModuleForm) (int64, error) {
	return call[int64](ctx, s.c, post(epProgram(programID, "plan", "module"), form))
}

// UpdatePlanModule edits a plan module and returns its ID.
// Staff only.
// PATCH /api/constructor-ep/programs/{program_id}/plan/module/{module_id}
func (s *ConstructorEPService) UpdatePlanModule(ctx context.Context, programID, moduleID int64, form EPModuleForm) (int64, error) {
	return call[int64](ctx, s.c, patch(epProgram(programID, "plan", "module", id(moduleID)), form))
}

// DeletePlanModule deletes a module and its subtree from the plan.
// Staff only.
// DELETE /api/constructor-ep/programs/{program_id}/plan/module/{module_id}
func (s *ConstructorEPService) DeletePlanModule(ctx context.Context, programID, moduleID int64) error {
	return exec(ctx, s.c, del(epProgram(programID, "plan", "module", id(moduleID)), nil))
}

// AddPlanDisciplines adds existing disciplines to a plan module.
// Staff only.
// POST /api/constructor-ep/programs/{program_id}/plan/module/{module_id}/discipline
func (s *ConstructorEPService) AddPlanDisciplines(ctx context.Context, programID, moduleID int64, disciplineIDs []int64) error {
	return exec(ctx, s.c, post(epProgram(programID, "plan", "module", id(moduleID), "discipline"), disciplineIDs))
}

// AddPlanBankModules inserts module bank modules into a plan module (bank roots only).
// Staff only.
// POST /api/constructor-ep/programs/{program_id}/plan/module/{module_id}/bank_module
func (s *ConstructorEPService) AddPlanBankModules(ctx context.Context, programID, moduleID int64, bankModuleIDs []int64) error {
	return exec(ctx, s.c, post(epProgram(programID, "plan", "module", id(moduleID), "bank_module"), bankModuleIDs))
}

// SetPlanDisciplineSemesters sets the semesters of a discipline in the plan.
// For a multi-semester discipline pass only the start semester.
// Staff only.
// PATCH /api/constructor-ep/programs/{program_id}/plan/module/{module_id}/disciplines/{discipline_id}
func (s *ConstructorEPService) SetPlanDisciplineSemesters(ctx context.Context, programID, moduleID, disciplineID int64, semesters []int) error {
	return exec(ctx, s.c, patch(epProgram(programID, "plan", "module", id(moduleID), "disciplines", id(disciplineID)), semesters))
}

// DeletePlanDiscipline removes a discipline from a plan module.
// Staff only.
// DELETE /api/constructor-ep/programs/{program_id}/plan/module/{module_id}/disciplines/{discipline_id}
func (s *ConstructorEPService) DeletePlanDiscipline(ctx context.Context, programID, moduleID, disciplineID int64) error {
	return exec(ctx, s.c, del(epProgram(programID, "plan", "module", id(moduleID), "disciplines", id(disciplineID)), nil))
}

// PlanToExpertise sends the plan to expertise.
// Staff only.
// POST /api/constructor-ep/programs/{program_id}/plan/to_expertise
func (s *ConstructorEPService) PlanToExpertise(ctx context.Context, programID int64) error {
	return exec(ctx, s.c, post(epProgram(programID, "plan", "to_expertise"), nil))
}

// PlanToRevision returns the plan to revision.
// Staff only.
// POST /api/constructor-ep/programs/{program_id}/plan/to_revision
func (s *ConstructorEPService) PlanToRevision(ctx context.Context, programID int64) error {
	return exec(ctx, s.c, post(epProgram(programID, "plan", "to_revision"), nil))
}

// PlanToSigning sends the plan to signing.
// Staff only.
// POST /api/constructor-ep/programs/{program_id}/plan/to_signing
func (s *ConstructorEPService) PlanToSigning(ctx context.Context, programID int64) error {
	return exec(ctx, s.c, post(epProgram(programID, "plan", "to_signing"), nil))
}

// SignedPlan downloads the signed plan PDF (plan status [EPStatusSigned]).
// Staff only.
// GET /api/constructor-ep/programs/{program_id}/signed
func (s *ConstructorEPService) SignedPlan(ctx context.Context, programID int64) (*File, error) {
	return download(ctx, s.c, get(epProgram(programID, "signed"), nil))
}

// PlanDocument exports the curriculum. variant is [EPPlanVariantAbit] (short,
// for applicants) or [EPPlanVariantDefault]; format is [EPFormatXLS] or [EPFormatPDF].
// Staff only.
// GET /api/constructor-ep/programs/{program_id}/plan/{variant}/{format}
func (s *ConstructorEPService) PlanDocument(ctx context.Context, programID int64, variant, format string) (*File, error) {
	return download(ctx, s.c, get(epProgram(programID, "plan", id(variant), id(format)), nil))
}

// SearchCompanies searches the company registry by name or INN (at least 3 characters).
// Staff only.
// GET /api/constructor-ep/companies
func (s *ConstructorEPService) SearchCompanies(ctx context.Context, query string) ([]EPCompany, error) {
	return call[[]EPCompany](ctx, s.c, get(epPath("companies"), q().set("query", strings.TrimSpace(query))))
}

// CompanyFile downloads the supporting file of a programme company (EPProgramCompany.ID).
// Staff only.
// GET /api/constructor-ep/companies/{company_id}/file
func (s *ConstructorEPService) CompanyFile(ctx context.Context, companyID int64) (*File, error) {
	return download(ctx, s.c, get(epPath("companies", id(companyID), "file"), nil))
}

// UploadCompanyFile uploads or replaces the supporting file of a programme
// company; the part is always sent as field "file".
// Staff only.
// POST /api/constructor-ep/companies/{company_id}/file
func (s *ConstructorEPService) UploadCompanyFile(ctx context.Context, companyID int64, file Upload) error {
	file.Field = "file"
	return exec(ctx, s.c, multipart(http.MethodPost, epPath("companies", id(companyID), "file"), nil, file))
}

// DeleteCompanyFile removes the supporting file of a programme company.
// Staff only.
// DELETE /api/constructor-ep/companies/{company_id}/file
func (s *ConstructorEPService) DeleteCompanyFile(ctx context.Context, companyID int64) error {
	return exec(ctx, s.c, del(epPath("companies", id(companyID), "file"), nil))
}

// ITTypes returns the IT-benefit company categories.
// Staff only.
// GET /api/constructor-ep/it_types
func (s *ConstructorEPService) ITTypes(ctx context.Context) ([]EPITType, error) {
	return call[[]EPITType](ctx, s.c, get(epPath("it_types"), nil))
}

// Disciplines searches disciplines to add to a plan or bank module.
// Staff only.
// GET /api/constructor-ep/disciplines/list
func (s *ConstructorEPService) Disciplines(ctx context.Context, p EPDisciplineSearchParams) ([]EPDisciplineSearchItem, error) {
	v := q().set("query", p.Query)
	epSetID(v, "education_level_id", p.EducationLevelID)
	if len(p.StatusIDs) > 0 {
		ids := make([]string, len(p.StatusIDs))
		for i, n := range p.StatusIDs {
			ids[i] = strconv.FormatInt(n, 10)
		}
		v.set("status", strings.Join(ids, ","))
	}
	v.set("program_type_id", p.ProgramTypeIDs)
	res, err := call[struct {
		Disciplines []EPDisciplineSearchItem `json:"disciplines"`
	}](ctx, s.c, get(epPath("disciplines", "list"), v))
	return res.Disciplines, err
}

// SearchPeople finds people by name or ISU number.
// Staff only.
// GET /api/constructor-ep/people/search/{query}
func (s *ConstructorEPService) SearchPeople(ctx context.Context, query string) ([]EPPerson, error) {
	return call[[]EPPerson](ctx, s.c, get(epPath("people", "search", id(query)), nil))
}

// IsAdmin reports whether the current user is a constructor administrator.
// Staff only.
// GET /api/constructor-ep/users/is_admin
func (s *ConstructorEPService) IsAdmin(ctx context.Context) (bool, error) {
	return call[bool](ctx, s.c, get(epPath("users", "is_admin"), nil))
}

func (s *ConstructorEPService) refs(ctx context.Context, name string) ([]EPRef, error) {
	return call[[]EPRef](ctx, s.c, get(epPath("references", name), nil))
}

// Levels returns the education levels.
// Staff only.
// GET /api/constructor-ep/references/levels
func (s *ConstructorEPService) Levels(ctx context.Context) ([]EPRef, error) {
	return s.refs(ctx, "levels")
}

// Languages returns the implementation languages.
// Staff only.
// GET /api/constructor-ep/references/languages
func (s *ConstructorEPService) Languages(ctx context.Context) ([]EPRef, error) {
	return s.refs(ctx, "languages")
}

// Implementers returns the implementing departments.
// Staff only.
// GET /api/constructor-ep/references/implementers
func (s *ConstructorEPService) Implementers(ctx context.Context) ([]EPRef, error) {
	return s.refs(ctx, "implementers")
}

// Forms returns the education forms.
// Staff only.
// GET /api/constructor-ep/references/forms
func (s *ConstructorEPService) Forms(ctx context.Context) ([]EPRef, error) {
	return s.refs(ctx, "forms")
}

// Standards returns the educational standards.
// Staff only.
// GET /api/constructor-ep/references/standards
func (s *ConstructorEPService) Standards(ctx context.Context) ([]EPStandard, error) {
	return call[[]EPStandard](ctx, s.c, get(epPath("references", "standards"), nil))
}

// StandardBlocks returns the standards with their plan blocks.
// Staff only.
// GET /api/constructor-ep/references/standards/blocks
func (s *ConstructorEPService) StandardBlocks(ctx context.Context) ([]EPStandardWithBlocks, error) {
	return call[[]EPStandardWithBlocks](ctx, s.c, get(epPath("references", "standards", "blocks"), nil))
}

// Directions returns the training directions, of one education level when
// educationLevelID is not 0.
// Staff only.
// GET /api/constructor-ep/references/directions
func (s *ConstructorEPService) Directions(ctx context.Context, educationLevelID int64) ([]EPDirection, error) {
	return call[[]EPDirection](ctx, s.c, get(epPath("references", "directions"), epSetID(q(), "education_level_id", educationLevelID)))
}

// PartnerUniversities returns the partner universities for academic mobility.
// Staff only.
// GET /api/constructor-ep/references/partner_universities
func (s *ConstructorEPService) PartnerUniversities(ctx context.Context) ([]EPPartnerUniversity, error) {
	return call[[]EPPartnerUniversity](ctx, s.c, get(epPath("references", "partner_universities"), nil))
}

// Statuses returns the status names (see the EPStatus* constants).
// Staff only.
// GET /api/constructor-ep/references/statuses
func (s *ConstructorEPService) Statuses(ctx context.Context) ([]EPRef, error) {
	return s.refs(ctx, "statuses")
}

// Countries returns the countries.
// Staff only.
// GET /api/constructor-ep/references/countries
func (s *ConstructorEPService) Countries(ctx context.Context) ([]EPRef, error) {
	return s.refs(ctx, "countries")
}

// Attributes returns the programme attributes.
// Staff only.
// GET /api/constructor-ep/references/attributes
func (s *ConstructorEPService) Attributes(ctx context.Context) ([]EPRef, error) {
	return s.refs(ctx, "attributes")
}

// EnrollmentYears returns the enrollment years.
// Staff only.
// GET /api/constructor-ep/references/enrollment_years
func (s *ConstructorEPService) EnrollmentYears(ctx context.Context) ([]EPEnrollmentYear, error) {
	return call[[]EPEnrollmentYear](ctx, s.c, get(epPath("references", "enrollment_years"), nil))
}

// Positions returns the manager positions.
// Staff only.
// GET /api/constructor-ep/references/positions
func (s *ConstructorEPService) Positions(ctx context.Context) ([]EPRef, error) {
	return s.refs(ctx, "positions")
}

// RuleTypes returns the module rule types (see the EPRule* constants).
// Staff only.
// GET /api/constructor-ep/references/rule_types
func (s *ConstructorEPService) RuleTypes(ctx context.Context) ([]EPRef, error) {
	return s.refs(ctx, "rule_types")
}

// ActivityTypes returns the calendar study schedule activity types.
// Staff only.
// GET /api/constructor-ep/references/activity_types
func (s *ConstructorEPService) ActivityTypes(ctx context.Context) ([]EPActivityType, error) {
	return call[[]EPActivityType](ctx, s.c, get(epPath("references", "activity_types"), nil))
}
