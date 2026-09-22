package myitmo

import (
	"context"
	"time"
)

// GIACommitteeMember is a committee member: an internal employee (ExpertISU
// set) or an external expert (ExpertID set).
type GIACommitteeMember struct {
	ExpertISU        int64  `json:"expert_isu"`
	ExpertID         int64  `json:"expert_id"`
	ISU              int64  `json:"isu"`
	ExpertSurname    string `json:"expert_surname"`
	ExpertName       string `json:"expert_name"`
	ExpertSecondName string `json:"expert_second_name"`
	ExpertDegreeType string `json:"expert_degree_type"`
	ExpertJobTitle   string `json:"expert_job_title"`
	ExpertWorkPlace  string `json:"expert_work_place"`
}

// External reports whether the member is an external expert.
func (m GIACommitteeMember) External() bool { return m.ExpertID != 0 }

// GIACommitteeListItem is a committee of the committee lists.
type GIACommitteeListItem struct {
	CommitteeID int64 `json:"committee_id"`
	// CommitteeNumber is a number or a string.
	CommitteeNumber    RawJSON           `json:"committee_number"`
	ChairmanSurname    string            `json:"chairman_surname"`
	ChairmanName       string            `json:"chairman_name"`
	ChairmanSecondName string            `json:"chairman_second_name"`
	EduDirections      []GIAEduDirection `json:"edu_directions"`
	// DirID, DirCode and DirName are a flattened variant of EduDirections.
	DirID      int64     `json:"dir_id"`
	DirCode    string    `json:"dir_code"`
	DirName    string    `json:"dir_name"`
	CreatedAt  time.Time `json:"created_at"`
	StatusID   int       `json:"status_id"`
	StatusName string    `json:"status_name"`
}

// GIACommitteeProgram is a programme with its committees.
type GIACommitteeProgram struct {
	GIACommitteeListItem
	ID               int64                  `json:"id"`
	EPName           string                 `json:"ep_name"`
	EPEnrollmentYear int                    `json:"ep_enrollment_year"`
	Committees       []GIACommitteeListItem `json:"committees"`
}

// GIACommitteeList is a page of committees grouped by programme.
type GIACommitteeList struct {
	Data  []GIACommitteeProgram `json:"data"`
	Total int                   `json:"total"`
}

// GIAGeneralStateCommitteeList is a page of ОГЭК committees.
type GIAGeneralStateCommitteeList struct {
	Committees []GIACommitteeListItem `json:"committees"`
	Total      int                    `json:"total"`
}

// GIACommittee is a state examination committee (состав ГЭК).
type GIACommittee struct {
	CommitteeID int64 `json:"committee_id"`
	// CommitteeNumber is a number or a string.
	CommitteeNumber RawJSON `json:"committee_number"`
	// ExpertID is the expert id of the chairman.
	ExpertID           int64                `json:"expert_id"`
	ChairmanSurname    string               `json:"chairman_surname"`
	ChairmanName       string               `json:"chairman_name"`
	ChairmanSecondName string               `json:"chairman_second_name"`
	ChairmanDegreeType string               `json:"chairman_degree_type"`
	ChairmanJobTitle   string               `json:"chairman_job_title"`
	ChairmanWorkPlace  string               `json:"chairman_work_place"`
	EduPrograms        []GIAEduProgram      `json:"edu_programs"`
	EduDirections      []GIAEduDirection    `json:"edu_directions"`
	InternalExperts    []GIACommitteeMember `json:"internal_experts"`
	ExternalExperts    []GIACommitteeMember `json:"external_experts"`
	StatusID           int                  `json:"status_id"`
	StatusName         string               `json:"status_name"`
	Comment            string               `json:"comment"`
	CommenterFIO       string               `json:"commenter_fio"`
	CommenterRole      string               `json:"commenter_role"`
}

// GIACommitteeInput is the body of committee create and update. A
// committee needs at least four members, two of them external.
type GIACommitteeInput struct {
	ChairmanExpertID int64   `json:"chairman_expert_id"`
	EduProgramIDs    []int64 `json:"edu_program_ids"`
	EduDirectionIDs  []int64 `json:"edu_direction_ids"`
	// InternalExperts are ISU numbers.
	InternalExperts []int64 `json:"internal_experts"`
	// ExternalExperts are expert ids.
	ExternalExperts []int64 `json:"external_experts"`
	// IsScratch saves a draft.
	IsScratch bool `json:"is_scratch"`
}

// GIACommitteeChairman is an approved chairman to pick for a committee.
type GIACommitteeChairman struct {
	ExpertID           int64  `json:"expert_id"`
	ChairmanSurname    string `json:"chairman_surname"`
	ChairmanName       string `json:"chairman_name"`
	ChairmanSecondName string `json:"chairman_second_name"`
	ChairmanDegreeType string `json:"chairman_degree_type"`
	ChairmanJobTitle   string `json:"chairman_job_title"`
	ChairmanWorkPlace  string `json:"chairman_work_place"`
}

// GIACommitteeFilters are the options of the committee list filters.
type GIACommitteeFilters struct {
	YearFilters          []GIAYearFilter `json:"year_filters"`
	StatusFilters        []GIAIDName     `json:"status_filters"`
	EduProgramFilters    []GIAIDName     `json:"edu_program_filters"`
	EduDirectionsFilters []GIAIDName     `json:"edu_directions_filters"`
	FacultyFilters       []IDValue       `json:"faculty_filters"`
	RoleFilters          []GIARoleOption `json:"role_filters"`
}

// GIACommitteesParams filters the committee lists. Zero values are not sent,
// except Limit and Offset.
type GIACommitteesParams struct {
	Limit, Offset  int
	Year           int
	StatusID       int
	EduProgramID   int64
	EduDirectionID int64
	FacultyID      int64
	Role           GIARole
	Query          string
}

func (p GIACommitteesParams) query() query {
	v := q().set("limit", p.Limit).set("offset", p.Offset)
	giaSet(v, "year", p.Year)
	giaSet(v, "status_id", p.StatusID)
	giaSet(v, "edu_program_id", p.EduProgramID)
	giaSet(v, "edu_direction_id", p.EduDirectionID)
	giaSet(v, "faculty_id", p.FacultyID)
	return v.set("role", p.Role).set("query", p.Query)
}

// GIACommitteeDirectionsParams filters [GIAService.CommitteeDirections].
type GIACommitteeDirectionsParams struct {
	Query            string
	Limit            int
	ChairmanExpertID int64
	// EduProgramIDs are sent comma-separated.
	EduProgramIDs []int64
}

// Committees returns a page of committees grouped by programme.
// Staff only.
// GET /api/gia/committees
func (s *GIAService) Committees(ctx context.Context, p GIACommitteesParams) (*GIACommitteeList, error) {
	return call[*GIACommitteeList](ctx, s.c, get("api/gia/committees", p.query()))
}

// CreateCommittee creates a committee.
// Staff only.
// POST /api/gia/committees
func (s *GIAService) CreateCommittee(ctx context.Context, in GIACommitteeInput) error {
	return exec(ctx, s.c, post("api/gia/committees", in))
}

// Committee returns a committee.
// Staff only.
// GET /api/gia/committees/{committee_id}
func (s *GIAService) Committee(ctx context.Context, committeeID int64) (*GIACommittee, error) {
	return call[*GIACommittee](ctx, s.c, get("api/gia/committees/"+id(committeeID), nil))
}

// UpdateCommittee updates a committee.
// Staff only.
// PATCH /api/gia/committees/{committee_id}
func (s *GIAService) UpdateCommittee(ctx context.Context, committeeID int64, in GIACommitteeInput) error {
	return exec(ctx, s.c, patch("api/gia/committees/"+id(committeeID), in))
}

// ApproveCommittee moves a committee to the next approval stage.
// Staff only.
// PATCH /api/gia/committees/{committee_id}/status
func (s *GIAService) ApproveCommittee(ctx context.Context, committeeID int64) error {
	return exec(ctx, s.c, patch("api/gia/committees/"+id(committeeID)+"/status", struct{}{}))
}

// RejectCommittee rejects a committee with a comment.
// Staff only.
// PATCH /api/gia/committees/{committee_id}/status
func (s *GIAService) RejectCommittee(ctx context.Context, committeeID int64, comment string) error {
	return exec(ctx, s.c, patch("api/gia/committees/"+id(committeeID)+"/status", giaComment{comment}))
}

// RevokeCommittee withdraws a committee.
// Staff only.
// POST /api/gia/committees/{committee_id}/revoke
func (s *GIAService) RevokeCommittee(ctx context.Context, committeeID int64) error {
	return exec(ctx, s.c, post("api/gia/committees/"+id(committeeID)+"/revoke", nil))
}

// CommitteeFilters returns the options of the committee list filters; showAs is optional.
// Staff only.
// GET /api/gia/committees/filters
func (s *GIAService) CommitteeFilters(ctx context.Context, showAs GIARole) (*GIACommitteeFilters, error) {
	return call[*GIACommitteeFilters](ctx, s.c, get("api/gia/committees/filters", q().set("show_as", showAs)))
}

// GeneralStateCommittees returns a page of ОГЭК committees.
// Staff only.
// GET /api/gia/committees/general-state-committees
func (s *GIAService) GeneralStateCommittees(ctx context.Context, p GIACommitteesParams) (*GIAGeneralStateCommitteeList, error) {
	return call[*GIAGeneralStateCommitteeList](ctx, s.c, get("api/gia/committees/general-state-committees", p.query()))
}

// GeneralStateCommitteeFilters returns the options of the ОГЭК committee list filters.
// Staff only.
// GET /api/gia/committees/general-state-committees/filters
func (s *GIAService) GeneralStateCommitteeFilters(ctx context.Context) (*GIACommitteeFilters, error) {
	return call[*GIACommitteeFilters](ctx, s.c, get("api/gia/committees/general-state-committees/filters", nil))
}

// CommitteeExperts searches employees and external experts to add to a committee.
// Staff only.
// GET /api/gia/committees/experts
func (s *GIAService) CommitteeExperts(ctx context.Context, query string, limit int) ([]GIACommitteeMember, error) {
	return call[[]GIACommitteeMember](ctx, s.c, get("api/gia/committees/experts", q().set("query", query).set("limit", limit)))
}

// CommitteeChairmen searches approved chairmen for a committee.
// Staff only.
// GET /api/gia/committees/chairmen
func (s *GIAService) CommitteeChairmen(ctx context.Context, query string, limit int) ([]GIACommitteeChairman, error) {
	return call[[]GIACommitteeChairman](ctx, s.c, get("api/gia/committees/chairmen", q().set("query", query).set("limit", limit)))
}

// CommitteePrograms returns the programmes available for a committee of the chairman.
// Staff only.
// GET /api/gia/committees/edu-programs
func (s *GIAService) CommitteePrograms(ctx context.Context, query string, limit int, chairmanExpertID int64) ([]GIAEduProgram, error) {
	v := q().set("query", query).set("limit", limit)
	return call[[]GIAEduProgram](ctx, s.c, get("api/gia/committees/edu-programs", giaSet(v, "chairman_expert_id", chairmanExpertID)))
}

// CommitteeDirections returns the fields of study available for a committee.
// Staff only.
// GET /api/gia/committees/edu-directions
func (s *GIAService) CommitteeDirections(ctx context.Context, p GIACommitteeDirectionsParams) ([]GIAEduDirection, error) {
	v := q().set("query", p.Query).set("limit", p.Limit)
	giaSet(v, "chairman_expert_id", p.ChairmanExpertID)
	v.set("edu_program_ids", giaJoin(p.EduProgramIDs))
	return call[[]GIAEduDirection](ctx, s.c, get("api/gia/committees/edu-directions", v))
}
