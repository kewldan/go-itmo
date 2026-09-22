package myitmo

import (
	"context"
)

// GIASecretary is a state examination committee secretary.
type GIASecretary struct {
	SecretaryID          int64  `json:"secretary_id"`
	SecretaryISU         int64  `json:"secretary_isu"`
	SecretarySurname     string `json:"secretary_surname"`
	SecretaryName        string `json:"secretary_name"`
	SecretarySecondName  string `json:"secretary_second_name"`
	SecretaryJobTitle    string `json:"secretary_job_title"`
	SecretaryEmail       string `json:"secretary_email"`
	SecretaryPhoneNumber string `json:"secretary_phone_number"`
}

// GIASecretaryRejection is why a secretary assignment was declined.
type GIASecretaryRejection struct {
	RejectedBy       string `json:"rejected_by"`
	RejectionReason  string `json:"rejection_reason"`
	RejectorJobTitle string `json:"rejector_job_title"`
}

// GIASecretaryGroup is a study group with its secretary assignment.
type GIASecretaryGroup struct {
	GroupID        string                 `json:"group_id"`
	StudentsCount  int                    `json:"students_count"`
	AssignedCount  int                    `json:"assigned_count"`
	IsGeneralState bool                   `json:"is_general_state"`
	StatusID       int                    `json:"status_id"`
	StatusName     string                 `json:"status_name"`
	Secretary      *GIASecretary          `json:"secretary"`
	RejectInfo     *GIASecretaryRejection `json:"reject_info"`
	DirsInfo       []GIAEduDirection      `json:"dirs_info"`
	// Students has an unknown item shape.
	Students RawJSON `json:"students,omitzero"`
}

// GIASecretaryProgram is a programme with the secretary assignments of its groups.
type GIASecretaryProgram struct {
	EPID                int64               `json:"ep_id"`
	EPName              string              `json:"ep_name"`
	EPEnrollmentYear    int                 `json:"ep_enrollment_year"`
	EPManagerISU        int64               `json:"ep_manager_isu"`
	EPManagerSurname    string              `json:"ep_manager_surname"`
	EPManagerName       string              `json:"ep_manager_name"`
	EPManagerSecondName string              `json:"ep_manager_second_name"`
	Groups              []GIASecretaryGroup `json:"groups"`
}

// GIASecretaryList is a page of secretary assignments by programme.
type GIASecretaryList struct {
	Secretaries []GIASecretaryProgram `json:"secretaries"`
	TotalCount  int                   `json:"total_count"`
}

// GIAGeneralStateSecretary is a student with the ОГЭК secretary.
type GIAGeneralStateSecretary struct {
	ID                  int64  `json:"id"`
	StudentISU          int64  `json:"student_isu"`
	StudentSurname      string `json:"student_surname"`
	StudentName         string `json:"student_name"`
	StudentSecondName   string `json:"student_second_name"`
	GroupID             string `json:"group_id"`
	EPName              string `json:"ep_name"`
	EPEnrollmentYear    int    `json:"ep_enrollment_year"`
	DirCode             string `json:"dir_code"`
	DirName             string `json:"dir_name"`
	SecretarySurname    string `json:"secretary_surname"`
	SecretaryName       string `json:"secretary_name"`
	SecretarySecondName string `json:"secretary_second_name"`
	StatusName          string `json:"status_name"`
}

// GIAGeneralStateSecretaryList is a page of [GIAGeneralStateSecretary].
type GIAGeneralStateSecretaryList struct {
	Secretaries []GIAGeneralStateSecretary `json:"secretaries"`
	TotalCount  int                        `json:"total_count"`
}

// GIAMyStudents is the current secretary with the assigned students.
type GIAMyStudents struct {
	// Students is an object, not a list: the secretary and their students.
	Students   GIAMySecretary `json:"students"`
	TotalCount int            `json:"total_count"`
}

// GIAMySecretary is the current secretary with the students by programme.
type GIAMySecretary struct {
	GIASecretary
	SecretaryStudents []GIAMyProgram `json:"secretary_students"`
}

// GIAMyProgram is a programme of [GIAMySecretary].
type GIAMyProgram struct {
	EPID             int64        `json:"ep_id"`
	EPName           string       `json:"ep_name"`
	EPEnrollmentYear int          `json:"ep_enrollment_year"`
	Groups           []GIAMyGroup `json:"groups"`
}

// GIAMyGroup is a study group of [GIAMyProgram].
type GIAMyGroup struct {
	GroupID       string           `json:"group_id"`
	StatusID      int              `json:"status_id"`
	StatusName    string           `json:"status_name"`
	StudentsCount int              `json:"students_count"`
	Students      []GIAMyGroupItem `json:"students"`
}

// GIAMyGroupItem is a student assignment of [GIAMyGroup].
type GIAMyGroupItem struct {
	SecretaryID int64           `json:"secretary_id"`
	StudentInfo GIANamedStudent `json:"student_info"`
}

// GIANamedStudent identifies a student by ISU number and name.
type GIANamedStudent struct {
	StudentISU        int64  `json:"student_isu"`
	StudentSurname    string `json:"student_surname"`
	StudentName       string `json:"student_name"`
	StudentSecondName string `json:"student_second_name"`
}

// GIAUnassignedStudent is a student without a secretary.
type GIAUnassignedStudent struct {
	StudentID         int64  `json:"student_id"`
	StudentISU        int64  `json:"student_isu"`
	StudentSurname    string `json:"student_surname"`
	StudentName       string `json:"student_name"`
	StudentSecondName string `json:"student_second_name"`
	GroupID           string `json:"group_id"`
	EPName            string `json:"ep_name"`
	DirCode           string `json:"dir_code"`
}

// GIAUnassignedStudents is a page of students without a secretary.
type GIAUnassignedStudents struct {
	Students []GIAUnassignedStudent `json:"students"`
	// TotalCount is the total number of students. The server spells the key
	// "total_сount" with a Cyrillic "с" (U+0441) instead of the Latin "c";
	// the tag reproduces it on purpose.
	TotalCount int `json:"total_сount"`
}

// GIACoordinator is an ОГЭК coordinator.
type GIACoordinator struct {
	ISU        int64  `json:"isu"`
	Surname    string `json:"surname"`
	Name       string `json:"name"`
	SecondName string `json:"second_name"`
}

// GIACoordinatorTerm appoints or removes an ОГЭК coordinator for an academic year.
type GIACoordinatorTerm struct {
	ISU       int64 `json:"isu"`
	YearStart int   `json:"year_start"`
	YearEnd   int   `json:"year_end"`
}

// GIASecretariesParams filters the secretary lists. Zero values are not
// sent, except Limit and Offset. SecretaryISU, StatusID and ShowAs are ignored by the
// routes that do not take them.
type GIASecretariesParams struct {
	Limit, Offset int
	Year          int
	EPID          int64
	GroupID       string
	SecretaryISU  int64
	StatusID      int
	ShowAs        GIARole
	Query         string
}

func (p GIASecretariesParams) query(secretary, showAs, status bool) query {
	v := q().set("limit", p.Limit).set("offset", p.Offset)
	giaSet(v, "year", p.Year)
	giaSet(v, "ep_id", p.EPID)
	v.set("group_id", p.GroupID)
	if secretary {
		giaSet(v, "secretary_isu", p.SecretaryISU)
	}
	if status {
		giaSet(v, "status_id", p.StatusID)
	}
	if showAs {
		v.set("show_as", p.ShowAs)
	}
	return v.set("query", p.Query)
}

// Secretaries returns a page of secretary assignments by programme and group.
// Staff only.
// GET /api/gia/secretaries/list
func (s *GIAService) Secretaries(ctx context.Context, p GIASecretariesParams) (*GIASecretaryList, error) {
	return call[*GIASecretaryList](ctx, s.c, get("api/gia/secretaries/list", p.query(true, true, true)))
}

// GeneralStateSecretaries returns a page of students with their ОГЭК secretaries.
// Staff only.
// GET /api/gia/secretaries/list_ogek
func (s *GIAService) GeneralStateSecretaries(ctx context.Context, p GIASecretariesParams) (*GIAGeneralStateSecretaryList, error) {
	return call[*GIAGeneralStateSecretaryList](ctx, s.c, get("api/gia/secretaries/list_ogek", p.query(true, false, true)))
}

// MyStudents returns the current secretary's contact data and assigned students.
// Staff only.
// GET /api/gia/secretaries/my_students
func (s *GIAService) MyStudents(ctx context.Context, p GIASecretariesParams) (*GIAMyStudents, error) {
	return call[*GIAMyStudents](ctx, s.c, get("api/gia/secretaries/my_students", p.query(false, false, true)))
}

// UnassignedStudents returns a page of students without a secretary.
// Staff only.
// GET /api/gia/secretaries/students/unassigned
func (s *GIAService) UnassignedStudents(ctx context.Context, p GIASecretariesParams) (*GIAUnassignedStudents, error) {
	return call[*GIAUnassignedStudents](ctx, s.c, get("api/gia/secretaries/students/unassigned", p.query(false, false, false)))
}

// AssignSecretary assigns a secretary (by ISU number) to students.
// Staff only.
// POST /api/gia/secretaries
func (s *GIAService) AssignSecretary(ctx context.Context, secretaryISU int64, studentIDs ...int64) error {
	body := struct {
		SecretaryISU int64   `json:"secretary_isu"`
		StudentID    []int64 `json:"student_id"`
	}{secretaryISU, studentIDs}
	if body.StudentID == nil {
		body.StudentID = []int64{}
	}
	return exec(ctx, s.c, post("api/gia/secretaries", body))
}

// ApproveSecretaries approves secretary assignments.
// Staff only.
// POST /api/gia/secretaries/approve
func (s *GIAService) ApproveSecretaries(ctx context.Context, secretaryIDs ...int64) error {
	body := struct {
		SecretaryID []int64 `json:"secretary_id"`
	}{secretaryIDs}
	if body.SecretaryID == nil {
		body.SecretaryID = []int64{}
	}
	return exec(ctx, s.c, post("api/gia/secretaries/approve", body))
}

// DeclineSecretaries declines secretary assignments with a reason.
// Staff only.
// DELETE /api/gia/secretaries/decline
func (s *GIAService) DeclineSecretaries(ctx context.Context, reason string, secretaryIDs ...int64) error {
	body := struct {
		SecretaryID  []int64 `json:"secretary_id"`
		RejectReason string  `json:"reject_reason"`
	}{secretaryIDs, reason}
	if body.SecretaryID == nil {
		body.SecretaryID = []int64{}
	}
	return exec(ctx, s.c, del("api/gia/secretaries/decline", body))
}

// Coordinators returns the ОГЭК coordinators of the academic year starting in year.
// Staff only.
// GET /api/gia/general_state_coordinator/list
func (s *GIAService) Coordinators(ctx context.Context, year int) ([]GIACoordinator, error) {
	return call[[]GIACoordinator](ctx, s.c, get("api/gia/general_state_coordinator/list", q().set("year", year)))
}

// AddCoordinators appoints ОГЭК coordinators.
// Staff only.
// POST /api/gia/general_state_coordinator
func (s *GIAService) AddCoordinators(ctx context.Context, terms ...GIACoordinatorTerm) error {
	if terms == nil {
		terms = []GIACoordinatorTerm{}
	}
	return exec(ctx, s.c, post("api/gia/general_state_coordinator", terms))
}

// RemoveCoordinators removes ОГЭК coordinators.
// Staff only.
// DELETE /api/gia/general_state_coordinator
func (s *GIAService) RemoveCoordinators(ctx context.Context, terms ...GIACoordinatorTerm) error {
	if terms == nil {
		terms = []GIACoordinatorTerm{}
	}
	return exec(ctx, s.c, del("api/gia/general_state_coordinator", terms))
}
