package myitmo

import (
	"context"
	"encoding/base64"
	"errors"
	"strings"
	"time"
)

// RequestsV2Service is the process-based requests platform (/api/requests/v2).
//
// Forms are SurveyJS models and processes are Flowable BPMN definitions. Fields
// whose name ends in JSON (SchemaJSON, PayloadJSON, ConfigJSON, ...) hold
// stringified JSON documents exactly as the server stores them; encode and
// decode them yourself. Errors carry result.details in [Error.Details].
//
// Staff methods are grouped by prefix: Categories, AccessList*, Dictionar*,
// PlatformRole*, Process*, Form*, Observed*, Tasks and ProcessInstance*.
type RequestsV2Service struct{ c *Client }

// RequestV2Status is the status of a v2 request.
type RequestV2Status string

// Request v2 statuses. PROCESSED and REJECTED are final.
const (
	RequestV2Draft                 RequestV2Status = "DRAFT"
	RequestV2Submitted             RequestV2Status = "SUBMITTED"
	RequestV2AcceptedForProcessing RequestV2Status = "ACCEPTED_FOR_PROCESSING"
	RequestV2InReview              RequestV2Status = "IN_REVIEW"
	RequestV2InProgress            RequestV2Status = "IN_PROGRESS"
	RequestV2Approved              RequestV2Status = "APPROVED"
	RequestV2Processed             RequestV2Status = "PROCESSED"
	RequestV2Rejected              RequestV2Status = "REJECTED"
	RequestV2ReturnedForRework     RequestV2Status = "RETURNED_FOR_REWORK"
	RequestV2Cancelled             RequestV2Status = "CANCELLED"
)

// Finished reports whether the status is final.
func (s RequestV2Status) Finished() bool { return s == RequestV2Processed || s == RequestV2Rejected }

// RequestV2CountPage is a page of items with the total number of matches.
type RequestV2CountPage[T any] struct {
	Items      []T `json:"items"`
	TotalCount int `json:"totalCount"`
}

// RequestV2Page is a numbered page of items (0-based Page).
type RequestV2Page[T any] struct {
	Items         []T `json:"items"`
	TotalElements int `json:"totalElements"`
	TotalPages    int `json:"totalPages"`
	Page          int `json:"page"`
	Size          int `json:"size"`
}

// RequestV2JSONText is a stringified JSON document. The server usually sends
// it as a JSON string; an inline object or array is kept as its JSON text.
type RequestV2JSONText string

// UnmarshalJSON accepts a string, an inline JSON value or null.
func (v *RequestV2JSONText) UnmarshalJSON(b []byte) error {
	s, err := requestsFlexText(b)
	*v = RequestV2JSONText(s)
	return err
}

// RequestV2FormCategory is a catalog category with the forms the user may submit.
type RequestV2FormCategory struct {
	// ID is empty for the "no category" group.
	ID   FlexID `json:"id"`
	Code string `json:"code"`
	// Name is empty for the "no category" group.
	Name string `json:"name"`
	// Color is a CSS color name such as "secondary-blue".
	Color     string          `json:"color"`
	Icon      string          `json:"icon"`
	SortOrder *int            `json:"sortOrder"`
	Forms     []RequestV2Form `json:"forms"`
}

// RequestV2Form is a submission form. The catalog may omit SchemaJSON.
type RequestV2Form struct {
	ID FlexID `json:"id"`
	// Alias is the form key for [RequestsV2Service.Form] and [RequestsV2Service.Submit].
	Alias       string `json:"alias"`
	Name        string `json:"name"`
	Description string `json:"description"`
	// SchemaJSON is the SurveyJS model as JSON text. Questions may carry the
	// custom properties dictionaryCode, dictionaryDependsOn, prefillCode and
	// prefillDependsOn that drive ResolveDictionary and ResolvePrefill.
	SchemaJSON RequestV2JSONText `json:"schemaJson"`
	// HTMLNote is an HTML note shown above the form.
	HTMLNote string `json:"htmlNote"`
}

// RequestV2Summary is a v2 request in a list.
type RequestV2Summary struct {
	ID                   int64           `json:"id"`
	FormCode             string          `json:"formCode"`
	FormName             string          `json:"formName"`
	Status               RequestV2Status `json:"status"`
	SubStatus            string          `json:"subStatus"`
	CreatedAt            time.Time       `json:"createdAt"`
	UpdatedAt            *time.Time      `json:"updatedAt"`
	Initiator            string          `json:"initiator"`
	InitiatorDisplayName string          `json:"initiatorDisplayName"`
}

// RequestV2Request is a v2 request with its answers and process result.
type RequestV2Request struct {
	ID                   int64           `json:"id"`
	FormCode             string          `json:"formCode"`
	FormName             string          `json:"formName"`
	Status               RequestV2Status `json:"status"`
	SubStatus            string          `json:"subStatus"`
	CreatedAt            time.Time       `json:"createdAt"`
	UpdatedAt            *time.Time      `json:"updatedAt"`
	Initiator            string          `json:"initiator"`
	InitiatorDisplayName string          `json:"initiatorDisplayName"`
	ReviewComment        string          `json:"reviewComment"`
	// PayloadJSON is the submitted SurveyJS answers as JSON text.
	PayloadJSON    string                   `json:"payloadJson"`
	RenderedFields []RequestV2RenderedField `json:"renderedFields"`
	Result         *RequestV2Result         `json:"result"`
}

// RequestV2RenderedField is a submitted answer formatted for display.
type RequestV2RenderedField struct {
	Name         string `json:"name"`
	Title        string `json:"title"`
	DisplayValue string `json:"displayValue"`
}

// RequestV2Result is the outcome the process attached to a request.
type RequestV2Result struct {
	// Tone is "success", "warning", "error" or "info".
	Tone       string                   `json:"tone"`
	Title      string                   `json:"title"`
	Message    string                   `json:"message"`
	Parameters []RequestV2ResultParam   `json:"parameters"`
	Files      []RequestV2ResultFileRef `json:"files"`
}

// RequestV2ResultParam is a label and value of a request result.
type RequestV2ResultParam struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

// RequestV2ResultFileRef is a file produced by the process; fetch it with [RequestsV2Service.ResultFile].
type RequestV2ResultFileRef struct {
	VariableName string `json:"variableName"`
	Label        string `json:"label"`
	Filename     string `json:"filename"`
}

// RequestV2FileContent is a result file sent inline as base64.
type RequestV2FileContent struct {
	Base64Content string `json:"base64Content"`
	ContentType   string `json:"contentType"`
	Filename      string `json:"filename"`
}

// Bytes decodes the file content.
func (f *RequestV2FileContent) Bytes() ([]byte, error) {
	if f == nil || f.Base64Content == "" {
		return nil, errors.New("myitmo: empty file content")
	}
	s := f.Base64Content
	if i := strings.Index(s, ";base64,"); i >= 0 && strings.HasPrefix(s, "data:") {
		s = s[i+len(";base64,"):]
	}
	return base64.StdEncoding.DecodeString(s)
}

// RequestV2Task is a workflow (Flowable) user task. Staff task lists and the
// applicant's own task share this shape; some fields are set by only one of them.
type RequestV2Task struct {
	// ID is the task id for [RequestsV2Service.CompleteTask].
	ID        string `json:"id"`
	RequestID int64  `json:"requestId"`
	Name      string `json:"name"`
	// TaskDefinitionKey is the BPMN element id (staff lists).
	TaskDefinitionKey string `json:"taskDefinitionKey,omitzero"`
	// Assignee is nil for unassigned tasks (staff lists).
	Assignee *string               `json:"assignee,omitzero"`
	Actions  []RequestV2TaskAction `json:"actions"`
	// TaskFormSchemaJSON is the SurveyJS model of an extra step form, as JSON text.
	TaskFormSchemaJSON string `json:"taskFormSchemaJson"`
	// RequestPayloadJSON is the request data to prefill the step form with, as JSON text.
	RequestPayloadJSON string `json:"requestPayloadJson,omitzero"`
	// SignatureTask is truthy (true or an object) for an e-signature step; such
	// tasks are completed through the sign service using SignatureTasks.
	SignatureTask  RawJSON                  `json:"signatureTask,omitzero"`
	SignatureTasks []RequestV2SignatureTask `json:"signatureTasks,omitzero"`
}

// RequestV2TaskAction is a button of a task.
type RequestV2TaskAction struct {
	// Code is sent as the outcome of [RequestsV2Service.CompleteTask].
	Code  string `json:"code"`
	Label string `json:"label"`
	// Kind is "danger", "ghost" or another value (primary); it only styles the button.
	Kind string `json:"kind"`
	// RequiresComment means the reviewComment variable must be set.
	RequiresComment bool `json:"requiresComment"`
}

// RequestV2SignatureTask points at a task of the sign service (/api/sign).
type RequestV2SignatureTask struct {
	TaskID      FlexID `json:"taskId"`
	SignatureID FlexID `json:"signatureId"`
}

// RequestV2TaskVariables are the process variables sent with a task outcome.
type RequestV2TaskVariables struct {
	ReviewComment string `json:"reviewComment,omitzero"`
	// RequestDataPatch is the data of the step form: a JSON object of field → value.
	RequestDataPatch RawJSON `json:"requestDataPatch,omitzero"`
}

// RequestV2TaskCounts are the badge counters of the requests tabs.
type RequestV2TaskCounts struct {
	// OwnRequestTasks is the number of open tasks on the user's own requests.
	OwnRequestTasks int `json:"ownRequestTasks"`
	// AssignedRequestTasks is the number of tasks assigned to the user.
	AssignedRequestTasks int `json:"assignedRequestTasks"`
}

// RequestV2DictionaryContext is the state of a SurveyJS question whose choices are resolved.
type RequestV2DictionaryContext struct {
	// Field and QuestionName are both the question name.
	Field        string `json:"field"`
	QuestionName string `json:"questionName"`
	// Filter and Query are both the search text.
	Filter string `json:"filter"`
	Query  string `json:"query"`
	Skip   int    `json:"skip"`
	// Take is the page size, typically 25.
	Take int `json:"take"`
	// Data is a JSON object of all current answers (dictionary answers reduced to their values).
	Data              RawJSON `json:"data,omitzero"`
	DependsOnQuestion string  `json:"dependsOnQuestion,omitzero"`
	// DependsOnValue is the comma-joined value of the dependency question.
	DependsOnValue string `json:"dependsOnValue,omitzero"`
	// Values asks for the labels of these values; Value repeats a single one.
	Values []RawJSON `json:"values,omitzero"`
	Value  RawJSON   `json:"value,omitzero"`
}

// RequestV2DictionaryItem is a choice of a dictionary question.
type RequestV2DictionaryItem struct {
	// Value is the stored value, usually a string or a number.
	Value RawJSON `json:"value,omitzero"`
	Text  string  `json:"text"`
}

// RequestV2PrefillContext is the state of a SurveyJS question whose value is prefilled.
type RequestV2PrefillContext struct {
	Field        string `json:"field"`
	QuestionName string `json:"questionName"`
	// Data is a JSON object of all current answers.
	Data              RawJSON `json:"data,omitzero"`
	CurrentValue      RawJSON `json:"currentValue,omitzero"`
	DependsOnQuestion string  `json:"dependsOnQuestion,omitzero"`
	DependsOnValue    string  `json:"dependsOnValue,omitzero"`
	// Dep repeats DependsOnValue under the default prefillDependsOnParam name.
	Dep string `json:"dep,omitzero"`
}

// RequestV2Prefill is a computed prefill value.
type RequestV2Prefill struct {
	// Value is the value to set; its type depends on the question.
	Value RawJSON `json:"value,omitzero"`
}

// Forms returns the catalog of forms the user may submit, grouped by category.
// GET /api/requests/v2/forms
func (s *RequestsV2Service) Forms(ctx context.Context) ([]RequestV2FormCategory, error) {
	return call[[]RequestV2FormCategory](ctx, s.c, get("api/requests/v2/forms", nil))
}

// Form returns a form with its SurveyJS schema.
// GET /api/requests/v2/forms/{alias}
func (s *RequestsV2Service) Form(ctx context.Context, alias string) (*RequestV2Form, error) {
	return call[*RequestV2Form](ctx, s.c, get("api/requests/v2/forms/"+id(alias), nil))
}

// Submit creates a request from a form. payloadJSON is the SurveyJS answers as
// JSON text, with dictionary answers reduced to their values.
// POST /api/requests/v2/requests
func (s *RequestsV2Service) Submit(ctx context.Context, alias, payloadJSON string) (*RequestV2Request, error) {
	body := struct {
		Alias       string `json:"alias"`
		PayloadJSON string `json:"payloadJson"`
	}{alias, payloadJSON}
	return call[*RequestV2Request](ctx, s.c, post("api/requests/v2/requests", body))
}

// Mine returns the requests created by the user.
// GET /api/requests/v2/requests/mine
func (s *RequestsV2Service) Mine(ctx context.Context) ([]RequestV2Summary, error) {
	return call[[]RequestV2Summary](ctx, s.c, get("api/requests/v2/requests/mine", nil))
}

// Get returns a request the user created, observes or has a task on.
// GET /api/requests/v2/requests/{id}
func (s *RequestsV2Service) Get(ctx context.Context, requestID int64) (*RequestV2Request, error) {
	return call[*RequestV2Request](ctx, s.c, get("api/requests/v2/requests/"+id(requestID), nil))
}

// ResultFile returns a file produced by the request's process; decode it with [RequestV2FileContent.Bytes].
// GET /api/requests/v2/requests/{id}/files/{variableName}
func (s *RequestsV2Service) ResultFile(ctx context.Context, requestID int64, variableName string) (*RequestV2FileContent, error) {
	return call[*RequestV2FileContent](ctx, s.c, get("api/requests/v2/requests/"+id(requestID)+"/files/"+id(variableName), nil))
}

// MyTaskRequestIDs returns the ids of requests on which the user has an open task.
// GET /api/requests/v2/tasks/mine/request-ids
func (s *RequestsV2Service) MyTaskRequestIDs(ctx context.Context) ([]int64, error) {
	return call[[]int64](ctx, s.c, get("api/requests/v2/tasks/mine/request-ids", nil))
}

// MyTask returns the user's open task on a request, or nil when there is none.
// GET /api/requests/v2/tasks/mine/by-request/{requestId}
func (s *RequestsV2Service) MyTask(ctx context.Context, requestID int64) (*RequestV2Task, error) {
	return call[*RequestV2Task](ctx, s.c, get("api/requests/v2/tasks/mine/by-request/"+id(requestID), nil))
}

// AssignedRequests returns requests of other users that wait for the user's action ("my tasks").
// GET /api/requests/v2/requests/my-tasks
func (s *RequestsV2Service) AssignedRequests(ctx context.Context) ([]RequestV2Summary, error) {
	return call[[]RequestV2Summary](ctx, s.c, get("api/requests/v2/requests/my-tasks", nil))
}

// TaskCounts returns the badge counters of the requests tabs.
// GET /api/requests/v2/requests/task-counts
func (s *RequestsV2Service) TaskCounts(ctx context.Context) (*RequestV2TaskCounts, error) {
	return call[*RequestV2TaskCounts](ctx, s.c, get("api/requests/v2/requests/task-counts", nil))
}

// CompleteTask completes a task with an outcome (RequestV2TaskAction.Code).
// Staff approve or reject with it; applicants rework or confirm.
// POST /api/requests/v2/tasks/{taskId}/complete
func (s *RequestsV2Service) CompleteTask(ctx context.Context, taskID, outcome string, vars RequestV2TaskVariables) error {
	body := struct {
		Outcome   string                 `json:"outcome"`
		Variables RequestV2TaskVariables `json:"variables"`
	}{outcome, vars}
	return exec(ctx, s.c, post("api/requests/v2/tasks/"+id(taskID)+"/complete", body))
}

// ResolveDictionary returns choices of a dictionary question. code is the
// question's dictionaryCode. A bare array result is returned as a page
// with TotalCount equal to its length.
// POST /api/requests/v2/dictionaries/resolve
func (s *RequestsV2Service) ResolveDictionary(ctx context.Context, code string, in RequestV2DictionaryContext) (*RequestV2CountPage[RequestV2DictionaryItem], error) {
	r := withQuery(post("api/requests/v2/dictionaries/resolve", in), q().set("code", code))
	raw, err := call[RawJSON](ctx, s.c, r)
	if err != nil {
		return nil, err
	}
	page := &RequestV2CountPage[RequestV2DictionaryItem]{}
	if len(raw) > 0 && raw.Kind() == '[' {
		err = decodeInto(raw, &page.Items)
		page.TotalCount = len(page.Items)
	} else {
		err = decodeInto(raw, page)
	}
	if err != nil {
		return nil, &DecodeError{Method: r.Method, Path: r.Path, Err: err}
	}
	return page, nil
}

// ResolvePrefill computes the prefill value of a question. code is the question's prefillCode.
// POST /api/requests/v2/prefills/resolve
func (s *RequestsV2Service) ResolvePrefill(ctx context.Context, code string, in RequestV2PrefillContext) (*RequestV2Prefill, error) {
	return call[*RequestV2Prefill](ctx, s.c, withQuery(post("api/requests/v2/prefills/resolve", in), q().set("code", code)))
}

// requestsTruthy reports whether v is truthy: anything but empty, null, false, 0 and "".
func requestsTruthy(v RawJSON) bool {
	switch strings.TrimSpace(string(v)) {
	case "", "null", "false", "0", `""`:
		return false
	}
	return true
}

// requestsProbe calls an access-check route and reports whether its result is truthy.
func (s *RequestsV2Service) requestsProbe(ctx context.Context, path string) (bool, error) {
	v, err := call[RawJSON](ctx, s.c, get(path, nil))
	return err == nil && requestsTruthy(v), err
}

// ---- Categories ----

// RequestV2Category is a catalog category (staff view).
type RequestV2Category struct {
	ID int64 `json:"id"`
	// Code is the unique alias (up to 100 characters) and the upsert key.
	Code string `json:"code"`
	Name string `json:"name"`
	// Icon is an icon name, "question" by default.
	Icon string `json:"icon"`
	// Color is primary-blue, primary-pink, secondary-blue, secondary-green,
	// secondary-orange, secondary-purple, secondary-red or secondary-yellow.
	Color     string `json:"color"`
	SortOrder int    `json:"sortOrder"`
}

// RequestV2CategoryInput creates or updates a category keyed by Code.
type RequestV2CategoryInput struct {
	Code      string `json:"code"`
	Name      string `json:"name"`
	Icon      string `json:"icon"`
	Color     string `json:"color"`
	SortOrder int    `json:"sortOrder"`
}

// Categories returns the catalog categories. Staff only.
// GET /api/requests/v2/categories
func (s *RequestsV2Service) Categories(ctx context.Context) ([]RequestV2Category, error) {
	return call[[]RequestV2Category](ctx, s.c, get("api/requests/v2/categories", nil))
}

// SaveCategory creates or updates a category by code. Staff only.
// PUT /api/requests/v2/categories
func (s *RequestsV2Service) SaveCategory(ctx context.Context, in RequestV2CategoryInput) (*RequestV2Category, error) {
	return call[*RequestV2Category](ctx, s.c, put("api/requests/v2/categories", in))
}

// DeleteCategory deletes a category; it fails while forms use it. Staff only.
// DELETE /api/requests/v2/categories/{id}
func (s *RequestsV2Service) DeleteCategory(ctx context.Context, categoryID int64) error {
	return exec(ctx, s.c, del("api/requests/v2/categories/"+id(categoryID), nil))
}

// ---- Access lists ----

// RequestV2AccessListType is the kind of an access list.
type RequestV2AccessListType string

// Access list types.
const (
	// RequestV2AccessListStatic holds an explicit user list and ConfigJSON "{}".
	RequestV2AccessListStatic RequestV2AccessListType = "STATIC"
	// RequestV2AccessListDynamic fetches users from an HTTP source described by ConfigJSON.
	RequestV2AccessListDynamic RequestV2AccessListType = "DYNAMIC"
)

// RequestV2AccessList is an access list (a set of users allowed to submit a form).
type RequestV2AccessList struct {
	Code string                  `json:"code"`
	Name string                  `json:"name"`
	Type RequestV2AccessListType `json:"type"`
	// ConfigJSON is a [RequestV2AccessListConfig] as JSON text ("{}" for STATIC lists).
	ConfigJSON        string     `json:"configJson"`
	UserCount         int        `json:"userCount"`
	LastAttemptAt     *time.Time `json:"lastAttemptAt"`
	LastUpdateStatus  string     `json:"lastUpdateStatus"`
	LastUpdatedBy     string     `json:"lastUpdatedBy"`
	LastUpdateMessage string     `json:"lastUpdateMessage"`
}

// RequestV2AccessListInput creates, updates or tests an access list keyed by Code.
type RequestV2AccessListInput struct {
	Code string                  `json:"code"`
	Name string                  `json:"name"`
	Type RequestV2AccessListType `json:"type"`
	// ConfigJSON is a [RequestV2AccessListConfig] as JSON text; "{}" for STATIC lists.
	ConfigJSON string `json:"configJson"`
}

// RequestV2AccessListConfig is the HTTP source of a DYNAMIC access list, stored as ConfigJSON.
type RequestV2AccessListConfig struct {
	URL string `json:"url"`
	// Method defaults to "GET".
	Method      string        `json:"method,omitzero"`
	QueryParams RawJSON       `json:"queryParams,omitzero"`
	Body        RawJSON       `json:"body,omitzero"`
	Auth        RequestV2Auth `json:"auth,omitzero"`
	// ResponseArrayPath is the path to the array of user ids in the response.
	ResponseArrayPath string `json:"responseArrayPath,omitzero"`
	// RefreshFrequency is "DAILY", "HOURLY" or "CRON".
	RefreshFrequency string `json:"refreshFrequency,omitzero"`
	// Cron is used with RefreshFrequency "CRON".
	Cron string `json:"cron,omitzero"`
}

// RequestV2Auth is the authentication of an outgoing HTTP call of an access
// list or a dictionary. Type is NONE, BASIC, API_KEY, BEARER (access lists) or
// BEARER_TOKEN (dictionaries), SERVICE_TOKEN (access lists) or SERVICE_REQUESTS
// (dictionaries). Placement is HEADER or QUERY for API_KEY.
type RequestV2Auth struct {
	Type      string `json:"type,omitzero"`
	Placement string `json:"placement,omitzero"`
	Username  string `json:"username,omitzero"`
	Password  string `json:"password,omitzero"`
	Token     string `json:"token,omitzero"`
	// Name and Value are the API key header or query parameter.
	Name  string `json:"name,omitzero"`
	Value string `json:"value,omitzero"`
}

// RequestV2AccessListTest is the result of a dry run of an access list source.
type RequestV2AccessListTest struct {
	// StatusCode is the upstream HTTP status.
	StatusCode int      `json:"statusCode"`
	UserCount  int      `json:"userCount"`
	Sample     []string `json:"sample"`
}

// RequestV2AccessListOption is an access list offered in the form editor.
type RequestV2AccessListOption struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

// CanManageAccessLists reports whether the user may manage access lists. Staff only.
// GET /api/requests/v2/access-definitions/admin-access
func (s *RequestsV2Service) CanManageAccessLists(ctx context.Context) (bool, error) {
	return s.requestsProbe(ctx, "api/requests/v2/access-definitions/admin-access")
}

// AccessLists returns all access lists with their configuration. Staff only.
// GET /api/requests/v2/access-definitions
func (s *RequestsV2Service) AccessLists(ctx context.Context) ([]RequestV2AccessList, error) {
	return call[[]RequestV2AccessList](ctx, s.c, get("api/requests/v2/access-definitions", nil))
}

// SaveAccessList creates or updates an access list by code. Staff only.
// PUT /api/requests/v2/access-definitions
func (s *RequestsV2Service) SaveAccessList(ctx context.Context, in RequestV2AccessListInput) error {
	return exec(ctx, s.c, put("api/requests/v2/access-definitions", in))
}

// AccessListUsers returns the user ids (SSO sub) of a STATIC access list. Staff only.
// GET /api/requests/v2/access-definitions/{code}/users
func (s *RequestsV2Service) AccessListUsers(ctx context.Context, code string) ([]string, error) {
	return call[[]string](ctx, s.c, get("api/requests/v2/access-definitions/"+id(code)+"/users", nil))
}

// SetAccessListUsers replaces the user ids (SSO sub) of a STATIC access list. Staff only.
// PUT /api/requests/v2/access-definitions/{code}/users
func (s *RequestsV2Service) SetAccessListUsers(ctx context.Context, code string, userIDs []string) error {
	return exec(ctx, s.c, put("api/requests/v2/access-definitions/"+id(code)+"/users", userIDs))
}

// PushAccessListExternalUsers replaces the members of an access list from an
// external system. The token needs the scope modify:access-lists:all.
// Staff only.
// PUT /api/requests/v2/access-definitions/{code}/external-users
func (s *RequestsV2Service) PushAccessListExternalUsers(ctx context.Context, code string, userIDs []string) error {
	return exec(ctx, s.c, put("api/requests/v2/access-definitions/"+id(code)+"/external-users", userIDs))
}

// TestAccessList dry-runs an access list configuration without saving it. Staff only.
// POST /api/requests/v2/access-definitions/test
func (s *RequestsV2Service) TestAccessList(ctx context.Context, in RequestV2AccessListInput) (*RequestV2AccessListTest, error) {
	return call[*RequestV2AccessListTest](ctx, s.c, post("api/requests/v2/access-definitions/test", in))
}

// RefreshAccessList reloads a DYNAMIC access list from its source now. Staff only.
// POST /api/requests/v2/access-definitions/{code}/refresh
func (s *RequestsV2Service) RefreshAccessList(ctx context.Context, code string) error {
	return exec(ctx, s.c, post("api/requests/v2/access-definitions/"+id(code)+"/refresh", nil))
}

// DeleteAccessList deletes an access list. Staff only.
// DELETE /api/requests/v2/access-definitions/{code}
func (s *RequestsV2Service) DeleteAccessList(ctx context.Context, code string) error {
	return exec(ctx, s.c, del("api/requests/v2/access-definitions/"+id(code), nil))
}

// AccessListOptions returns the access lists offered in the form editor. Staff only.
// GET /api/requests/v2/access-definitions/options
func (s *RequestsV2Service) AccessListOptions(ctx context.Context) ([]RequestV2AccessListOption, error) {
	return call[[]RequestV2AccessListOption](ctx, s.c, get("api/requests/v2/access-definitions/options", nil))
}

// ---- Dictionaries, prefills, context enrichments ----

// RequestV2DictionaryRef is a dictionary that form questions can bind to via dictionaryCode.
type RequestV2DictionaryRef struct {
	Code     string `json:"code"`
	Name     string `json:"name"`
	Provider string `json:"provider"`
}

// RequestV2Dictionary is a dictionary with its HTTP configuration (admin view).
type RequestV2Dictionary struct {
	Code string `json:"code"`
	Name string `json:"name"`
	// ConfigJSON is a [RequestV2DictionaryConfig] as JSON text.
	ConfigJSON string `json:"configJson"`
}

// RequestV2DictionaryInput creates or updates a dictionary keyed by Code.
type RequestV2DictionaryInput struct {
	Code string `json:"code"`
	Name string `json:"name"`
	// ConfigJSON is a [RequestV2DictionaryConfig] as JSON text.
	ConfigJSON string `json:"configJson"`
}

// RequestV2DictionaryConfig is the HTTP source of a dictionary, stored as
// ConfigJSON. URL, query and header values may contain ${payload.x},
// ${user.*} and ${process.*} placeholders.
type RequestV2DictionaryConfig struct {
	URL string `json:"url"`
	// Method defaults to "GET"; POST, PUT and PATCH send Body.
	Method      string            `json:"method,omitzero"`
	QueryParams map[string]string `json:"queryParams,omitzero"`
	Headers     map[string]string `json:"headers,omitzero"`
	Body        RawJSON           `json:"body,omitzero"`
	Auth        RequestV2Auth     `json:"auth,omitzero"`
	// ItemsPath defaults to "items", TotalCountPath to "totalCount",
	// ValuePath to "id" and TextPath to "text".
	ItemsPath      string `json:"itemsPath,omitzero"`
	TotalCountPath string `json:"totalCountPath,omitzero"`
	ValuePath      string `json:"valuePath,omitzero"`
	TextPath       string `json:"textPath,omitzero"`
	// TimeoutSeconds defaults to 30.
	TimeoutSeconds int `json:"timeoutSeconds,omitzero"`
}

// RequestV2DictionaryTest is the result of a dry run of a dictionary configuration.
type RequestV2DictionaryTest struct {
	StatusCode int  `json:"statusCode"`
	TotalCount *int `json:"totalCount"`
	// Items are the extracted items as returned by the upstream.
	Items []RawJSON `json:"items"`
}

// RequestV2PrefillRef is a prefill source that form questions can bind to via prefillCode.
// The key names are not confirmed.
type RequestV2PrefillRef struct {
	Code     string `json:"code"`
	Name     string `json:"name"`
	Title    string `json:"title"`
	Provider string `json:"provider"`
}

// RequestV2ContextEnrichment is an integration a BPMN service task can use.
type RequestV2ContextEnrichment struct {
	// Code is stored as definitionCode in the task configuration.
	Code     string `json:"code"`
	Provider string `json:"provider"`
}

// Dictionaries returns the dictionaries available to form questions. Staff only.
// GET /api/requests/v2/dictionaries
func (s *RequestsV2Service) Dictionaries(ctx context.Context) ([]RequestV2DictionaryRef, error) {
	return call[[]RequestV2DictionaryRef](ctx, s.c, get("api/requests/v2/dictionaries", nil))
}

// CanManageDictionaries reports whether the user may manage dictionaries. Staff only.
// GET /api/requests/v2/dictionaries/admin-access
func (s *RequestsV2Service) CanManageDictionaries(ctx context.Context) (bool, error) {
	return s.requestsProbe(ctx, "api/requests/v2/dictionaries/admin-access")
}

// DictionaryConfigs returns the dictionaries with their configuration. Staff only.
// GET /api/requests/v2/dictionaries/admin
func (s *RequestsV2Service) DictionaryConfigs(ctx context.Context) ([]RequestV2Dictionary, error) {
	return call[[]RequestV2Dictionary](ctx, s.c, get("api/requests/v2/dictionaries/admin", nil))
}

// SaveDictionary creates or updates a dictionary by code. Staff only.
// PUT /api/requests/v2/dictionaries/admin
func (s *RequestsV2Service) SaveDictionary(ctx context.Context, in RequestV2DictionaryInput) error {
	return exec(ctx, s.c, put("api/requests/v2/dictionaries/admin", in))
}

// TestDictionary runs a dictionary configuration against its upstream without
// saving it. configJSON is a [RequestV2DictionaryConfig] as JSON text; testContext
// is a JSON object of placeholder values without the "payload." prefix, for
// example {"faculty":{"id":"1"}}. Staff only.
// POST /api/requests/v2/dictionaries/test
func (s *RequestsV2Service) TestDictionary(ctx context.Context, configJSON string, testContext RawJSON) (*RequestV2DictionaryTest, error) {
	if len(testContext) == 0 {
		testContext = RawJSON("{}")
	}
	body := struct {
		ConfigJSON string  `json:"configJson"`
		Context    RawJSON `json:"context"`
	}{configJSON, testContext}
	return call[*RequestV2DictionaryTest](ctx, s.c, post("api/requests/v2/dictionaries/test", body))
}

// DeleteDictionary deletes a dictionary. Staff only.
// DELETE /api/requests/v2/dictionaries/admin/{code}
func (s *RequestsV2Service) DeleteDictionary(ctx context.Context, code string) error {
	return exec(ctx, s.c, del("api/requests/v2/dictionaries/admin/"+id(code), nil))
}

// Prefills returns the prefill sources available to form questions. Staff only.
// GET /api/requests/v2/prefills
func (s *RequestsV2Service) Prefills(ctx context.Context) ([]RequestV2PrefillRef, error) {
	return call[[]RequestV2PrefillRef](ctx, s.c, get("api/requests/v2/prefills", nil))
}

// ContextEnrichments returns the integrations available to BPMN service tasks. Staff only.
// GET /api/requests/v2/context-enrichments
func (s *RequestsV2Service) ContextEnrichments(ctx context.Context) ([]RequestV2ContextEnrichment, error) {
	return call[[]RequestV2ContextEnrichment](ctx, s.c, get("api/requests/v2/context-enrichments", nil))
}

// ---- Platform roles and users ----

// RequestV2Role is a platform role of the requests service.
type RequestV2Role string

// Platform roles.
const (
	// RequestV2RoleAdmin administers the platform.
	RequestV2RoleAdmin RequestV2Role = "ADMIN"
	// RequestV2RoleEditor edits processes.
	RequestV2RoleEditor RequestV2Role = "EDITOR"
	// RequestV2RoleRequestEditor edits own and assigned requests; the default for users without an assignment.
	RequestV2RoleRequestEditor RequestV2Role = "REQUEST_EDITOR"
)

// RequestV2PlatformRole is the platform role of a user.
type RequestV2PlatformRole struct {
	// UserID is the SSO sub; empty in [RequestsV2Service.MyPlatformRole].
	UserID string        `json:"userId,omitzero"`
	Role   RequestV2Role `json:"role"`
}

// RequestV2User is a user found by [RequestsV2Service.Users].
type RequestV2User struct {
	// UserID is the SSO sub.
	UserID   string `json:"userId"`
	FullName string `json:"fullName"`
	// PersonalID is the ISU number.
	PersonalID FlexID `json:"personalId"`
	Email      string `json:"email"`
	Phone      string `json:"phone"`
}

// MyPlatformRole returns the user's platform role.
// GET /api/requests/v2/platform-roles/me
func (s *RequestsV2Service) MyPlatformRole(ctx context.Context) (*RequestV2PlatformRole, error) {
	return call[*RequestV2PlatformRole](ctx, s.c, get("api/requests/v2/platform-roles/me", nil))
}

// PlatformRoles returns the explicit role assignments. Staff only.
// GET /api/requests/v2/platform-roles
func (s *RequestsV2Service) PlatformRoles(ctx context.Context) ([]RequestV2PlatformRole, error) {
	return call[[]RequestV2PlatformRole](ctx, s.c, get("api/requests/v2/platform-roles", nil))
}

// SetPlatformRole assigns a role to a user (SSO sub). Staff only.
// PUT /api/requests/v2/platform-roles/{userId}
func (s *RequestsV2Service) SetPlatformRole(ctx context.Context, userID string, role RequestV2Role) error {
	return exec(ctx, s.c, withQuery(put("api/requests/v2/platform-roles/"+id(userID), nil), q().set("role", string(role))))
}

// ResetPlatformRole removes the explicit assignment of a user, restoring the default role. Staff only.
// DELETE /api/requests/v2/platform-roles/{userId}
func (s *RequestsV2Service) ResetPlatformRole(ctx context.Context, userID string) error {
	return exec(ctx, s.c, del("api/requests/v2/platform-roles/"+id(userID), nil))
}

// Users finds users by name, email, ISU number or user id. Queries
// usually have at least two characters and ask for 20 results. Staff only.
// GET /api/requests/v2/users
func (s *RequestsV2Service) Users(ctx context.Context, search string, limit int) ([]RequestV2User, error) {
	return call[[]RequestV2User](ctx, s.c, get("api/requests/v2/users", q().set("query", search).set("limit", limit)))
}

// ---- Processes and forms (designer) ----

// RequestV2ProcessStatus is the publication status of a process.
type RequestV2ProcessStatus string

// Process publication statuses.
const (
	RequestV2ProcessDraft     RequestV2ProcessStatus = "DRAFT"
	RequestV2ProcessPublished RequestV2ProcessStatus = "PUBLISHED"
	// RequestV2ProcessSuspended no longer accepts new requests.
	RequestV2ProcessSuspended RequestV2ProcessStatus = "SUSPENDED"
)

// RequestV2Process is a process definition; its alias equals the form alias.
type RequestV2Process struct {
	Alias string `json:"alias"`
	Name  string `json:"name"`
}

// RequestV2ProcessConfig holds the catalog, observer and publication settings of a process.
type RequestV2ProcessConfig struct {
	Status RequestV2ProcessStatus `json:"status"`
	// VisibleInCatalog shows the form in the applicants' catalog.
	VisibleInCatalog bool                       `json:"visibleInCatalog"`
	Observers        []RequestV2ProcessObserver `json:"observers"`
}

// RequestV2ProcessObserver is a user who can monitor the requests of a process.
type RequestV2ProcessObserver struct {
	UserID   string `json:"userId"`
	FullName string `json:"fullName"`
}

// RequestV2FormInput creates or updates the submission form of a process keyed by Alias.
type RequestV2FormInput struct {
	// Alias is the process key.
	Alias       string  `json:"alias"`
	Name        string  `json:"name"`
	Description *string `json:"description"`
	HTMLNote    *string `json:"htmlNote"`
	// CategoryID is RequestV2Category.ID, or nil.
	CategoryID *int64 `json:"categoryId"`
	// Version is usually 1.
	Version int `json:"version"`
	// SchemaJSON is the SurveyJS model as JSON text.
	SchemaJSON string `json:"schemaJson"`
	// ProcessVariableMappingsJSON is a JSON array (as text) mapping form fields to process variables; "[]" by default.
	ProcessVariableMappingsJSON string `json:"processVariableMappingsJson"`
	// AccessCodes are the access list codes allowed to submit; empty means unrestricted.
	AccessCodes []string `json:"accessCodes"`
}

// Processes returns the process definitions. Staff only.
// GET /api/requests/v2/processes
func (s *RequestsV2Service) Processes(ctx context.Context) ([]RequestV2Process, error) {
	return call[[]RequestV2Process](ctx, s.c, get("api/requests/v2/processes", nil))
}

// ProcessXML returns the BPMN 2.0 XML of a process. Staff only.
// GET /api/requests/v2/processes/{key}
func (s *RequestsV2Service) ProcessXML(ctx context.Context, key string) (string, error) {
	return call[string](ctx, s.c, get("api/requests/v2/processes/"+id(key), nil))
}

// DeployProcess saves BPMN XML as a new process definition version and publishes it.
// The result shape is not confirmed. Staff only.
// POST /api/requests/v2/processes/deploy
func (s *RequestsV2Service) DeployProcess(ctx context.Context, bpmnXML string) (RawJSON, error) {
	body := struct {
		BPMNXML string `json:"bpmnXml"`
	}{bpmnXML}
	return call[RawJSON](ctx, s.c, post("api/requests/v2/processes/deploy", body))
}

// CreateProcessDraft creates an empty draft process from a name (up to 120
// characters) and returns it with the generated alias. Staff only.
// POST /api/requests/v2/processes/draft
func (s *RequestsV2Service) CreateProcessDraft(ctx context.Context, name string) (*RequestV2Process, error) {
	body := struct {
		Name string `json:"name"`
	}{name}
	return call[*RequestV2Process](ctx, s.c, post("api/requests/v2/processes/draft", body))
}

// ProcessConfig returns the settings of a process. A process without settings
// answers 404 ([Error.IsNotFound]). Staff only.
// GET /api/requests/v2/processes/{alias}/config
func (s *RequestsV2Service) ProcessConfig(ctx context.Context, alias string) (*RequestV2ProcessConfig, error) {
	return call[*RequestV2ProcessConfig](ctx, s.c, get("api/requests/v2/processes/"+id(alias)+"/config", nil))
}

// UpdateProcessConfig sets catalog visibility and observers (SSO subs) of a process. Staff only.
// PUT /api/requests/v2/processes/{alias}/config
func (s *RequestsV2Service) UpdateProcessConfig(ctx context.Context, alias string, visibleInCatalog bool, observerUserIDs []string) (*RequestV2ProcessConfig, error) {
	body := struct {
		VisibleInCatalog bool     `json:"visibleInCatalog"`
		ObserverUserIDs  []string `json:"observerUserIds"`
	}{visibleInCatalog, observerUserIDs}
	return call[*RequestV2ProcessConfig](ctx, s.c, put("api/requests/v2/processes/"+id(alias)+"/config", body))
}

// PublishProcess lets new requests be submitted. Staff only.
// POST /api/requests/v2/processes/{alias}/publish
func (s *RequestsV2Service) PublishProcess(ctx context.Context, alias string) (*RequestV2ProcessConfig, error) {
	return call[*RequestV2ProcessConfig](ctx, s.c, post("api/requests/v2/processes/"+id(alias)+"/publish", nil))
}

// SuspendProcess stops accepting new requests. Staff only.
// POST /api/requests/v2/processes/{alias}/suspend
func (s *RequestsV2Service) SuspendProcess(ctx context.Context, alias string) (*RequestV2ProcessConfig, error) {
	return call[*RequestV2ProcessConfig](ctx, s.c, post("api/requests/v2/processes/"+id(alias)+"/suspend", nil))
}

// DeleteProcess irreversibly deletes a process and all related data. Staff only.
// DELETE /api/requests/v2/processes/{alias}
func (s *RequestsV2Service) DeleteProcess(ctx context.Context, alias string) error {
	return exec(ctx, s.c, del("api/requests/v2/processes/"+id(alias), nil))
}

// SaveForm creates or updates the submission form of a process by alias. The
// result shape is not confirmed. Staff only.
// POST /api/requests/v2/forms
func (s *RequestsV2Service) SaveForm(ctx context.Context, in RequestV2FormInput) (RawJSON, error) {
	return call[RawJSON](ctx, s.c, post("api/requests/v2/forms", in))
}

// SetFormAccess replaces the access lists allowed to submit an existing form. Staff only.
// PUT /api/requests/v2/forms/{alias}/access
func (s *RequestsV2Service) SetFormAccess(ctx context.Context, alias string, accessCodes []string) error {
	body := struct {
		AccessCodes []string `json:"accessCodes"`
	}{accessCodes}
	return exec(ctx, s.c, put("api/requests/v2/forms/"+id(alias)+"/access", body))
}

// ---- Processing ----

// RequestV2ObservedParams filters the monitoring list. Zero values mean "any".
type RequestV2ObservedParams struct {
	// Page is 0-based.
	Page int
	// Size defaults to 30.
	Size      int
	FormCode  string
	Status    RequestV2Status
	RequestID int64
	// Initiator is free text matched against the initiator.
	Initiator string
}

// RequestV2ObservedType is a form the user may observe.
type RequestV2ObservedType struct {
	// Code is the form alias, used as RequestV2ObservedParams.FormCode.
	Code string `json:"code"`
	Name string `json:"name"`
}

// ObservedRequests returns a page of requests the user may monitor. Staff only.
// GET /api/requests/v2/requests/observed
func (s *RequestsV2Service) ObservedRequests(ctx context.Context, p RequestV2ObservedParams) (*RequestV2Page[RequestV2Summary], error) {
	if p.Size == 0 {
		p.Size = 30
	}
	// All six parameters are always sent, empty ones included.
	qv := query{"formCode": {p.FormCode}, "status": {string(p.Status)}, "initiator": {p.Initiator}}.
		set("page", p.Page).set("size", p.Size).set("requestId", p.RequestID)
	return call[*RequestV2Page[RequestV2Summary]](ctx, s.c, get("api/requests/v2/requests/observed", qv))
}

// ObservedRequestTypes returns the forms the user may observe. Staff only.
// GET /api/requests/v2/requests/observed/types
func (s *RequestsV2Service) ObservedRequestTypes(ctx context.Context) ([]RequestV2ObservedType, error) {
	return call[[]RequestV2ObservedType](ctx, s.c, get("api/requests/v2/requests/observed/types", nil))
}

// Tasks returns the workflow tasks assigned to or claimable by the user. Staff only.
// GET /api/requests/v2/tasks
func (s *RequestsV2Service) Tasks(ctx context.Context) ([]RequestV2Task, error) {
	return call[[]RequestV2Task](ctx, s.c, get("api/requests/v2/tasks", nil))
}

// ---- Process instances (Flowable diagnostics) ----

// RequestV2InstanceStatus is the state of a process instance.
type RequestV2InstanceStatus string

// Process instance statuses; RequestV2InstanceAll is a filter only.
const (
	RequestV2InstanceAll        RequestV2InstanceStatus = "ALL"
	RequestV2InstanceRunning    RequestV2InstanceStatus = "RUNNING"
	RequestV2InstanceFailed     RequestV2InstanceStatus = "FAILED"
	RequestV2InstanceSuspended  RequestV2InstanceStatus = "SUSPENDED"
	RequestV2InstanceCompleted  RequestV2InstanceStatus = "COMPLETED"
	RequestV2InstanceTerminated RequestV2InstanceStatus = "TERMINATED"
)

// RequestV2DeadLetter is the status of a job whose retries are exhausted; see RequestV2JobError.
const RequestV2DeadLetter = "DEAD_LETTER"

// RequestV2InstanceParams filters the process instance list.
type RequestV2InstanceParams struct {
	// Status defaults to ALL on the server.
	Status RequestV2InstanceStatus
	Search string
	// Page is 0-based.
	Page int
	// Size defaults to 50.
	Size int
}

// RequestV2InstancePage is a page of process instances with per-status counters.
type RequestV2InstancePage struct {
	RequestV2CountPage[RequestV2Instance]
	// StatusCounts is keyed by RequestV2InstanceStatus values, ALL included.
	StatusCounts map[string]int `json:"statusCounts"`
}

// RequestV2Instance is a Flowable process instance.
type RequestV2Instance struct {
	ID                       string                     `json:"id"`
	Status                   RequestV2InstanceStatus    `json:"status"`
	HasErrors                bool                       `json:"hasErrors"`
	ProcessDefinitionName    string                     `json:"processDefinitionName"`
	Alias                    string                     `json:"alias"`
	ProcessDefinitionVersion int                        `json:"processDefinitionVersion"`
	CurrentActivities        []RequestV2CurrentActivity `json:"currentActivities"`
	RequestID                *int64                     `json:"requestId"`
	RequestStatus            RequestV2Status            `json:"requestStatus"`
	Initiator                string                     `json:"initiator"`
	StartTime                time.Time                  `json:"startTime"`
	// EndTime is sent by [RequestsV2Service.ProcessInstance] only.
	EndTime          *time.Time `json:"endTime,omitzero"`
	DurationInMillis *int64     `json:"durationInMillis"`
}

// RequestV2CurrentActivity is an active step of a process instance.
type RequestV2CurrentActivity struct {
	ActivityID string `json:"activityId"`
	Name       string `json:"name"`
	// Type is the activity type, e.g. "async-job".
	Type string `json:"type"`
}

// RequestV2InstanceDetail is the diagnostics of a process instance.
type RequestV2InstanceDetail struct {
	Process      RequestV2Instance          `json:"process"`
	Errors       []RequestV2JobError        `json:"errors"`
	Variables    []RequestV2ProcessVariable `json:"variables"`
	Activities   []RequestV2ActivityHistory `json:"activities"`
	DeleteReason string                     `json:"deleteReason"`
}

// RequestV2JobError is a failed job of a process instance.
type RequestV2JobError struct {
	ID string `json:"id"`
	// Status is RequestV2DeadLetter when retries are exhausted; other values wait for a retry.
	Status      string `json:"status"`
	Retries     int    `json:"retries"`
	ElementID   string `json:"elementId"`
	ElementName string `json:"elementName"`
	Message     string `json:"message"`
	StackTrace  string `json:"stackTrace"`
}

// RequestV2ProcessVariable is a variable of a process instance.
type RequestV2ProcessVariable struct {
	Name string `json:"name"`
	Type string `json:"type"`
	// Value is a string or any JSON value.
	Value       RawJSON   `json:"value,omitzero"`
	ExecutionID string    `json:"executionId"`
	TaskID      string    `json:"taskId"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// RequestV2ActivityHistory is a step in the history of a process instance.
type RequestV2ActivityHistory struct {
	ActivityID string `json:"activityId"`
	Name       string `json:"name"`
	Type       string `json:"type"`
	// Status is "ACTIVE", "COMPLETED" or "CANCELLED".
	Status           string    `json:"status"`
	StartTime        time.Time `json:"startTime"`
	DurationInMillis *int64    `json:"durationInMillis"`
	Assignee         string    `json:"assignee"`
}

// ProcessInstances returns a page of process instances. Staff only.
// GET /api/requests/v2/admin/process-instances
func (s *RequestsV2Service) ProcessInstances(ctx context.Context, p RequestV2InstanceParams) (*RequestV2InstancePage, error) {
	if p.Size == 0 {
		p.Size = 50
	}
	qv := q().set("status", string(p.Status)).set("search", p.Search).set("page", p.Page).set("size", p.Size)
	return call[*RequestV2InstancePage](ctx, s.c, get("api/requests/v2/admin/process-instances", qv))
}

// ProcessInstance returns the diagnostics of a process instance. Staff only.
// GET /api/requests/v2/admin/process-instances/{id}
func (s *RequestsV2Service) ProcessInstance(ctx context.Context, instanceID string) (*RequestV2InstanceDetail, error) {
	return call[*RequestV2InstanceDetail](ctx, s.c, get("api/requests/v2/admin/process-instances/"+id(instanceID), nil))
}

// RetryProcessJob puts a dead-letter job back into the queue with new retries;
// its external calls may run again. Staff only.
// POST /api/requests/v2/admin/process-instances/{id}/jobs/{jobId}/retry
func (s *RequestsV2Service) RetryProcessJob(ctx context.Context, instanceID, jobID string) error {
	return exec(ctx, s.c, post("api/requests/v2/admin/process-instances/"+id(instanceID)+"/jobs/"+id(jobID)+"/retry", nil))
}

// RestartProcessActivity cancels and recreates the current execution of a
// step; its external calls may run again. Staff only.
// POST /api/requests/v2/admin/process-instances/{id}/activities/{activityId}/restart
func (s *RequestsV2Service) RestartProcessActivity(ctx context.Context, instanceID, activityID string) error {
	return exec(ctx, s.c, post("api/requests/v2/admin/process-instances/"+id(instanceID)+"/activities/"+id(activityID)+"/restart", nil))
}

// DeleteProcessInstance irreversibly deletes a process instance (steps, jobs,
// history); the request and its data remain. Staff only.
// DELETE /api/requests/v2/admin/process-instances/{id}
func (s *RequestsV2Service) DeleteProcessInstance(ctx context.Context, instanceID string) error {
	return exec(ctx, s.c, del("api/requests/v2/admin/process-instances/"+id(instanceID), nil))
}
