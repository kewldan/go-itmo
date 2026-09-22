package myitmo

import (
	"context"
	"time"

	"github.com/kewldan/go-itmo/internal/jsonx"
)

// GIAOrderStatusAwaitingApproval is the order status_id shown when the order
// awaits the current user's approval.
const GIAOrderStatusAwaitingApproval = 20

// GIAOrderPerson is the author of an order or a comment.
type GIAOrderPerson struct {
	FIO string `json:"fio"`
	ISU int64  `json:"isu"`
}

// GIAOrderRow is an order of the order list.
type GIAOrderRow struct {
	OrderID        int64     `json:"order_id"`
	Number         string    `json:"number"`
	OrderType      GIAIDName `json:"order_type"`
	EducationLevel GIAIDName `json:"education_level"`
	// Implementer, Contingent and CreatedBy have unknown shapes.
	Implementer  RawJSON   `json:"implementer,omitzero"`
	Contingent   RawJSON   `json:"contingent,omitzero"`
	FormDate     time.Time `json:"form_date"`
	ApproveDate  time.Time `json:"approve_date"`
	CreatedBy    RawJSON   `json:"created_by,omitzero"`
	Status       GIAStatus `json:"status"`
	ScanFileLink string    `json:"scan_file_link"`
}

// GIAOrderList is a page of orders. The server nests it as
// {"result": {"result": [...], "count": n}}.
type GIAOrderList struct {
	Result []GIAOrderRow `json:"result"`
	Count  int           `json:"count"`
}

// GIAOrderSigner is a person who approves an order.
type GIAOrderSigner struct {
	FIO       string    `json:"fio"`
	ISU       int64     `json:"isu"`
	JobTitle  string    `json:"job_title"`
	Status    GIAStatus `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

// GIAOrderComment is a comment on an order.
type GIAOrderComment struct {
	Comment   string         `json:"comment"`
	CreatedAt time.Time      `json:"created_at"`
	CreatedBy GIAOrderPerson `json:"created_by"`
}

// GIAOrderReviewer is a reviewer listed in an order.
type GIAOrderReviewer struct {
	FIO           string `json:"fio"`
	AcademicTitle string `json:"academic_title"`
	JobTitle      string `json:"job_title"`
	WorkPlace     string `json:"work_place"`
	Degree        string `json:"degree"`
}

// GIAOrderStudent is a student listed in an order.
type GIAOrderStudent struct {
	StudentID  int64  `json:"student_id"`
	FIO        string `json:"fio"`
	ISU        int64  `json:"isu"`
	Theme      string `json:"theme"`
	Supervisor string `json:"supervisor"`
	// Consultants has an unknown item shape.
	Consultants RawJSON            `json:"consultants,omitzero"`
	Reviewers   []GIAOrderReviewer `json:"reviewers"`
}

// GIAOrderImplementer is the faculty that implements the programme of an order block.
type GIAOrderImplementer struct {
	FacultyShortName string `json:"faculty_short_name"`
	Faculty          string `json:"faculty"`
}

// GIAOrderAttachment is a block of an order: a group of one programme.
type GIAOrderAttachment struct {
	EducationLevel GIAIDName           `json:"education_level"`
	Implementer    GIAOrderImplementer `json:"implementer"`
	Direction      GIAEduDirection     `json:"direction"`
	EP             GIAEduProgram       `json:"ep"`
	GroupID        string              `json:"group_id"`
	Students       []GIAOrderStudent   `json:"students"`
}

// GIAOrder is an order (приказ) with its signers, comments and content.
type GIAOrder struct {
	ID        int64             `json:"id"`
	OrderType GIAIDName         `json:"order_type"`
	FormDate  time.Time         `json:"form_date"`
	UpdatedAt time.Time         `json:"updated_at"`
	CreatedBy GIAOrderPerson    `json:"created_by"`
	Status    GIAStatus         `json:"status"`
	Signers   []GIAOrderSigner  `json:"signers"`
	Comments  []GIAOrderComment `json:"comments"`
	// Attachment is one [GIAOrderAttachment] or a list of them; see [GIAOrder.Attachments].
	Attachment RawJSON `json:"attachment,omitzero"`
}

// Attachments decodes Attachment, which the server sends as an object or a list.
func (o *GIAOrder) Attachments() ([]GIAOrderAttachment, error) {
	switch o.Attachment.Kind() {
	case '[':
		var list []GIAOrderAttachment
		err := jsonx.Unmarshal(o.Attachment, &list)
		return list, err
	case '{':
		var one GIAOrderAttachment
		err := jsonx.Unmarshal(o.Attachment, &one)
		return []GIAOrderAttachment{one}, err
	}
	return nil, nil
}

// GIAOrderCandidate is a student who can be put in a new order.
type GIAOrderCandidate struct {
	StudentID  int64  `json:"student_id"`
	FIO        string `json:"fio"`
	ISU        int64  `json:"isu"`
	GroupID    string `json:"group_id"`
	Theme      string `json:"theme"`
	Supervisor string `json:"supervisor"`
	Consultant string `json:"consultant"`
	// Reviewer and Reviewers have unknown shapes.
	Reviewer  RawJSON         `json:"reviewer,omitzero"`
	Reviewers RawJSON         `json:"reviewers,omitzero"`
	Direction GIAEduDirection `json:"direction"`
}

// GIAOrderInput is the body of [GIAService.CreateOrder].
type GIAOrderInput struct {
	FormDate    Date  `json:"form_date"`
	OrderTypeID int64 `json:"order_type_id"`
	// StudyYear is the end year of the academic year.
	StudyYear        int     `json:"study_year"`
	EducationLevelID int64   `json:"education_level_id"`
	ImplementerID    int64   `json:"implementer_id"`
	Students         []int64 `json:"students"`
}

// GIAOrdersParams filters [GIAService.Orders]. Zero values are not sent,
// except Limit and Offset.
type GIAOrdersParams struct {
	Limit, Offset    int
	Query            string
	OrderTypeID      int64
	ImplementerID    int64
	EducationLevelID int64
	StatusID         int
	// NeedApprove lists only the orders awaiting the user's approval.
	NeedApprove bool
}

// GIAOrderCandidatesParams filters [GIAService.OrderCandidates]. Zero values
// are not sent.
type GIAOrderCandidatesParams struct {
	Year             int
	ImplementerID    int64
	EducationLevelID int64
	// TypeID is the order type.
	TypeID        int64
	Limit, Offset int
	Query         string
}

// Orders returns a page of orders.
// Staff only.
// GET /api/gia/orders/list
func (s *GIAService) Orders(ctx context.Context, p GIAOrdersParams) (*GIAOrderList, error) {
	v := q().set("limit", p.Limit).set("offset", p.Offset).set("query", p.Query)
	giaSet(v, "order_type_id", p.OrderTypeID)
	giaSet(v, "implementer_id", p.ImplementerID)
	giaSet(v, "education_level_id", p.EducationLevelID)
	giaSet(v, "status_id", p.StatusID)
	if p.NeedApprove {
		v.set("need_approve", 1)
	}
	return call[*GIAOrderList](ctx, s.c, get("api/gia/orders/list", v))
}

// CreateOrder creates an order. The result shape is not known; it is
// returned as is. A non-zero
// error_code is an *Error.
// Staff only.
// POST /api/gia/orders/
func (s *GIAService) CreateOrder(ctx context.Context, in GIAOrderInput) (RawJSON, error) {
	if in.Students == nil {
		in.Students = []int64{}
	}
	return call[RawJSON](ctx, s.c, post("api/gia/orders/", in))
}

// Order returns an order.
// Staff only.
// GET /api/gia/orders/{order_id}
func (s *GIAService) Order(ctx context.Context, orderID int64) (*GIAOrder, error) {
	return call[*GIAOrder](ctx, s.c, get("api/gia/orders/"+id(orderID), nil))
}

// UpdateOrderStudents replaces the students of a returned order.
// Staff only.
// PATCH /api/gia/orders/{order_id}
func (s *GIAService) UpdateOrderStudents(ctx context.Context, orderID int64, studentIDs ...int64) error {
	body := struct {
		Students []int64 `json:"students"`
	}{studentIDs}
	if body.Students == nil {
		body.Students = []int64{}
	}
	return exec(ctx, s.c, patch("api/gia/orders/"+id(orderID), body))
}

// ApproveOrder approves (signs off) an order.
// Staff only.
// POST /api/gia/orders/{order_id}/approve
func (s *GIAService) ApproveOrder(ctx context.Context, orderID int64) error {
	return exec(ctx, s.c, post("api/gia/orders/"+id(orderID)+"/approve", struct{}{}))
}

// DeclineOrder declines an order; an empty comment sends {}.
// Staff only.
// POST /api/gia/orders/{order_id}/decline
func (s *GIAService) DeclineOrder(ctx context.Context, orderID int64, comment string) error {
	body := struct {
		Comment string `json:"comment,omitzero"`
	}{comment}
	return exec(ctx, s.c, post("api/gia/orders/"+id(orderID)+"/decline", body))
}

// OrderFile downloads the order document as PDF.
// Staff only.
// GET /api/gia/orders/{order_id}/file
func (s *GIAService) OrderFile(ctx context.Context, orderID int64) (*File, error) {
	return download(ctx, s.c, get("api/gia/orders/"+id(orderID)+"/file", q().set("format", GIAFormatPDF)))
}

// OrderCandidates returns the students who can be put in a new order.
// Staff only.
// GET /api/gia/orders/students/list
func (s *GIAService) OrderCandidates(ctx context.Context, p GIAOrderCandidatesParams) ([]GIAOrderCandidate, error) {
	v := giaSet(q(), "year", p.Year)
	giaSet(v, "implementer_id", p.ImplementerID)
	giaSet(v, "education_level_id", p.EducationLevelID)
	giaSet(v, "type_id", p.TypeID)
	giaSet(v, "limit", p.Limit)
	giaSet(v, "offset", p.Offset)
	v.set("query", p.Query)
	return call[[]GIAOrderCandidate](ctx, s.c, get("api/gia/orders/students/list", v))
}
