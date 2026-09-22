package myitmo

import (
	"context"
	"encoding/json/jsontext"
	"errors"
	"fmt"
	"iter"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/kewldan/go-itmo/internal/jsonx"
	"github.com/kewldan/go-itmo/internal/rest"
)

// jsonValue is raw JSON kept for decoding later.
type jsonValue = jsontext.Value

// maxBody bounds JSON responses; the largest known payload (a full study plan) is a few MB.
const maxBody = 64 << 20

// envelope is the standard MyITMO wrapper: error_code 0 means success.
type envelope[T any] struct {
	ErrorCode    int    `json:"error_code"`
	ErrorMessage string `json:"error_message"`
	Result       T      `json:"result"`
}

// dataEnvelope is the wrapper of the older schedule service: code 0 means success.
type dataEnvelope[T any] struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    T      `json:"data"`
}

// call sends r and decodes the "result" of the standard envelope. The result
// is decoded only after error_code is checked, so an error payload of another
// shape still surfaces as *Error.
func call[T any](ctx context.Context, c *Client, r *rest.Request) (T, error) {
	var out T
	var env envelope[jsonValue]
	if err := c.send(ctx, r, &env); err != nil {
		return out, err
	}
	if env.ErrorCode != 0 {
		return out, &Error{Method: r.Method, Path: r.Path, StatusCode: http.StatusOK, Code: env.ErrorCode, Message: env.ErrorMessage, Details: details(env.Result)}
	}
	if err := decodeInto(env.Result, &out); err != nil {
		return out, &DecodeError{Method: r.Method, Path: r.Path, Err: err}
	}
	return out, nil
}

// details extracts result.details of an error envelope (requests v2 sends a string or a list).
func details(result jsonValue) RawJSON {
	var d struct {
		Details RawJSON `json:"details"`
	}
	if len(result) == 0 || jsonx.Unmarshal(result, &d) != nil {
		return nil
	}
	return d.Details
}

// callData sends r and decodes the "data" of the schedule envelope.
func callData[T any](ctx context.Context, c *Client, r *rest.Request) (T, error) {
	var env dataEnvelope[T]
	if err := c.send(ctx, r, &env); err != nil {
		return env.Data, err
	}
	if env.Code != 0 {
		return env.Data, &Error{Method: r.Method, Path: r.Path, StatusCode: http.StatusOK, Code: env.Code, Message: env.Message}
	}
	return env.Data, nil
}

// callRaw sends r and decodes the whole body; for services without an envelope.
func callRaw[T any](ctx context.Context, c *Client, r *rest.Request) (T, error) {
	var out T
	err := c.send(ctx, r, &out)
	return out, err
}

// exec sends r and only checks the envelope; the result is discarded.
func exec(ctx context.Context, c *Client, r *rest.Request) error {
	_, err := call[jsonValue](ctx, c, r)
	return err
}

// download sends r and returns the body as a file. The caller closes it.
func download(ctx context.Context, c *Client, r *rest.Request) (*File, error) {
	if r.Header == nil {
		r.Header = http.Header{}
	}
	r.Header.Set("Accept", "*/*")
	resp, err := c.rest.Send(ctx, r) //nolint:bodyclose // the caller closes File.Body
	if err != nil {
		return nil, err
	}
	if resp.StatusCode/100 != 2 {
		return nil, errorFrom(r, resp)
	}
	return rest.FileFrom(resp), nil
}

func (c *Client) send(ctx context.Context, r *rest.Request, out any) error {
	resp, err := c.rest.Send(ctx, r) //nolint:bodyclose // closed by errorFrom or rest.ReadAll
	if err != nil {
		return err
	}
	if resp.StatusCode/100 != 2 {
		return errorFrom(r, resp)
	}
	body, err := rest.ReadAll(resp, maxBody)
	if err != nil {
		return fmt.Errorf("myitmo: read %s %s: %w", r.Method, r.Path, err)
	}
	if len(body) == 0 || resp.StatusCode == http.StatusNoContent {
		return nil
	}
	if err := jsonx.Unmarshal(body, out); err != nil {
		return &DecodeError{Method: r.Method, Path: r.Path, Err: err}
	}
	return nil
}

func errorFrom(r *rest.Request, resp *http.Response) error {
	body, _ := rest.ReadAll(resp, 64<<10)
	e := &Error{Method: r.Method, Path: r.Path, StatusCode: resp.StatusCode}
	var env struct {
		ErrorCode    int       `json:"error_code"`
		ErrorMessage string    `json:"error_message"`
		Message      string    `json:"message"`
		Detail       any       `json:"detail"`
		Result       jsonValue `json:"result"`
	}
	if jsonx.Unmarshal(body, &env) == nil {
		e.Code = env.ErrorCode
		e.Details = details(env.Result)
		e.Message = env.ErrorMessage
		if e.Message == "" {
			e.Message = env.Message
		}
		if e.Message == "" && env.Detail != nil {
			e.Message = fmt.Sprint(env.Detail)
		}
	}
	return e
}

func decodeInto(v jsonValue, out any) error {
	if len(v) == 0 || string(v) == "null" {
		return nil
	}
	return jsonx.Unmarshal(v, out)
}

// Error is a failed MyITMO call: an HTTP error or a non-zero error_code.
type Error struct {
	Method     string
	Path       string
	StatusCode int
	// Code is the service error_code; 0 when the service sent none.
	Code int
	// Message is the localised server message, safe to show to the user.
	Message string
	// Details is result.details of the error envelope, when the service sends it.
	Details RawJSON
}

func (e *Error) Error() string {
	var b strings.Builder
	fmt.Fprintf(&b, "myitmo: %s %s: HTTP %d", e.Method, e.Path, e.StatusCode)
	if e.Code != 0 {
		fmt.Fprintf(&b, ", error_code %d", e.Code)
	}
	if e.Message != "" {
		b.WriteString(": ")
		b.WriteString(e.Message)
	}
	return b.String()
}

// IsUnauthorized reports whether the access token was rejected.
func (e *Error) IsUnauthorized() bool { return e.StatusCode == http.StatusUnauthorized }

// IsForbidden reports whether the user lacks access to the section (e.g. staff-only services).
func (e *Error) IsForbidden() bool { return e.StatusCode == http.StatusForbidden }

// IsNotFound reports whether the resource does not exist.
func (e *Error) IsNotFound() bool { return e.StatusCode == http.StatusNotFound }

// ErrorCode returns the service error_code of err, or 0.
func ErrorCode(err error) int {
	if e, ok := errors.AsType[*Error](err); ok {
		return e.Code
	}
	return 0
}

// DecodeError means the response did not match the model. It usually means
// the server changed a field type; please report it.
type DecodeError struct {
	Method, Path string
	Err          error
}

func (e *DecodeError) Error() string {
	return fmt.Sprintf("myitmo: decode %s %s: %v", e.Method, e.Path, e.Err)
}
func (e *DecodeError) Unwrap() error { return e.Err }

// File is a downloaded document. The caller must close Body.
type File = rest.File

// Upload is a file sent in a multipart request.
type Upload = rest.Upload

// query builds url.Values, skipping zero values. Supported value types are
// strings, integers, bools, Date, *int64, *bool and slices of int64/string
// (repeated keys).
type query url.Values

func q() query { return query{} }

func (v query) set(key string, value any) query {
	switch x := value.(type) {
	case string:
		if x != "" {
			url.Values(v).Set(key, x)
		}
	case int:
		url.Values(v).Set(key, strconv.Itoa(x))
	case int64:
		url.Values(v).Set(key, strconv.FormatInt(x, 10))
	case bool:
		url.Values(v).Set(key, strconv.FormatBool(x))
	case *int64:
		if x != nil {
			url.Values(v).Set(key, strconv.FormatInt(*x, 10))
		}
	case *int:
		if x != nil {
			url.Values(v).Set(key, strconv.Itoa(*x))
		}
	case *bool:
		if x != nil {
			url.Values(v).Set(key, strconv.FormatBool(*x))
		}
	case *string:
		if x != nil {
			url.Values(v).Set(key, *x)
		}
	case Date:
		if !x.IsZero() {
			url.Values(v).Set(key, x.String())
		}
	case []int64:
		for _, n := range x {
			url.Values(v).Add(key, strconv.FormatInt(n, 10))
		}
	case []string:
		for _, s := range x {
			url.Values(v).Add(key, s)
		}
	default:
		panic(fmt.Sprintf("myitmo: unsupported query value %T", value))
	}
	return v
}

func (v query) values() url.Values { return url.Values(v) }

// id formats a path segment.
func id[T ~int | ~int64 | ~string](v T) string {
	switch x := any(v).(type) {
	case string:
		return url.PathEscape(x)
	case int:
		return strconv.Itoa(x)
	case int64:
		return strconv.FormatInt(x, 10)
	}
	return url.PathEscape(fmt.Sprint(v))
}

func get(path string, qv query) *rest.Request {
	return &rest.Request{Method: http.MethodGet, Path: path, Query: qv.values()}
}

func post(path string, body any) *rest.Request {
	return &rest.Request{Method: http.MethodPost, Path: path, JSON: body}
}

func put(path string, body any) *rest.Request {
	return &rest.Request{Method: http.MethodPut, Path: path, JSON: body}
}

func patch(path string, body any) *rest.Request {
	return &rest.Request{Method: http.MethodPatch, Path: path, JSON: body}
}

func del(path string, body any) *rest.Request {
	return &rest.Request{Method: http.MethodDelete, Path: path, JSON: body}
}

// multipart builds a multipart/form-data request.
func multipart(method, path string, fields map[string]string, files ...Upload) *rest.Request {
	body, ct := rest.Multipart(fields, files...)
	return &rest.Request{Method: method, Path: path, Body: body, ContentType: ct}
}

// withQuery attaches query parameters to a request.
func withQuery(r *rest.Request, qv query) *rest.Request {
	r.Query = qv.values()
	return r
}

// Event is one Server-Sent Event of a streaming endpoint.
type Event = rest.Event

// stream sends r and returns its Server-Sent Events. Iteration closes the body.
func stream(ctx context.Context, c *Client, r *rest.Request) (iter.Seq2[Event, error], error) {
	if r.Header == nil {
		r.Header = http.Header{}
	}
	r.Header.Set("Accept", "text/event-stream")
	resp, err := c.rest.Send(ctx, r) //nolint:bodyclose // closed when iteration stops
	if err != nil {
		return nil, err
	}
	if resp.StatusCode/100 != 2 {
		return nil, errorFrom(r, resp)
	}
	return rest.Events(resp.Body), nil
}

// withHeader sets a request header.
func withHeader(r *rest.Request, key, value string) *rest.Request {
	if r.Header == nil {
		r.Header = http.Header{}
	}
	r.Header.Set(key, value)
	return r
}
