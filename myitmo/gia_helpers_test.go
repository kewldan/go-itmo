package myitmo_test

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"testing"

	"github.com/kewldan/go-itmo/myitmo"
)

// giaCase is one request expectation of a GIA method.
type giaCase struct {
	name   string
	call   func(c *myitmo.Client) error
	method string
	path   string
	query  url.Values
	// body is the expected JSON body; "" asserts an empty body.
	body string
}

func runGIACases(t *testing.T, cases []giaCase) {
	t.Helper()
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			f, c := newFake(t)
			if err := tt.call(c); err != nil {
				t.Fatal(err)
			}
			r := f.expect(tt.method, tt.path)
			if len(r.Query) != 0 || len(tt.query) != 0 {
				if !reflect.DeepEqual(r.Query, tt.query) {
					t.Errorf("query = %v, want %v", r.Query, tt.query)
				}
			}
			if tt.body != "" {
				r.sameJSON(t, tt.body)
			} else if len(r.Body) != 0 {
				t.Errorf("unexpected body %s", r.Body)
			}
		})
	}
}

// giaDownload is one expectation of a GIA file download.
type giaDownload struct {
	name   string
	call   func(c *myitmo.Client) (*myitmo.File, error)
	method string
	path   string
	query  url.Values
	body   string
}

func runGIADownloads(t *testing.T, cases []giaDownload) {
	t.Helper()
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			f, c := newFake(t)
			f.replyWith(http.StatusOK, http.Header{
				"Content-Type":        {"application/pdf"},
				"Content-Disposition": {`attachment; filename="doc.pdf"`},
			}, "%PDF-1.4")
			file := must[*myitmo.File](t)(tt.call(c))
			defer file.Body.Close()
			r := f.expect(tt.method, tt.path)
			if len(r.Query) != 0 || len(tt.query) != 0 {
				if !reflect.DeepEqual(r.Query, tt.query) {
					t.Errorf("query = %v, want %v", r.Query, tt.query)
				}
			}
			if tt.body != "" {
				r.sameJSON(t, tt.body)
			}
			body, _ := io.ReadAll(file.Body)
			if file.Name != "doc.pdf" || string(body) != "%PDF-1.4" {
				t.Errorf("file = %q %q", file.Name, body)
			}
		})
	}
}

// giaf is fmt.Sprintf for expected bodies.
func giaf(format string, a ...any) string { return fmt.Sprintf(format, a...) }
