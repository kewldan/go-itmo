package myitmo

import (
	"context"
	"net/http"
	"strconv"
	"time"
)

// DMSService is voluntary medical insurance (/api/dms).
type DMSService struct{ c *Client }

// Request statuses in DMSRequest.ReqStatus, DMSRequestRow.StatusID and
// DMSRequestDetails.StatusID. [DMSService.Decide] takes DMSStatusApproved or
// DMSStatusReturned.
const (
	// DMSStatusDeclined: the employee declined the university policy.
	DMSStatusDeclined = -1
	// DMSStatusAwaitingSignature: waiting for the signature to be confirmed.
	DMSStatusAwaitingSignature = -2
	// DMSStatusNew: insurance approved; the employee checks the data and signs.
	DMSStatusNew = 1
	// DMSStatusOnApproval: the data awaits HR approval.
	DMSStatusOnApproval = 2
	// DMSStatusApproved: the data was approved and must be signed again.
	DMSStatusApproved = 3
	// DMSStatusReturned: returned for rework with a comment.
	DMSStatusReturned = 4
	// DMSStatusConfirmed: the data is confirmed; the policy is being issued.
	DMSStatusConfirmed = 5
	// DMSStatusActive: the policy is active.
	DMSStatusActive = 6
)

// Values of DMSRequest.Gender and DMSPersonalData.Gender.
const (
	DMSGenderMale   = "M"
	DMSGenderFemale = "W"
)

// DMSCitizenshipRussia is the DMSDictItem.ID of Russia in [DMSService.Citizenships].
const DMSCitizenshipRussia = 1

// DMSRequest is the employee's own insurance request with personal data.
type DMSRequest struct {
	RequestID int64 `json:"request_id"`
	// ReqStatus is one of the DMSStatus* constants.
	ReqStatus            int    `json:"req_status"`
	LastName             string `json:"last_name"`
	FirstName            string `json:"first_name"`
	ParentalName         string `json:"parental_name"`
	BirthDate            Date   `json:"birth_date"`
	Gender               string `json:"gender"`
	Citizenship          string `json:"citizenship"`
	CitizenshipID        int64  `json:"citizenship_id"`
	PassportSeries       string `json:"passport_series"`
	PassportNumber       string `json:"passport_number"`
	PassportIssueDate    Date   `json:"passport_issue_date"`
	PassportIssueDep     string `json:"passport_issue_dep"`
	PassportIssuedBy     string `json:"passport_issued_by"`
	ResidenceFactAddress string `json:"residence_fact_address"`
	MobileNumber         string `json:"mobile_number"`
	// CorpMail is required before signing; see [DMSService.CreateCorpMail].
	CorpMail string `json:"corp_mail"`
	// Comment is the HR comment when the request was returned for rework.
	Comment                 string     `json:"comment"`
	CommentAuthorISU        *int64     `json:"comment_author_isu"`
	CommentAuthorSurname    string     `json:"comment_author_surname"`
	CommentAuthorName       string     `json:"comment_author_name"`
	CommentAuthorSecondName string     `json:"comment_author_second_name"`
	CommentCreatedAt        *time.Time `json:"comment_created_at"`
}

// DMSContacts is support contacts and insurance programme information.
type DMSContacts struct {
	UniversitySupport DMSUniversitySupport `json:"university_support"`
	DMSSupport        DMSSupport           `json:"DMS_support"`
	InsuranceSupport  DMSInsuranceSupport  `json:"insurance_support"`
	InformationLink   string               `json:"information_link"`
	InsuranceData     DMSInsuranceData     `json:"insurance_data"`
}

// DMSUniversitySupport is the university support contact.
type DMSUniversitySupport struct {
	ContactLink         string `json:"contact_link"`
	ContactSupportEmail string `json:"contact_support_email"`
	ContactPhoneNumber  string `json:"contact_phone_number"`
}

// DMSSupport is the insurance programme support contact.
type DMSSupport struct {
	ContactLink  string `json:"contact_link"`
	ContactEmail string `json:"contact_email"`
}

// DMSInsuranceSupport is the insurer's contact.
type DMSInsuranceSupport struct {
	InsuranceCompany              string `json:"insurance_company"`
	InsuranceCompanyContactNumber string `json:"insurance_company_contact_number"`
	// ContactSupportEmail is the doctor curator's e-mail.
	ContactSupportEmail string `json:"contact_support_email"`
}

// DMSInsuranceData describes the insurance programme.
type DMSInsuranceData struct {
	ProgramName                   string      `json:"program_name"`
	InsuranceCompany              string      `json:"insurance_company"`
	InsuranceCompanyLink          string      `json:"insurance_company_link"`
	InsuranceCompanyContactNumber string      `json:"insurance_company_contact_number"`
	QRCodes                       []DMSQRCode `json:"qr_codes"`
}

// DMSQRCode is a QR code image of the programme.
type DMSQRCode struct {
	// QRCode is a base64-encoded PNG.
	QRCode     string `json:"qr_code"`
	QRCodeDesc string `json:"qr_code_desc"`
}

// DMSPersonalData is the edited personal data sent for HR approval.
type DMSPersonalData struct {
	LastName     string
	FirstName    string
	ParentalName string // optional
	BirthDate    Date
	// Gender is DMSGenderMale or DMSGenderFemale.
	Gender            string
	CitizenshipID     int64
	PassportSeries    string
	PassportNumber    string
	PassportIssueDate Date
	// PassportIssueDep is the issuing department code.
	PassportIssueDep     string
	PassportIssuedBy     string
	ResidenceFactAddress string
	// ResidenceApartment is the apartment number: the last number of the
	// address, or "1" when there is none (the default here).
	ResidenceApartment string
	MobileNumber       string
	CorpMail           string
}

func (d DMSPersonalData) fields() map[string]string {
	apartment := d.ResidenceApartment
	if apartment == "" {
		apartment = "1"
	}
	f := map[string]string{
		"last_name":              d.LastName,
		"first_name":             d.FirstName,
		"birth_date":             d.BirthDate.String(),
		"gender":                 d.Gender,
		"citizenship_id":         strconv.FormatInt(d.CitizenshipID, 10),
		"passport_series":        d.PassportSeries,
		"passport_number":        d.PassportNumber,
		"passport_issue_date":    d.PassportIssueDate.String(),
		"passport_issue_dep":     d.PassportIssueDep,
		"passport_issued_by":     d.PassportIssuedBy,
		"residence_fact_address": d.ResidenceFactAddress,
		"residence_apartment":    apartment,
		"mobile_number":          d.MobileNumber,
		"corp_mail":              d.CorpMail,
	}
	if d.ParentalName != "" {
		f["parental_name"] = d.ParentalName
	}
	return f
}

// DMSSignTask identifies an e-signature task for the consent (see [SignService]).
type DMSSignTask struct {
	SignatureID FlexID `json:"signature_id"`
	TaskID      FlexID `json:"task_id"`
}

// DMSDictItem is an entry of a DMS dictionary (citizenships, status filters).
type DMSDictItem struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// DMSListParams filters the HR request list.
type DMSListParams struct {
	// Limit is the page size (typically 20); Offset is an item offset.
	Limit  int
	Offset int
	// StatusID filters by status (DMSStatus* constants).
	StatusID *int64
	// SortBy is a comma list of full_name, corp_mail, last_update, status_id.
	SortBy string
	// SortOrder is a comma list of asc/desc matching SortBy.
	SortOrder string
	// ApprovalNeeded keeps only the requests awaiting the user's approval.
	ApprovalNeeded bool
	// Query is the search text.
	Query string
}

// DMSRequestList is a page of the HR request list.
type DMSRequestList struct {
	Count           int             `json:"count"`
	OnApprovalCount int             `json:"on_approval_count"`
	Requests        []DMSRequestRow `json:"requests"`
}

// DMSRequestRow is a row of the HR request list.
type DMSRequestRow struct {
	RequestID int64 `json:"request_id"`
	// PersID is the employee's ISU number.
	PersID   int64  `json:"pers_id"`
	FullName string `json:"full_name"`
	CorpMail string `json:"corp_mail"`
	// LastUpdate is a preformatted date.
	LastUpdate string `json:"last_update"`
	StatusID   int    `json:"status_id"`
	StatusName string `json:"status_name"`
}

// DMSRequestDetails is the HR view of one request. Dates are preformatted strings.
type DMSRequestDetails struct {
	RequestID            int64  `json:"request_id"`
	PersID               int64  `json:"pers_id"`
	LastName             string `json:"last_name"`
	FirstName            string `json:"first_name"`
	ParentalName         string `json:"parental_name"`
	BirthDate            string `json:"birth_date"`
	Gender               string `json:"gender"`
	Citizenship          string `json:"citizenship"`
	PassportSeries       string `json:"passport_series"`
	PassportNumber       string `json:"passport_number"`
	PassportIssueDate    string `json:"passport_issue_date"`
	PassportIssueDep     string `json:"passport_issue_dep"`
	PassportIssuedBy     string `json:"passport_issued_by"`
	ResidenceFactAddress string `json:"residence_fact_address"`
	MobileNumber         string `json:"mobile_number"`
	CorpMail             string `json:"corp_mail"`
	// UserFileName is the supporting document; see [DMSService.File].
	UserFileName string `json:"user_file_name"`
	// StatusID DMSStatusOnApproval means a decision is expected.
	StatusID   int            `json:"status_id"`
	StatusName string         `json:"status_name"`
	Rejections []DMSRejection `json:"rejections"`
}

// DMSRejection is an earlier return for rework.
type DMSRejection struct {
	Comment           string `json:"comment"`
	CommenterPersID   int64  `json:"commenter_pers_id"`
	CommenterFullName string `json:"commenter_full_name"`
	CommenterDate     string `json:"commenter_date"`
	CommenterTime     string `json:"commenter_time"`
}

// Access reports whether the user may use the HR approval view.
// Staff only.
// GET /api/dms/me/access
func (s *DMSService) Access(ctx context.Context) (bool, error) {
	return call[bool](ctx, s.c, get("api/dms/me/access", nil))
}

// MyRequest returns the employee's insurance request, or nil when there is none.
// Staff only.
// GET /api/dms/me/request
func (s *DMSService) MyRequest(ctx context.Context) (*DMSRequest, error) {
	return call[*DMSRequest](ctx, s.c, get("api/dms/me/request", nil))
}

// Contacts returns support contacts and insurance programme information.
// Staff only.
// GET /api/dms/me/request/contacts
func (s *DMSService) Contacts(ctx context.Context) (*DMSContacts, error) {
	return call[*DMSContacts](ctx, s.c, get("api/dms/me/request/contacts", nil))
}

// SubmitPersonalData sends edited personal data for HR approval. files are
// supporting PDF documents; each is sent as a "files" part whatever its Field.
// Staff only.
// POST /api/dms/me/request/{request_id}
func (s *DMSService) SubmitPersonalData(ctx context.Context, requestID int64, data DMSPersonalData, files ...Upload) error {
	parts := make([]Upload, len(files))
	for i, f := range files {
		f.Field = "files"
		parts[i] = f
	}
	return exec(ctx, s.c, multipart(http.MethodPost, "api/dms/me/request/"+id(requestID), data.fields(), parts...))
}

// Sign creates the e-signature tasks for the consent to transfer data to the insurer.
// parentalName may be empty.
// Staff only.
// POST /api/dms/me/request/{request_id}/sign
func (s *DMSService) Sign(ctx context.Context, requestID int64, lastName, firstName, parentalName string) ([]DMSSignTask, error) {
	body := struct {
		LastName     string `json:"last_name"`
		FirstName    string `json:"first_name"`
		ParentalName string `json:"parental_name"`
	}{lastName, firstName, parentalName}
	return call[[]DMSSignTask](ctx, s.c, post("api/dms/me/request/"+id(requestID)+"/sign", body))
}

// Reject declines the university insurance policy.
// Staff only.
// POST /api/dms/me/request/{request_id}/reject
func (s *DMSService) Reject(ctx context.Context, requestID int64) error {
	return exec(ctx, s.c, post("api/dms/me/request/"+id(requestID)+"/reject", nil))
}

// CreateCorpMail creates the corporate @itmo.ru mailbox that the insurance
// requires. username is the local part.
// Staff only.
// POST /api/dms/me/corp_mail
func (s *DMSService) CreateCorpMail(ctx context.Context, username, password string) error {
	body := struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}{username, password}
	return exec(ctx, s.c, post("api/dms/me/corp_mail", body))
}

// Citizenships returns the citizenship dictionary.
// Staff only.
// GET /api/dms/citizenships
func (s *DMSService) Citizenships(ctx context.Context) ([]DMSDictItem, error) {
	return call[[]DMSDictItem](ctx, s.c, get("api/dms/citizenships", nil))
}

// StatusFilters returns the status filter options of the HR request list.
// Staff only (HR approvers).
// GET /api/dms/filters/statuses
func (s *DMSService) StatusFilters(ctx context.Context) ([]DMSDictItem, error) {
	return call[[]DMSDictItem](ctx, s.c, get("api/dms/filters/statuses", nil))
}

// List returns a page of employees' insurance requests.
// Staff only (HR approvers).
// GET /api/dms/request/list
func (s *DMSService) List(ctx context.Context, p DMSListParams) (*DMSRequestList, error) {
	qv := q().set("limit", p.Limit).set("offset", p.Offset).
		set("status_id", p.StatusID).
		set("sort_by", p.SortBy).
		set("sort_order", p.SortOrder).
		set("approval_needed", p.ApprovalNeeded).
		set("query", p.Query)
	return call[*DMSRequestList](ctx, s.c, get("api/dms/request/list", qv))
}

// Details returns one request of the HR list.
// Staff only (HR approvers).
// GET /api/dms/request/list/{request_id}
func (s *DMSService) Details(ctx context.Context, requestID int64) (*DMSRequestDetails, error) {
	return call[*DMSRequestDetails](ctx, s.c, get("api/dms/request/list/"+id(requestID), nil))
}

// Decide approves a request (DMSStatusApproved) or returns it for rework
// (DMSStatusReturned, comment required).
// Staff only (HR approvers).
// POST /api/dms/request/me/{request_id}/decision
func (s *DMSService) Decide(ctx context.Context, requestID int64, statusID int, comment string) error {
	body := struct {
		RequestID int64  `json:"request_id"`
		StatusID  int    `json:"status_id"`
		Comment   string `json:"comment,omitzero"`
	}{requestID, statusID, comment}
	return exec(ctx, s.c, post("api/dms/request/me/"+id(requestID)+"/decision", body))
}

// File downloads the supporting document of a request (DMSRequestDetails.UserFileName).
// Staff only (HR approvers).
// GET /api/dms/request/{request_id}/files/{file_name}
func (s *DMSService) File(ctx context.Context, requestID int64, fileName string) (*File, error) {
	return download(ctx, s.c, get("api/dms/request/"+id(requestID)+"/files/"+id(fileName), nil))
}
