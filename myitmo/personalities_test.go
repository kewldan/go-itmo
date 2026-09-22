package myitmo_test

import (
	"net/http"
	"testing"

	"github.com/kewldan/go-itmo/myitmo"
)

func TestPersonalitiesGet(t *testing.T) {
	f, c := newFake(t)
	f.result(`{"isu":100003,"fio":"Тестов Тест Тестович","gender":"male","photo":"https://photo.example.test/p/1/",
		"contacts":[{"contact":["+7 000 000-00-00"],"contact_alias":"phone"}],
		"rooms":[{"room_number":"101","bld_name":"Главный корпус"}],
		"positions":[{"department_name":"Кафедра","department_link":"https://example.test/dep","position_name":"доцент","sorting":1,
			"vacation":{"type":"annual"},"start_vacation":"2026-07-01","end_vacation":"2026-07-28"},
			{"department_name":"Лаборатория","department_link":"","position_name":"научный сотрудник","sorting":"2","vacation":null,"start_vacation":null,"end_vacation":null}],
		"powers":[{"power_id":5,"power_name":"Руководитель программы","dep_name":"Факультет","dep_link":"https://example.test/f"}],
		"levels":{"rank":null,"degree":"кандидат наук"},
		"education":[{"course":"1","faculty_name":"Факультет","group":"A0000"}],
		"activities":[],
		"exchange_training":false}`)
	p := must[*myitmo.Person](t)(c.Personalities.Get(ctx, 100003))
	f.expect(http.MethodGet, "/api/personalities/persons/100003")
	if p.ISU != 100003 || len(p.Positions) != 2 || len(p.Powers) != 1 || p.Powers[0].PowerID != 5 ||
		p.Levels == nil || p.Levels.Rank != "" || p.Levels.Degree != "кандидат наук" || p.Contacts[0].ContactAlias != "phone" {
		t.Fatalf("person = %+v", p)
	}
	pos := p.Positions
	if !pos[0].OnVacation() || pos[0].StartVacation != myitmo.NewDate(2026, 7, 1) || pos[0].Sorting != "1" ||
		pos[1].OnVacation() || !pos[1].EndVacation.IsZero() || pos[1].Sorting != "2" {
		t.Errorf("positions = %+v", pos)
	}

	f.result(`{"isu":100004,"fio":"Студент","levels":null,"powers":[],"positions":[]}`)
	if p := must[*myitmo.Person](t)(c.Personalities.Get(ctx, 100004)); p.Levels != nil {
		t.Errorf("levels = %+v", p.Levels)
	}
}

func TestPersonalitiesSearch(t *testing.T) {
	f, c := newFake(t)
	f.result(`{"count":1,"data":[{"id":100003,"fio":"Тестов Тест","gender":"female","phone":null,"email":"t@example.test","work":null,
		"education":"Факультет, группа A0000","photo":null}]}`)
	page := must[*myitmo.Page[myitmo.PersonSummary]](t)(c.Personalities.Search(ctx, "Тестов", 12, 24))
	r := f.expect(http.MethodGet, "/api/personalities/persons")
	if r.Query.Get("q") != "Тестов" || r.Query.Get("limit") != "12" || r.Query.Get("offset") != "24" {
		t.Errorf("query = %v", r.Query)
	}
	if page.Count != 1 || page.Data[0].Education != "Факультет, группа A0000" || page.Data[0].Work != "" {
		t.Errorf("page = %+v", page)
	}
}

func TestPersonalitiesActivities(t *testing.T) {
	f, c := newFake(t)
	f.result(`{"count":2,"data":[
		{"type":"project","type_name":"Проект","name":"Тестовый проект","url":"https://example.test/p","date_begin":"01.02.2025","date_end":"01.02.2026","date_doc":null,"year":null,"info":null,"authors":null},
		{"type":"article","type_name":"Статья","name":"Статья о тестах","url":null,"date_begin":null,"date_end":null,"date_doc":null,"year":2024,"info":"Журнал","authors":"Тестов Т."}]}`)
	page := must[*myitmo.Page[myitmo.PersonActivity]](t)(c.Personalities.Activities(ctx, 100003,
		myitmo.PersonActivitiesParams{Query: "тест", Type: myitmo.ActivityArticle, Limit: 20, Offset: 40}))
	r := f.expect(http.MethodGet, "/api/personalities/persons/100003/activities")
	if got, want := r.Query.Encode(), "limit=20&offset=40&q=%D1%82%D0%B5%D1%81%D1%82&type=article"; got != want {
		t.Errorf("query = %q, want %q", got, want)
	}
	if page.Count != 2 || page.Data[0].Type != myitmo.ActivityProject || page.Data[0].DateBegin != "01.02.2025" ||
		page.Data[1].Year != "2024" || page.Data[0].Year != "" || page.Data[1].Authors != "Тестов Т." {
		t.Errorf("page = %+v", page)
	}

	must[*myitmo.Page[myitmo.PersonActivity]](t)(c.Personalities.Activities(ctx, 100003, myitmo.PersonActivitiesParams{}))
	if got := f.last().Query.Encode(); got != "offset=0" {
		t.Errorf("default query = %q", got)
	}
}
