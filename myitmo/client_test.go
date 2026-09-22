package myitmo_test

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/kewldan/go-itmo/myitmo"
)

var ctx = context.Background()

func TestRequestsCarryTokenAndLanguage(t *testing.T) {
	f, c := newFake(t)
	f.result(`[]`)
	must[[]myitmo.Specialization](t)(c.RecordBook.Specializations(ctx))
	r := f.expect(http.MethodGet, "/api/record_book/specializations")
	if got := r.Header.Get("Authorization"); got != "Bearer test-token" {
		t.Errorf("Authorization = %q", got)
	}
	if got := r.Header.Get("Accept-Language"); got != "ru" {
		t.Errorf("Accept-Language = %q", got)
	}
}

func TestForeignHostsGetNoToken(t *testing.T) {
	f, c := newFake(t)
	other, _ := newFake(t)
	other.result(`null`)
	// An absolute URL to another host must not receive the MyITMO token.
	if err := c.Do(ctx, http.MethodGet, other.srv.URL+"/anything", nil, nil, nil); err != nil {
		t.Fatal(err)
	}
	if got := other.last().Header.Get("Authorization"); got != "" {
		t.Errorf("token leaked to foreign host: %q", got)
	}
	_ = f
}

func TestErrorCodeInsideOKResponse(t *testing.T) {
	f, c := newFake(t)
	f.reply(http.StatusOK, `{"error_code":137,"error_message":"нельзя записать студента: [Выбрано 1 занятие в этот день]","result":{"details":["x"]}}`)
	_, err := c.RecordBook.Specializations(ctx)
	var e *myitmo.Error
	if !errors.As(err, &e) {
		t.Fatalf("err = %v, want *Error", err)
	}
	if e.Code != 137 || e.StatusCode != http.StatusOK || e.Message == "" || string(e.Details) != `["x"]` {
		t.Errorf("error = %+v", e)
	}
	if myitmo.ErrorCode(err) != 137 {
		t.Errorf("ErrorCode = %d", myitmo.ErrorCode(err))
	}
}

func TestHTTPErrorKeepsServerMessage(t *testing.T) {
	f, c := newFake(t)
	f.reply(http.StatusForbidden, `{"error_code":403,"error_message":"Нет доступа","result":null}`)
	_, err := c.RecordBook.Specializations(ctx)
	var e *myitmo.Error
	if !errors.As(err, &e) || !e.IsForbidden() || e.Message != "Нет доступа" {
		t.Fatalf("err = %#v", err)
	}
}

func TestScheduleUsesDataEnvelopeAndDates(t *testing.T) {
	f, c := newFake(t)
	f.reply(http.StatusOK, `{"code":0,"message":null,"data":[{"day_number":1,"week_number":3,"date":"2026-09-21","note":null,"lessons":[
		{"pair_id":1,"subject":"Math","subject_id":2,"time_start":"08:20","time_end":"09:50","teacher_id":null,"bld_id":13,"work_type_id":1,"format_id":1,"flow_id":5,"flow_type_id":2}],
		"type":0,"intersections":[]}]}`)
	days := must[[]myitmo.ScheduleDay](t)(c.Schedule.Personal(ctx, myitmo.NewDate(2026, 9, 21), myitmo.NewDate(2026, 9, 27)))
	r := f.expect(http.MethodGet, "/api/schedule/schedule/personal")
	if r.Query.Get("date_start") != "2026-09-21" || r.Query.Get("date_end") != "2026-09-27" {
		t.Errorf("query = %v", r.Query)
	}
	if len(days) != 1 || days[0].Date != myitmo.NewDate(2026, 9, 21) || days[0].Lessons[0].TeacherID != nil || *days[0].Lessons[0].BuildingID != 13 {
		t.Errorf("days = %+v", days)
	}
}

func TestLenientTimestamps(t *testing.T) {
	f, c := newFake(t)
	f.result(`[{"control_name":"a","date":"2026-01-15T10:00:00+03:00"},{"control_name":"b","date":"2026-01-15T10:00"},{"control_name":"c","date":null},{"control_name":"d","date":""}]`)
	entries := must[[]myitmo.ControlEntry](t)(c.RecordBook.ControlEntries(ctx, 1))
	want := time.Date(2026, 1, 15, 10, 0, 0, 0, myitmo.MSK)
	if !entries[0].Date.Equal(want) || !entries[1].Date.Equal(want) {
		t.Errorf("dates = %v, %v", entries[0].Date, entries[1].Date)
	}
	if entries[2].Date != nil {
		t.Errorf("null date = %v", entries[2].Date)
	}
}

func TestSearchAllPaginates(t *testing.T) {
	f, c := newFake(t)
	f.result(`{"count":3,"data":[{"id":1},{"id":2}]}`).result(`{"count":3,"data":[{"id":3}]}`)
	var ids []int64
	for p, err := range c.Personalities.SearchAll(ctx, "Иванов", 2) {
		if err != nil {
			t.Fatal(err)
		}
		ids = append(ids, p.ID)
	}
	if len(ids) != 3 || ids[2] != 3 {
		t.Errorf("ids = %v", ids)
	}
	if r := f.last(); r.Query.Get("offset") != "2" || r.Query.Get("q") != "Иванов" {
		t.Errorf("query = %v", r.Query)
	}
}

func TestQRPass(t *testing.T) {
	f, c := newFake(t)
	f.reply(http.StatusOK, `{"response":{"qr_hex":"12AB"}}`)
	pass := must[*myitmo.Pass](t)(c.QR.Pass(ctx))
	r := f.expect(http.MethodGet, "/qr/v1/user/pass")
	if pass.Hex != "12AB" || r.Header.Get("Authorization") == "" {
		t.Errorf("pass = %+v, auth = %q", pass, r.Header.Get("Authorization"))
	}
}

func TestDateText(t *testing.T) {
	d, err := myitmo.ParseDate("2026-02-28T00:00:00+03:00")
	if err != nil || d.AddDays(1) != myitmo.NewDate(2026, 3, 1) || d.String() != "2026-02-28" {
		t.Errorf("date = %v, %v", d, err)
	}
}
