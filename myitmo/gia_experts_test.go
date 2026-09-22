package myitmo_test

import (
	"net/url"
	"testing"
	"time"

	"github.com/kewldan/go-itmo/myitmo"
)

func TestGIAExpertRequests(t *testing.T) {
	in := myitmo.GIAExpertInput{
		Name: "Тест", Surname: "Тестов", Email: "expert@example.com", Gender: myitmo.Ptr("male"),
		AcademicTitleID: myitmo.Ptr[int64](2),
		Degrees:         []myitmo.GIAExpertDegree{{DegreeTypeID: 1, Number: "1", Series: "A", Date: time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC)}},
		WorkPlaces:      []myitmo.GIAExpertWorkPlace{{OrganizationRegistryID: 1, INN: 7700000000, OrganizationName: "ООО Тест", JobTitle: "инженер", OrganizationStatusID: 3, Stages: myitmo.RawJSON(`{"blocks":[]}`)}},
		StatusID:        myitmo.GIAExpertStatusDraft,
	}
	inJSON := `{"name":"Тест","surname":"Тестов","second_name":"","email":"expert@example.com","phone_number":"",
		"gender":"male","education":"","speciality":"","academic_title_id":2,"edu_level_id":null,
		"degrees":[{"degree_type_id":1,"number":"1","series":"A","date":"2020-01-02T00:00:00Z"}],
		"work_places":[{"organization_registry_id":1,"inn":7700000000,"organization_name":"ООО Тест","line_of_business":"",
		"job_title":"инженер","organization_status_id":3,"stages":{"blocks":[]}}],"status_id":%d}`
	chair := myitmo.GIAChairmanInput{ExpertID: 5, EPID: 7, DirID: []int64{8, 9}, WorkPlaceID: 10, StatusID: myitmo.GIAChairmanStatusOnApproval}
	chairJSON := `{"expert_id":5,"ep_id":7,"dir_id":[8,9],"work_place_id":10,"status_id":2}`
	cp := myitmo.GIAChairmenParams{Limit: 15, Offset: 30, Year: 2025, EPID: 7, DirID: 8, FacultyID: 3, PersonID: 5, StatusID: 2, Query: "тест"}
	cpQuery := url.Values{"limit": {"15"}, "offset": {"30"}, "year": {"2025"}, "ep_id": {"7"}, "dir_id": {"8"},
		"faculty_id": {"3"}, "person_id": {"5"}, "status_id": {"2"}, "query": {"тест"}}
	runGIACases(t, []giaCase{
		{"Experts", func(c *myitmo.Client) error {
			_, err := c.GIA.Experts(ctx, myitmo.GIAExpertsParams{Query: "тест", Limit: 25, Status: myitmo.GIAExpertStatusApproved})
			return err
		}, "GET", "/api/gia/experts", url.Values{"query": {"тест"}, "limit": {"25"}, "offset": {"0"}, "status": {"4"}}, ""},
		{"CreateExpert", func(c *myitmo.Client) error { _, err := c.GIA.CreateExpert(ctx, in); return err },
			"POST", "/api/gia/experts", nil, giaf(inJSON, 1)},
		{"Expert", func(c *myitmo.Client) error { _, err := c.GIA.Expert(ctx, 5); return err },
			"GET", "/api/gia/experts/5", nil, ""},
		{"UpdateExpert", func(c *myitmo.Client) error { return c.GIA.UpdateExpert(ctx, 5, in) },
			"PATCH", "/api/gia/experts/5", nil, giaf(inJSON, 1)},
		{"DeleteExpert", func(c *myitmo.Client) error { return c.GIA.DeleteExpert(ctx, 5) },
			"DELETE", "/api/gia/experts/5", nil, ""},
		{"ExpertExists", func(c *myitmo.Client) error { _, err := c.GIA.ExpertExists(ctx, "expert@example.com"); return err },
			"GET", "/api/gia/experts/exists", url.Values{"email": {"expert@example.com"}}, ""},
		{"Organization", func(c *myitmo.Client) error { _, err := c.GIA.Organization(ctx, "7700000000"); return err },
			"GET", "/api/gia/experts/organization", url.Values{"inn": {"7700000000"}}, ""},
		{"ExpertsSimple", func(c *myitmo.Client) error { _, err := c.GIA.ExpertsSimple(ctx); return err },
			"GET", "/api/gia/experts/simple", nil, ""},
		{"ApproveExpert", func(c *myitmo.Client) error { return c.GIA.ApproveExpert(ctx, 5) },
			"POST", "/api/gia/experts/approve/5", nil, ""},
		{"RejectExpert", func(c *myitmo.Client) error { return c.GIA.RejectExpert(ctx, 5, "нет данных") },
			"DELETE", "/api/gia/experts/reject/5", nil, `{"message":"нет данных"}`},
		{"SubmitExpert", func(c *myitmo.Client) error {
			draft := in
			draft.StatusID = 0
			return c.GIA.SubmitExpert(ctx, 5, draft)
		}, "POST", "/api/gia/experts/under-approval/5", nil, giaf(inJSON, 2)},
		{"Chairmen", func(c *myitmo.Client) error { _, err := c.GIA.Chairmen(ctx, cp); return err },
			"GET", "/api/gia/chairman/list", cpQuery, ""},
		{"ChairmenSimple", func(c *myitmo.Client) error { _, err := c.GIA.ChairmenSimple(ctx); return err },
			"GET", "/api/gia/chairman/list/simple", nil, ""},
		{"ProgramsWithoutChairman", func(c *myitmo.Client) error { _, err := c.GIA.ProgramsWithoutChairman(ctx, cp); return err },
			"GET", "/api/gia/chairman/ep_missing", cpQuery, ""},
		{"CreateChairman", func(c *myitmo.Client) error { return c.GIA.CreateChairman(ctx, chair) },
			"POST", "/api/gia/chairman", nil, chairJSON},
		{"UpdateChairman", func(c *myitmo.Client) error { return c.GIA.UpdateChairman(ctx, 12, chair) },
			"PATCH", "/api/gia/chairman/12", nil, chairJSON},
		{"DeleteChairman", func(c *myitmo.Client) error { return c.GIA.DeleteChairman(ctx, 12) },
			"DELETE", "/api/gia/chairman/12", nil, ""},
		{"ApproveChairman", func(c *myitmo.Client) error { return c.GIA.ApproveChairman(ctx, 12) },
			"POST", "/api/gia/chairman/12/approve", nil, ""},
		{"DeclineChairman", func(c *myitmo.Client) error { return c.GIA.DeclineChairman(ctx, 12, "нет") },
			"POST", "/api/gia/chairman/12/decline", nil, `{"message":"нет"}`},
	})
	runGIADownloads(t, []giaDownload{
		{"ChairmenOrder", func(c *myitmo.Client) (*myitmo.File, error) { return c.GIA.ChairmenOrder(ctx) },
			"GET", "/api/gia/chairman/order", nil, ""},
		{"ChairmanNomination", func(c *myitmo.Client) (*myitmo.File, error) { return c.GIA.ChairmanNomination(ctx, 12) },
			"GET", "/api/gia/chairman/12/represent", nil, ""},
	})
}

func TestGIAExpertDecodes(t *testing.T) {
	f, c := newFake(t)
	f.result(`{"expert_id":42}`)
	if id := must[int64](t)(c.GIA.CreateExpert(ctx, myitmo.GIAExpertInput{})); id != 42 {
		t.Errorf("id = %d", id)
	}

	f.result(`{"experts":[{"id":5,"fio":"Тестов Тест","academic_title":"доцент","degree_type":"к.т.н.","job_title":"инженер",
		"work_place":"ООО Тест","status_id":4,"created_at":"2025-03-01T10:00:00","created_by":null,"created_by_fio":""}],"count":1}`)
	list := must[*myitmo.GIAExpertList](t)(c.GIA.Experts(ctx, myitmo.GIAExpertsParams{}))
	if list.Count != 1 || list.Experts[0].StatusID != myitmo.GIAExpertStatusApproved || list.Experts[0].CreatedBy != nil ||
		list.Experts[0].CreatedAt.Day() != 1 {
		t.Errorf("list = %+v", list)
	}

	f.result(`{"id":5,"name":"Тест","surname":"Тестов","gender":"female","academic_title_id":null,"edu_level_id":2,
		"degrees":[{"degree_type_id":1,"degree_type":"к.т.н.","number":"1","series":"A","date":"2020-01-02T00:00:00Z"}],
		"work_places":[{"organization_registry_id":2,"inn":null,"organization_name":"Test Ltd","organization_status_id":1,
		"organization_status":"частная","stages":{"time":1,"blocks":[{"type":"paragraph"}]}}],"status_id":3,"approved":false,
		"comment":"исправить","comment_author":"Проверяющий"}`)
	e := must[*myitmo.GIAExpert](t)(c.GIA.Expert(ctx, 5))
	if e.AcademicTitleID != nil || *e.EduLevelID != 2 || e.WorkPlaces[0].INN != 0 || len(e.WorkPlaces[0].Stages) == 0 ||
		e.Degrees[0].Date.Year() != 2020 || e.Approved == nil || *e.Approved {
		t.Errorf("expert = %+v", e)
	}

	f.result(`true`)
	if ok := must[bool](t)(c.GIA.ExpertExists(ctx, "x@example.com")); !ok {
		t.Error("exists = false")
	}
	f.result(`null`)
	if ok := must[bool](t)(c.GIA.ExpertExists(ctx, "x@example.com")); ok {
		t.Error("exists = true")
	}

	f.result(`{"Chairman":[{"chairman_id":12,"expert_id":5,"chairman_surname":"Тестов","ep_id":7,"ep_name":"Программа",
		"direction_id":8,"status_id":2,"status_name":"На согласовании","created_at":"2025-01-10T09:00:00Z",
		"academic_titles":[],"degrees":null,"is_general_state":false}],"total_count":1}`)
	chairs := must[*myitmo.GIAChairmanList](t)(c.GIA.Chairmen(ctx, myitmo.GIAChairmenParams{}))
	if chairs.TotalCount != 1 || chairs.Chairman[0].ChairmanID != 12 || chairs.Chairman[0].DirectionID != 8 {
		t.Errorf("chairmen = %+v", chairs)
	}

	f.result(`{"programs":[{"ep_id":7,"ep_name":"Программа","faculty_name":"Факультет","direction_code":"09.03.01","directions":[]}],"total_count":3}`)
	eps := must[*myitmo.GIAEPWithoutChairmanList](t)(c.GIA.ProgramsWithoutChairman(ctx, myitmo.GIAChairmenParams{}))
	if eps.TotalCount != 3 || eps.Programs[0].DirectionCode != "09.03.01" {
		t.Errorf("programs = %+v", eps)
	}

	f.result(`[{"expert_id":5,"job_titles":[{"work_place_id":10,"job_title":"инженер","stages":{"blocks":[]}}]}]`)
	simple := must[[]myitmo.GIAExpertSimple](t)(c.GIA.ChairmenSimple(ctx))
	if simple[0].JobTitles[0].WorkPlaceID != 10 {
		t.Errorf("simple = %+v", simple)
	}
}
