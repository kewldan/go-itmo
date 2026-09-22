package myitmo

import (
	"context"
	"net/http"
	"time"
)

// ConstructorService is the work-programme constructor (/api/constructor):
// discipline (РПД), practice (РПП) and state final certification (ГИА)
// programmes, their workflow, people, contents and references. Almost every
// route is for staff: authors, editors, experts and administrators.
type ConstructorService struct{ c *Client }

// Programme types (program_type_id / type_id). A continuing-education (ДПО)
// programme is a discipline with dpo=1.
const (
	ConstructorTypeDiscipline    = 1 // РПД
	ConstructorTypePractice      = 2 // РПП
	ConstructorTypeCertification = 3 // ГИА
)

// Programme statuses (status_id). Names come from [ConstructorService.Statuses].
const (
	ConstructorStatusDraft       = 1 // editable; main info may change only here
	ConstructorStatusPublished   = 2
	ConstructorStatusReturned    = 3 // returned for revision
	ConstructorStatusArchived    = 4
	ConstructorStatusOnExpertise = 5
	ConstructorStatusSigned      = 6 // a signed PDF exists
	ConstructorStatusOnSigning   = 7
)

// Work type ids (work_type_id).
const (
	ConstructorWorkLectures      = 1
	ConstructorWorkLabs          = 2
	ConstructorWorkPractices     = 3
	ConstructorWorkCourseWork    = 7
	ConstructorWorkCourseProject = 8
	ConstructorWorkConsultations = 12
	ConstructorWorkIndependent   = 22  // СРО, independent work
	ConstructorWorkContactTotal  = 128 // contact work in total
)

// ConstructorIDName is a dictionary entry of the constructor.
type ConstructorIDName struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// ConstructorPerson is a developer, author or expert attached to a programme.
type ConstructorPerson struct {
	// ID is the row id inside the programme, used to remove the person.
	ID int64 `json:"id"`
	// ISU is nil for external authors.
	ISU *int64 `json:"isu"`
	FIO string `json:"fio"`
}

// ConstructorPersonRef is a people-search result, also sent to add a
// developer or an author.
type ConstructorPersonRef struct {
	// ISU is nil for an external author entered as free text.
	ISU *int64 `json:"isu"`
	FIO string `json:"fio"`
}

// ConstructorValidationError is a condition that blocks publishing or
// expertise.
type ConstructorValidationError struct {
	Text    string   `json:"text"`
	Objects []string `json:"objects"`
}

// ConstructorEditorJS is EditorJS output data. Besides the standard blocks the
// constructor uses the block type "math" with data {"math": "<LaTeX>"}.
type ConstructorEditorJS struct {
	Time    int64                      `json:"time,omitzero"`
	Blocks  []ConstructorEditorJSBlock `json:"blocks"`
	Version string                     `json:"version,omitzero"`
}

// ConstructorEditorJSBlock is one EditorJS block.
type ConstructorEditorJSBlock struct {
	ID   string `json:"id,omitzero"`
	Type string `json:"type"`
	// Data depends on Type.
	Data RawJSON `json:"data"`
}

// ConstructorSignTask is a signing task of a programme document; its files
// are served by the signing service (/api/sign/tasks/{task_id}).
type ConstructorSignTask struct {
	TaskID     FlexID                   `json:"task_id"`
	Name       string                   `json:"name"`
	Type       string                   `json:"type"`   // "pdf"
	Status     string                   `json:"status"` // e.g. "SIGNED"
	Signatures []ConstructorSignatureID `json:"signatures"`
}

// ConstructorSignatureID is a signature slot of a signing task.
type ConstructorSignatureID struct {
	SignatureID FlexID `json:"signature_id"`
}

// ConstructorProgramList is a page of work programmes.
type ConstructorProgramList struct {
	Count    int                      `json:"count"`
	Programs []ConstructorProgramItem `json:"programs"`
	// CanSign reports whether the user may sign documents.
	CanSign bool `json:"can_sign"`
	// IsDPOExpert reports whether the user is a continuing-education expert.
	IsDPOExpert bool `json:"is_dpo_expert"`
}

// ConstructorProgramItem is a row of a programme list.
type ConstructorProgramItem struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	// Capacity is in credits (з.е.).
	Capacity        float64             `json:"capacity"`
	ImplementerName string              `json:"implementer_name"`
	FinAttrName     string              `json:"fin_attr_name"`
	LanguageCode    string              `json:"language_code"` // e.g. "RU"
	EducationLevels []ConstructorIDName `json:"education_levels"`
	DevelopersCount int                 `json:"developers_count"`
	Developers      []ConstructorPerson `json:"developers"`
	StatusID        int64               `json:"status_id"`
	StatusName      string              `json:"status_name"`
	// Experts is sent by the expertise list only.
	Experts []ConstructorPerson `json:"experts,omitzero"`
}

// ConstructorProgramListParams filters [ConstructorService.Programs].
type ConstructorProgramListParams struct {
	Limit  int
	Offset int
	// Query searches by name or id.
	Query string
	// ProgramTypeID is one of the ConstructorType* constants.
	ProgramTypeID    *int64
	StatusID         *int64
	EducationLevelID *int64
	ImplementerID    *int64
	LanguageID       *int64
	// MyDisciplines limits the list to the user's programmes.
	MyDisciplines bool
	// Archive shows archived programmes (usually together with StatusID 4).
	Archive bool
	// DPO lists continuing-education programmes.
	DPO bool
}

// ConstructorExpertiseListParams filters [ConstructorService.ExpertisePrograms].
type ConstructorExpertiseListParams struct {
	Limit         int
	Offset        int
	Query         string
	MyDisciplines bool
	DPO           bool
}

// ConstructorSignTaskParams filters [ConstructorService.SignTasks].
type ConstructorSignTaskParams struct {
	Query            string
	EducationLevelID *int64
	ImplementerID    *int64
	LanguageID       *int64
}

// ConstructorChanges is a page of the change history.
type ConstructorChanges struct {
	Count int                 `json:"count"`
	Rows  []ConstructorChange `json:"rows"`
}

// ConstructorChange is one change of a programme.
type ConstructorChange struct {
	Timestamp time.Time `json:"timestamp"`
	// ProgramTypeName and ProgramID are set in the global history only.
	ProgramTypeName string `json:"program_type_name,omitzero"`
	ProgramID       int64  `json:"program_id,omitzero"`
	// ProgramPartName is set in the history of one programme only.
	ProgramPartName string `json:"program_part_name,omitzero"`
	// Name is the title of the change.
	Name           string `json:"name"`
	ActorISU       int64  `json:"actor_isu"`
	ActorFIO       string `json:"actor_fio"`
	PriorState     string `json:"prior_state"`
	ResultingState string `json:"resulting_state"`
}

// ConstructorProgramStatus is the header of a programme page with the user's
// permissions.
type ConstructorProgramStatus struct {
	Menu       []ConstructorMenuItem `json:"menu"`
	StatusID   int64                 `json:"status_id"`
	StatusName string                `json:"status_name"`
	Name       string                `json:"name"`
	// TypeID is one of the ConstructorType* constants.
	TypeID   int64 `json:"type_id"`
	CanEdit  bool  `json:"can_edit"`
	IsAdmin  bool  `json:"is_admin"`
	IsExpert bool  `json:"is_expert"`
	CanSign  bool  `json:"can_sign"`
	// Errors block publishing and sending to expertise while non-empty.
	Errors []ConstructorValidationError `json:"errors"`
	// ContentsCount is the number of parts; copying is allowed only when it is 1.
	ContentsCount    int                  `json:"contents_count"`
	SignatureTask    *ConstructorSignTask `json:"signature_task"`
	CanEditCompanies bool                 `json:"can_edit_companies"`
}

// ConstructorMenuItem is a tab (programme part) of a programme page.
type ConstructorMenuItem struct {
	// ID is the programme part id, used by comments and the change filter.
	ID   int64  `json:"id"`
	Name string `json:"name"`
	// Query names the tab: mainRPD, chapters, assessments, mainGIA,
	// preparations, stages, assessmentsGIA, changes, sources, skills, ep,
	// mainRPP, content or assessmentsRPP.
	Query string `json:"query"`
	Order int    `json:"order"`
}

// ConstructorStatusAction is a workflow transition of a programme.
type ConstructorStatusAction string

// Workflow transitions.
const (
	// ConstructorPublish publishes a draft.
	ConstructorPublish ConstructorStatusAction = "publish"
	// ConstructorToExamination sends a published or returned programme to expertise.
	ConstructorToExamination ConstructorStatusAction = "to_examination"
	// ConstructorToRevision returns a programme from expertise or signing.
	ConstructorToRevision ConstructorStatusAction = "to_revision"
	// ConstructorToSigning sends a programme from expertise to signing.
	ConstructorToSigning ConstructorStatusAction = "to_signing"
	// ConstructorArchive archives a programme; it can no longer be edited or used in plans.
	ConstructorArchive ConstructorStatusAction = "archive"
)

// ConstructorDocumentFormat is the format of a generated programme document.
type ConstructorDocumentFormat string

// Document formats.
const (
	ConstructorDOCX ConstructorDocumentFormat = "docx"
	ConstructorPDF  ConstructorDocumentFormat = "pdf"
)

// ConstructorProgramJSON is the JSON dump of a work programme. Only the
// annotation is known; the rest is kept in Unknown.
type ConstructorProgramJSON struct {
	DisciplineFields ConstructorProgramJSONFields `json:"discipline_fields"`
	// Unknown holds the other members of the dump.
	Unknown RawJSON `json:",embed"`
}

// ConstructorProgramJSONFields is the discipline part of [ConstructorProgramJSON].
type ConstructorProgramJSONFields struct {
	AnnotationRU string  `json:"annotation_ru"`
	Unknown      RawJSON `json:",embed"`
}

// ConstructorComment is an expert comment on a programme part.
type ConstructorComment struct {
	ID int64 `json:"id"`
	// ProgramPartID matches [ConstructorMenuItem].ID.
	ProgramPartID int64     `json:"program_part_id"`
	Comment       string    `json:"comment"`
	FIO           string    `json:"fio"`
	CreatedAt     time.Time `json:"created_at"`
}

// ConstructorStatus is a programme status of the reference list.
type ConstructorStatus struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	// Style is a Bootstrap badge variant.
	Style string `json:"style"`
}

// ConstructorFormat is an implementation format.
type ConstructorFormat struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// ConstructorIntermediateAssessmentType is a form of intermediate control.
type ConstructorIntermediateAssessmentType struct {
	// ID equals the work type id (7 course work, 8 course project).
	ID           int64  `json:"id"`
	Name         string `json:"name"`
	IsCourseWork bool   `json:"is_course_work"`
}

func constructorFlag(b bool) int {
	if b {
		return 1
	}
	return 0
}

func constructorProgram(programID int64) string {
	return "api/constructor/programs/" + id(programID)
}

// Programs searches work programmes. Staff only.
// GET /api/constructor/programs/list
func (s *ConstructorService) Programs(ctx context.Context, p ConstructorProgramListParams) (*ConstructorProgramList, error) {
	qv := q().set("limit", p.Limit).set("offset", p.Offset).set("query", p.Query).
		set("program_type_id", p.ProgramTypeID).set("my_disciplines", constructorFlag(p.MyDisciplines)).
		set("archive", constructorFlag(p.Archive)).set("status_id", p.StatusID).
		set("education_level_id", p.EducationLevelID).set("implementer_id", p.ImplementerID).
		set("language_id", p.LanguageID).set("dpo", constructorFlag(p.DPO))
	return call[*ConstructorProgramList](ctx, s.c, get("api/constructor/programs/list", qv))
}

// ExpertisePrograms lists programmes assigned for expertise. Staff only.
// GET /api/constructor/programs/expertise/list
func (s *ConstructorService) ExpertisePrograms(ctx context.Context, p ConstructorExpertiseListParams) (*ConstructorProgramList, error) {
	qv := q().set("limit", p.Limit).set("offset", p.Offset).set("my_disciplines", constructorFlag(p.MyDisciplines)).
		set("dpo", constructorFlag(p.DPO)).set("query", p.Query)
	return call[*ConstructorProgramList](ctx, s.c, get("api/constructor/programs/expertise/list", qv))
}

// SignTasks lists programme documents the user can sign. Staff only.
// GET /api/constructor/programs/tasks_list
func (s *ConstructorService) SignTasks(ctx context.Context, p ConstructorSignTaskParams) ([]ConstructorSignTask, error) {
	qv := q().set("query", p.Query).set("education_level_id", p.EducationLevelID).
		set("implementer_id", p.ImplementerID).set("language_id", p.LanguageID)
	res, err := call[struct {
		Rows []ConstructorSignTask `json:"rows"`
	}](ctx, s.c, get("api/constructor/programs/tasks_list", qv))
	return res.Rows, err
}

// Changes returns the change history of all programmes; programTypeID and
// query are optional filters. Staff only.
// GET /api/constructor/programs/changes
func (s *ConstructorService) Changes(ctx context.Context, limit, offset int, programTypeID *int64, query string) (*ConstructorChanges, error) {
	qv := q().set("limit", limit).set("offset", offset).set("program_type_id", programTypeID).set("query", query)
	return call[*ConstructorChanges](ctx, s.c, get("api/constructor/programs/changes", qv))
}

// ProgramChanges returns the change history of a programme; part optionally
// filters by programme part ([ConstructorMenuItem].ID). Staff only.
// GET /api/constructor/programs/{id}/changes
func (s *ConstructorService) ProgramChanges(ctx context.Context, programID int64, limit, offset int, part *int64) (*ConstructorChanges, error) {
	qv := q().set("limit", limit).set("offset", offset).set("part", part)
	return call[*ConstructorChanges](ctx, s.c, get(constructorProgram(programID)+"/changes", qv))
}

// Status returns the programme header, its tabs and the user's permissions. Staff only.
// GET /api/constructor/programs/{id}/status
func (s *ConstructorService) Status(ctx context.Context, programID int64) (*ConstructorProgramStatus, error) {
	return call[*ConstructorProgramStatus](ctx, s.c, get(constructorProgram(programID)+"/status", nil))
}

// SetStatus performs a workflow transition, including archiving. Staff only.
// POST /api/constructor/programs/{id}/status/{action}
func (s *ConstructorService) SetStatus(ctx context.Context, programID int64, action ConstructorStatusAction) error {
	return exec(ctx, s.c, post(constructorProgram(programID)+"/status/"+id(string(action)), nil))
}

// Copy duplicates a programme with a single part and returns the new id. Staff only.
// POST /api/constructor/programs/{id}/copy
func (s *ConstructorService) Copy(ctx context.Context, programID int64) (int64, error) {
	return call[int64](ctx, s.c, post(constructorProgram(programID)+"/copy", nil))
}

// Document downloads the generated programme document. The caller closes the file. Staff only.
// GET /api/constructor/programs/{id}/document/{format}
func (s *ConstructorService) Document(ctx context.Context, programID int64, format ConstructorDocumentFormat) (*File, error) {
	return download(ctx, s.c, get(constructorProgram(programID)+"/document/"+id(string(format)), nil))
}

// SignedDocument downloads the signed PDF of a signed programme. The caller closes the file. Staff only.
// GET /api/constructor/programs/{id}/signed
func (s *ConstructorService) SignedDocument(ctx context.Context, programID int64) (*File, error) {
	return download(ctx, s.c, get(constructorProgram(programID)+"/signed", nil))
}

// UploadFile attaches a file to a programme on signing. Staff only.
// POST /api/constructor/programs/{id}/files
func (s *ConstructorService) UploadFile(ctx context.Context, programID int64, file Upload) error {
	file.Field = "file"
	return exec(ctx, s.c, multipart(http.MethodPost, constructorProgram(programID)+"/files", nil, file))
}

// ProgramJSON returns the JSON dump of a work programme; the individual plan
// uses it to show a discipline annotation.
// GET /api/constructor/programs/{id}/json
func (s *ConstructorService) ProgramJSON(ctx context.Context, programID int64) (*ConstructorProgramJSON, error) {
	return call[*ConstructorProgramJSON](ctx, s.c, get(constructorProgram(programID)+"/json", nil))
}

// AddDeveloper adds an editor, usually a [ConstructorService.SearchPeople] result. Staff only.
// POST /api/constructor/programs/{id}/developers
func (s *ConstructorService) AddDeveloper(ctx context.Context, programID int64, person ConstructorPersonRef) error {
	return exec(ctx, s.c, post(constructorProgram(programID)+"/developers", person))
}

// RemoveDeveloper removes an editor by [ConstructorPerson].ID. Staff only.
// DELETE /api/constructor/programs/{id}/developers/{developerId}
func (s *ConstructorService) RemoveDeveloper(ctx context.Context, programID, developerID int64) error {
	return exec(ctx, s.c, del(constructorProgram(programID)+"/developers/"+id(developerID), nil))
}

// AddAuthor adds an author; an external author has a nil ISU and a free-text FIO. Staff only.
// POST /api/constructor/programs/{id}/authors
func (s *ConstructorService) AddAuthor(ctx context.Context, programID int64, person ConstructorPersonRef) error {
	return exec(ctx, s.c, post(constructorProgram(programID)+"/authors", person))
}

// RemoveAuthor removes an author by [ConstructorPerson].ID. Staff only.
// DELETE /api/constructor/programs/{id}/authors/{authorId}
func (s *ConstructorService) RemoveAuthor(ctx context.Context, programID, authorID int64) error {
	return exec(ctx, s.c, del(constructorProgram(programID)+"/authors/"+id(authorID), nil))
}

// AddExpert assigns one expert by ISU number. Staff only.
// POST /api/constructor/programs/{id}/experts
func (s *ConstructorService) AddExpert(ctx context.Context, programID, isu int64) error {
	body := struct {
		ISU int64 `json:"isu"`
	}{isu}
	return exec(ctx, s.c, post(constructorProgram(programID)+"/experts", body))
}

// SetExperts assigns experts by ISU number; send the whole list,
// including those already assigned. Staff only.
// PATCH /api/constructor/programs/{id}/experts
func (s *ConstructorService) SetExperts(ctx context.Context, programID int64, isus []int64) error {
	if isus == nil {
		isus = []int64{}
	}
	return exec(ctx, s.c, patch(constructorProgram(programID)+"/experts", isus))
}

// RemoveExpert removes an expert by [ConstructorPerson].ID. Staff only.
// DELETE /api/constructor/programs/{id}/experts/{expertId}
func (s *ConstructorService) RemoveExpert(ctx context.Context, programID, expertID int64) error {
	return exec(ctx, s.c, del(constructorProgram(programID)+"/experts/"+id(expertID), nil))
}

// Comments returns the expert comments of a programme, one per part. Staff only.
// GET /api/constructor/programs/{id}/comments
func (s *ConstructorService) Comments(ctx context.Context, programID int64) ([]ConstructorComment, error) {
	return call[[]ConstructorComment](ctx, s.c, get(constructorProgram(programID)+"/comments", nil))
}

// AddComment adds an expert comment to a programme part ([ConstructorMenuItem].ID). Staff only.
// POST /api/constructor/programs/{id}/comments
func (s *ConstructorService) AddComment(ctx context.Context, programID, partID int64, comment string) error {
	body := struct {
		ProgramPartID int64  `json:"program_part_id"`
		Comment       string `json:"comment"`
	}{partID, comment}
	return exec(ctx, s.c, post(constructorProgram(programID)+"/comments", body))
}

// UpdateComment edits an expert comment. Staff only.
// PATCH /api/constructor/programs/{id}/comments/{commentId}
func (s *ConstructorService) UpdateComment(ctx context.Context, programID, commentID int64, comment string) error {
	body := struct {
		Comment string `json:"comment"`
	}{comment}
	return exec(ctx, s.c, patch(constructorProgram(programID)+"/comments/"+id(commentID), body))
}

// DeleteComment deletes an expert comment. Staff only.
// DELETE /api/constructor/programs/{id}/comments/{commentId}
func (s *ConstructorService) DeleteComment(ctx context.Context, programID, commentID int64) error {
	return exec(ctx, s.c, del(constructorProgram(programID)+"/comments/"+id(commentID), nil))
}

// SearchPeople finds people by name or ISU number for the people pickers. Staff only.
// GET /api/constructor/people/search/{query}
func (s *ConstructorService) SearchPeople(ctx context.Context, query string) ([]ConstructorPersonRef, error) {
	return call[[]ConstructorPersonRef](ctx, s.c, get("api/constructor/people/search/"+id(query), nil))
}

// Role returns the user's role level in the constructor. Lower is more
// privileged: below 3 manages experts, 2 is a super expert, below 4 sees the
// expertise tab. Staff only.
// GET /api/constructor/people/roles
func (s *ConstructorService) Role(ctx context.Context) (int, error) {
	return call[int](ctx, s.c, get("api/constructor/people/roles", nil))
}

func (s *ConstructorService) idNames(ctx context.Context, path string) ([]ConstructorIDName, error) {
	return call[[]ConstructorIDName](ctx, s.c, get("api/constructor/references/"+path, nil))
}

// Statuses returns the programme statuses. Staff only.
// GET /api/constructor/references/statuses
func (s *ConstructorService) Statuses(ctx context.Context) ([]ConstructorStatus, error) {
	return call[[]ConstructorStatus](ctx, s.c, get("api/constructor/references/statuses", nil))
}

// ProgramTypes returns the programme types (РПД, РПП, ГИА). Staff only.
// GET /api/constructor/references/program_types
func (s *ConstructorService) ProgramTypes(ctx context.Context) ([]ConstructorIDName, error) {
	return s.idNames(ctx, "program_types")
}

// Languages returns the implementation languages. Staff only.
// GET /api/constructor/references/languages
func (s *ConstructorService) Languages(ctx context.Context) ([]ConstructorIDName, error) {
	return s.idNames(ctx, "languages")
}

// Implementers returns the implementing departments. Staff only.
// GET /api/constructor/references/implementers
func (s *ConstructorService) Implementers(ctx context.Context) ([]ConstructorIDName, error) {
	return s.idNames(ctx, "implementers")
}

// Formats returns the implementation formats. Staff only.
// GET /api/constructor/references/formats
func (s *ConstructorService) Formats(ctx context.Context) ([]ConstructorFormat, error) {
	return call[[]ConstructorFormat](ctx, s.c, get("api/constructor/references/formats", nil))
}

// IntermediateAssessmentTypes returns the forms of intermediate control,
// including course work and course project. Staff only.
// GET /api/constructor/references/intermediate_assessments
func (s *ConstructorService) IntermediateAssessmentTypes(ctx context.Context) ([]ConstructorIntermediateAssessmentType, error) {
	return call[[]ConstructorIntermediateAssessmentType](ctx, s.c, get("api/constructor/references/intermediate_assessments", nil))
}

// WorkTypes returns the work (class) types. Staff only.
// GET /api/constructor/references/work_types
func (s *ConstructorService) WorkTypes(ctx context.Context) ([]ConstructorIDName, error) {
	return s.idNames(ctx, "work_types")
}

// EducationLevels returns the education levels. Staff only.
// GET /api/constructor/references/education_levels
func (s *ConstructorService) EducationLevels(ctx context.Context) ([]ConstructorIDName, error) {
	return s.idNames(ctx, "education_levels")
}

// CurrentAssessmentTypes returns the types of current-assessment tools. Staff only.
// GET /api/constructor/references/current_assessments
func (s *ConstructorService) CurrentAssessmentTypes(ctx context.Context) ([]ConstructorIDName, error) {
	return s.idNames(ctx, "current_assessments")
}

// MaxContents returns the maximum number of parts a programme may have. Staff only.
// GET /api/constructor/references/max_contents
func (s *ConstructorService) MaxContents(ctx context.Context) (int, error) {
	return call[int](ctx, s.c, get("api/constructor/references/max_contents", nil))
}

// FinAttrs returns the financial implementers. Staff only.
// GET /api/constructor/references/fin_attrs
func (s *ConstructorService) FinAttrs(ctx context.Context) ([]ConstructorIDName, error) {
	return s.idNames(ctx, "fin_attrs")
}

// FinAttrImplementers returns the departments allowed for a financial implementer. Staff only.
// GET /api/constructor/references/fin_attrs/{finAttrId}/implementers
func (s *ConstructorService) FinAttrImplementers(ctx context.Context, finAttrID int64) ([]ConstructorIDName, error) {
	return s.idNames(ctx, "fin_attrs/"+id(finAttrID)+"/implementers")
}

// CertificationAssessmentTypes returns the ГИА assessment-tool types. Staff only.
// GET /api/constructor/references/certification_assessment_types
func (s *ConstructorService) CertificationAssessmentTypes(ctx context.Context) ([]ConstructorIDName, error) {
	return s.idNames(ctx, "certification_assessment_types")
}

// PracticeForms returns the practice kinds. Staff only.
// GET /api/constructor/references/practice_forms
func (s *ConstructorService) PracticeForms(ctx context.Context) ([]ConstructorIDName, error) {
	return s.idNames(ctx, "practice_forms")
}

// PracticePassTypes returns the practice pass methods. Staff only.
// GET /api/constructor/references/practice_pass_types
func (s *ConstructorService) PracticePassTypes(ctx context.Context) ([]ConstructorIDName, error) {
	return s.idNames(ctx, "practice_pass_types")
}

// PracticeFormats returns the practice formats. Staff only.
// GET /api/constructor/references/practice_formats
func (s *ConstructorService) PracticeFormats(ctx context.Context) ([]ConstructorIDName, error) {
	return s.idNames(ctx, "practice_formats")
}
