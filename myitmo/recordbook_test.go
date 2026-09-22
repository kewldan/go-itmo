package myitmo_test

import (
	"net/http"
	"testing"

	"github.com/kewldan/go-itmo/myitmo"
)

func TestRecordBookSpecializations(t *testing.T) {
	f, c := newFake(t)
	f.result(`[{"main_plan":500,"specialization_name":"Programme","semesters":[{"study_year":"2026/2027","semester":3,"course":2,"actual":true}]}]`)
	specs := must[[]myitmo.Specialization](t)(c.RecordBook.Specializations(ctx))
	f.expect(http.MethodGet, "/api/record_book/specializations")
	if specs[0].MainPlan != 500 || !specs[0].Semesters[0].Actual || specs[0].Semesters[0].Course != 2 {
		t.Errorf("specializations = %+v", specs)
	}
}

func TestRecordBookEntries(t *testing.T) {
	f, c := newFake(t)
	f.result(`[{"name":" Math ","discipline_id":7,"est_id":70,"current_score":null,"rate":"","attempt":1,"control_type":"Экзамен",
		"control_type_id":1,"exam_date":"2027-01-15T10:00:00+03:00","have_tree":true,"lms_link":null,
		"teacher":{"surname":"Surname","name":"Name","patronymic":""}}]`)
	entries := must[[]myitmo.RecordBookEntry](t)(c.RecordBook.Entries(ctx, 500, 3))
	f.expect(http.MethodGet, "/api/record_book/500/3")
	e := entries[0]
	if e.EstID != 70 || e.CurrentScore != nil || e.ExamDate == nil || !e.HaveTree || e.Teacher.Surname != "Surname" {
		t.Errorf("entry = %+v", e)
	}
}

func TestRecordBookControlEntries(t *testing.T) {
	f, c := newFake(t)
	f.result(`[{"id":1,"control_name":"Total","parent_id":null,"lower_value":60,"max_value":100,"min_value":0,"required":true,"rate":null,"date":null,"teacher":null},
		{"id":2,"control_name":"Lab 1","parent_id":1,"max_value":10,"required":false,"rate":8.5}]`)
	entries := must[[]myitmo.ControlEntry](t)(c.RecordBook.ControlEntries(ctx, 70))
	f.expect(http.MethodGet, "/api/record_book/70")
	if entries[0].ParentID != nil || *entries[1].ParentID != 1 || *entries[1].Rate != 8.5 || *entries[0].LowerValue != 60 {
		t.Errorf("entries = %+v", entries)
	}
}
