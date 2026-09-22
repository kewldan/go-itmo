package myitmo

import (
	"context"
)

// GIAReviewer is an external thesis reviewer.
type GIAReviewer struct {
	// ReviewerID is sent by lists, ID by some pages; either may be zero.
	ReviewerID         int64  `json:"reviewer_id"`
	ID                 int64  `json:"id"`
	ReviewerSurname    string `json:"reviewer_surname"`
	ReviewerName       string `json:"reviewer_name"`
	ReviewerSecondName string `json:"reviewer_second_name"`
	ReviewerEmail      string `json:"reviewer_email"`
	AcademicTitle      string `json:"academic_title"`
	DegreeType         string `json:"degree_type"`
	JobTitle           string `json:"job_title"`
	Workplace          string `json:"workplace"`
	INN                *int64 `json:"inn"`
	// Registered reports whether the reviewer accepted the invitation.
	Registered bool `json:"registered"`
	// DegreeTypeID, AcademicTitleID and OrganizationStatusID are not confirmed.
	DegreeTypeID         *int64 `json:"degree_type_id"`
	AcademicTitleID      *int64 `json:"academic_title_id"`
	OrganizationStatusID *int64 `json:"organization_status_id"`
}

// GIAReviewerInput is the body of reviewer create and update; nil pointers send null.
type GIAReviewerInput struct {
	ReviewerName         string  `json:"reviewer_name"`
	ReviewerSurname      string  `json:"reviewer_surname"`
	ReviewerSecondName   string  `json:"reviewer_second_name"`
	ReviewerEmail        string  `json:"reviewer_email"`
	DegreeTypeID         *int64  `json:"degree_type_id"`
	AcademicTitleID      *int64  `json:"academic_title_id"`
	Workplace            *string `json:"workplace"`
	INN                  *int64  `json:"inn"`
	JobTitle             *string `json:"job_title"`
	OrganizationStatusID *int64  `json:"organization_status_id"`
}

// GIAReviewerList is a page of reviewers.
type GIAReviewerList struct {
	Reviewers []GIAReviewer `json:"reviewers"`
	Total     int           `json:"total"`
}

// GIAReviewerStudent is a student a reviewer can be appointed to.
type GIAReviewerStudent struct {
	StudentID         int64  `json:"student_id"`
	StudentISU        int64  `json:"student_isu"`
	StudentSurname    string `json:"student_surname"`
	StudentName       string `json:"student_name"`
	StudentSecondName string `json:"student_second_name"`
	EPName            string `json:"ep_name"`
	EPEnrollmentYear  int    `json:"ep_enrollment_year"`
}

// GIAReviewedStudent is a thesis the current reviewer reviews.
type GIAReviewedStudent struct {
	DiplomaID         int64  `json:"diploma_id"`
	StudentISU        int64  `json:"student_isu"`
	StudentSurname    string `json:"student_surname"`
	StudentName       string `json:"student_name"`
	StudentSecondName string `json:"student_second_name"`
	GroupID           string `json:"group_id"`
	EPName            string `json:"ep_name"`
	EPEnrollmentYear  int    `json:"ep_enrollment_year"`
	DirCode           string `json:"dir_code"`
	DirName           string `json:"dir_name"`
	// Review has an unknown shape.
	Review RawJSON `json:"review,omitzero"`
}

// GIAReviewedStudentList is a page of [GIAReviewedStudent].
type GIAReviewedStudentList struct {
	Students []GIAReviewedStudent `json:"students"`
	Total    int                  `json:"total"`
}

// GIAReviewedStudentFilters are the options of the reviewer's student list filters.
type GIAReviewedStudentFilters struct {
	EduProgramFilters    []GIAIDName      `json:"edu_program_filters"`
	EduDirectionsFilters []GIAIDName      `json:"edu_directions_filters"`
	GroupFilters         []GIAGroupOption `json:"group_filters"`
}

// GIAAppointment is a student with the appointed reviewers.
type GIAAppointment struct {
	ID                   int64            `json:"id"`
	StudentISU           int64            `json:"student_isu"`
	StudentSurname       string           `json:"student_surname"`
	StudentName          string           `json:"student_name"`
	StudentSecondName    string           `json:"student_second_name"`
	GroupID              string           `json:"group_id"`
	EPName               string           `json:"ep_name"`
	EPEnrollmentYear     int              `json:"ep_enrollment_year"`
	EPEducationLevelName string           `json:"ep_education_level_name"`
	EPFacultyShortName   string           `json:"ep_faculty_short_name"`
	DirCode              string           `json:"dir_code"`
	DirName              string           `json:"dir_name"`
	SupervisorISU        int64            `json:"supervisor_isu"`
	SupervisorSurname    string           `json:"supervisor_surname"`
	SupervisorName       string           `json:"supervisor_name"`
	SupervisorSecondName string           `json:"supervisor_second_name"`
	ReviewerID           int64            `json:"reviewer_id"`
	ReviewerSurname      string           `json:"reviewer_surname"`
	ReviewerName         string           `json:"reviewer_name"`
	ReviewerSecondName   string           `json:"reviewer_second_name"`
	Reviewers            []GIAReviewerRef `json:"reviewers"`
	// Review has an unknown shape.
	Review RawJSON `json:"review,omitzero"`
}

// GIAAppointmentList is a page of reviewer appointments.
type GIAAppointmentList struct {
	Appointments []GIAAppointment `json:"appointments"`
	Total        int              `json:"total"`
	// IsAppointmentsRestricted is true after [GIAService.RestrictAppointments].
	IsAppointmentsRestricted bool `json:"is_appointments_restricted"`
}

// GIAAppointmentFilters are the options of the appointment list filters.
type GIAAppointmentFilters struct {
	EduProgramFilters       []GIAIDName      `json:"edu_program_filters"`
	EduDirectionsFilters    []GIAIDName      `json:"edu_directions_filters"`
	EducationalLevelFilters []GIAIDName      `json:"educational_level_filters"`
	GroupFilters            []GIAGroupOption `json:"group_filters"`
}

// GIAReviewersParams filters [GIAService.Reviewers].
type GIAReviewersParams struct {
	Limit, Offset int
	Query         string
	ShowAs        GIARole
}

// GIAReviewedStudentsParams filters [GIAService.ReviewedStudents]. Zero
// values are not sent, except Limit and Offset.
type GIAReviewedStudentsParams struct {
	Limit, Offset int
	Year          int
	GroupID       string
	EduDirection  int64
	EduProgram    int64
	Query         string
}

// GIAAppointmentsParams filters [GIAService.Appointments]. Zero values are
// not sent, except Limit and Offset.
type GIAAppointmentsParams struct {
	Limit, Offset int
	Year          int
	GroupID       string
	EduDirection  int64
	EduProgram    int64
	Reviewer      int64
	EduLevel      int64
	HasReview     *bool
	HasReviewer   *bool
	ShowAs        GIARole
	Query         string
}

// CreateReviewer creates an external reviewer and returns its id; send the
// invitation with [GIAService.NotifyReviewers]. A duplicate e-mail fails with
// "reviewer with this email already exists".
// Staff only.
// POST /api/gia/reviewers
func (s *GIAService) CreateReviewer(ctx context.Context, in GIAReviewerInput) (int64, error) {
	res, err := call[struct {
		ReviewerID int64 `json:"reviewer_id"`
	}](ctx, s.c, post("api/gia/reviewers", in))
	return res.ReviewerID, err
}

// Reviewer returns a reviewer card.
// Staff only.
// GET /api/gia/reviewers/{reviewer_id}
func (s *GIAService) Reviewer(ctx context.Context, reviewerID int64) (*GIAReviewer, error) {
	return call[*GIAReviewer](ctx, s.c, get("api/gia/reviewers/"+id(reviewerID), nil))
}

// UpdateReviewer edits a reviewer card.
// Staff only.
// PATCH /api/gia/reviewers/{reviewer_id}
func (s *GIAService) UpdateReviewer(ctx context.Context, reviewerID int64, in GIAReviewerInput) error {
	return exec(ctx, s.c, patch("api/gia/reviewers/"+id(reviewerID), in))
}

// ReviewerStudents returns the students appointed to a reviewer.
// Staff only.
// GET /api/gia/reviewers/{reviewer_id}/students
func (s *GIAService) ReviewerStudents(ctx context.Context, reviewerID int64) ([]GIAReviewerStudent, error) {
	return call[[]GIAReviewerStudent](ctx, s.c, get("api/gia/reviewers/"+id(reviewerID)+"/students", nil))
}

// Reviewers returns a page of reviewers.
// Staff only.
// GET /api/gia/reviewers/reviewers
func (s *GIAService) Reviewers(ctx context.Context, p GIAReviewersParams) (*GIAReviewerList, error) {
	v := q().set("limit", p.Limit).set("offset", p.Offset).set("query", p.Query).set("show_as", p.ShowAs)
	return call[*GIAReviewerList](ctx, s.c, get("api/gia/reviewers/reviewers", v))
}

// ReviewedStudents returns a page of the theses the current reviewer reviews.
// Staff only.
// GET /api/gia/reviewers/students
func (s *GIAService) ReviewedStudents(ctx context.Context, p GIAReviewedStudentsParams) (*GIAReviewedStudentList, error) {
	v := q().set("limit", p.Limit).set("offset", p.Offset)
	giaSet(v, "year", p.Year)
	v.set("group_id", p.GroupID)
	giaSet(v, "edu_direction", p.EduDirection)
	giaSet(v, "edu_program", p.EduProgram)
	return call[*GIAReviewedStudentList](ctx, s.c, get("api/gia/reviewers/students", v.set("query", p.Query)))
}

// ReviewedStudentFilters returns the options of the reviewer's student list filters.
// Staff only.
// GET /api/gia/reviewers/students/filters
func (s *GIAService) ReviewedStudentFilters(ctx context.Context) (*GIAReviewedStudentFilters, error) {
	return call[*GIAReviewedStudentFilters](ctx, s.c, get("api/gia/reviewers/students/filters", nil))
}

// SearchReviewerStudents searches students to appoint a reviewer to; showAs is optional.
// Staff only.
// GET /api/gia/reviewers/students/search
func (s *GIAService) SearchReviewerStudents(ctx context.Context, query string, limit int, showAs GIARole) ([]GIAReviewerStudent, error) {
	v := q().set("query", query).set("limit", limit).set("show_as", showAs)
	return call[[]GIAReviewerStudent](ctx, s.c, get("api/gia/reviewers/students/search", v))
}

// Appointments returns a page of students with their appointed reviewers.
// Staff only.
// GET /api/gia/reviewers/appointments
func (s *GIAService) Appointments(ctx context.Context, p GIAAppointmentsParams) (*GIAAppointmentList, error) {
	v := q().set("limit", p.Limit).set("offset", p.Offset)
	giaSet(v, "year", p.Year)
	v.set("group_id", p.GroupID)
	giaSet(v, "edu_direction", p.EduDirection)
	giaSet(v, "edu_program", p.EduProgram)
	giaSet(v, "reviewer", p.Reviewer)
	giaSet(v, "edu_level", p.EduLevel)
	v.set("has_review", p.HasReview).set("has_reviewer", p.HasReviewer).set("show_as", p.ShowAs).set("query", p.Query)
	return call[*GIAAppointmentList](ctx, s.c, get("api/gia/reviewers/appointments", v))
}

// AppointmentFilters returns the options of the appointment list filters; showAs is optional.
// Staff only.
// GET /api/gia/reviewers/appointments/filters
func (s *GIAService) AppointmentFilters(ctx context.Context, showAs GIARole) (*GIAAppointmentFilters, error) {
	return call[*GIAAppointmentFilters](ctx, s.c, get("api/gia/reviewers/appointments/filters", q().set("show_as", showAs)))
}

// AppointReviewer appoints a reviewer to students.
// Staff only.
// POST /api/gia/reviewers/appoint/students
func (s *GIAService) AppointReviewer(ctx context.Context, reviewerID int64, studentIDs ...int64) error {
	body := struct {
		ReviewerID int64   `json:"reviewer_id"`
		StudentIDs []int64 `json:"student_ids"`
	}{reviewerID, studentIDs}
	if body.StudentIDs == nil {
		body.StudentIDs = []int64{}
	}
	return exec(ctx, s.c, post("api/gia/reviewers/appoint/students", body))
}

// NotifyReviewers (re)sends the registration invitation to reviewers.
// Staff only.
// POST /api/gia/reviewers/notify
func (s *GIAService) NotifyReviewers(ctx context.Context, reviewerIDs ...int64) error {
	body := struct {
		ReviewerIDs []int64 `json:"reviewer_ids"`
	}{reviewerIDs}
	if body.ReviewerIDs == nil {
		body.ReviewerIDs = []int64{}
	}
	return exec(ctx, s.c, post("api/gia/reviewers/notify", body))
}

// RestrictAppointments locks editing of reviewer appointments.
// Staff only.
// POST /api/gia/reviewers/edit-restriction
func (s *GIAService) RestrictAppointments(ctx context.Context) error {
	return exec(ctx, s.c, post("api/gia/reviewers/edit-restriction", struct{}{}))
}

// RegisterReviewer completes the registration of an external reviewer with
// the invitation token; the reviewer calls it from their own account.
// POST /api/gia/reviewers/register
func (s *GIAService) RegisterReviewer(ctx context.Context, token string) error {
	body := struct {
		Token string `json:"token"`
	}{token}
	return exec(ctx, s.c, post("api/gia/reviewers/register", body))
}

func giaReviewerDiploma(diplomaID int64) string {
	return "api/gia/reviewers/diplomas/" + id(diplomaID)
}

// ReviewerDiploma returns the thesis card as a reviewer sees it.
// Staff only.
// GET /api/gia/reviewers/diplomas/{diploma_id}/student
func (s *GIAService) ReviewerDiploma(ctx context.Context, diplomaID int64) (*GIADiploma, error) {
	return call[*GIADiploma](ctx, s.c, get(giaReviewerDiploma(diplomaID)+"/student", nil))
}

// SubmitReviewerReview submits (Approved true) or revokes (Approved false) the reviewer's review.
// Staff only.
// PUT /api/gia/reviewers/diplomas/{diploma_id}/review/reviewer
func (s *GIAService) SubmitReviewerReview(ctx context.Context, diplomaID int64, review GIAReviewerReview) error {
	return exec(ctx, s.c, put(giaReviewerDiploma(diplomaID)+"/review/reviewer", review))
}

type giaQuestion struct {
	Question string `json:"question"`
}

// AddReviewerQuestion adds a reviewer question for the defense.
// Staff only.
// POST /api/gia/reviewers/diplomas/{diploma_id}/review/reviewer/questions
func (s *GIAService) AddReviewerQuestion(ctx context.Context, diplomaID int64, question string) error {
	return exec(ctx, s.c, post(giaReviewerDiploma(diplomaID)+"/review/reviewer/questions", giaQuestion{question}))
}

// UpdateReviewerQuestion edits a reviewer question.
// Staff only.
// PATCH /api/gia/reviewers/diplomas/{diploma_id}/review/reviewer/questions/{question_id}
func (s *GIAService) UpdateReviewerQuestion(ctx context.Context, diplomaID, questionID int64, question string) error {
	return exec(ctx, s.c, patch(giaReviewerDiploma(diplomaID)+"/review/reviewer/questions/"+id(questionID), giaQuestion{question}))
}

// DeleteReviewerQuestion deletes a reviewer question.
// Staff only.
// DELETE /api/gia/reviewers/diplomas/{diploma_id}/review/reviewer/questions/{question_id}
func (s *GIAService) DeleteReviewerQuestion(ctx context.Context, diplomaID, questionID int64) error {
	return exec(ctx, s.c, del(giaReviewerDiploma(diplomaID)+"/review/reviewer/questions/"+id(questionID), nil))
}
