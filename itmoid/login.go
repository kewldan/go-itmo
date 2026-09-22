package itmoid

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json/v2"
	"errors"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"golang.org/x/oauth2"
)

// Code is an authorization code together with the PKCE verifier and state it was requested with.
type Code struct {
	Value    string
	Verifier string
	State    string
}

// NewState returns a random OAuth state value.
func NewState() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}

// AuthCodeURL returns the ITMO.ID login page for app. Pass an empty verifier
// for apps without PKCE. Use it to drive the login in a browser or WebView,
// then read the code with [ParseCallback].
func (a *Authenticator) AuthCodeURL(app App, state, verifier string) string {
	var opts []oauth2.AuthCodeOption
	if app.PKCE && verifier != "" {
		opts = append(opts, oauth2.S256ChallengeOption(verifier))
	}
	return a.OAuth2Config(app).AuthCodeURL(state, opts...)
}

// maxSteps bounds the number of ITMO.ID pages a single login may walk through.
const maxSteps = 8

// Authorize obtains an authorization code for app. With nil creds it relies on
// the SSO cookies of a previous login and returns [ErrLoginRequired] when
// ITMO.ID asks for a password.
func (a *Authenticator) Authorize(ctx context.Context, app App, creds *Credentials) (*Code, error) {
	code := &Code{State: NewState()}
	if app.PKCE {
		code.Verifier = oauth2.GenerateVerifier()
	}
	next, err := http.NewRequestWithContext(ctx, http.MethodGet, a.AuthCodeURL(app, code.State, code.Verifier), nil)
	if err != nil {
		return nil, err
	}

	credentialsSent, otpSent := false, false
	for range maxSteps {
		resp, err := a.http.Do(next)
		if err != nil {
			return nil, fmt.Errorf("itmoid: %w", err)
		}
		body, err := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
		_ = resp.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("itmoid: read page: %w", err)
		}

		if loc := resp.Header.Get("Location"); resp.StatusCode/100 == 3 && loc != "" {
			target, err := resp.Request.URL.Parse(loc)
			if err != nil {
				return nil, fmt.Errorf("itmoid: bad redirect: %w", err)
			}
			if isCallback(app, target) {
				value, err := ParseCallback(app, a.issuer, target.String(), code.State)
				if err != nil {
					return nil, err
				}
				code.Value = value
				return code, nil
			}
			// Keycloak walks through its own pages (required actions and similar) via redirects.
			if !sameHost(target, a.issuer) {
				return nil, fmt.Errorf("%w: redirect to %s", ErrUnexpectedPage, target.Host)
			}
			if next, err = http.NewRequestWithContext(ctx, http.MethodGet, target.String(), nil); err != nil {
				return nil, err
			}
			continue
		}
		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("itmoid: unexpected HTTP %d from ITMO.ID", resp.StatusCode)
		}

		page := parsePage(body)
		if page.action == "" {
			return nil, fmt.Errorf("%w: no form on %q", ErrUnexpectedPage, page.template)
		}
		form := url.Values{}
		switch page.template {
		case "login-otp.ftl", "login-otp":
			if otpSent {
				return nil, &FormError{Err: ErrInvalidCredentials, Message: page.message}
			}
			if creds == nil || creds.OTP == nil {
				return nil, ErrOTPRequired
			}
			otp, err := creds.OTP(ctx)
			if err != nil {
				return nil, fmt.Errorf("itmoid: otp: %w", err)
			}
			form.Set("otp", otp)
			otpSent = true
		case "", "login.ftl", "login":
			if creds == nil {
				return nil, ErrLoginRequired
			}
			if credentialsSent {
				return nil, &FormError{Err: ErrInvalidCredentials, Message: page.message}
			}
			form.Set("username", creds.Username)
			form.Set("password", creds.Password)
			form.Set("rememberMe", "on")
			credentialsSent = true
		default:
			return nil, &FormError{Err: fmt.Errorf("%w: %s", ErrUnexpectedPage, page.template), Message: page.message}
		}

		action, err := resp.Request.URL.Parse(page.action)
		if err != nil || !sameHost(action, a.issuer) {
			return nil, fmt.Errorf("%w: form posts outside ITMO.ID", ErrUnexpectedPage)
		}
		next, err = http.NewRequestWithContext(ctx, http.MethodPost, action.String(), strings.NewReader(form.Encode()))
		if err != nil {
			return nil, err
		}
		next.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	return nil, fmt.Errorf("%w: too many steps", ErrUnexpectedPage)
}

// Exchange trades a code for tokens. Apps without PKCE (BARS) exchange codes
// on their own backend instead; see the bars package.
func (a *Authenticator) Exchange(ctx context.Context, app App, code *Code) (*oauth2.Token, error) {
	var opts []oauth2.AuthCodeOption
	if code.Verifier != "" {
		opts = append(opts, oauth2.VerifierOption(code.Verifier))
	}
	tok, err := a.OAuth2Config(app).Exchange(a.ctx(ctx), code.Value, opts...)
	if err != nil {
		return nil, fmt.Errorf("itmoid: exchange code: %w", err)
	}
	return tok, nil
}

// Login signs in with a password and returns tokens for app.
func (a *Authenticator) Login(ctx context.Context, app App, creds Credentials) (*oauth2.Token, error) {
	code, err := a.Authorize(ctx, app, &creds)
	if err != nil {
		return nil, err
	}
	return a.Exchange(ctx, app, code)
}

// LoginSSO returns tokens for app using the SSO session of a previous login.
func (a *Authenticator) LoginSSO(ctx context.Context, app App) (*oauth2.Token, error) {
	code, err := a.Authorize(ctx, app, nil)
	if err != nil {
		return nil, err
	}
	return a.Exchange(ctx, app, code)
}

// ParseCallback extracts the code from a redirect to app.RedirectURL. It
// rejects anything but the exact HTTPS callback with the expected state, a
// single code, no error, no fragment and, when present, the expected issuer.
func ParseCallback(app App, issuer, callbackURL, expectedState string) (string, error) {
	u, err := url.Parse(callbackURL)
	if err != nil || expectedState == "" || !isCallback(app, u) || u.Fragment != "" || u.RawFragment != "" {
		return "", fmt.Errorf("itmoid: not a valid %s callback", app.ClientID)
	}
	q, err := url.ParseQuery(u.RawQuery)
	if err != nil {
		return "", errors.New("itmoid: malformed callback query")
	}
	for _, vs := range q {
		if len(vs) > 1 {
			return "", errors.New("itmoid: repeated callback parameter")
		}
	}
	if e := q.Get("error"); e != "" {
		return "", fmt.Errorf("itmoid: authorization denied: %s", e)
	}
	if q.Get("state") != expectedState {
		return "", errors.New("itmoid: callback state mismatch")
	}
	if iss := q.Get("iss"); iss != "" && issuer != "" && iss != issuer {
		return "", errors.New("itmoid: callback from another issuer")
	}
	code := q.Get("code")
	if code == "" || len(code) > 4096 {
		return "", errors.New("itmoid: callback has no code")
	}
	return code, nil
}

func isCallback(app App, u *url.URL) bool {
	cb, err := url.Parse(app.RedirectURL)
	if err != nil {
		return false
	}
	return u.Scheme == cb.Scheme && strings.EqualFold(u.Host, cb.Host) && u.User == nil && u.EscapedPath() == cb.EscapedPath()
}

func sameHost(u *url.URL, base string) bool {
	b, err := url.Parse(base)
	return err == nil && u.Scheme == "https" && strings.EqualFold(u.Host, b.Host) && u.User == nil
}

type page struct {
	template string
	action   string
	message  string
}

var (
	reTemplate   = regexp.MustCompile(`"templateName"\s*:\s*"((?:[^"\\]|\\.)*)"`)
	reAction     = regexp.MustCompile(`"loginAction"\s*:\s*"((?:[^"\\]|\\.)*)"`)
	reMessage    = regexp.MustCompile(`"message"\s*:\s*\{[^{}]*?"summary"\s*:\s*"((?:[^"\\]|\\.)*)"`)
	reFormAction = regexp.MustCompile(`<form[^>]*\baction="([^"]+)"`)
	reHTMLError  = regexp.MustCompile(`(?s)<span[^>]*(?:id="input-error"|class="[^"]*kc-feedback-text[^"]*")[^>]*>(.*?)</span>`)
)

// parsePage reads the Keycloak login page. The ITMO.ID theme embeds a
// kcContext JSON object; the classic theme is an HTML form.
func parsePage(body []byte) page {
	var p page
	if m := reTemplate.FindSubmatch(body); m != nil {
		p.template = unquote(m[1])
	}
	if m := reAction.FindSubmatch(body); m != nil {
		p.action = unquote(m[1])
	} else if m := reFormAction.FindSubmatch(body); m != nil {
		p.action = html.UnescapeString(string(m[1]))
	}
	if m := reMessage.FindSubmatch(body); m != nil {
		p.message = unquote(m[1])
	} else if m := reHTMLError.FindSubmatch(body); m != nil {
		p.message = strings.TrimSpace(html.UnescapeString(string(m[1])))
	}
	return p
}

func unquote(raw []byte) string {
	var s string
	if err := json.Unmarshal(append(append([]byte{'"'}, raw...), '"'), &s); err != nil {
		return string(raw)
	}
	return s
}
