// Package rest is the HTTP plumbing shared by the MyITMO and BARS clients:
// request building, JSON bodies, multipart uploads and file downloads.
// Response envelopes and error mapping stay in the client packages.
package rest

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"mime"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/kewldan/go-itmo/internal/jsonx"
)

// Client sends requests relative to BaseURL.
type Client struct {
	HTTP      *http.Client
	BaseURL   *url.URL
	UserAgent string
	Logger    *slog.Logger
	// Prepare runs right before a request is sent; clients use it for auth headers.
	Prepare func(*http.Request) error
}

// Request describes one call. Path is resolved against Client.BaseURL unless it is absolute.
type Request struct {
	Method string
	Path   string
	Query  url.Values
	Header http.Header
	// JSON is marshalled as the body when non-nil.
	JSON any
	// Body is sent as is with ContentType when JSON is nil.
	Body        io.Reader
	ContentType string
}

// Send performs the request. The caller owns the response body.
func (c *Client) Send(ctx context.Context, r *Request) (*http.Response, error) {
	target, err := c.BaseURL.Parse(r.Path)
	if err != nil {
		return nil, fmt.Errorf("rest: bad path %q: %w", r.Path, err)
	}
	if len(r.Query) > 0 {
		q := target.Query()
		for k, vs := range r.Query {
			for _, v := range vs {
				q.Add(k, v)
			}
		}
		target.RawQuery = q.Encode()
	}

	body, contentType := r.Body, r.ContentType
	if r.JSON != nil {
		data, err := jsonx.Marshal(r.JSON)
		if err != nil {
			return nil, fmt.Errorf("rest: encode %s %s: %w", r.Method, r.Path, err)
		}
		body, contentType = bytes.NewReader(data), "application/json"
	}

	req, err := http.NewRequestWithContext(ctx, r.Method, target.String(), body)
	if err != nil {
		return nil, err
	}
	for k, vs := range r.Header {
		req.Header[k] = vs
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	if req.Header.Get("Accept") == "" {
		req.Header.Set("Accept", "application/json")
	}
	if c.UserAgent != "" {
		req.Header.Set("User-Agent", c.UserAgent)
	}
	if c.Prepare != nil {
		if err := c.Prepare(req); err != nil {
			return nil, err
		}
	}

	start := time.Now()
	resp, err := c.HTTP.Do(req)
	if c.Logger != nil {
		// Only method, path and status: queries and bodies may carry personal data.
		attrs := []slog.Attr{slog.String("method", r.Method), slog.String("path", target.Path), slog.Duration("elapsed", time.Since(start))}
		if resp != nil {
			attrs = append(attrs, slog.Int("status", resp.StatusCode))
		}
		if err != nil {
			attrs = append(attrs, slog.Any("error", err))
		}
		c.Logger.LogAttrs(ctx, slog.LevelDebug, "itmo request", attrs...)
	}
	return resp, err
}

// ReadAll reads at most limit bytes of the body and closes it.
func ReadAll(resp *http.Response, limit int64) ([]byte, error) {
	defer resp.Body.Close()
	return io.ReadAll(io.LimitReader(resp.Body, limit))
}

// File is a downloaded document. The caller must close Body.
type File struct {
	Name        string
	ContentType string
	Size        int64
	Body        io.ReadCloser
}

// FileFrom wraps a successful response as a File.
func FileFrom(resp *http.Response) *File {
	f := &File{ContentType: resp.Header.Get("Content-Type"), Size: resp.ContentLength, Body: resp.Body}
	if _, params, err := mime.ParseMediaType(resp.Header.Get("Content-Disposition")); err == nil {
		f.Name = params["filename"]
	}
	return f
}

// Upload is one file part of a multipart request.
type Upload struct {
	Field       string
	Name        string
	ContentType string
	Content     io.Reader
}

// Multipart encodes fields and files into a multipart/form-data body.
// Content is streamed through a pipe, so large files are not buffered.
func Multipart(fields map[string]string, files ...Upload) (io.Reader, string) {
	pr, pw := io.Pipe()
	mw := multipart.NewWriter(pw)
	go func() {
		err := func() error {
			for k, v := range fields {
				if err := mw.WriteField(k, v); err != nil {
					return err
				}
			}
			for _, f := range files {
				h := make(map[string][]string)
				h["Content-Disposition"] = []string{fmt.Sprintf(`form-data; name="%s"; filename="%s"`, escapeQuotes(f.Field), escapeQuotes(f.Name))}
				ct := f.ContentType
				if ct == "" {
					ct = "application/octet-stream"
				}
				h["Content-Type"] = []string{ct}
				part, err := mw.CreatePart(h)
				if err != nil {
					return err
				}
				if _, err := io.Copy(part, f.Content); err != nil {
					return err
				}
			}
			return mw.Close()
		}()
		pw.CloseWithError(err)
	}()
	return pr, mw.FormDataContentType()
}

var quoteEscaper = strings.NewReplacer("\\", "\\\\", `"`, "\\\"")

func escapeQuotes(s string) string { return quoteEscaper.Replace(s) }
