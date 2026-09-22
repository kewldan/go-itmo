package myitmo_test

import (
	"net/url"
	"testing"

	"github.com/kewldan/go-itmo/myitmo"
)

func TestGIASecretaryRequests(t *testing.T) {
	p := myitmo.GIASecretariesParams{Limit: 15, Offset: 15, Year: 2025, EPID: 7, GroupID: "Z3400", SecretaryISU: 100002,
		StatusID: 2, ShowAs: myitmo.GIARoleEPManager, Query: "тест"}
	base := func(extra url.Values) url.Values {
		v := url.Values{"limit": {"15"}, "offset": {"15"}, "year": {"2025"}, "ep_id": {"7"}, "group_id": {"Z3400"}, "query": {"тест"}}
		for k, vs := range extra {
			v[k] = vs
		}
		return v
	}
	terms := []myitmo.GIACoordinatorTerm{{ISU: 100002, YearStart: 2024, YearEnd: 2025}}
	termsJSON := `[{"isu":100002,"year_start":2024,"year_end":2025}]`
	runGIACases(t, []giaCase{
		{"Secretaries", func(c *myitmo.Client) error { _, err := c.GIA.Secretaries(ctx, p); return err },
			"GET", "/api/gia/secretaries/list", base(url.Values{"secretary_isu": {"100002"}, "status_id": {"2"}, "show_as": {"ep_manager"}}), ""},
		{"GeneralStateSecretaries", func(c *myitmo.Client) error { _, err := c.GIA.GeneralStateSecretaries(ctx, p); return err },
			"GET", "/api/gia/secretaries/list_ogek", base(url.Values{"secretary_isu": {"100002"}, "status_id": {"2"}}), ""},
		{"MyStudents", func(c *myitmo.Client) error { _, err := c.GIA.MyStudents(ctx, p); return err },
			"GET", "/api/gia/secretaries/my_students", base(url.Values{"status_id": {"2"}}), ""},
		{"UnassignedStudents", func(c *myitmo.Client) error { _, err := c.GIA.UnassignedStudents(ctx, p); return err },
			"GET", "/api/gia/secretaries/students/unassigned", base(nil), ""},
		{"AssignSecretary", func(c *myitmo.Client) error { return c.GIA.AssignSecretary(ctx, 100002, 11, 12) },
			"POST", "/api/gia/secretaries", nil, `{"secretary_isu":100002,"student_id":[11,12]}`},
		{"ApproveSecretaries", func(c *myitmo.Client) error { return c.GIA.ApproveSecretaries(ctx, 3) },
			"POST", "/api/gia/secretaries/approve", nil, `{"secretary_id":[3]}`},
		{"DeclineSecretaries", func(c *myitmo.Client) error {
			return c.GIA.DeclineSecretaries(ctx, "другой секретарь", 3, 4)
		},
			"DELETE", "/api/gia/secretaries/decline", nil, `{"secretary_id":[3,4],"reject_reason":"другой секретарь"}`},
		{"Coordinators", func(c *myitmo.Client) error { _, err := c.GIA.Coordinators(ctx, 2024); return err },
			"GET", "/api/gia/general_state_coordinator/list", url.Values{"year": {"2024"}}, ""},
		{"AddCoordinators", func(c *myitmo.Client) error { return c.GIA.AddCoordinators(ctx, terms...) },
			"POST", "/api/gia/general_state_coordinator", nil, termsJSON},
		{"RemoveCoordinators", func(c *myitmo.Client) error { return c.GIA.RemoveCoordinators(ctx, terms...) },
			"DELETE", "/api/gia/general_state_coordinator", nil, termsJSON},
	})
}

func TestGIASecretaryDecodes(t *testing.T) {
	f, c := newFake(t)
	// The server spells the total with a Cyrillic "с" in "total_сount".
	f.result(`{"students":[{"student_id":11,"student_isu":100001,"student_surname":"Студентов","group_id":"Z3400","ep_name":"Программа",
		"dir_code":"09.03.01"}],"total_сount":42}`)
	un := must[*myitmo.GIAUnassignedStudents](t)(c.GIA.UnassignedStudents(ctx, myitmo.GIASecretariesParams{}))
	if un.TotalCount != 42 || un.Students[0].DirCode != "09.03.01" {
		t.Errorf("unassigned = %+v", un)
	}

	f.result(`{"secretaries":[{"ep_id":7,"ep_name":"Программа","ep_manager_isu":100003,"groups":[{"group_id":"Z3400","students_count":20,
		"assigned_count":20,"is_general_state":false,"status_id":2,"status_name":"На согласовании",
		"secretary":{"secretary_id":3,"secretary_isu":100002,"secretary_surname":"Секретарев","secretary_email":"sec@example.com"},
		"reject_info":null,"dirs_info":[{"dir_id":8,"dir_code":"09.03.01"}],"students":[]}]}],"total_count":1}`)
	sl := must[*myitmo.GIASecretaryList](t)(c.GIA.Secretaries(ctx, myitmo.GIASecretariesParams{}))
	g := sl.Secretaries[0].Groups[0]
	if sl.TotalCount != 1 || g.Secretary.SecretaryID != 3 || g.RejectInfo != nil || g.DirsInfo[0].DirID != 8 {
		t.Errorf("secretaries = %+v", sl)
	}

	f.result(`{"students":{"secretary_id":3,"secretary_isu":100002,"secretary_surname":"Секретарев","secretary_phone_number":"+70000000000",
		"secretary_students":[{"ep_id":7,"ep_name":"Программа","groups":[{"group_id":"Z3400","status_id":4,"students_count":1,
		"students":[{"secretary_id":3,"student_info":{"student_isu":100001,"student_surname":"Студентов"}}]}]}]},"total_count":1}`)
	my := must[*myitmo.GIAMyStudents](t)(c.GIA.MyStudents(ctx, myitmo.GIASecretariesParams{}))
	if my.Students.SecretaryISU != 100002 || my.Students.SecretaryStudents[0].Groups[0].Students[0].StudentInfo.StudentISU != 100001 {
		t.Errorf("my students = %+v", my)
	}

	f.result(`{"secretaries":[{"id":1,"student_isu":100001,"secretary_surname":"Секретарев","status_name":"Назначен"}],"total_count":1}`)
	og := must[*myitmo.GIAGeneralStateSecretaryList](t)(c.GIA.GeneralStateSecretaries(ctx, myitmo.GIASecretariesParams{}))
	if og.TotalCount != 1 || og.Secretaries[0].StatusName != "Назначен" {
		t.Errorf("ogek = %+v", og)
	}

	f.result(`[{"isu":100002,"surname":"Координаторов","name":"Тест","second_name":""}]`)
	co := must[[]myitmo.GIACoordinator](t)(c.GIA.Coordinators(ctx, 2024))
	if co[0].ISU != 100002 {
		t.Errorf("coordinators = %+v", co)
	}
}
