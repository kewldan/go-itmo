package myitmo

import (
	"context"
	"time"

	"github.com/kewldan/go-itmo/internal/jsonx"
)

// Defense-day statuses with a known meaning.
const (
	// GIAPresentationStatusDocuments is the first status in which the
	// defense-day documents can be downloaded (lower statuses cannot).
	GIAPresentationStatusDocuments = 4
	// GIAPresentationStatusMarksSent means the marks were sent to ОСОП.
	GIAPresentationStatusMarksSent = 13
)

// Room reservation statuses of a defense day.
const (
	GIAReservationBooked   = 14
	GIAReservationPending  = 15
	GIAReservationDeclined = 16
)

// GIAErrPresentationNotDefensePassed is the string error_code of
// [GIAService.ApproveMarks] when the defense has not ended yet; it starts
// the Message of the returned *Error.
const GIAErrPresentationNotDefensePassed = "presentation_not_defense_passed"

// GIAPresentation is a defense day (день защиты).
type GIAPresentation struct {
	ID               int64     `json:"id"`
	EPID             int64     `json:"ep_id"`
	EPName           string    `json:"ep_name"`
	EPEnrollmentYear int       `json:"ep_enrollment_year"`
	Language         string    `json:"language"`
	DefenseDate      time.Time `json:"defense_date"`
	StartTime        time.Time `json:"start_time"`
	EndTime          time.Time `json:"end_time"`
	// GeneralStateDate is the start of the committee session.
	GeneralStateDate            time.Time `json:"general_state_date"`
	Link                        string    `json:"link"`
	PresentationDefenseTypeID   int64     `json:"presentation_defense_type_id"`
	PresentationDefenseTypeName string    `json:"presentation_defense_type_name"`
	CanChooseDefenseDate        bool      `json:"can_choose_defense_date"`
	StatusID                    int       `json:"status_id"`
	StatusName                  string    `json:"status_name"`
	// ReservationStatusID is one of the GIAReservation constants.
	ReservationStatusID   int    `json:"reservation_status_id"`
	ReservationStatusName string `json:"reservation_status_name"`
	RoomID                int64  `json:"room_id"`
	RoomName              string `json:"room_name"`
	Address               string `json:"address"`
	CategoryID            int64  `json:"category_id"`
	// GroupID is the room group of the booking service.
	GroupID       int64                  `json:"group_id"`
	CreatedBy     int64                  `json:"created_by"`
	Comment       string                 `json:"comment"`
	CommenterRole string                 `json:"commenter_role"`
	Committee     *GIACommittee          `json:"committee"`
	Students      []GIAPresentationEntry `json:"students"`
}

// GIAPresentationEntry is a student of a defense day.
type GIAPresentationEntry struct {
	StudentID         int64  `json:"student_id"`
	StudentISU        int64  `json:"student_isu"`
	ISU               int64  `json:"isu"`
	StudentSurname    string `json:"student_surname"`
	StudentName       string `json:"student_name"`
	StudentSecondName string `json:"student_second_name"`
	// Approved reports whether the sign-up was confirmed.
	Approved             bool    `json:"approved"`
	ID                   int64   `json:"id"`
	AverageMark          float64 `json:"average_mark"`
	RedDiploma           bool    `json:"red_diploma"`
	HasThrees            bool    `json:"has_threes"`
	DiplomaReady         bool    `json:"diploma_ready"`
	HasQuestions         bool    `json:"has_questions"`
	SupervisorSurname    string  `json:"supervisor_surname"`
	SupervisorName       string  `json:"supervisor_name"`
	SupervisorSecondName string  `json:"supervisor_second_name"`
	// SupervisorMark and ReviewerMark are a number or a string.
	SupervisorMark     RawJSON `json:"supervisor_mark,omitzero"`
	ReviewerSurname    string  `json:"reviewer_surname"`
	ReviewerName       string  `json:"reviewer_name"`
	ReviewerSecondName string  `json:"reviewer_second_name"`
	ReviewerMark       RawJSON `json:"reviewer_mark,omitzero"`
	// GeneralStateMark is a [GIAGrade] id.
	GeneralStateMark int64 `json:"general_state_mark"`
}

// GIAPresentationListItem is a defense day of the lists.
type GIAPresentationListItem struct {
	PresentationID     int64             `json:"presentation_id"`
	DefenseDate        time.Time         `json:"defense_date"`
	StartTime          time.Time         `json:"start_time"`
	GeneralStateDate   time.Time         `json:"general_state_date"`
	Place              string            `json:"place"`
	EPName             string            `json:"ep_name"`
	EPEnrollmentYear   int               `json:"ep_enrollment_year"`
	EPFacultyShortName string            `json:"ep_faculty_short_name"`
	DirCode            string            `json:"dir_code"`
	EduDirections      []GIAEduDirection `json:"edu_directions"`
	// Chairman is a name string or an object.
	Chairman                  RawJSON `json:"chairman,omitzero"`
	PresentationDefenseTypeID int64   `json:"presentation_defense_type_id"`
	StatusID                  int     `json:"status_id"`
	StatusName                string  `json:"status_name"`
	ReservationStatusID       int     `json:"reservation_status_id"`
	ReservationStatusName     string  `json:"reservation_status_name"`
	// CreateAt is the creation time; the server spells it without "d".
	CreateAt            time.Time `json:"create_at"`
	CreatedByISU        int64     `json:"created_by_isu"`
	CreatedBySurname    string    `json:"created_by_surname"`
	CreatedByName       string    `json:"created_by_name"`
	CreatedBySecondName string    `json:"created_by_second_name"`
}

// GIAPresentationList is a page of defense days.
type GIAPresentationList struct {
	Presentations []GIAPresentationListItem `json:"presentations"`
	Total         int                       `json:"total"`
}

// GIAPresentationFilters are the options of the defense-day list filters.
type GIAPresentationFilters struct {
	YearFilters             []GIAYearFilter `json:"year_filters"`
	StatusFilters           []GIAIDName     `json:"status_filters"`
	FacultyFilters          []IDValue       `json:"faculty_filters"`
	EduProgramFilters       []GIAIDName     `json:"edu_program_filters"`
	DirFilters              []GIAIDName     `json:"dir_filters"`
	EducationalLevelFilters []GIAIDName     `json:"educational_level_filters"`
	DefenseDateFilters      []GIADateOption `json:"defense_date_filters"`
	// RoleFilters carry the role key in ID.
	RoleFilters []GIARoleOption `json:"role_filters"`
}

// GIAPresentationCommittee is a committee that can be attached to a defense day.
type GIAPresentationCommittee struct {
	CommitteeID int64 `json:"committee_id"`
	// CommitteeNumber and Number are a number or a string.
	CommitteeNumber    RawJSON           `json:"committee_number"`
	Number             RawJSON           `json:"number,omitzero"`
	ChairmanSurname    string            `json:"chairman_surname"`
	ChairmanName       string            `json:"chairman_name"`
	ChairmanSecondName string            `json:"chairman_second_name"`
	CreatedAt          time.Time         `json:"created_at"`
	EduDirections      []GIAEduDirection `json:"edu_directions"`
}

// GIAPresentationCandidate is a student who can be put on a defense day.
type GIAPresentationCandidate struct {
	StudentID         int64  `json:"student_id"`
	StudentISU        int64  `json:"student_isu"`
	StudentSurname    string `json:"student_surname"`
	StudentName       string `json:"student_name"`
	StudentSecondName string `json:"student_second_name"`
	// Date is the defense day the student chose; empty if none.
	Date string `json:"date"`
	// Appointed reports whether the student is on the day.
	Appointed bool `json:"appointed"`
}

// GIADefenseQuestion is a question asked at the defense. ID is not sent on save.
type GIADefenseQuestion struct {
	ID            int64  `json:"id,omitzero"`
	Question      string `json:"question"`
	AnswerQuality string `json:"answer_quality"`
}

// GIADefenseStudent is the card of a student at the defense.
type GIADefenseStudent struct {
	StudentID            int64  `json:"student_id"`
	StudentISU           int64  `json:"student_isu"`
	StudentSurname       string `json:"student_surname"`
	StudentName          string `json:"student_name"`
	StudentSecondName    string `json:"student_second_name"`
	StudentGroup         string `json:"student_group"`
	DiplomaID            int64  `json:"diploma_id"`
	ThemeRU              string `json:"theme_ru"`
	DirCode              string `json:"dir_code"`
	DirName              string `json:"dir_name"`
	EPName               string `json:"ep_name"`
	EPEnrollmentYear     int    `json:"ep_enrollment_year"`
	EPEducationLevelName string `json:"ep_education_level_name"`
	EPFaculty            string `json:"ep_faculty"`
	SupervisorSurname    string `json:"supervisor_surname"`
	SupervisorName       string `json:"supervisor_name"`
	SecretarySurname     string `json:"secretary_surname"`
	SecretaryName        string `json:"secretary_name"`
	SecretarySecondName  string `json:"secretary_second_name"`
	Mark                 string `json:"mark"`
	MarkID               int64  `json:"mark_id"`
	IsAuthor             bool   `json:"is_author"`
	HasSupervisorReview  bool   `json:"has_supervisor_review"`
	HasReviewerReview    bool   `json:"has_reviewer_review"`
	// ResultFile, Presentation and ExtraFiles download with [GIAService.DiplomaFile].
	ResultFile   *GIAFileRef          `json:"result_file"`
	Presentation *GIAFileRef          `json:"presentation"`
	ExtraFiles   []GIAFileRef         `json:"extra_files"`
	Questions    []GIADefenseQuestion `json:"questions"`
}

// GIABookingInfo is the room booking of a defense day. Times are sent as
// "2006-01-02 15:04" in their own location; pass Moscow times.
type GIABookingInfo struct {
	Name           string    `json:"name,omitzero"`
	AdditionalInfo string    `json:"additional_info,omitzero"`
	Participants   RawJSON   `json:"participants,omitzero"`
	ContactPhone   string    `json:"contact_phone,omitzero"`
	CoBookers      []int64   `json:"co_bookers"`
	StartDatetime  time.Time `json:"-"`
	EndDatetime    time.Time `json:"-"`
	EventID        *int64    `json:"event_id"`
	RoomID         int64     `json:"room_id"`
	GroupID        int64     `json:"group_id"`
	CategoryID     int64     `json:"category_id"`
	Equipment      RawJSON   `json:"equipment,omitzero"`
	RoomName       string    `json:"room_name"`
	Address        string    `json:"address"`
}

// MarshalJSON implements json.Marshaler; it formats the booking times the
// way the booking service expects them.
func (b GIABookingInfo) MarshalJSON() ([]byte, error) {
	type plain GIABookingInfo
	const layout = "2006-01-02 15:04"
	return jsonx.Marshal(struct {
		plain
		StartDatetime string `json:"start_datetime"`
		EndDatetime   string `json:"end_datetime"`
	}{plain(b), b.StartDatetime.Format(layout), b.EndDatetime.Format(layout)})
}

// GIAPresentationInput is the body of defense-day create and update.
type GIAPresentationInput struct {
	// EduProgramID is omitted for ОГЭК days.
	EduProgramID              int64     `json:"edu_program_id,omitzero"`
	DefenseDate               time.Time `json:"defense_date"`
	PresentationDefenseTypeID int64     `json:"presentation_defense_type_id"`
	CanChooseDefenseDate      bool      `json:"can_choose_defense_date"`
	// Building is the room category name, Auditorium the room name.
	Building         string    `json:"building"`
	Auditorium       string    `json:"auditorium"`
	GeneralStateDate time.Time `json:"general_state_date"`
	StartTime        time.Time `json:"start_time"`
	EndTime          time.Time `json:"end_time"`
	// IsScratch saves a draft.
	IsScratch bool   `json:"is_scratch"`
	Link      string `json:"link"`
	// IsGeneralState creates an ОГЭК day; create only.
	IsGeneralState bool `json:"is_general_state,omitzero"`
	// BookingInfo books a room; nil when no room is chosen.
	BookingInfo *GIABookingInfo `json:"booking_info,omitzero"`
}

// GIAQueuePosition is the place of a student in the defense queue (1-based).
type GIAQueuePosition struct {
	StudentID    int64 `json:"student_id"`
	OrderInQueue int   `json:"order_in_queue"`
}

// GIAMarkEntry is the committee mark of a student. Mark is a [GIAGrade] id; nil clears it.
type GIAMarkEntry struct {
	StudentID int64  `json:"student_id"`
	Mark      *int64 `json:"mark"`
}

// GIAPresentationsParams filters the defense-day lists. Zero values are not
// sent, except Limit and Offset.
type GIAPresentationsParams struct {
	Limit, Offset int
	Year          int
	StatusID      int
	FacultyID     int64
	EduProgramID  int64
	DirID         int64
	EduLevelID    int64
	// Date is a defense date from [GIAPresentationFilters.DefenseDateFilters].
	Date   string
	ShowAs GIARole
	Query  string
}

func (p GIAPresentationsParams) query() query {
	v := q().set("limit", p.Limit).set("offset", p.Offset)
	giaSet(v, "year", p.Year)
	giaSet(v, "status_id", p.StatusID)
	giaSet(v, "faculty_id", p.FacultyID)
	giaSet(v, "edu_program_id", p.EduProgramID)
	giaSet(v, "dir_id", p.DirID)
	giaSet(v, "edu_level_id", p.EduLevelID)
	return v.set("date", p.Date).set("show_as", p.ShowAs).set("query", p.Query)
}

func giaPresentation(presentationID int64) string {
	return "api/gia/presentations/" + id(presentationID)
}

// Presentations returns a page of defense days.
// Staff only.
// GET /api/gia/presentations
func (s *GIAService) Presentations(ctx context.Context, p GIAPresentationsParams) (*GIAPresentationList, error) {
	return call[*GIAPresentationList](ctx, s.c, get("api/gia/presentations", p.query()))
}

// CreatePresentation creates a defense day, optionally booking a room.
// Staff only.
// POST /api/gia/presentations
func (s *GIAService) CreatePresentation(ctx context.Context, in GIAPresentationInput) error {
	return exec(ctx, s.c, post("api/gia/presentations", in))
}

// Presentation returns a defense day with its committee and students.
// Staff only.
// GET /api/gia/presentations/{presentation_id}
func (s *GIAService) Presentation(ctx context.Context, presentationID int64) (*GIAPresentation, error) {
	return call[*GIAPresentation](ctx, s.c, get(giaPresentation(presentationID), nil))
}

// UpdatePresentation edits a defense day; IsGeneralState is ignored.
// Staff only.
// PATCH /api/gia/presentations/{presentation_id}
func (s *GIAService) UpdatePresentation(ctx context.Context, presentationID int64, in GIAPresentationInput) error {
	in.IsGeneralState = false
	return exec(ctx, s.c, patch(giaPresentation(presentationID), in))
}

// ApprovePresentation approves the date of a defense day.
// Staff only.
// PATCH /api/gia/presentations/{presentation_id}/status
func (s *GIAService) ApprovePresentation(ctx context.Context, presentationID int64) error {
	return exec(ctx, s.c, patch(giaPresentation(presentationID)+"/status", struct{}{}))
}

// RejectPresentation rejects a defense day with a comment.
// Staff only.
// PATCH /api/gia/presentations/{presentation_id}/status
func (s *GIAService) RejectPresentation(ctx context.Context, presentationID int64, comment string) error {
	return exec(ctx, s.c, patch(giaPresentation(presentationID)+"/status", giaComment{comment}))
}

// RevokePresentation withdraws a defense day.
// Staff only.
// POST /api/gia/presentations/{presentation_id}/revoke
func (s *GIAService) RevokePresentation(ctx context.Context, presentationID int64) error {
	return exec(ctx, s.c, post(giaPresentation(presentationID)+"/revoke", nil))
}

// SetPresentationLink sets the online meeting link of a defense day.
// Staff only.
// POST /api/gia/presentations/{presentation_id}/links
func (s *GIAService) SetPresentationLink(ctx context.Context, presentationID int64, link string) error {
	body := struct {
		Link string `json:"link"`
	}{link}
	return exec(ctx, s.c, post(giaPresentation(presentationID)+"/links", body))
}

// AttachCommittee attaches a committee to a defense day.
// Staff only.
// POST /api/gia/presentations/{presentation_id}/committees
func (s *GIAService) AttachCommittee(ctx context.Context, presentationID, committeeID int64) error {
	body := struct {
		CommitteeID int64 `json:"committee_id"`
	}{committeeID}
	return exec(ctx, s.c, post(giaPresentation(presentationID)+"/committees", body))
}

// ApproveMarks sends the final marks of a defense day to ОСОП.
//
// This route answers errors with a string error_code, which the generic
// envelope (int error_code) cannot decode, so the call uses its own
// envelope. A string code becomes the start of Error.Message
// ("<code>: <message>") and Error.Code stays 0; see
// [GIAErrPresentationNotDefensePassed] for the defense that has not ended.
// Staff only.
// POST /api/gia/presentations/{presentation_id}/marks/approve
func (s *GIAService) ApproveMarks(ctx context.Context, presentationID int64) error {
	return giaExecLoose(ctx, s.c, post(giaPresentation(presentationID)+"/marks/approve", nil))
}

// SetPresentationStudents replaces the students assigned to a defense day.
// Staff only.
// PATCH /api/gia/presentations/{presentation_id}/students
func (s *GIAService) SetPresentationStudents(ctx context.Context, presentationID int64, studentIDs ...int64) error {
	if studentIDs == nil {
		studentIDs = []int64{}
	}
	return exec(ctx, s.c, patch(giaPresentation(presentationID)+"/students", studentIDs))
}

// OrderPresentationStudents sets the order of students in the defense queue.
// Staff only.
// PATCH /api/gia/presentations/{presentation_id}/students/order
func (s *GIAService) OrderPresentationStudents(ctx context.Context, presentationID int64, order []GIAQueuePosition) error {
	return exec(ctx, s.c, patch(giaPresentation(presentationID)+"/students/order", order))
}

// SetPresentationMarks sets the committee marks of students.
// Staff only.
// PATCH /api/gia/presentations/{presentation_id}/students/marks
func (s *GIAService) SetPresentationMarks(ctx context.Context, presentationID int64, marks []GIAMarkEntry) error {
	return exec(ctx, s.c, patch(giaPresentation(presentationID)+"/students/marks", marks))
}

// ApprovePresentationStudent confirms the sign-up of a student for a defense day.
// Staff only.
// PATCH /api/gia/presentations/{presentation_id}/students/{student_id}/approve
func (s *GIAService) ApprovePresentationStudent(ctx context.Context, presentationID, studentID int64) error {
	return exec(ctx, s.c, patch(giaPresentation(presentationID)+"/students/"+id(studentID)+"/approve", nil))
}

// RejectPresentationStudent rejects the sign-up of a student for a defense day.
// Staff only.
// PATCH /api/gia/presentations/{presentation_id}/students/{student_id}/reject
func (s *GIAService) RejectPresentationStudent(ctx context.Context, presentationID, studentID int64) error {
	return exec(ctx, s.c, patch(giaPresentation(presentationID)+"/students/"+id(studentID)+"/reject", nil))
}

// DefenseQuestions returns the questions asked to a student at the defense.
// Staff only.
// GET /api/gia/presentations/{presentation_id}/students/{student_id}/questions
func (s *GIAService) DefenseQuestions(ctx context.Context, presentationID, studentID int64) ([]GIADefenseQuestion, error) {
	return call[[]GIADefenseQuestion](ctx, s.c, get(giaPresentation(presentationID)+"/students/"+id(studentID)+"/questions", nil))
}

// SetDefenseQuestions saves the defense questions of a student and the approval flag.
// Staff only.
// PATCH /api/gia/presentations/{presentation_id}/students/{student_id}/questions
func (s *GIAService) SetDefenseQuestions(ctx context.Context, presentationID, studentID int64, approved bool, questions []GIADefenseQuestion) error {
	body := struct {
		Approved  bool                 `json:"approved"`
		Questions []GIADefenseQuestion `json:"questions"`
	}{approved, make([]GIADefenseQuestion, 0, len(questions))}
	for _, qn := range questions {
		body.Questions = append(body.Questions, GIADefenseQuestion{Question: qn.Question, AnswerQuality: qn.AnswerQuality})
	}
	return exec(ctx, s.c, patch(giaPresentation(presentationID)+"/students/"+id(studentID)+"/questions", body))
}

func (s *GIAService) presentationDoc(ctx context.Context, presentationID int64, doc string, format GIAFormat) (*File, error) {
	return download(ctx, s.c, get(giaPresentation(presentationID)+"/"+doc, q().set("format", format)))
}

// IndividualProtocols downloads a ZIP of the individual protocols of all students of a defense day.
// Staff only.
// GET /api/gia/presentations/{presentation_id}/indv-protocols-zip
func (s *GIAService) IndividualProtocols(ctx context.Context, presentationID int64, format GIAFormat) (*File, error) {
	return s.presentationDoc(ctx, presentationID, "indv-protocols-zip", format)
}

// AttendanceSheet downloads the attendance sheet (явочный лист) of a defense day.
// Staff only.
// GET /api/gia/presentations/{presentation_id}/attendance-sheet
func (s *GIAService) AttendanceSheet(ctx context.Context, presentationID int64, format GIAFormat) (*File, error) {
	return s.presentationDoc(ctx, presentationID, "attendance-sheet", format)
}

// DefenseOrder downloads the order of the day (порядок дня).
// Staff only.
// GET /api/gia/presentations/{presentation_id}/defense-order
func (s *GIAService) DefenseOrder(ctx context.Context, presentationID int64, format GIAFormat) (*File, error) {
	return s.presentationDoc(ctx, presentationID, "defense-order", format)
}

// DefenseAct downloads the act on procedure compliance.
// Staff only.
// GET /api/gia/presentations/{presentation_id}/defense-act
func (s *GIAService) DefenseAct(ctx context.Context, presentationID int64, format GIAFormat) (*File, error) {
	return s.presentationDoc(ctx, presentationID, "defense-act", format)
}

// ChairmanConclusion downloads the conclusion of the committee chairman.
// Staff only.
// GET /api/gia/presentations/{presentation_id}/chairman-conclusion
func (s *GIAService) ChairmanConclusion(ctx context.Context, presentationID int64, format GIAFormat) (*File, error) {
	return s.presentationDoc(ctx, presentationID, "chairman-conclusion", format)
}

// MarkTemplate downloads the thesis mark template.
// Staff only.
// GET /api/gia/presentations/{presentation_id}/mark-template
func (s *GIAService) MarkTemplate(ctx context.Context, presentationID int64, format GIAFormat) (*File, error) {
	return s.presentationDoc(ctx, presentationID, "mark-template", format)
}

// PresentationFiles downloads a ZIP of all documents of a defense day.
// Staff only.
// GET /api/gia/presentations/{presentation_id}/all-files
func (s *GIAService) PresentationFiles(ctx context.Context, presentationID int64, format GIAFormat) (*File, error) {
	return s.presentationDoc(ctx, presentationID, "all-files", format)
}

// PresentationMaterials downloads a ZIP of the students' materials of a defense day.
// Staff only.
// GET /api/gia/presentations/{presentation_id}/materials-zip
func (s *GIAService) PresentationMaterials(ctx context.Context, presentationID int64) (*File, error) {
	return download(ctx, s.c, get(giaPresentation(presentationID)+"/materials-zip", nil))
}

// IndividualProtocol downloads the individual defense protocol of a student;
// format is optional (the server then presumably sends PDF).
// Staff only.
// GET /api/gia/presentations/students/{student_id}/indv-protocol
func (s *GIAService) IndividualProtocol(ctx context.Context, studentID int64, format GIAFormat) (*File, error) {
	return download(ctx, s.c, get("api/gia/presentations/students/"+id(studentID)+"/indv-protocol", q().set("format", format)))
}

// DefenseStudent returns the card of a student at the defense.
// Staff only.
// GET /api/gia/presentations/students/{student_id}/main-info
func (s *GIAService) DefenseStudent(ctx context.Context, studentID int64) (*GIADefenseStudent, error) {
	return call[*GIADefenseStudent](ctx, s.c, get("api/gia/presentations/students/"+id(studentID)+"/main-info", nil))
}

// PresentationCandidates returns the students of a programme who can be put
// on a defense day, keyed by study group.
// Staff only.
// GET /api/gia/presentations/students
func (s *GIAService) PresentationCandidates(ctx context.Context, eduProgramID int64) (map[string][]GIAPresentationCandidate, error) {
	return call[map[string][]GIAPresentationCandidate](ctx, s.c, get("api/gia/presentations/students", q().set("edu_program_id", eduProgramID)))
}

// SecretaryPresentations returns the defense days visible to the current secretary.
// Staff only.
// GET /api/gia/presentations/secretary
func (s *GIAService) SecretaryPresentations(ctx context.Context, p GIAPresentationsParams) ([]GIAPresentation, error) {
	return call[[]GIAPresentation](ctx, s.c, get("api/gia/presentations/secretary", p.query()))
}

// GeneralStatePresentations returns a page of ОГЭК defense days.
// Staff only.
// GET /api/gia/presentations/general-state
func (s *GIAService) GeneralStatePresentations(ctx context.Context, p GIAPresentationsParams) (*GIAPresentationList, error) {
	return call[*GIAPresentationList](ctx, s.c, get("api/gia/presentations/general-state", p.query()))
}

// PresentationFilters returns the options of the defense-day list filters; showAs is optional.
// Staff only.
// GET /api/gia/presentations/filters
func (s *GIAService) PresentationFilters(ctx context.Context, showAs GIARole) (*GIAPresentationFilters, error) {
	return call[*GIAPresentationFilters](ctx, s.c, get("api/gia/presentations/filters", q().set("show_as", showAs)))
}

// GeneralStatePresentationFilters returns the options of the ОГЭК defense-day
// list filters; showAs is optional.
// Staff only.
// GET /api/gia/presentations/general-state/filters
func (s *GIAService) GeneralStatePresentationFilters(ctx context.Context, showAs GIARole) (*GIAPresentationFilters, error) {
	return call[*GIAPresentationFilters](ctx, s.c, get("api/gia/presentations/general-state/filters", q().set("show_as", showAs)))
}

// PresentationCommittees returns the committees that can be attached to a
// defense day of the programme.
// Staff only.
// GET /api/gia/presentations/committees
func (s *GIAService) PresentationCommittees(ctx context.Context, eduProgramID int64) ([]GIAPresentationCommittee, error) {
	return call[[]GIAPresentationCommittee](ctx, s.c, get("api/gia/presentations/committees", q().set("edu_program_id", eduProgramID)))
}

// PresentationPrograms returns the programmes for which a defense day can be created.
// Staff only.
// GET /api/gia/presentations/edu-programs
func (s *GIAService) PresentationPrograms(ctx context.Context, query string, limit int) ([]GIAEduProgram, error) {
	return call[[]GIAEduProgram](ctx, s.c, get("api/gia/presentations/edu-programs", q().set("query", query).set("limit", limit)))
}

// DefenseTypes returns the defense formats (offline, online, ...).
// Staff only.
// GET /api/gia/presentations/defense-types
func (s *GIAService) DefenseTypes(ctx context.Context) ([]GIAIDName, error) {
	return call[[]GIAIDName](ctx, s.c, get("api/gia/presentations/defense-types", nil))
}
