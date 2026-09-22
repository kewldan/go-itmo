package myitmo

import "context"

// ConstructorPracticeForm is the main info of a practice programme (РПП).
type ConstructorPracticeForm struct {
	NameRU        string `json:"name_ru"`
	NameEN        string `json:"name_en"`
	FinAttrID     int64  `json:"fin_attr_id"`
	Abbreviation  string `json:"abbreviation"`
	LanguageID    int64  `json:"language_id"`
	FormatID      int64  `json:"format_id"`
	ImplementerID int64  `json:"implementer_id"`
	// EducationLevels usually holds a single id.
	EducationLevels []int64 `json:"education_levels"`
}

// ConstructorPracticeCreate creates a practice programme.
type ConstructorPracticeCreate struct {
	ConstructorPracticeForm
	// Contents lists the parts with their credits (1..30 each).
	Contents []ConstructorPartCapacity `json:"contents"`
}

// ConstructorPracticeInfo is the main page of a practice programme.
type ConstructorPracticeInfo struct {
	ID              int64               `json:"id"`
	NameRU          string              `json:"name_ru"`
	NameEN          string              `json:"name_en"`
	Abbreviation    string              `json:"abbreviation"`
	FinAttrID       int64               `json:"fin_attr_id"`
	FinAttrName     string              `json:"fin_attr_name"`
	LanguageID      int64               `json:"language_id"`
	LanguageName    string              `json:"language_name"`
	ImplementerID   int64               `json:"implementer_id"`
	ImplementerName string              `json:"implementer_name"`
	FormatID        int64               `json:"format_id"`
	EducationLevels []ConstructorIDName `json:"education_levels"`
	// FormID, TypeID, PassTypeID and PracticeFormatID are nil until set by
	// [ConstructorService.UpdatePracticeAbout].
	FormID           *int64 `json:"form_id"`
	TypeID           *int64 `json:"type_id"`
	PassTypeID       *int64 `json:"pass_type_id"`
	PracticeFormatID *int64 `json:"practice_format_id"`
	// Description is the fixed main provisions (HTML).
	Description           string              `json:"description"`
	AdditionalDescription string              `json:"additional_description"`
	OutTags               []ConstructorOutTag `json:"out_tags"`
	// Contents use work type 128 for contact hours.
	Contents   []ConstructorContent `json:"contents"`
	Developers []ConstructorPerson  `json:"developers"`
	Authors    []ConstructorPerson  `json:"authors"`
}

// ConstructorPracticePart sets the credits and form of control of one practice part.
type ConstructorPracticePart struct {
	Capacity                int                      `json:"capacity"` // up to 30
	ContentOrder            int                      `json:"content_order"`
	IntermediateAssessments []ConstructorWorkTypeRef `json:"intermediate_assessments"`
}

// ConstructorPracticeAbout is the kind, type, pass method and format of a practice.
type ConstructorPracticeAbout struct {
	// FormID is a [ConstructorService.PracticeForms] id.
	FormID int64 `json:"form_id"`
	// TypeID is a [ConstructorService.PracticeTypes] id.
	TypeID int64 `json:"type_id"`
	// PassTypeID is a [ConstructorService.PracticePassTypes] id.
	PassTypeID int64 `json:"pass_type_id"`
	// PracticeFormatID is a [ConstructorService.PracticeFormats] id.
	PracticeFormatID int64 `json:"practice_format_id"`
}

// ConstructorPracticeContent is the content, structure and report materials
// of a practice.
type ConstructorPracticeContent struct {
	Content           *ConstructorEditorJS `json:"content"`
	AdditionalContent string               `json:"additional_content"`
	Report            string               `json:"report"`
	// ReportDescription is HTML.
	ReportDescription string `json:"report_description"`
	// DisabledDescription is the provisions for students with disabilities.
	DisabledDescription string `json:"disabled_description"`
}

// ConstructorPracticeContentWrite edits the content and structure of a practice.
type ConstructorPracticeContentWrite struct {
	AdditionalContent *string             `json:"additional_content"`
	Content           ConstructorEditorJS `json:"content"`
}

// ConstructorGradeCriterion is a grading criterion of a practice.
type ConstructorGradeCriterion struct {
	ID int64 `json:"id"`
	// Name is the grade.
	Name string `json:"name"`
	// Description is newline-separated.
	Description string `json:"description"`
}

// ConstructorGradeDescription updates the description of a grading criterion.
type ConstructorGradeDescription struct {
	ID          int64  `json:"id"`
	Description string `json:"description"`
}

// ConstructorCertificationForm is the main info of a ГИА programme.
type ConstructorCertificationForm struct {
	Abbreviation  string `json:"abbreviation"`
	LanguageID    int64  `json:"language_id"`
	FinAttrID     int64  `json:"fin_attr_id"`
	ImplementerID int64  `json:"implementer_id"`
	// EducationLevels usually holds a single id.
	EducationLevels []int64 `json:"education_levels"`
	FormatID        int64   `json:"format_id"`
}

// ConstructorCertificationInfo is the main page of a ГИА programme.
type ConstructorCertificationInfo struct {
	ID              int64               `json:"id"`
	Abbreviation    string              `json:"abbreviation"`
	EducationLevels []ConstructorIDName `json:"education_levels"`
	FinAttrID       int64               `json:"fin_attr_id"`
	FinAttrName     string              `json:"fin_attr_name"`
	LanguageID      int64               `json:"language_id"`
	LanguageName    string              `json:"language_name"`
	ImplementerID   int64               `json:"implementer_id"`
	ImplementerName string              `json:"implementer_name"`
	FormatID        int64               `json:"format_id"`
	// DisabledDescription is the provisions for students with disabilities, newline-separated.
	DisabledDescription string              `json:"disabled_description"`
	Developers          []ConstructorPerson `json:"developers"`
	Authors             []ConstructorPerson `json:"authors"`
}

// ConstructorPreparationStage is a ГИА preparation stage.
type ConstructorPreparationStage struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	// Description is HTML.
	Description           string `json:"description"`
	AdditionalDescription string `json:"additional_description"`
}

// ConstructorPreparationStageUpdate sets the additional description of a
// preparation stage; newlines are written as <br>.
type ConstructorPreparationStageUpdate struct {
	ID                    int64   `json:"id"`
	AdditionalDescription *string `json:"additional_description"`
}

// ConstructorMainStage is a main ГИА stage with its deadline.
type ConstructorMainStage struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	// Deadline is free text as the server sends it; its format is not confirmed.
	Deadline   string `json:"deadline"`
	IsEditable bool   `json:"is_editable"`
}

// ConstructorMainStageUpdate sets the deadline of a main stage.
type ConstructorMainStageUpdate struct {
	ID       int64  `json:"id"`
	Deadline string `json:"deadline"`
}

// ConstructorCertificationAssessment is a ГИА assessment tool with its grades.
type ConstructorCertificationAssessment struct {
	ID     int64                  `json:"id"`
	Name   string                 `json:"name"`
	Grades []ConstructorCertGrade `json:"grades"`
}

// ConstructorCertGrade is a grade of a ГИА assessment tool.
type ConstructorCertGrade struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

func constructorPractice(programID int64) string {
	return "api/constructor/practices/" + id(programID)
}

func constructorCertification(programID int64) string {
	return "api/constructor/certifications/" + id(programID)
}

// CreatePractice creates a practice programme and returns its id. Staff only.
// POST /api/constructor/practices/create
func (s *ConstructorService) CreatePractice(ctx context.Context, form ConstructorPracticeCreate) (int64, error) {
	return call[int64](ctx, s.c, post("api/constructor/practices/create", form))
}

// UpdatePractice edits the main info of a draft practice programme via the
// discipline route. Staff only.
// PATCH /api/constructor/disciplines/{id}/info
func (s *ConstructorService) UpdatePractice(ctx context.Context, programID int64, form ConstructorPracticeForm) error {
	return exec(ctx, s.c, patch("api/constructor/disciplines/"+id(programID)+"/info", form))
}

// Practice returns the main info of a practice programme. Staff only.
// GET /api/constructor/practices/{id}/info
func (s *ConstructorService) Practice(ctx context.Context, programID int64) (*ConstructorPracticeInfo, error) {
	return call[*ConstructorPracticeInfo](ctx, s.c, get(constructorPractice(programID)+"/info", nil))
}

// SetPracticeCapacity sets the parts and credits of a draft practice programme. Staff only.
// PATCH /api/constructor/practices/{id}/capacity
func (s *ConstructorService) SetPracticeCapacity(ctx context.Context, programID int64, parts []ConstructorPracticePart) error {
	if parts == nil {
		parts = []ConstructorPracticePart{}
	}
	return exec(ctx, s.c, patch(constructorPractice(programID)+"/capacity", parts))
}

// UpdatePracticeAbout sets the kind, type, pass method and format of a practice. Staff only.
// PATCH /api/constructor/practices/{id}/about
func (s *ConstructorService) UpdatePracticeAbout(ctx context.Context, programID int64, about ConstructorPracticeAbout) error {
	return exec(ctx, s.c, patch(constructorPractice(programID)+"/about", about))
}

// UpdatePracticeDescription edits the "Особенности" text of a practice. Staff only.
// PATCH /api/constructor/practices/{id}/description
func (s *ConstructorService) UpdatePracticeDescription(ctx context.Context, programID int64, additionalDescription string) error {
	body := struct {
		AdditionalDescription string `json:"additional_description"`
	}{additionalDescription}
	return exec(ctx, s.c, patch(constructorPractice(programID)+"/description", body))
}

// PracticeTypes returns the practice types available for a programme. Staff only.
// GET /api/constructor/practices/{id}/types
func (s *ConstructorService) PracticeTypes(ctx context.Context, programID int64) ([]ConstructorIDName, error) {
	return call[[]ConstructorIDName](ctx, s.c, get(constructorPractice(programID)+"/types", nil))
}

// PracticeContent returns the content, structure and report materials of a practice. Staff only.
// GET /api/constructor/practices/{id}/content
func (s *ConstructorService) PracticeContent(ctx context.Context, programID int64) (*ConstructorPracticeContent, error) {
	return call[*ConstructorPracticeContent](ctx, s.c, get(constructorPractice(programID)+"/content", nil))
}

// UpdatePracticeContent edits the content and structure of a practice. Staff only.
// PATCH /api/constructor/practices/{id}/content
func (s *ConstructorService) UpdatePracticeContent(ctx context.Context, programID int64, content ConstructorPracticeContentWrite) error {
	return exec(ctx, s.c, patch(constructorPractice(programID)+"/content", content))
}

// UpdatePracticeReport edits the additional report materials of a practice. Staff only.
// PATCH /api/constructor/practices/{id}/report
func (s *ConstructorService) UpdatePracticeReport(ctx context.Context, programID int64, report string) error {
	body := struct {
		Report string `json:"report"`
	}{report}
	return exec(ctx, s.c, patch(constructorPractice(programID)+"/report", body))
}

// PracticeAssessments returns the grading criteria of a practice. Staff only.
// GET /api/constructor/practices/{id}/assessments
func (s *ConstructorService) PracticeAssessments(ctx context.Context, programID int64) ([]ConstructorGradeCriterion, error) {
	res, err := call[struct {
		Assessments []ConstructorGradeCriterion `json:"assessments"`
	}](ctx, s.c, get(constructorPractice(programID)+"/assessments", nil))
	return res.Assessments, err
}

// UpdatePracticeGrades edits the descriptions of the grading criteria. Staff only.
// PATCH /api/constructor/practices/{id}/grades
func (s *ConstructorService) UpdatePracticeGrades(ctx context.Context, programID int64, grades []ConstructorGradeDescription) error {
	if grades == nil {
		grades = []ConstructorGradeDescription{}
	}
	return exec(ctx, s.c, patch(constructorPractice(programID)+"/grades", grades))
}

// CreateCertification creates a ГИА programme and returns its id. Staff only.
// POST /api/constructor/certifications/create
func (s *ConstructorService) CreateCertification(ctx context.Context, form ConstructorCertificationForm) (int64, error) {
	return call[int64](ctx, s.c, post("api/constructor/certifications/create", form))
}

// Certification returns the main info of a ГИА programme. Staff only.
// GET /api/constructor/certifications/{id}/info
func (s *ConstructorService) Certification(ctx context.Context, programID int64) (*ConstructorCertificationInfo, error) {
	return call[*ConstructorCertificationInfo](ctx, s.c, get(constructorCertification(programID)+"/info", nil))
}

// UpdateCertification edits the main info of a draft ГИА programme. Staff only.
// PATCH /api/constructor/certifications/{id}/info
func (s *ConstructorService) UpdateCertification(ctx context.Context, programID int64, form ConstructorCertificationForm) error {
	return exec(ctx, s.c, patch(constructorCertification(programID)+"/info", form))
}

// PreparationStages returns the ГИА preparation stages. Staff only.
// GET /api/constructor/certifications/{id}/preparation_stages
func (s *ConstructorService) PreparationStages(ctx context.Context, programID int64) ([]ConstructorPreparationStage, error) {
	res, err := call[struct {
		Stages []ConstructorPreparationStage `json:"stages"`
	}](ctx, s.c, get(constructorCertification(programID)+"/preparation_stages", nil))
	return res.Stages, err
}

// UpdatePreparationStages saves the additional descriptions of preparation stages. Staff only.
// PATCH /api/constructor/certifications/{id}/preparation_stages
func (s *ConstructorService) UpdatePreparationStages(ctx context.Context, programID int64, stages []ConstructorPreparationStageUpdate) error {
	if stages == nil {
		stages = []ConstructorPreparationStageUpdate{}
	}
	return exec(ctx, s.c, patch(constructorCertification(programID)+"/preparation_stages", stages))
}

// MainStages returns the main ГИА stages and their deadlines. Staff only.
// GET /api/constructor/certifications/{id}/main_stages
func (s *ConstructorService) MainStages(ctx context.Context, programID int64) ([]ConstructorMainStage, error) {
	res, err := call[struct {
		Stages []ConstructorMainStage `json:"stages"`
	}](ctx, s.c, get(constructorCertification(programID)+"/main_stages", nil))
	return res.Stages, err
}

// UpdateMainStages saves the deadlines of editable main stages. Staff only.
// PATCH /api/constructor/certifications/{id}/main_stages
func (s *ConstructorService) UpdateMainStages(ctx context.Context, programID int64, stages []ConstructorMainStageUpdate) error {
	if stages == nil {
		stages = []ConstructorMainStageUpdate{}
	}
	return exec(ctx, s.c, patch(constructorCertification(programID)+"/main_stages", stages))
}

// CertificationAssessments returns the ГИА assessment tools of one type
// ([ConstructorService.CertificationAssessmentTypes] id). Staff only.
// GET /api/constructor/certifications/{id}/assessments/types/{typeId}
func (s *ConstructorService) CertificationAssessments(ctx context.Context, programID, typeID int64) ([]ConstructorCertificationAssessment, error) {
	res, err := call[struct {
		Assessments []ConstructorCertificationAssessment `json:"assessments"`
	}](ctx, s.c, get(constructorCertification(programID)+"/assessments/types/"+id(typeID), nil))
	return res.Assessments, err
}

// UpdateCertificationAssessment edits the grade descriptions of a ГИА assessment tool. Staff only.
// PATCH /api/constructor/certifications/{id}/assessments/{assessmentId}
func (s *ConstructorService) UpdateCertificationAssessment(ctx context.Context, programID, assessmentID int64, grades []ConstructorCertGrade) error {
	if grades == nil {
		grades = []ConstructorCertGrade{}
	}
	return exec(ctx, s.c, patch(constructorCertification(programID)+"/assessments/"+id(assessmentID), grades))
}
