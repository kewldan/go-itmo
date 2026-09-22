package myitmo

import (
	"context"
	"encoding/json/jsontext"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/kewldan/go-itmo/internal/jsonx"
	"github.com/kewldan/go-itmo/internal/rest"
)

// RequestsService is legacy requests and applications (/api/requests).
type RequestsService struct{ c *Client }

// requestsFlexText returns the contents of a JSON string, "" for null, and the raw
// text of any other value (numbers, booleans, objects, arrays).
func requestsFlexText(b []byte) (string, error) {
	b = []byte(strings.TrimSpace(string(b)))
	switch {
	case len(b) == 0 || string(b) == "null":
		return "", nil
	case b[0] == '"':
		var s string
		err := jsonx.Unmarshal(b, &s)
		return s, err
	case jsontext.Value(b).IsValid():
		return string(b), nil
	}
	return "", fmt.Errorf("myitmo: invalid JSON value %q", b)
}

// RequestStatus is the numeric status of a legacy request in [RequestSummary].
type RequestStatus int

// Legacy request statuses (as the "My requests" page renders them).
const (
	RequestStatusInProgress RequestStatus = 0
	RequestStatusDone       RequestStatus = 1
	RequestStatusRejected   RequestStatus = 2
)

// RequestFieldType is the type of a legacy form field.
type RequestFieldType string

// Legacy form field types; they also select the value encoding of [RequestFieldValue].
const (
	RequestFieldText     RequestFieldType = "text"
	RequestFieldNumber   RequestFieldType = "number"
	RequestFieldDate     RequestFieldType = "date"
	RequestFieldDateTime RequestFieldType = "date_time"
	// RequestFieldDictionary takes values from [RequestsService.Dictionary].
	RequestFieldDictionary RequestFieldType = "dictionary"
	// RequestFieldFile takes the name returned by [RequestsService.Upload].
	RequestFieldFile RequestFieldType = "file"
	// RequestFieldMultiple is a repeatable block of sub-fields (a table).
	RequestFieldMultiple RequestFieldType = "multiple_field"
)

// RequestCategory is a group of legacy request templates in the catalog.
type RequestCategory struct {
	ID   FlexID `json:"id"`
	Name string `json:"name"`
	// Color is a palette key: "0".."10" or "default".
	Color FlexID `json:"color"`
	// Icon is a CSS class such as "icon icon-other".
	Icon     string            `json:"icon"`
	Requests []RequestTemplate `json:"requests"`
}

// RequestTemplate is a legacy request type that can be filled in.
type RequestTemplate struct {
	// ID is the template id for [RequestsService.Template] and [RequestsService.Send].
	ID         int64  `json:"id"`
	Name       string `json:"name"`
	CategoryID FlexID `json:"category_id,omitzero"`
}

// RequestForm is the form definition of a legacy request template.
type RequestForm struct {
	// TemplateDescription is HTML shown above the form.
	TemplateDescription string                `json:"template_description"`
	TemplateFiles       []RequestTemplateFile `json:"template_files"`
	// UserCanApplyNow is false when the form must not be shown; see UserCannotApplyNowReason.
	UserCanApplyNow          bool           `json:"user_can_apply_now"`
	UserCannotApplyNowReason string         `json:"user_cannot_apply_now_reason"`
	Fields                   []RequestField `json:"fields_data"`
}

// RequestTemplateFile is a blank attached to a template; download it with [RequestsService.TemplateFile].
type RequestTemplateFile struct {
	FileID   FlexID `json:"file_id"`
	FileName string `json:"file_name"`
}

// RequestField is one field of a legacy form.
type RequestField struct {
	FieldID   int64  `json:"field_id"`
	FieldName string `json:"field_name"`
	FieldNote string `json:"field_note"`
	// FieldType is one of the RequestField* constants.
	FieldType RequestFieldType `json:"field_type"`
	// FieldSize "S"/"s" marks a narrow field; for text it selects input over textarea.
	FieldSize         string `json:"field_size"`
	RequiredFieldFlag bool   `json:"required_field_flag"`
	// DefaultValue is the initial value; its type depends on the field.
	DefaultValue RawJSON `json:"default_value,omitzero"`
	// MultipleChoice allows several dictionary values.
	MultipleChoice bool `json:"multiple_choice"`
	// DictionaryID is the dictionary for [RequestsService.Dictionary].
	DictionaryID FlexID `json:"dictionary_id"`
	// DependentFieldID is the field whose value is passed as dep to the dictionary.
	DependentFieldID FlexID `json:"dependent_field_id"`
	// InitDictionary is the preselected dictionary item for DefaultValue.
	InitDictionary *RequestDictItem `json:"init_dictionary"`
	// ShowConditionFlag is the initial visibility; see [RequestsService.FormUpdate].
	ShowConditionFlag     bool              `json:"show_condition_flag"`
	MultipleFieldMaxBlock int               `json:"multiple_field_max_block"`
	MultipleFieldData     []RequestSubField `json:"multiple_field_data"`
}

// RequestSubField is a column of a repeatable (multiple_field) block.
type RequestSubField struct {
	Name       string           `json:"name"`
	Type       RequestFieldType `json:"type"`
	Size       string           `json:"size"`
	Required   bool             `json:"required"`
	Notice     string           `json:"notice"`
	Dictionary FlexID           `json:"dictionary"`
	DependsID  FlexID           `json:"dependsId"`
}

// RequestDictItem is a value of a legacy dictionary.
type RequestDictItem struct {
	// ID is the value submitted in [RequestFieldValue].
	ID   FlexID `json:"id"`
	Text string `json:"text"`
}

// RequestFieldValue is the value of one field sent to [RequestsService.FormUpdate] and [RequestsService.Send].
//
// Value encoding by type: dictionary → the selected id (several ids joined by ","),
// date → "DD.MM.YYYY", date_time → "DD.MM.YYYY HH:mm", file → the name from
// [RequestsService.Upload], multiple_field → a JSON string of a 2-D array of
// cell values (rows of columns, dictionary cells replaced by their id), other
// values as text ("" when empty).
type RequestFieldValue struct {
	FieldID string `json:"field_id"`
	// FieldType is required by Send; FormUpdate accepts it empty.
	FieldType RequestFieldType `json:"field_type,omitzero"`
	Value     string           `json:"value"`
}

// RequestFormUpdate lists the fields to show and hide after a value change.
type RequestFormUpdate struct {
	Show []FlexID `json:"show"`
	Hide []FlexID `json:"hide"`
}

// RequestSummary is a submitted legacy request in the user's list.
type RequestSummary struct {
	ID     int64  `json:"id"`
	Name   string `json:"name"`
	Notice string `json:"notice"`
	// Status is one of the RequestStatus* constants.
	Status     RequestStatus `json:"status"`
	StatusName string        `json:"status_name"`
	CreatedAt  time.Time     `json:"created_at"`
	UpdatedAt  time.Time     `json:"updated_at"`
}

// Legacy request status tags of [RequestDetail]; any other tag means in progress.
const (
	RequestStatusTagProcessed = "processed"
	RequestStatusTagRejected  = "rejected"
)

// RequestDetail is a submitted legacy request with its answers and approvals.
type RequestDetail struct {
	ID            int64  `json:"id"`
	RequestName   string `json:"request_name"`
	RequestStatus string `json:"request_status"`
	RequestState  string `json:"request_state"`
	// RequestStatusTag is RequestStatusTagProcessed, RequestStatusTagRejected or another value (in progress).
	RequestStatusTag  string                `json:"request_status_tag"`
	RequestCreateDate time.Time             `json:"request_create_date"`
	Approvals         []RequestApproval     `json:"agreement_list"`
	EnteredFields     []RequestEnteredField `json:"entered_fields"`
}

// RequestApproval is one step of the approval chain of a legacy request.
type RequestApproval struct {
	FIO         string `json:"fio"`
	ContextName string `json:"context_name"`
	// Decision is "Согласовано", "Отклонено" or another text while pending.
	Decision     string `json:"decision"`
	DecisionDate string `json:"decision_date"`
}

// RequestEnteredField is an answer of a submitted legacy request.
type RequestEnteredField struct {
	Name string           `json:"name"`
	Type RequestFieldType `json:"type"`
	// Data is the display value; for file fields it is the id for [RequestsService.AttachedFile].
	Data     string `json:"data"`
	FileName string `json:"file_name"`
	// MultipleData is set for multiple_field answers.
	MultipleData *RequestMultipleData `json:"multiple_data"`
}

// RequestMultipleData is a table answer.
type RequestMultipleData struct {
	Headers []string   `json:"headers"`
	Values  [][]string `json:"values"`
}

// RequestValidationError is returned by [RequestsService.Send] when the server
// rejects individual fields (result.error_list). It unwraps to *Error.
type RequestValidationError struct {
	Err    *Error
	Fields []RequestFieldError
}

func (e *RequestValidationError) Error() string {
	return fmt.Sprintf("%v (%d invalid fields)", e.Err, len(e.Fields))
}

func (e *RequestValidationError) Unwrap() error { return e.Err }

// RequestFieldError is the validation error of one field.
type RequestFieldError struct {
	// ID is the field_id of the field.
	ID   FlexID                `json:"id"`
	Info RequestFieldErrorInfo `json:"error_info"`
}

// RequestFieldErrorInfo is the message of a field, or per-cell messages of a multiple_field.
type RequestFieldErrorInfo struct {
	Text  string                  `json:"text"`
	Cells []RequestFieldErrorCell `json:"obj_list"`
}

// RequestFieldErrorCell is a validation error of one cell of a multiple_field.
type RequestFieldErrorCell struct {
	// FieldNumber is the 1-based column.
	FieldNumber int `json:"field_number"`
	// PartNumber is the 1-based row (block).
	PartNumber int    `json:"part_number"`
	ErrorText  string `json:"error_text"`
}

// Catalog returns the legacy request templates grouped by category.
// GET /api/requests/all
func (s *RequestsService) Catalog(ctx context.Context) ([]RequestCategory, error) {
	return call[[]RequestCategory](ctx, s.c, get("api/requests/all", nil))
}

// Template returns the form of a legacy request template. templateID is RequestTemplate.ID.
// GET /api/requests/{requestId}
func (s *RequestsService) Template(ctx context.Context, templateID int64) (*RequestForm, error) {
	return call[*RequestForm](ctx, s.c, get("api/requests/"+id(templateID), nil))
}

// Dictionary searches the values of a legacy dictionary for a field.
// dictID is RequestField.DictionaryID; search may be empty for the initial list;
// dep is the value of the field named by DependentFieldID, or "".
// GET /api/requests/dict/{dictId}
func (s *RequestsService) Dictionary(ctx context.Context, dictID string, fieldID int64, search, dep string) ([]RequestDictItem, error) {
	qv := query{"q": {search}}.set("field", fieldID).set("dep", dep)
	return call[[]RequestDictItem](ctx, s.c, get("api/requests/dict/"+id(dictID), qv))
}

// FormUpdate recomputes which fields are visible after changedField changed.
// POST /api/requests/form_update
func (s *RequestsService) FormUpdate(ctx context.Context, changedField int64, values []RequestFieldValue) (*RequestFormUpdate, error) {
	body := struct {
		ChangedField  int64               `json:"changed_field"`
		CurrentValues []RequestFieldValue `json:"current_values"`
	}{changedField, values}
	return call[*RequestFormUpdate](ctx, s.c, post("api/requests/form_update", body))
}

// Send submits a filled legacy request and returns the id of the created request.
// Field-level rejections come back as *RequestValidationError.
// POST /api/requests/send
func (s *RequestsService) Send(ctx context.Context, templateID int64, values []RequestFieldValue) (int64, error) {
	body := struct {
		RequestID int64               `json:"request_id"`
		Values    []RequestFieldValue `json:"values"`
	}{templateID, values}
	r := post("api/requests/send", body)
	resp, err := s.c.rest.Send(ctx, r) //nolint:bodyclose // closed by rest.ReadAll
	if err != nil {
		return 0, err
	}
	raw, err := rest.ReadAll(resp, maxBody)
	if err != nil {
		return 0, fmt.Errorf("myitmo: read %s %s: %w", r.Method, r.Path, err)
	}
	var env struct {
		ErrorCode    int       `json:"error_code"`
		ErrorMessage string    `json:"error_message"`
		Message      string    `json:"message"`
		Result       jsonValue `json:"result"`
	}
	parsed := jsonx.Unmarshal(raw, &env) == nil
	if resp.StatusCode/100 != 2 || env.ErrorCode != 0 {
		e := &Error{Method: r.Method, Path: r.Path, StatusCode: resp.StatusCode, Code: env.ErrorCode, Message: env.ErrorMessage}
		if e.Message == "" {
			e.Message = env.Message
		}
		var res struct {
			ErrorList []RequestFieldError `json:"error_list"`
		}
		if parsed && decodeInto(env.Result, &res) == nil && len(res.ErrorList) > 0 {
			return 0, &RequestValidationError{Err: e, Fields: res.ErrorList}
		}
		e.Details = details(env.Result)
		return 0, e
	}
	if !parsed {
		return 0, &DecodeError{Method: r.Method, Path: r.Path, Err: errors.New("response is not a JSON envelope")}
	}
	var res struct {
		ReqID *int64 `json:"reqId"`
	}
	if err := decodeInto(env.Result, &res); err != nil {
		return 0, &DecodeError{Method: r.Method, Path: r.Path, Err: err}
	}
	if res.ReqID == nil {
		return 0, &Error{Method: r.Method, Path: r.Path, StatusCode: resp.StatusCode, Message: "request was not created"}
	}
	return *res.ReqID, nil
}

// Upload stores a file for a "file" field and returns the name to put into its value.
// The part name is always "file".
// POST /api/requests/upload
func (s *RequestsService) Upload(ctx context.Context, file Upload) (string, error) {
	file.Field = "file"
	res, err := call[struct {
		Name string `json:"name"`
	}](ctx, s.c, multipart(http.MethodPost, "api/requests/upload", nil, file))
	return res.Name, err
}

// TemplateFile downloads a blank attached to a template. fileID is RequestTemplateFile.FileID.
// GET /api/requests/file/{fileId}
func (s *RequestsService) TemplateFile(ctx context.Context, fileID string) (*File, error) {
	return download(ctx, s.c, get("api/requests/file/"+id(fileID), nil))
}

// List returns the user's submitted legacy requests.
// GET /api/requests/my
func (s *RequestsService) List(ctx context.Context) ([]RequestSummary, error) {
	return call[[]RequestSummary](ctx, s.c, get("api/requests/my", nil))
}

// Get returns a submitted legacy request. requestID is RequestSummary.ID.
// GET /api/requests/my/{requestId}
func (s *RequestsService) Get(ctx context.Context, requestID int64) (*RequestDetail, error) {
	return call[*RequestDetail](ctx, s.c, get("api/requests/my/"+id(requestID), nil))
}

// Delete withdraws a submitted legacy request that is still in progress.
// DELETE /api/requests/my/{requestId}
func (s *RequestsService) Delete(ctx context.Context, requestID int64) error {
	return requestsExecLoose(ctx, s.c, del("api/requests/my/"+id(requestID), nil))
}

// Download returns the printable PDF of a submitted legacy request.
// GET /api/requests/my/{requestId}/download
func (s *RequestsService) Download(ctx context.Context, requestID int64) (*File, error) {
	return download(ctx, s.c, get("api/requests/my/"+id(requestID)+"/download", nil))
}

// AttachedFile downloads a file attached to a submitted request. fileID is
// RequestEnteredField.Data of a file field.
// GET /api/requests/files/my/{fileId}
func (s *RequestsService) AttachedFile(ctx context.Context, fileID string) (*File, error) {
	return download(ctx, s.c, get("api/requests/files/my/"+id(fileID), nil))
}
