package myitmo

import (
	"context"
	"time"
)

// Expert card statuses (status_id).
const (
	GIAExpertStatusDraft      = 1
	GIAExpertStatusOnApproval = 2
	GIAExpertStatusRejected   = 3
	GIAExpertStatusApproved   = 4
)

// Chairman appointment statuses sent in [GIAChairmanInput.StatusID].
const (
	GIAChairmanStatusDraft      = 1
	GIAChairmanStatusOnApproval = 2
)

// GIAExpertListItem is a row of the expert list.
type GIAExpertListItem struct {
	ID            int64     `json:"id"`
	FIO           string    `json:"fio"`
	AcademicTitle string    `json:"academic_title"`
	DegreeType    string    `json:"degree_type"`
	JobTitle      string    `json:"job_title"`
	WorkPlace     string    `json:"work_place"`
	StatusID      int       `json:"status_id"`
	CreatedAt     time.Time `json:"created_at"`
	CreatedBy     *int64    `json:"created_by"`
	CreatedByFIO  string    `json:"created_by_fio"`
}

// GIAExpertList is a page of experts.
type GIAExpertList struct {
	Experts []GIAExpertListItem `json:"experts"`
	Count   int                 `json:"count"`
}

// GIAExpert is an expert card: a candidate member or chairman of a state
// examination committee (ГЭК).
type GIAExpert struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Surname     string `json:"surname"`
	SecondName  string `json:"second_name"`
	Email       string `json:"email"`
	PhoneNumber string `json:"phone_number"`
	// Gender is "male" or "female".
	Gender          string               `json:"gender"`
	Education       string               `json:"education"`
	Speciality      string               `json:"speciality"`
	AcademicTitleID *int64               `json:"academic_title_id"`
	AcademicTitle   string               `json:"academic_title"`
	EduLevelID      *int64               `json:"edu_level_id"`
	Degrees         []GIAExpertDegree    `json:"degrees"`
	WorkPlaces      []GIAExpertWorkPlace `json:"work_places"`
	// StatusID is one of the GIAExpertStatus constants.
	StatusID      int    `json:"status_id"`
	Approved      *bool  `json:"approved"`
	Comment       string `json:"comment"`
	CommentAuthor string `json:"comment_author"`
	CreatedBy     *int64 `json:"created_by"`
}

// GIAExpertDegree is an academic degree of an expert.
type GIAExpertDegree struct {
	DegreeTypeID int64     `json:"degree_type_id"`
	DegreeType   string    `json:"degree_type,omitzero"`
	Number       string    `json:"number"`
	Series       string    `json:"series"`
	Date         time.Time `json:"date"`
}

// GIAExpertWorkPlace is a work place of an expert.
type GIAExpertWorkPlace struct {
	// OrganizationRegistryID 1 means registered in Russia; INN is then required.
	OrganizationRegistryID int64  `json:"organization_registry_id"`
	INN                    int64  `json:"inn,omitzero"`
	OrganizationName       string `json:"organization_name"`
	LineOfBusiness         string `json:"line_of_business"`
	JobTitle               string `json:"job_title"`
	OrganizationStatusID   int64  `json:"organization_status_id"`
	OrganizationStatus     string `json:"organization_status,omitzero"`
	// Stages is an Editor.js document: the main stages of work and scientific activity.
	Stages RawJSON `json:"stages,omitzero"`
}

// GIAExpertInput is the body of expert create, update and send-for-approval.
type GIAExpertInput struct {
	Name        string `json:"name"`
	Surname     string `json:"surname"`
	SecondName  string `json:"second_name"`
	Email       string `json:"email"`
	PhoneNumber string `json:"phone_number"`
	// Gender is "male" or "female"; nil sends null.
	Gender          *string              `json:"gender"`
	Education       string               `json:"education"`
	Speciality      string               `json:"speciality"`
	AcademicTitleID *int64               `json:"academic_title_id"`
	EduLevelID      *int64               `json:"edu_level_id"`
	Degrees         []GIAExpertDegree    `json:"degrees"`
	WorkPlaces      []GIAExpertWorkPlace `json:"work_places"`
	// StatusID is GIAExpertStatusDraft or GIAExpertStatusOnApproval.
	StatusID int `json:"status_id"`
}

// GIAExpertSimple is an expert with the job titles of their work places.
type GIAExpertSimple struct {
	ExpertID  int64               `json:"expert_id"`
	JobTitles []GIAExpertJobTitle `json:"job_titles"`
}

// GIAExpertJobTitle is a work place of [GIAExpertSimple].
type GIAExpertJobTitle struct {
	WorkPlaceID int64  `json:"work_place_id"`
	JobTitle    string `json:"job_title"`
	// Stages is an Editor.js document; approval needs it to be non-empty.
	Stages RawJSON `json:"stages"`
}

// GIAOrganization is an organisation found by INN.
type GIAOrganization struct {
	OrganizationName      string `json:"organization_name"`
	OrganizationOkvedName string `json:"organization_okved_name"`
}

// GIAExpertsParams filters [GIAService.Experts].
type GIAExpertsParams struct {
	Query         string
	Limit, Offset int
	// Status is an expert status_id; 0 means any.
	Status int
}

// Experts returns a page of experts.
// Staff only.
// GET /api/gia/experts
func (s *GIAService) Experts(ctx context.Context, p GIAExpertsParams) (*GIAExpertList, error) {
	v := q().set("query", p.Query).set("limit", p.Limit).set("offset", p.Offset)
	return call[*GIAExpertList](ctx, s.c, get("api/gia/experts", giaSet(v, "status", p.Status)))
}

// CreateExpert creates an expert card and returns its id. Create a draft
// and then call [GIAService.SubmitExpert].
// Staff only.
// POST /api/gia/experts
func (s *GIAService) CreateExpert(ctx context.Context, in GIAExpertInput) (int64, error) {
	res, err := call[struct {
		ExpertID int64 `json:"expert_id"`
	}](ctx, s.c, post("api/gia/experts", in))
	return res.ExpertID, err
}

// Expert returns an expert card.
// Staff only.
// GET /api/gia/experts/{expert_id}
func (s *GIAService) Expert(ctx context.Context, expertID int64) (*GIAExpert, error) {
	return call[*GIAExpert](ctx, s.c, get("api/gia/experts/"+id(expertID), nil))
}

// UpdateExpert updates an expert card.
// Staff only.
// PATCH /api/gia/experts/{expert_id}
func (s *GIAService) UpdateExpert(ctx context.Context, expertID int64, in GIAExpertInput) error {
	return exec(ctx, s.c, patch("api/gia/experts/"+id(expertID), in))
}

// DeleteExpert deletes an expert card.
// Staff only.
// DELETE /api/gia/experts/{expert_id}
func (s *GIAService) DeleteExpert(ctx context.Context, expertID int64) error {
	return exec(ctx, s.c, del("api/gia/experts/"+id(expertID), nil))
}

// ExpertExists reports whether an expert with this e-mail already exists.
// The server answer is treated as a truthy value.
// Staff only.
// GET /api/gia/experts/exists
func (s *GIAService) ExpertExists(ctx context.Context, email string) (bool, error) {
	res, err := call[RawJSON](ctx, s.c, get("api/gia/experts/exists", q().set("email", email)))
	return giaTruthy(res), err
}

// Organization looks up an organisation by INN.
// Staff only.
// GET /api/gia/experts/organization
func (s *GIAService) Organization(ctx context.Context, inn string) (*GIAOrganization, error) {
	return call[*GIAOrganization](ctx, s.c, get("api/gia/experts/organization", q().set("inn", inn)))
}

// ExpertsSimple returns all experts with the job titles of their work places.
// Staff only.
// GET /api/gia/experts/simple
func (s *GIAService) ExpertsSimple(ctx context.Context) ([]GIAExpertSimple, error) {
	return call[[]GIAExpertSimple](ctx, s.c, get("api/gia/experts/simple", nil))
}

// ApproveExpert approves an expert card.
// Staff only.
// POST /api/gia/experts/approve/{expert_id}
func (s *GIAService) ApproveExpert(ctx context.Context, expertID int64) error {
	return exec(ctx, s.c, post("api/gia/experts/approve/"+id(expertID), nil))
}

// RejectExpert rejects an expert card with a comment.
// Staff only.
// DELETE /api/gia/experts/reject/{expert_id}
func (s *GIAService) RejectExpert(ctx context.Context, expertID int64, message string) error {
	return exec(ctx, s.c, del("api/gia/experts/reject/"+id(expertID), giaMessage{message}))
}

// SubmitExpert sends an expert card for approval with its full content;
// a zero StatusID is set to GIAExpertStatusOnApproval.
// Staff only.
// POST /api/gia/experts/under-approval/{expert_id}
func (s *GIAService) SubmitExpert(ctx context.Context, expertID int64, in GIAExpertInput) error {
	if in.StatusID == 0 {
		in.StatusID = GIAExpertStatusOnApproval
	}
	return exec(ctx, s.c, post("api/gia/experts/under-approval/"+id(expertID), in))
}

// giaMessage is the {message} body of reject and decline calls.
type giaMessage struct {
	Message string `json:"message"`
}

// giaComment is the {comment} body of reject calls.
type giaComment struct {
	Comment string `json:"comment"`
}

// GIAChairman is a chairman appointment of a state examination committee
// for an educational programme.
type GIAChairman struct {
	ChairmanID         int64  `json:"chairman_id"`
	ExpertID           int64  `json:"expert_id"`
	ChairmanSurname    string `json:"chairman_surname"`
	ChairmanName       string `json:"chairman_name"`
	ChairmanSecondName string `json:"chairman_second_name"`
	AcademicTitleName  string `json:"academic_title_name"`
	// AcademicTitles and Degrees are present; shapes not known.
	AcademicTitles       RawJSON   `json:"academic_titles,omitzero"`
	Degrees              RawJSON   `json:"degrees,omitzero"`
	JobTitle             string    `json:"job_title"`
	WorkPlace            string    `json:"work_place"`
	WorkPlaceID          int64     `json:"work_place_id"`
	WorkPlaceStatusName  string    `json:"work_place_status_name"`
	TypeName             string    `json:"type_name"`
	EPID                 int64     `json:"ep_id"`
	EPName               string    `json:"ep_name"`
	EPEnrollmentYear     int       `json:"ep_enrollment_year"`
	EPManagerSurname     string    `json:"ep_manager_surname"`
	EPManagerName        string    `json:"ep_manager_name"`
	EPManagerSecondName  string    `json:"ep_manager_second_name"`
	EPManagerISU         int64     `json:"ep_manager_isu"`
	FacultyManagerISU    int64     `json:"faculty_manager_isu"`
	FacultyShortName     string    `json:"faculty_short_name"`
	FacultyName          string    `json:"faculty_name"`
	DirectionID          int64     `json:"direction_id"`
	DirectionCode        string    `json:"direction_code"`
	DirectionName        string    `json:"direction_name"`
	StatusID             int       `json:"status_id"`
	StatusName           string    `json:"status_name"`
	Comment              string    `json:"comment"`
	RejectedBySurname    string    `json:"rejected_by_surname"`
	RejectedByName       string    `json:"rejected_by_name"`
	RejectedBySecondName string    `json:"rejected_by_second_name"`
	RejectorRole         string    `json:"rejector_role"`
	WhoCreatedFIO        string    `json:"who_created_fio"`
	WhoCreatedISU        int64     `json:"who_created_isu"`
	CreatedAt            time.Time `json:"created_at"`
	IsGeneralState       bool      `json:"is_general_state"`
}

// GIAChairmanList is a page of chairman appointments. The server spells the
// list key with a capital C.
type GIAChairmanList struct {
	Chairman   []GIAChairman `json:"Chairman"`
	TotalCount int           `json:"total_count"`
}

// GIAEPWithoutChairman is a programme that has no chairman yet.
type GIAEPWithoutChairman struct {
	EPID          int64  `json:"ep_id"`
	EPName        string `json:"ep_name"`
	FacultyName   string `json:"faculty_name"`
	DirectionCode string `json:"direction_code"`
	// Directions is a list whose item shape is not known.
	Directions RawJSON `json:"directions,omitzero"`
}

// GIAEPWithoutChairmanList is a page of programmes without a chairman.
type GIAEPWithoutChairmanList struct {
	Programs   []GIAEPWithoutChairman `json:"programs"`
	TotalCount int                    `json:"total_count"`
}

// GIAChairmanInput is the body of chairman create and update.
type GIAChairmanInput struct {
	ExpertID    int64   `json:"expert_id"`
	EPID        int64   `json:"ep_id"`
	DirID       []int64 `json:"dir_id"`
	WorkPlaceID int64   `json:"work_place_id"`
	// StatusID is GIAChairmanStatusDraft or GIAChairmanStatusOnApproval;
	// updating with GIAChairmanStatusDraft revokes a sent appointment.
	StatusID int `json:"status_id"`
}

// GIAChairmenParams filters the chairman lists. Zero values are not sent,
// except Limit and Offset.
type GIAChairmenParams struct {
	Limit, Offset int
	Year          int
	EPID          int64
	DirID         int64
	FacultyID     int64
	// PersonID is an expert id.
	PersonID int64
	StatusID int
	Query    string
}

func (p GIAChairmenParams) query() query {
	v := q().set("limit", p.Limit).set("offset", p.Offset)
	giaSet(v, "year", p.Year)
	giaSet(v, "ep_id", p.EPID)
	giaSet(v, "dir_id", p.DirID)
	giaSet(v, "faculty_id", p.FacultyID)
	giaSet(v, "person_id", p.PersonID)
	giaSet(v, "status_id", p.StatusID)
	return v.set("query", p.Query)
}

// Chairmen returns a page of chairman appointments.
// Staff only.
// GET /api/gia/chairman/list
func (s *GIAService) Chairmen(ctx context.Context, p GIAChairmenParams) (*GIAChairmanList, error) {
	return call[*GIAChairmanList](ctx, s.c, get("api/gia/chairman/list", p.query()))
}

// ChairmenSimple returns all chairmen with their job titles and work stages.
// Staff only.
// GET /api/gia/chairman/list/simple
func (s *GIAService) ChairmenSimple(ctx context.Context) ([]GIAExpertSimple, error) {
	return call[[]GIAExpertSimple](ctx, s.c, get("api/gia/chairman/list/simple", nil))
}

// ProgramsWithoutChairman returns the programmes that have no chairman yet.
// Staff only.
// GET /api/gia/chairman/ep_missing
func (s *GIAService) ProgramsWithoutChairman(ctx context.Context, p GIAChairmenParams) (*GIAEPWithoutChairmanList, error) {
	return call[*GIAEPWithoutChairmanList](ctx, s.c, get("api/gia/chairman/ep_missing", p.query()))
}

// ChairmenOrder downloads the generated draft order on chairmen (DOCX).
// Staff only.
// GET /api/gia/chairman/order
func (s *GIAService) ChairmenOrder(ctx context.Context) (*File, error) {
	return download(ctx, s.c, get("api/gia/chairman/order", nil))
}

// CreateChairman creates a chairman appointment.
// Staff only.
// POST /api/gia/chairman
func (s *GIAService) CreateChairman(ctx context.Context, in GIAChairmanInput) error {
	return exec(ctx, s.c, post("api/gia/chairman", in))
}

// UpdateChairman edits a chairman appointment or changes its status.
// Staff only.
// PATCH /api/gia/chairman/{chairman_id}
func (s *GIAService) UpdateChairman(ctx context.Context, chairmanID int64, in GIAChairmanInput) error {
	return exec(ctx, s.c, patch("api/gia/chairman/"+id(chairmanID), in))
}

// DeleteChairman deletes a chairman appointment.
// Staff only.
// DELETE /api/gia/chairman/{chairman_id}
func (s *GIAService) DeleteChairman(ctx context.Context, chairmanID int64) error {
	return exec(ctx, s.c, del("api/gia/chairman/"+id(chairmanID), nil))
}

// ApproveChairman approves a chairman appointment.
// Staff only.
// POST /api/gia/chairman/{chairman_id}/approve
func (s *GIAService) ApproveChairman(ctx context.Context, chairmanID int64) error {
	return exec(ctx, s.c, post("api/gia/chairman/"+id(chairmanID)+"/approve", nil))
}

// DeclineChairman declines a chairman appointment with a comment.
// Staff only.
// POST /api/gia/chairman/{chairman_id}/decline
func (s *GIAService) DeclineChairman(ctx context.Context, chairmanID int64, message string) error {
	return exec(ctx, s.c, post("api/gia/chairman/"+id(chairmanID)+"/decline", giaMessage{message}))
}

// ChairmanNomination downloads the nomination document (представление) of a chairman (PDF).
// Staff only.
// GET /api/gia/chairman/{chairman_id}/represent
func (s *GIAService) ChairmanNomination(ctx context.Context, chairmanID int64) (*File, error) {
	return download(ctx, s.c, get("api/gia/chairman/"+id(chairmanID)+"/represent", nil))
}
