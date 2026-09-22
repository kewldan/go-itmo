package myitmo_test

import (
	"net/url"
	"testing"

	"github.com/kewldan/go-itmo/myitmo"
)

func TestConstructorEPBankRequests(t *testing.T) {
	ctx := t.Context()
	form := myitmo.EPBankModuleForm{NameRU: "Модуль", NameEN: "Module", BlockID: 1, StandardID: 7}
	formJSON := `{"name_ru":"Модуль","name_en":"Module","block_id":1,"standard_id":7}`
	sub := myitmo.EPModuleForm{NameRU: "Подмодуль", NameEN: "Sub", RuleTypeID: 41, RuleValue: myitmo.EPRuleValue{6}, ParentID: 50, BlockID: 1, StandardID: 7}
	subJSON := `{"name_ru":"Подмодуль","name_en":"Sub","rule_type_id":41,"rule_value":[6],"parent_id":50,"block_id":1,"standard_id":7}`

	runEPCases(t, []epCase{
		{"CreateBankModule", func(t *testing.T, c *myitmo.Client) error {
			return epErr(c.ConstructorEP.CreateBankModule(ctx, form))
		}, "POST", "/api/constructor-ep/bank/modules/create", nil, formJSON},
		{"BankModules", func(t *testing.T, c *myitmo.Client) error {
			return epErr(c.ConstructorEP.BankModules(ctx, myitmo.EPBankModuleListParams{
				Query: "мод", EducationLevelID: 1, StartYear: 2024, StatusID: myitmo.EPStatusSigned, BlockID: 2, ProgramID: 42, Limit: 20, Offset: 40,
			}))
		}, "GET", "/api/constructor-ep/bank/modules/list", url.Values{
			"query": {"мод"}, "education_level_id": {"1"}, "start_year": {"2024"}, "status_id": {"6"},
			"block_id": {"2"}, "program_id": {"42"}, "limit": {"20"}, "offset": {"40"},
		}, ""},
		{"BankModule", func(t *testing.T, c *myitmo.Client) error {
			return epErr(c.ConstructorEP.BankModule(ctx, 50))
		}, "GET", "/api/constructor-ep/bank/modules/50/info", nil, ""},
		{"BankModuleTree", func(t *testing.T, c *myitmo.Client) error {
			return epErr(c.ConstructorEP.BankModuleTree(ctx, 50))
		}, "GET", "/api/constructor-ep/bank/modules/50/tree", nil, ""},
		{"UpdateBankModule", func(t *testing.T, c *myitmo.Client) error {
			return c.ConstructorEP.UpdateBankModule(ctx, 50, form)
		}, "PATCH", "/api/constructor-ep/bank/modules/50", nil, formJSON},
		{"SetBankModuleDescription", func(t *testing.T, c *myitmo.Client) error {
			return c.ConstructorEP.SetBankModuleDescription(ctx, 50, "Описание")
		}, "PATCH", "/api/constructor-ep/bank/modules/50/description", nil, `{"description":"Описание"}`},
		{"SetBankModulePrograms", func(t *testing.T, c *myitmo.Client) error {
			return c.ConstructorEP.SetBankModulePrograms(ctx, 50, []int64{42, 43})
		}, "PATCH", "/api/constructor-ep/bank/modules/50/accessible_programs", nil, `{"accessible_programs":[42,43]}`},
		{"SetBankModuleProgramsAll", func(t *testing.T, c *myitmo.Client) error {
			return c.ConstructorEP.SetBankModulePrograms(ctx, 50, nil)
		}, "PATCH", "/api/constructor-ep/bank/modules/50/accessible_programs", nil, `{"accessible_programs":null}`},
		{"ApproveBankModule", func(t *testing.T, c *myitmo.Client) error {
			return c.ConstructorEP.ApproveBankModule(ctx, 50)
		}, "POST", "/api/constructor-ep/bank/modules/50/approve", nil, ""},
		{"RevertBankModule", func(t *testing.T, c *myitmo.Client) error {
			return c.ConstructorEP.RevertBankModule(ctx, 50)
		}, "POST", "/api/constructor-ep/bank/modules/50/revert", nil, ""},
		{"ArchiveBankModule", func(t *testing.T, c *myitmo.Client) error {
			return c.ConstructorEP.ArchiveBankModule(ctx, 50)
		}, "POST", "/api/constructor-ep/bank/modules/50/archive", nil, ""},
		{"AddBankModuleEditor", func(t *testing.T, c *myitmo.Client) error {
			return c.ConstructorEP.AddBankModuleEditor(ctx, 50, 100001)
		}, "POST", "/api/constructor-ep/bank/modules/50/100001", nil, ""},
		{"DeleteBankModuleEditor", func(t *testing.T, c *myitmo.Client) error {
			return c.ConstructorEP.DeleteBankModuleEditor(ctx, 50, 100001)
		}, "DELETE", "/api/constructor-ep/bank/modules/50/100001", nil, ""},
		{"CreateBankSubmodule", func(t *testing.T, c *myitmo.Client) error {
			return epErr(c.ConstructorEP.CreateBankSubmodule(ctx, 50, sub))
		}, "POST", "/api/constructor-ep/bank/modules/50/tree/module", nil, subJSON},
		{"UpdateBankSubmodule", func(t *testing.T, c *myitmo.Client) error {
			return epErr(c.ConstructorEP.UpdateBankSubmodule(ctx, 50, 51, sub))
		}, "PATCH", "/api/constructor-ep/bank/modules/50/tree/module/51", nil, subJSON},
		{"DeleteBankSubmodule", func(t *testing.T, c *myitmo.Client) error {
			return c.ConstructorEP.DeleteBankSubmodule(ctx, 50, 51)
		}, "DELETE", "/api/constructor-ep/bank/modules/50/tree/module/51", nil, ""},
		{"AddBankDisciplines", func(t *testing.T, c *myitmo.Client) error {
			return c.ConstructorEP.AddBankDisciplines(ctx, 50, 51, []int64{101})
		}, "POST", "/api/constructor-ep/bank/modules/50/tree/module/51/discipline", nil, `[101]`},
		{"SetBankDisciplineSemesters", func(t *testing.T, c *myitmo.Client) error {
			return c.ConstructorEP.SetBankDisciplineSemesters(ctx, 50, 51, 101, []int{3})
		}, "PATCH", "/api/constructor-ep/bank/modules/50/tree/module/51/discipline/101", nil, `[3]`},
		{"DeleteBankDiscipline", func(t *testing.T, c *myitmo.Client) error {
			return c.ConstructorEP.DeleteBankDiscipline(ctx, 50, 51, 101)
		}, "DELETE", "/api/constructor-ep/bank/modules/50/tree/module/51/discipline/101", nil, ""},
	})
}

func TestConstructorEPKUGRequests(t *testing.T) {
	ctx := t.Context()
	act := myitmo.EPKUGActivityForm{
		TypeID: myitmo.EPActivityExamSession, SemesterID: myitmo.EPKUGSemesterID(2024, myitmo.EPSemesterAutumn),
		DateStart: myitmo.NewDate(2025, 1, 9), DateEnd: myitmo.NewDate(2025, 1, 25),
	}
	sems := []myitmo.EPSemesterDatesForm{
		{Year: 2025, Order: myitmo.EPSemesterAutumn, DateStart: myitmo.NewDate(2025, 9, 1), DateEnd: myitmo.NewDate(2026, 1, 31)},
		{Year: 2025, Order: myitmo.EPSemesterSpring, DateStart: myitmo.NewDate(2026, 2, 1), DateEnd: myitmo.NewDate(2026, 8, 31)},
	}
	semsJSON := `[{"year":2025,"order":1,"date_start":"01-09-2025","date_end":"31-01-2026"},{"year":2025,"order":2,"date_start":"01-02-2026","date_end":"31-08-2026"}]`

	runEPCases(t, []epCase{
		{"KUGTemplates", func(t *testing.T, c *myitmo.Client) error {
			return epErr(c.ConstructorEP.KUGTemplates(ctx, myitmo.EPKUGTemplateListParams{
				Query: "шаблон", StandardID: 7, YearStart: 2024, UsedInEP: myitmo.Ptr(false), Personal: true, Valid: myitmo.Ptr(true), Limit: 20, Offset: 20,
			}))
		}, "GET", "/api/constructor-ep/kug/templates/list", url.Values{
			"query": {"шаблон"}, "standard_id": {"7"}, "year_start": {"2024"}, "used_in_ep": {"false"}, "personal": {"1"},
			"valid": {"true"}, "limit": {"20"}, "offset": {"20"},
		}, ""},
		{"CreateKUGTemplate", func(t *testing.T, c *myitmo.Client) error {
			return epErr(c.ConstructorEP.CreateKUGTemplate(ctx, myitmo.EPKUGTemplateForm{Name: "Шаблон", YearStart: 2025, StandardID: 7}))
		}, "POST", "/api/constructor-ep/kug/templates/create", nil, `{"name":"Шаблон","year_start":2025,"standard_id":7}`},
		{"KUGTemplate", func(t *testing.T, c *myitmo.Client) error {
			return epErr(c.ConstructorEP.KUGTemplate(ctx, 3))
		}, "GET", "/api/constructor-ep/kug/templates/3", nil, ""},
		{"RenameKUGTemplate", func(t *testing.T, c *myitmo.Client) error {
			return c.ConstructorEP.RenameKUGTemplate(ctx, 3, "Новое имя")
		}, "PATCH", "/api/constructor-ep/kug/templates/update_name/3", nil, `{"name":"Новое имя"}`},
		{"KUGTemplateActivities", func(t *testing.T, c *myitmo.Client) error {
			return epErr(c.ConstructorEP.KUGTemplateActivities(ctx, 3))
		}, "GET", "/api/constructor-ep/kug/templates/3/activities", nil, ""},
		{"KUGTemplateHoursByYear", func(t *testing.T, c *myitmo.Client) error {
			return epErr(c.ConstructorEP.KUGTemplateHoursByYear(ctx, 3, 2024))
		}, "GET", "/api/constructor-ep/kug/templates/hours_by_year/3/2024", nil, ""},
		{"KUGTemplateSemesterDates", func(t *testing.T, c *myitmo.Client) error {
			return epErr(c.ConstructorEP.KUGTemplateSemesterDates(ctx, 3, 2024))
		}, "GET", "/api/constructor-ep/kug/templates/semesters_dates/3/2024", nil, ""},
		{"ValidateKUGTemplate", func(t *testing.T, c *myitmo.Client) error {
			return epErr(c.ConstructorEP.ValidateKUGTemplate(ctx, 3))
		}, "GET", "/api/constructor-ep/kug/templates/validate/3", nil, ""},
		{"KUGTemplateUsage", func(t *testing.T, c *myitmo.Client) error {
			return epErr(c.ConstructorEP.KUGTemplateUsage(ctx, 3, "", 8, 0))
		}, "GET", "/api/constructor-ep/kug/templates/using/3", url.Values{"limit": {"8"}, "offset": {"0"}}, ""},
		{"CreateKUGTemplateActivity", func(t *testing.T, c *myitmo.Client) error {
			return c.ConstructorEP.CreateKUGTemplateActivity(ctx, 3, act)
		}, "POST", "/api/constructor-ep/kug/templates/activities/create", nil,
			`{"type_id":11,"kug_template_id":3,"semester_id":202420251,"date_start":"09-01-2025","date_end":"25-01-2025"}`},
		{"ReplaceKUGTemplateActivity", func(t *testing.T, c *myitmo.Client) error {
			return c.ConstructorEP.ReplaceKUGTemplateActivity(ctx, 3, act)
		}, "POST", "/api/constructor-ep/kug/templates/activities/replace", nil,
			`{"type_id":11,"kug_template_id":3,"semester_id":202420251,"date_start":"09-01-2025","date_end":"25-01-2025"}`},
		{"ApplyKUGTemplate", func(t *testing.T, c *myitmo.Client) error {
			return c.ConstructorEP.ApplyKUGTemplate(ctx, 3, []int64{42, 43})
		}, "POST", "/api/constructor-ep/kug/add_to_ep/3", nil, `{"ep_ids":[42,43]}`},
		{"KUGByProgram", func(t *testing.T, c *myitmo.Client) error {
			return epErr(c.ConstructorEP.KUGByProgram(ctx, 42))
		}, "GET", "/api/constructor-ep/kug/by_ep/42", nil, ""},
		{"KUGActivities", func(t *testing.T, c *myitmo.Client) error {
			return epErr(c.ConstructorEP.KUGActivities(ctx, 8))
		}, "GET", "/api/constructor-ep/kug/8/activities", nil, ""},
		{"KUGHoursByYear", func(t *testing.T, c *myitmo.Client) error {
			return epErr(c.ConstructorEP.KUGHoursByYear(ctx, 8, 2025))
		}, "GET", "/api/constructor-ep/kug/hours_by_year/8/2025", nil, ""},
		{"KUGSemesterDates", func(t *testing.T, c *myitmo.Client) error {
			return epErr(c.ConstructorEP.KUGSemesterDates(ctx, 8, 2025))
		}, "GET", "/api/constructor-ep/kug/activities/kug_semesters_dates/8/2025", nil, ""},
		{"ValidateKUG", func(t *testing.T, c *myitmo.Client) error {
			return epErr(c.ConstructorEP.ValidateKUG(ctx, 8))
		}, "GET", "/api/constructor-ep/kug/validate/8", nil, ""},
		{"CreateKUGActivity", func(t *testing.T, c *myitmo.Client) error {
			return c.ConstructorEP.CreateKUGActivity(ctx, 8, act)
		}, "POST", "/api/constructor-ep/kug/activities/create", nil,
			`{"type_id":11,"kug_ep_id":8,"semester_id":202420251,"date_start":"09-01-2025","date_end":"25-01-2025"}`},
		{"ReplaceKUGActivity", func(t *testing.T, c *myitmo.Client) error {
			return c.ConstructorEP.ReplaceKUGActivity(ctx, 8, act)
		}, "POST", "/api/constructor-ep/kug/activities/replace", nil,
			`{"type_id":11,"kug_ep_id":8,"semester_id":202420251,"date_start":"09-01-2025","date_end":"25-01-2025"}`},
		{"KUGComments", func(t *testing.T, c *myitmo.Client) error {
			return epErr(c.ConstructorEP.KUGComments(ctx, 8))
		}, "GET", "/api/constructor-ep/kug/8/comments", nil, ""},
		{"AddKUGComment", func(t *testing.T, c *myitmo.Client) error {
			return c.ConstructorEP.AddKUGComment(ctx, 8, "Замечание")
		}, "POST", "/api/constructor-ep/kug/8/comments/create", nil, `{"text":"Замечание"}`},
		{"UpdateKUGComment", func(t *testing.T, c *myitmo.Client) error {
			return c.ConstructorEP.UpdateKUGComment(ctx, 4, "Исправлено")
		}, "PATCH", "/api/constructor-ep/kug/comments/4/update", nil, `{"text":"Исправлено"}`},
		{"DeleteKUGComment", func(t *testing.T, c *myitmo.Client) error {
			return c.ConstructorEP.DeleteKUGComment(ctx, 4)
		}, "DELETE", "/api/constructor-ep/kug/comments/4/delete", nil, ""},
		{"KUGToExpertise", func(t *testing.T, c *myitmo.Client) error {
			return c.ConstructorEP.KUGToExpertise(ctx, 8)
		}, "POST", "/api/constructor-ep/kug/8/to_expertise", nil, ""},
		{"KUGToRevision", func(t *testing.T, c *myitmo.Client) error {
			return c.ConstructorEP.KUGToRevision(ctx, 8)
		}, "POST", "/api/constructor-ep/kug/8/to_revision", nil, ""},
		{"KUGToSigning", func(t *testing.T, c *myitmo.Client) error {
			return c.ConstructorEP.KUGToSigning(ctx, 8)
		}, "POST", "/api/constructor-ep/kug/8/to_signing", nil, ""},
		{"CreateProductionCalendar", func(t *testing.T, c *myitmo.Client) error {
			return c.ConstructorEP.CreateProductionCalendar(ctx, 2026)
		}, "POST", "/api/constructor-ep/kug/production_calendar/create/for_year", nil, `{"year":2026}`},
		{"ProductionCalendar", func(t *testing.T, c *myitmo.Client) error {
			return epErr(c.ConstructorEP.ProductionCalendar(ctx, 2026))
		}, "GET", "/api/constructor-ep/kug/production_calendar/list/by_year/2026", nil, ""},
		{"ProductionCalendarYears", func(t *testing.T, c *myitmo.Client) error {
			return epErr(c.ConstructorEP.ProductionCalendarYears(ctx))
		}, "GET", "/api/constructor-ep/kug/production_calendar/years/list", nil, ""},
		{"AddSemesters", func(t *testing.T, c *myitmo.Client) error {
			return c.ConstructorEP.AddSemesters(ctx, sems)
		}, "POST", "/api/constructor-ep/kug/semesters/add", nil, semsJSON},
		{"UpdateSemesters", func(t *testing.T, c *myitmo.Client) error {
			return c.ConstructorEP.UpdateSemesters(ctx, sems)
		}, "PATCH", "/api/constructor-ep/kug/semesters/update", nil, semsJSON},
		{"SemestersByYear", func(t *testing.T, c *myitmo.Client) error {
			return epErr(c.ConstructorEP.SemestersByYear(ctx, 2025))
		}, "GET", "/api/constructor-ep/kug/semesters/by_year/2025", nil, ""},
		{"Semesters", func(t *testing.T, c *myitmo.Client) error {
			return epErr(c.ConstructorEP.Semesters(ctx))
		}, "GET", "/api/constructor-ep/kug/semesters/list", nil, ""},
	})
}

func TestConstructorEPDecodeBank(t *testing.T) {
	f, c := newFake(t)
	ctx := t.Context()

	f.result(`{"modules":[{"id":50,"name":"Модуль","education_level":{"name":"Бакалавриат"},"capacity":12,
		"created_at":"2025-03-01T10:00:00+03:00","rule_type":null,"rule_value":null,"start_semester":3,
		"creator":{"fio":"Тестов Тест"},"status":{"id":6,"name":"Одобрен"}}],"count":1,"can_create":true}`)
	list := must[*myitmo.EPBankModuleList](t)(c.ConstructorEP.BankModules(ctx, myitmo.EPBankModuleListParams{}))
	m := list.Modules[0]
	if !list.CanCreate || m.EducationLevel.Name != "Бакалавриат" || m.RuleType != nil || *m.StartSemester != 3 ||
		m.CreatedAt.Month() != 3 || m.Creator.FIO != "Тестов Тест" || m.Status.ID != myitmo.EPStatusSigned {
		t.Fatalf("list = %+v", list)
	}

	f.result(`{"id":50,"name":"Модуль","name_ru":"Модуль","name_en":"Module","block":{"id":1,"name":"Блок 1"},
		"educational_standard":{"id":7,"name":"ОС ИТМО","education_level":{"id":1,"name":"Бакалавриат"}},
		"rule_type":{"name":"Выбор"},"rule_value":[1],"description":null,"status":{"id":2,"name":"В работе"},
		"canEdit":true,"is_admin":false,"root_module":{"id":51},"editors":[{"isu":100001,"fio":"Тестов Тест"}],
		"accessible_programs":null,"belongs_to":[{"id":42,"name":"Программа","module_id":60}]}`)
	bm := must[*myitmo.EPBankModule](t)(c.ConstructorEP.BankModule(ctx, 50))
	if !bm.CanEdit || bm.EducationalStandard.EducationLevel.ID != 1 || bm.RuleType.Name != "Выбор" ||
		string(bm.RootModule) != `{"id":51}` || bm.Editors[0].ISU != 100001 || bm.AccessiblePrograms != nil || bm.BelongsTo[0].ModuleID != 60 {
		t.Fatalf("module = %+v", bm)
	}

	f.result(`{"id":51,"name":"Модуль","type":"module","is_root":true,"is_shop":false,"rule_types":[1,41],"parent_id":null,
		"children":[{"id":101,"type":"discipline","name":"Дисциплина","module_id":51,"semesters":[3],"semesters_count":1,"status":{"id":1}}]}`)
	tree := must[*myitmo.EPPlanNode](t)(c.ConstructorEP.BankModuleTree(ctx, 50))
	if !tree.IsRoot || tree.Children[0].ModuleID != 51 || tree.Children[0].Status.ID != myitmo.EPStatusDraft {
		t.Fatalf("tree = %+v", tree)
	}
}

func TestConstructorEPDecodeKUG(t *testing.T) {
	f, c := newFake(t)
	ctx := t.Context()

	f.result(`{"kug_templates":[{"id":3,"name":"Шаблон","year_start":2024,"standard_name":"ОС ИТМО","used_in_ep":true}],"count":1}`)
	tl := must[*myitmo.EPKUGTemplateList](t)(c.ConstructorEP.KUGTemplates(ctx, myitmo.EPKUGTemplateListParams{}))
	if tl.Count != 1 || !tl.KUGTemplates[0].UsedInEP || tl.KUGTemplates[0].YearStart != 2024 {
		t.Fatalf("templates = %+v", tl)
	}

	f.result(`{"id":3,"name":"Шаблон","standard_id":7,"year_start":2024,"used_in_ep":false}`)
	if tpl := must[*myitmo.EPKUGTemplate](t)(c.ConstructorEP.KUGTemplate(ctx, 3)); tpl.StandardID != 7 {
		t.Fatalf("template = %+v", tpl)
	}

	f.result(`[{"type":{"id":9},"date_start":"2024-09-01","date_end":"2024-12-28"}]`)
	acts := must[[]myitmo.EPKUGActivity](t)(c.ConstructorEP.KUGActivities(ctx, 8))
	if acts[0].Type.ID != myitmo.EPActivityTheory || acts[0].DateEnd != myitmo.NewDate(2024, 12, 28) {
		t.Fatalf("activities = %+v", acts)
	}

	f.result(`{"autumn_sem_hours":{"9":100,"11":14},"spring_sem_hours":{"2":24}}`)
	h := must[*myitmo.EPKUGHoursByYear](t)(c.ConstructorEP.KUGHoursByYear(ctx, 8, 2024))
	if h.AutumnSemHours["11"] != 14 || h.SpringSemHours["2"] != 24 {
		t.Fatalf("hours = %+v", h)
	}

	f.result(`{"years_errors":{"2024":{"semesters_errors":{"2":{"missed_activities_error":{"missed_periods":[{"date_start":"2025-06-01","date_end":"2025-06-07"}]}}}}}}`)
	v := must[*myitmo.EPKUGValidation](t)(c.ConstructorEP.ValidateKUG(ctx, 8))
	if v.YearsErrors["2024"].SemestersErrors["2"].MissedActivitiesError.MissedPeriods[0].DateEnd != myitmo.NewDate(2025, 6, 7) {
		t.Fatalf("validation = %+v", v)
	}

	f.result(`null`)
	if v := must[*myitmo.EPKUGValidation](t)(c.ConstructorEP.ValidateKUGTemplate(ctx, 3)); v != nil {
		t.Fatalf("validation = %+v, want nil", v)
	}

	f.result(`{"count":1,"programs":[{"id":1,"ep_id":42,"ep_name":"Программа","implementer_name":"Факультет"}]}`)
	if u := must[*myitmo.EPKUGTemplateUsage](t)(c.ConstructorEP.KUGTemplateUsage(ctx, 3, "", 8, 0)); u.Programs[0].EPID != 42 {
		t.Fatalf("usage = %+v", u)
	}

	f.result(`{"id":8,"year_start":2024}`)
	if k := must[*myitmo.EPKUG](t)(c.ConstructorEP.KUGByProgram(ctx, 42)); k.ID != 8 || k.YearStart != 2024 {
		t.Fatalf("kug = %+v", k)
	}

	f.result(`[{"id":4,"text":"Замечание","date":"2025-04-01T09:30:00+03:00","user":{"isu":100001,"fio":"Тестов Тест"}}]`)
	cm := must[[]myitmo.EPKUGComment](t)(c.ConstructorEP.KUGComments(ctx, 8))
	if cm[0].User.ISU != 100001 || cm[0].Date.Hour() != 9 {
		t.Fatalf("comments = %+v", cm)
	}

	f.result(`[{"date":"2026-01-01","type_id":8},{"date":"2026-01-03","type_id":1}]`)
	days := must[[]myitmo.EPProductionCalendarDay](t)(c.ConstructorEP.ProductionCalendar(ctx, 2026))
	if days[0].TypeID != myitmo.EPActivityPublicHoliday || days[1].Date != myitmo.NewDate(2026, 1, 3) {
		t.Fatalf("days = %+v", days)
	}

	f.result(`[2024,2025]`)
	if ys := must[[]int](t)(c.ConstructorEP.ProductionCalendarYears(ctx)); len(ys) != 2 || ys[1] != 2025 {
		t.Fatalf("years = %v", ys)
	}

	f.result(`[{"date_start":"2025-09-01","date_end":"2026-01-31"},{"date_start":"2026-02-01","date_end":"2026-08-31"}]`)
	if p := must[[]myitmo.EPPeriod](t)(c.ConstructorEP.SemestersByYear(ctx, 2025)); p[1].DateStart != myitmo.NewDate(2026, 2, 1) {
		t.Fatalf("semesters = %+v", p)
	}

	f.result(`[{"id":202520261,"date_start":"2025-09-01","date_end":"2026-01-31"}]`)
	s := must[[]myitmo.EPKUGSemester](t)(c.ConstructorEP.Semesters(ctx))
	if s[0].ID != myitmo.EPKUGSemesterID(2025, myitmo.EPSemesterAutumn) {
		t.Fatalf("semesters = %+v", s)
	}
}
