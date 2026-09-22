package myitmo

import (
	"context"
	"encoding/json/jsontext"
	"fmt"
	"strconv"
	"strings"

	"github.com/kewldan/go-itmo/internal/jsonx"
	"github.com/kewldan/go-itmo/internal/rest"
)

// giaSet adds key=n to v unless n is zero. Most GIA filters are optional ids
// that are omitted when unset.
func giaSet[T ~int | ~int64](v query, key string, n T) query {
	if n != 0 {
		v.set(key, int64(n))
	}
	return v
}

// giaSetFloat adds key=f to v unless f is zero.
func giaSetFloat(v query, key string, f float64) query {
	if f != 0 {
		v.set(key, strconv.FormatFloat(f, 'f', -1, 64))
	}
	return v
}

// giaJoin joins ids with commas, the format of list filters.
func giaJoin(ids []int64) string {
	parts := make([]string, len(ids))
	for i, n := range ids {
		parts[i] = strconv.FormatInt(n, 10)
	}
	return strings.Join(parts, ",")
}

// giaTruthy reports whether a JSON value is truthy: anything but empty, null, false, 0 and "".
func giaTruthy(v jsontext.Value) bool {
	switch strings.TrimSpace(string(v)) {
	case "", "null", "false", "0", `""`:
		return false
	}
	return true
}

// giaLooseEnvelope is the standard envelope whose error_code may be a
// number or a string (POST .../marks/approve answers
// {"error_code":"presentation_not_defense_passed", ...}).
type giaLooseEnvelope struct {
	ErrorCode    jsontext.Value `json:"error_code"`
	ErrorMessage string         `json:"error_message"`
	Message      string         `json:"message"`
	Detail       jsontext.Value `json:"detail"`
}

// giaExecLoose sends r and turns an HTTP error or a non-zero error_code into
// *Error. Unlike exec and callRaw it accepts a string error_code: the
// generic helpers decode error_code as an int and would drop the whole
// payload. A string code is kept at the start of Error.Message
// ("<code>: <message>"), and Error.Code stays 0.
func giaExecLoose(ctx context.Context, c *Client, r *rest.Request) error {
	resp, err := c.rest.Send(ctx, r) //nolint:bodyclose // closed by rest.ReadAll
	if err != nil {
		return err
	}
	body, err := rest.ReadAll(resp, maxBody)
	if err != nil {
		return fmt.Errorf("myitmo: read %s %s: %w", r.Method, r.Path, err)
	}
	var env giaLooseEnvelope
	decoded := len(body) > 0 && jsonx.Unmarshal(body, &env) == nil
	e := &Error{Method: r.Method, Path: r.Path, StatusCode: resp.StatusCode}
	failed := resp.StatusCode/100 != 2
	if decoded {
		msg := env.ErrorMessage
		if msg == "" {
			msg = env.Message
		}
		if msg == "" && len(env.Detail) > 0 && string(env.Detail) != "null" {
			var s string
			if jsonx.Unmarshal(env.Detail, &s) == nil {
				msg = s
			} else {
				msg = string(env.Detail)
			}
		}
		e.Message = msg
		switch code := strings.TrimSpace(string(env.ErrorCode)); {
		case code == "" || code == "null" || code == "0" || code == `""`:
		case code[0] == '"':
			var s string
			_ = jsonx.Unmarshal(env.ErrorCode, &s)
			if n, err := strconv.Atoi(s); err == nil {
				e.Code = n
				failed = failed || n != 0
			} else {
				e.Message = s
				if msg != "" {
					e.Message += ": " + msg
				}
				failed = true
			}
		default:
			if n, err := strconv.Atoi(code); err == nil {
				e.Code = n
			}
			failed = true
		}
	}
	if !failed {
		return nil
	}
	return e
}

// giaForm builds an empty application/x-www-form-urlencoded request, the way
// a plain HTML form without inputs is submitted.
func giaForm(method, path string) *rest.Request {
	return &rest.Request{Method: method, Path: path, Body: strings.NewReader(""), ContentType: "application/x-www-form-urlencoded"}
}
