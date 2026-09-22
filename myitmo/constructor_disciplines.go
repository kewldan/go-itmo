package myitmo

import (
	"context"
	"net/http"
)

// ConstructorDisciplineForm is the main info of a discipline programme (РПД or ДПО).
type ConstructorDisciplineForm struct {
	NameRU string `json:"name_ru"` // up to 200 characters
	NameEN string `json:"name_en"` // up to 200 characters, no Cyrillic
	// Abbreviation is up to 30 characters.
	Abbreviation    string  `json:"abbreviation"`
	FinAttrID       int64   `json:"fin_attr_id"`
	LanguageID      int64   `json:"language_id"`
	FormatID        int64   `json:"format_id"`
	ImplementerID   int64   `json:"implementer_id"`
	BarsRealization bool    `json:"bars_realization"`
	EducationLevels []int64 `json:"education_levels"`
	// ISUPublish is sent for continuing-education (ДПО) programmes only.
	ISUPublish bool `json:"isu_publish,omitzero"`
}

// ConstructorDisciplineCreate creates a discipline programme.
type ConstructorDisciplineCreate struct {
	ConstructorDisciplineForm
	// Contents lists the parts with their credits (1..30 each); nil (sent as
	// null) for continuing-education programmes.
	Contents []ConstructorPartCapacity `json:"contents"`
}

// ConstructorPartCapacity is the credit value of a new part.
type ConstructorPartCapacity struct {
	Capacity int `json:"capacity"`
}

// ConstructorDisciplineInfo is the main page of a discipline programme.
type ConstructorDisciplineInfo struct {
	ID              int64               `json:"id"`
	NameRU          string              `json:"name_ru"`
	NameEN          string              `json:"name_en"`
	Abbreviation    string              `json:"abbreviation"`
	LanguageID      int64               `json:"language_id"`
	LanguageName    string              `json:"language_name"`
	ImplementerID   int64               `json:"implementer_id"`
	ImplementerName string              `json:"implementer_name"`
	FinAttrID       int64               `json:"fin_attr_id"`
	FinAttrName     string              `json:"fin_attr_name"`
	FormatID        int64               `json:"format_id"`
	FormatName      string              `json:"format_name"`
	EducationLevels []ConstructorIDName `json:"education_levels"`
	BarsRealization bool                `json:"bars_realization"`
	IsDPO           bool                `json:"is_dpo"`
	ISUPublish      bool                `json:"isu_publish"`
	// Capacity is the total in credits.
	Capacity       float64              `json:"capacity"`
	Contents       []ConstructorContent `json:"contents"`
	AnnotationRU   string               `json:"annotation_ru"`
	AnnotationEN   string               `json:"annotation_en"`
	AdditionalInfo string               `json:"additional_info"`
	InTags         []ConstructorIDName  `json:"in_tags"`
	OutTags        []ConstructorOutTag  `json:"out_tags"`
	Companies      []ConstructorCompany `json:"companies"`
	Developers     []ConstructorPerson  `json:"developers"`
	Authors        []ConstructorPerson  `json:"authors"`
	Experts        []ConstructorPerson  `json:"experts"`
}

// ConstructorContent is a part (usually a semester) of a programme.
type ConstructorContent struct {
	// ID is the contentID of the /contents/{contentId} routes.
	ID int64 `json:"id"`
	// ContentOrder is the part number.
	ContentOrder            int                          `json:"content_order"`
	Capacity                float64                      `json:"capacity"`
	IntermediateAssessments []ConstructorContentWorkType `json:"intermediate_assessments"`
	WorkTypes               []ConstructorContentWorkType `json:"work_types"`
}

// ConstructorContentWorkType is the hours of one work type in a part.
type ConstructorContentWorkType struct {
	WorkTypeID int64  `json:"work_type_id"`
	Name       string `json:"name"`
	// Hours is nil for forms of intermediate control.
	Hours *float64 `json:"hours"`
}

// ConstructorOutTag is a learning outcome of a programme.
type ConstructorOutTag struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	// TypeID is a [ConstructorService.TagTypes] id (1 knowledge, 2 abilities, 3 skills).
	TypeID          int64 `json:"type_id"`
	ProgramOutTagID int64 `json:"program_out_tag_id"`
	// IsAttached means the outcome is used in chapters and cannot be removed.
	IsAttached bool `json:"is_attached"`
}

// ConstructorCompany is a company attached to a programme for the IT benefit.
type ConstructorCompany struct {
	// ID is the programme-company link id used by the file routes and removal.
	ID int64 `json:"id"`
	// TypeID is a [ConstructorService.ITTypes] id.
	TypeID           int64  `json:"type_id"`
	INN              string `json:"inn"`
	NameShortWithOPF string `json:"name_short_with_opf"`
	NameFullWithOPF  string `json:"name_full_with_opf"`
	FileAttached     bool   `json:"file_attached"`
	FileName         string `json:"file_name"`
	FileType         string `json:"file_type"`
}

// ConstructorDisciplineDescription is the "О дисциплине" section.
type ConstructorDisciplineDescription struct {
	AnnotationRU string `json:"annotation_ru"` // under 1000 characters
	// AnnotationEN is required unless the language is English (id 4); no Cyrillic.
	AnnotationEN   string `json:"annotation_en"`
	AdditionalInfo string `json:"additional_info"`
}

// ConstructorCapacityPart sets the credits and hours of one part.
type ConstructorCapacityPart struct {
	Capacity     int `json:"capacity"` // up to 30
	ContentOrder int `json:"content_order"`
	// IntermediateAssessments are the forms of control (hours nil), plus
	// optionally course work (7) or course project (8).
	IntermediateAssessments []ConstructorWorkTypeRef   `json:"intermediate_assessments"`
	WorkTypes               []ConstructorWorkTypeHours `json:"work_types"`
}

// ConstructorWorkTypeRef names a form of control; Hours is sent as null.
type ConstructorWorkTypeRef struct {
	WorkTypeID int64    `json:"work_type_id"`
	Hours      *float64 `json:"hours"`
}

// ConstructorWorkTypeHours is the hours of one work type.
type ConstructorWorkTypeHours struct {
	WorkTypeID int64   `json:"work_type_id"`
	Hours      float64 `json:"hours"`
}

// ConstructorDisciplineUse is a place where a discipline is used.
type ConstructorDisciplineUse struct {
	EPID       int64  `json:"ep_id"`
	EPName     string `json:"ep_name"`
	Link       string `json:"link"`
	ModuleID   int64  `json:"module_id"`
	ModuleName string `json:"module_name"`
	Semesters  []int  `json:"semesters"`
}

// ConstructorOutcomeRef selects a learning outcome tag of a type.
type ConstructorOutcomeRef struct {
	ID     int64 `json:"id"`
	TypeID int64 `json:"type_id"`
}

// ConstructorTagType is a learning outcome type.
type ConstructorTagType struct {
	// ID is 1 knowledge, 2 abilities, 3 skills.
	ID   int64  `json:"id"`
	Name string `json:"name"`
	// ThirdPerson is the verb shown before a tag of this type.
	ThirdPerson string `json:"third_person"`
}

// ConstructorITType is an IT-benefit company type.
type ConstructorITType struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Order int    `json:"order"`
	// FileAllowed means a supporting file may be attached.
	FileAllowed bool `json:"file_allowed"`
}

// ConstructorCompanyRef is a company search result.
type ConstructorCompanyRef struct {
	INN              string `json:"inn"`
	NameShortWithOPF string `json:"name_short_with_opf"`
	NameFullWithOPF  string `json:"name_full_with_opf"`
}

// ConstructorChapters is the chapter structure of a part.
type ConstructorChapters struct {
	Chapters         []ConstructorChapter         `json:"chapters"`
	WorkTypeLimits   []ConstructorWorkTypeLimit   `json:"work_type_limits"`
	Capacity         float64                      `json:"capacity"`
	DistributedTotal []ConstructorWorkTypeTotal   `json:"distributed_total"`
	Errors           []ConstructorValidationError `json:"errors"`
}

// ConstructorWorkTypeLimit is the hours of a work type available to chapters.
type ConstructorWorkTypeLimit struct {
	WorkTypeID int64   `json:"work_type_id"`
	Total      float64 `json:"total"`
	// ProgramWorkTypeID is sent back in [ConstructorChapterWrite].
	ProgramWorkTypeID int64 `json:"program_work_type_id"`
}

// ConstructorWorkTypeTotal is the hours of a work type already distributed.
type ConstructorWorkTypeTotal struct {
	WorkTypeID int64   `json:"work_type_id"`
	Total      float64 `json:"total"`
}

// ConstructorChapter is a chapter ("Раздел") of a part.
type ConstructorChapter struct {
	ID        int64                      `json:"id"`
	Name      string                     `json:"name"`
	Order     int                        `json:"order"`
	WorkTypes []ConstructorWorkTypeHours `json:"work_types"`
	Themes    []ConstructorTheme         `json:"themes"`
}

// ConstructorTheme is a theme of a chapter.
type ConstructorTheme struct {
	ID        int64                 `json:"id,omitzero"`
	Name      string                `json:"name"`
	Order     int                   `json:"order"`
	Resources []ConstructorResource `json:"resources"`
}

// ConstructorResource is a link attached to a theme.
type ConstructorResource struct {
	// Name is usually null.
	Name *string `json:"name"`
	Link string  `json:"link"`
}

// ConstructorChapterWrite creates or updates a chapter.
type ConstructorChapterWrite struct {
	Name  string `json:"name"`
	Order int    `json:"order"`
	// Themes are sent without ids.
	Themes           []ConstructorTheme                `json:"themes"`
	ProgramWorkTypes []ConstructorProgramWorkTypeHours `json:"program_work_types"`
}

// ConstructorProgramWorkTypeHours assigns hours of a work type to a chapter.
type ConstructorProgramWorkTypeHours struct {
	// ProgramWorkTypeID comes from [ConstructorWorkTypeLimit].
	ProgramWorkTypeID int64   `json:"program_work_type_id"`
	Hours             float64 `json:"hours"`
}

// ConstructorContentCapacity sets the credits and hours of a single part;
// zero entries may be left out.
type ConstructorContentCapacity struct {
	Capacity  float64                    `json:"capacity"`
	WorkTypes []ConstructorWorkTypeHours `json:"work_types"`
}

// ConstructorAssessments is the assessment tools of a part.
type ConstructorAssessments struct {
	IntermediateAssessments ConstructorIntermediateAssessments `json:"intermediate_assessments"`
	CurrentAssessments      []ConstructorCurrentAssessment     `json:"current_assessments"`
}

// ConstructorIntermediateAssessments is the intermediate control of a part.
type ConstructorIntermediateAssessments struct {
	List []ConstructorIntermediateAssessment `json:"list"`
	// AdditionalScore is 0 or 3 (the "3 additional points" flag).
	AdditionalScore int `json:"additional_score"`
}

// ConstructorIntermediateAssessment is a form of intermediate control of a part.
type ConstructorIntermediateAssessment struct {
	ID           int64                `json:"id"`
	Name         string               `json:"name"`
	MinScore     float64              `json:"min_score"`
	MaxScore     float64              `json:"max_score"`
	IsCourseWork bool                 `json:"is_course_work"`
	Description  *ConstructorEditorJS `json:"description"`
}

// ConstructorCurrentAssessment is a current-assessment tool of a part.
type ConstructorCurrentAssessment struct {
	ID           int64   `json:"id"`
	Name         string  `json:"name"`
	Order        int     `json:"order"`
	MinScore     float64 `json:"min_score"`
	MaxScore     float64 `json:"max_score"`
	ControlPoint bool    `json:"control_point"`
	// CurrentAssessmentID is a [ConstructorService.CurrentAssessmentTypes] id.
	CurrentAssessmentID   int64                   `json:"current_assessment_id"`
	CurrentAssessmentName string                  `json:"current_assessment_name"`
	Chapters              []ConstructorChapterRef `json:"chapters"`
	Description           *ConstructorEditorJS    `json:"description"`
}

// ConstructorChapterRef names a chapter covered by an assessment.
type ConstructorChapterRef struct {
	ChapterID   int64  `json:"chapter_id"`
	ChapterName string `json:"chapter_name"`
}

// ConstructorIntermediateAssessmentWrite edits an intermediate assessment.
type ConstructorIntermediateAssessmentWrite struct {
	Description ConstructorEditorJS `json:"description"`
	MinScore    float64             `json:"min_score"`
	MaxScore    float64             `json:"max_score"`
}

// ConstructorCurrentAssessmentWrite creates or edits a current-assessment tool.
type ConstructorCurrentAssessmentWrite struct {
	Description  ConstructorEditorJS `json:"description"`
	Name         string              `json:"name"`
	Order        int                 `json:"order"`
	ControlPoint bool                `json:"control_point"`
	MinScore     float64             `json:"min_score"`
	// MaxScore is at most 100 for BARS programmes.
	MaxScore float64 `json:"max_score"`
	// CurrentAssessmentID is a [ConstructorService.CurrentAssessmentTypes] id.
	CurrentAssessmentID int64 `json:"current_assessment_id"`
	// ChaptersID lists the covered chapter ids.
	ChaptersID []int64 `json:"chapters_id"`
}

// ConstructorSources is the literature of a programme.
type ConstructorSources struct {
	LibrarySources []ConstructorIDName `json:"library_sources"`
	OtherSources   []ConstructorIDName `json:"other_sources"`
}

func constructorContent(programID, contentID int64) string {
	return constructorProgram(programID) + "/contents/" + id(contentID)
}

// CreateDiscipline creates a discipline programme and returns its id. Staff only.
// POST /api/constructor/disciplines/create
func (s *ConstructorService) CreateDiscipline(ctx context.Context, form ConstructorDisciplineCreate) (int64, error) {
	body := struct {
		ConstructorDisciplineForm
		Contents *[]ConstructorPartCapacity `json:"contents"`
	}{ConstructorDisciplineForm: form.ConstructorDisciplineForm}
	if form.Contents != nil {
		body.Contents = &form.Contents
	}
	return call[int64](ctx, s.c, post("api/constructor/disciplines/create", body))
}

// CreateDPO creates a continuing-education programme from an existing
// discipline programme and returns the new id. Staff only.
// POST /api/constructor/disciplines/{programId}/create_dpo
func (s *ConstructorService) CreateDPO(ctx context.Context, sourceID int64) (int64, error) {
	return call[int64](ctx, s.c, post("api/constructor/disciplines/"+id(sourceID)+"/create_dpo", nil))
}

// Discipline returns the main info of a discipline programme. Staff only.
// GET /api/constructor/disciplines/{id}/info
func (s *ConstructorService) Discipline(ctx context.Context, programID int64) (*ConstructorDisciplineInfo, error) {
	return call[*ConstructorDisciplineInfo](ctx, s.c, get("api/constructor/disciplines/"+id(programID)+"/info", nil))
}

// UpdateDiscipline edits the main info of a draft discipline programme. Staff only.
// PATCH /api/constructor/disciplines/{id}/info
func (s *ConstructorService) UpdateDiscipline(ctx context.Context, programID int64, form ConstructorDisciplineForm) error {
	return exec(ctx, s.c, patch("api/constructor/disciplines/"+id(programID)+"/info", form))
}

// UpdateDisciplineDescription edits the annotation of a discipline programme. Staff only.
// PATCH /api/constructor/disciplines/{id}/description
func (s *ConstructorService) UpdateDisciplineDescription(ctx context.Context, programID int64, d ConstructorDisciplineDescription) error {
	return exec(ctx, s.c, patch("api/constructor/disciplines/"+id(programID)+"/description", d))
}

// SetDisciplineCapacity sets the parts, credits and hours of a draft
// discipline programme. Staff only.
// PATCH /api/constructor/disciplines/{id}/capacity
func (s *ConstructorService) SetDisciplineCapacity(ctx context.Context, programID int64, parts []ConstructorCapacityPart) error {
	if parts == nil {
		parts = []ConstructorCapacityPart{}
	}
	return exec(ctx, s.c, patch("api/constructor/disciplines/"+id(programID)+"/capacity", parts))
}

// ImportDiscipline copies the description, tags, outcomes, authors, chapters,
// sources and assessments of programme from into programme to, overwriting
// them. Staff only.
// POST /api/constructor/disciplines/{fromProgramId}/{toProgramId}
func (s *ConstructorService) ImportDiscipline(ctx context.Context, from, to int64) error {
	return exec(ctx, s.c, post("api/constructor/disciplines/"+id(from)+"/"+id(to), nil))
}

// DisciplineUsage returns the educational programmes and modules that use a
// discipline. Staff only.
// GET /api/constructor/disciplines/{id}/use
func (s *ConstructorService) DisciplineUsage(ctx context.Context, programID int64) ([]ConstructorDisciplineUse, error) {
	return call[[]ConstructorDisciplineUse](ctx, s.c, get("api/constructor/disciplines/"+id(programID)+"/use", nil))
}

// SetIncomingTags sets the input knowledge tags. Staff only.
// PATCH /api/constructor/programs/{id}/incoming_tags
func (s *ConstructorService) SetIncomingTags(ctx context.Context, programID int64, tagIDs []int64) error {
	if tagIDs == nil {
		tagIDs = []int64{}
	}
	return exec(ctx, s.c, patch(constructorProgram(programID)+"/incoming_tags", tagIDs))
}

// SetOutcomingTags sets the learning outcomes. Staff only.
// PATCH /api/constructor/programs/{id}/outcoming_tags
func (s *ConstructorService) SetOutcomingTags(ctx context.Context, programID int64, tags []ConstructorOutcomeRef) error {
	if tags == nil {
		tags = []ConstructorOutcomeRef{}
	}
	return exec(ctx, s.c, patch(constructorProgram(programID)+"/outcoming_tags", tags))
}

// TagTypes returns the learning outcome types. Staff only.
// GET /api/constructor/tags/types
func (s *ConstructorService) TagTypes(ctx context.Context) ([]ConstructorTagType, error) {
	return call[[]ConstructorTagType](ctx, s.c, get("api/constructor/tags/types", nil))
}

// SearchTags autocompletes tags. Staff only.
// GET /api/constructor/tags/search
func (s *ConstructorService) SearchTags(ctx context.Context, query string) ([]ConstructorIDName, error) {
	return call[[]ConstructorIDName](ctx, s.c, get("api/constructor/tags/search", q().set("query", query)))
}

// CreateTag creates a tag (up to 200 characters) and returns its id. Staff only.
// POST /api/constructor/tags/create
func (s *ConstructorService) CreateTag(ctx context.Context, name string) (int64, error) {
	body := struct {
		Name string `json:"name"`
	}{name}
	return call[int64](ctx, s.c, post("api/constructor/tags/create", body))
}

// ITTypes returns the IT-benefit company types. Staff only.
// GET /api/constructor/it_types
func (s *ConstructorService) ITTypes(ctx context.Context) ([]ConstructorITType, error) {
	return call[[]ConstructorITType](ctx, s.c, get("api/constructor/it_types", nil))
}

// SearchCompanies finds companies by name or INN (at least 3 characters). Staff only.
// GET /api/constructor/companies
func (s *ConstructorService) SearchCompanies(ctx context.Context, query string) ([]ConstructorCompanyRef, error) {
	return call[[]ConstructorCompanyRef](ctx, s.c, get("api/constructor/companies", q().set("query", query)))
}

// AddCompany attaches a company to a programme under an IT-benefit type
// ([ConstructorService.ITTypes] id). Staff only.
// POST /api/constructor/programs/{id}/companies
func (s *ConstructorService) AddCompany(ctx context.Context, programID int64, company ConstructorCompanyRef, typeID int64) error {
	body := struct {
		ConstructorCompanyRef
		TypeID int64 `json:"type_id"`
	}{company, typeID}
	return exec(ctx, s.c, post(constructorProgram(programID)+"/companies", body))
}

// RemoveCompany detaches a company by [ConstructorCompany].ID. Staff only.
// DELETE /api/constructor/programs/{id}/companies/{companyId}
func (s *ConstructorService) RemoveCompany(ctx context.Context, programID, companyID int64) error {
	return exec(ctx, s.c, del(constructorProgram(programID)+"/companies/"+id(companyID), nil))
}

// UploadCompanyFile attaches a supporting file to a company link
// ([ConstructorCompany].ID). Staff only.
// POST /api/constructor/companies/{companyId}/file
func (s *ConstructorService) UploadCompanyFile(ctx context.Context, companyID int64, file Upload) error {
	file.Field = "file"
	return exec(ctx, s.c, multipart(http.MethodPost, "api/constructor/companies/"+id(companyID)+"/file", nil, file))
}

// CompanyFile downloads the supporting file of a company link. The caller closes the file. Staff only.
// GET /api/constructor/companies/{companyId}/file
func (s *ConstructorService) CompanyFile(ctx context.Context, companyID int64) (*File, error) {
	return download(ctx, s.c, get("api/constructor/companies/"+id(companyID)+"/file", nil))
}

// DeleteCompanyFile removes the supporting file of a company link. Staff only.
// DELETE /api/constructor/companies/{companyId}/file
func (s *ConstructorService) DeleteCompanyFile(ctx context.Context, companyID int64) error {
	return exec(ctx, s.c, del("api/constructor/companies/"+id(companyID)+"/file", nil))
}

// Chapters returns the chapters of a part with the hour limits. Staff only.
// GET /api/constructor/programs/{id}/contents/{contentId}/chapters
func (s *ConstructorService) Chapters(ctx context.Context, programID, contentID int64) (*ConstructorChapters, error) {
	return call[*ConstructorChapters](ctx, s.c, get(constructorContent(programID, contentID)+"/chapters", nil))
}

// CreateChapter creates a chapter in a part. Staff only.
// POST /api/constructor/programs/{id}/contents/{contentId}/chapters/create
func (s *ConstructorService) CreateChapter(ctx context.Context, programID, contentID int64, chapter ConstructorChapterWrite) error {
	return exec(ctx, s.c, post(constructorContent(programID, contentID)+"/chapters/create", chapter))
}

// UpdateChapter edits a chapter. Staff only.
// PATCH /api/constructor/programs/{id}/contents/{contentId}/chapters/{chapterId}
func (s *ConstructorService) UpdateChapter(ctx context.Context, programID, contentID, chapterID int64, chapter ConstructorChapterWrite) error {
	return exec(ctx, s.c, patch(constructorContent(programID, contentID)+"/chapters/"+id(chapterID), chapter))
}

// DeleteChapter deletes a chapter. Staff only.
// DELETE /api/constructor/programs/{id}/contents/{contentId}/chapters/{chapterId}
func (s *ConstructorService) DeleteChapter(ctx context.Context, programID, contentID, chapterID int64) error {
	return exec(ctx, s.c, del(constructorContent(programID, contentID)+"/chapters/"+id(chapterID), nil))
}

// SetContentCapacity changes the credits and hours of one part. Staff only.
// PATCH /api/constructor/programs/{id}/contents/{contentId}/capacity
func (s *ConstructorService) SetContentCapacity(ctx context.Context, programID, contentID int64, capacity ConstructorContentCapacity) error {
	return exec(ctx, s.c, patch(constructorContent(programID, contentID)+"/capacity", capacity))
}

// SplitIndependentWork distributes the independent-work hours evenly across
// the chapters of a part. Staff only.
// POST /api/constructor/programs/{id}/contents/{contentId}/capacity/split
func (s *ConstructorService) SplitIndependentWork(ctx context.Context, programID, contentID int64) error {
	return exec(ctx, s.c, post(constructorContent(programID, contentID)+"/capacity/split", nil))
}

// Assessments returns the intermediate and current assessment tools of a part. Staff only.
// GET /api/constructor/programs/{id}/contents/{contentId}/work_types
func (s *ConstructorService) Assessments(ctx context.Context, programID, contentID int64) (*ConstructorAssessments, error) {
	return call[*ConstructorAssessments](ctx, s.c, get(constructorContent(programID, contentID)+"/work_types", nil))
}

// UpdateIntermediateAssessment edits the scores and description of an
// intermediate assessment. Staff only.
// PATCH /api/constructor/programs/{id}/contents/{contentId}/work_types/{assessmentId}
func (s *ConstructorService) UpdateIntermediateAssessment(ctx context.Context, programID, contentID, assessmentID int64, a ConstructorIntermediateAssessmentWrite) error {
	return exec(ctx, s.c, patch(constructorContent(programID, contentID)+"/work_types/"+id(assessmentID), a))
}

// DeleteIntermediateAssessment deletes an intermediate assessment of a draft. Staff only.
// DELETE /api/constructor/programs/{id}/contents/{contentId}/work_types/{assessmentId}
func (s *ConstructorService) DeleteIntermediateAssessment(ctx context.Context, programID, contentID, assessmentID int64) error {
	return exec(ctx, s.c, del(constructorContent(programID, contentID)+"/work_types/"+id(assessmentID), nil))
}

// ToggleAdditionalScore toggles the "3 additional points" flag of a part;
// the state is reported as AdditionalScore. Staff only.
// PATCH /api/constructor/programs/{id}/contents/{contentId}/work_types/extra
func (s *ConstructorService) ToggleAdditionalScore(ctx context.Context, programID, contentID int64) error {
	return exec(ctx, s.c, patch(constructorContent(programID, contentID)+"/work_types/extra", nil))
}

// CreateCurrentAssessment adds a current-assessment tool to a part. Staff only.
// POST /api/constructor/programs/{id}/contents/{contentId}/assessments/current/create
func (s *ConstructorService) CreateCurrentAssessment(ctx context.Context, programID, contentID int64, a ConstructorCurrentAssessmentWrite) error {
	return exec(ctx, s.c, post(constructorContent(programID, contentID)+"/assessments/current/create", a))
}

// UpdateCurrentAssessment edits a current-assessment tool. Staff only.
// PATCH /api/constructor/programs/{id}/contents/{contentId}/assessments/current/{assessmentId}
func (s *ConstructorService) UpdateCurrentAssessment(ctx context.Context, programID, contentID, assessmentID int64, a ConstructorCurrentAssessmentWrite) error {
	return exec(ctx, s.c, patch(constructorContent(programID, contentID)+"/assessments/current/"+id(assessmentID), a))
}

// DeleteCurrentAssessment deletes a current-assessment tool. Staff only.
// DELETE /api/constructor/programs/{id}/contents/{contentId}/assessments/current/{assessmentId}
func (s *ConstructorService) DeleteCurrentAssessment(ctx context.Context, programID, contentID, assessmentID int64) error {
	return exec(ctx, s.c, del(constructorContent(programID, contentID)+"/assessments/current/"+id(assessmentID), nil))
}

// UploadImage uploads an image for an EditorJS description and returns its URL. Staff only.
// POST /api/constructor/static/image
func (s *ConstructorService) UploadImage(ctx context.Context, image Upload) (string, error) {
	image.Field = "image"
	return call[string](ctx, s.c, multipart(http.MethodPost, "api/constructor/static/image", nil, image))
}

// Sources returns the literature of a programme. Staff only.
// GET /api/constructor/programs/{id}/sources
func (s *ConstructorService) Sources(ctx context.Context, programID int64) (*ConstructorSources, error) {
	return call[*ConstructorSources](ctx, s.c, get(constructorProgram(programID)+"/sources", nil))
}

// AddLibrarySources adds library catalogue sources by id. Staff only.
// POST /api/constructor/programs/{id}/sources
func (s *ConstructorService) AddLibrarySources(ctx context.Context, programID int64, sourceIDs []int64) error {
	if sourceIDs == nil {
		sourceIDs = []int64{}
	}
	return exec(ctx, s.c, post(constructorProgram(programID)+"/sources", sourceIDs))
}

// RemoveLibrarySource removes a library source. Staff only.
// DELETE /api/constructor/programs/{id}/sources/{sourceId}
func (s *ConstructorService) RemoveLibrarySource(ctx context.Context, programID, sourceID int64) error {
	return exec(ctx, s.c, del(constructorProgram(programID)+"/sources/"+id(sourceID), nil))
}

// AddOtherSource adds a free-text resource. Staff only.
// POST /api/constructor/programs/{id}/sources/other
func (s *ConstructorService) AddOtherSource(ctx context.Context, programID int64, name string) error {
	body := struct {
		Name string `json:"name"`
	}{name}
	return exec(ctx, s.c, post(constructorProgram(programID)+"/sources/other", body))
}

// RemoveOtherSource removes a free-text resource. Staff only.
// DELETE /api/constructor/programs/{id}/sources/other/{sourceId}
func (s *ConstructorService) RemoveOtherSource(ctx context.Context, programID, sourceID int64) error {
	return exec(ctx, s.c, del(constructorProgram(programID)+"/sources/other/"+id(sourceID), nil))
}

// SearchLibrary searches the university library catalogue; Name is the
// bibliographic description. Staff only.
// GET /api/constructor/sources/library/search/{query}
func (s *ConstructorService) SearchLibrary(ctx context.Context, query string) ([]ConstructorIDName, error) {
	return call[[]ConstructorIDName](ctx, s.c, get("api/constructor/sources/library/search/"+id(query), nil))
}
