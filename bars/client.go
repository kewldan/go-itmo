// Package bars is a client for BARS, the ITMO grade book (https://bars.itmo.ru).
//
// BARS has its own session, separate from MyITMO: an opaque "Bearer ..."
// header that the server returns after exchanging an ITMO.ID authorization
// code. It lives about 30 minutes and cannot be refreshed; instead the client
// can silently obtain a new code through the ITMO.ID SSO session:
//
//	auth := itmoid.New()
//	client := bars.New(bars.WithAuthenticator(auth))
//	err := client.LoginPassword(ctx, itmoid.Credentials{Username: u, Password: p})
//	user, err := client.CurrentUser(ctx)
//
// Catalogues and journals are read in the context of the academic period
// stored on the server for the user (shared with other BARS sessions); use
// [Client.WithPeriod] to switch it safely.
package bars

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/kewldan/go-itmo/internal/jsonx"
	"github.com/kewldan/go-itmo/internal/rest"
	"github.com/kewldan/go-itmo/itmoid"
)

// DefaultBaseURL is the REST root of BARS.
const DefaultBaseURL = "https://bars.itmo.ru/backend/rest/"

// SessionStore keeps the session header between runs.
type SessionStore interface {
	// Load returns the stored "Bearer ..." header or "" when there is none.
	Load(ctx context.Context) (string, error)
	// Save stores the header; "" deletes the session.
	Save(ctx context.Context, authorization string) error
}

// MemoryStore is the default in-memory SessionStore.
type MemoryStore struct {
	mu    sync.RWMutex
	value string
}

// Load implements SessionStore.
func (m *MemoryStore) Load(context.Context) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.value, nil
}

// Save implements SessionStore.
func (m *MemoryStore) Save(_ context.Context, v string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.value = v
	return nil
}

// CodeSource returns a fresh ITMO.ID authorization code for the BARS client
// without user interaction, or [itmoid.ErrLoginRequired].
type CodeSource func(ctx context.Context) (string, error)

// Client talks to BARS. It is safe for concurrent use.
type Client struct {
	rest        rest.Client
	store       SessionStore
	codes       CodeSource
	auth        *itmoid.Authenticator
	redirectURI string

	renewMu  sync.Mutex
	periodMu sync.Mutex
}

type config struct {
	httpClient  *http.Client
	baseURL     string
	store       SessionStore
	codes       CodeSource
	auth        *itmoid.Authenticator
	redirectURI string
	userAgent   string
	logger      *slog.Logger
}

// Option configures a [Client].
type Option func(*config)

// WithHTTPClient sets the HTTP client for BARS requests.
func WithHTTPClient(c *http.Client) Option { return func(o *config) { o.httpClient = c } }

// WithBaseURL overrides the REST root (must end with /backend/rest/ or similar).
func WithBaseURL(u string) Option { return func(o *config) { o.baseURL = u } }

// WithStore persists the session; the default keeps it in memory.
func WithStore(s SessionStore) Option { return func(o *config) { o.store = s } }

// WithCodeSource enables silent renewal: on HTTP 401 the client asks for a
// new code once, exchanges it and repeats the request.
func WithCodeSource(src CodeSource) Option { return func(o *config) { o.codes = src } }

// WithAuthenticator uses an ITMO.ID authenticator for password and SSO logins
// and, unless WithCodeSource is given, for silent renewal through its SSO session.
func WithAuthenticator(a *itmoid.Authenticator) Option { return func(o *config) { o.auth = a } }

// WithRedirectURI overrides the callback registered for the BARS client in ITMO.ID.
func WithRedirectURI(u string) Option { return func(o *config) { o.redirectURI = u } }

// WithUserAgent sets the User-Agent header.
func WithUserAgent(ua string) Option { return func(o *config) { o.userAgent = ua } }

// WithLogger logs method, path, status and latency of every request at debug level.
func WithLogger(l *slog.Logger) Option { return func(o *config) { o.logger = l } }

// New returns a BARS client.
func New(opts ...Option) *Client {
	cfg := config{
		httpClient:  &http.Client{Timeout: 30 * time.Second},
		baseURL:     DefaultBaseURL,
		store:       &MemoryStore{},
		redirectURI: itmoid.BARS.RedirectURL,
		userAgent:   "go-itmo",
	}
	for _, opt := range opts {
		opt(&cfg)
	}
	if !strings.HasSuffix(cfg.baseURL, "/") {
		cfg.baseURL += "/"
	}
	base, err := url.Parse(cfg.baseURL)
	if err != nil {
		panic("bars: invalid base URL: " + err.Error())
	}
	c := &Client{store: cfg.store, codes: cfg.codes, auth: cfg.auth, redirectURI: cfg.redirectURI}
	if c.codes == nil && c.auth != nil {
		c.codes = func(ctx context.Context) (string, error) {
			code, err := c.auth.Authorize(ctx, itmoid.BARS, nil)
			if err != nil {
				return "", err
			}
			return code.Value, nil
		}
	}
	c.rest = rest.Client{HTTP: cfg.httpClient, BaseURL: base, UserAgent: cfg.userAgent, Logger: cfg.logger}
	return c
}

// Error is a failed BARS call. Bodies and URLs are not included: they may
// contain grades or personal data.
type Error struct {
	Method     string
	Path       string
	StatusCode int
}

func (e *Error) Error() string {
	return fmt.Sprintf("bars: %s %s: HTTP %d", e.Method, e.Path, e.StatusCode)
}

// IsUnauthorized reports whether the session is missing, expired or rejected.
func (e *Error) IsUnauthorized() bool { return e.StatusCode == http.StatusUnauthorized }

// IsForbidden reports HTTP 403: the selected role may not use the route.
func (e *Error) IsForbidden() bool { return e.StatusCode == http.StatusForbidden }

// IsNotFound reports HTTP 404.
func (e *Error) IsNotFound() bool { return e.StatusCode == http.StatusNotFound }

// IsLocked reports HTTP 423, which means "the user is not a student in the
// selected year and term" on journal and deadline routes.
func (e *Error) IsLocked() bool { return e.StatusCode == http.StatusLocked }

// ErrNoSession means there is no session and it could not be renewed.
var ErrNoSession = errors.New("bars: no session")

// ValidAuthorization reports whether v looks like a BARS session header: a
// single "Bearer ..." line of sane length.
func ValidAuthorization(v string) bool {
	return strings.HasPrefix(v, "Bearer ") && len(v) >= 16 && len(v) <= 16384 && !strings.ContainsAny(v, "\r\n")
}

// Login exchanges an ITMO.ID authorization code for a session and stores it.
func (c *Client) Login(ctx context.Context, code string) error {
	resp, err := c.rest.Send(ctx, &rest.Request{
		Method: http.MethodGet,
		Path:   "login",
		Query:  url.Values{"code": {code}, "customRedirectUri": {c.redirectURI}},
	})
	if err != nil {
		return err
	}
	_ = resp.Body.Close()
	if resp.StatusCode/100 != 2 {
		return &Error{Method: http.MethodGet, Path: "login", StatusCode: resp.StatusCode}
	}
	authorization := resp.Header.Get("Authorization")
	if !ValidAuthorization(authorization) {
		return &Error{Method: http.MethodGet, Path: "login", StatusCode: http.StatusUnauthorized}
	}
	return c.store.Save(ctx, authorization)
}

// LoginPassword signs in through the ITMO.ID form; needs WithAuthenticator.
func (c *Client) LoginPassword(ctx context.Context, creds itmoid.Credentials) error {
	if c.auth == nil {
		return errors.New("bars: LoginPassword needs WithAuthenticator")
	}
	code, err := c.auth.Authorize(ctx, itmoid.BARS, &creds)
	if err != nil {
		return err
	}
	return c.Login(ctx, code.Value)
}

// LoginSSO signs in with the SSO session of the authenticator, for example
// right after a MyITMO login with the same authenticator.
func (c *Client) LoginSSO(ctx context.Context) error {
	if c.auth == nil {
		return errors.New("bars: LoginSSO needs WithAuthenticator")
	}
	code, err := c.auth.Authorize(ctx, itmoid.BARS, nil)
	if err != nil {
		return err
	}
	return c.Login(ctx, code.Value)
}

// HasSession reports whether a session header is stored.
func (c *Client) HasSession(ctx context.Context) bool {
	v, err := c.store.Load(ctx)
	return err == nil && v != ""
}

// Logout forgets the stored session.
func (c *Client) Logout(ctx context.Context) error { return c.store.Save(ctx, "") }

// renew obtains a new session unless another goroutine already replaced the rejected one.
func (c *Client) renew(ctx context.Context, rejected string) (bool, error) {
	if c.codes == nil {
		return false, nil
	}
	c.renewMu.Lock()
	defer c.renewMu.Unlock()
	if current, _ := c.store.Load(ctx); current != "" && current != rejected {
		return true, nil
	}
	code, err := c.codes(ctx)
	if errors.Is(err, itmoid.ErrLoginRequired) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if err := c.Login(ctx, code); err != nil {
		return false, err
	}
	return true, nil
}

// do sends r with the session, renewing it once on 401, and decodes the body into out.
func (c *Client) do(ctx context.Context, r *rest.Request, out any) error {
	resp, err := c.send(ctx, r) //nolint:bodyclose // closed by rest.ReadAll
	if err != nil {
		return err
	}
	body, err := rest.ReadAll(resp, 64<<20)
	if err != nil {
		return fmt.Errorf("bars: read %s %s: %w", r.Method, r.Path, err)
	}
	if out == nil || len(body) == 0 {
		return nil
	}
	if err := jsonx.Unmarshal(body, out); err != nil {
		return fmt.Errorf("bars: decode %s %s: %w", r.Method, r.Path, err)
	}
	return nil
}

// send performs r with auth and renewal; on success the caller owns the body.
// A non-2xx answer becomes *Error and a rotated session header is stored.
func (c *Client) send(ctx context.Context, r *rest.Request) (*http.Response, error) {
	resp, sent, err := c.exchange(ctx, r) //nolint:bodyclose // returned to the caller or closed below
	if err != nil {
		return nil, err
	}
	if resp.StatusCode/100 != 2 {
		_ = resp.Body.Close()
		return nil, &Error{Method: r.Method, Path: r.Path, StatusCode: resp.StatusCode}
	}
	if err := c.rotate(ctx, resp, sent); err != nil {
		_ = resp.Body.Close()
		return nil, err
	}
	return resp, nil
}

// rotate stores a new session header sent with a successful response.
func (c *Client) rotate(ctx context.Context, resp *http.Response, sent string) error {
	if rotated := resp.Header.Get("Authorization"); ValidAuthorization(rotated) && rotated != sent {
		return c.store.Save(ctx, rotated)
	}
	return nil
}

// exchange performs r with the session, renewing it once on 401, and returns
// the response whatever its status together with the header that was sent.
func (c *Client) exchange(ctx context.Context, r *rest.Request) (*http.Response, string, error) {
	authorization, err := c.store.Load(ctx)
	if err != nil {
		return nil, "", err
	}
	if authorization == "" {
		ok, err := c.renew(ctx, "")
		if err != nil {
			return nil, "", err
		}
		if !ok {
			return nil, "", ErrNoSession
		}
	}
	resp, sent, err := c.attempt(ctx, r) //nolint:bodyclose // returned to the caller or closed below
	if err != nil {
		return nil, "", err
	}
	// Streaming bodies (uploads) cannot be sent twice; JSON bodies are re-encoded per attempt.
	if resp.StatusCode == http.StatusUnauthorized && r.Body == nil {
		_ = resp.Body.Close()
		ok, err := c.renew(ctx, sent)
		if err != nil {
			return nil, "", err
		}
		if !ok {
			return nil, "", &Error{Method: r.Method, Path: r.Path, StatusCode: http.StatusUnauthorized}
		}
		if resp, sent, err = c.attempt(ctx, r); err != nil {
			return nil, "", err
		}
	}
	return resp, sent, nil
}

// anonymous performs r without a session (login and registration forms).
func (c *Client) anonymous(ctx context.Context, r *rest.Request) (*http.Response, error) {
	resp, err := c.rest.Send(ctx, r)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode/100 != 2 {
		_ = resp.Body.Close()
		return nil, &Error{Method: r.Method, Path: r.Path, StatusCode: resp.StatusCode}
	}
	return resp, nil
}

func (c *Client) attempt(ctx context.Context, r *rest.Request) (*http.Response, string, error) {
	authorization, err := c.store.Load(ctx)
	if err != nil {
		return nil, "", err
	}
	cp := *r
	cp.Header = r.Header.Clone()
	if cp.Header == nil {
		cp.Header = http.Header{}
	}
	cp.Header.Set("Authorization", authorization)
	resp, err := c.rest.Send(ctx, &cp)
	return resp, authorization, err
}
