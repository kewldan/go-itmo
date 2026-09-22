package myitmo

import (
	"context"
	"encoding/json/v2"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

// VacationService is employee vacations (/api/vacation). The whole module is
// for university employees; management and agreement routes need the rights
// reported by [VacationService.Access].
type VacationService struct{ c *Client }

// VacationStatus is the status_id of a vacation application. Human-readable
// names come from [VacationFilters.Statuses]; 2, 5 (pending of another kind)
// and negative values (rejected or cancelled) have no constant.
type VacationStatus int

// Vacation statuses with a known meaning.
const (
	// VacationStatusDraft is a created application that was not sent yet; it can be deleted.
	VacationStatusDraft VacationStatus = 0
	// VacationStatusPending awaits signature or approval; the owner can cancel it.
	VacationStatusPending VacationStatus = 1
	// VacationStatusSigned is a digitally signed application.
	VacationStatusSigned VacationStatus = 3
	// VacationStatusPaper is a paper application; it can be deleted.
	VacationStatusPaper VacationStatus = 4
)

// VacationType is the type_id of a vacation; names come from [VacationService.Types].
type VacationType int

// Vacation types with a known meaning.
const (
	// VacationTypePaid is the annual paid vacation.
	VacationTypePaid VacationType = 1
	// VacationTypeUnpaid is unpaid leave ("UTO"); create it with [VacationService.CreateUnpaid].
	VacationTypeUnpaid VacationType = 2
)

// VacationReasonPersonal is the unpaid-leave reason id that requires a free-text
// [VacationReason.PersonalReason].
const VacationReasonPersonal = 7

// VacationDayType is the type of a production-calendar day.
type VacationDayType int

// Production-calendar day types.
const (
	VacationDayWeekend VacationDayType = 1
	// VacationDayHoliday is a public holiday; holidays are not counted as vacation days.
	VacationDayHoliday VacationDayType = 2
)

// VacationAccess is the current user's rights in the vacation module.
type VacationAccess struct {
	// Access reports whether the user may use the module at all.
	Access bool `json:"access"`
	// EVacation allows electronic (digitally signed) applications; otherwise paper only.
	EVacation         bool `json:"e_vacation"`
	Admin             bool `json:"admin"`
	Supervisor        bool `json:"supervisor"`
	HR                bool `json:"hr"`
	VacCreationAccess bool `json:"vac_creation_access"`
	Timekeeper        bool `json:"timekeeper"`
	// Employer allows signing agreements.
	Employer bool `json:"employer"`
}

// VacationFilters are the dictionaries of the vacation filters.
type VacationFilters struct {
	Positions   []IDValue `json:"positions"`
	Departments []IDValue `json:"departments"`
	// Statuses maps status_id to its name.
	Statuses []IDValue              `json:"statuses"`
	Projects []VacationFilterOption `json:"projects"`
	// Category lists staff categories; filters send the Value.
	Category []VacationFilterOption `json:"category"`
}

// VacationFilterOption is a dictionary entry whose id type is not known.
type VacationFilterOption struct {
	// ID is a number or a string.
	ID    RawJSON `json:"id"`
	Value string  `json:"value"`
}

// VacationPosition is one of the current user's employments.
type VacationPosition struct {
	EmployeeID     int64  `json:"employee_id"`
	Name           string `json:"name"`
	DepartmentName string `json:"department_name"`
	ProjectName    string `json:"project_name"`
}

// VacationPerson is a people-search result.
type VacationPerson struct {
	ISU  int64  `json:"isu"`
	Name string `json:"name"`
	FIO  string `json:"fio"`
}

// VacationCalendarDay is a non-working day of the production calendar.
type VacationCalendarDay struct {
	Date Date            `json:"date"`
	Type VacationDayType `json:"type"`
}

// VacationBalance is the vacation-day balance of one employment.
type VacationBalance struct {
	EmployeeID int64 `json:"employee_id"`
	// Remaining is the number of accumulated days available.
	Remaining float64 `json:"remaining"`
	// Normative is the annual norm of days.
	Normative float64 `json:"normative"`
}

// VacationEmployment is the employment a vacation belongs to.
type VacationEmployment struct {
	EmployeeID     int64  `json:"employee_id"`
	PositionName   string `json:"position_name"`
	DepartmentName string `json:"department_name"`
	ProjectName    string `json:"project_name"`
	// ProjectID is the project code, a number or a string.
	ProjectID RawJSON `json:"project_id,omitzero"`
}

// Vacation is an application of the current user.
type Vacation struct {
	VacationID int64 `json:"vacation_id"`
	ID         int64 `json:"id"`
	// VacationType is the human-readable type name.
	VacationType string         `json:"vacation_type"`
	DateStart    Date           `json:"date_start"`
	DateEnd      Date           `json:"date_end"`
	Duration     int            `json:"duration"`
	StatusID     VacationStatus `json:"status_id"`
	// TaskID and SignatureID identify the signing task; zero when there is none.
	TaskID         FlexInt `json:"task_id"`
	SignatureID    FlexInt `json:"signature_id"`
	NeedsSignature bool    `json:"needs_signature"`
	// OrgRepresenter is the organisation representative who signs the application.
	OrgRepresenter string             `json:"org_representer"`
	Employee       VacationEmployment `json:"employee"`
}

// VacationHistoryParams filters [VacationService.History]. Zero values are not sent.
type VacationHistoryParams struct {
	From, To     Date
	DepartmentID int64
	PositionID   int64
	Type         VacationType
	// Status is a pointer because 0 (draft) is a valid filter.
	Status *VacationStatus
	// Project is the project id from [VacationFilters.Projects].
	Project string
}

// VacationReason is the reason of an unpaid leave.
type VacationReason struct {
	// ReasonID comes from [VacationService.UnpaidReasons].
	ReasonID int64 `json:"reason_id"`
	// PersonalReason is required when ReasonID is [VacationReasonPersonal].
	PersonalReason string `json:"personal_reason,omitzero"`
}

// VacationApplication is a new vacation application for one employment.
type VacationApplication struct {
	// EmployeeID is the employment id (own, or of a subordinate for managers).
	EmployeeID int64        `json:"employee_id"`
	TypeID     VacationType `json:"type_id"`
	DateStart  Date         `json:"date_start"`
	DateEnd    Date         `json:"date_end"`
	// Duration is the day count, see [VacationConflict.NumOfDays].
	Duration int `json:"duration"`
	// Digital files an electronic application to be signed; false files a paper one.
	Digital bool `json:"digital"`
	// AdditionalData is the unpaid-leave reason; nil otherwise.
	AdditionalData *VacationReason `json:"additional_data"`
}

// vacationApplicationWire adds the fields that are always sent.
type vacationApplicationWire struct {
	VacationApplication
	FileNames   []string `json:"file_names,omitzero"`
	ExtraFields RawJSON  `json:"extra_fields"`
}

// VacationCreated is a created application.
type VacationCreated struct {
	ID int64 `json:"id"`
	// SignatureID and TaskID are used to sign a digital application.
	SignatureID FlexInt `json:"signature_id"`
	TaskID      FlexInt `json:"task_id"`
}

// VacationConflict is the result of a period check.
type VacationConflict struct {
	// Conflict is false or null when there is none; otherwise true or a message.
	Conflict RawJSON `json:"conflict"`
	// NumOfDays is the vacation day count of the period.
	NumOfDays int `json:"num_of_days"`
}

// HasConflict reports whether Conflict is truthy.
func (c *VacationConflict) HasConflict() bool {
	switch string(c.Conflict) {
	case "", "null", "false", `""`, "0":
		return false
	}
	return true
}

// VacationListParams filters [VacationService.List]. Zero values are not sent,
// except Limit and Offset.
type VacationListParams struct {
	Limit, Offset int
	PositionID    int64
	DepartmentID  int64
	Type          VacationType
	// ISU limits the list to one person.
	ISU      int64
	Category string
	Project  string
	From, To Date
	Statuses []VacationStatus
}

// VacationOverview is the management view: vacations grouped by department,
// person and position.
type VacationOverview struct {
	Departments []VacationDepartment `json:"departments"`
	// Count is the total for pagination.
	Count int `json:"count"`
}

// VacationDepartment is a department of [VacationOverview].
type VacationDepartment struct {
	DepartmentID   int64            `json:"department_id"`
	DepartmentName string           `json:"department_name"`
	People         []VacationMember `json:"people"`
}

// VacationMember is a person of [VacationDepartment].
type VacationMember struct {
	Name string `json:"name"`
	ISU  int64  `json:"isu"`
	// IsUpper marks a manager.
	IsUpper   bool                     `json:"is_upper"`
	Positions []VacationMemberPosition `json:"positions"`
}

// VacationMemberPosition is an employment of [VacationMember] with its vacations.
type VacationMemberPosition struct {
	EmployeeID           int64           `json:"employee_id"`
	PositionName         string          `json:"position_name"`
	CurrentRemainingDays float64         `json:"current_remaining_days"`
	ProjectCode          string          `json:"project_code"`
	ProjectName          string          `json:"project_name"`
	Category             string          `json:"category"`
	Vacations            []VacationEntry `json:"vacations"`
}

// VacationEntry is a vacation in the management view.
type VacationEntry struct {
	VacationID int64 `json:"vacation_id"`
	EmployeeID int64 `json:"employee_id"`
	Employee   *struct {
		EmployeeID int64 `json:"employee_id"`
	} `json:"employee"`
	DateStart Date `json:"date_start"`
	DateEnd   Date `json:"date_end"`
	Duration  int  `json:"duration"`
	// RemainingDaysOnVacStart is the balance on the first day ("N of M days").
	RemainingDaysOnVacStart *float64       `json:"remaining_days_on_vac_start"`
	VacationType            string         `json:"vacation_type"`
	StatusID                VacationStatus `json:"status_id"`
	// Status1C reports whether the vacation was sent to 1C.
	Status1C       bool    `json:"status_1c"`
	TaskID         FlexInt `json:"task_id"`
	SignatureID    FlexInt `json:"signature_id"`
	NeedsSignature bool    `json:"needs_signature"`
	// ReadOnly is false when the vacation can be edited.
	ReadOnly       bool       `json:"read_only"`
	OrgRepresenter string     `json:"org_representer"`
	ApprovedBy     string     `json:"approved_by"`
	ApprovedAt     *time.Time `json:"approved_at"`
	UpperName      string     `json:"upper_name"`
	AssignedTo     string     `json:"assigned_to"`
}

// VacationEmployeesParams filters [VacationService.Employees]. Zero values are not sent.
type VacationEmployeesParams struct {
	PositionID   int64
	DepartmentID int64
	ISU          int64
	Category     string
	Project      string
	From, To     Date
}

// VacationEmployeeGroup is a department with the employees a manager may file vacations for.
type VacationEmployeeGroup struct {
	DepartmentID   int64                  `json:"department_id"`
	DepartmentName string                 `json:"department_name"`
	Employees      []VacationEmployeeItem `json:"employees"`
}

// VacationEmployeeItem is an employment with its balance.
type VacationEmployeeItem struct {
	EmployeeID int64  `json:"employee_id"`
	PersName   string `json:"pers_name"`
	// PersID is the person id (ISU number).
	PersID       int64  `json:"pers_id"`
	PositionName string `json:"position_name"`
	Category     string `json:"category"`
	ProjectCode  string `json:"project_code"`
	ProjectName  string `json:"project_name"`
	// Remaining is the number of accumulated days.
	Remaining float64 `json:"remaining"`
}

// VacationAgreementParams filters [VacationService.PendingApproval] and
// [VacationService.Unapproved]. Zero values are not sent.
type VacationAgreementParams struct {
	PositionID   int64
	DepartmentID int64
	Type         VacationType
	// Assigned is the ISU number of the assignee, see [VacationService.Assignees].
	Assigned int64
	// ISU limits the list to one applicant.
	ISU           int64
	Limit, Offset int
}

// VacationAgreementTasks is a page of agreement tasks.
type VacationAgreementTasks struct {
	Tasks []VacationAgreementTask `json:"tasks"`
	Count int                     `json:"count"`
}

// VacationAgreementTask is an application awaiting approval or signature.
type VacationAgreementTask struct {
	VacationID int64 `json:"vacation_id"`
	// EmpID is the applicant's employment id.
	EmpID                   int64    `json:"emp_id"`
	EmpName                 string   `json:"emp_name"`
	EmpISU                  int64    `json:"emp_isu"`
	DateStart               Date     `json:"date_start"`
	DateEnd                 Date     `json:"date_end"`
	Duration                int      `json:"duration"`
	RemainingDaysOnVacStart *float64 `json:"remaining_days_on_vac_start"`
	VacationType            string   `json:"vacation_type"`
	PositionName            string   `json:"position_name"`
	// ProjectID is the project code, a number or a string.
	ProjectID      RawJSON        `json:"project_id,omitzero"`
	ProjectName    string         `json:"project_name"`
	DepartmentName string         `json:"department_name"`
	Status         VacationStatus `json:"status"`
	// UpperName is who pre-approved the application.
	UpperName      string `json:"upper_name"`
	OrgRepresenter string `json:"org_representer"`
	// ApprovalOnly means the current user can only pre-approve.
	ApprovalOnly bool `json:"approval_only"`
	// Refusable is false when the application cannot be rejected.
	Refusable   *bool          `json:"refusable"`
	SignatureID FlexInt        `json:"signature_id"`
	SignTaskID  FlexInt        `json:"sign_task_id"`
	Files       []VacationFile `json:"files"`
}

// VacationFile is a supporting file of an unpaid-leave application.
type VacationFile struct {
	// FileLink is an absolute URL.
	FileLink string `json:"file_link"`
	FileName string `json:"file_name"`
}

func setID(qv query, key string, v int64) query {
	if v != 0 {
		qv.set(key, v)
	}
	return qv
}

// Access returns the current user's rights in the vacation module. Staff only.
// GET /api/vacation/me/access
func (s *VacationService) Access(ctx context.Context) (*VacationAccess, error) {
	return call[*VacationAccess](ctx, s.c, get("api/vacation/me/access", nil))
}

// Filters returns the dictionaries of the management and agreement filters. Staff only.
// GET /api/vacation/filters
func (s *VacationService) Filters(ctx context.Context) (*VacationFilters, error) {
	return call[*VacationFilters](ctx, s.c, get("api/vacation/filters", nil))
}

// PersonalFilters returns the filter dictionaries limited to the current user's employments. Staff only.
// GET /api/vacation/filters/personal
func (s *VacationService) PersonalFilters(ctx context.Context) (*VacationFilters, error) {
	return call[*VacationFilters](ctx, s.c, get("api/vacation/filters/personal", nil))
}

// Types returns the vacation types (id is a [VacationType]). Staff only.
// GET /api/vacation/filters/types
func (s *VacationService) Types(ctx context.Context) ([]IDValue, error) {
	return call[[]IDValue](ctx, s.c, get("api/vacation/filters/types", nil))
}

// UnpaidReasons returns the reasons of an unpaid leave (uto_reasons). Staff only.
// GET /api/vacation/filters/uto_reasons
func (s *VacationService) UnpaidReasons(ctx context.Context) ([]IDValue, error) {
	return call[[]IDValue](ctx, s.c, get("api/vacation/filters/uto_reasons", nil))
}

// Positions returns the current user's employments. Staff only.
// GET /api/vacation/filters/positions
func (s *VacationService) Positions(ctx context.Context) ([]VacationPosition, error) {
	return call[[]VacationPosition](ctx, s.c, get("api/vacation/filters/positions", nil))
}

// People searches employees by name or ISU number. Staff only.
// GET /api/vacation/filters/people
func (s *VacationService) People(ctx context.Context, query string) ([]VacationPerson, error) {
	return call[[]VacationPerson](ctx, s.c, get("api/vacation/filters/people", q().set("query", query)))
}

// Assignees returns the people agreement tasks can be assigned to (id is the ISU number). Staff only.
// GET /api/vacation/filters/assigned
func (s *VacationService) Assignees(ctx context.Context) ([]IDValue, error) {
	return call[[]IDValue](ctx, s.c, get("api/vacation/filters/assigned", nil))
}

// Calendar returns the weekends and holidays of a year; empty when the year is not available. Staff only.
// GET /api/vacation/calendar
func (s *VacationService) Calendar(ctx context.Context, year int) ([]VacationCalendarDay, error) {
	return call[[]VacationCalendarDay](ctx, s.c, get("api/vacation/calendar", q().set("year", year)))
}

// MyCalendar returns the current user's vacations overlapping a period. Staff only.
// GET /api/vacation/me/calendar
func (s *VacationService) MyCalendar(ctx context.Context, from, to Date) ([]Vacation, error) {
	return call[[]Vacation](ctx, s.c, get("api/vacation/me/calendar", q().set("date_start", from).set("date_end", to)))
}

// Balance returns the vacation-day balance of each of the current user's employments on a date. Staff only.
// GET /api/vacation/me/balance
func (s *VacationService) Balance(ctx context.Context, on Date) ([]VacationBalance, error) {
	return call[[]VacationBalance](ctx, s.c, get("api/vacation/me/balance", q().set("date_start", on)))
}

// History returns the current user's vacation applications. Staff only.
// GET /api/vacation/me/vacations
func (s *VacationService) History(ctx context.Context, p VacationHistoryParams) ([]Vacation, error) {
	qv := q().set("date_start", p.From).set("date_end", p.To)
	setID(qv, "depId", p.DepartmentID)
	setID(qv, "posId", p.PositionID)
	setID(qv, "type", int64(p.Type))
	if p.Status != nil {
		qv.set("stat", int(*p.Status))
	}
	qv.set("proj", p.Project)
	return call[[]Vacation](ctx, s.c, get("api/vacation/me/vacations", qv))
}

// Create files vacation applications of any type but unpaid leave, one per
// employment. A digital application must then be signed with the returned
// ids; a paper one is printed from [VacationService.Download]. Staff only.
// POST /api/vacation/me/vacations
func (s *VacationService) Create(ctx context.Context, apps []VacationApplication) ([]VacationCreated, error) {
	return call[[]VacationCreated](ctx, s.c, post("api/vacation/me/vacations", wireApplications(apps, nil)))
}

// CreateUnpaid files unpaid-leave (UTO) applications with optional supporting
// files. Each application needs AdditionalData with a reason. The field name
// of every upload is set to "files". Staff only.
// POST /api/vacation/me/vacations/uto
func (s *VacationService) CreateUnpaid(ctx context.Context, apps []VacationApplication, files ...Upload) ([]VacationCreated, error) {
	names := make([]string, 0, len(files))
	parts := make([]Upload, len(files))
	for i, f := range files {
		f.Field = "files"
		parts[i] = f
		names = append(names, f.Name)
	}
	body, err := json.Marshal(wireApplications(apps, names))
	if err != nil {
		return nil, fmt.Errorf("myitmo: encode vacations: %w", err)
	}
	return call[[]VacationCreated](ctx, s.c, multipart(http.MethodPost, "api/vacation/me/vacations/uto", map[string]string{"vacations": string(body)}, parts...))
}

func wireApplications(apps []VacationApplication, fileNames []string) []vacationApplicationWire {
	out := make([]vacationApplicationWire, len(apps))
	for i, a := range apps {
		out[i] = vacationApplicationWire{VacationApplication: a, FileNames: fileNames, ExtraFields: RawJSON("null")}
		if fileNames != nil && out[i].FileNames == nil {
			out[i].FileNames = []string{}
		}
	}
	return out
}

// Delete deletes a draft or paper application. Staff only.
// DELETE /api/vacation/me/vacations/{vacation_id}
func (s *VacationService) Delete(ctx context.Context, vacationID int64) error {
	return exec(ctx, s.c, del("api/vacation/me/vacations/"+id(vacationID), nil))
}

// Upcoming returns the current user's upcoming vacations. Staff only.
// GET /api/vacation/me/vacations/upcoming
func (s *VacationService) Upcoming(ctx context.Context) ([]Vacation, error) {
	return call[[]Vacation](ctx, s.c, get("api/vacation/me/vacations/upcoming", nil))
}

// Conflicts checks a planned period for the given employments and counts its
// vacation days. Staff only.
// GET /api/vacation/me/vacation_conflict_with_days
func (s *VacationService) Conflicts(ctx context.Context, from, to Date, employeeIDs ...int64) (*VacationConflict, error) {
	qv := q().set("date_start", from).set("date_end", to).set("employee_id", employeeIDs)
	return call[*VacationConflict](ctx, s.c, get("api/vacation/me/vacation_conflict_with_days", qv))
}

// List returns the vacations of subordinate employees grouped by department. Staff only.
// GET /api/vacation/vacations
func (s *VacationService) List(ctx context.Context, p VacationListParams) (*VacationOverview, error) {
	qv := q().set("limit", p.Limit).set("offset", p.Offset)
	setID(qv, "posId", p.PositionID)
	setID(qv, "depId", p.DepartmentID)
	setID(qv, "type", int64(p.Type))
	setID(qv, "query", p.ISU)
	qv.set("category", p.Category).set("proj", p.Project).set("date_start", p.From).set("date_end", p.To)
	for _, st := range p.Statuses {
		qv.values().Add("stat", strconv.Itoa(int(st)))
	}
	return call[*VacationOverview](ctx, s.c, get("api/vacation/vacations", qv))
}

// Employees returns the employees the current manager may file vacations for, with balances. Staff only.
// GET /api/vacation/employees
func (s *VacationService) Employees(ctx context.Context, p VacationEmployeesParams) ([]VacationEmployeeGroup, error) {
	qv := q()
	setID(qv, "posId", p.PositionID)
	setID(qv, "depId", p.DepartmentID)
	setID(qv, "query", p.ISU)
	qv.set("category", p.Category).set("proj", p.Project).set("date_start", p.From).set("date_end", p.To)
	return call[[]VacationEmployeeGroup](ctx, s.c, get("api/vacation/employees", qv))
}

func agreementQuery(p VacationAgreementParams) query {
	qv := q()
	setID(qv, "posId", p.PositionID)
	setID(qv, "depId", p.DepartmentID)
	setID(qv, "type", int64(p.Type))
	setID(qv, "assigned", p.Assigned)
	setID(qv, "query", p.ISU)
	if p.Limit > 0 {
		qv.set("limit", p.Limit).set("offset", p.Offset)
	}
	return qv
}

// PendingApproval returns the applications awaiting the current user's approval or signature. Staff only.
// GET /api/vacation/pending-approval
func (s *VacationService) PendingApproval(ctx context.Context, p VacationAgreementParams) (*VacationAgreementTasks, error) {
	return call[*VacationAgreementTasks](ctx, s.c, get("api/vacation/pending-approval", agreementQuery(p)))
}

// PendingApprovalCount returns the number of applications awaiting the current user. Staff only.
// GET /api/vacation/pending-approval/count
func (s *VacationService) PendingApprovalCount(ctx context.Context) (int, error) {
	r, err := call[struct {
		Count int `json:"count"`
	}](ctx, s.c, get("api/vacation/pending-approval/count", nil))
	return r.Count, err
}

// PendingApprovalTask returns one agreement task; VacationID is 0 when it was not found. Staff only.
// GET /api/vacation/pending-approval/{vacation_id}
func (s *VacationService) PendingApprovalTask(ctx context.Context, vacationID int64) (*VacationAgreementTask, error) {
	return call[*VacationAgreementTask](ctx, s.c, get("api/vacation/pending-approval/"+id(vacationID), nil))
}

// Unapproved returns the applications not yet pre-approved, for signers. Staff only.
// GET /api/vacation/unapproved-vacations
func (s *VacationService) Unapproved(ctx context.Context, p VacationAgreementParams) (*VacationAgreementTasks, error) {
	return call[*VacationAgreementTasks](ctx, s.c, get("api/vacation/unapproved-vacations", agreementQuery(p)))
}

// Approve approves applications without a signature, at most 10 per call. Staff only.
// PATCH /api/vacation/vacations/approve
func (s *VacationService) Approve(ctx context.Context, vacationIDs ...int64) error {
	if vacationIDs == nil {
		vacationIDs = []int64{}
	}
	return exec(ctx, s.c, patch("api/vacation/vacations/approve", vacationIDs))
}

// Reject rejects an application (managers) or cancels the user's own pending
// one (reason may be empty). Staff only.
// PATCH /api/vacation/vacation/{vacation_id}/reject
func (s *VacationService) Reject(ctx context.Context, vacationID int64, reason string) error {
	body := struct {
		ID              int64  `json:"id"`
		RejectionReason string `json:"rejection_reason"`
	}{vacationID, reason}
	return exec(ctx, s.c, patch("api/vacation/vacation/"+id(vacationID)+"/reject", body))
}

// PDF downloads the generated (unsigned) application. employeeID is the
// applicant's employment id. Staff only.
// GET /api/vacation/static/vacations/{employee_id}/{vacation_id}/pdf
func (s *VacationService) PDF(ctx context.Context, employeeID, vacationID int64) (*File, error) {
	return download(ctx, s.c, get("api/vacation/static/vacations/"+id(employeeID)+"/"+id(vacationID)+"/pdf", nil))
}

// Signed downloads the digitally signed application (status 3). Staff only.
// GET /api/vacation/vacations/{vacation_id}/signed
func (s *VacationService) Signed(ctx context.Context, vacationID int64) (*File, error) {
	return download(ctx, s.c, get("api/vacation/vacations/"+id(vacationID)+"/signed", nil))
}

// Original downloads the original of a paper application (status 4). Staff only.
// GET /api/vacation/vacations/{vacation_id}/signed?type=original
func (s *VacationService) Original(ctx context.Context, vacationID int64) (*File, error) {
	return download(ctx, s.c, get("api/vacation/vacations/"+id(vacationID)+"/signed", q().set("type", "original")))
}

// Archive downloads the archive with a signed application and its signatures. Staff only.
// GET /api/vacation/vacations/{vacation_id}/file-archive
func (s *VacationService) Archive(ctx context.Context, vacationID int64) (*File, error) {
	return download(ctx, s.c, get("api/vacation/vacations/"+id(vacationID)+"/file-archive", nil))
}

// Download downloads the paper applications of several vacations (a single
// file or an archive). Staff only.
// GET /api/vacation/vacations/download
func (s *VacationService) Download(ctx context.Context, vacationIDs ...int64) (*File, error) {
	return download(ctx, s.c, get("api/vacation/vacations/download", q().set("ids", vacationIDs)))
}
