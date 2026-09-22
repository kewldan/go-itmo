//go:build integration

package myitmo_test

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"golang.org/x/oauth2"

	"github.com/kewldan/go-itmo/itmoid"
	"github.com/kewldan/go-itmo/myitmo"
)

// liveClient builds a client from ITMO_REFRESH_TOKEN_FILE (preferred, updated
// in place because ITMO.ID rotates refresh tokens) or ITMO_REFRESH_TOKEN.
// Only read-only endpoints are called. Nothing about the responses is logged
// beyond counts: they contain personal data.
func liveClient(t *testing.T) *myitmo.Client {
	t.Helper()
	file := os.Getenv("ITMO_REFRESH_TOKEN_FILE")
	token := os.Getenv("ITMO_REFRESH_TOKEN")
	if file != "" {
		b, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("read token file: %v", err)
		}
		token = strings.TrimSpace(string(b))
	}
	if token == "" {
		t.Skip("set ITMO_REFRESH_TOKEN_FILE or ITMO_REFRESH_TOKEN")
	}
	save := func(tok *oauth2.Token) {
		if file != "" && tok.RefreshToken != "" {
			_ = os.WriteFile(file, []byte(tok.RefreshToken), 0o600)
		}
	}
	auth := itmoid.New()
	return myitmo.New(auth.FromRefreshToken(context.Background(), itmoid.MyITMO, token, save))
}

func TestLiveReadOnly(t *testing.T) {
	c := liveClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	checks := map[string]func() (int, error){
		"schedule": func() (int, error) {
			d, err := c.Schedule.Personal(ctx, myitmo.Today(), myitmo.Today().AddDays(7))
			return len(d), err
		},
		"time slots": func() (int, error) { s, err := c.Schedule.TimeSlots(ctx); return len(s), err },
		"record book": func() (int, error) {
			s, err := c.RecordBook.Specializations(ctx)
			if err != nil || len(s) == 0 || len(s[0].Semesters) == 0 {
				return len(s), err
			}
			e, err := c.RecordBook.Entries(ctx, s[0].MainPlan, s[0].Semesters[0].Semester)
			return len(e), err
		},
		"study plans":    func() (int, error) { p, err := c.StudyPlan.Programs(ctx); return countPrograms(p), err },
		"sport semester": func() (int, error) { _, err := c.Sport.CurrentSemester(ctx); return 1, err },
		"sport filters":  func() (int, error) { _, err := c.Sport.Filters(ctx); return 1, err },
		"sport chosen":   func() (int, error) { s, err := c.Sport.Chosen(ctx); return len(s), err },
		"menu":           func() (int, error) { m, err := c.System.Menu(ctx); return countMenu(m), err },
		"dashboard":      func() (int, error) { d, err := c.System.Dashboard(ctx); return len(d), err },
		"requests":       func() (int, error) { r, err := c.Requests.List(ctx); return len(r), err },
		"people": func() (int, error) {
			p, err := c.Personalities.Search(ctx, "Иванов", 5, 0)
			if p == nil {
				return 0, err
			}
			return len(p.Data), err
		},
		"election": func() (int, error) { _, err := c.Election.Availability(ctx); return 1, err },
		"qr":       func() (int, error) { _, err := c.QR.Pass(ctx); return 1, err },
	}
	for name, check := range checks {
		t.Run(name, func(t *testing.T) {
			n, err := check()
			var apiErr *myitmo.Error
			var decErr *myitmo.DecodeError
			switch {
			case errors.As(err, &decErr):
				t.Errorf("model mismatch: %v", err)
			case errors.As(err, &apiErr):
				t.Logf("server error (not a model problem): HTTP %d code %d", apiErr.StatusCode, apiErr.Code)
			case err != nil:
				t.Error(err)
			default:
				t.Logf("ok, %d items", n)
			}
		})
	}
}

func countPrograms(p *myitmo.StudyPlanPrograms) int {
	if p == nil {
		return 0
	}
	return len(p.Programs)
}

func countMenu(m *myitmo.MenuResponse) int {
	if m == nil {
		return 0
	}
	return len(m.Menu)
}
