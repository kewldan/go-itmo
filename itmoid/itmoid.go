// Package itmoid signs users in to ITMO.ID, the Keycloak realm behind every
// ITMO service, and turns the result into OAuth 2.0 tokens.
//
// Each ITMO service is a separate OIDC client ([App]). MyITMO uses the
// authorization code flow with PKCE and issues refresh tokens; BARS only needs
// the authorization code, which its own backend exchanges for a session.
//
// An [Authenticator] keeps the ITMO.ID SSO cookies in its HTTP client, so after
// one password login it can obtain codes for other apps without the password:
//
//	auth := itmoid.New()
//	tok, err := auth.Login(ctx, itmoid.MyITMO, itmoid.Credentials{Username: u, Password: p})
//	// later, for BARS, without asking for the password again:
//	code, err := auth.Authorize(ctx, itmoid.BARS, nil)
package itmoid

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"strings"

	"golang.org/x/net/publicsuffix"
	"golang.org/x/oauth2"
)

// DefaultIssuer is the ITMO.ID realm.
const DefaultIssuer = "https://id.itmo.ru/auth/realms/itmo"

// App is an OIDC client registered in ITMO.ID.
type App struct {
	ClientID    string
	RedirectURL string
	Scopes      []string
	// PKCE enables the S256 code challenge; required by MyITMO.
	PKCE bool
}

// Registered clients of the ITMO services.
var (
	// MyITMO is the student and staff personal cabinet at my.itmo.ru.
	MyITMO = App{
		ClientID:    "student-personal-cabinet",
		RedirectURL: "https://my.itmo.ru/login/callback",
		Scopes:      []string{"openid", "profile"},
		PKCE:        true,
	}
	// MyITMODev is the staging cabinet at dev.my.itmo.su.
	MyITMODev = App{
		ClientID:    "student-personal-cabinet-dev",
		RedirectURL: "https://dev.my.itmo.su/login/callback",
		Scopes:      []string{"openid", "profile"},
		PKCE:        true,
	}
	// BARS is the grade book at bars.itmo.ru. Its backend exchanges the code
	// itself, so there is no PKCE and no refresh token.
	BARS = App{
		ClientID:    "bars",
		RedirectURL: "https://bars.itmo.ru/rest/login",
		Scopes:      []string{"openid"},
	}
)

// Credentials are an ITMO.ID login and password. They are used for a single
// form submission and never stored.
type Credentials struct {
	Username string
	Password string
	// OTP returns a one-time code when the account has two-factor
	// authentication enabled. Nil makes such logins fail with [ErrOTPRequired].
	OTP func(ctx context.Context) (string, error)
}

// Errors returned by the login flow.
var (
	// ErrLoginRequired means the SSO session is absent or expired and
	// ITMO.ID wants the login form filled in.
	ErrLoginRequired = errors.New("itmoid: interactive login required")
	// ErrInvalidCredentials means ITMO.ID rejected the login or password.
	ErrInvalidCredentials = errors.New("itmoid: invalid username or password")
	// ErrOTPRequired means the account needs a one-time code and Credentials.OTP is nil.
	ErrOTPRequired = errors.New("itmoid: one-time code required")
	// ErrUnexpectedPage means ITMO.ID showed a page the flow does not know
	// (password change, terms acceptance and so on); finish it in a browser.
	ErrUnexpectedPage = errors.New("itmoid: unexpected ITMO.ID page")
)

// FormError carries the message ITMO.ID displayed on the login form.
type FormError struct {
	Err     error
	Message string
}

func (e *FormError) Error() string {
	if e.Message == "" {
		return e.Err.Error()
	}
	return fmt.Sprintf("%v: %s", e.Err, e.Message)
}

func (e *FormError) Unwrap() error { return e.Err }

// Authenticator talks to ITMO.ID. It is safe for concurrent use.
type Authenticator struct {
	issuer string
	http   *http.Client
}

// Option configures an [Authenticator].
type Option func(*Authenticator)

// WithIssuer overrides the realm URL, for tests or a staging realm.
func WithIssuer(issuer string) Option {
	return func(a *Authenticator) { a.issuer = strings.TrimSuffix(issuer, "/") }
}

// WithHTTPClient sets the client used for ITMO.ID. The Authenticator replaces
// its CheckRedirect (the code arrives in a redirect) and adds a cookie jar when
// it has none (the SSO session lives in cookies). The client is copied.
func WithHTTPClient(c *http.Client) Option {
	return func(a *Authenticator) { cp := *c; a.http = &cp }
}

// New returns an Authenticator for ITMO.ID.
func New(opts ...Option) *Authenticator {
	a := &Authenticator{issuer: DefaultIssuer, http: &http.Client{}}
	for _, opt := range opts {
		opt(a)
	}
	if a.http.Jar == nil {
		jar, _ := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
		a.http.Jar = jar
	}
	a.http.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	return a
}

// Issuer returns the realm URL.
func (a *Authenticator) Issuer() string { return a.issuer }

// HTTPClient returns the client holding the SSO cookies. It does not follow redirects.
func (a *Authenticator) HTTPClient() *http.Client { return a.http }

// OAuth2Config returns the oauth2 configuration of app in this realm.
func (a *Authenticator) OAuth2Config(app App) *oauth2.Config {
	return &oauth2.Config{
		ClientID:    app.ClientID,
		RedirectURL: app.RedirectURL,
		Scopes:      app.Scopes,
		Endpoint: oauth2.Endpoint{
			AuthURL:   a.issuer + "/protocol/openid-connect/auth",
			TokenURL:  a.issuer + "/protocol/openid-connect/token",
			AuthStyle: oauth2.AuthStyleInParams,
		},
	}
}

func (a *Authenticator) ctx(ctx context.Context) context.Context {
	return context.WithValue(ctx, oauth2.HTTPClient, a.http)
}
