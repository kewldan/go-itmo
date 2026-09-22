package myitmo

import (
	"context"
	"net/http"
	"strconv"
)

// GIAStudentsService is state final attestation, student side (/api/gia-students):
// the final qualification work (ВКР) workflow, the diploma supplement and the
// European diploma supplement.
//
// A diploma goes through five stages in order: application, task, annotation,
// file and defense. [GIAStudentDiplomaInfo.Stages] lists only the stages that
// are already open. Every call returns the standard envelope; errors carry
// the server's localised message.
type GIAStudentsService struct{ c *Client }

// Stage names, used by [GIAStudentsService.Revoke] and as the entity of
// [GIAStudentsService.Statuses]. The index in [GIAStudentDiplomaInfo.Stages]
// follows the same order.
const (
	GIAStudentStageApplication = "application"
	GIAStudentStageTask        = "task"
	GIAStudentStageAnnotation  = "annotation"
	GIAStudentStageFile        = "file"
	GIAStudentStageDefense     = "defense"
)

// Stage status_id values of [GIAStudentStage] and the stage update requests.
// Human-readable names come from [GIAStudentsService.Statuses].
const (
	// GIAStudentStatusNotStarted: the stage is open but not started. For the
	// application stage it means the agreement must be accepted first; other
	// stages are started with the Start* methods (which send 7).
	GIAStudentStatusNotStarted = 6
	// GIAStudentStatusInWork: draft, editable.
	GIAStudentStatusInWork = 7
	// GIAStudentStatusSentToApproval: sent for approval (application, task, annotation).
	GIAStudentStatusSentToApproval = 8
	// GIAStudentStatusRejected: rejected; the form stays editable and every
	// later stage is blocked.
	GIAStudentStatusRejected = 3
	// GIAStudentStatusApproved: approved or done.
	GIAStudentStatusApproved = 4
	// GIAStudentStatusFileSentToApproval: the file stage uses 10 instead of 8.
	GIAStudentStatusFileSentToApproval = 10
	// GIAStudentStatusDefenseDone and GIAStudentStatusDefenseDoneAlt: the defense stage is done.
	GIAStudentStatusDefenseDone    = 13
	GIAStudentStatusDefenseDoneAlt = 17
)

// Diploma type ids of [GIAStudentDiplomaType] with a known meaning.
const (
	// GIAStudentDiplomaTypeNoJustification needs no topic justification.
	GIAStudentDiplomaTypeNoJustification = 1
	// GIAStudentDiplomaTypeOther requires a custom DiplomaTypeName.
	GIAStudentDiplomaTypeOther = 6
)

// GIAStudentDefenseTypeOnline is the presentation_defense_type_id of an online defense.
const GIAStudentDefenseTypeOnline = 1

// GIAStudentAssessmentAbsent is the assessment_id of a student absent from the
// defense (shown as "Н").
const GIAStudentAssessmentAbsent = 35

// Status of a diploma supplement block ([GIAStudentSupplementStatus]).
const (
	GIAStudentSupplementEditable  = 7
	GIAStudentSupplementConfirmed = 4
)

// Status of the European diploma supplement ([GIAStudentEUSupplement].Status).
const (
	// GIAStudentEUNotSubmitted is not sent by the server: it stands for
	// a 404 answer of [GIAStudentsService.EUSupplement].
	GIAStudentEUNotSubmitted  = -1
	GIAStudentEUSubmitted     = 27
	GIAStudentEUStudentReview = 28
	GIAStudentEUOnReview      = 29
	GIAStudentEURejected      = 3
	GIAStudentEUApproved      = 30
	GIAStudentEUDone          = 4
)

// GIAStudentDiploma is one of the student's diplomas, one per educational programme.
type GIAStudentDiploma struct {
	DiplomaID          int64  `json:"diploma_id"`
	EducationLevelName string `json:"education_level_name"`
	EPName             string `json:"ep_name"`
	// EPEnrollmentYear is a number or a string.
	EPEnrollmentYear RawJSON `json:"ep_enrollment_year,omitzero"`
	FacultyName      string  `json:"faculty_name"`
	EPDirectionCode  string  `json:"ep_direction_code"`
	EPDirectionName  string  `json:"ep_direction_name"`
	EPLanguage       string  `json:"ep_language"`
}

// GIAStudentDiplomaInfo is the state of the diploma workflow.
type GIAStudentDiplomaInfo struct {
	DiplomaID int64 `json:"diploma_id"`
	// Stages are the open stages in the order application, task, annotation,
	// file, defense; a stage is available when its index is present.
	Stages   []GIAStudentStage   `json:"stages"`
	Contacts []GIAStudentContact `json:"contacts"`
	// RPDGIALink is the URL of the GIA programme document.
	RPDGIALink string `json:"rpd_gia_link"`
}

// GIAStudentStage is the status of one workflow stage.
type GIAStudentStage struct {
	// StatusID is one of the GIAStudentStatus* constants.
	StatusID int `json:"status_id"`
	// Deadline is preformatted by the server and shown as is; empty when there is none.
	// For the defense stage it is the defense date deadline.
	Deadline string `json:"deadline"`
}

// GIAStudentContact is a person responsible for the diploma workflow.
type GIAStudentContact struct {
	Role  string `json:"role"`
	FIO   string `json:"fio"`
	ISU   *int64 `json:"isu"`
	Email string `json:"email"`
	Phone string `json:"phone"`
	// PhotoURL is a base URL; append "cover/90/90/" for a thumbnail.
	PhotoURL string `json:"photo_url"`
}

// GIAStudentComment is the reviewer's or approver's comment attached to a stage.
type GIAStudentComment struct {
	Comment           string `json:"comment"`
	CommentAuthorRole string `json:"comment_author_role"`
	CommentAuthorFIO  string `json:"comment_author_fio"`
}

// GIAStudentApplication is the application stage: topic, type and supervisors.
type GIAStudentApplication struct {
	// DiplomaTypeID is a [GIAStudentDiplomaType] id.
	DiplomaTypeID int64 `json:"diploma_type_id"`
	// DiplomaTypeName is the custom type name when DiplomaTypeID is [GIAStudentDiplomaTypeOther].
	DiplomaTypeName            string `json:"diploma_type_name"`
	DiplomaTheme               string `json:"diploma_theme"`
	DiplomaThemeEN             string `json:"diploma_theme_en"`
	DiplomaJustification       string `json:"diploma_justification"`
	DiplomaExternalPartner     bool   `json:"diploma_external_partner"`
	DiplomaExternalPartnerName string `json:"diploma_external_partner_name"`
	// DiplomaFundamental is true for a fundamental area, false for an applied one.
	DiplomaFundamental bool   `json:"diploma_fundamental"`
	SupervisorISU      *int64 `json:"supervisor_isu"`
	// SupervisorFIO is "Surname Name Patronymic"; blank (spaces) when there is no supervisor.
	SupervisorFIO      string              `json:"supervisor_fio"`
	SupervisorJobTitle string              `json:"supervisor_job_title"`
	Consultants        []GIAStudentAdviser `json:"consultants"`
	CoSupervisors      []GIAStudentAdviser `json:"co_supervisors"`
	GIAStudentComment
}

// GIAStudentAdviser is a consultant or co-supervisor: an ITMO worker (ISU set)
// or an external person (name and workplace set).
type GIAStudentAdviser struct {
	ISU             *int64 `json:"isu,omitzero"`
	Surname         string `json:"surname,omitzero"`
	Name            string `json:"name,omitzero"`
	SecondName      string `json:"second_name,omitzero"`
	WorkPlace       string `json:"work_place,omitzero"`
	JobTitle        string `json:"job_title,omitzero"`
	DegreeTypeID    *int64 `json:"degree_type_id,omitzero"`
	AcademicTitleID *int64 `json:"academic_title_id,omitzero"`
}

// GIAStudentApplicationUpdate saves or sends the application.
type GIAStudentApplicationUpdate struct {
	// StatusID is the current stage status to save a draft, or
	// [GIAStudentStatusSentToApproval] to send it.
	StatusID      int   `json:"status_id"`
	DiplomaTypeID int64 `json:"diploma_type_id"`
	// DiplomaTypeName is set only for [GIAStudentDiplomaTypeOther].
	DiplomaTypeName *string `json:"diploma_type_name"`
	DiplomaTheme    string  `json:"diploma_theme"`
	// DiplomaThemeEN is sent only when given; it must contain no Cyrillic.
	DiplomaThemeEN string `json:"diploma_theme_en,omitzero"`
	// DiplomaJustification is nil for [GIAStudentDiplomaTypeNoJustification].
	DiplomaJustification       *string `json:"diploma_justification"`
	DiplomaExternalPartner     bool    `json:"diploma_external_partner"`
	DiplomaExternalPartnerName *string `json:"diploma_external_partner_name"`
	DiplomaFundamental         bool    `json:"diploma_fundamental"`
	SupervisorISU              *int64  `json:"supervisor_isu"`
	// Consultants and CoSupervisors are sent as null when empty. An ITMO worker
	// is {ISU}; an external person has names, WorkPlace and JobTitle.
	Consultants   []GIAStudentAdviser `json:"consultants"`
	CoSupervisors []GIAStudentAdviser `json:"co_supervisors"`
}

// giaStudentApplicationWire sends empty adviser lists as null.
type giaStudentApplicationWire struct {
	GIAStudentApplicationUpdate
	Consultants   *[]GIAStudentAdviser `json:"consultants"`
	CoSupervisors *[]GIAStudentAdviser `json:"co_supervisors"`
}

func giaStudentAdvisers(v []GIAStudentAdviser) *[]GIAStudentAdviser {
	if len(v) == 0 {
		return nil
	}
	return &v
}

// GIAStudentTask is the individual task stage.
type GIAStudentTask struct {
	MainQuestions *EditorJS `json:"main_questions"`
	// Format is free text: the presentation or format of the work.
	Format    string    `json:"format"`
	ExtraInfo *EditorJS `json:"extra_info"`
	GIAStudentComment
}

// GIAStudentTaskUpdate saves or sends the task stage.
type GIAStudentTaskUpdate struct {
	// StatusID is the current stage status to save a draft, or
	// [GIAStudentStatusSentToApproval] to send it.
	StatusID      int       `json:"status_id"`
	MainQuestions *EditorJS `json:"main_questions"`
	Format        string    `json:"format"`
	ExtraInfo     *EditorJS `json:"extra_info"`
}

// GIAStudentAnnotation is the annotation stage.
type GIAStudentAnnotation struct {
	Goal         *EditorJS               `json:"goal"`
	Tasks        *EditorJS               `json:"tasks"`
	Results      *EditorJS               `json:"results"`
	ExtraInfo    *EditorJS               `json:"extra_info"`
	Grants       []GIAStudentAchievement `json:"grants"`
	Publications []GIAStudentAchievement `json:"publications"`
	Speeches     []GIAStudentAchievement `json:"speeches"`
	GIAStudentComment
}

// GIAStudentAnnotationUpdate saves or sends the annotation stage.
type GIAStudentAnnotationUpdate struct {
	// StatusID is the current stage status to save a draft, or
	// [GIAStudentStatusSentToApproval] to send it.
	StatusID int       `json:"status_id"`
	Goal     *EditorJS `json:"goal"`
	Tasks    *EditorJS `json:"tasks"`
	Results  *EditorJS `json:"results"`
	// GrantIDs, PublicationIDs and SpeechIDs come from [GIAStudentsService.Grants],
	// [GIAStudentsService.Publications] and [GIAStudentsService.Speeches].
	GrantIDs       []int64 `json:"grant_ids"`
	PublicationIDs []int64 `json:"publication_ids"`
	SpeechIDs      []int64 `json:"speech_ids"`
	// ExtraInfo may be nil.
	ExtraInfo *EditorJS `json:"extra_info"`
}

// GIAStudentAchievement is a grant, publication or talk of the student.
type GIAStudentAchievement struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	// Year is a number or a string; shown for grants and publications.
	Year RawJSON `json:"year,omitzero"`
}

// GIAStudentDiplomaFile is the file stage: the uploaded text, antiplagiat
// results, the final file and the reviews.
type GIAStudentDiplomaFile struct {
	// StudentFileKey is "" when nothing is uploaded; see [GIAStudentsService.DownloadFile].
	StudentFileKey  string `json:"student_file_key"`
	StudentFileName string `json:"student_file_name"`
	// Publish is the permission to publish the text.
	Publish bool `json:"publish"`
	// ResultFileKey is the final generated PDF; empty until it exists.
	ResultFileKey           string `json:"result_file_key"`
	AntiplagiatReportWebURL string `json:"antiplagiat_report_web_url"`
	// Originality, Citations, Similarity and SelfCitations are percentages;
	// they are meaningful for statuses 3 and 4.
	Originality             *float64                   `json:"originality"`
	Citations               *float64                   `json:"citations"`
	Similarity              *float64                   `json:"similarity"`
	SelfCitations           *float64                   `json:"self_citations"`
	SupervisorReview        *GIAStudentReview          `json:"supervisor_review"`
	ReviewerReviews         []GIAStudentReview         `json:"reviewer_reviews"`
	ReviewerReviewQuestions []GIAStudentReviewQuestion `json:"reviewer_review_questions"`
	// StatusID and Name: meaning not confirmed.
	StatusID int    `json:"status_id"`
	Name     string `json:"name"`
	GIAStudentComment
}

// GIAStudentReview is a supervisor's or reviewer's review. Criterion values are
// display strings; supervisor and reviewer reviews use different criteria.
type GIAStudentReview struct {
	// ReviewerID is set on reviewer reviews only.
	ReviewerID int64 `json:"reviewer_id,omitzero"`
	// StudentSignature is the date the student acknowledged the review; nil
	// until then. Its format is not confirmed.
	StudentSignature *string `json:"student_signature"`

	// Supervisor criteria.
	InitiativeGoalsName     string `json:"initiative_goals_name,omitzero"`
	InitiativeMethodsName   string `json:"initiative_methods_name,omitzero"`
	ConsistencyName         string `json:"consistency_name,omitzero"`
	InvolvementName         string `json:"involvement_name,omitzero"`
	PreparednessName        string `json:"preparedness_name,omitzero"`
	ApprobationName         string `json:"approbation_name,omitzero"`
	PublicationName         string `json:"publication_name,omitzero"`
	PersonalInvolvementName string `json:"personal_involvement_name,omitzero"`

	// Criteria of both kinds.
	QualityLogicName           string `json:"quality_logic_name,omitzero"`
	JustificationRelevanceName string `json:"justification_relevance_name,omitzero"`
	IntegrationName            string `json:"integration_name,omitzero"`

	// Reviewer criteria.
	RelevantContentName         string `json:"relevant_content_name,omitzero"`
	RelevantEduName             string `json:"relevant_edu_name,omitzero"`
	CorrectMethodsName          string `json:"correct_methods_name,omitzero"`
	JustificationAssertionsName string `json:"justification_assertions_name,omitzero"`
	ValueName                   string `json:"value_name,omitzero"`

	Advantages     string `json:"advantages"`
	Disadvantages  string `json:"disadvantages"`
	Comment        string `json:"comment"`
	AssessmentName string `json:"assessment_name"`
	ThesisComplete bool   `json:"thesis_complete"`
	// AwardingQualifications is a bool or a string.
	AwardingQualifications RawJSON `json:"awarding_qualifications,omitzero"`
}

// GIAStudentReviewQuestion is a question a reviewer asked about the work.
type GIAStudentReviewQuestion struct {
	ID         int64  `json:"id"`
	ReviewerID int64  `json:"reviewer_id"`
	Question   string `json:"question"`
}

// GIAStudentPossiblePresentations lists the defense sessions open for sign-up.
type GIAStudentPossiblePresentations struct {
	PossiblePresentations []GIAStudentDefenseSession `json:"possible_presentations"`
	StudentSignUp         *GIAStudentSignUp          `json:"student_sign_up"`
}

// GIAStudentSignUp is the student's current defense sign-up.
type GIAStudentSignUp struct {
	// PresentationID is nil when the student is not signed up.
	PresentationID *int64 `json:"presentation_id"`
	// Approved fixes the day; the sign-up can no longer change.
	Approved bool `json:"approved"`
}

// GIAStudentDefenseSession is a defense session (day) of a commission.
type GIAStudentDefenseSession struct {
	PresentationID int64 `json:"presentation_id"`
	// PresentationDefenseTypeID is [GIAStudentDefenseTypeOnline] for an online defense.
	PresentationDefenseTypeID   int64  `json:"presentation_defense_type_id"`
	PresentationDefenseTypeName string `json:"presentation_defense_type_name"`
	DefenseDate                 Date   `json:"defense_date"`
	// StartTime is "15:04:05".
	StartTime string `json:"start_time"`
	Link      string `json:"link"`
	Address   string `json:"address"`
	RoomName  string `json:"room_name"`
	// CanChooseDefenseDate allows signing up for this session.
	CanChooseDefenseDate bool `json:"can_choose_defense_date"`
}

// GIAStudentDefense is the defense stage: schedule, result, presentation files
// and the students of the same session.
type GIAStudentDefense struct {
	MainInfo     GIAStudentDefenseInfo       `json:"main_info"`
	Presentation *GIAStudentStoredFile       `json:"presentation"`
	ExtraFiles   *GIAStudentDefenseFiles     `json:"extra_files"`
	Documents    *GIAStudentDefenseDocuments `json:"documents"`
	StudentList  GIAStudentDefenseStudents   `json:"student_list"`
}

// GIAStudentDefenseInfo is the schedule, place and result of a defense.
type GIAStudentDefenseInfo struct {
	DefenseDate Date `json:"defense_date"`
	// StartTime is "15:04:05".
	StartTime                   string         `json:"start_time"`
	PresentationDefenseTypeName string         `json:"presentation_defense_type_name"`
	Address                     string         `json:"address"`
	RoomName                    string         `json:"room_name"`
	Link                        string         `json:"link"`
	Group                       string         `json:"group"`
	Secretary                   GIAStudentName `json:"secretary"`
	Supervisor                  GIAStudentName `json:"supervisor"`
	ThemeRU                     string         `json:"theme_ru"`
	// Assessment is the mark (5 excellent … 2 unsatisfactory) as a number or a
	// string; null before the defense.
	Assessment RawJSON `json:"assessment,omitzero"`
	// AssessmentID is [GIAStudentAssessmentAbsent] when the student was absent.
	AssessmentID        *int64   `json:"assessment_id"`
	AssessmentName      string   `json:"assessment_name"`
	AverageMark         *float64 `json:"average_mark"`
	RedDiploma          bool     `json:"red_diploma"`
	HasSupervisorReview bool     `json:"has_supervisor_review"`
	HasReviewerReview   bool     `json:"has_reviewer_review"`
}

// GIAStudentName is a person's full name.
type GIAStudentName struct {
	Surname    string `json:"surname"`
	Name       string `json:"name"`
	SecondName string `json:"second_name"`
}

// GIAStudentStoredFile is a file stored with the diploma; download it with
// [GIAStudentsService.DownloadFile].
type GIAStudentStoredFile struct {
	FileKey  string `json:"file_key"`
	FileName string `json:"file_name"`
}

// GIAStudentDefenseFiles are the extra defense materials.
type GIAStudentDefenseFiles struct {
	Files []GIAStudentStoredFile `json:"files"`
}

// GIAStudentDefenseDocuments are the documents produced by the defense.
type GIAStudentDefenseDocuments struct {
	ResultFileKey string `json:"result_file_key"`
}

// GIAStudentDefenseStudents are the students of a defense session.
type GIAStudentDefenseStudents struct {
	Students []GIAStudentClassmate `json:"students"`
}

// GIAStudentClassmate is a student of the same defense session.
type GIAStudentClassmate struct {
	ISU int64 `json:"isu"`
	GIAStudentName
}

// GIAStudentSupplement is the diploma supplement the student confirms block by block.
type GIAStudentSupplement struct {
	PersInfo    GIAStudentSupplementPersInfo    `json:"pers_info"`
	Disciplines GIAStudentSupplementDisciplines `json:"disciplines"`
	// ElectiveDisciplines is the "faculties" block of [GIAStudentsService.ApproveSupplementElectives].
	ElectiveDisciplines GIAStudentSupplementDisciplines `json:"elective_disciplines"`
	AdditionalInfo      GIAStudentSupplementAddInfo     `json:"additional_info"`
	// Deadline is shown as is; empty when there is none.
	Deadline string `json:"deadline"`
}

// GIAStudentSupplementStatus is the confirmation status of a supplement block.
type GIAStudentSupplementStatus struct {
	// StatusID is [GIAStudentSupplementEditable] or [GIAStudentSupplementConfirmed].
	StatusID   int    `json:"status_id"`
	StatusName string `json:"status_name"`
}

// GIAStudentSupplementPersInfo is the personal data block of the supplement.
// It cannot be confirmed without SNILS unless IsForeigner.
type GIAStudentSupplementPersInfo struct {
	Status     GIAStudentSupplementStatus `json:"status"`
	Surname    string                     `json:"surname"`
	Name       string                     `json:"name"`
	SecondName string                     `json:"second_name"`
	// BirthDate is shown as is; its format is not confirmed.
	BirthDate   string `json:"birth_date"`
	SNILS       string `json:"snils"`
	IsForeigner bool   `json:"is_foreigner"`
	DirCode     string `json:"dir_code"`
	DirName     string `json:"dir_name"`
	EduLevel    string `json:"edu_level"`
	// EduPeriod is a number or a string.
	EduPeriod RawJSON `json:"edu_period,omitzero"`
	// DocObr is the previous education document.
	DocObr  string `json:"doc_obr"`
	ThemeEN string `json:"theme_en"`
	// BirthPlace and Email are not confirmed.
	BirthPlace string `json:"birth_place,omitzero"`
	Email      string `json:"email,omitzero"`
}

// GIAStudentSupplementDisciplines is a disciplines block of the supplement.
type GIAStudentSupplementDisciplines struct {
	Status        GIAStudentSupplementStatus `json:"status"`
	Blocks        []GIAStudentSupplementPart `json:"blocks"`
	AvgGrade      float64                    `json:"avg_grade"`
	UngradedCount int                        `json:"ungraded_count"`
	TotalCredits  float64                    `json:"total_credits"`
	EPVolume      float64                    `json:"ep_volume"`
	// Changes and Diff are the changes since the last confirmation; either
	// key may be used. Entry keys are not confirmed (type, a
	// discipline name, credit_units, hours, grade, old_value, new_value).
	Changes RawJSON `json:"changes,omitzero"`
	Diff    RawJSON `json:"diff,omitzero"`
}

// GIAStudentSupplementPart is a curriculum block (e.g. "Блок 1") of a supplement.
type GIAStudentSupplementPart struct {
	BlockID      int64                            `json:"block_id"`
	BlockName    string                           `json:"block_name"`
	TotalCredits *float64                         `json:"total_credits"`
	DiscInfo     []GIAStudentSupplementDiscipline `json:"disc_info"`
}

// GIAStudentSupplementDiscipline is a discipline line of the supplement.
type GIAStudentSupplementDiscipline struct {
	DiscID      int64   `json:"disc_id"`
	DiscName    string  `json:"disc_name"`
	CreditUnits float64 `json:"credit_units"`
	HoursAmount float64 `json:"hours_amount"`
	// AvgGrade is a number or a string.
	AvgGrade       RawJSON `json:"avg_grade,omitzero"`
	AvgGradeLetter string  `json:"avg_grade_letter"`
	// Choice is whether an elective is included (not confirmed).
	Choice bool `json:"choice"`
}

// GIAStudentSupplementAddInfo is the additional information block of the supplement.
type GIAStudentSupplementAddInfo struct {
	Status  GIAStudentSupplementStatus    `json:"status"`
	AddInfo []GIAStudentSupplementAddItem `json:"add_info"`
}

// GIAStudentSupplementAddItem is an additional information line.
type GIAStudentSupplementAddItem struct {
	Description string `json:"description"`
	// ID or AdditionalInfoID identifies the line; either may be set, type not confirmed.
	ID               RawJSON `json:"id,omitzero"`
	AdditionalInfoID RawJSON `json:"additional_info_id,omitzero"`
	Choice           bool    `json:"choice"`
}

// GIAStudentElectiveChoice includes or excludes an elective discipline.
type GIAStudentElectiveChoice struct {
	// ObjectID is [GIAStudentSupplementDiscipline].DiscID.
	ObjectID int64 `json:"object_id"`
	Choice   bool  `json:"choice"`
}

// GIAStudentAddInfoChoice includes or excludes an additional information line.
type GIAStudentAddInfoChoice struct {
	// ObjectID is the line's Description.
	ObjectID string `json:"object_id"`
	Choice   bool   `json:"choice"`
}

// GIAStudentEUSupplement is the European diploma supplement request.
type GIAStudentEUSupplement struct {
	// Status.StatusID is one of the GIAStudentEU* constants.
	Status          GIAStudentSupplementStatus  `json:"status"`
	PaymentRequired bool                        `json:"payment_required"`
	PaymentDone     bool                        `json:"payment_done"`
	RejectReason    string                      `json:"reject_reason"`
	PrintReady      bool                        `json:"print_ready"`
	PersInfo        *GIAStudentEUPersInfo       `json:"pers_info"`
	DiplomaInfo     *GIAStudentEUDiplomaInfo    `json:"diploma_info"`
	Disciplines     *GIAStudentEUDisciplineList `json:"disciplines"`
	// ElectiveDisciplines has the same shape as Disciplines.
	ElectiveDisciplines *GIAStudentEUDisciplineList `json:"elective_disciplines"`
}

// GIAStudentEUPersInfo is the personal data of the EU supplement, in Latin script.
type GIAStudentEUPersInfo struct {
	ISU        int64  `json:"isu"`
	Surname    string `json:"surname"`
	Name       string `json:"name"`
	SecondName string `json:"second_name"`
	// BirthDate is shown as is; its format is not confirmed.
	BirthDate       string `json:"birth_date"`
	BirthPlace      string `json:"birth_place"`
	Email           string `json:"email"`
	Foreigner       bool   `json:"foreigner"`
	PrintAttachment bool   `json:"print_attachment"`
}

// GIAStudentEUDiplomaInfo is the diploma data of the EU supplement.
type GIAStudentEUDiplomaInfo struct {
	Qualification string `json:"qualification"`
	EPName        string `json:"ep_name"`
	DirName       string `json:"dir_name"`
	Theme         string `json:"theme"`
	Surname       string `json:"surname"`
	Name          string `json:"name"`
	// BirthDate and DefenceDate are shown as is; their format is not confirmed.
	BirthDate  string `json:"birth_date"`
	BirthPlace string `json:"birth_place"`
	// EduPeriod, TotalCredits and TotalVolume are shown as is; numbers or strings.
	EduPeriod     RawJSON `json:"edu_period,omitzero"`
	TotalCredits  RawJSON `json:"total_credits,omitzero"`
	TotalVolume   RawJSON `json:"total_volume,omitzero"`
	EduLevel      string  `json:"edu_level"`
	EduLevelEN    string  `json:"edu_level_en"`
	DiplomaNumber string  `json:"diploma_number"`
	DefenceDate   string  `json:"defence_date"`
}

// GIAStudentEUDisciplineList is a disciplines block of the EU supplement.
type GIAStudentEUDisciplineList struct {
	Blocks []GIAStudentSupplementPart `json:"blocks"`
}

// GIAStudentEUSupplementRequest creates or resubmits the EU supplement request.
type GIAStudentEUSupplementRequest struct {
	// Name, Surname and BirthPlace must be in Latin script.
	Name       string `json:"name"`
	Surname    string `json:"surname"`
	BirthPlace string `json:"birth_place"`
	// Theme is the English topic.
	Theme string `json:"theme"`
	// PrintAttachment requests a paper copy.
	PrintAttachment bool   `json:"print_attachment"`
	Email           string `json:"email"`
}

// GIAStudentWorker is an ITMO worker who can be a supervisor, consultant or co-supervisor.
type GIAStudentWorker struct {
	ISU        int64  `json:"isu"`
	Surname    string `json:"surname"`
	Name       string `json:"name"`
	SecondName string `json:"second_name"`
	JobTitle   string `json:"job_title"`
}

// GIAStudentDiplomaType is a type of final qualification work.
type GIAStudentDiplomaType struct {
	// DiplomaTypeID: see [GIAStudentDiplomaTypeOther] and [GIAStudentDiplomaTypeNoJustification].
	DiplomaTypeID   int64  `json:"diploma_type_id"`
	DiplomaTypeName string `json:"diploma_type_name"`
}

// GIAStudentStatus is a status of a workflow entity with its badge colour.
type GIAStudentStatus struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	// Color is a hex badge colour: "#E9F7E8" success, "#FEEDF2" danger,
	// "#F7B500" or "#FEF8E5" warning, "#E5F4FF" info.
	Color string `json:"color"`
}

// GIAStudentTitle is an academic title or degree, for external advisers.
type GIAStudentTitle struct {
	ID     int64  `json:"id"`
	NameRU string `json:"name_ru"`
}

func giaStudentDiploma(diplomaID int64, rest string) string {
	return "api/gia-students/diplomas/" + id(diplomaID) + rest
}

func giaStudentAttachment(diplomaID int64, rest string) string {
	return "api/gia-students/attachment/" + id(diplomaID) + rest
}

// giaStudentUpload names the multipart part "file" unless the caller chose a field.
func giaStudentUpload(u Upload) Upload {
	if u.Field == "" {
		u.Field = "file"
	}
	return u
}

// Diplomas returns the student's diplomas, one per educational programme.
// GET /api/gia-students/diplomas/my
func (s *GIAStudentsService) Diplomas(ctx context.Context) ([]GIAStudentDiploma, error) {
	return call[[]GIAStudentDiploma](ctx, s.c, get("api/gia-students/diplomas/my", nil))
}

// DiplomaInfo returns the workflow state of a diploma: open stages, deadlines and contacts.
// GET /api/gia-students/diplomas/{diplomaId}/info
func (s *GIAStudentsService) DiplomaInfo(ctx context.Context, diplomaID int64) (*GIAStudentDiplomaInfo, error) {
	return call[*GIAStudentDiplomaInfo](ctx, s.c, get(giaStudentDiploma(diplomaID, "/info"), nil))
}

// AcceptAgreement accepts the terms; required while the application stage is
// [GIAStudentStatusNotStarted].
// POST /api/gia-students/diplomas/{diplomaId}/agreement
func (s *GIAStudentsService) AcceptAgreement(ctx context.Context, diplomaID int64) error {
	return exec(ctx, s.c, post(giaStudentDiploma(diplomaID, "/agreement"), nil))
}

// Application returns the application stage.
// GET /api/gia-students/diplomas/{diplomaId}/application
func (s *GIAStudentsService) Application(ctx context.Context, diplomaID int64) (*GIAStudentApplication, error) {
	return call[*GIAStudentApplication](ctx, s.c, get(giaStudentDiploma(diplomaID, "/application"), nil))
}

// UpdateApplication saves the application draft or sends it for approval.
// The form is editable in statuses 7 and 3.
// PUT /api/gia-students/diplomas/{diplomaId}/application
func (s *GIAStudentsService) UpdateApplication(ctx context.Context, diplomaID int64, a GIAStudentApplicationUpdate) error {
	w := giaStudentApplicationWire{a, giaStudentAdvisers(a.Consultants), giaStudentAdvisers(a.CoSupervisors)}
	return exec(ctx, s.c, put(giaStudentDiploma(diplomaID, "/application"), w))
}

// Revoke recalls a stage sent for approval back to work. stage is
// [GIAStudentStageApplication], [GIAStudentStageTask] or
// [GIAStudentStageAnnotation]; the file stage is recalled with
// [GIAStudentsService.UploadFile] and [GIAStudentStatusInWork].
// DELETE /api/gia-students/diplomas/{diplomaId}/revoke
func (s *GIAStudentsService) Revoke(ctx context.Context, diplomaID int64, stage string) error {
	return exec(ctx, s.c, withQuery(del(giaStudentDiploma(diplomaID, "/revoke"), nil), q().set("stage", stage)))
}

// Task returns the individual task stage.
// GET /api/gia-students/diplomas/{diplomaId}/task
func (s *GIAStudentsService) Task(ctx context.Context, diplomaID int64) (*GIAStudentTask, error) {
	return call[*GIAStudentTask](ctx, s.c, get(giaStudentDiploma(diplomaID, "/task"), nil))
}

// StartTask moves a not started task stage into work (status_id 7).
// PUT /api/gia-students/diplomas/{diplomaId}/task
func (s *GIAStudentsService) StartTask(ctx context.Context, diplomaID int64) error {
	return exec(ctx, s.c, put(giaStudentDiploma(diplomaID, "/task"), giaStudentStart{GIAStudentStatusInWork}))
}

// UpdateTask saves the task draft or sends it for approval.
// PUT /api/gia-students/diplomas/{diplomaId}/task
func (s *GIAStudentsService) UpdateTask(ctx context.Context, diplomaID int64, t GIAStudentTaskUpdate) error {
	return exec(ctx, s.c, put(giaStudentDiploma(diplomaID, "/task"), t))
}

// Annotation returns the annotation stage.
// GET /api/gia-students/diplomas/{diplomaId}/annotation
func (s *GIAStudentsService) Annotation(ctx context.Context, diplomaID int64) (*GIAStudentAnnotation, error) {
	return call[*GIAStudentAnnotation](ctx, s.c, get(giaStudentDiploma(diplomaID, "/annotation"), nil))
}

// StartAnnotation moves a not started annotation stage into work (status_id 7).
// PUT /api/gia-students/diplomas/{diplomaId}/annotation
func (s *GIAStudentsService) StartAnnotation(ctx context.Context, diplomaID int64) error {
	return exec(ctx, s.c, put(giaStudentDiploma(diplomaID, "/annotation"), giaStudentStart{GIAStudentStatusInWork}))
}

// UpdateAnnotation saves the annotation draft or sends it for approval.
// PUT /api/gia-students/diplomas/{diplomaId}/annotation
func (s *GIAStudentsService) UpdateAnnotation(ctx context.Context, diplomaID int64, a GIAStudentAnnotationUpdate) error {
	return exec(ctx, s.c, put(giaStudentDiploma(diplomaID, "/annotation"), a))
}

// giaStudentStart is the body that opens a not started stage.
type giaStudentStart struct {
	StatusID int `json:"status_id"`
}

// Grants returns the student's grants that can be linked to the annotation.
// GET /api/gia-students/me/grants
func (s *GIAStudentsService) Grants(ctx context.Context) ([]GIAStudentAchievement, error) {
	return call[[]GIAStudentAchievement](ctx, s.c, get("api/gia-students/me/grants", nil))
}

// Publications returns the student's publications that can be linked to the annotation.
// GET /api/gia-students/me/publications
func (s *GIAStudentsService) Publications(ctx context.Context) ([]GIAStudentAchievement, error) {
	return call[[]GIAStudentAchievement](ctx, s.c, get("api/gia-students/me/publications", nil))
}

// Speeches returns the student's conference talks that can be linked to the annotation.
// GET /api/gia-students/me/speeches
func (s *GIAStudentsService) Speeches(ctx context.Context) ([]GIAStudentAchievement, error) {
	return call[[]GIAStudentAchievement](ctx, s.c, get("api/gia-students/me/speeches", nil))
}

// File returns the file stage: the uploaded text, antiplagiat results and reviews.
// GET /api/gia-students/diplomas/{diplomaId}/file
func (s *GIAStudentsService) File(ctx context.Context, diplomaID int64) (*GIAStudentDiplomaFile, error) {
	return call[*GIAStudentDiplomaFile](ctx, s.c, get(giaStudentDiploma(diplomaID, "/file"), nil))
}

// StartFileStage moves a not started file stage into work: status_id 7 in the
// query and no body.
// PUT /api/gia-students/diplomas/{diplomaId}/file
func (s *GIAStudentsService) StartFileStage(ctx context.Context, diplomaID int64) error {
	r := put(giaStudentDiploma(diplomaID, "/file"), nil)
	return exec(ctx, s.c, withQuery(r, q().set("status_id", GIAStudentStatusInWork)))
}

// UploadFile uploads the work text (a PDF) and sets the file stage status:
// the current status to save, [GIAStudentStatusFileSentToApproval] to send,
// [GIAStudentStatusInWork] to recall. publish permits publishing the text.
// status_id and publish go both into the query and into the form. file.Field defaults to "file".
// PUT /api/gia-students/diplomas/{diplomaId}/file
func (s *GIAStudentsService) UploadFile(ctx context.Context, diplomaID int64, statusID int, publish bool, file Upload) error {
	fields := map[string]string{"status_id": strconv.Itoa(statusID), "publish": strconv.FormatBool(publish)}
	r := multipart(http.MethodPut, giaStudentDiploma(diplomaID, "/file"), fields, giaStudentUpload(file))
	return exec(ctx, s.c, withQuery(r, q().set("status_id", statusID).set("publish", publish)))
}

// DownloadFile downloads a stored file by key: StudentFileKey, ResultFileKey,
// the presentation or an extra defense file. The caller closes the body.
// GET /api/gia-students/diplomas/{diplomaId}/file/download/{fileKey}
func (s *GIAStudentsService) DownloadFile(ctx context.Context, diplomaID int64, fileKey string) (*File, error) {
	return download(ctx, s.c, get(giaStudentDiploma(diplomaID, "/file/download/"+id(fileKey)), nil))
}

// AntiplagiatReport downloads the antiplagiat verification certificate (PDF).
// GET /api/gia-students/diplomas/{diplomaId}/file/antiplagiat/report
func (s *GIAStudentsService) AntiplagiatReport(ctx context.Context, diplomaID int64) (*File, error) {
	return download(ctx, s.c, get(giaStudentDiploma(diplomaID, "/file/antiplagiat/report"), nil))
}

// SupervisorReviewFile downloads the supervisor's review (PDF).
// GET /api/gia-students/diplomas/{diplomaId}/review/supervisor/file
func (s *GIAStudentsService) SupervisorReviewFile(ctx context.Context, diplomaID int64) (*File, error) {
	return download(ctx, s.c, get(giaStudentDiploma(diplomaID, "/review/supervisor/file"), nil))
}

// ViewSupervisorReview acknowledges the supervisor's review (sets student_signature).
// POST /api/gia-students/diplomas/{diplomaId}/review/supervisor/view
func (s *GIAStudentsService) ViewSupervisorReview(ctx context.Context, diplomaID int64) error {
	return exec(ctx, s.c, post(giaStudentDiploma(diplomaID, "/review/supervisor/view"), nil))
}

// ReviewerReviewFile downloads a reviewer's review (PDF). reviewerID is
// [GIAStudentReview].ReviewerID.
// GET /api/gia-students/diplomas/{diplomaId}/review/reviewer/{reviewerId}/file
func (s *GIAStudentsService) ReviewerReviewFile(ctx context.Context, diplomaID, reviewerID int64) (*File, error) {
	return download(ctx, s.c, get(giaStudentDiploma(diplomaID, "/review/reviewer/"+id(reviewerID)+"/file"), nil))
}

// ViewReviewerReview acknowledges a reviewer's review.
// POST /api/gia-students/diplomas/{diplomaId}/review/reviewer/{reviewerId}/view
func (s *GIAStudentsService) ViewReviewerReview(ctx context.Context, diplomaID, reviewerID int64) error {
	return exec(ctx, s.c, post(giaStudentDiploma(diplomaID, "/review/reviewer/"+id(reviewerID)+"/view"), nil))
}

// PossiblePresentations returns the defense sessions open for sign-up and the current sign-up.
// GET /api/gia-students/diplomas/{diplomaId}/presentations/possible
func (s *GIAStudentsService) PossiblePresentations(ctx context.Context, diplomaID int64) (*GIAStudentPossiblePresentations, error) {
	return call[*GIAStudentPossiblePresentations](ctx, s.c, get(giaStudentDiploma(diplomaID, "/presentations/possible"), nil))
}

// SignUpPresentation signs up for a defense session; one sign-up at a time.
// POST /api/gia-students/diplomas/{diplomaId}/presentations/{presentationId}/signup
func (s *GIAStudentsService) SignUpPresentation(ctx context.Context, diplomaID, presentationID int64) error {
	return exec(ctx, s.c, post(giaStudentDiploma(diplomaID, "/presentations/"+id(presentationID)+"/signup"), nil))
}

// CancelPresentationSignUp cancels the defense session sign-up.
// DELETE /api/gia-students/diplomas/{diplomaId}/presentations/{presentationId}/signup
func (s *GIAStudentsService) CancelPresentationSignUp(ctx context.Context, diplomaID, presentationID int64) error {
	return exec(ctx, s.c, del(giaStudentDiploma(diplomaID, "/presentations/"+id(presentationID)+"/signup"), nil))
}

// Presentation returns the defense stage: schedule, result, files and session students.
// GET /api/gia-students/diplomas/{diplomaId}/presentation
func (s *GIAStudentsService) Presentation(ctx context.Context, diplomaID int64) (*GIAStudentDefense, error) {
	return call[*GIAStudentDefense](ctx, s.c, get(giaStudentDiploma(diplomaID, "/presentation"), nil))
}

// UploadPresentation uploads or replaces the defense presentation. file.Field defaults to "file".
// PUT /api/gia-students/diplomas/{diplomaId}/presentation/file
func (s *GIAStudentsService) UploadPresentation(ctx context.Context, diplomaID int64, file Upload) error {
	return exec(ctx, s.c, multipart(http.MethodPut, giaStudentDiploma(diplomaID, "/presentation/file"), nil, giaStudentUpload(file)))
}

// AddPresentationExtraFile adds an extra defense material. file.Field defaults to "file".
// POST /api/gia-students/diplomas/{diplomaId}/presentation/extra-files
func (s *GIAStudentsService) AddPresentationExtraFile(ctx context.Context, diplomaID int64, file Upload) error {
	return exec(ctx, s.c, multipart(http.MethodPost, giaStudentDiploma(diplomaID, "/presentation/extra-files"), nil, giaStudentUpload(file)))
}

// DeletePresentationExtraFile deletes an extra defense material by its file key.
// DELETE /api/gia-students/diplomas/{diplomaId}/presentation/extra-files/{fileKey}
func (s *GIAStudentsService) DeletePresentationExtraFile(ctx context.Context, diplomaID int64, fileKey string) error {
	return exec(ctx, s.c, del(giaStudentDiploma(diplomaID, "/presentation/extra-files/"+id(fileKey)), nil))
}

// Supplement returns the diploma supplement the student confirms block by block.
// GET /api/gia-students/attachment/{diplomaId}
func (s *GIAStudentsService) Supplement(ctx context.Context, diplomaID int64) (*GIAStudentSupplement, error) {
	return call[*GIAStudentSupplement](ctx, s.c, get(giaStudentAttachment(diplomaID, ""), nil))
}

func (s *GIAStudentsService) approveSupplement(ctx context.Context, part string, diplomaID int64, body any) error {
	return exec(ctx, s.c, post("api/gia-students/attachment/approve/"+part+"/"+id(diplomaID), body))
}

// ApproveSupplementPersInfo confirms the personal data block (part pers_info).
// POST /api/gia-students/attachment/approve/{part}/{diplomaId}
func (s *GIAStudentsService) ApproveSupplementPersInfo(ctx context.Context, diplomaID int64) error {
	return s.approveSupplement(ctx, "pers_info", diplomaID, nil)
}

// ApproveSupplementDisciplines confirms the disciplines block (part disciplines).
// POST /api/gia-students/attachment/approve/{part}/{diplomaId}
func (s *GIAStudentsService) ApproveSupplementDisciplines(ctx context.Context, diplomaID int64) error {
	return s.approveSupplement(ctx, "disciplines", diplomaID, nil)
}

// ApproveSupplementElectives confirms the elective disciplines block (part
// faculties). choices must cover every elective discipline.
// POST /api/gia-students/attachment/approve/{part}/{diplomaId}
func (s *GIAStudentsService) ApproveSupplementElectives(ctx context.Context, diplomaID int64, choices []GIAStudentElectiveChoice) error {
	if choices == nil {
		choices = []GIAStudentElectiveChoice{}
	}
	return s.approveSupplement(ctx, "faculties", diplomaID, choices)
}

// ApproveSupplementAdditionalInfo confirms the additional information block
// (part additional_info) with the lines to include.
// POST /api/gia-students/attachment/approve/{part}/{diplomaId}
func (s *GIAStudentsService) ApproveSupplementAdditionalInfo(ctx context.Context, diplomaID int64, choices []GIAStudentAddInfoChoice) error {
	if choices == nil {
		choices = []GIAStudentAddInfoChoice{}
	}
	return s.approveSupplement(ctx, "additional_info", diplomaID, choices)
}

// EUSupplement returns the European diploma supplement request. The server
// answers 404 when nothing was submitted: check [Error.IsNotFound] and treat
// it as [GIAStudentEUNotSubmitted].
// GET /api/gia-students/attachment/{diplomaId}/eu
func (s *GIAStudentsService) EUSupplement(ctx context.Context, diplomaID int64) (*GIAStudentEUSupplement, error) {
	return call[*GIAStudentEUSupplement](ctx, s.c, get(giaStudentAttachment(diplomaID, "/eu"), nil))
}

// SubmitEUSupplement creates or resubmits the European supplement request.
// All four supplement blocks must be confirmed first. When the result has
// PaymentRequired and not PaymentDone, continue with [GIAStudentsService.InitEUSupplementPayment].
// POST /api/gia-students/attachment/{diplomaId}/eu
func (s *GIAStudentsService) SubmitEUSupplement(ctx context.Context, diplomaID int64, r GIAStudentEUSupplementRequest) (*GIAStudentEUSupplement, error) {
	return call[*GIAStudentEUSupplement](ctx, s.c, post(giaStudentAttachment(diplomaID, "/eu"), r))
}

// DeleteEUSupplement withdraws the European supplement request.
// DELETE /api/gia-students/attachment/{diplomaId}/eu
func (s *GIAStudentsService) DeleteEUSupplement(ctx context.Context, diplomaID int64) error {
	return exec(ctx, s.c, del(giaStudentAttachment(diplomaID, "/eu"), nil))
}

// ApproveEUSupplement confirms the prepared European supplement (student review → approved).
// POST /api/gia-students/attachment/{diplomaId}/eu/approve
func (s *GIAStudentsService) ApproveEUSupplement(ctx context.Context, diplomaID int64) error {
	return exec(ctx, s.c, post(giaStudentAttachment(diplomaID, "/eu/approve"), nil))
}

// RollbackEUSupplement rejects the prepared European supplement with a comment.
// POST /api/gia-students/attachment/{diplomaId}/eu/rollback
func (s *GIAStudentsService) RollbackEUSupplement(ctx context.Context, diplomaID int64, comment string) error {
	body := struct {
		StudentComment string `json:"student_comment"`
	}{comment}
	return exec(ctx, s.c, post(giaStudentAttachment(diplomaID, "/eu/rollback"), body))
}

// InitEUSupplementPayment starts paying for the European supplement and returns
// the payment gateway URL. The gateway redirects to successURL or failureURL.
// POST /api/gia-students/attachment/{diplomaId}/eu/init_payment
func (s *GIAStudentsService) InitEUSupplementPayment(ctx context.Context, diplomaID int64, successURL, failureURL string) (string, error) {
	body := struct {
		SuccessURL string `json:"SuccessUrl"`
		FailureURL string `json:"FailureUrl"`
	}{successURL, failureURL}
	return call[string](ctx, s.c, post(giaStudentAttachment(diplomaID, "/eu/init_payment"), body))
}

// People searches ITMO workers for the supervisor, consultants and
// co-supervisors. A query of " " returns the default list; an empty
// query omits the parameter.
// GET /api/gia-students/people
func (s *GIAStudentsService) People(ctx context.Context, query string) ([]GIAStudentWorker, error) {
	return call[[]GIAStudentWorker](ctx, s.c, get("api/gia-students/people", q().set("query", query)))
}

// DiplomaTypes returns the types of final qualification work.
// GET /api/gia-students/references/diploma-types
func (s *GIAStudentsService) DiplomaTypes(ctx context.Context) ([]GIAStudentDiplomaType, error) {
	return call[[]GIAStudentDiplomaType](ctx, s.c, get("api/gia-students/references/diploma-types", nil))
}

// Statuses returns the status dictionary of a workflow entity: one of the
// GIAStudentStage* names.
// GET /api/gia-students/references/statuses
func (s *GIAStudentsService) Statuses(ctx context.Context, entity string) ([]GIAStudentStatus, error) {
	return call[[]GIAStudentStatus](ctx, s.c, get("api/gia-students/references/statuses", q().set("entity", entity)))
}

// AcademicTitles returns the academic titles for external advisers.
// GET /api/gia-students/references/academic-titles
func (s *GIAStudentsService) AcademicTitles(ctx context.Context) ([]GIAStudentTitle, error) {
	return call[[]GIAStudentTitle](ctx, s.c, get("api/gia-students/references/academic-titles", nil))
}

// DegreeTypes returns the academic degrees for external advisers.
// GET /api/gia-students/references/degree-types
func (s *GIAStudentsService) DegreeTypes(ctx context.Context) ([]GIAStudentTitle, error) {
	return call[[]GIAStudentTitle](ctx, s.c, get("api/gia-students/references/degree-types", nil))
}

// OrganizationStatuses returns the organization status dictionary. Shape
// not confirmed.
// GET /api/gia-students/references/organization-statuses
func (s *GIAStudentsService) OrganizationStatuses(ctx context.Context) (RawJSON, error) {
	return call[RawJSON](ctx, s.c, get("api/gia-students/references/organization-statuses", nil))
}

// AgreementText returns the agreement text. Shape not confirmed.
// GET /api/gia-students/references/agreement
func (s *GIAStudentsService) AgreementText(ctx context.Context) (RawJSON, error) {
	return call[RawJSON](ctx, s.c, get("api/gia-students/references/agreement", nil))
}

// UserStatus returns the user's GIA role flags. Shape not
// confirmed.
// GET /api/gia-students/users/status
func (s *GIAStudentsService) UserStatus(ctx context.Context) (RawJSON, error) {
	return call[RawJSON](ctx, s.c, get("api/gia-students/users/status", nil))
}
