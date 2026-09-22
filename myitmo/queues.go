package myitmo

import (
	"bytes"
	"context"
	"net/http"
	"time"

	"github.com/kewldan/go-itmo/internal/jsonx"
	"github.com/kewldan/go-itmo/internal/rest"
)

// QueuesService is electronic queues (/api/queues).
type QueuesService struct{ c *Client }

// QueueTable is a queue the user can sign up for.
type QueueTable struct {
	TableID      int64  `json:"table_id"`
	TypeID       int64  `json:"type_id"`
	TableName    string `json:"table_name"`
	TableAddress string `json:"table_address"`
	// Comment is an HTML description.
	Comment string `json:"comment"`
	// TemplateID is set when a request form must be submitted with the
	// sign-up; pass the id of that request as QueueSignUp.RequestID.
	TemplateID *int64 `json:"template_id"`
}

// QueueEntry is one of the user's queue sign-ups.
type QueueEntry struct {
	ApplicationID int64     `json:"application_id"`
	TypeID        int64     `json:"type_id"`
	TableID       int64     `json:"table_id"`
	TableName     string    `json:"table_name"`
	TableAddress  string    `json:"table_address"`
	Date          time.Time `json:"date"`
	Phone         string    `json:"phone"`
	// Comment is the user's comment.
	Comment string `json:"comment"`
	// QueueComment is an HTML note of the queue.
	QueueComment string `json:"queue_comment"`
}

// QueueSlot is a free time slot of a queue.
type QueueSlot struct {
	TimeTableID int64     `json:"time_table_id"`
	Date        time.Time `json:"date"`
}

// QueueSignUp is a sign-up for a queue slot.
type QueueSignUp struct {
	// TimeTableID is QueueSlot.TimeTableID.
	TimeTableID int64 `json:"time_table_id"`
	// Phone must contain at least 11 digits, e.g. "+7(###)###-##-##".
	Phone   string  `json:"phone"`
	Comment *string `json:"comment"`
	// RequestID is the id of the submitted request when QueueTable.TemplateID is set.
	RequestID *int64 `json:"request_id"`
}

// queueCancel is the body of a cancellation.
type queueCancel struct {
	TypeID        int64 `json:"type_id"`
	ApplicationID int64 `json:"application_id"`
}

// List returns the queues the user can sign up for.
// GET /api/queues/
func (s *QueuesService) List(ctx context.Context) ([]QueueTable, error) {
	return call[[]QueueTable](ctx, s.c, get("api/queues/", nil))
}

// Get returns one queue; nil when none matches.
// GET /api/queues/
func (s *QueuesService) Get(ctx context.Context, tableID, typeID int64) (*QueueTable, error) {
	list, err := call[[]QueueTable](ctx, s.c, get("api/queues/", q().set("table_id", tableID).set("type_id", typeID)))
	if err != nil || len(list) == 0 {
		return nil, err
	}
	return &list[0], nil
}

// Current returns the user's active sign-ups.
// GET /api/queues/current
func (s *QueuesService) Current(ctx context.Context) ([]QueueEntry, error) {
	return call[[]QueueEntry](ctx, s.c, get("api/queues/current", nil))
}

// Archive returns the user's past sign-ups.
// GET /api/queues/archive
func (s *QueuesService) Archive(ctx context.Context) ([]QueueEntry, error) {
	return call[[]QueueEntry](ctx, s.c, get("api/queues/archive", nil))
}

// Slots returns the free time slots of a queue.
// GET /api/queues/slots
func (s *QueuesService) Slots(ctx context.Context, tableID, typeID int64) ([]QueueSlot, error) {
	return call[[]QueueSlot](ctx, s.c, get("api/queues/slots", q().set("table_id", tableID).set("type_id", typeID)))
}

// SignUp takes a queue slot. The body is JSON sent with the content type
// application/x-www-form-urlencoded, as the server expects.
// POST /api/queues/
func (s *QueuesService) SignUp(ctx context.Context, p QueueSignUp) error {
	r, err := queueStringBody(http.MethodPost, "api/queues/", p)
	if err != nil {
		return err
	}
	return exec(ctx, s.c, r)
}

// Cancel removes a sign-up.
// DELETE /api/queues/
func (s *QueuesService) Cancel(ctx context.Context, typeID, applicationID int64) error {
	return exec(ctx, s.c, del("api/queues/", queueCancel{TypeID: typeID, ApplicationID: applicationID}))
}

// queueStringBody sends body as JSON text with the content type
// application/x-www-form-urlencoded.
func queueStringBody(method, path string, body any) (*rest.Request, error) {
	data, err := jsonx.Marshal(body)
	if err != nil {
		return nil, err
	}
	return &rest.Request{Method: method, Path: path, Body: bytes.NewReader(data), ContentType: "application/x-www-form-urlencoded"}, nil
}
