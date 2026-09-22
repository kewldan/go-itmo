package bars_test

import (
	"encoding/json/v2"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/kewldan/go-itmo/bars"
)

type step struct {
	status int
	header map[string]string
	body   string
}

type seen struct {
	method, path, auth, body string
}

type fakeBARS struct {
	t     *testing.T
	srv   *httptest.Server
	mu    sync.Mutex
	steps []step
	seen  []seen
}

func newFakeBARS(t *testing.T) *fakeBARS {
	f := &fakeBARS{t: t}
	f.srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		f.mu.Lock()
		defer f.mu.Unlock()
		path := r.URL.EscapedPath()
		if r.URL.RawQuery != "" {
			path += "?" + r.URL.RawQuery
		}
		f.seen = append(f.seen, seen{r.Method, path, r.Header.Get("Authorization"), string(body)})
		if len(f.steps) == 0 {
			t.Errorf("unexpected request %s %s", r.Method, path)
			w.WriteHeader(http.StatusTeapot)
			return
		}
		s := f.steps[0]
		f.steps = f.steps[1:]
		for k, v := range s.header {
			w.Header().Set(k, v)
		}
		if s.status == 0 {
			s.status = http.StatusOK
		}
		w.WriteHeader(s.status)
		_, _ = io.WriteString(w, s.body)
	}))
	t.Cleanup(f.srv.Close)
	return f
}

func (f *fakeBARS) queue(steps ...step) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.steps = append(f.steps, steps...)
}

func (f *fakeBARS) requests() []seen {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]seen(nil), f.seen...)
}

func (f *fakeBARS) client(opts ...bars.Option) *bars.Client {
	return bars.New(append([]bars.Option{bars.WithBaseURL(f.srv.URL + "/backend/rest/")}, opts...)...)
}

func withSession(v string) bars.Option {
	s := &bars.MemoryStore{}
	_ = s.Save(ctx, v)
	return bars.WithStore(s)
}

// session returns a fake and a client with a stored synthetic session.
func session(t *testing.T) (*fakeBARS, *bars.Client) {
	t.Helper()
	f := newFakeBARS(t)
	return f, f.client(withSession("Bearer synthetic-token"))
}

// reply queues a 200 answer with body.
func (f *fakeBARS) reply(body string) *fakeBARS {
	f.queue(step{body: body})
	return f
}

// last returns the most recent request.
func (f *fakeBARS) last() seen {
	f.t.Helper()
	reqs := f.requests()
	if len(reqs) == 0 {
		f.t.Fatal("no request was made")
	}
	return reqs[len(reqs)-1]
}

// expect asserts method and escaped path with query (keys sorted) of the last request.
func (f *fakeBARS) expect(method, path string) seen {
	f.t.Helper()
	r := f.last()
	if r.method != method || r.path != "/backend/rest/"+path {
		f.t.Fatalf("request = %s %s, want %s /backend/rest/%s", r.method, r.path, method, path)
	}
	return r
}

// noBody asserts that the request had no body.
func (s seen) noBody(t *testing.T) {
	t.Helper()
	if s.body != "" {
		t.Fatalf("request body = %s, want none", s.body)
	}
}

// sameJSON asserts that the request body is JSON equal to want.
func (s seen) sameJSON(t *testing.T, want string) {
	t.Helper()
	var got, exp any
	if err := json.Unmarshal([]byte(s.body), &got); err != nil {
		t.Fatalf("request body %q is not JSON: %v", s.body, err)
	}
	if err := json.Unmarshal([]byte(want), &exp); err != nil {
		t.Fatalf("bad expectation %q: %v", want, err)
	}
	g, _ := json.Marshal(got, json.Deterministic(true))
	e, _ := json.Marshal(exp, json.Deterministic(true))
	if string(g) != string(e) {
		t.Fatalf("request body = %s, want %s", g, e)
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

func mustJSON(t *testing.T, data string, v any) {
	t.Helper()
	if err := json.Unmarshal([]byte(data), v); err != nil {
		t.Fatalf("decode %q: %v", data, err)
	}
}

func ok(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
