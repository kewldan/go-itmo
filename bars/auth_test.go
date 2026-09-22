package bars_test

import (
	"errors"
	"net/http"
	"testing"

	"github.com/kewldan/go-itmo/bars"
)

func TestLoginParent(t *testing.T) {
	f := newFakeBARS(t)
	store := &bars.MemoryStore{}
	c := f.client(bars.WithStore(store))
	f.queue(step{header: map[string]string{"Authorization": "Bearer synthetic-parent"}})
	ok(t, c.LoginParent(ctx, "parent-login", "synthetic-password"))
	r := f.expect(http.MethodPost, "login")
	r.sameJSON(t, `{"login":"parent-login","password":"synthetic-password"}`)
	if r.auth != "" {
		t.Errorf("login sent a session: %q", r.auth)
	}
	if v, _ := store.Load(ctx); v != "Bearer synthetic-parent" {
		t.Errorf("session = %q", v)
	}

	f.queue(step{status: http.StatusNotFound})
	var e *bars.Error
	if err := c.LoginParent(ctx, "nobody", "x"); !errors.As(err, &e) || !e.IsNotFound() {
		t.Fatalf("err = %v", err)
	}
	f.queue(step{})
	if err := c.LoginParent(ctx, "parent-login", "x"); !errors.As(err, &e) || !e.IsUnauthorized() {
		t.Fatalf("missing header: err = %v", err)
	}
}

func TestImpersonateStoresRotatedSession(t *testing.T) {
	f := newFakeBARS(t)
	store := &bars.MemoryStore{}
	_ = store.Save(ctx, "Bearer synthetic-super")
	c := f.client(bars.WithStore(store))
	f.queue(step{header: map[string]string{"Authorization": "Bearer synthetic-other"}})
	ok(t, c.Impersonate(ctx, "200 000/1"))
	f.expect(http.MethodGet, "set_me_to/200%20000%2F1").noBody(t)
	if v, _ := store.Load(ctx); v != "Bearer synthetic-other" {
		t.Errorf("session = %q", v)
	}
}

func TestRegisterParent(t *testing.T) {
	f := newFakeBARS(t)
	c := f.client()
	f.reply("")
	ok(t, c.RegisterParent(ctx, "tok/en", "parent", "secret-7"))
	r := f.expect(http.MethodPost, "registration/tok%2Fen")
	r.sameJSON(t, `{"login":"parent","password":"secret-7","retryPassword":"secret-7"}`)
	if r.auth != "" || c.HasSession(ctx) {
		t.Errorf("registration used or created a session")
	}
	f.queue(step{status: http.StatusBadRequest})
	var e *bars.Error
	if err := c.RegisterParent(ctx, "t", "parent", "secret-7"); !errors.As(err, &e) || e.StatusCode != http.StatusBadRequest {
		t.Fatalf("err = %v", err)
	}
}

func TestRegistrationTokens(t *testing.T) {
	f, c := session(t)
	f.reply(`["aaa","bbb"]`)
	tokens := must[[]string](t)(c.RegistrationTokens(ctx))
	f.expect(http.MethodGet, "registration/token")
	if len(tokens) != 2 || tokens[1] != "bbb" {
		t.Errorf("tokens = %v", tokens)
	}

	f.reply("")
	ok(t, c.CreateRegistrationToken(ctx))
	f.expect(http.MethodPost, "registration/token").noBody(t)

	f.reply("")
	ok(t, c.DeleteRegistrationToken(ctx, "a b"))
	f.expect(http.MethodDelete, "registration/token/a%20b")
}

func TestParents(t *testing.T) {
	f, c := session(t)
	f.reply(`[{"id":1}]`)
	v := must[bars.RawJSON](t)(c.Parents(ctx, 42))
	f.expect(http.MethodGet, "users/42/parent")
	if string(v) != `[{"id":1}]` {
		t.Errorf("parents = %s", v)
	}

	f.reply(`{"ok":true}`)
	must[bars.RawJSON](t)(c.AddParent(ctx, 42, bars.RawJSON(`{"login":"p","password":"x"}`)))
	f.expect(http.MethodPost, "users/42/parent").sameJSON(t, `{"login":"p","password":"x"}`)

	f.reply("")
	must[bars.RawJSON](t)(c.AddParent(ctx, 42, nil))
	f.expect(http.MethodPost, "users/42/parent").noBody(t)
}
