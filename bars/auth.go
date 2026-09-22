package bars

import (
	"context"
	"net/http"

	"github.com/kewldan/go-itmo/internal/rest"
)

type passwordLogin struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

// LoginParent signs in with a BARS login and password and stores the session.
// Such accounts are created for parents through [Client.RegisterParent]
// ("Вход для родителей"). The server answers 404 for
// an unknown login and 401 for a wrong password. Silent renewal through
// ITMO.ID does not apply to these accounts.
//
// POST /login
//
// Audience: Parent.
func (c *Client) LoginParent(ctx context.Context, login, password string) error {
	r := &rest.Request{Method: http.MethodPost, Path: "login", JSON: passwordLogin{Login: login, Password: password}}
	resp, err := c.anonymous(ctx, r)
	if err != nil {
		return err
	}
	_ = resp.Body.Close()
	authorization := resp.Header.Get("Authorization")
	if !ValidAuthorization(authorization) {
		return &Error{Method: r.Method, Path: r.Path, StatusCode: http.StatusUnauthorized}
	}
	return c.store.Save(ctx, authorization)
}

// Impersonate switches the session to the user with the given personal
// (ISU) number; pass User.SuperUserPersonalNumber to switch back. A rotated
// session header in the answer is stored. Requires User.CanChangeUser.
//
// Warning: this GET changes the identity of the session on the server.
//
// GET /set_me_to/{personalNumber}
//
// Audience: Admin (superuser, office manager).
func (c *Client) Impersonate(ctx context.Context, personalNumber string) error {
	return c.exec(ctx, get(route("set_me_to", personalNumber), nil))
}

type parentRegistration struct {
	Login         string `json:"login"`
	Password      string `json:"password"`
	RetryPassword string `json:"retryPassword"`
}

// RegisterParent creates a parent account from a registration token that a
// student issued with [Client.CreateRegistrationToken]. It needs no session
// and does not sign in; call [Client.LoginParent] afterwards. The login
// must have at most 255 characters and the password 7 to 255 characters; the server answers 400 when the user already exists.
//
// POST /registration/{token}
//
// Audience: Parent (guest).
func (c *Client) RegisterParent(ctx context.Context, token, login, password string) error {
	resp, err := c.anonymous(ctx, post(route("registration", token), parentRegistration{Login: login, Password: password, RetryPassword: password})) //nolint:bodyclose // closed by rest.ReadAll
	if err != nil {
		return err
	}
	_, err = rest.ReadAll(resp, 1<<20)
	return err
}

// RegistrationTokens returns the student's parent registration tokens. The
// registration link is https://bars.itmo.ru/register/{token}.
//
// GET /registration/token
//
// Audience: Student.
func (c *Client) RegistrationTokens(ctx context.Context) ([]string, error) {
	return fetch[[]string](ctx, c, get("registration/token", nil))
}

// CreateRegistrationToken issues a new parent registration token; list the
// tokens afterwards to read it. Keep one token at a time.
//
// POST /registration/token
//
// Audience: Student.
func (c *Client) CreateRegistrationToken(ctx context.Context) error {
	return c.exec(ctx, post("registration/token", nil))
}

// DeleteRegistrationToken revokes a parent registration token.
//
// DELETE /registration/token/{token}
//
// Audience: Student.
func (c *Client) DeleteRegistrationToken(ctx context.Context, token string) error {
	return c.exec(ctx, del(route("registration", "token", token), nil))
}

// Parents returns the parents of a student. The response shape is unknown.
//
// GET /users/{studentId}/parent
//
// Audience: unknown.
func (c *Client) Parents(ctx context.Context, studentID int64) (RawJSON, error) {
	return fetch[RawJSON](ctx, c, get(route("users", id(studentID), "parent"), nil))
}

// AddParent adds a parent to a student. The body and the response are
// unknown, so both are raw JSON.
//
// POST /users/{studentId}/parent
//
// Audience: unknown.
func (c *Client) AddParent(ctx context.Context, studentID int64, body RawJSON) (RawJSON, error) {
	return fetch[RawJSON](ctx, c, post(route("users", id(studentID), "parent"), raw(body)))
}
