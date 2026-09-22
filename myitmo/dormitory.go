package myitmo

import (
	"context"
	"time"
)

// DormitoryService is dormitory settlement and payments (/api/dormitory).
type DormitoryService struct{ c *Client }

// Settlement statuses in DormStatus.StatusID.
const (
	// DormStatusQueued: waiting for a place, in the queue.
	DormStatusQueued = 1
	// DormStatusOffered: a place is offered; a check-in time must be chosen
	// within 48 hours (24 hours for some buildings).
	DormStatusOffered = 2
	// DormStatusAppointed: a check-in time is booked.
	DormStatusAppointed = 3
	// DormStatusSettled: settled; contracts and payments are available.
	DormStatusSettled = 4
)

// DormTypeDorm is DormStatus.DormType for a regular dormitory; other values mean apartments.
const DormTypeDorm = "dorm"

// DormStatus is the user's settlement status.
type DormStatus struct {
	// StatusID is one of the DormStatus* constants.
	StatusID int    `json:"statusId"`
	DormType string `json:"dormType"`
	// QueuePlace is the place in the waiting queue.
	QueuePlace        *int                   `json:"queuePlace"`
	AssignedDormitory *DormAssignedDormitory `json:"assignedDormitory"`
	Appointment       *DormAppointment       `json:"appointment"`
}

// DormAssignedDormitory is the dormitory offered or assigned to the user.
type DormAssignedDormitory struct {
	BuildingID int64  `json:"buildingId"`
	Name       string `json:"name"`
	Address    string `json:"address"`
	// SettlementProcedureText is HTML.
	SettlementProcedureText string `json:"settlementProcedureText"`
}

// DormAppointment is a booked check-in time.
type DormAppointment struct {
	TimeSlot time.Time `json:"timeSlot"`
	// CanRemoveUntil is the deadline for [DormitoryService.RemoveAppointment].
	CanRemoveUntil time.Time `json:"canRemoveUntil"`
}

// Medical document statuses in DormDocument.Status.
const (
	DormDocumentRejected    = -1
	DormDocumentNotUploaded = 0
	DormDocumentInReview    = 1
	DormDocumentAccepted    = 2
)

// DormDocument is a required medical document and its check status.
type DormDocument struct {
	DocumentTypeName string `json:"documentTypeName"`
	// Status is one of the DormDocument* constants.
	Status int `json:"status"`
	// TemplateID is the request template used to upload the document (requests service).
	TemplateID FlexID `json:"templateId"`
}

// DormApartment is an ITMO.Aparts option.
type DormApartment struct {
	ID      FlexID `json:"id"`
	Name    string `json:"name"`
	Address string `json:"address"`
}

// DormTimeSlot is a free check-in time.
type DormTimeSlot struct {
	TimeSlotID int64     `json:"timeSlotId"`
	TimeSlot   time.Time `json:"timeSlot"`
}

// DormPeriod is a period with accommodation contracts.
type DormPeriod struct {
	Year     int  `json:"year"`
	DateFrom Date `json:"dateFrom"`
	DateTo   Date `json:"dateTo"`
}

// Status returns the settlement status, or nil when the user has no settlement.
// GET /api/dormitory/status/full
func (s *DormitoryService) Status(ctx context.Context) (*DormStatus, error) {
	return call[*DormStatus](ctx, s.c, get("api/dormitory/status/full", nil))
}

// Documents returns the required medical documents and their status.
// GET /api/dormitory/documents
func (s *DormitoryService) Documents(ctx context.Context) ([]DormDocument, error) {
	return call[[]DormDocument](ctx, s.c, get("api/dormitory/documents", nil))
}

// Apartments returns the ITMO.Aparts options offered instead of a dormitory.
// GET /api/dormitory/apartments
func (s *DormitoryService) Apartments(ctx context.Context) ([]DormApartment, error) {
	return call[[]DormApartment](ctx, s.c, get("api/dormitory/apartments", nil))
}

// Slots returns the free check-in times (status DormStatusOffered).
// GET /api/dormitory/settlement/slots
func (s *DormitoryService) Slots(ctx context.Context) ([]DormTimeSlot, error) {
	return call[[]DormTimeSlot](ctx, s.c, get("api/dormitory/settlement/slots", nil))
}

// Register books a check-in time.
// POST /api/dormitory/settlement/register
func (s *DormitoryService) Register(ctx context.Context, timeSlotID int64) error {
	body := struct {
		TimeSlotID int64 `json:"timeSlotId"`
	}{timeSlotID}
	return exec(ctx, s.c, post("api/dormitory/settlement/register", body))
}

// RemoveAppointment cancels the booked check-in time so that another can be chosen.
// Allowed until DormAppointment.CanRemoveUntil.
// DELETE /api/dormitory/settlement/remove
func (s *DormitoryService) RemoveAppointment(ctx context.Context) error {
	return exec(ctx, s.c, del("api/dormitory/settlement/remove", nil))
}

// Cancel withdraws from settlement. Request template 4827 is submitted first.
// DELETE /api/dormitory/settlement/cancel
func (s *DormitoryService) Cancel(ctx context.Context) error {
	return exec(ctx, s.c, del("api/dormitory/settlement/cancel", nil))
}

// ChangeCategory switches from a dormitory to ITMO.Aparts. Request template
// 5587 with the chosen apartments is submitted first.
// POST /api/dormitory/settlement/change_category
func (s *DormitoryService) ChangeCategory(ctx context.Context) error {
	return exec(ctx, s.c, post("api/dormitory/settlement/change_category", nil))
}

// ContractPeriods returns the periods that have accommodation contracts.
// GET /api/dormitory/payments/contracts/periods
func (s *DormitoryService) ContractPeriods(ctx context.Context) ([]DormPeriod, error) {
	return call[[]DormPeriod](ctx, s.c, get("api/dormitory/payments/contracts/periods", nil))
}

// Contracts returns the accommodation contracts of a period (DormPeriod.DateFrom and DateTo).
// GET /api/dormitory/payments/contracts
func (s *DormitoryService) Contracts(ctx context.Context, from, to Date) ([]PaymentContract, error) {
	return call[[]PaymentContract](ctx, s.c, get("api/dormitory/payments/contracts", q().set("from", from).set("to", to)))
}

// Pay starts an online accommodation payment and returns the payment gateway URL.
// POST /api/dormitory/payments/pay
func (s *DormitoryService) Pay(ctx context.Context, req PayRequest) (string, error) {
	return call[string](ctx, s.c, post("api/dormitory/payments/pay", req))
}
