package myitmo_test

import (
	"net/http"
	"net/url"
	"reflect"
	"testing"
	"time"

	"github.com/kewldan/go-itmo/myitmo"
)

func TestAdviserRoutes(t *testing.T) {
	cases := []struct {
		name  string
		path  string
		query url.Values
		do    func(*myitmo.AdviserService) error
	}{
		{"StudentsDefault", "/api/services/adviser/students", url.Values{"limit": {"20"}, "offset": {"0"}}, func(s *myitmo.AdviserService) error {
			_, err := s.Students(ctx, myitmo.AdviserStudentsParams{})
			return err
		}},
		{"Students", "/api/services/adviser/students", url.Values{"query": {"name"}, "course": {"2"}, "op": {"77"}, "group": {"A3100"}, "limit": {"50"}, "offset": {"100"}}, func(s *myitmo.AdviserService) error {
			_, err := s.Students(ctx, myitmo.AdviserStudentsParams{Query: "name", Course: myitmo.Ptr(2), Op: myitmo.Ptr[int64](77), Group: "A3100", Limit: 50, Offset: 100})
			return err
		}},
		{"Student", "/api/services/adviser/students/100001", nil, func(s *myitmo.AdviserService) error { _, err := s.Student(ctx, 100001); return err }},
		{"Activities", "/api/services/adviser/students/100001/activities", url.Values{"q": {"robot"}, "type": {"project"}, "limit": {"20"}, "offset": {"40"}}, func(s *myitmo.AdviserService) error {
			_, err := s.Activities(ctx, 100001, myitmo.AdviserActivitiesParams{Query: "robot", Type: "project", Offset: 40})
			return err
		}},
		{"Scholarship", "/api/services/adviser/students/100001/scholarship", nil, func(s *myitmo.AdviserService) error { _, err := s.Scholarship(ctx, 100001); return err }},
		{"CourseFilter", "/api/services/adviser/filters/course", nil, func(s *myitmo.AdviserService) error { _, err := s.CourseFilter(ctx); return err }},
		{"OpFilter", "/api/services/adviser/filters/op", nil, func(s *myitmo.AdviserService) error { _, err := s.OpFilter(ctx); return err }},
		{"GroupFilter", "/api/services/adviser/filters/groups", nil, func(s *myitmo.AdviserService) error { _, err := s.GroupFilter(ctx); return err }},
		{"RecordBookSpecializations", "/api/services/adviser/services/record_book/specializations", url.Values{"isu": {"100001"}}, func(s *myitmo.AdviserService) error {
			_, err := s.RecordBookSpecializations(ctx, 100001)
			return err
		}},
		{"RecordBookEntries", "/api/services/adviser/services/record_book/555/3", url.Values{"isu": {"100001"}}, func(s *myitmo.AdviserService) error {
			_, err := s.RecordBookEntries(ctx, 100001, 555, 3)
			return err
		}},
		{"RecordBookControls", "/api/services/adviser/services/record_book/details", url.Values{"est_id": {"999"}, "isu": {"100001"}}, func(s *myitmo.AdviserService) error {
			_, err := s.RecordBookControls(ctx, 100001, 999)
			return err
		}},
		{"Programs", "/api/services/adviser/services/program/100001", nil, func(s *myitmo.AdviserService) error { _, err := s.Programs(ctx, 100001); return err }},
		{"PersonPlan", "/api/services/adviser/services/person_plan/100001/555", nil, func(s *myitmo.AdviserService) error { _, err := s.PersonPlan(ctx, 100001, 555); return err }},
		{"Choice", "/api/services/adviser/services/choice/100001/555", nil, func(s *myitmo.AdviserService) error { _, err := s.Choice(ctx, 100001, 555); return err }},
		{"Replacements", "/api/services/adviser/services/replace/100001/555", nil, func(s *myitmo.AdviserService) error { _, err := s.Replacements(ctx, 100001, 555); return err }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			f, c := newFake(t)
			if err := tc.do(c.Adviser); err != nil {
				t.Fatal(err)
			}
			r := f.expect(http.MethodGet, tc.path)
			if len(tc.query) == 0 && len(r.Query) == 0 {
				return
			}
			if !reflect.DeepEqual(r.Query, tc.query) {
				t.Errorf("query = %v, want %v", r.Query, tc.query)
			}
		})
	}
}

func TestAdviserSchedule(t *testing.T) {
	f, c := newFake(t)
	f.reply(http.StatusOK, `{"code":0,"message":"","data":[{"date":"2026-09-07","lessons":[{"time_start":"08:20","time_end":"09:50","subject":"Course","work_type":"Лекция","work_type_id":1,"teacher_name":"Teacher A","teacher_id":7,"room":"101","building":"Main","group":"A3100","note":"","zoom_url":null}]}]}`)
	days := must[[]myitmo.ScheduleDay](t)(c.Adviser.Schedule(ctx, 100001, myitmo.NewDate(2026, time.September, 7), myitmo.NewDate(2026, time.September, 13)))
	r := f.expect(http.MethodGet, "/api/services/adviser/services/schedule")
	want := url.Values{"date_start": {"2026-09-07"}, "date_end": {"2026-09-13"}, "isu": {"100001"}}
	if !reflect.DeepEqual(r.Query, want) {
		t.Errorf("query = %v", r.Query)
	}
	if len(days) != 1 || days[0].Lessons[0].Subject != "Course" || *days[0].Lessons[0].TeacherID != 7 {
		t.Errorf("days = %+v", days)
	}

	f.reply(http.StatusOK, `{"code":5,"message":"denied","data":null}`)
	if _, err := c.Adviser.Schedule(ctx, 100001, myitmo.NewDate(2026, time.September, 7), myitmo.NewDate(2026, time.September, 7)); myitmo.ErrorCode(err) != 5 {
		t.Errorf("err = %v", err)
	}
}

func TestAdviserStudentsDecode(t *testing.T) {
	f, c := newFake(t)
	f.result(`{"count":1,"data":[{"isu":100001,"last_name":"Last","first_name":"First","patronymic":null,"photo":null,
		"education":[{"course":2,"level":"bachelor","faculty_name":"Faculty","group":"A3100","program_code":"09.03.04","program_name":"Program","dir_name":"Direction","status":"studying","is_debt":true}],
		"contacts":[{"contact_alias":"email","contact":["student@example.com"]}]}]}`)
	p := must[*myitmo.Page[myitmo.AdviserStudent]](t)(c.Adviser.Students(ctx, myitmo.AdviserStudentsParams{}))
	if p.Count != 1 || p.Data[0].ISU != 100001 || p.Data[0].Patronymic != "" || !p.Data[0].Education[0].IsDebt || p.Data[0].Contacts[0].Contact[0] != "student@example.com" {
		t.Errorf("page = %+v", p)
	}
}

func TestAdviserStudentDecode(t *testing.T) {
	f, c := newFake(t)
	f.result(`{"fio":"Last First","eng_fio":"Last First","photo":null,"birth_date":"2005-03-04","country":"Country","address":"Address",
		"contacts":[],"rooms":[{"room_number":"101","bld_name":"Dorm"}],"order":{"form":"full-time"},"education":[{"course":1,"group":"A3100"}]}`)
	p := must[*myitmo.AdviserPerson](t)(c.Adviser.Student(ctx, 100001))
	if p.BirthDate != myitmo.NewDate(2005, time.March, 4) || p.Rooms[0].BldName != "Dorm" || p.Order.Form != "full-time" || p.Education[0].Group != "A3100" {
		t.Errorf("person = %+v", p)
	}
}

func TestAdviserActivitiesAndScholarshipDecode(t *testing.T) {
	f, c := newFake(t)
	f.result(`{"count":1,"data":[{"id":1,"type":"project","type_name":"Project","name":"Robot","url":null,"date_begin":"01.09.2025","date_end":null,"date_doc":null,"year":2025,"info":null,"authors":"A. Author","group":null}]}`)
	a := must[*myitmo.Page[myitmo.AdviserActivity]](t)(c.Adviser.Activities(ctx, 100001, myitmo.AdviserActivitiesParams{}))
	if a.Data[0].Type != "project" || *a.Data[0].Year != 2025 || a.Data[0].DateBegin != "01.09.2025" {
		t.Errorf("activities = %+v", a)
	}

	f.result(`{"data":[{"Type":"Academic scholarship","type":3},{"Type":"Social scholarship","type":"social"}]}`)
	s := must[[]myitmo.AdviserScholarship](t)(c.Adviser.Scholarship(ctx, 100001))
	if len(s) != 2 || s[0].Name != "Academic scholarship" || string(s[0].Type) != "3" || string(s[1].Type) != `"social"` {
		t.Errorf("scholarship = %+v", s)
	}
}

func TestAdviserFiltersDecode(t *testing.T) {
	f, c := newFake(t)
	f.result(`[{"value":1},{"value":"2"}]`)
	courses := must[[]myitmo.AdviserCourseOption](t)(c.Adviser.CourseFilter(ctx))
	if len(courses) != 2 || string(courses[0].Value) != "1" {
		t.Errorf("courses = %+v", courses)
	}
	f.result(`[{"id":77,"value":"Program"}]`)
	ops := must[[]myitmo.IDValue](t)(c.Adviser.OpFilter(ctx))
	if ops[0].ID != 77 || ops[0].Value != "Program" {
		t.Errorf("ops = %+v", ops)
	}
	f.result(`[{"value":"A3100"}]`)
	groups := must[[]myitmo.AdviserGroupOption](t)(c.Adviser.GroupFilter(ctx))
	if groups[0].Value != "A3100" {
		t.Errorf("groups = %+v", groups)
	}
}

func TestAdviserRecordBookDecode(t *testing.T) {
	f, c := newFake(t)
	f.result(`[{"main_plan":555,"specialization_name":"Program","semesters":[{"semester":3,"actual":true}]}]`)
	sp := must[[]myitmo.Specialization](t)(c.Adviser.RecordBookSpecializations(ctx, 100001))
	if sp[0].MainPlan != 555 || !sp[0].Semesters[0].Actual {
		t.Errorf("specializations = %+v", sp)
	}
	f.result(`[{"name":"Course","discipline_id":1,"est_id":999,"current_score":71.5,"rate":"4/C","attempt":1,"control_type":"Exam","control_type_id":5,"exam_date":"2026-01-15T10:00:00+03:00","teacher":{"surname":"Teacher","name":"A","patronymic":""}}]`)
	en := must[[]myitmo.RecordBookEntry](t)(c.Adviser.RecordBookEntries(ctx, 100001, 555, 3))
	if en[0].EstID != 999 || *en[0].CurrentScore != 71.5 || en[0].Teacher.Surname != "Teacher" {
		t.Errorf("entries = %+v", en)
	}
	f.result(`[{"id":1,"parent_id":null,"control_name":"Test","date":"2026-01-10T00:00:00+03:00","rate":10,"min_value":0,"max_value":20,"lower_value":5,"required":true,"teacher":{"surname":"Teacher","name":"A","patronymic":""}}]`)
	ce := must[[]myitmo.ControlEntry](t)(c.Adviser.RecordBookControls(ctx, 100001, 999))
	if ce[0].ParentID != nil || *ce[0].Rate != 10 || !ce[0].Required {
		t.Errorf("controls = %+v", ce)
	}
}

func TestAdviserPlanDecode(t *testing.T) {
	f, c := newFake(t)
	f.result(`{"programs":[{"mainPlanId":555,"opName":"Program"}]}`)
	pr := must[*myitmo.AdviserPrograms](t)(c.Adviser.Programs(ctx, 100001))
	if pr.Programs[0].MainPlanID != 555 || pr.Programs[0].OpName != "Program" {
		t.Errorf("programs = %+v", pr)
	}

	f.result(`{"semestersCount":8,"currentSemester":3,"structure":[{"id":1,"type":"module","name":"Module","choiceParameterId":2,"rules":[1],"needChoice":true,"choiceAvailable":true,"creditPoints":12,"rpdUrl":null,"children":[
		{"id":2,"type":"discipline","name":"Course","moduleId":1,"choiceParameterId":1,"creditPoints":4.5,"rpdUrl":"https://example.com/rpd","children":[],
		 "contents":{"3":[{"semester":3,"creditPoints":4.5,"activities":[{"workTypeId":1,"volume":32}]}]}}]}]}`)
	pl := must[*myitmo.AdviserPlan](t)(c.Adviser.PersonPlan(ctx, 100001, 555))
	if pl.SemestersCount != 8 || !pl.Structure[0].NeedChoice {
		t.Fatalf("plan = %+v", pl)
	}
	d := pl.Structure[0].Children[0]
	if *d.ModuleID != 1 || d.CreditPoints != 4.5 || d.Contents["3"][0].Activities[0].Volume != 32 {
		t.Errorf("discipline = %+v", d)
	}

	f.result(`{"chosenContents":[{"moduleId":1,"disciplineId":2,"semester":3}]}`)
	ch := must[*myitmo.AdviserChoice](t)(c.Adviser.Choice(ctx, 100001, 555))
	if ch.ChosenContents[0].DisciplineID != 2 {
		t.Errorf("choice = %+v", ch)
	}

	f.result(`[{"disciplineIdFrom":2,"moduleId":1,"name":"Other course"}]`)
	rp := must[[]myitmo.AdviserReplacement](t)(c.Adviser.Replacements(ctx, 100001, 555))
	if rp[0].DisciplineIDFrom != 2 || rp[0].Name != "Other course" {
		t.Errorf("replacements = %+v", rp)
	}
}
