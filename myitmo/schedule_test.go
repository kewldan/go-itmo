package myitmo_test

import (
	"net/http"
	"reflect"
	"testing"

	"github.com/kewldan/go-itmo/myitmo"
)

func TestSchedulePersonalSubjectFilterAndIntersections(t *testing.T) {
	f, c := newFake(t)
	f.reply(http.StatusOK, `{"code":0,"message":null,"data":[{"day_number":1,"week_number":4,"date":"2026-09-21","note":null,"type":0,
		"lessons":[
			{"pair_id":10,"subject":"Math","subject_id":2,"work_type":"Лекция","work_type_id":1,"time_start":"08:20","time_end":"09:50",
			 "teacher_id":1,"teacher_name":"Teacher A","room":"101","building":"Main","group":"A1","zoom_url":null,"bld_id":13,"main_bld_id":13,
			 "format":"Очно","format_id":1,"flow_type_id":2,"flow_id":5},
			{"pair_id":11,"subject":"Physics","subject_id":3,"work_type_id":3,"time_start":"09:00","time_end":"10:30","teacher_id":null,"bld_id":null}],
		"intersections":[[10,11]]}]}`)
	days := must[[]myitmo.ScheduleDay](t)(c.Schedule.Personal(ctx, myitmo.NewDate(2026, 9, 21), myitmo.NewDate(2026, 9, 21), 2, 3))
	r := f.expect(http.MethodGet, "/api/schedule/schedule/personal")
	if r.Query.Get("date_start") != "2026-09-21" || !reflect.DeepEqual(r.Query["subject_id[]"], []string{"2", "3"}) {
		t.Errorf("query = %v", r.Query)
	}
	d := days[0]
	if !reflect.DeepEqual(d.Intersections, [][]int64{{10, 11}}) || d.WeekNumber != 4 {
		t.Errorf("day = %+v", d)
	}
	if d.Lessons[0].WorkTypeID != myitmo.WorkTypeLecture || *d.Lessons[0].TeacherID != 1 || d.Lessons[1].BuildingID != nil {
		t.Errorf("lessons = %+v", d.Lessons)
	}

	f.reply(http.StatusOK, `{"code":0,"message":null,"data":[]}`)
	must[[]myitmo.ScheduleDay](t)(c.Schedule.Personal(ctx, myitmo.NewDate(2026, 9, 21), myitmo.NewDate(2026, 9, 27)))
	if r := f.last(); r.Query.Has("subject_id[]") {
		t.Errorf("empty filter sent: %v", r.Query)
	}
}

func TestScheduleTimeSlots(t *testing.T) {
	f, c := newFake(t)
	f.reply(http.StatusOK, `{"code":0,"message":null,"data":[{"id":1,"time_start":"08:20","time_end":"09:50","order":1}]}`)
	slots := must[[]myitmo.TimeSlot](t)(c.Schedule.TimeSlots(ctx))
	f.expect(http.MethodGet, "/api/schedule/meta/time_slots")
	if slots[0].Order != 1 || slots[0].TimeStart != "08:20" {
		t.Errorf("slots = %+v", slots)
	}
}

func TestScheduleSubjects(t *testing.T) {
	f, c := newFake(t)
	f.reply(http.StatusOK, `{"code":0,"message":null,"data":[{"id":2,"name":"Math"}]}`)
	raw := must[myitmo.RawJSON](t)(c.Schedule.Subjects(ctx))
	f.expect(http.MethodGet, "/api/schedule/subjects")
	if string(raw) != `[{"id":2,"name":"Math"}]` {
		t.Errorf("subjects = %s", raw)
	}

	f.reply(http.StatusOK, `{"code":3,"message":"failure","data":null}`)
	if _, err := c.Schedule.Subjects(ctx); myitmo.ErrorCode(err) != 3 {
		t.Errorf("err = %v", err)
	}
}
