package myitmo

import (
	"context"
)

// GIAStatusEntity selects the dictionary of [GIAService.Statuses].
type GIAStatusEntity = string

// Status dictionaries of the GIA staff section.
const (
	GIAEntityCommittees    GIAStatusEntity = "committees"
	GIAEntityPresentations GIAStatusEntity = "presentations"
	GIAEntityChairman      GIAStatusEntity = "chairman"
	GIAEntityExperts       GIAStatusEntity = "experts"
	GIAEntityOrders        GIAStatusEntity = "orders"
	GIAEntitySecretaries   GIAStatusEntity = "secretaries"
	// GIAEntityReservations are room reservations: 14 booked, 15 pending, 16 declined.
	GIAEntityReservations GIAStatusEntity = "reservations"
	GIAEntityEUSupplement GIAStatusEntity = "attachment_eu"
)

// GIAEPStudents is a programme with its students grouped by study group.
type GIAEPStudents struct {
	EPID             int64        `json:"ep_id"`
	EPName           string       `json:"ep_name"`
	EPEnrollmentYear int          `json:"ep_enrollment_year"`
	Groups           []GIAEPGroup `json:"groups"`
}

// GIAEPGroup is a study group of [GIAEPStudents].
type GIAEPGroup struct {
	GroupID string `json:"group_id"`
	// IsSecretarySingle reports whether the whole group has one secretary.
	IsSecretarySingle bool                `json:"is_secretary_single"`
	StudentsCount     int                 `json:"students_count"`
	StudentsData      []GIAEPGroupStudent `json:"students_data"`
}

// GIAEPGroupStudent is a student with the assigned secretary.
type GIAEPGroupStudent struct {
	StudentID         int64  `json:"student_id"`
	StudentISU        int64  `json:"student_isu"`
	StudentSurname    string `json:"student_surname"`
	StudentName       string `json:"student_name"`
	StudentSecondName string `json:"student_second_name"`
	// SecretaryISU is 0 when no secretary is assigned.
	SecretaryISU        int64  `json:"secretary_isu"`
	SecretarySurname    string `json:"secretary_surname"`
	SecretaryName       string `json:"secretary_name"`
	SecretarySecondName string `json:"secretary_second_name"`
}

// UserStatus returns the role flags of the current user in the GIA staff section.
// Staff only.
// GET /api/gia/users/status
func (s *GIAService) UserStatus(ctx context.Context) (*GIAUserStatus, error) {
	return call[*GIAUserStatus](ctx, s.c, get("api/gia/users/status", nil))
}

// PeopleSearch searches staff by name or ISU number. The result shape is not
// known; it is returned as is.
// Staff only.
// GET /api/gia/people/search
func (s *GIAService) PeopleSearch(ctx context.Context, query string) (RawJSON, error) {
	return call[RawJSON](ctx, s.c, get("api/gia/people/search", q().set("query", query)))
}

// SetSNILS saves the SNILS number of a person, e.g. for a student
// supplement.
// POST /api/gia/people/snils
func (s *GIAService) SetSNILS(ctx context.Context, isu int64, snils string) error {
	body := struct {
		ISU   int64  `json:"isu"`
		SNILS string `json:"snils"`
	}{isu, snils}
	return exec(ctx, s.c, post("api/gia/people/snils", body))
}

// EPStudents returns the students of a programme grouped by study group, with
// their secretaries. groupID is optional; the first entry is the relevant one.
// Staff only.
// GET /api/gia/ep/students
func (s *GIAService) EPStudents(ctx context.Context, epID int64, groupID string) ([]GIAEPStudents, error) {
	return call[[]GIAEPStudents](ctx, s.c, get("api/gia/ep/students", q().set("ep_id", epID).set("group_id", groupID)))
}

// EPGroups returns the study group ids of a programme.
// Staff only.
// GET /api/gia/ep/{ep_id}/groups
func (s *GIAService) EPGroups(ctx context.Context, epID int64) ([]string, error) {
	return call[[]string](ctx, s.c, get("api/gia/ep/"+id(epID)+"/groups", nil))
}

// Years returns the academic years of the year filters.
// Staff only.
// GET /api/gia/filters/years
func (s *GIAService) Years(ctx context.Context) ([]GIAYearOption, error) {
	return call[[]GIAYearOption](ctx, s.c, get("api/gia/filters/years", nil))
}

// FilterPrograms returns the programmes of the filters; showAs is optional.
// Staff only.
// GET /api/gia/filters/ep
func (s *GIAService) FilterPrograms(ctx context.Context, showAs GIARole) ([]IDValue, error) {
	return call[[]IDValue](ctx, s.c, get("api/gia/filters/ep", q().set("show_as", showAs)))
}

// FilterGroups returns the study groups of the filters; showAs is optional.
// Staff only.
// GET /api/gia/filters/groups
func (s *GIAService) FilterGroups(ctx context.Context, showAs GIARole) ([]GIAGroupValue, error) {
	return call[[]GIAGroupValue](ctx, s.c, get("api/gia/filters/groups", q().set("show_as", showAs)))
}

// FilterDirections returns the fields of study of the filters; both
// arguments are optional (0 and "").
// Staff only.
// GET /api/gia/filters/directions
func (s *GIAService) FilterDirections(ctx context.Context, epID int64, showAs GIARole) ([]GIADirectionOption, error) {
	return call[[]GIADirectionOption](ctx, s.c, get("api/gia/filters/directions", giaSet(q(), "ep_id", epID).set("show_as", showAs)))
}

// FilterSecretaries returns the secretaries of the filters; showAs is optional.
// Staff only.
// GET /api/gia/filters/secretaries
func (s *GIAService) FilterSecretaries(ctx context.Context, showAs GIARole) ([]GIAPersonOption, error) {
	return call[[]GIAPersonOption](ctx, s.c, get("api/gia/filters/secretaries", q().set("show_as", showAs)))
}

// FilterStudents returns the students of the filters.
// Staff only.
// GET /api/gia/filters/students
func (s *GIAService) FilterStudents(ctx context.Context) ([]IDValue, error) {
	return call[[]IDValue](ctx, s.c, get("api/gia/filters/students", nil))
}

// Employees searches employees, the candidates for secretaries and ОГЭК coordinators.
// Staff only.
// GET /api/gia/filters/employed
func (s *GIAService) Employees(ctx context.Context, query string) ([]GIAPersonOption, error) {
	return call[[]GIAPersonOption](ctx, s.c, get("api/gia/filters/employed", q().set("query", query)))
}

// Statuses returns the status dictionary of an entity; showAs is optional.
// Staff only.
// GET /api/gia/filters/statuses
func (s *GIAService) Statuses(ctx context.Context, entity GIAStatusEntity, showAs GIARole) ([]GIAStatus, error) {
	return call[[]GIAStatus](ctx, s.c, get("api/gia/filters/statuses", q().set("entity", entity).set("show_as", showAs)))
}

// FilterExperts returns the experts of the filters.
// Staff only.
// GET /api/gia/filters/experts
func (s *GIAService) FilterExperts(ctx context.Context) ([]IDValue, error) {
	return call[[]IDValue](ctx, s.c, get("api/gia/filters/experts", nil))
}

// FilterFaculties returns the faculties of the order filters.
// Staff only.
// GET /api/gia/filters/faculties
func (s *GIAService) FilterFaculties(ctx context.Context) ([]GIANamedOption, error) {
	return call[[]GIANamedOption](ctx, s.c, get("api/gia/filters/faculties", nil))
}

// FilterEduLevels returns the education levels of the order filters.
// Staff only.
// GET /api/gia/filters/edu-levels
func (s *GIAService) FilterEduLevels(ctx context.Context) ([]GIANamedOption, error) {
	return call[[]GIANamedOption](ctx, s.c, get("api/gia/filters/edu-levels", nil))
}

// OrderTypes returns the order types.
// Staff only.
// GET /api/gia/filters/order-types
func (s *GIAService) OrderTypes(ctx context.Context) ([]GIAIDName, error) {
	return call[[]GIAIDName](ctx, s.c, get("api/gia/filters/order-types", nil))
}

// ChairmanFaculties returns the faculties of the chairman and supplement filters.
// Staff only.
// GET /api/gia/filters/chairman/faculties
func (s *GIAService) ChairmanFaculties(ctx context.Context) ([]IDValue, error) {
	return call[[]IDValue](ctx, s.c, get("api/gia/filters/chairman/faculties", nil))
}

// ChairmanDirections returns the fields of study of the chairman filters.
// Staff only.
// GET /api/gia/filters/chairman/dir
func (s *GIAService) ChairmanDirections(ctx context.Context) ([]IDValue, error) {
	return call[[]IDValue](ctx, s.c, get("api/gia/filters/chairman/dir", nil))
}

// ChairmanPrograms returns the programmes of the chairman filters.
// Staff only.
// GET /api/gia/filters/chairman/ep
func (s *GIAService) ChairmanPrograms(ctx context.Context) ([]IDValue, error) {
	return call[[]IDValue](ctx, s.c, get("api/gia/filters/chairman/ep", nil))
}

// AttachmentCredits returns the distinct credit totals of the supplement list filter.
// Staff only.
// GET /api/gia/filters/attachment/credits
func (s *GIAService) AttachmentCredits(ctx context.Context) ([]float64, error) {
	res, err := call[struct {
		Credits []float64 `json:"credits"`
	}](ctx, s.c, get("api/gia/filters/attachment/credits", nil))
	return res.Credits, err
}

// AcademicTitles returns the academic titles dictionary.
// Staff only.
// GET /api/gia/references/academic-titles
func (s *GIAService) AcademicTitles(ctx context.Context) ([]GIARefItem, error) {
	return call[[]GIARefItem](ctx, s.c, get("api/gia/references/academic-titles", nil))
}

// DegreeTypes returns the academic degrees dictionary.
// Staff only.
// GET /api/gia/references/degree-types
func (s *GIAService) DegreeTypes(ctx context.Context) ([]GIARefItem, error) {
	return call[[]GIARefItem](ctx, s.c, get("api/gia/references/degree-types", nil))
}

// ExpertEduLevels returns the education levels of expert cards.
// Staff only.
// GET /api/gia/references/edu-levels
func (s *GIAService) ExpertEduLevels(ctx context.Context) ([]GIAIDName, error) {
	return call[[]GIAIDName](ctx, s.c, get("api/gia/references/edu-levels", nil))
}

// Grades returns the criterion grades (assessment=false) or the final
// assessment grades (assessment=true).
// Staff only.
// GET /api/gia/references/grades
func (s *GIAService) Grades(ctx context.Context, assessment bool) ([]GIAGrade, error) {
	return call[[]GIAGrade](ctx, s.c, get("api/gia/references/grades", q().set("assessment", assessment)))
}

// OrganizationRegistries returns the organisation registry types; id 1 means
// registered in Russia, where the INN is required.
// Staff only.
// GET /api/gia/references/organization-registries
func (s *GIAService) OrganizationRegistries(ctx context.Context) ([]GIAIDName, error) {
	return call[[]GIAIDName](ctx, s.c, get("api/gia/references/organization-registries", nil))
}

// OrganizationStatuses returns the organisation status dictionary.
// Staff only.
// GET /api/gia/references/organization-statuses
func (s *GIAService) OrganizationStatuses(ctx context.Context) ([]GIARefItem, error) {
	return call[[]GIARefItem](ctx, s.c, get("api/gia/references/organization-statuses", nil))
}
