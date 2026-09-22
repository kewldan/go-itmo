package myitmo_test

import (
	"net/http"
	"testing"

	"github.com/kewldan/go-itmo/myitmo"
)

func TestIndividualPlanRoutes(t *testing.T) {
	const base = "/api/individual-plan/plan/77"
	const disc = base + "/module/3/discipline/100"
	module := myitmo.IndividualPlanModuleInput{
		NameRu: "Модуль", NameEn: "Module", RuleTypeID: 1, RuleValue: []int{1, 2},
		ParentID: 2, ParentName: "Parent", BlockID: 1, RuleTypeName: "choose",
	}
	moduleJSON := `{"name_ru":"Модуль","name_en":"Module","rule_type_id":1,"rule_value":[1,2],"parent_id":2,"parent_name":"Parent","block_id":1,"rule_type_name":"choose"}`
	noRule := module
	noRule.RuleValue = nil
	noRuleJSON := `{"name_ru":"Модуль","name_en":"Module","rule_type_id":1,"rule_value":null,"parent_id":2,"parent_name":"Parent","block_id":1,"rule_type_name":"choose"}`

	runElRoutes(t, []elRoute{
		{"Role", func(c *myitmo.Client) error { return elDiscard(c.IndividualPlan.Role(ctx)) },
			http.MethodGet, "/api/individual-plan/plan/role", "", ""},
		{"Programs", func(c *myitmo.Client) error { return elDiscard(c.IndividualPlan.Programs(ctx)) },
			http.MethodGet, "/api/individual-plan/plan/programs", "", ""},
		{"Directions", func(c *myitmo.Client) error { return elDiscard(c.IndividualPlan.Directions(ctx)) },
			http.MethodGet, "/api/individual-plan/plan/directions", "", ""},
		{"Implementers", func(c *myitmo.Client) error { return elDiscard(c.IndividualPlan.Implementers(ctx)) },
			http.MethodGet, "/api/individual-plan/plan/implementers", "", ""},
		{"Statuses", func(c *myitmo.Client) error { return elDiscard(c.IndividualPlan.Statuses(ctx)) },
			http.MethodGet, "/api/individual-plan/plan/statuses", "", ""},
		{"CurrentSemester", func(c *myitmo.Client) error { return elDiscard(c.IndividualPlan.CurrentSemester(ctx)) },
			http.MethodGet, "/api/individual-plan/semesters/current", "", ""},
		{"Flags", func(c *myitmo.Client) error { return elDiscard(c.IndividualPlan.Flags(ctx)) },
			http.MethodGet, "/api/individual-plan/flags", "", ""},
		{"ListDefault", func(c *myitmo.Client) error {
			return elDiscard(c.IndividualPlan.List(ctx, myitmo.IndividualPlanListParams{}))
		}, http.MethodGet, "/api/individual-plan/plan", "archive=0", ""},
		{"ListFiltered", func(c *myitmo.Client) error {
			return elDiscard(c.IndividualPlan.List(ctx, myitmo.IndividualPlanListParams{
				Status: myitmo.Ptr[int64](2), EPID: myitmo.Ptr[int64](5), DirectionID: myitmo.Ptr[int64](6),
				ImplementerID: myitmo.Ptr[int64](8), Archive: true, Query: "abc",
				ErrorFilter: myitmo.IndividualPlanChoicePastErrors, Limit: 15, Offset: 30,
			}))
		}, http.MethodGet, "/api/individual-plan/plan",
			"archive=1&choice_past_errors=1&direction_id=6&ep_id=5&implementer_id=8&limit=15&offset=30&query=abc&status=2", ""},
		{"Create", func(c *myitmo.Client) error {
			return elDiscard(c.IndividualPlan.Create(ctx, myitmo.IndividualPlanCreate{StartYear: 1, EPID: 5, DirectionID: 6, ImplementerID: 8}))
		}, http.MethodPost, "/api/individual-plan/plan", "",
			`{"start_year":1,"ep_id":5,"direction_id":6,"implementer_id":8,"isu":[],"manual":false}`},
		{"Get", func(c *myitmo.Client) error { return elDiscard(c.IndividualPlan.Get(ctx, 77)) },
			http.MethodGet, base, "", ""},
		{"Content", func(c *myitmo.Client) error { return elDiscard(c.IndividualPlan.Content(ctx, 77)) },
			http.MethodGet, base + "/plan", "", ""},
		{"Draft", func(c *myitmo.Client) error { return elDiscard(c.IndividualPlan.Draft(ctx, 77)) },
			http.MethodGet, base + "/draft", "", ""},
		{"CreateDraft", func(c *myitmo.Client) error { return c.IndividualPlan.CreateDraft(ctx, 77) },
			http.MethodPost, base + "/draft", "", ""},
		{"DeleteDraft", func(c *myitmo.Client) error { return c.IndividualPlan.DeleteDraft(ctx, 77) },
			http.MethodDelete, base + "/draft", "", ""},
		{"PublishDraft", func(c *myitmo.Client) error { return c.IndividualPlan.PublishDraft(ctx, 77) },
			http.MethodPost, base + "/draft/save", "", ""},
		{"Changes", func(c *myitmo.Client) error {
			return elDiscard(c.IndividualPlan.Changes(ctx, 77, myitmo.IndividualPlanChangesParams{Category: "choice", Query: "x", Limit: 10, Offset: 20}))
		}, http.MethodGet, base + "/changes", "category=choice&limit=10&offset=20&query=x", ""},
		{"Semesters", func(c *myitmo.Client) error { return elDiscard(c.IndividualPlan.Semesters(ctx, 77)) },
			http.MethodGet, base + "/semesters", "", ""},
		{"AvailableDisciplines", func(c *myitmo.Client) error {
			return elDiscard(c.IndividualPlan.AvailableDisciplines(ctx, 77, "math", []int64{1, 2}))
		}, http.MethodGet, base + "/availableDisciplines", "program_type_id=1&program_type_id=2&query=math", ""},
		{"DisciplineUse", func(c *myitmo.Client) error { return elDiscard(c.IndividualPlan.DisciplineUse(ctx, 77, 100, true)) },
			http.MethodGet, base + "/discipline/100/use", "draft=true", ""},
		{"UpperModules", func(c *myitmo.Client) error { return elDiscard(c.IndividualPlan.UpperModules(ctx, 77)) },
			http.MethodGet, base + "/upperModules", "", ""},
		{"LowerModules", func(c *myitmo.Client) error { return elDiscard(c.IndividualPlan.LowerModules(ctx, 77)) },
			http.MethodGet, base + "/lowerModules", "", ""},
		{"AddChoice", func(c *myitmo.Client) error {
			return c.IndividualPlan.AddChoice(ctx, 77, 3, 100, []int64{202620271})
		}, http.MethodPost, disc + "/choice", "", `[202620271]`},
		{"UpdateChoice", func(c *myitmo.Client) error {
			return c.IndividualPlan.UpdateChoice(ctx, 77, 3, 100, []int64{202620271, 202620272})
		}, http.MethodPatch, disc + "/choice", "", `[202620271,202620272]`},
		{"DeleteChoice", func(c *myitmo.Client) error { return c.IndividualPlan.DeleteChoice(ctx, 77, 3, 100) },
			http.MethodDelete, disc + "/choice", "", ""},
		{"CreateModule", func(c *myitmo.Client) error { return elDiscard(c.IndividualPlan.CreateModule(ctx, 77, module)) },
			http.MethodPost, base + "/module", "", moduleJSON},
		{"UpdateModule", func(c *myitmo.Client) error { return elDiscard(c.IndividualPlan.UpdateModule(ctx, 77, 3, noRule)) },
			http.MethodPatch, base + "/module/3", "", noRuleJSON},
		{"DeleteModule", func(c *myitmo.Client) error { return c.IndividualPlan.DeleteModule(ctx, 77, 3) },
			http.MethodDelete, base + "/module/3", "", ""},
		{"AddDisciplines", func(c *myitmo.Client) error { return c.IndividualPlan.AddDisciplines(ctx, 77, 3, []int64{100, 101}) },
			http.MethodPost, base + "/module/3/discipline", "", `[100,101]`},
		{"DeleteDiscipline", func(c *myitmo.Client) error { return c.IndividualPlan.DeleteDiscipline(ctx, 77, 3, 100) },
			http.MethodDelete, disc, "", ""},
		{"SetDisciplineSemesters", func(c *myitmo.Client) error {
			return c.IndividualPlan.SetDisciplineSemesters(ctx, 77, 3, 100, []int{3})
		}, http.MethodPatch, disc, "", `[3]`},
		{"MoveDiscipline", func(c *myitmo.Client) error { return c.IndividualPlan.MoveDiscipline(ctx, 77, 3, 100, 4) },
			http.MethodPatch, disc + "/move", "", `{"module_id":4}`},
		{"ToggleFlowChoice", func(c *myitmo.Client) error { return c.IndividualPlan.ToggleFlowChoice(ctx, 77, 3, 100) },
			http.MethodPatch, disc + "/flowChoice", "", ""},
		{"ValidatePlanContent", func(c *myitmo.Client) error { return elDiscard(c.IndividualPlan.ValidatePlanContent(ctx, 77, true)) },
			http.MethodPost, base + "/plan/validate/plan-content", "draft=true", ""},
		{"ValidatePersonalPlanContent", func(c *myitmo.Client) error {
			return elDiscard(c.IndividualPlan.ValidatePersonalPlanContent(ctx, 77, false))
		}, http.MethodPost, base + "/plan/validate/personal-plan-content", "draft=false", ""},
		{"ValidateSelection", func(c *myitmo.Client) error {
			return elDiscard(c.IndividualPlan.ValidateSelection(ctx, 77, false, false))
		}, http.MethodPost, base + "/plan/validate/selection", "consider_stated_semester=false&draft=false", ""},
	})
}

func TestIndividualPlanListDecode(t *testing.T) {
	f, c := newFake(t)
	f.result(`{"forbidden":false,"can_create":true,"is_admin":false,"edu_levels":[1,2]}`)
	role := must[*myitmo.IndividualPlanRole](t)(c.IndividualPlan.Role(ctx))
	if role.Forbidden || !role.CanCreate || len(role.EduLevels) != 2 {
		t.Errorf("role = %+v", role)
	}

	f.result(`{"sem_id":202620271}`)
	if id := must[int64](t)(c.IndividualPlan.CurrentSemester(ctx)); id != 202620271 {
		t.Errorf("current semester = %d", id)
	}

	f.result(`[{"key":"past_choice","enabled":true}]`)
	if fl := must[[]myitmo.IndividualPlanFlag](t)(c.IndividualPlan.Flags(ctx)); fl[0].Key != myitmo.IndividualPlanFlagPastChoice || !fl[0].Enabled {
		t.Errorf("flags = %+v", fl)
	}

	f.result(`{"count":2,"items":[
		{"id":77,"isu":100000,"fio":"Student A","ep_direction":{"code":"01.03.02","name":"Direction"},
			"ep":{"id":5,"name":"Programme","enrollment_year":2024},"implementer":{"short_name":"DEP"},
			"date_start":"2024-09-01","date_end":null,"status":{"short_name":"active"},"has_draft":true},
		{"id":78,"isu":null,"fio":null,"ep_direction":{"code":"01.03.02","name":"Direction"},
			"ep":{"id":5,"name":"Programme","enrollment_year":"2024"},"implementer":{"short_name":"DEP"},
			"date_start":null,"date_end":null,"status":null,"has_draft":false}]}`)
	l := must[*myitmo.IndividualPlanList](t)(c.IndividualPlan.List(ctx, myitmo.IndividualPlanListParams{}))
	a, b := l.Items[0], l.Items[1]
	if l.Count != 2 || *a.ISU != 100000 || *a.DateStart != myitmo.NewDate(2024, 9, 1) || a.DateEnd != nil || a.Status.ShortName != "active" || !a.HasDraft {
		t.Errorf("first = %+v", a)
	}
	if b.ISU != nil || b.Status != nil || string(b.EP.EnrollmentYear) != `"2024"` {
		t.Errorf("second = %+v", b)
	}

	f.result(`[{"id":77}]`)
	if cr := must[[]myitmo.IndividualPlanCreated](t)(c.IndividualPlan.Create(ctx, myitmo.IndividualPlanCreate{})); cr[0].ID != 77 {
		t.Errorf("created = %+v", cr)
	}

	f.result(`{"id":77,"isu":100000,"fio":"Student A","status":{"id":2,"name":"active"},"date_start":"2024-09-01","date_end":"2028-06-30",
		"implementer":{"short_name":"DEP"},"group":"G0000","ep":{"id":5,"name":"Programme","enrollment_year":2024},
		"ep_direction":{"code":"01.03.02","name":"Direction"},"edu_standard":{"name":"Standard"},"can_edit":true}`)
	info := must[*myitmo.IndividualPlanInfo](t)(c.IndividualPlan.Get(ctx, 77))
	if info.Status.ID != myitmo.IndividualPlanStatusActive || info.EduStandard.Name != "Standard" || !info.CanEdit || info.DateEnd.Year != 2028 {
		t.Errorf("info = %+v", info)
	}
}

func TestIndividualPlanContentDecode(t *testing.T) {
	f, c := newFake(t)
	f.result(`{"plan":[{"id":1,"name":"Root","name_ru":"Root","name_en":"Root","type":"module","is_root":true,"block_id":1,"parent_id":null,
		"rule":{"type_id":2,"type_name":"all","rule_value":null},"min_capacity":0,"max_capacity":240,"can_delete":false,
		"children_types":[1,2],"rule_types":[1,2,21],
		"children":[{"id":3,"name":"Module","name_ru":"Module","name_en":"Module","type":"module","is_root":false,"block_id":1,"parent_id":1,
			"rule":{"type_id":1,"type_name":"choose","rule_value":[1]},"min_capacity":3,"max_capacity":6,"target_capacity":null,"can_delete":true,
			"children":[{"id":100,"name":"Discipline A","type":"discipline","module_id":3,"is_root":false,"status":{"id":6,"name":"approved"},
				"implementer":{"name":"Department","short_name":"DEP"},"capacity":3,"semesters_count":1,"semesters":[5],"no_flow_choice":false,
				"can_delete":true,"grades":[{"semester_id":202620271,"grade":5,"letter":"A","work_type_name":"exam"}],
				"selection":{"semesters":[202620271],"protocol":false,"protocol_semesters":[]},
				"program_contents":[{"id":11,"content_order":1,"capacity":3,"program_work_types":[
					{"id":1,"work_type_id":128,"work_type_name":"contact","hours":48,"is_control":false},
					{"id":2,"work_type_id":22,"work_type_name":"self-study","hours":60,"is_control":false}]}]}]}]}],
		"semesters_count":8,"updated_at":"2026-09-01T12:00:00+03:00","filled_min_capacity":3,"filled_max_capacity":6,"full_capacity":240,
		"has_draft":true,"changed_choice":true}`)
	pc := must[*myitmo.IndividualPlanContent](t)(c.IndividualPlan.Draft(ctx, 77))
	root := pc.Plan[0]
	if !root.IsRoot || root.ParentID != nil || root.Rule.RuleValue != nil || len(root.ChildrenTypes) != 2 || !pc.ChangedChoice || *pc.FullCapacity != 240 {
		t.Fatalf("root = %+v", root)
	}
	m := root.Children[0]
	if *m.ParentID != 1 || m.Rule.RuleValue[0] != 1 || m.TargetCapacity != nil {
		t.Errorf("module = %+v", m)
	}
	d := m.Children[0]
	wt := d.ProgramContents[0].ProgramWorkTypes
	if d.ModuleID != 3 || d.Implementer.ShortName != "DEP" || string(d.Grades[0].Grade) != "5" || d.Selection.Semesters[0] != 202620271 ||
		wt[0].WorkTypeID != myitmo.IndividualPlanWorkTypeContact || wt[1].Hours != 60 {
		t.Errorf("discipline = %+v", d)
	}
}

func TestIndividualPlanChangesDecode(t *testing.T) {
	f, c := newFake(t)
	f.result(`{"count":2,"items":[
		{"id":1,"date":"2026-09-01T12:00:00+03:00","category":"plan","entity_type":"module","summary":"changed",
			"author":{"id":9,"fio":"Employee A"},"discipline_id":null,"discipline_name":null,"module_id":3,"module_name":"Module",
			"description":{"before":{"name":"Old","name_en":"Old","rule":{"name":"choose","value":[1]},"parent_module":{"id":1,"name":"Root","rule_name":"all","rule_value":null}},
				"after":{"name":"New","name_en":"New","rule":{"name":"choose","value":[2]},"parent_module":{"id":1,"name":"Root","rule_name":"all","rule_value":null}}}},
		{"id":2,"date":"2026-09-02T12:00:00+03:00","category":"choice","entity_type":"choice","summary":"added",
			"author":null,"discipline_id":100,"discipline_name":"Discipline A","module_id":3,"module_name":"Module",
			"description":{"before":null,"after":{"semesters":[202620271],"block_name":"Block"}}}]}`)
	ch := must[*myitmo.IndividualPlanChanges](t)(c.IndividualPlan.Changes(ctx, 77, myitmo.IndividualPlanChangesParams{}))
	a, b := ch.Items[0], ch.Items[1]
	if a.Author.Fio != "Employee A" || a.DisciplineID != nil || a.Description.Before.Rule.Value[0] != 1 || a.Description.After.ParentModule.ID != 1 {
		t.Errorf("first = %+v", a)
	}
	if b.Author != nil || *b.DisciplineID != 100 || b.Description.Before != nil || b.Description.After.BlockName != "Block" || b.Description.After.Semesters[0] != 202620271 {
		t.Errorf("second = %+v", b)
	}
}

func TestIndividualPlanLookupsDecode(t *testing.T) {
	f, c := newFake(t)
	f.result(`{"count":1,"disciplines":[{"id":100,"name":"Discipline A","capacity":3,"implementer_short_name":"DEP","semesters_count":1,
		"status_name":"approved","annotation_ru":null,"program_contents":[]}]}`)
	s := must[*myitmo.IndividualPlanDisciplineSearch](t)(c.IndividualPlan.AvailableDisciplines(ctx, 77, "", nil))
	if s.Count != 1 || s.Disciplines[0].ImplementerShortName != "DEP" {
		t.Errorf("search = %+v", s)
	}

	f.result(`[{"id":3,"name_ru":"Module","rule_type_name":"choose","rule_value":[1]}]`)
	if u := must[[]myitmo.IndividualPlanModuleUse](t)(c.IndividualPlan.DisciplineUse(ctx, 77, 100, false)); u[0].RuleValue[0] != 1 {
		t.Errorf("use = %+v", u)
	}

	f.result(`[{"id":3,"block_id":1,"name_ru":"Module","name_en":"Module","parent_id":1,"parent_name":"Root","rule_type_id":1,"rule_type_name":"choose","rule_value":[1]}]`)
	if m := must[[]myitmo.IndividualPlanModuleRef](t)(c.IndividualPlan.LowerModules(ctx, 77)); *m[0].ParentID != 1 || m[0].ParentName != "Root" {
		t.Errorf("modules = %+v", m)
	}

	f.result(`[202420251,202420252]`)
	if sem := must[[]int64](t)(c.IndividualPlan.Semesters(ctx, 77)); len(sem) != 2 {
		t.Errorf("semesters = %v", sem)
	}

	f.result(`[{"id":5,"name":"Programme","enrollment_year":2024}]`)
	if p := must[[]myitmo.IndividualPlanProgram](t)(c.IndividualPlan.Programs(ctx)); string(p[0].EnrollmentYear) != "2024" {
		t.Errorf("programs = %+v", p)
	}

	f.result(`[{"id":6,"code":"01.03.02","name":"Direction"}]`)
	if d := must[[]myitmo.IndividualPlanDirection](t)(c.IndividualPlan.Directions(ctx)); d[0].Code != "01.03.02" {
		t.Errorf("directions = %+v", d)
	}

	f.result(`17`)
	if id := must[int64](t)(c.IndividualPlan.CreateModule(ctx, 77, myitmo.IndividualPlanModuleInput{})); id != 17 {
		t.Errorf("module id = %d", id)
	}
}

func TestIndividualPlanValidationDecode(t *testing.T) {
	f, c := newFake(t)
	f.result(`{"status":"error","errors":[{"module_id":3,"module_name":"Module","discipline_id":null,"discipline_name":null,
		"errors":[{"ru":"not enough"}]}],"impossible_modules":[3]}`)
	v := must[*myitmo.IndividualPlanValidation](t)(c.IndividualPlan.ValidateSelection(ctx, 77, true, true))
	if v.Status != myitmo.IndividualPlanValidationStatusError || *v.Errors[0].ModuleID != 3 || v.Errors[0].DisciplineID != nil ||
		v.Errors[0].Errors[0].Ru != "not enough" || v.ImpossibleModules[0] != 3 {
		t.Errorf("validation = %+v", v)
	}
}
