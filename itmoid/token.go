package itmoid

import (
	"context"
	"encoding/base64"
	"encoding/json/v2"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"golang.org/x/oauth2"
)

// TokenSource returns a source that refreshes tok through ITMO.ID when it
// expires. Keycloak rotates refresh tokens, so persist every token passed to
// onRefresh (it may be nil) or the stored refresh token soon stops working.
func (a *Authenticator) TokenSource(ctx context.Context, app App, tok *oauth2.Token, onRefresh func(*oauth2.Token)) oauth2.TokenSource {
	base := a.OAuth2Config(app).TokenSource(a.ctx(context.WithoutCancel(ctx)), tok)
	if onRefresh == nil {
		return base
	}
	return &notifyingSource{base: base, last: tok.AccessToken, notify: onRefresh}
}

// FromRefreshToken returns a source that starts from a stored refresh token.
func (a *Authenticator) FromRefreshToken(ctx context.Context, app App, refreshToken string, onRefresh func(*oauth2.Token)) oauth2.TokenSource {
	return a.TokenSource(ctx, app, &oauth2.Token{RefreshToken: refreshToken}, onRefresh)
}

// Refresh forces a refresh and returns the new tokens.
func (a *Authenticator) Refresh(ctx context.Context, app App, refreshToken string) (*oauth2.Token, error) {
	tok, err := a.OAuth2Config(app).TokenSource(a.ctx(ctx), &oauth2.Token{RefreshToken: refreshToken}).Token()
	if err != nil {
		return nil, fmt.Errorf("itmoid: refresh: %w", err)
	}
	return tok, nil
}

type notifyingSource struct {
	base   oauth2.TokenSource
	mu     sync.Mutex
	last   string
	notify func(*oauth2.Token)
}

func (s *notifyingSource) Token() (*oauth2.Token, error) {
	tok, err := s.base.Token()
	if err != nil {
		return nil, err
	}
	s.mu.Lock()
	changed := tok.AccessToken != s.last
	s.last = tok.AccessToken
	s.mu.Unlock()
	if changed {
		s.notify(tok)
	}
	return tok, nil
}

// Logout ends the SSO session the refresh token belongs to.
func (a *Authenticator) Logout(ctx context.Context, app App, refreshToken string) error {
	form := url.Values{"client_id": {app.ClientID}, "refresh_token": {refreshToken}}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, a.issuer+"/protocol/openid-connect/logout", strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := a.http.Do(req)
	if err != nil {
		return fmt.Errorf("itmoid: logout: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode/100 != 2 {
		return fmt.Errorf("itmoid: logout: HTTP %d", resp.StatusCode)
	}
	return nil
}

// Impersonate exchanges the access token of a privileged user for tokens of
// another user (RFC 8693), the way the MyITMO support console does. Ordinary
// accounts get an error from ITMO.ID.
func (a *Authenticator) Impersonate(ctx context.Context, app App, subjectToken, requestedSubject string) (*oauth2.Token, error) {
	form := url.Values{
		"client_id":            {app.ClientID},
		"grant_type":           {"urn:ietf:params:oauth:grant-type:token-exchange"},
		"subject_token":        {subjectToken},
		"requested_subject":    {requestedSubject},
		"audience":             {app.ClientID},
		"requested_token_type": {"urn:ietf:params:oauth:token-type:refresh_token"},
		"scope":                {strings.Join(app.Scopes, " ")},
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, a.issuer+"/protocol/openid-connect/token", strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := a.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("itmoid: impersonate: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("itmoid: impersonate: HTTP %d", resp.StatusCode)
	}
	var raw struct {
		AccessToken      string `json:"access_token"`
		RefreshToken     string `json:"refresh_token"`
		IDToken          string `json:"id_token"`
		TokenType        string `json:"token_type"`
		ExpiresIn        int64  `json:"expires_in"`
		RefreshExpiresIn int64  `json:"refresh_expires_in"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("itmoid: impersonate: %w", err)
	}
	tok := &oauth2.Token{
		AccessToken:  raw.AccessToken,
		RefreshToken: raw.RefreshToken,
		TokenType:    raw.TokenType,
		ExpiresIn:    raw.ExpiresIn,
		Expiry:       time.Now().Add(time.Duration(raw.ExpiresIn) * time.Second),
	}
	return tok.WithExtra(map[string]any{"id_token": raw.IDToken, "refresh_expires_in": raw.RefreshExpiresIn}), nil
}

// Claims are the ITMO.ID identity claims of the signed-in user.
type Claims struct {
	Subject           string `json:"sub"`
	ISU               int64  `json:"isu,omitzero"`
	Name              string `json:"name,omitzero"`
	GivenName         string `json:"given_name,omitzero"`
	MiddleName        string `json:"middle_name,omitzero"`
	FamilyName        string `json:"family_name,omitzero"`
	PreferredUsername string `json:"preferred_username,omitzero"`
	Email             string `json:"email,omitzero"`
	EmailVerified     bool   `json:"email_verified,omitzero"`
	Picture           string `json:"picture,omitzero"`
	Gender            string `json:"gender,omitzero"`
	Birthdate         string `json:"birthdate,omitzero"`
	IssuedAt          int64  `json:"iat,omitzero"`
	Expires           int64  `json:"exp,omitzero"`
	// Raw holds every claim, including ITMO-specific ones not listed above.
	Raw map[string]any `json:"-"`
}

// IDClaims decodes the ID token that came with tok. The signature is NOT
// verified: use the result for display only, never for authorization.
func IDClaims(tok *oauth2.Token) (*Claims, error) {
	id, _ := tok.Extra("id_token").(string)
	if id == "" {
		return nil, errors.New("itmoid: token has no id_token")
	}
	return ParseJWTClaims(id)
}

// ParseJWTClaims decodes the payload of a JWT without verifying it.
func ParseJWTClaims(jwt string) (*Claims, error) {
	parts := strings.Split(jwt, ".")
	if len(parts) != 3 {
		return nil, errors.New("itmoid: malformed JWT")
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("itmoid: malformed JWT payload: %w", err)
	}
	var c Claims
	if err := json.Unmarshal(payload, &c); err != nil {
		return nil, fmt.Errorf("itmoid: malformed JWT claims: %w", err)
	}
	if err := json.Unmarshal(payload, &c.Raw); err != nil {
		return nil, fmt.Errorf("itmoid: malformed JWT claims: %w", err)
	}
	return &c, nil
}

// RefreshExpiry reports when the refresh token stops working. Keycloak refresh
// tokens are JWTs, so this reads their exp claim without verifying them.
func RefreshExpiry(tok *oauth2.Token) (time.Time, bool) {
	if tok == nil || tok.RefreshToken == "" {
		return time.Time{}, false
	}
	c, err := ParseJWTClaims(tok.RefreshToken)
	if err != nil || c.Expires == 0 {
		return time.Time{}, false
	}
	return time.Unix(c.Expires, 0), true
}

// UserInfo calls the OIDC userinfo endpoint with the access token.
func (a *Authenticator) UserInfo(ctx context.Context, ts oauth2.TokenSource) (*Claims, error) {
	tok, err := ts.Token()
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, a.issuer+"/protocol/openid-connect/userinfo", nil)
	if err != nil {
		return nil, err
	}
	tok.SetAuthHeader(req)
	resp, err := a.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("itmoid: userinfo: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("itmoid: userinfo: HTTP %d", resp.StatusCode)
	}
	var c Claims
	if err := json.Unmarshal(body, &c); err != nil {
		return nil, fmt.Errorf("itmoid: userinfo: %w", err)
	}
	if err := json.Unmarshal(body, &c.Raw); err != nil {
		return nil, fmt.Errorf("itmoid: userinfo: %w", err)
	}
	return &c, nil
}
