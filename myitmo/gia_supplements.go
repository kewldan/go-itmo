package myitmo

import (
	"context"
	"net/http"
	"time"
)

// Statuses of a national supplement block.
const (
	GIASupplementBlockConfirmed = 4
	GIASupplementBlockNotFilled = 7
)

// GIASupplementBlock is a block of the national supplement that can be
// returned to the student with [GIAService.RollbackSupplementBlock].
type GIASupplementBlock = string

// National supplement blocks.
const (
	GIABlockPersInfo       GIASupplementBlock = "pers_info"
	GIABlockDisciplines    GIASupplementBlock = "disciplines"
	GIABlockFaculties      GIASupplementBlock = "faculties"
	GIABlockAdditionalInfo GIASupplementBlock = "additional_info"
)

// GIASupplementStudentRef identifies the student of a supplement row.
type GIASupplementStudentRef struct {
	StudentID int64 `json:"student_id"`
	GIANamedStudent
}

// GIASupplementFaculty is the faculty or megafaculty of a supplement row;
// the faculty sends FacultyShortName, the megafaculty ShortName.
type GIASupplementFaculty struct {
	FacultyShortName string `json:"faculty_short_name"`
	ShortName        string `json:"short_name"`
}

// GIASupplementRow is a student of the national supplement list.
type GIASupplementRow struct {
	StudentInfo GIASupplementStudentRef `json:"student_info"`
	EPInfo      GIAEduProgram           `json:"ep_info"`
	Faculty     GIASupplementFaculty    `json:"faculty"`
	Megafaculty GIASupplementFaculty    `json:"megafaculty"`
	GroupID     string                  `json:"group_id"`
	// SecretaryInfo has an unknown shape.
	SecretaryInfo    RawJSON   `json:"secretary_info,omitzero"`
	AttachmentStatus int       `json:"attachment_status"`
	DiplomaStatus    GIAStatus `json:"diploma_status"`
	AvgGrade         float64   `json:"avg_grade"`
	CreditUnits      float64   `json:"credit_units"`
	DefenceDate      time.Time `json:"defence_date"`
	// DefenceGrade, SupervisorGrade and ReviewerGrade are grade names.
	DefenceGrade    string `json:"defence_grade"`
	SupervisorGrade string `json:"supervisor_grade"`
	ReviewerGrade   string `json:"reviewer_grade"`
	ProtocolNumber  string `json:"protocol_number"`
	HasThrees       bool   `json:"has_threes"`
	PerfectDiploma  bool   `json:"perfect_diploma"`
}

// GIASupplementStudentList is a page of the national supplement list.
type GIASupplementStudentList struct {
	AttachmentInfo []GIASupplementRow `json:"attachment_info"`
	TotalCount     int                `json:"total_count"`
}

// GIASupplementPersInfo is the personal block of a national supplement.
type GIASupplementPersInfo struct {
	Surname    string `json:"surname"`
	Name       string `json:"name"`
	SecondName string `json:"second_name"`
	BirthDate  Date   `json:"birth_date"`
	SNILS      string `json:"snils"`
	// DocObr is the previous education document.
	DocObr    string     `json:"doc_obr"`
	EduLevel  string     `json:"edu_level"`
	EduPeriod string     `json:"edu_period"`
	DirCode   string     `json:"dir_code"`
	DirName   string     `json:"dir_name"`
	Status    *GIAStatus `json:"status"`
}

// GIASupplementSection is a block of a national supplement with a status;
// the block rows are kept raw because their container key is not known.
type GIASupplementSection struct {
	// Status.StatusID is GIASupplementBlockConfirmed or GIASupplementBlockNotFilled.
	Status        *GIAStatus `json:"status"`
	AvgGrade      float64    `json:"avg_grade"`
	UngradedCount int        `json:"ungraded_count"`
	// AddInfo is set on the additional-info block.
	AddInfo string `json:"add_info"`
	// Rest holds the other members: sections of {title, total, rows:
	// [{disc_id, disc_name, credit_units, avg_grade}]} for disciplines, rows
	// of {disc_id, disc_name, avg_grade, choice} for electives, rows of
	// {description, value, choice} for additional info.
	Rest RawJSON `json:",embed"`
}

// GIASupplementStudent is the national supplement data of a student.
type GIASupplementStudent struct {
	PersInfo            GIASupplementPersInfo `json:"pers_info"`
	Disciplines         GIASupplementSection  `json:"disciplines"`
	ElectiveDisciplines GIASupplementSection  `json:"elective_disciplines"`
	AdditionalInfo      GIASupplementSection  `json:"additional_info"`
}

// GIASupplementParams filters the national supplement list. Zero values are
// not sent, except Limit and Offset; the export ignores both.
type GIASupplementParams struct {
	Limit, Offset int
	Year          int
	EPID          int64
	FacultyID     int64
	GroupID       string
	Credits       float64
	ZECount       float64
	// AttachmentStatus is sent when set.
	AttachmentStatus *int
	DiplomaStatus    string
	// ValidCredits sends valid_credits=1.
	ValidCredits  bool
	Qualification string
	AvgGrade      string
	ShowAs        GIARole
	Query         string
}

func (p GIASupplementParams) query(page bool) query {
	v := q()
	if page {
		v.set("limit", p.Limit).set("offset", p.Offset)
	}
	giaSet(v, "year", p.Year)
	giaSet(v, "ep_id", p.EPID)
	giaSet(v, "faculty_id", p.FacultyID)
	v.set("group_id", p.GroupID)
	giaSetFloat(v, "credits", p.Credits)
	giaSetFloat(v, "ze_count", p.ZECount)
	if p.AttachmentStatus != nil {
		v.set("attachment_status", *p.AttachmentStatus)
	}
	v.set("diploma_status", p.DiplomaStatus)
	if p.ValidCredits {
		v.set("valid_credits", 1)
	}
	return v.set("qualification", p.Qualification).set("avg_grade", p.AvgGrade).set("query", p.Query).set("show_as", p.ShowAs)
}

// SupplementStudents returns a page of students for the national diploma supplement.
// Staff only.
// GET /api/gia/attachment/students/list
func (s *GIAService) SupplementStudents(ctx context.Context, p GIASupplementParams) (*GIASupplementStudentList, error) {
	return call[*GIASupplementStudentList](ctx, s.c, get("api/gia/attachment/students/list", p.query(true)))
}

// ExportSupplementStudents downloads the national supplement list as a spreadsheet.
// Staff only.
// GET /api/gia/attachment/students/file
func (s *GIAService) ExportSupplementStudents(ctx context.Context, p GIASupplementParams) (*File, error) {
	return download(ctx, s.c, get("api/gia/attachment/students/file", p.query(false)))
}

// SupplementStudent returns the national supplement data of a student.
// Staff only.
// GET /api/gia/attachment/students/{student_id}
func (s *GIAService) SupplementStudent(ctx context.Context, studentID int64) (*GIASupplementStudent, error) {
	return call[*GIASupplementStudent](ctx, s.c, get("api/gia/attachment/students/"+id(studentID), nil))
}

// SyncConfirmedDisciplines refreshes the supplement disciplines from the confirmed academic records.
// Staff only.
// POST /api/gia/attachment/students/{student_id}/sync_confirmed_discs
func (s *GIAService) SyncConfirmedDisciplines(ctx context.Context, studentID int64) error {
	return exec(ctx, s.c, post("api/gia/attachment/students/"+id(studentID)+"/sync_confirmed_discs", nil))
}

// RollbackSupplementBlock returns a supplement block to the student for
// rework. A non-zero
// error_code is an *Error.
// Staff only.
// POST /api/gia/attachment/rollback/{block}/{student_id}
func (s *GIAService) RollbackSupplementBlock(ctx context.Context, block GIASupplementBlock, studentID int64) error {
	return exec(ctx, s.c, post("api/gia/attachment/rollback/"+id(block)+"/"+id(studentID), nil))
}

// GIAEUSupplementRow is a European supplement request of the list.
type GIAEUSupplementRow struct {
	// The id comes under one of these keys; see [GIAEUSupplementRow.Key].
	DiplomaID      int64 `json:"diploma_id"`
	AttachmentID   int64 `json:"attachment_id"`
	ID             int64 `json:"id"`
	EUAttachmentID int64 `json:"eu_attachment_id"`

	Surname string `json:"surname"`
	Name    string `json:"name"`
	ISU     int64  `json:"isu"`
	GroupID string `json:"group_id"`
	EPName  string `json:"ep_name"`
	DirName string `json:"dir_name"`
	// LanguageInfo is "rus" or "eng".
	LanguageInfo    string    `json:"language_info"`
	PaymentRequired bool      `json:"payment_required"`
	PaymentDone     bool      `json:"payment_done"`
	PrintAttachment bool      `json:"print_attachment"`
	CreatedAt       time.Time `json:"created_at"`
	Status          GIAStatus `json:"status"`
}

// Key returns the id to pass to the European supplement methods: the first
// non-zero of attachment_id, eu_attachment_id, id and diploma_id.
func (r GIAEUSupplementRow) Key() int64 {
	for _, n := range []int64{r.AttachmentID, r.EUAttachmentID, r.ID, r.DiplomaID} {
		if n != 0 {
			return n
		}
	}
	return 0
}

// GIAEUSupplementList is a page of European supplement requests.
type GIAEUSupplementList struct {
	AttachmentInfo []GIAEUSupplementRow `json:"attachment_info"`
	TotalCount     int                  `json:"total_count"`
}

// GIAEUPersInfo is the personal block of a European supplement.
type GIAEUPersInfo struct {
	Surname         string `json:"surname"`
	Name            string `json:"name"`
	SecondName      string `json:"second_name"`
	BirthDate       Date   `json:"birth_date"`
	BirthPlace      string `json:"birth_place"`
	Email           string `json:"email"`
	ISU             int64  `json:"isu"`
	Foreigner       bool   `json:"foreigner"`
	PaymentRequired bool   `json:"payment_required"`
	PaymentDone     bool   `json:"payment_done"`
	PrintAttachment bool   `json:"print_attachment"`
}

// GIAEUDiplomaInfo is the diploma block of a European supplement.
type GIAEUDiplomaInfo struct {
	// DiplomaNumber is a number or a string.
	DiplomaNumber RawJSON   `json:"diploma_number,omitzero"`
	DefenceDate   time.Time `json:"defence_date"`
	Direction     string    `json:"direction"`
	DirName       string    `json:"dir_name"`
	EPName        string    `json:"ep_name"`
	EduLevelEN    string    `json:"edu_level_en"`
	Qualification string    `json:"qualification"`
	Theme         string    `json:"theme"`
	EduPeriod     string    `json:"edu_period"`
	TotalCredits  float64   `json:"total_credits"`
	TotalVolume   float64   `json:"total_volume"`
	FileName      string    `json:"file_name"`
	// File has an unknown shape.
	File RawJSON `json:"file,omitzero"`
}

// GIAEUFileMeta describes the uploaded European supplement file.
type GIAEUFileMeta struct {
	FileName string `json:"file_name"`
	Name     string `json:"name"`
}

// GIAEUSupplement is a European diploma supplement.
type GIAEUSupplement struct {
	Status          GIAStatus        `json:"status"`
	Comment         string           `json:"comment"`
	RejectReason    string           `json:"reject_reason"`
	PaymentRequired bool             `json:"payment_required"`
	PaymentDone     bool             `json:"payment_done"`
	StudentData     GIANamedStudent  `json:"student_data"`
	PersInfo        GIAEUPersInfo    `json:"pers_info"`
	DiplomaInfo     GIAEUDiplomaInfo `json:"diploma_info"`
	// Disciplines and ElectiveDisciplines are blocks whose disc_info rows are
	// {disc_id, disc_name, credit_units, avg_grade, avg_grade_letter}.
	Disciplines         RawJSON        `json:"disciplines,omitzero"`
	ElectiveDisciplines RawJSON        `json:"elective_disciplines,omitzero"`
	File                RawJSON        `json:"file,omitzero"`
	AttachmentFile      *GIAEUFileMeta `json:"attachment_file"`
}

// GIAEUDiplomaPatch is the English diploma data of a European supplement.
type GIAEUDiplomaPatch struct {
	Surname    string `json:"surname"`
	Name       string `json:"name"`
	BirthPlace string `json:"birth_place"`
	Direction  string `json:"direction"`
	EPName     string `json:"ep_name"`
	Theme      string `json:"theme"`
}

// GIAEURename is a new English name of a discipline.
type GIAEURename struct {
	DiscID  int64  `json:"disc_id"`
	NewName string `json:"new_name"`
}

// GIAEUDisciplinesPatch renames disciplines of a European supplement;
// Faculties are the elective disciplines.
type GIAEUDisciplinesPatch struct {
	DisciplinesToPatch []GIAEURename `json:"disciplines_to_patch"`
	FacultiesToPatch   []GIAEURename `json:"faculties_to_patch"`
}

// GIAEUSupplementsParams filters [GIAService.EUSupplements]. Zero values are
// not sent, except Limit and Offset.
type GIAEUSupplementsParams struct {
	Limit, Offset int
	Year          int
	EPID          int64
	GroupID       string
	StatusID      int
	// EPLanguage is the programme language, "rus" or "eng".
	EPLanguage string
	Query      string
}

func giaEU(attachmentID int64) string { return "api/gia/attachment/eu/" + id(attachmentID) }

// EUSupplements returns a page of European supplement requests.
// Staff only.
// GET /api/gia/attachment/eu/list
func (s *GIAService) EUSupplements(ctx context.Context, p GIAEUSupplementsParams) (*GIAEUSupplementList, error) {
	v := q().set("limit", p.Limit).set("offset", p.Offset)
	giaSet(v, "year", p.Year)
	giaSet(v, "ep_id", p.EPID)
	v.set("group_id", p.GroupID)
	giaSet(v, "status_id", p.StatusID)
	v.set("ep_language", p.EPLanguage).set("query", p.Query)
	return call[*GIAEUSupplementList](ctx, s.c, get("api/gia/attachment/eu/list", v))
}

// EUSupplement returns a European supplement.
// Staff only.
// GET /api/gia/attachment/eu/{attachment_id}
func (s *GIAService) EUSupplement(ctx context.Context, attachmentID int64) (*GIAEUSupplement, error) {
	return call[*GIAEUSupplement](ctx, s.c, get(giaEU(attachmentID), nil))
}

// EUSupplementFile downloads the European supplement file. Students may call
// it too, passing their diploma id.
// GET /api/gia/attachment/eu/{attachment_id}/file
func (s *GIAService) EUSupplementFile(ctx context.Context, attachmentID int64) (*File, error) {
	return download(ctx, s.c, get(giaEU(attachmentID)+"/file", nil))
}

// UploadEUSupplement uploads the prepared European supplement file (sent as
// the "scan" part; its Field is overwritten).
// Staff only.
// POST /api/gia/attachment/eu/{attachment_id}/file
func (s *GIAService) UploadEUSupplement(ctx context.Context, attachmentID int64, file Upload) error {
	file.Field = "scan"
	return exec(ctx, s.c, multipart(http.MethodPost, giaEU(attachmentID)+"/file", nil, file))
}

// SendEUSupplement sends a European supplement to the student.
// Staff only.
// POST /api/gia/attachment/eu/{attachment_id}/send
func (s *GIAService) SendEUSupplement(ctx context.Context, attachmentID int64) error {
	return exec(ctx, s.c, post(giaEU(attachmentID)+"/send", nil))
}

// RevokeEUSupplement revokes a European supplement.
// Staff only.
// POST /api/gia/attachment/eu/{attachment_id}/revoke
func (s *GIAService) RevokeEUSupplement(ctx context.Context, attachmentID int64) error {
	return exec(ctx, s.c, post(giaEU(attachmentID)+"/revoke", struct{}{}))
}

// RollbackEUSupplement returns a European supplement for rework.
// Staff only.
// POST /api/gia/attachment/eu/{attachment_id}/rollback
func (s *GIAService) RollbackEUSupplement(ctx context.Context, attachmentID int64) error {
	return exec(ctx, s.c, post(giaEU(attachmentID)+"/rollback", struct{}{}))
}

// ConfirmEUSupplement approves a European supplement.
// Staff only.
// POST /api/gia/attachment/eu/{attachment_id}/confirm
func (s *GIAService) ConfirmEUSupplement(ctx context.Context, attachmentID int64) error {
	return exec(ctx, s.c, post(giaEU(attachmentID)+"/confirm", struct{}{}))
}

// NotifyEUSupplementPrinted tells the student that the printed supplement is ready.
// Staff only.
// POST /api/gia/attachment/eu/{attachment_id}/print_notification
func (s *GIAService) NotifyEUSupplementPrinted(ctx context.Context, attachmentID int64) error {
	return exec(ctx, s.c, post(giaEU(attachmentID)+"/print_notification", nil))
}

// UpdateEUSupplementDiploma edits the English personal and diploma data of a European supplement.
// Staff only.
// PATCH /api/gia/attachment/eu/{attachment_id}/modify/diploma
func (s *GIAService) UpdateEUSupplementDiploma(ctx context.Context, attachmentID int64, in GIAEUDiplomaPatch) error {
	return exec(ctx, s.c, patch(giaEU(attachmentID)+"/modify/diploma", in))
}

// RenameEUSupplementDisciplines sets the English names of disciplines in a European supplement.
// Staff only.
// PATCH /api/gia/attachment/eu/{attachment_id}/modify/disciplines
func (s *GIAService) RenameEUSupplementDisciplines(ctx context.Context, attachmentID int64, in GIAEUDisciplinesPatch) error {
	if in.DisciplinesToPatch == nil {
		in.DisciplinesToPatch = []GIAEURename{}
	}
	if in.FacultiesToPatch == nil {
		in.FacultiesToPatch = []GIAEURename{}
	}
	return exec(ctx, s.c, patch(giaEU(attachmentID)+"/modify/disciplines", in))
}

// EUSupplementAttachment downloads the uploaded European supplement file with
// a form POST with an empty url-encoded body.
// Staff only.
// POST /api/gia/attachment/eu/{attachment_id}/attachment
func (s *GIAService) EUSupplementAttachment(ctx context.Context, attachmentID int64) (*File, error) {
	return download(ctx, s.c, giaForm(http.MethodPost, giaEU(attachmentID)+"/attachment"))
}
