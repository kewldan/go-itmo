package myitmo_test

import (
	"net/url"
	"testing"

	"github.com/kewldan/go-itmo/myitmo"
)

func TestGIAFilterRequests(t *testing.T) {
	runGIACases(t, []giaCase{
		{"UserStatus", func(c *myitmo.Client) error { _, err := c.GIA.UserStatus(ctx); return err },
			"GET", "/api/gia/users/status", nil, ""},
		{"PeopleSearch", func(c *myitmo.Client) error { _, err := c.GIA.PeopleSearch(ctx, "Тестов"); return err },
			"GET", "/api/gia/people/search", url.Values{"query": {"Тестов"}}, ""},
		{"SetSNILS", func(c *myitmo.Client) error { return c.GIA.SetSNILS(ctx, 100001, "000-000-000 00") },
			"POST", "/api/gia/people/snils", nil, `{"isu":100001,"snils":"000-000-000 00"}`},
		{"EPStudents", func(c *myitmo.Client) error { _, err := c.GIA.EPStudents(ctx, 7, "Z3400"); return err },
			"GET", "/api/gia/ep/students", url.Values{"ep_id": {"7"}, "group_id": {"Z3400"}}, ""},
		{"EPGroups", func(c *myitmo.Client) error { _, err := c.GIA.EPGroups(ctx, 7); return err },
			"GET", "/api/gia/ep/7/groups", nil, ""},
		{"Years", func(c *myitmo.Client) error { _, err := c.GIA.Years(ctx); return err },
			"GET", "/api/gia/filters/years", nil, ""},
		{"FilterPrograms", func(c *myitmo.Client) error { _, err := c.GIA.FilterPrograms(ctx, myitmo.GIARoleSecretary); return err },
			"GET", "/api/gia/filters/ep", url.Values{"show_as": {"secretary"}}, ""},
		{"FilterGroups", func(c *myitmo.Client) error { _, err := c.GIA.FilterGroups(ctx, ""); return err },
			"GET", "/api/gia/filters/groups", nil, ""},
		{"FilterDirections", func(c *myitmo.Client) error {
			_, err := c.GIA.FilterDirections(ctx, 7, myitmo.GIARoleEPManager)
			return err
		}, "GET", "/api/gia/filters/directions", url.Values{"ep_id": {"7"}, "show_as": {"ep_manager"}}, ""},
		{"FilterSecretaries", func(c *myitmo.Client) error { _, err := c.GIA.FilterSecretaries(ctx, myitmo.GIARoleOSOP); return err },
			"GET", "/api/gia/filters/secretaries", url.Values{"show_as": {"osop"}}, ""},
		{"FilterStudents", func(c *myitmo.Client) error { _, err := c.GIA.FilterStudents(ctx); return err },
			"GET", "/api/gia/filters/students", nil, ""},
		{"Employees", func(c *myitmo.Client) error { _, err := c.GIA.Employees(ctx, "тест"); return err },
			"GET", "/api/gia/filters/employed", url.Values{"query": {"тест"}}, ""},
		{"Statuses", func(c *myitmo.Client) error {
			_, err := c.GIA.Statuses(ctx, myitmo.GIAEntityReservations, myitmo.GIARoleAdmin)
			return err
		}, "GET", "/api/gia/filters/statuses", url.Values{"entity": {"reservations"}, "show_as": {"admin"}}, ""},
		{"FilterExperts", func(c *myitmo.Client) error { _, err := c.GIA.FilterExperts(ctx); return err },
			"GET", "/api/gia/filters/experts", nil, ""},
		{"FilterFaculties", func(c *myitmo.Client) error { _, err := c.GIA.FilterFaculties(ctx); return err },
			"GET", "/api/gia/filters/faculties", nil, ""},
		{"FilterEduLevels", func(c *myitmo.Client) error { _, err := c.GIA.FilterEduLevels(ctx); return err },
			"GET", "/api/gia/filters/edu-levels", nil, ""},
		{"OrderTypes", func(c *myitmo.Client) error { _, err := c.GIA.OrderTypes(ctx); return err },
			"GET", "/api/gia/filters/order-types", nil, ""},
		{"ChairmanFaculties", func(c *myitmo.Client) error { _, err := c.GIA.ChairmanFaculties(ctx); return err },
			"GET", "/api/gia/filters/chairman/faculties", nil, ""},
		{"ChairmanDirections", func(c *myitmo.Client) error { _, err := c.GIA.ChairmanDirections(ctx); return err },
			"GET", "/api/gia/filters/chairman/dir", nil, ""},
		{"ChairmanPrograms", func(c *myitmo.Client) error { _, err := c.GIA.ChairmanPrograms(ctx); return err },
			"GET", "/api/gia/filters/chairman/ep", nil, ""},
		{"AttachmentCredits", func(c *myitmo.Client) error { _, err := c.GIA.AttachmentCredits(ctx); return err },
			"GET", "/api/gia/filters/attachment/credits", nil, ""},
		{"AcademicTitles", func(c *myitmo.Client) error { _, err := c.GIA.AcademicTitles(ctx); return err },
			"GET", "/api/gia/references/academic-titles", nil, ""},
		{"DegreeTypes", func(c *myitmo.Client) error { _, err := c.GIA.DegreeTypes(ctx); return err },
			"GET", "/api/gia/references/degree-types", nil, ""},
		{"ExpertEduLevels", func(c *myitmo.Client) error { _, err := c.GIA.ExpertEduLevels(ctx); return err },
			"GET", "/api/gia/references/edu-levels", nil, ""},
		{"Grades", func(c *myitmo.Client) error { _, err := c.GIA.Grades(ctx, true); return err },
			"GET", "/api/gia/references/grades", url.Values{"assessment": {"true"}}, ""},
		{"OrganizationRegistries", func(c *myitmo.Client) error { _, err := c.GIA.OrganizationRegistries(ctx); return err },
			"GET", "/api/gia/references/organization-registries", nil, ""},
		{"OrganizationStatuses", func(c *myitmo.Client) error { _, err := c.GIA.OrganizationStatuses(ctx); return err },
			"GET", "/api/gia/references/organization-statuses", nil, ""},
	})
}

func TestGIAUserStatusDecode(t *testing.T) {
	f, c := newFake(t)
	f.result(`{"is_admin":false,"is_root":false,"is_osop":true,"is_general_state_coordinator":false,
		"is_general_state_secretary":false,"is_supervisor":true,"is_ep_manager":false,"is_faculty_manager":false,
		"is_deputy_faculty_manager":false,"is_secretary":true,"is_reviewer":false,"is_manager":false}`)
	st := must[*myitmo.GIAUserStatus](t)(c.GIA.UserStatus(ctx))
	if !st.IsOSOP || !st.IsSupervisor || !st.IsSecretary || st.IsAdmin {
		t.Errorf("status = %+v", st)
	}
}

func TestGIAFilterDecodes(t *testing.T) {
	f, c := newFake(t)
	f.result(`[{"value":"Z3400"},"Z3401"]`)
	groups := must[[]myitmo.GIAGroupValue](t)(c.GIA.FilterGroups(ctx, ""))
	if len(groups) != 2 || groups[0].Value != "Z3400" || groups[1].Value != "Z3401" {
		t.Errorf("groups = %+v", groups)
	}

	f.result(`{"credits":[240,300.5]}`)
	credits := must[[]float64](t)(c.GIA.AttachmentCredits(ctx))
	if len(credits) != 2 || credits[1] != 300.5 {
		t.Errorf("credits = %v", credits)
	}

	f.result(`[{"status_id":14,"status_name":"Забронировано","status_color":"#E9F7E8","color":null}]`)
	st := must[[]myitmo.GIAStatus](t)(c.GIA.Statuses(ctx, myitmo.GIAEntityReservations, ""))
	if st[0].StatusID != myitmo.GIAReservationBooked || st[0].StatusColor != "#E9F7E8" {
		t.Errorf("statuses = %+v", st)
	}

	f.result(`[{"value":"2024/2025"},{"id":3,"year":2026,"value":2026}]`)
	years := must[[]myitmo.GIAYearOption](t)(c.GIA.Years(ctx))
	if string(years[0].Value) != `"2024/2025"` || years[1].Year != 2026 {
		t.Errorf("years = %+v", years)
	}

	f.result(`[{"id":1,"name_ru":"отлично","number":5}]`)
	grades := must[[]myitmo.GIAGrade](t)(c.GIA.Grades(ctx, true))
	if grades[0].Number != 5 || grades[0].NameRU != "отлично" {
		t.Errorf("grades = %+v", grades)
	}

	f.result(`[{"ep_id":7,"ep_name":"Тестовая программа","ep_enrollment_year":2022,"groups":[{"group_id":"Z3400",
		"is_secretary_single":true,"students_count":1,"students_data":[{"student_id":11,"student_isu":100001,
		"student_surname":"Тестов","student_name":"Тест","student_second_name":"","secretary_isu":0}]}]}]`)
	eps := must[[]myitmo.GIAEPStudents](t)(c.GIA.EPStudents(ctx, 7, ""))
	if eps[0].Groups[0].StudentsData[0].StudentISU != 100001 || !eps[0].Groups[0].IsSecretarySingle {
		t.Errorf("ep students = %+v", eps)
	}
}
