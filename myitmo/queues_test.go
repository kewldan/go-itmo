package myitmo_test

import (
	"net/http"
	"net/url"
	"reflect"
	"testing"
	"time"

	"github.com/kewldan/go-itmo/myitmo"
)

func TestQueuesRequests(t *testing.T) {
	tests := []struct {
		name  string
		call  func(c *myitmo.Client) error
		path  string
		query url.Values
	}{
		{"List", func(c *myitmo.Client) error { _, err := c.Queues.List(ctx); return err }, "/api/queues/", url.Values{}},
		{"Get", func(c *myitmo.Client) error { _, err := c.Queues.Get(ctx, 4, 2); return err },
			"/api/queues/", url.Values{"table_id": {"4"}, "type_id": {"2"}}},
		{"Current", func(c *myitmo.Client) error { _, err := c.Queues.Current(ctx); return err }, "/api/queues/current", url.Values{}},
		{"Archive", func(c *myitmo.Client) error { _, err := c.Queues.Archive(ctx); return err }, "/api/queues/archive", url.Values{}},
		{"Slots", func(c *myitmo.Client) error { _, err := c.Queues.Slots(ctx, 4, 2); return err },
			"/api/queues/slots", url.Values{"table_id": {"4"}, "type_id": {"2"}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, c := newFake(t)
			if err := tt.call(c); err != nil {
				t.Fatal(err)
			}
			r := f.expect(http.MethodGet, tt.path)
			if !reflect.DeepEqual(r.Query, tt.query) {
				t.Errorf("query = %v, want %v", r.Query, tt.query)
			}
		})
	}
}

func TestQueuesSignUp(t *testing.T) {
	f, c := newFake(t)
	err := c.Queues.SignUp(ctx, myitmo.QueueSignUp{TimeTableID: 901, Phone: "+7(900)000-00-00"})
	if err != nil {
		t.Fatal(err)
	}
	r := f.expect(http.MethodPost, "/api/queues/")
	if ct := r.Header.Get("Content-Type"); ct != "application/x-www-form-urlencoded" {
		t.Errorf("Content-Type = %q", ct)
	}
	// Exact JSON text expected by the server, nulls included.
	if want := `{"time_table_id":901,"phone":"+7(900)000-00-00","comment":null,"request_id":null}`; string(r.Body) != want {
		t.Errorf("body = %s, want %s", r.Body, want)
	}

	must[struct{}](t)(struct{}{}, c.Queues.SignUp(ctx, myitmo.QueueSignUp{
		TimeTableID: 901, Phone: "+7(900)000-00-00", Comment: myitmo.Ptr("<b>&</b>"), RequestID: myitmo.Ptr[int64](33),
	}))
	r = f.expect(http.MethodPost, "/api/queues/")
	if want := `{"time_table_id":901,"phone":"+7(900)000-00-00","comment":"<b>&</b>","request_id":33}`; string(r.Body) != want {
		t.Errorf("body = %s, want %s", r.Body, want)
	}
}

func TestQueuesCancel(t *testing.T) {
	f, c := newFake(t)
	if err := c.Queues.Cancel(ctx, 2, 5005); err != nil {
		t.Fatal(err)
	}
	f.expect(http.MethodDelete, "/api/queues/").sameJSON(t, `{"type_id":2,"application_id":5005}`)
}

func TestQueuesDecode(t *testing.T) {
	f, c := newFake(t)
	f.result(`[{"table_id":4,"type_id":2,"table_name":"Student office","table_address":null,
		"comment":"<p>Bring your ID</p>","template_id":12}]`)
	q := must[*myitmo.QueueTable](t)(c.Queues.Get(ctx, 4, 2))
	if q.TableName != "Student office" || *q.TemplateID != 12 || q.TableAddress != "" {
		t.Errorf("queue = %+v", q)
	}
	f.result(`[]`)
	if q := must[*myitmo.QueueTable](t)(c.Queues.Get(ctx, 4, 3)); q != nil {
		t.Errorf("queue = %+v, want nil", q)
	}

	f.result(`[{"application_id":5005,"type_id":2,"table_id":4,"table_name":"Student office","table_address":"Room 101",
		"date":"2026-03-02T10:15:00+03:00","phone":"+7(900)000-00-00","comment":null,"queue_comment":"<p>Note</p>"}]`)
	e := must[[]myitmo.QueueEntry](t)(c.Queues.Current(ctx))
	if len(e) != 1 || e[0].ApplicationID != 5005 || e[0].Date.Minute() != 15 || e[0].QueueComment != "<p>Note</p>" {
		t.Errorf("entries = %+v", e)
	}

	f.result(`[{"time_table_id":901,"date":"2026-03-02T10:15:00+03:00"},{"time_table_id":902,"date":"2026-03-02T10:30:00+03:00"}]`)
	s := must[[]myitmo.QueueSlot](t)(c.Queues.Slots(ctx, 4, 2))
	if len(s) != 2 || s[1].TimeTableID != 902 || s[1].Date.Sub(s[0].Date) != 15*time.Minute {
		t.Errorf("slots = %+v", s)
	}
}
