package itmoid_test

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/kewldan/go-itmo/itmoid"
)

const callback = "https://app.example/login/callback"

var testApp = itmoid.App{ClientID: "test-client", RedirectURL: callback, Scopes: []string{"openid"}, PKCE: true}

// fakeKeycloak mimics the ITMO.ID pages that matter to the flow.
type fakeKeycloak struct {
	srv *httptest.Server

	mu        sync.Mutex
	state     string
	challenge string
	verifier  string
	posts     int
}

func newFakeKeycloak(t *testing.T) *fakeKeycloak {
	t.Helper()
	k := &fakeKeycloak{}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /auth/realms/itmo/protocol/openid-connect/auth", k.auth)
	mux.HandleFunc("POST /auth/realms/itmo/login-actions/authenticate", k.authenticate)
	mux.HandleFunc("POST /auth/realms/itmo/protocol/openid-connect/token", k.token)
	k.srv = httptest.NewTLSServer(mux)
	t.Cleanup(k.srv.Close)
	return k
}

func (k *fakeKeycloak) issuer() string { return k.srv.URL + "/auth/realms/itmo" }

func (k *fakeKeycloak) authenticator() *itmoid.Authenticator {
	return itmoid.New(itmoid.WithIssuer(k.issuer()), itmoid.WithHTTPClient(k.srv.Client()))
}

func (k *fakeKeycloak) page(w http.ResponseWriter, template, message string) {
	action := strings.ReplaceAll(k.srv.URL+"/auth/realms/itmo/login-actions/authenticate?session_code=s&execution=e", "/", `\/`)
	msg := ""
	if message != "" {
		msg = fmt.Sprintf(`"message": { "type": "error", "summary": %q },`, message)
	}
	fmt.Fprintf(w, `<html><script>const kcContext = { "url": { "loginAction": "%s" }, %s "templateName": "%s" };</script></html>`, action, msg, template)
}

func (k *fakeKeycloak) redirectWithCode(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, callback+"?state="+k.state+"&code=the-code&iss="+k.issuer(), http.StatusFound)
}

func (k *fakeKeycloak) auth(w http.ResponseWriter, r *http.Request) {
	k.mu.Lock()
	defer k.mu.Unlock()
	q := r.URL.Query()
	k.state, k.challenge = q.Get("state"), q.Get("code_challenge")
	if _, err := r.Cookie("KEYCLOAK_IDENTITY"); err == nil {
		k.redirectWithCode(w, r)
		return
	}
	k.page(w, "login.ftl", "")
}

func (k *fakeKeycloak) authenticate(w http.ResponseWriter, r *http.Request) {
	k.mu.Lock()
	defer k.mu.Unlock()
	k.posts++
	_ = r.ParseForm()
	switch {
	case r.PostForm.Get("otp") == "123456":
	case r.PostForm.Get("otp") != "":
		k.page(w, "login-otp.ftl", "Неверный код")
		return
	case r.PostForm.Get("username") == "otp-user" && r.PostForm.Get("password") == "good":
		k.page(w, "login-otp.ftl", "")
		return
	case r.PostForm.Get("password") != "good":
		k.page(w, "login.ftl", "Неверное имя пользователя или пароль.")
		return
	}
	http.SetCookie(w, &http.Cookie{Name: "KEYCLOAK_IDENTITY", Value: "sso", Path: "/auth/realms/itmo/"})
	k.redirectWithCode(w, r)
}

func (k *fakeKeycloak) token(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	k.mu.Lock()
	k.verifier = r.PostForm.Get("code_verifier")
	k.mu.Unlock()
	if r.PostForm.Get("code") != "the-code" || r.PostForm.Get("client_id") != testApp.ClientID {
		http.Error(w, `{"error":"invalid_grant"}`, http.StatusBadRequest)
		return
	}
	payload := base64.RawURLEncoding.EncodeToString([]byte(`{"sub":"u1","isu":123456,"name":"Test User","exp":4102444800}`))
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"access_token":"at","expires_in":300,"refresh_token":"x.%s.y","refresh_expires_in":1800,"token_type":"Bearer","id_token":"x.%s.y"}`, payload, payload)
}

func TestPasswordLoginExchangesCodeWithPKCE(t *testing.T) {
	k := newFakeKeycloak(t)
	auth := k.authenticator()

	tok, err := auth.Login(context.Background(), testApp, itmoid.Credentials{Username: "user", Password: "good"})
	if err != nil {
		t.Fatalf("Login: %v", err)
	}
	if tok.AccessToken != "at" {
		t.Errorf("access token = %q", tok.AccessToken)
	}
	if k.verifier == "" || k.challenge == "" {
		t.Errorf("PKCE not used: verifier=%q challenge=%q", k.verifier, k.challenge)
	}
	claims, err := itmoid.IDClaims(tok)
	if err != nil {
		t.Fatalf("IDClaims: %v", err)
	}
	if claims.ISU != 123456 || claims.Name != "Test User" || claims.Raw["sub"] != "u1" {
		t.Errorf("claims = %+v", claims)
	}
	if exp, ok := itmoid.RefreshExpiry(tok); !ok || exp.Year() != 2100 {
		t.Errorf("refresh expiry = %v, %v", exp, ok)
	}

	// The SSO cookie now lets the same authenticator log in without a password.
	before := k.posts
	if _, err := auth.LoginSSO(context.Background(), testApp); err != nil {
		t.Fatalf("LoginSSO: %v", err)
	}
	if k.posts != before {
		t.Errorf("SSO login posted the form")
	}
}

func TestWrongPasswordReportsFormMessage(t *testing.T) {
	k := newFakeKeycloak(t)
	_, err := k.authenticator().Login(context.Background(), testApp, itmoid.Credentials{Username: "user", Password: "bad"})
	if !errors.Is(err, itmoid.ErrInvalidCredentials) {
		t.Fatalf("err = %v, want ErrInvalidCredentials", err)
	}
	var fe *itmoid.FormError
	if !errors.As(err, &fe) || fe.Message != "Неверное имя пользователя или пароль." {
		t.Errorf("form message = %v", err)
	}
	if k.posts != 1 {
		t.Errorf("credentials posted %d times, want 1", k.posts)
	}
}

func TestSSOWithoutSessionNeedsLogin(t *testing.T) {
	k := newFakeKeycloak(t)
	if _, err := k.authenticator().Authorize(context.Background(), testApp, nil); !errors.Is(err, itmoid.ErrLoginRequired) {
		t.Fatalf("err = %v, want ErrLoginRequired", err)
	}
}

func TestOTP(t *testing.T) {
	k := newFakeKeycloak(t)
	auth := k.authenticator()
	creds := itmoid.Credentials{Username: "otp-user", Password: "good"}
	if _, err := auth.Login(context.Background(), testApp, creds); !errors.Is(err, itmoid.ErrOTPRequired) {
		t.Fatalf("without OTP: err = %v", err)
	}
	creds.OTP = func(context.Context) (string, error) { return "123456", nil }
	if _, err := auth.Login(context.Background(), testApp, creds); err != nil {
		t.Fatalf("with OTP: %v", err)
	}
}

func TestParseCallbackIsStrict(t *testing.T) {
	iss := itmoid.DefaultIssuer
	ok, err := itmoid.ParseCallback(testApp, iss, callback+"?state=s1&code=abc&iss="+iss, "s1")
	if err != nil || ok != "abc" {
		t.Fatalf("valid callback: %q, %v", ok, err)
	}
	bad := []string{
		"http://app.example/login/callback?state=s1&code=abc",
		"https://app.example.evil/login/callback?state=s1&code=abc",
		"https://user@app.example/login/callback?state=s1&code=abc",
		"https://app.example:8443/login/callback?state=s1&code=abc",
		callback + "/extra?state=s1&code=abc",
		callback + "?state=other&code=abc",
		callback + "?state=s1&code=abc&error=access_denied",
		callback + "?state=s1&code=abc#fragment",
		callback + "?state=s1&code=abc&code=def",
		callback + "?state=s1&code=abc&iss=https%3A%2F%2Fevil.example",
		callback + "?state=s1",
	}
	for _, u := range bad {
		if code, err := itmoid.ParseCallback(testApp, iss, u, "s1"); err == nil {
			t.Errorf("accepted %s -> %q", u, code)
		}
	}
	if _, err := itmoid.ParseCallback(testApp, iss, callback+"?state=s1&code=abc", ""); err == nil {
		t.Error("accepted empty expected state")
	}
}

func TestAuthCodeURL(t *testing.T) {
	u := itmoid.New().AuthCodeURL(itmoid.BARS, "s1", "")
	for _, want := range []string{
		"https://id.itmo.ru/auth/realms/itmo/protocol/openid-connect/auth?",
		"client_id=bars",
		"redirect_uri=https%3A%2F%2Fbars.itmo.ru%2Frest%2Flogin",
		"state=s1",
		"response_type=code",
	} {
		if !strings.Contains(u, want) {
			t.Errorf("%s lacks %s", u, want)
		}
	}
	if strings.Contains(u, "code_challenge") {
		t.Error("BARS must not use PKCE")
	}
}
