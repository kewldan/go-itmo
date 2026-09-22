package myitmo

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// GIAReviewRole selects the review of [GIAService.ReviewFile].
type GIAReviewRole = string

// Review roles.
const (
	GIAReviewSupervisor GIAReviewRole = "supervisor"
	GIAReviewReviewer   GIAReviewRole = "reviewer"
)

// GIAReviewerRef is a reviewer appointed to a thesis.
type GIAReviewerRef struct {
	ReviewerID         int64  `json:"reviewer_id"`
	ReviewerSurname    string `json:"reviewer_surname"`
	ReviewerName       string `json:"reviewer_name"`
	ReviewerSecondName string `json:"reviewer_second_name"`
}

// GIAStageStatus is the status of a thesis stage or review in the monitoring list.
type GIAStageStatus struct {
	// Stage is set for stage statuses, Reviewer (a role) for review statuses.
	Stage      GIAStage `json:"stage,omitzero"`
	Reviewer   string   `json:"reviewer,omitzero"`
	StatusID   int      `json:"status_id"`
	StatusName string   `json:"status_name"`
	// Extra holds the other status fields, whose shape is not known.
	Extra RawJSON `json:",embed"`
}

// GIAMonitoringRow is a thesis of the monitoring list.
type GIAMonitoringRow struct {
	DiplomaID         int64  `json:"diploma_id"`
	StudentISU        int64  `json:"student_isu"`
	StudentSurname    string `json:"student_surname"`
	StudentName       string `json:"student_name"`
	StudentSecondName string `json:"student_second_name"`
	// StudentStatus contains "академ" or "отчисл" for students on leave or expelled.
	StudentStatus            string           `json:"student_status"`
	GroupID                  string           `json:"group_id"`
	EPID                     int64            `json:"ep_id"`
	EPName                   string           `json:"ep_name"`
	EPEnrollmentYear         int              `json:"ep_enrollment_year"`
	EPEducationLevelName     string           `json:"ep_education_level_name"`
	EPFacultyShortName       string           `json:"ep_faculty_short_name"`
	EPMegafacultyShortName   string           `json:"ep_megafaculty_short_name"`
	DirCode                  string           `json:"dir_code"`
	DirName                  string           `json:"dir_name"`
	EPManagerSurname         string           `json:"ep_manager_surname"`
	EPManagerName            string           `json:"ep_manager_name"`
	EPManagerSecondName      string           `json:"ep_manager_second_name"`
	SupervisorISU            int64            `json:"supervisor_isu"`
	SupervisorSurname        string           `json:"supervisor_surname"`
	SupervisorName           string           `json:"supervisor_name"`
	SupervisorSecondName     string           `json:"supervisor_second_name"`
	SecretarySurname         string           `json:"secretary_surname"`
	SecretaryName            string           `json:"secretary_name"`
	SecretarySecondName      string           `json:"secretary_second_name"`
	HasGeneralStateSecretary bool             `json:"has_general_state_secretary"`
	DefenseDate              time.Time        `json:"defense_date"`
	Reviewers                []GIAReviewerRef `json:"reviewers"`
	StageStatuses            []GIAStageStatus `json:"stage_statuses"`
	ReviewStatuses           []GIAStageStatus `json:"review_statuses"`
}

// GIAMonitoringList is a page of the monitoring list.
type GIAMonitoringList struct {
	Result []GIAMonitoringRow `json:"result"`
	Count  int                `json:"count"`
}

// GIAStageRow is a thesis of a stage list.
type GIAStageRow struct {
	DiplomaID            int64  `json:"diploma_id"`
	DiplomaTheme         string `json:"diploma_theme"`
	StudentISU           int64  `json:"student_isu"`
	StudentSurname       string `json:"student_surname"`
	StudentName          string `json:"student_name"`
	StudentSecondName    string `json:"student_second_name"`
	StudentStatus        string `json:"student_status"`
	GroupID              string `json:"group_id"`
	EPID                 int64  `json:"ep_id"`
	EPName               string `json:"ep_name"`
	EPEnrollmentYear     int    `json:"ep_enrollment_year"`
	EPEducationLevelName string `json:"ep_education_level_name"`
	EPFacultyShortName   string `json:"ep_faculty_short_name"`
	DirCode              string `json:"dir_code"`
	DirName              string `json:"dir_name"`
	EPManagerSurname     string `json:"ep_manager_surname"`
	EPManagerName        string `json:"ep_manager_name"`
	EPManagerSecondName  string `json:"ep_manager_second_name"`
	SupervisorISU        int64  `json:"supervisor_isu"`
	SupervisorSurname    string `json:"supervisor_surname"`
	SupervisorName       string `json:"supervisor_name"`
	SupervisorSecondName string `json:"supervisor_second_name"`
	SupervisorDegree     string `json:"supervisor_degree"`
	SupervisorJobTitle   string `json:"supervisor_job_title"`
	StatusID             int    `json:"status_id"`
	StatusName           string `json:"status_name"`
	// The file stage adds the originality, the final file and the review statuses.
	Originality            float64    `json:"originality"`
	ResultFileKey          string     `json:"result_file_key"`
	SupervisorReviewStatus *GIAStatus `json:"supervisor_review_status"`
	ReviewerReviewStatus   *GIAStatus `json:"reviewer_review_status"`
}

// GIAStageList is a page of a stage list.
type GIAStageList struct {
	Result []GIAStageRow `json:"result"`
	Count  int           `json:"count"`
}

// GIAMonitoringFilters are the options of the monitoring list filters.
type GIAMonitoringFilters struct {
	RoleFilters             []GIARoleOption     `json:"role_filters"`
	FacultyFilters          []IDValue           `json:"faculty_filters"`
	EduProgramFilters       []GIAIDName         `json:"edu_program_filters"`
	EduDirectionsFilters    []GIAIDName         `json:"edu_directions_filters"`
	GroupFilters            []GIAGroupOption    `json:"group_filters"`
	CompressedStatusFilters []GIAApprovedOption `json:"compressed_status_filters"`
	DefenseDateFilters      []GIADateOption     `json:"defense_date_filters"`
}

// GIAStageFilters are the options of the stage list filters.
type GIAStageFilters struct {
	RoleFilters          []GIARoleOption  `json:"role_filters"`
	FacultyFilters       []IDValue        `json:"faculty_filters"`
	EduProgramFilters    []GIAIDName      `json:"edu_program_filters"`
	EduDirectionsFilters []GIAIDName      `json:"edu_directions_filters"`
	GroupFilters         []GIAGroupOption `json:"group_filters"`
	StatusFilters        []GIAIDName      `json:"status_filters"`
}

// GIAStageCount is the number of theses awaiting the user's approval at a stage.
type GIAStageCount struct {
	Stage GIAStage `json:"stage"`
	Count int      `json:"count"`
}

// GIAStageComment is the last comment on a thesis stage.
type GIAStageComment struct {
	Comment       string `json:"comment"`
	CommenterFIO  string `json:"commenter_fio"`
	CommenterRole string `json:"commenter_role"`
}

// GIADiplomaMainInfo is the student and programme block of a thesis card.
type GIADiplomaMainInfo struct {
	StudentISU              int64  `json:"student_isu"`
	StudentSurname          string `json:"student_surname"`
	StudentName             string `json:"student_name"`
	StudentSecondName       string `json:"student_second_name"`
	StudentGroup            string `json:"student_group"`
	ThemeRU                 string `json:"theme_ru"`
	DirCode                 string `json:"dir_code"`
	DirName                 string `json:"dir_name"`
	EPName                  string `json:"ep_name"`
	EPFaculty               string `json:"ep_faculty"`
	EPDiplomaYear           int    `json:"ep_diploma_year"`
	EPEnrollmentYear        int    `json:"ep_enrollment_year"`
	EPEducationLevelName    string `json:"ep_education_level_name"`
	EPEducationLanguageName string `json:"ep_education_language_name"`
	EPManagerISU            int64  `json:"ep_manager_isu"`
	FacultyManagerISU       int64  `json:"faculty_manager_isu"`
	SecretaryISU            int64  `json:"secretary_isu"`
	SecretarySurname        string `json:"secretary_surname"`
	SecretaryName           string `json:"secretary_name"`
	SecretarySecondName     string `json:"secretary_second_name"`
	SupervisorISU           int64  `json:"supervisor_isu"`
	SupervisorSurname       string `json:"supervisor_surname"`
	SupervisorName          string `json:"supervisor_name"`
	SupervisorSecondName    string `json:"supervisor_second_name"`
}

// GIADiplomaApplication is the application stage (заявление) of a thesis.
type GIADiplomaApplication struct {
	StatusID             int         `json:"status_id"`
	StatusName           string      `json:"status_name"`
	ThemeRU              string      `json:"theme_ru"`
	ThemeEN              string      `json:"theme_en"`
	Justification        string      `json:"justification"`
	DiplomaTypeName      string      `json:"diploma_type_name"`
	ExternalPartnerName  string      `json:"external_partner_name"`
	SupervisorISU        int64       `json:"supervisor_isu"`
	SupervisorSurname    string      `json:"supervisor_surname"`
	SupervisorName       string      `json:"supervisor_name"`
	SupervisorSecondName string      `json:"supervisor_second_name"`
	SupervisorJob        string      `json:"supervisor_job"`
	SupervisorJobTitle   string      `json:"supervisor_job_title"`
	CoSupervisors        []GIAPerson `json:"co_supervisors"`
	Consultants          []GIAPerson `json:"consultants"`
	GIAStageComment
}

// GIADiplomaTask is the task stage (задание) of a thesis.
type GIADiplomaTask struct {
	StatusID      int       `json:"status_id"`
	StatusName    string    `json:"status_name"`
	IssuedAt      time.Time `json:"issued_at"`
	Format        string    `json:"format"`
	MainQuestions string    `json:"main_questions"`
	ExtraInfo     string    `json:"extra_info"`
	GIAStageComment
}

// GIADiplomaAnnotation is the annotation stage of a thesis.
type GIADiplomaAnnotation struct {
	StatusID     int                    `json:"status_id"`
	StatusName   string                 `json:"status_name"`
	Goal         string                 `json:"goal"`
	Tasks        string                 `json:"tasks"`
	Results      string                 `json:"results"`
	ExtraInfo    string                 `json:"extra_info"`
	Grants       []GIAAnnotationGrant   `json:"grants"`
	Publications []GIAAnnotationPublish `json:"publications"`
	Speeches     []GIAAnnotationSpeech  `json:"speeches"`
	GIAStageComment
}

// GIAAnnotationGrant is a grant listed in the annotation.
type GIAAnnotationGrant struct {
	Name string `json:"name"`
	// Year is a number or a string.
	Year RawJSON `json:"year,omitzero"`
}

// GIAAnnotationPublish is a publication listed in the annotation.
type GIAAnnotationPublish struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	Type   string `json:"type"`
}

// GIAAnnotationSpeech is a conference talk listed in the annotation.
type GIAAnnotationSpeech struct {
	Name string `json:"name"`
}

// GIADiplomaFileStage is the final-file stage of a thesis with the
// anti-plagiarism results.
type GIADiplomaFileStage struct {
	StatusID   int    `json:"status_id"`
	StatusName string `json:"status_name"`
	// StudentFileKey and ResultFileKey download with [GIAService.DiplomaFile].
	StudentFileKey          string  `json:"student_file_key"`
	ResultFileKey           string  `json:"result_file_key"`
	ReportFileName          string  `json:"report_file_name"`
	Originality             float64 `json:"originality"`
	Citations               float64 `json:"citations"`
	Similarity              float64 `json:"similarity"`
	SelfCitations           float64 `json:"self_citations"`
	AntiplagiatReportWebURL string  `json:"antiplagiat_report_web_url"`
	GIAStageComment
}

// GIAReview is a supervisor or reviewer review of a thesis.
type GIAReview struct {
	Approved bool `json:"approved"`
	// ReviewerID is set on reviewer reviews.
	ReviewerID    int64  `json:"reviewer_id"`
	Advantages    string `json:"advantages"`
	Disadvantages string `json:"disadvantages"`
	Comment       string `json:"comment"`
	// Assessment is a final assessment [GIAGrade] id.
	Assessment             int64  `json:"assessment"`
	AssessmentName         string `json:"assessment_name"`
	ThesisComplete         bool   `json:"thesis_complete"`
	AwardingQualifications bool   `json:"awarding_qualifications"`
	// Signatures have an unknown shape.
	SupervisorSignature RawJSON `json:"supervisor_signature,omitzero"`
	ReviewerSignature   RawJSON `json:"reviewer_signature,omitzero"`
	StudentSignature    RawJSON `json:"student_signature,omitzero"`
	// Criteria holds the other members: the criterion grades (the keys of
	// [GIASupervisorReview] or [GIAReviewerReview], [GIAGrade] ids) and their
	// names under "<key>_name".
	Criteria RawJSON `json:",embed"`
}

// GIAReviewerQuestion is a question of a reviewer for the defense.
type GIAReviewerQuestion struct {
	ID         int64  `json:"id"`
	Question   string `json:"question"`
	ReviewerID int64  `json:"reviewer_id"`
}

// GIADiploma is the full thesis card with all stages and reviews.
type GIADiploma struct {
	DiplomaID               int64                 `json:"diploma_id"`
	RequesterReviewerID     *int64                `json:"requester_reviewer_id"`
	MainInfo                GIADiplomaMainInfo    `json:"main_info"`
	Application             GIADiplomaApplication `json:"application"`
	Task                    GIADiplomaTask        `json:"task"`
	Annotation              GIADiplomaAnnotation  `json:"annotation"`
	File                    GIADiplomaFileStage   `json:"file"`
	SupervisorReview        *GIAReview            `json:"supervisor_review"`
	ReviewerReviews         []GIAReview           `json:"reviewer_reviews"`
	ReviewerReviewQuestions []GIAReviewerQuestion `json:"reviewer_review_questions"`
}

// GIADiplomaHistoryItem is a status change or comment of a thesis stage.
type GIADiplomaHistoryItem struct {
	StatusID      int       `json:"status_id"`
	Comment       string    `json:"comment"`
	CommenterFIO  string    `json:"commenter_fio"`
	CommenterISU  int64     `json:"commenter_isu"`
	CommenterRole string    `json:"commenter_role"`
	CreatedAt     time.Time `json:"created_at"`
}

// GIAAntiplagiatStatus is the state of the anti-plagiarism check of a thesis.
type GIAAntiplagiatStatus struct {
	Status         string  `json:"status"`
	IsReady        bool    `json:"is_ready"`
	IsFailed       bool    `json:"is_failed"`
	FailDetails    string  `json:"fail_details"`
	Originality    float64 `json:"originality"`
	Similarity     float64 `json:"similarity"`
	Citations      float64 `json:"citations"`
	SelfCitations  float64 `json:"self_citations"`
	ReportFileName string  `json:"report_file_name"`
	ReportWebURL   string  `json:"report_web_url"`
}

// GIAAntiplagiatInfo is the manually entered anti-plagiarism result.
type GIAAntiplagiatInfo struct {
	Originality             float64
	Citations               float64
	Similarity              float64
	SelfCitations           float64
	AntiplagiatReportWebURL string
}

// GIASupervisorReview is the supervisor's review. Criterion fields are
// criterion [GIAGrade] ids; Assessment is a final assessment grade id.
// Approved false revokes a submitted review.
type GIASupervisorReview struct {
	Approved               bool   `json:"approved"`
	InitiativeGoals        int64  `json:"initiative_goals"`
	InitiativeMethods      int64  `json:"initiative_methods"`
	QualityLogic           int64  `json:"quality_logic"`
	JustificationRelevance int64  `json:"justification_relevance"`
	Consistency            int64  `json:"consistency"`
	Involvement            int64  `json:"involvement"`
	Preparedness           int64  `json:"preparedness"`
	Approbation            int64  `json:"approbation"`
	Publication            int64  `json:"publication"`
	PersonalInvolvement    int64  `json:"personal_involvement"`
	Integration            int64  `json:"integration"`
	Advantages             string `json:"advantages"`
	Disadvantages          string `json:"disadvantages"`
	Comment                string `json:"comment"`
	Assessment             int64  `json:"assessment"`
	ThesisComplete         bool   `json:"thesis_complete"`
	AwardingQualifications bool   `json:"awarding_qualifications"`
}

// GIAReviewerReview is the external reviewer's review. Criterion fields are
// criterion [GIAGrade] ids; Assessment is a final assessment grade id.
// Approved false revokes a submitted review.
type GIAReviewerReview struct {
	Approved                bool   `json:"approved"`
	RelevantContent         int64  `json:"relevant_content"`
	JustificationRelevance  int64  `json:"justification_relevance"`
	RelevantEdu             int64  `json:"relevant_edu"`
	CorrectMethods          int64  `json:"correct_methods"`
	QualityLogic            int64  `json:"quality_logic"`
	JustificationAssertions int64  `json:"justification_assertions"`
	Value                   int64  `json:"value"`
	Integration             int64  `json:"integration"`
	Advantages              string `json:"advantages"`
	Disadvantages           string `json:"disadvantages"`
	Comment                 string `json:"comment"`
	Assessment              int64  `json:"assessment"`
	ThesisComplete          bool   `json:"thesis_complete"`
	AwardingQualifications  bool   `json:"awarding_qualifications"`
}

// GIAPrintItem is a diploma or supplement of the print list.
type GIAPrintItem struct {
	ID                int64     `json:"id"`
	DiplomaID         int64     `json:"diploma_id"`
	StudentFIO        string    `json:"student_fio"`
	StudentISU        int64     `json:"student_isu"`
	GroupID           string    `json:"group_id"`
	MegafacultyName   string    `json:"megafaculty_name"`
	ImplementerName   string    `json:"implementer_name"`
	QualificationName string    `json:"qualification_name"`
	DefenseDate       time.Time `json:"defense_date"`
	ProtocolNumber    string    `json:"protocol_number"`
	DocSeries         string    `json:"doc_series"`
	// DocNumber is a number or a string.
	DocNumber  RawJSON `json:"doc_number,omitzero"`
	WithHonors bool    `json:"with_honors"`
	SignerFIO  string  `json:"signer_fio"`
	SignerISU  int64   `json:"signer_isu"`
	StatusID   int     `json:"status_id"`
	StatusName string  `json:"status_name"`
}

// GIAPrintUnit is a megafaculty or implementing unit of the print filters.
type GIAPrintUnit struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	ShortName string `json:"short_name"`
}

// GIAPrintFilters are the options of the print list filters.
type GIAPrintFilters struct {
	Megafaculties   []GIAPrintUnit `json:"megafaculties"`
	Implementers    []GIAPrintUnit `json:"implementers"`
	EducationLevels []GIAIDName    `json:"education_levels"`
	// Groups has an unknown item shape.
	Groups   RawJSON     `json:"groups,omitzero"`
	Statuses []GIAIDName `json:"statuses"`
}

// GIAPrintList is a page of the print list with its filters.
type GIAPrintList struct {
	Array   []GIAPrintItem  `json:"array"`
	Count   int             `json:"count"`
	Filters GIAPrintFilters `json:"filters"`
}

// GIAMonitoringParams filters [GIAService.Monitoring]. Zero values are not
// sent, except Limit, Offset and HideDeducted.
type GIAMonitoringParams struct {
	Limit, Offset int
	Year          int
	HideDeducted  bool
	ShowAs        GIARole
	Faculty       int64
	GroupID       string
	EduDirection  int64
	EduProgram    int64
	// ApprovalNeeded and Approved are sent when set.
	ApprovalNeeded *bool
	Approved       *bool
	DefenseDate    string
	// SortBy and SortOrder ("asc"/"desc") are sent comma-separated.
	SortBy    []string
	SortOrder []string
	// Type defaults to "monitoring".
	Type  string
	Query string
}

// GIAStageMonitoringParams filters [GIAService.StageMonitoring]. Zero values
// are not sent, except Stage, Limit, Offset and HideDeducted.
type GIAStageMonitoringParams struct {
	Stage          GIAStage
	Limit, Offset  int
	Year           int
	StatusID       int
	ShowAs         GIARole
	HideDeducted   bool
	GroupID        string
	EduProgram     int64
	EduDirection   int64
	SortBy         []string
	SortOrder      []string
	ApprovalNeeded *bool
	Query          string
}

// GIAPrintParams filters [GIAService.PrintList]. Zero values are not sent,
// except DocType, Limit and Offset.
type GIAPrintParams struct {
	DocType          GIADocType
	Limit, Offset    int
	Year             int
	MegafacultyID    int64
	ImplementerID    int64
	EducationLevelID int64
	GroupID          string
	WithHonors       *bool
	StatusID         int
	Query            string
}

// GIAPrintPreviewParams sets the rendering of [GIAService.PrintPreview].
type GIAPrintPreviewParams struct {
	DocType  GIADocType
	FontSize float64
	// AttachmentSecondFontSize is sent for supplements only.
	AttachmentSecondFontSize float64
}

func giaDiploma(diplomaID int64) string { return "api/gia/diplomas/" + id(diplomaID) }

// Monitoring returns a page of the thesis monitoring list.
// Staff only.
// GET /api/gia/diplomas/monitoring
func (s *GIAService) Monitoring(ctx context.Context, p GIAMonitoringParams) (*GIAMonitoringList, error) {
	if p.Type == "" {
		p.Type = "monitoring"
	}
	v := q().set("limit", p.Limit).set("offset", p.Offset)
	giaSet(v, "year", p.Year)
	v.set("hide_deducted", p.HideDeducted).set("show_as", p.ShowAs)
	giaSet(v, "faculty", p.Faculty)
	v.set("group_id", p.GroupID)
	giaSet(v, "edu_direction", p.EduDirection)
	giaSet(v, "edu_program", p.EduProgram)
	v.set("approval_needed", p.ApprovalNeeded).set("approved", p.Approved).set("defense_date", p.DefenseDate)
	v.set("sort_by", strings.Join(p.SortBy, ",")).set("sort_order", strings.Join(p.SortOrder, ","))
	v.set("type", p.Type).set("query", p.Query)
	return call[*GIAMonitoringList](ctx, s.c, get("api/gia/diplomas/monitoring", v))
}

// MonitoringFilters returns the options of the monitoring list filters; showAs is optional.
// Staff only.
// GET /api/gia/diplomas/monitoring/filters
func (s *GIAService) MonitoringFilters(ctx context.Context, showAs GIARole) (*GIAMonitoringFilters, error) {
	return call[*GIAMonitoringFilters](ctx, s.c, get("api/gia/diplomas/monitoring/filters", q().set("show_as", showAs)))
}

// StageMonitoring returns a page of theses at an approval stage.
// Staff only.
// GET /api/gia/diplomas/monitoring/stage
func (s *GIAService) StageMonitoring(ctx context.Context, p GIAStageMonitoringParams) (*GIAStageList, error) {
	v := q().set("stage", p.Stage).set("limit", p.Limit).set("offset", p.Offset)
	giaSet(v, "year", p.Year)
	giaSet(v, "status_id", p.StatusID)
	v.set("show_as", p.ShowAs).set("hide_deducted", p.HideDeducted).set("group_id", p.GroupID)
	giaSet(v, "edu_program", p.EduProgram)
	giaSet(v, "edu_direction", p.EduDirection)
	v.set("sort_by", strings.Join(p.SortBy, ",")).set("sort_order", strings.Join(p.SortOrder, ","))
	v.set("approval_needed", p.ApprovalNeeded).set("query", p.Query)
	return call[*GIAStageList](ctx, s.c, get("api/gia/diplomas/monitoring/stage", v))
}

// StageFilters returns the options of the stage list filters; both arguments are optional.
// Staff only.
// GET /api/gia/diplomas/monitoring/stage/filters
func (s *GIAService) StageFilters(ctx context.Context, stage GIAStage, showAs GIARole) (*GIAStageFilters, error) {
	return call[*GIAStageFilters](ctx, s.c, get("api/gia/diplomas/monitoring/stage/filters", q().set("stage", stage).set("show_as", showAs)))
}

// PendingApprovalCounts returns the number of theses awaiting the user's approval per stage.
// Staff only.
// GET /api/gia/diplomas/stage/onApproval/count
func (s *GIAService) PendingApprovalCounts(ctx context.Context) ([]GIAStageCount, error) {
	return call[[]GIAStageCount](ctx, s.c, get("api/gia/diplomas/stage/onApproval/count", nil))
}

// SupervisorReviewCount returns the number of reviews the current supervisor still has to write.
// Staff only.
// GET /api/gia/diplomas/review/supervisor/count
func (s *GIAService) SupervisorReviewCount(ctx context.Context) (int, error) {
	res, err := call[struct {
		Count int `json:"count"`
	}](ctx, s.c, get("api/gia/diplomas/review/supervisor/count", nil))
	return res.Count, err
}

// ApproveStages approves a stage (application, task or annotation) of several theses at once.
// Staff only.
// PUT /api/gia/diplomas/statuses
func (s *GIAService) ApproveStages(ctx context.Context, stage GIAStage, diplomaIDs ...int64) error {
	body := struct {
		Stage      GIAStage `json:"stage"`
		DiplomaIDs []int64  `json:"diploma_ids"`
	}{stage, diplomaIDs}
	if body.DiplomaIDs == nil {
		body.DiplomaIDs = []int64{}
	}
	return exec(ctx, s.c, put("api/gia/diplomas/statuses", body))
}

// Diploma returns the full thesis card.
// Staff only.
// GET /api/gia/diplomas/{diploma_id}/student
func (s *GIAService) Diploma(ctx context.Context, diplomaID int64) (*GIADiploma, error) {
	return call[*GIADiploma](ctx, s.c, get(giaDiploma(diplomaID)+"/student", nil))
}

// DiplomaHistory returns the status and comment history of a thesis stage.
// Staff only.
// GET /api/gia/diplomas/{diploma_id}/history
func (s *GIAService) DiplomaHistory(ctx context.Context, diplomaID int64, stage GIAStage) ([]GIADiplomaHistoryItem, error) {
	return call[[]GIADiplomaHistoryItem](ctx, s.c, get(giaDiploma(diplomaID)+"/history", q().set("stage", stage)))
}

// AntiplagiatStatus returns the state of the anti-plagiarism check.
// Staff only.
// GET /api/gia/diplomas/{diploma_id}/antiplagiat/status
func (s *GIAService) AntiplagiatStatus(ctx context.Context, diplomaID int64) (*GIAAntiplagiatStatus, error) {
	return call[*GIAAntiplagiatStatus](ctx, s.c, get(giaDiploma(diplomaID)+"/antiplagiat/status", nil))
}

// SetAntiplagiatInfo enters the anti-plagiarism results manually and uploads
// the report (sent as the "report" part; its Field is overwritten).
// Staff only.
// PUT /api/gia/diplomas/{diploma_id}/antiplagiat/info
func (s *GIAService) SetAntiplagiatInfo(ctx context.Context, diplomaID int64, info GIAAntiplagiatInfo, report Upload) error {
	num := func(f float64) string { return strconv.FormatFloat(f, 'f', -1, 64) }
	fields := map[string]string{
		"originality":                num(info.Originality),
		"citations":                  num(info.Citations),
		"similarity":                 num(info.Similarity),
		"self_citations":             num(info.SelfCitations),
		"antiplagiat_report_web_url": info.AntiplagiatReportWebURL,
	}
	report.Field = "report"
	return exec(ctx, s.c, multipart(http.MethodPut, giaDiploma(diplomaID)+"/antiplagiat/info", fields, report))
}

// SendToAntiplagiat sends the thesis file to the anti-plagiarism system; the
// check takes about ten minutes.
// Staff only.
// POST /api/gia/diplomas/{diploma_id}/antiplagiat/send
func (s *GIAService) SendToAntiplagiat(ctx context.Context, diplomaID int64) error {
	return exec(ctx, s.c, post(giaDiploma(diplomaID)+"/antiplagiat/send", nil))
}

// AntiplagiatReport downloads the anti-plagiarism report.
// Staff only.
// GET /api/gia/diplomas/{diploma_id}/antiplagiat/verification-report
func (s *GIAService) AntiplagiatReport(ctx context.Context, diplomaID int64) (*File, error) {
	return download(ctx, s.c, get(giaDiploma(diplomaID)+"/antiplagiat/verification-report", nil))
}

// ApproveStage approves a thesis stage acting as the given role: the role
// whose approval the stage awaits (status 2 osop, 8 supervisor, 9
// ep_manager, 10 secretary, 11 faculty_manager or deputy_faculty_manager).
// Staff only.
// PUT /api/gia/diplomas/{diploma_id}/{stage}/status (?status={role})
func (s *GIAService) ApproveStage(ctx context.Context, diplomaID int64, stage GIAStage, as GIARole) error {
	r := put(giaDiploma(diplomaID)+"/"+id(stage)+"/status", struct{}{})
	return exec(ctx, s.c, withQuery(r, q().set("status", as)))
}

// RejectStage returns a thesis stage to the student for rework with a message.
// Staff only.
// PUT /api/gia/diplomas/{diploma_id}/{stage}/status (?status=reject)
func (s *GIAService) RejectStage(ctx context.Context, diplomaID int64, stage GIAStage, message string) error {
	r := put(giaDiploma(diplomaID)+"/"+id(stage)+"/status", giaMessage{message})
	return exec(ctx, s.c, withQuery(r, q().set("status", "reject")))
}

// ApplicationFile downloads the generated application (заявление) PDF.
// Staff only.
// GET /api/gia/diplomas/{diploma_id}/application/file
func (s *GIAService) ApplicationFile(ctx context.Context, diplomaID int64) (*File, error) {
	return download(ctx, s.c, get(giaDiploma(diplomaID)+"/application/file", nil))
}

// DiplomaFile downloads a stored thesis file by its key ([GIAFileRef.Key],
// [GIADiplomaFileStage.StudentFileKey], ...).
// Staff only.
// GET /api/gia/diplomas/{diploma_id}/download/{file_key}
func (s *GIAService) DiplomaFile(ctx context.Context, diplomaID int64, fileKey string) (*File, error) {
	return download(ctx, s.c, get(giaDiploma(diplomaID)+"/download/"+id(fileKey), nil))
}

// SubmitSupervisorReview submits (Approved true) or revokes (Approved false) the supervisor's review.
// Staff only.
// PUT /api/gia/diplomas/{diploma_id}/review/supervisor
func (s *GIAService) SubmitSupervisorReview(ctx context.Context, diplomaID int64, review GIASupervisorReview) error {
	return exec(ctx, s.c, put(giaDiploma(diplomaID)+"/review/supervisor", review))
}

// ReviewFile downloads the generated review PDF of the supervisor or the reviewer.
// Staff only.
// GET /api/gia/diplomas/{diploma_id}/review/{role}/file
func (s *GIAService) ReviewFile(ctx context.Context, diplomaID int64, role GIAReviewRole) (*File, error) {
	return download(ctx, s.c, get(giaDiploma(diplomaID)+"/review/"+id(role)+"/file", nil))
}

// PrintList returns a page of diplomas or supplements to print, with the filter options.
// Staff only.
// GET /api/gia/diplomas/print/list
func (s *GIAService) PrintList(ctx context.Context, p GIAPrintParams) (*GIAPrintList, error) {
	v := q().set("doctype", p.DocType).set("limit", p.Limit).set("offset", p.Offset)
	giaSet(v, "year", p.Year)
	giaSet(v, "megafaculty_id", p.MegafacultyID)
	giaSet(v, "implementer_id", p.ImplementerID)
	giaSet(v, "education_level_id", p.EducationLevelID)
	v.set("group_id", p.GroupID).set("with_honors", p.WithHonors)
	giaSet(v, "status_id", p.StatusID)
	v.set("query", p.Query)
	return call[*GIAPrintList](ctx, s.c, get("api/gia/diplomas/print/list", v))
}

// PrintPreview renders the print PDF of a diploma or supplement.
// Staff only.
// POST /api/gia/diplomas/print/{diploma_id}
func (s *GIAService) PrintPreview(ctx context.Context, diplomaID int64, p GIAPrintPreviewParams) (*File, error) {
	v := giaSetFloat(q(), "font_size", p.FontSize)
	if p.DocType == GIADocTypeAttachment {
		giaSetFloat(v, "attachment_second_font_size", p.AttachmentSecondFontSize)
	}
	v.set("doctype", p.DocType)
	return download(ctx, s.c, withQuery(post("api/gia/diplomas/print/"+id(diplomaID), struct{}{}), v))
}

// SetBlankNumber sets the number of the blank form of a diploma or supplement.
// Staff only.
// PATCH /api/gia/diplomas/print/{diploma_id}/number
func (s *GIAService) SetBlankNumber(ctx context.Context, diplomaID int64, docType GIADocType, number int64) error {
	body := struct {
		Number int64 `json:"number"`
	}{number}
	r := patch("api/gia/diplomas/print/"+id(diplomaID)+"/number", body)
	return exec(ctx, s.c, withQuery(r, q().set("doctype", docType)))
}

// UploadPrintScan uploads the scan of a printed diploma or supplement (sent
// as the "scan" part; its Field is overwritten).
// Staff only.
// POST /api/gia/diplomas/print/{diploma_id}/scan
func (s *GIAService) UploadPrintScan(ctx context.Context, diplomaID int64, docType GIADocType, scan Upload) error {
	scan.Field = "scan"
	r := multipart(http.MethodPost, "api/gia/diplomas/print/"+id(diplomaID)+"/scan", nil, scan)
	return exec(ctx, s.c, withQuery(r, q().set("doctype", docType)))
}

// EUSupplementTemplate downloads the DOCX template of the European supplement;
// diploma_id is the European supplement id.
// Staff only.
// POST /api/gia/diplomas/print/{diploma_id}/eu_attachment
func (s *GIAService) EUSupplementTemplate(ctx context.Context, diplomaID int64) (*File, error) {
	return download(ctx, s.c, post("api/gia/diplomas/print/"+id(diplomaID)+"/eu_attachment", nil))
}
