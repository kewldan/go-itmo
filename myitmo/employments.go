package myitmo

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/kewldan/go-itmo/internal/rest"
)

// EmploymentsService is the electronic hiring workflow of new employees
// (/api/employments). Every route is staff only: department heads approve
// requests and HR checks the candidate's documents; see [EmploymentsService.Rights].
type EmploymentsService struct{ c *Client }

// EmploymentStatus is the stage of a hiring request.
type EmploymentStatus int

// Values of EmploymentStatus.
const (
	EmploymentStopped            EmploymentStatus = -2
	EmploymentRejected           EmploymentStatus = -1
	EmploymentAwaitingApproval   EmploymentStatus = 0
	EmploymentApproved           EmploymentStatus = 1
	EmploymentDocumentCollection EmploymentStatus = 2
	EmploymentDocumentCheck      EmploymentStatus = 3
	EmploymentDocumentCorrection EmploymentStatus = 4
	EmploymentOPSSigning         EmploymentStatus = 5
	EmploymentOPSSigned          EmploymentStatus = 6
	EmploymentContractSigned     EmploymentStatus = 7
	EmploymentCompleted          EmploymentStatus = 10
)

// EmploymentDocumentStatus is the state of a candidate document; briefings
// and the paper work book use the same scale.
type EmploymentDocumentStatus int

// Values of EmploymentDocumentStatus.
const (
	EmploymentDocumentRejected   EmploymentDocumentStatus = -1
	EmploymentDocumentAbsent     EmploymentDocumentStatus = 0
	EmploymentDocumentInProgress EmploymentDocumentStatus = 1
	EmploymentDocumentOnCheck    EmploymentDocumentStatus = 2
	EmploymentDocumentApproved   EmploymentDocumentStatus = 3
)

// EmploymentDocumentType is the type of a candidate document, used in paths.
type EmploymentDocumentType string

// Values of EmploymentDocumentType.
const (
	EmploymentDocPDAgreement        EmploymentDocumentType = "pd_agreement"
	EmploymentDocIdentityPhoto      EmploymentDocumentType = "identity_photo"
	EmploymentDocPassport           EmploymentDocumentType = "passport"
	EmploymentDocSNILS              EmploymentDocumentType = "snils"
	EmploymentDocINN                EmploymentDocumentType = "inn"
	EmploymentDocMilitary           EmploymentDocumentType = "military"
	EmploymentDocEducation          EmploymentDocumentType = "education"
	EmploymentDocCriminalRecord     EmploymentDocumentType = "criminal_record"
	EmploymentDocWorkBook           EmploymentDocumentType = "work_book"
	EmploymentDocMedicalExamination EmploymentDocumentType = "medical_examination"
	EmploymentDocBriefing           EmploymentDocumentType = "briefing"
	EmploymentDocRequisites         EmploymentDocumentType = "requisites"
	EmploymentDocCivilService       EmploymentDocumentType = "civil_service"
)

// EmploymentSignedDocumentType is the type of a signed contract document uploaded by HR.
type EmploymentSignedDocumentType string

// Values of EmploymentSignedDocumentType.
const (
	EmploymentSignedContract                        EmploymentSignedDocumentType = "contract"
	EmploymentSignedElectronicInteractionConsent    EmploymentSignedDocumentType = "electronic_interaction_consent"
	EmploymentSignedPersonalDataDistributionConsent EmploymentSignedDocumentType = "personal_data_distribution_consent"
	EmploymentSignedCivilServiceNotice              EmploymentSignedDocumentType = "civil_service_notice"
	EmploymentSignedCivilServiceAssurance           EmploymentSignedDocumentType = "civil_service_assurance"
)

// Values of EmploymentFormat.ID.
const (
	EmploymentFormatOnsite = 1
	EmploymentFormatRemote = 2
)

// Values of EmploymentCreateRequest.EmploymentTypeID.
const (
	EmploymentTypeMain             = 1
	EmploymentTypeExternalPartTime = 2
	EmploymentTypeInternalPartTime = 3
)

// Values of EmploymentCreateRequest.WorkingHoursTypeID.
const (
	EmploymentHoursNonStandard = 0
	EmploymentHoursStandard    = 1
)

// Values of EmploymentCreateRequest.ProbationDurationID.
const (
	EmploymentProbation2Weeks = 1
	EmploymentProbation1Month = 2
	EmploymentProbation2Month = 3
	EmploymentProbation3Month = 4
	EmploymentProbation1Week  = 5
	EmploymentProbation3Weeks = 6
)

// EmploymentCitizenshipRussia is the citizenship code of the Russian Federation (OKSM 643).
const EmploymentCitizenshipRussia = 643

// Employment is a hiring request.
type Employment struct {
	EmploymentID int64 `json:"employmentId"`
	// ID is a fallback for when EmploymentID is absent.
	ID int64 `json:"id"`
	// EmploymentIDAlt is another spelling that may appear after creation.
	EmploymentIDAlt int64            `json:"employmentID"`
	Status          EmploymentStatus `json:"status"`
	// BriefingStatus is nil when the briefing has not started.
	BriefingStatus    *EmploymentDocumentStatus `json:"briefingStatus"`
	StopReason        *EmploymentStopInfo       `json:"stopReason"`
	RejectComment     string                    `json:"rejectComment"`
	CandidateInfo     *EmploymentCandidate      `json:"candidateInfo"`
	Contact           *EmploymentContact        `json:"contact"`
	PositionInfo      *EmploymentPositionInfo   `json:"positionInfo"`
	Format            *EmploymentFormat         `json:"format"`
	EmploymentType    *EmploymentName           `json:"employmentType"`
	ProbationDuration *EmploymentName           `json:"probationDuration"`
	// FixedTermReason is set for fixed-term contracts; its shape is unknown.
	FixedTermReason   RawJSON `json:"fixedTermReason"`
	ContractDateStart Date    `json:"contractDateStart"`
	// ContractDateEnd is zero for open-ended contracts.
	ContractDateEnd Date `json:"contractDateEnd"`
	// Schedule is the work days joined with ", " ("Пн, Вт, ...").
	Schedule       string `json:"schedule"`
	IsInoagent     bool   `json:"isInoagent"`
	IsDisqualified bool   `json:"isDisqualified"`
}

// Key returns the request ID from whichever field the server filled.
func (e *Employment) Key() int64 {
	switch {
	case e.EmploymentID != 0:
		return e.EmploymentID
	case e.EmploymentIDAlt != 0:
		return e.EmploymentIDAlt
	}
	return e.ID
}

// EmploymentStopInfo is why a hiring request was stopped.
type EmploymentStopInfo struct {
	Name    string `json:"name"`
	Comment string `json:"comment"`
}

// EmploymentCandidate is the person being hired.
type EmploymentCandidate struct {
	LastName   string `json:"lastName"`
	FirstName  string `json:"firstName"`
	Patronymic string `json:"patronymic"`
	// PersonID is nil until the candidate has an ITMO person record.
	PersonID *int64 `json:"personId"`
	Phone    string `json:"phone"`
	Email    string `json:"email"`
	// Citizenship shape is unknown.
	Citizenship RawJSON `json:"citizenship"`
}

// EmploymentContact is the contact person of a hiring request.
type EmploymentContact struct {
	PersonID int64  `json:"personId"`
	Name     string `json:"name"`
	Phone    string `json:"phone"`
	Email    string `json:"email"`
}

// EmploymentPositionInfo is the position the candidate is hired for.
type EmploymentPositionInfo struct {
	Position   EmploymentName       `json:"position"`
	Department EmploymentDepartment `json:"department"`
	// StaffUnit is the rate, 0.01 to 1.
	StaffUnit float64 `json:"staffUnit"`
	// Category is the staff category; "ППС" means teaching staff.
	Category string `json:"category"`
}

// EmploymentDepartment is a department of the university.
type EmploymentDepartment struct {
	ID             int64  `json:"id"`
	Name           string `json:"name"`
	ShortName      string `json:"shortName"`
	DepartmentCode string `json:"departmentCode"`
}

// EmploymentFormat is the work format.
type EmploymentFormat struct {
	// ID is EmploymentFormatOnsite or EmploymentFormatRemote.
	ID         int    `json:"id"`
	Name       string `json:"name"`
	FromRussia bool   `json:"fromRussia"`
}

// EmploymentName is a named reference value.
type EmploymentName struct {
	Name string `json:"name"`
}

// EmploymentList is a page of hiring requests.
type EmploymentList struct {
	Employments []Employment `json:"employments"`
	Count       int          `json:"count"`
}

// EmploymentListParams filters [EmploymentsService.List].
type EmploymentListParams struct {
	// Limit defaults to 20.
	Limit  int
	Offset int
	// Department is a department ID; 0 means any.
	Department int64
	// Status filters by stage; nil means any.
	Status *EmploymentStatus
	// Query is a search text.
	Query string
}

// EmploymentCreateRequest is a new hiring request.
type EmploymentCreateRequest struct {
	CandidateInfo EmploymentCandidateInput `json:"candidateInfo"`
	PositionInfo  EmploymentPositionInput  `json:"positionInfo"`
	// ProbationDurationID is one of the EmploymentProbation constants.
	ProbationDurationID int                   `json:"probationDurationId"`
	Format              EmploymentFormatInput `json:"format"`
	// EmploymentTypeID is one of the EmploymentType constants.
	EmploymentTypeID int `json:"employmentTypeId"`
	// Schedule is the work days joined with ", " ("Пн, Вт, Ср").
	Schedule string `json:"schedule"`
	// WorkingHoursTypeID is EmploymentHoursStandard or EmploymentHoursNonStandard.
	WorkingHoursTypeID int  `json:"workingHoursTypeId"`
	ContractDateStart  Date `json:"contractDateStart"`
	// ContactPersonID is set when the applicant is not the contact person.
	ContactPersonID int64 `json:"contactPersonId,omitzero"`
	// ContractDateEnd, FixedTermReasonID and AbsentEmployeePersonID are for fixed-term contracts.
	ContractDateEnd        Date  `json:"contractDateEnd,omitzero"`
	FixedTermReasonID      int64 `json:"fixedTermReasonId,omitzero"`
	AbsentEmployeePersonID int64 `json:"absentEmployeePersonId,omitzero"`
}

// EmploymentCandidateInput is the candidate of a new hiring request.
type EmploymentCandidateInput struct {
	LastName   string `json:"lastName"`
	FirstName  string `json:"firstName"`
	Patronymic string `json:"patronymic"`
	// CitizenshipCode is an OKSM country code (EmploymentCitizenshipRussia).
	CitizenshipCode int    `json:"citizenshipCode"`
	Phone           string `json:"phone"`
	Email           string `json:"email"`
}

// EmploymentPositionInput is the position of a new hiring request.
type EmploymentPositionInput struct {
	// StaffUnit is the rate, 0.01 to 1.
	StaffUnit float64 `json:"staffUnit"`
	// StaffUnitID is EmploymentStaffUnit.StaffUnitID; nil is sent as null.
	StaffUnitID *int64 `json:"staffUnitId"`
}

// EmploymentFormatInput is the work format of a new hiring request.
type EmploymentFormatInput struct {
	// ID is EmploymentFormatOnsite or EmploymentFormatRemote.
	ID         int  `json:"id"`
	FromRussia bool `json:"fromRussia"`
}

// EmploymentRights is what the current user may do with a hiring request.
type EmploymentRights struct {
	// IsLead allows approving or rejecting the request.
	IsLead bool `json:"isLead"`
	// IsHR allows checking documents and the contract.
	IsHR bool `json:"isHR"`
}

// EmploymentDocument is a document filled in by the candidate.
type EmploymentDocument struct {
	Type       EmploymentDocumentType   `json:"type"`
	Status     EmploymentDocumentStatus `json:"status"`
	UploadedAt *time.Time               `json:"uploadedAt"`
	Comment    string                   `json:"comment"`
	// Data is type-specific: e.g. passport has series, number, issueDate and
	// scan URLs; work_book has xmlFileUrl and pdfFileUrl; requisites has
	// bankName, bik and accountNumber. File URLs can be fetched with
	// [EmploymentsService.DocumentFile].
	Data RawJSON `json:"data"`
}

// EmploymentSignedDocuments is the signed contract documents of a hiring request.
type EmploymentSignedDocuments struct {
	Documents []EmploymentSignedDocument `json:"documents"`
	// EmploymentStatus of 7 or more locks uploading and removing.
	EmploymentStatus EmploymentStatus `json:"employmentStatus"`
}

// EmploymentSignedDocument is one signed contract document.
type EmploymentSignedDocument struct {
	Type EmploymentSignedDocumentType `json:"type"`
	// Status is a server-defined string; empty when nothing is uploaded.
	Status   string `json:"status"`
	FileURL  string `json:"fileUrl"`
	FileName string `json:"fileName"`
	// FileSize and Size are alternative spellings of the size in bytes.
	FileSize int64 `json:"fileSize"`
	Size     int64 `json:"size"`
}

// EmploymentSigner is an employer-side signer of the contract.
type EmploymentSigner struct {
	KeycloakID string `json:"keycloakId"`
	FullName   string `json:"fullName"`
	// ProxyDateTo is when the signer's power of attorney expires; zero when unlimited.
	ProxyDateTo Date   `json:"proxyDateTo"`
	Position    string `json:"position"`
}

// EmploymentStepStatus is the state of the briefings or the paper work book.
// Only Status is known.
type EmploymentStepStatus struct {
	Status EmploymentDocumentStatus `json:"status"`
}

// EmploymentStaffUnit is a vacant position in a department.
type EmploymentStaffUnit struct {
	StaffUnitID int64                `json:"staffUnitId"`
	Position    EmploymentName       `json:"position"`
	Department  EmploymentDepartment `json:"department"`
	Category    string               `json:"category"`
	UnitsFree   float64              `json:"unitsFree"`
	// StaffUnitAvailableBefore is zero when the position has no deadline.
	StaffUnitAvailableBefore Date `json:"staffUnitAvailableBefore"`
}

// EmploymentStopReason is a reason to stop a hiring process.
type EmploymentStopReason struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// EmploymentFixedTermReason is a reason for a fixed-term contract.
type EmploymentFixedTermReason struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	// RequiresAbsentEmployee means AbsentEmployeePersonID must be set.
	RequiresAbsentEmployee bool `json:"requiresAbsentEmployee"`
}

type employmentComment struct {
	Comment string `json:"comment"`
}

type employmentStop struct {
	ReasonID int64  `json:"reasonId"`
	Comment  string `json:"comment"`
}

type employmentSigner struct {
	SignerKeycloakID string `json:"signerKeycloakId"`
}

func employmentPath(employmentID int64) string {
	return "api/employments/employments/" + id(employmentID)
}

// List returns hiring requests. Staff only.
// GET /api/employments/employments/
func (s *EmploymentsService) List(ctx context.Context, p EmploymentListParams) (*EmploymentList, error) {
	limit := p.Limit
	if limit == 0 {
		limit = 20
	}
	qv := q().set("limit", limit).set("offset", p.Offset).set("query", p.Query)
	if p.Department != 0 {
		qv.set("department", p.Department)
	}
	if p.Status != nil {
		qv.set("status", int(*p.Status))
	}
	return call[*EmploymentList](ctx, s.c, get("api/employments/employments/", qv))
}

// Create files a new hiring request; see [Employment.Key] for its ID. Staff only.
// POST /api/employments/employments/
func (s *EmploymentsService) Create(ctx context.Context, req EmploymentCreateRequest) (*Employment, error) {
	return call[*Employment](ctx, s.c, post("api/employments/employments/", req))
}

// Get returns a hiring request. Staff only.
// GET /api/employments/employments/{employmentId}
func (s *EmploymentsService) Get(ctx context.Context, employmentID int64) (*Employment, error) {
	return call[*Employment](ctx, s.c, get(employmentPath(employmentID), nil))
}

// Rights returns what the current user may do with a hiring request. Staff only.
// GET /api/employments/employments/{employmentId}/rights
func (s *EmploymentsService) Rights(ctx context.Context, employmentID int64) (*EmploymentRights, error) {
	return call[*EmploymentRights](ctx, s.c, get(employmentPath(employmentID)+"/rights", nil))
}

// Approve lets the department head approve a request (0 → 1); the result
// carries at least the new Status. Staff only.
// POST /api/employments/employments/{employmentId}/approve
func (s *EmploymentsService) Approve(ctx context.Context, employmentID int64) (*Employment, error) {
	return call[*Employment](ctx, s.c, post(employmentPath(employmentID)+"/approve", nil))
}

// Reject rejects a request (→ -1); the result carries at least the new Status. Staff only.
// POST /api/employments/employments/{employmentId}/reject
func (s *EmploymentsService) Reject(ctx context.Context, employmentID int64, comment string) (*Employment, error) {
	return call[*Employment](ctx, s.c, post(employmentPath(employmentID)+"/reject", employmentComment{comment}))
}

// Stop stops the hiring process (→ -2) for a reason from [EmploymentsService.StopReasons]. Staff only.
// POST /api/employments/employments/{employmentId}/stop
func (s *EmploymentsService) Stop(ctx context.Context, employmentID, reasonID int64, comment string) (*Employment, error) {
	return call[*Employment](ctx, s.c, post(employmentPath(employmentID)+"/stop", employmentStop{reasonID, comment}))
}

// Documents returns the documents filled in by the candidate. Staff only.
// GET /api/employments/employments/{employmentId}/documents
func (s *EmploymentsService) Documents(ctx context.Context, employmentID int64) ([]EmploymentDocument, error) {
	return call[[]EmploymentDocument](ctx, s.c, get(employmentPath(employmentID)+"/documents", nil))
}

// ApproveDocument approves one candidate document. Staff only.
// POST /api/employments/employments/{employmentId}/documents/{documentType}/approve
func (s *EmploymentsService) ApproveDocument(ctx context.Context, employmentID int64, docType EmploymentDocumentType) error {
	return exec(ctx, s.c, post(employmentPath(employmentID)+"/documents/"+id(string(docType))+"/approve", nil))
}

// RejectDocument rejects one candidate document with a comment. Staff only.
// POST /api/employments/employments/{employmentId}/documents/{documentType}/reject
func (s *EmploymentsService) RejectDocument(ctx context.Context, employmentID int64, docType EmploymentDocumentType, comment string) error {
	return exec(ctx, s.c, post(employmentPath(employmentID)+"/documents/"+id(string(docType))+"/reject", employmentComment{comment}))
}

// ApproveDocuments approves the whole document set (→ 5); every document
// must be approved first. Staff only.
// POST /api/employments/employments/{employmentId}/documents/approve
func (s *EmploymentsService) ApproveDocuments(ctx context.Context, employmentID int64) (*Employment, error) {
	return call[*Employment](ctx, s.c, post(employmentPath(employmentID)+"/documents/approve", nil))
}

// ReturnDocuments sends the documents back to the candidate for correction
// (→ 4); at least one document must be rejected. Staff only.
// POST /api/employments/employments/{employmentId}/documents/reject
func (s *EmploymentsService) ReturnDocuments(ctx context.Context, employmentID int64, comment string) (*Employment, error) {
	return call[*Employment](ctx, s.c, post(employmentPath(employmentID)+"/documents/reject", employmentComment{comment}))
}

// RejectDocuments rejects the document set without a comment; the
// request then counts as rejected. Staff only.
// POST /api/employments/employments/{employmentId}/documents/reject
func (s *EmploymentsService) RejectDocuments(ctx context.Context, employmentID int64) (*Employment, error) {
	return call[*Employment](ctx, s.c, post(employmentPath(employmentID)+"/documents/reject", nil))
}

// SendDocuments sends the checked documents onward (presumably to signing). Staff only.
// POST /api/employments/employments/{employmentId}/documents/send
func (s *EmploymentsService) SendDocuments(ctx context.Context, employmentID int64) error {
	return exec(ctx, s.c, post(employmentPath(employmentID)+"/documents/send", nil))
}

// SignedDocuments returns the signed contract documents uploaded by HR. Staff only.
// GET /api/employments/employments/{employmentId}/documents/signed
func (s *EmploymentsService) SignedDocuments(ctx context.Context, employmentID int64) (*EmploymentSignedDocuments, error) {
	return call[*EmploymentSignedDocuments](ctx, s.c, get(employmentPath(employmentID)+"/documents/signed", nil))
}

// UploadSignedDocument uploads a signed contract document; the file field name is set to "file". Staff only.
// POST /api/employments/employments/{employmentId}/documents/signed/upload
func (s *EmploymentsService) UploadSignedDocument(ctx context.Context, employmentID int64, docType EmploymentSignedDocumentType, file Upload) error {
	file.Field = "file"
	return exec(ctx, s.c, multipart(http.MethodPost, employmentPath(employmentID)+"/documents/signed/upload",
		map[string]string{"documentType": string(docType)}, file))
}

// DeleteSignedDocument removes an uploaded signed document. Staff only.
// DELETE /api/employments/employments/{employmentId}/documents/signed/{documentType}
func (s *EmploymentsService) DeleteSignedDocument(ctx context.Context, employmentID int64, docType EmploymentSignedDocumentType) error {
	return exec(ctx, s.c, del(employmentPath(employmentID)+"/documents/signed/"+id(string(docType)), nil))
}

// Signers returns who may sign the contract on the employer side. Staff only.
// GET /api/employments/employments/{employmentId}/signers
func (s *EmploymentsService) Signers(ctx context.Context, employmentID int64) ([]EmploymentSigner, error) {
	return call[[]EmploymentSigner](ctx, s.c, get(employmentPath(employmentID)+"/signers", nil))
}

func signerBody(signerKeycloakID string) any {
	if signerKeycloakID == "" {
		return nil
	}
	return employmentSigner{signerKeycloakID}
}

// Contract downloads the generated employment contract; signerKeycloakID is
// EmploymentSigner.KeycloakID or empty. Staff only.
// POST /api/employments/employments/{employmentId}/contract
func (s *EmploymentsService) Contract(ctx context.Context, employmentID int64, signerKeycloakID string) (*File, error) {
	return download(ctx, s.c, post(employmentPath(employmentID)+"/contract", signerBody(signerKeycloakID)))
}

// ContractPreview downloads the contract as a PDF for preview; signerKeycloakID
// is EmploymentSigner.KeycloakID or empty. Staff only.
// POST /api/employments/employments/{employmentId}/contract/file
func (s *EmploymentsService) ContractPreview(ctx context.Context, employmentID int64, signerKeycloakID string) (*File, error) {
	return download(ctx, s.c, post(employmentPath(employmentID)+"/contract/file", signerBody(signerKeycloakID)))
}

// Briefing returns the state of the safety and induction briefings. Staff only.
// GET /api/employments/employments/{employmentId}/briefing
func (s *EmploymentsService) Briefing(ctx context.Context, employmentID int64) (*EmploymentStepStatus, error) {
	return call[*EmploymentStepStatus](ctx, s.c, get(employmentPath(employmentID)+"/briefing", nil))
}

// ApproveBriefing approves the briefings. Staff only.
// POST /api/employments/employments/{employmentId}/briefing/approve
func (s *EmploymentsService) ApproveBriefing(ctx context.Context, employmentID int64) error {
	return exec(ctx, s.c, post(employmentPath(employmentID)+"/briefing/approve", nil))
}

// PaperWorkBook returns the state of the candidate's paper work-record book. Staff only.
// GET /api/employments/employments/{employmentId}/paper_work_book
func (s *EmploymentsService) PaperWorkBook(ctx context.Context, employmentID int64) (*EmploymentStepStatus, error) {
	return call[*EmploymentStepStatus](ctx, s.c, get(employmentPath(employmentID)+"/paper_work_book", nil))
}

// ApprovePaperWorkBook confirms that the candidate brought the paper work book to HR. Staff only.
// POST /api/employments/employments/{employmentId}/paper_work_book/approve
func (s *EmploymentsService) ApprovePaperWorkBook(ctx context.Context, employmentID int64) error {
	return exec(ctx, s.c, post(employmentPath(employmentID)+"/paper_work_book/approve", nil))
}

// Departments searches departments; an empty query lists them. Staff only.
// GET /api/employments/references/departments
func (s *EmploymentsService) Departments(ctx context.Context, query string) ([]EmploymentDepartment, error) {
	out, err := call[struct {
		Departments []EmploymentDepartment `json:"departments"`
	}](ctx, s.c, get("api/employments/references/departments", q().set("query", query)))
	return out.Departments, err
}

// StaffUnits returns vacant positions; department 0 means any, limit 0 means 100. Staff only.
// GET /api/employments/references/staff_units
func (s *EmploymentsService) StaffUnits(ctx context.Context, department int64, limit, offset int) ([]EmploymentStaffUnit, error) {
	if limit == 0 {
		limit = 100
	}
	qv := q().set("limit", limit).set("offset", offset)
	if department != 0 {
		qv.set("department", department)
	}
	out, err := call[struct {
		StaffUnits []EmploymentStaffUnit `json:"staffUnits"`
	}](ctx, s.c, get("api/employments/references/staff_units", qv))
	return out.StaffUnits, err
}

// StopReasons returns the reasons to stop a hiring process. Staff only.
// GET /api/employments/references/employment_stop_reasons
func (s *EmploymentsService) StopReasons(ctx context.Context) ([]EmploymentStopReason, error) {
	return call[[]EmploymentStopReason](ctx, s.c, get("api/employments/references/employment_stop_reasons", nil))
}

// FixedTermReasons returns the reasons for a fixed-term contract. Staff only.
// GET /api/employments/references/fixed_term_reasons
func (s *EmploymentsService) FixedTermReasons(ctx context.Context) ([]EmploymentFixedTermReason, error) {
	return call[[]EmploymentFixedTermReason](ctx, s.c, get("api/employments/references/fixed_term_reasons", nil))
}

// StaticDocument downloads a candidate file through the static-documents
// proxy. path is the part after "/static/documents/" of a backend file URL
// and may keep its query string. Staff only.
// GET /api/employments/static/documents/{path}
func (s *EmploymentsService) StaticDocument(ctx context.Context, path string) (*File, error) {
	r, err := employmentProxy("api/employments/static/documents/", path)
	if err != nil {
		return nil, err
	}
	return download(ctx, s.c, r)
}

// V1File downloads a file through the generic employment-backend proxy. path
// is the part after "/api/v1/" of a backend URL and may keep its query string. Staff only.
// GET /api/employments/api/v1/{path}
func (s *EmploymentsService) V1File(ctx context.Context, path string) (*File, error) {
	r, err := employmentProxy("api/employments/api/v1/", path)
	if err != nil {
		return nil, err
	}
	return download(ctx, s.c, r)
}

var employmentStaticRe = regexp.MustCompile(`(?:/employment)?/api/v1/static/documents/(.+)$`)

// DocumentFile downloads a file URL found in EmploymentDocument.Data or
// EmploymentSignedDocument.FileURL. Backend URLs are routed through the
// proxies ([EmploymentsService.StaticDocument] or
// [EmploymentsService.V1File]); a path starting with "/api/employments/" is
// fetched as is. Staff only.
func (s *EmploymentsService) DocumentFile(ctx context.Context, fileURL string) (*File, error) {
	if tail, ok := strings.CutPrefix(fileURL, "/api/employments/static/documents/"); ok {
		return s.StaticDocument(ctx, tail)
	}
	if tail, ok := strings.CutPrefix(fileURL, "/api/employments/api/v1/"); ok {
		return s.V1File(ctx, tail)
	}
	u, err := url.Parse(fileURL)
	if err != nil || u.Scheme == "" {
		return nil, errEmploymentFileURL
	}
	suffix := ""
	if u.RawQuery != "" {
		suffix = "?" + u.RawQuery
	}
	p := u.EscapedPath()
	if m := employmentStaticRe.FindStringSubmatch(p); m != nil {
		return s.StaticDocument(ctx, m[1]+suffix)
	}
	// Everything up to the last "/api/v1" is stripped.
	if i := strings.LastIndex(p, "/api/v1/"); i >= 0 {
		return s.V1File(ctx, p[i+len("/api/v1/"):]+suffix)
	}
	return nil, errEmploymentFileURL
}

var errEmploymentFileURL = errors.New("myitmo: not an employment file URL")

// employmentProxy builds a request for a proxied file path, keeping it under prefix.
func employmentProxy(prefix, path string) (*rest.Request, error) {
	path = strings.TrimLeft(path, "/")
	u, err := url.Parse(path)
	if err != nil || u.Scheme != "" || u.Host != "" || u.Path == "" {
		return nil, errors.New("myitmo: invalid employment file path")
	}
	for seg := range strings.SplitSeq(u.Path, "/") {
		if seg == ".." || seg == "." {
			return nil, errors.New("myitmo: invalid employment file path")
		}
	}
	target := prefix + u.EscapedPath()
	if u.RawQuery != "" {
		// Kept verbatim: backend URLs may be signed.
		target += "?" + u.RawQuery
	}
	return get(target, nil), nil
}
