package myitmo_test

import (
	"net/url"
	"testing"

	"github.com/kewldan/go-itmo/myitmo"
)

func TestGIACommitteeRequests(t *testing.T) {
	in := myitmo.GIACommitteeInput{ChairmanExpertID: 5, EduProgramIDs: []int64{7}, EduDirectionIDs: []int64{8},
		InternalExperts: []int64{100001, 100002}, ExternalExperts: []int64{6, 9}, IsScratch: true}
	inJSON := `{"chairman_expert_id":5,"edu_program_ids":[7],"edu_direction_ids":[8],"internal_experts":[100001,100002],
		"external_experts":[6,9],"is_scratch":true}`
	p := myitmo.GIACommitteesParams{Limit: 15, Year: 2025, StatusID: 3, EduProgramID: 7, EduDirectionID: 8, FacultyID: 2,
		Role: myitmo.GIARoleOSOP, Query: "тест"}
	pq := url.Values{"limit": {"15"}, "offset": {"0"}, "year": {"2025"}, "status_id": {"3"}, "edu_program_id": {"7"},
		"edu_direction_id": {"8"}, "faculty_id": {"2"}, "role": {"osop"}, "query": {"тест"}}
	runGIACases(t, []giaCase{
		{"Committees", func(c *myitmo.Client) error { _, err := c.GIA.Committees(ctx, p); return err },
			"GET", "/api/gia/committees", pq, ""},
		{"CreateCommittee", func(c *myitmo.Client) error { return c.GIA.CreateCommittee(ctx, in) },
			"POST", "/api/gia/committees", nil, inJSON},
		{"Committee", func(c *myitmo.Client) error { _, err := c.GIA.Committee(ctx, 3); return err },
			"GET", "/api/gia/committees/3", nil, ""},
		{"UpdateCommittee", func(c *myitmo.Client) error { return c.GIA.UpdateCommittee(ctx, 3, in) },
			"PATCH", "/api/gia/committees/3", nil, inJSON},
		{"ApproveCommittee", func(c *myitmo.Client) error { return c.GIA.ApproveCommittee(ctx, 3) },
			"PATCH", "/api/gia/committees/3/status", nil, `{}`},
		{"RejectCommittee", func(c *myitmo.Client) error { return c.GIA.RejectCommittee(ctx, 3, "мало членов") },
			"PATCH", "/api/gia/committees/3/status", nil, `{"comment":"мало членов"}`},
		{"RevokeCommittee", func(c *myitmo.Client) error { return c.GIA.RevokeCommittee(ctx, 3) },
			"POST", "/api/gia/committees/3/revoke", nil, ""},
		{"CommitteeFilters", func(c *myitmo.Client) error { _, err := c.GIA.CommitteeFilters(ctx, myitmo.GIARoleAdmin); return err },
			"GET", "/api/gia/committees/filters", url.Values{"show_as": {"admin"}}, ""},
		{"GeneralStateCommittees", func(c *myitmo.Client) error { _, err := c.GIA.GeneralStateCommittees(ctx, p); return err },
			"GET", "/api/gia/committees/general-state-committees", pq, ""},
		{"GeneralStateCommitteeFilters", func(c *myitmo.Client) error { _, err := c.GIA.GeneralStateCommitteeFilters(ctx); return err },
			"GET", "/api/gia/committees/general-state-committees/filters", nil, ""},
		{"CommitteeExperts", func(c *myitmo.Client) error { _, err := c.GIA.CommitteeExperts(ctx, "тест", 15); return err },
			"GET", "/api/gia/committees/experts", url.Values{"query": {"тест"}, "limit": {"15"}}, ""},
		{"CommitteeChairmen", func(c *myitmo.Client) error { _, err := c.GIA.CommitteeChairmen(ctx, "", 15); return err },
			"GET", "/api/gia/committees/chairmen", url.Values{"limit": {"15"}}, ""},
		{"CommitteePrograms", func(c *myitmo.Client) error { _, err := c.GIA.CommitteePrograms(ctx, "", 15, 5); return err },
			"GET", "/api/gia/committees/edu-programs", url.Values{"limit": {"15"}, "chairman_expert_id": {"5"}}, ""},
		{"CommitteeDirections", func(c *myitmo.Client) error {
			_, err := c.GIA.CommitteeDirections(ctx, myitmo.GIACommitteeDirectionsParams{Limit: 15, ChairmanExpertID: 5, EduProgramIDs: []int64{7, 11}})
			return err
		}, "GET", "/api/gia/committees/edu-directions", url.Values{"limit": {"15"}, "chairman_expert_id": {"5"}, "edu_program_ids": {"7,11"}}, ""},
	})
}

func TestGIACommitteeDecodes(t *testing.T) {
	f, c := newFake(t)
	f.result(`{"data":[{"id":1,"ep_name":"Программа","ep_enrollment_year":2021,"committees":[{"committee_id":3,
		"committee_number":"12-А","chairman_surname":"Тестов","edu_directions":[{"dir_id":8,"dir_code":"09.03.01","dir_name":"Информатика"}],
		"created_at":"2025-02-01T12:00:00","status_id":4,"status_name":"Утвержден"}]}],"total":1}`)
	list := must[*myitmo.GIACommitteeList](t)(c.GIA.Committees(ctx, myitmo.GIACommitteesParams{}))
	if list.Total != 1 || list.Data[0].EPName != "Программа" || list.Data[0].Committees[0].EduDirections[0].DirCode != "09.03.01" ||
		string(list.Data[0].Committees[0].CommitteeNumber) != `"12-А"` {
		t.Errorf("list = %+v", list)
	}

	f.result(`{"committee_id":3,"committee_number":12,"expert_id":5,"chairman_surname":"Тестов",
		"edu_programs":[{"ep_id":7,"ep_name":"Программа","ep_enrollment_year":2021}],
		"internal_experts":[{"expert_isu":100002,"expert_surname":"Сотрудников"}],
		"external_experts":[{"expert_id":6,"expert_surname":"Внешний","expert_work_place":"ООО Тест"}],
		"status_id":3,"status_name":"Отклонен","comment":"мало членов","commenter_fio":"Проверяющий","commenter_role":"osop"}`)
	cm := must[*myitmo.GIACommittee](t)(c.GIA.Committee(ctx, 3))
	if cm.InternalExperts[0].External() || !cm.ExternalExperts[0].External() || cm.EduPrograms[0].EPID != 7 || cm.CommenterRole != "osop" {
		t.Errorf("committee = %+v", cm)
	}

	f.result(`{"year_filters":[{"year":2025}],"status_filters":[{"id":1,"name":"Черновик"}],"faculty_filters":[{"id":2,"value":"ФИТ"}],
		"role_filters":[{"key":"osop","name":"ОСОП"}]}`)
	fl := must[*myitmo.GIACommitteeFilters](t)(c.GIA.CommitteeFilters(ctx, ""))
	if fl.YearFilters[0].Year != 2025 || fl.RoleFilters[0].Key != myitmo.GIARoleOSOP || fl.FacultyFilters[0].Value != "ФИТ" {
		t.Errorf("filters = %+v", fl)
	}

	f.result(`{"committees":[{"committee_id":4,"committee_number":1}],"total":1}`)
	gs := must[*myitmo.GIAGeneralStateCommitteeList](t)(c.GIA.GeneralStateCommittees(ctx, myitmo.GIACommitteesParams{}))
	if gs.Total != 1 || gs.Committees[0].CommitteeID != 4 {
		t.Errorf("general state = %+v", gs)
	}
}
