package myitmo_test

import (
	"encoding/json/v2"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync"
	"testing"

	"golang.org/x/oauth2"

	"github.com/kewldan/go-itmo/myitmo"
)

// recorded is what the fake server saw.
type recorded struct {
	Method string
	Path   string
	Query  url.Values
	Header http.Header
	Body   []byte
}

// fake is a MyITMO stand-in that answers every request with the next queued response.
type fake struct {
	t   *testing.T
	srv *httptest.Server

	mu        sync.Mutex
	requests  []recorded
	responses []response
}

type response struct {
	status int
	header http.Header
	body   string
}

func newFake(t *testing.T) (*fake, *myitmo.Client) {
	t.Helper()
	f := &fake{t: t}
	f.srv = httptest.NewServer(http.HandlerFunc(f.serve))
	t.Cleanup(f.srv.Close)
	client := myitmo.New(
		oauth2.StaticTokenSource(&oauth2.Token{AccessToken: "test-token"}),
		myitmo.WithBaseURL(f.srv.URL),
		myitmo.WithQRBaseURL(f.srv.URL+"/qr/"),
	)
	return f, client
}

func (f *fake) serve(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	f.mu.Lock()
	f.requests = append(f.requests, recorded{r.Method, r.URL.EscapedPath(), r.URL.Query(), r.Header.Clone(), body})
	var resp response
	if len(f.responses) > 0 {
		resp, f.responses = f.responses[0], f.responses[1:]
	} else {
		resp = response{status: http.StatusOK, body: `{"error_code":0,"result":null}`}
	}
	f.mu.Unlock()
	for k, vs := range resp.header {
		w.Header()[k] = vs
	}
	if w.Header().Get("Content-Type") == "" {
		w.Header().Set("Content-Type", "application/json")
	}
	w.WriteHeader(resp.status)
	_, _ = io.WriteString(w, resp.body)
}

// result queues a successful standard envelope around result JSON.
func (f *fake) result(resultJSON string) *fake {
	return f.reply(http.StatusOK, `{"error_code":0,"error_message":null,"result":`+resultJSON+`}`)
}

// reply queues a raw response.
func (f *fake) reply(status int, body string) *fake {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.responses = append(f.responses, response{status: status, body: body})
	return f
}

// replyWith queues a raw response with headers.
func (f *fake) replyWith(status int, header http.Header, body string) *fake {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.responses = append(f.responses, response{status: status, header: header, body: body})
	return f
}

// last returns the most recent request.
func (f *fake) last() recorded {
	f.t.Helper()
	f.mu.Lock()
	defer f.mu.Unlock()
	if len(f.requests) == 0 {
		f.t.Fatal("no request was made")
	}
	return f.requests[len(f.requests)-1]
}

// expect asserts method and escaped path (with a leading slash) of the last request.
func (f *fake) expect(method, path string) recorded {
	f.t.Helper()
	r := f.last()
	if r.Method != method || r.Path != path {
		f.t.Fatalf("request = %s %s, want %s %s", r.Method, r.Path, method, path)
	}
	return r
}

// jsonBody decodes the request body into a generic value for comparison.
func (r recorded) jsonBody(t *testing.T) any {
	t.Helper()
	var v any
	if err := json.Unmarshal(r.Body, &v); err != nil {
		t.Fatalf("request body %q is not JSON: %v", r.Body, err)
	}
	return v
}

// sameJSON asserts that the request body is JSON equal to want.
func (r recorded) sameJSON(t *testing.T, want string) {
	t.Helper()
	var w any
	if err := json.Unmarshal([]byte(want), &w); err != nil {
		t.Fatalf("bad expectation %q: %v", want, err)
	}
	got, _ := json.Marshal(r.jsonBody(t), json.Deterministic(true))
	exp, _ := json.Marshal(w, json.Deterministic(true))
	if string(got) != string(exp) {
		t.Fatalf("request body = %s, want %s", got, exp)
	}
}

func must[T any](t *testing.T) func(T, error) T {
	return func(v T, err error) T {
		t.Helper()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		return v
	}
}
