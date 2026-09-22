package bars

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/kewldan/go-itmo/internal/rest"
)

// route joins escaped path segments. Every segment goes through
// url.PathEscape, so identifiers containing "/" stay one segment.
func route(segments ...string) string {
	escaped := make([]string, len(segments))
	for i, s := range segments {
		escaped[i] = url.PathEscape(s)
	}
	return strings.Join(escaped, "/")
}

// id formats a numeric path segment or query value.
func id(v int64) string { return strconv.FormatInt(v, 10) }

// ids joins numbers with commas, the format of list parameters.
func ids(vs []int64) string {
	parts := make([]string, len(vs))
	for i, v := range vs {
		parts[i] = id(v)
	}
	return strings.Join(parts, ",")
}

// query builds url.Values. Values are encoded by rest.Client.
type query url.Values

func q() query { return query{} }

// str sets a non-empty string.
func (v query) str(key, value string) query {
	if value != "" {
		url.Values(v).Set(key, value)
	}
	return v
}

// num sets a non-zero number.
func (v query) num(key string, value int64) query {
	if value != 0 {
		url.Values(v).Set(key, id(value))
	}
	return v
}

// flag sets key=true when value is true; false is never sent.
func (v query) flag(key string, value bool) query {
	if value {
		url.Values(v).Set(key, "true")
	}
	return v
}

// list sets a comma-joined list when value is non-nil; an empty non-nil
// slice sends an empty value.
func (v query) list(key string, value []int64) query {
	if value != nil {
		url.Values(v).Set(key, ids(value))
	}
	return v
}

func (v query) values() url.Values {
	if len(v) == 0 {
		return nil
	}
	return url.Values(v)
}

func put(path string, body any) *rest.Request {
	return &rest.Request{Method: http.MethodPut, Path: path, JSON: body}
}

func del(path string, qv url.Values) *rest.Request {
	return &rest.Request{Method: http.MethodDelete, Path: path, Query: qv}
}

// exec performs r and ignores the response body.
func (c *Client) exec(ctx context.Context, r *rest.Request) error {
	return c.do(ctx, r, nil)
}

// download performs r and returns the body as a file; the caller closes it.
func (c *Client) download(ctx context.Context, r *rest.Request) (*File, error) {
	if r.Header == nil {
		r.Header = http.Header{}
	}
	r.Header.Set("Accept", "*/*")
	resp, err := c.send(ctx, r) //nolint:bodyclose // owned by the returned File
	if err != nil {
		return nil, err
	}
	return rest.FileFrom(resp), nil
}

func withQuery(r *rest.Request, qv url.Values) *rest.Request {
	r.Query = qv
	return r
}

// raw turns an optional raw JSON body into a request body; empty means none.
func raw(v RawJSON) any {
	if len(v) == 0 {
		return nil
	}
	return v
}
