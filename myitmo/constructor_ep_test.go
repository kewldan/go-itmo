package myitmo_test

import (
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/kewldan/go-itmo/myitmo"
)

// epCase is one request-shape expectation of the constructor-ep service.
type epCase struct {
	name   string
	call   func(t *testing.T, c *myitmo.Client) error
	method string
	path   string
	// query is compared exactly; nil means no query.
	query url.Values
	// body is the expected JSON body; empty means no body.
	body string
}

func runEPCases(t *testing.T, cases []epCase) {
	t.Helper()
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			f, c := newFake(t)
			if err := tc.call(t, c); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			r := f.expect(tc.method, tc.path)
			if len(tc.query) == 0 {
				if len(r.Query) != 0 {
					t.Fatalf("query = %v, want none", r.Query)
				}
			} else if !reflect.DeepEqual(r.Query, tc.query) {
				t.Fatalf("query = %v, want %v", r.Query, tc.query)
			}
			if tc.body == "" {
				if len(r.Body) != 0 {
					t.Fatalf("body = %s, want none", r.Body)
				}
			} else {
				r.sameJSON(t, tc.body)
			}
		})
	}
}

func epErr[T any](_ T, err error) error { return err }

func TestConstructorEPProgramRequests(t *testing.T) {
	ctx := t.Context()
	note := "Synthetic note"
	form := myitmo.EPProgramForm{
		NameRU: "Программа", NameEN: "Programme", EnrollmentYearID: 5, EducationalStandardID: 7,
		ImplementerID: 9, Languages: []int64{1}, ActualEducationDuration: 2, DirectionsID: []int64{3, 4},
		Attributes: []int64{}, Note: &note, AbitExport: true,
	}
	formJSON := `{"name_ru":"Программа","name_en":"Programme","enrollment_year_id":5,"educational_standard_id":7,
		"implementer_id":9,"languages":[1],"actual_education_duration":2,"directions_id":[3,4],"attributes":[],
		"note":"Synthetic note","abit_export":true}`
	start := time.Date(2024, 9, 1, 0, 0, 0, 0, time.UTC)
	module := myitmo.EPModuleForm{NameRU: "Модуль", NameEN: "Module", RuleTypeID: 1, RuleValue: myitmo.EPRuleValue{2}, ParentID: 10, BlockID: 11, StandardID: 12}
	moduleJSON := `{"name_ru":"Модуль","name_en":"Module","rule_type_id":1,"rule_value":[2],"parent_id":10,"block_id":11,"standard_id":12}`

	runEPCases(t, []epCase{
		{"List", func(t *testing.T, c *myitmo.Client) error {
			return epErr(c.ConstructorEP.List(ctx, myitmo.EPProgramListParams{
				Query: "data", PlanStatusIDs: []int64{2, 3}, KUGStatusIDs: []int64{5}, EducationLevelIDs: []int64{1},
				ImplementerIDs: []int64{8}, DirectionIDs: []int64{4}, EnrollmentYears: []int64{6},
				EducationFormIDs: []int64{1}, CharacteristicsStatusIDs: []int64{6}, StandardID: 7, KUGFree: true, Limit: 15, Offset: 30,
			}))
		}, "GET", "/api/constructor-ep/programs/list", url.Values{
			"query": {"data"}, "plan_status_id": {"2", "3"}, "kug_status_id": {"5"}, "education_level_id": {"1"},
			"implementer_id": {"8"}, "direction_id": {"4"}, "enrollment_year": {"6"}, "education_form_id": {"1"},
			"characteristics_status_id": {"6"}, "standard_id": {"7"}, "kug_free": {"true"}, "archive": {"0"},
			"limit": {"15"}, "offset": {"30"},
		}, ""},
		{"ListArchived", func(t *testing.T, c *myitmo.Client) error {
			return epErr(c.ConstructorEP.List(ctx, myitmo.EPProgramListParams{Archived: true}))
		}, "GET", "/api/constructor-ep/programs/list", url.Values{"archive": {"1"}, "offset": {"0"}}, ""},
		{"SimpleList", func(t *testing.T, c *myitmo.Client) error {
			return epErr(c.ConstructorEP.SimpleList(ctx, myitmo.EPSimpleListParams{Query: "x", EnrollmentYears: []int64{2023, 2024}, ImplementerID: 3, EducationLevelID: 2}))
		}, "GET", "/api/constructor-ep/programs/simple_list", url.Values{
			"query": {"x"}, "enrollment_year": {"2023", "2024"}, "implementer_id": {"3"}, "education_level_id": {"2"}, "archive": {"0"},
		}, ""},
		{"Tasks", func(t *testing.T, c *myitmo.Client) error {
			return epErr(c.ConstructorEP.Tasks(ctx, myitmo.EPTasksParams{Query: "q", ImplementerID: 4, TaskType: myitmo.EPTaskKUG}))
		}, "GET", "/api/constructor-ep/programs/tasks_list", url.Values{"query": {"q"}, "implementer_id": {"4"}, "task_type": {"kug"}}, ""},
		{"Create", func(t *testing.T, c *myitmo.Client) error {
			return epErr(c.ConstructorEP.Create(ctx, form))
		}, "POST", "/api/constructor-ep/programs/create", nil, formJSON},
		{"Update", func(t *testing.T, c *myitmo.Client) error {
			return c.ConstructorEP.Update(ctx, 42, form)
		}, "PATCH", "/api/constructor-ep/programs/42", nil, formJSON},
		{"Info", func(t *testing.T, c *myitmo.Client) error {
			return epErr(c.ConstructorEP.Info(ctx, 42))
		}, "GET", "/api/constructor-ep/programs/42/info", nil, ""},
		{"Status", func(t *testing.T, c *myitmo.Client) error {
			return epErr(c.ConstructorEP.Status(ctx, 42))
		}, "GET", "/api/constructor-ep/programs/42/status", nil, ""},
		{"SetAnnotation", func(t *testing.T, c *myitmo.Client) error {
			return c.ConstructorEP.SetAnnotation(ctx, 42, "Аннотация", "")
		}, "PATCH", "/api/constructor-ep/programs/42/annotation", nil, `{"annotation_ru":"Аннотация","annotation_en":null}`},
		{"SetNote", func(t *testing.T, c *myitmo.Client) error {
			return c.ConstructorEP.SetNote(ctx, 42, "")
		}, "PATCH", "/api/constructor-ep/programs/42/note", nil, `{"note":null}`},
		{"SetMobility", func(t *testing.T, c *myitmo.Client) error {
			return c.ConstructorEP.SetMobility(ctx, 42, myitmo.EPMobilityForm{
				PartnerUniversityID:   myitmo.Ptr[int64](3),
				AcademicMobilityStart: myitmo.Ptr(myitmo.NewDate(2025, 2, 1)),
				AcademicMobilityEnd:   myitmo.Ptr(myitmo.NewDate(2025, 6, 30)),
			})
		}, "PATCH", "/api/constructor-ep/programs/42/mobility", nil,
			`{"partner_university_id":3,"academic_mobility_start":"2025-02-01","academic_mobility_end":"2025-06-30"}`},
		{"ClearMobility", func(t *testing.T, c *myitmo.Client) error {
			return c.ConstructorEP.SetMobility(ctx, 42, myitmo.EPMobilityForm{})
		}, "PATCH", "/api/constructor-ep/programs/42/mobility", nil,
			`{"partner_university_id":null,"academic_mobility_start":null,"academic_mobility_end":null}`},
		{"SetKCP", func(t *testing.T, c *myitmo.Client) error {
			return c.ConstructorEP.SetKCP(ctx, 42, []myitmo.EPKCP{{KCPID: 1, DirectionID: 2, Code: "01.03.02", Name: "Направление", BudgetForm: myitmo.Ptr(25), AbitExport: true}})
		}, "PATCH", "/api/constructor-ep/programs/42/kcp", nil,
			`[{"kcp_id":1,"direction_id":2,"code":"01.03.02","name":"Направление","budget_form":25,"contract_form":null,"other_form":null,"abit_export":true,"military_department":false}]`},
		{"AddManagers", func(t *testing.T, c *myitmo.Client) error {
			return c.ConstructorEP.AddManagers(ctx, 42, []myitmo.EPManagerForm{{ISU: 100001, Positions: []myitmo.EPManagerPositionForm{{PositionID: 1, DateStart: start, DateEnd: start.AddDate(1, 0, 0)}}}})
		}, "POST", "/api/constructor-ep/programs/42/managers", nil,
			`[{"isu":100001,"positions":[{"position_id":1,"date_start":"2024-09-01T00:00:00Z","date_end":"2025-09-01T00:00:00Z"}]}]`},
		{"UpdateManager", func(t *testing.T, c *myitmo.Client) error {
			return c.ConstructorEP.UpdateManager(ctx, 42, 100001, []myitmo.EPManagerPositionForm{{PositionID: 3, DateStart: start, DateEnd: start}})
		}, "PATCH", "/api/constructor-ep/programs/42/managers/100001", nil,
			`[{"position_id":3,"date_start":"2024-09-01T00:00:00Z","date_end":"2024-09-01T00:00:00Z"}]`},
		{"DeleteManager", func(t *testing.T, c *myitmo.Client) error {
			return c.ConstructorEP.DeleteManager(ctx, 42, 100001)
		}, "DELETE", "/api/constructor-ep/programs/42/managers/100001", nil, ""},
		{"Archive", func(t *testing.T, c *myitmo.Client) error {
			return c.ConstructorEP.Archive(ctx, 42)
		}, "POST", "/api/constructor-ep/programs/42/archive", nil, ""},
		{"SendToISU", func(t *testing.T, c *myitmo.Client) error {
			return c.ConstructorEP.SendToISU(ctx, 42)
		}, "POST", "/api/constructor-ep/programs/42/to_isu", nil, ""},
		{"AddCompany", func(t *testing.T, c *myitmo.Client) error {
			return c.ConstructorEP.AddCompany(ctx, 42, myitmo.EPCompany{INN: "7700000000", NameShortWithOPF: "ООО Пример", NameFullWithOPF: "Общество Пример"}, 2)
		}, "POST", "/api/constructor-ep/programs/42/companies", nil,
			`{"inn":"7700000000","name_short_with_opf":"ООО Пример","name_full_with_opf":"Общество Пример","type_id":2}`},
		{"DeleteCompany", func(t *testing.T, c *myitmo.Client) error {
			return c.ConstructorEP.DeleteCompany(ctx, 42, 7)
		}, "DELETE", "/api/constructor-ep/programs/42/companies/7", nil, ""},
		{"Comments", func(t *testing.T, c *myitmo.Client) error {
			return epErr(c.ConstructorEP.Comments(ctx, 42))
		}, "GET", "/api/constructor-ep/programs/42/comments", nil, ""},
		{"AddComment", func(t *testing.T, c *myitmo.Client) error {
			return c.ConstructorEP.AddComment(ctx, 42, "Проверить", myitmo.EPPartPlan)
		}, "POST", "/api/constructor-ep/programs/42/comments", nil, `{"comment":"Проверить","part_id":2}`},
		{"UpdateComment", func(t *testing.T, c *myitmo.Client) error {
			return c.ConstructorEP.UpdateComment(ctx, 42, 5, "Исправлено")
		}, "PATCH", "/api/constructor-ep/programs/42/comments/5", nil, `{"comment":"Исправлено"}`},
		{"DeleteComment", func(t *testing.T, c *myitmo.Client) error {
			return c.ConstructorEP.DeleteComment(ctx, 42, 5)
		}, "DELETE", "/api/constructor-ep/programs/42/comments/5", nil, ""},
		{"Plan", func(t *testing.T, c *myitmo.Client) error {
			return epErr(c.ConstructorEP.Plan(ctx, 42))
		}, "GET", "/api/constructor-ep/programs/42/plan", nil, ""},
		{"PlanPracticeCredits", func(t *testing.T, c *myitmo.Client) error {
			return epErr(c.ConstructorEP.PlanPracticeCredits(ctx, 42))
		}, "GET", "/api/constructor-ep/programs/42/plan/capacity_sum/by_semesters", nil, ""},
		{"CreatePlanModule", func(t *testing.T, c *myitmo.Client) error {
			return epErr(c.ConstructorEP.CreatePlanModule(ctx, 42, module))
		}, "POST", "/api/constructor-ep/programs/42/plan/module", nil, moduleJSON},
		{"UpdatePlanModule", func(t *testing.T, c *myitmo.Client) error {
			m := module
			m.RuleValue = nil
			return epErr(c.ConstructorEP.UpdatePlanModule(ctx, 42, 13, m))
		}, "PATCH", "/api/constructor-ep/programs/42/plan/module/13", nil,
			`{"name_ru":"Модуль","name_en":"Module","rule_type_id":1,"rule_value":null,"parent_id":10,"block_id":11,"standard_id":12}`},
		{"DeletePlanModule", func(t *testing.T, c *myitmo.Client) error {
			return c.ConstructorEP.DeletePlanModule(ctx, 42, 13)
		}, "DELETE", "/api/constructor-ep/programs/42/plan/module/13", nil, ""},
		{"AddPlanDisciplines", func(t *testing.T, c *myitmo.Client) error {
			return c.ConstructorEP.AddPlanDisciplines(ctx, 42, 13, []int64{101, 102})
		}, "POST", "/api/constructor-ep/programs/42/plan/module/13/discipline", nil, `[101,102]`},
		{"AddPlanBankModules", func(t *testing.T, c *myitmo.Client) error {
			return c.ConstructorEP.AddPlanBankModules(ctx, 42, 13, []int64{201})
		}, "POST", "/api/constructor-ep/programs/42/plan/module/13/bank_module", nil, `[201]`},
		{"SetPlanDisciplineSemesters", func(t *testing.T, c *myitmo.Client) error {
			return c.ConstructorEP.SetPlanDisciplineSemesters(ctx, 42, 13, 101, []int{1, 2})
		}, "PATCH", "/api/constructor-ep/programs/42/plan/module/13/disciplines/101", nil, `[1,2]`},
		{"DeletePlanDiscipline", func(t *testing.T, c *myitmo.Client) error {
			return c.ConstructorEP.DeletePlanDiscipline(ctx, 42, 13, 101)
		}, "DELETE", "/api/constructor-ep/programs/42/plan/module/13/disciplines/101", nil, ""},
		{"PlanToExpertise", func(t *testing.T, c *myitmo.Client) error {
			return c.ConstructorEP.PlanToExpertise(ctx, 42)
		}, "POST", "/api/constructor-ep/programs/42/plan/to_expertise", nil, ""},
		{"PlanToRevision", func(t *testing.T, c *myitmo.Client) error {
			return c.ConstructorEP.PlanToRevision(ctx, 42)
		}, "POST", "/api/constructor-ep/programs/42/plan/to_revision", nil, ""},
		{"PlanToSigning", func(t *testing.T, c *myitmo.Client) error {
			return c.ConstructorEP.PlanToSigning(ctx, 42)
		}, "POST", "/api/constructor-ep/programs/42/plan/to_signing", nil, ""},
		{"SearchCompanies", func(t *testing.T, c *myitmo.Client) error {
			return epErr(c.ConstructorEP.SearchCompanies(ctx, " Пример "))
		}, "GET", "/api/constructor-ep/companies", url.Values{"query": {"Пример"}}, ""},
		{"DeleteCompanyFile", func(t *testing.T, c *myitmo.Client) error {
			return c.ConstructorEP.DeleteCompanyFile(ctx, 7)
		}, "DELETE", "/api/constructor-ep/companies/7/file", nil, ""},
		{"ITTypes", func(t *testing.T, c *myitmo.Client) error {
			return epErr(c.ConstructorEP.ITTypes(ctx))
		}, "GET", "/api/constructor-ep/it_types", nil, ""},
		{"Disciplines", func(t *testing.T, c *myitmo.Client) error {
			return epErr(c.ConstructorEP.Disciplines(ctx, myitmo.EPDisciplineSearchParams{Query: "мат", EducationLevelID: 1, StatusIDs: []int64{2, 3, 5, 6, 7}, ProgramTypeIDs: []int64{1, 2}}))
		}, "GET", "/api/constructor-ep/disciplines/list", url.Values{
			"query": {"мат"}, "education_level_id": {"1"}, "status": {"2,3,5,6,7"}, "program_type_id": {"1", "2"},
		}, ""},
		{"SearchPeople", func(t *testing.T, c *myitmo.Client) error {
			return epErr(c.ConstructorEP.SearchPeople(ctx, "Иван Ив/"))
		}, "GET", "/api/constructor-ep/people/search/%D0%98%D0%B2%D0%B0%D0%BD%20%D0%98%D0%B2%2F", nil, ""},
		{"IsAdmin", func(t *testing.T, c *myitmo.Client) error {
			return epErr(c.ConstructorEP.IsAdmin(ctx))
		}, "GET", "/api/constructor-ep/users/is_admin", nil, ""},
	})
}

func TestConstructorEPReferenceRequests(t *testing.T) {
	ctx := t.Context()
	ep := func(c *myitmo.Client) *myitmo.ConstructorEPService { return c.ConstructorEP }
	refs := map[string]func(*myitmo.Client) error{
		"levels":               func(c *myitmo.Client) error { return epErr(ep(c).Levels(ctx)) },
		"languages":            func(c *myitmo.Client) error { return epErr(ep(c).Languages(ctx)) },
		"implementers":         func(c *myitmo.Client) error { return epErr(ep(c).Implementers(ctx)) },
		"forms":                func(c *myitmo.Client) error { return epErr(ep(c).Forms(ctx)) },
		"standards":            func(c *myitmo.Client) error { return epErr(ep(c).Standards(ctx)) },
		"standards/blocks":     func(c *myitmo.Client) error { return epErr(ep(c).StandardBlocks(ctx)) },
		"directions":           func(c *myitmo.Client) error { return epErr(ep(c).Directions(ctx, 0)) },
		"partner_universities": func(c *myitmo.Client) error { return epErr(ep(c).PartnerUniversities(ctx)) },
		"statuses":             func(c *myitmo.Client) error { return epErr(ep(c).Statuses(ctx)) },
		"countries":            func(c *myitmo.Client) error { return epErr(ep(c).Countries(ctx)) },
		"attributes":           func(c *myitmo.Client) error { return epErr(ep(c).Attributes(ctx)) },
		"enrollment_years":     func(c *myitmo.Client) error { return epErr(ep(c).EnrollmentYears(ctx)) },
		"positions":            func(c *myitmo.Client) error { return epErr(ep(c).Positions(ctx)) },
		"rule_types":           func(c *myitmo.Client) error { return epErr(ep(c).RuleTypes(ctx)) },
		"activity_types":       func(c *myitmo.Client) error { return epErr(ep(c).ActivityTypes(ctx)) },
	}
	var cases []epCase
	for name, fn := range refs {
		cases = append(cases, epCase{name, func(t *testing.T, c *myitmo.Client) error { return fn(c) },
			"GET", "/api/constructor-ep/references/" + name, nil, ""})
	}
	cases = append(cases, epCase{"directions by level", func(t *testing.T, c *myitmo.Client) error {
		return epErr(c.ConstructorEP.Directions(ctx, 3))
	}, "GET", "/api/constructor-ep/references/directions", url.Values{"education_level_id": {"3"}}, ""})
	runEPCases(t, cases)
}

func TestConstructorEPDownloads(t *testing.T) {
	ctx := t.Context()
	cases := []struct {
		name, method, path string
		call               func(c *myitmo.Client) (*myitmo.File, error)
	}{
		{"SignedPlan", "GET", "/api/constructor-ep/programs/42/signed", func(c *myitmo.Client) (*myitmo.File, error) {
			return c.ConstructorEP.SignedPlan(ctx, 42)
		}},
		{"PlanDocument", "GET", "/api/constructor-ep/programs/42/plan/abit/xls", func(c *myitmo.Client) (*myitmo.File, error) {
			return c.ConstructorEP.PlanDocument(ctx, 42, myitmo.EPPlanVariantAbit, myitmo.EPFormatXLS)
		}},
		{"CompanyFile", "GET", "/api/constructor-ep/companies/7/file", func(c *myitmo.Client) (*myitmo.File, error) {
			return c.ConstructorEP.CompanyFile(ctx, 7)
		}},
		{"KUGDocument", "POST", "/api/constructor-ep/kug/42/xlsx/generate", func(c *myitmo.Client) (*myitmo.File, error) {
			return c.ConstructorEP.KUGDocument(ctx, 42, myitmo.EPFormatXLSX)
		}},
		{"KUGOrder", "POST", "/api/constructor-ep/kug/order/2025/pdf/generate", func(c *myitmo.Client) (*myitmo.File, error) {
			return c.ConstructorEP.KUGOrder(ctx, 2025)
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			f, c := newFake(t)
			f.replyWith(http.StatusOK, http.Header{
				"Content-Type":        {"application/pdf"},
				"Content-Disposition": {`attachment; filename="doc.pdf"`},
			}, "%PDF-1.4 synthetic")
			file := must[*myitmo.File](t)(tc.call(c))
			defer file.Body.Close()
			r := f.expect(tc.method, tc.path)
			if len(r.Body) != 0 {
				t.Fatalf("body = %s, want none", r.Body)
			}
			data, _ := io.ReadAll(file.Body)
			if file.Name != "doc.pdf" || file.ContentType != "application/pdf" || string(data) != "%PDF-1.4 synthetic" {
				t.Fatalf("file = %+v %q", file, data)
			}
		})
	}
}

func TestConstructorEPUploadCompanyFile(t *testing.T) {
	f, c := newFake(t)
	err := c.ConstructorEP.UploadCompanyFile(t.Context(), 7, myitmo.Upload{Name: "letter.pdf", ContentType: "application/pdf", Content: strings.NewReader("synthetic")})
	if err != nil {
		t.Fatal(err)
	}
	r := f.expect("POST", "/api/constructor-ep/companies/7/file")
	_, params, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil {
		t.Fatal(err)
	}
	part, err := multipart.NewReader(strings.NewReader(string(r.Body)), params["boundary"]).NextPart()
	if err != nil {
		t.Fatal(err)
	}
	data, _ := io.ReadAll(part)
	if part.FormName() != "file" || part.FileName() != "letter.pdf" || string(data) != "synthetic" {
		t.Fatalf("part %q %q %q", part.FormName(), part.FileName(), data)
	}
}

func TestConstructorEPDecodeList(t *testing.T) {
	f, c := newFake(t)
	f.result(`{"rows":[{"id":42,"name":"Программа","admission_year":"2024","directions":[{"id":4,"code":"09.03.01","name":"Информатика"}],
		"education_level":{"id":1,"name":"Бакалавриат"},"implementer":{"id":8,"name":"Факультет"},"manager":null,
		"education_plan_status":{"id":2,"name":"В работе"},"characteristics_status":{"id":1,"name":"Черновик"},
		"kug_status":{"id":6,"name":"Подписан"}}],"count":31,"can_create":true,"can_sign":false}`)
	page := must[*myitmo.EPProgramListPage](t)(c.ConstructorEP.List(t.Context(), myitmo.EPProgramListParams{Limit: 1}))
	row := page.Rows[0]
	if page.Count != 31 || !page.CanCreate || row.AdmissionYear != 2024 || row.Manager != nil ||
		row.Directions[0].Code != "09.03.01" || row.KUGStatus.ID != myitmo.EPStatusSigned {
		t.Fatalf("page = %+v", page)
	}
}

func TestConstructorEPDecodeInfo(t *testing.T) {
	f, c := newFake(t)
	f.result(`{"id":42,"name_ru":"Программа","name_en":"Programme","enrollment_year":{"id":6,"year":2024},
		"educational_standard":{"id":7,"name":"ОС ИТМО","education_level":{"id":1,"name":"Бакалавриат"},
		"education_form":{"id":1,"name":"Очная"},"standard_education_duration":{"duration":4},"semesters_count":8},
		"implementer":{"id":8,"name":"Факультет"},"languages":[{"id":1,"name":"Русский"}],"attributes":[],
		"educational_directions":[{"id":4,"code":"09.03.01","name":"Информатика"}],"actual_education_duration":4,
		"note":null,"abit_export":true,"exists_in_isu":false,"annotation_ru":"Текст","annotation_en":null,
		"kcp":[{"kcp_id":1,"direction_id":4,"code":"09.03.01","name":"Информатика","budget_form":25,"contract_form":null,"other_form":0,"abit_export":true,"military_department":false}],
		"managers":[{"isu":100001,"fio":"Тестов Тест","positions":[{"position_id":1,"position_name":"Руководитель","date_start":"2024-09-01","date_end":"9999-09-09"}]}],
		"companies":[{"id":7,"type_id":2,"inn":"7700000000","name_short_with_opf":"ООО Пример","name_full_with_opf":"Общество Пример","file_attached":true,"file_name":"letter","file_type":"pdf"}],
		"additional_info":{"partner_university":{"id":3,"name":"University","country":{"id":1,"name":"Страна"}},"academic_mobility_start":"2025-02-01","academic_mobility_end":null}}`)
	info := must[*myitmo.EPProgramInfo](t)(c.ConstructorEP.Info(t.Context(), 42))
	if info.EnrollmentYear.Year != 2024 || info.EducationalStandard.StandardEducationDuration.Duration != 4 ||
		*info.KCP[0].BudgetForm != 25 || info.KCP[0].ContractForm != nil ||
		info.Managers[0].Positions[0].DateEnd != myitmo.NewDate(9999, 9, 9) ||
		!info.Companies[0].FileAttached || info.AdditionalInfo.PartnerUniversity.Country.ID != 1 ||
		*info.AdditionalInfo.AcademicMobilityStart != myitmo.NewDate(2025, 2, 1) || info.AdditionalInfo.AcademicMobilityEnd != nil {
		t.Fatalf("info = %+v", info)
	}
}

func TestConstructorEPDecodeHeaderAndPlan(t *testing.T) {
	f, c := newFake(t)
	f.result(`{"id":42,"name":"Программа","is_admin":true,"is_manager":false,"can_edit":true,"archived_at":"2025-01-10T12:00:00+03:00",
		"plan_status":{"id":2},"kug_status":{"id":1},"characteristics_status":{"id":7}}`)
	h := must[*myitmo.EPProgramHeader](t)(c.ConstructorEP.Status(t.Context(), 42))
	if !h.IsAdmin || h.ArchivedAt == nil || h.ArchivedAt.Day() != 10 || h.CharacteristicsStatus.ID != myitmo.EPStatusSigning {
		t.Fatalf("header = %+v", h)
	}

	f.result(`{"name":"Программа","enrollment_year":2024,"full_capacity":240,"filled_min_capacity":200,"filled_max_capacity":250,
		"plan":[{"id":1,"type":"module","name":"Блок 1","is_root":true,"is_bank":false,"block_id":1,"rule":null,"rule_types":[1,41],
		"children_types":[1],"same_capacity_children":false,"capacity":0,"min_capacity":180,"max_capacity":200,"target_capacity":null,
		"children":[{"id":2,"type":"module","name":"Модуль","parent_id":1,"rule":{"type_id":1,"type_name":"Выбор","rule_value":[1]},
		"children":[{"id":101,"type":"discipline","name":"Дисциплина","module_id":2,"status":{"id":6,"name":"Подписан"},
		"implementer":{"name":"Факультет","short_name":"Ф"},"type_id":1,"semesters":[1,2],"semesters_count":2,"language_id":4,"language_code":"RU","capacity":6}]}]}]}`)
	plan := must[*myitmo.EPPlan](t)(c.ConstructorEP.Plan(t.Context(), 42))
	d := plan.Plan[0].Children[0].Children[0]
	if plan.EnrollmentYear != 2024 || *plan.FullCapacity != 240 || plan.Plan[0].ParentID != nil ||
		plan.Plan[0].Children[0].Rule.TypeID != myitmo.EPRuleChooseN || d.Type != myitmo.EPNodeDiscipline ||
		d.Implementer.ShortName != "Ф" || d.Semesters[1] != 2 || d.Capacity != 6 {
		t.Fatalf("plan = %+v", plan)
	}
}

func TestConstructorEPDecodeMisc(t *testing.T) {
	f, c := newFake(t)
	ctx := t.Context()

	f.result(`{"rows":[{"task_id":5,"name":"План","type":"pdf","status":"SIGNED","signatures":[{"signature_id":9}]}]}`)
	tasks := must[*myitmo.EPTaskList](t)(c.ConstructorEP.Tasks(ctx, myitmo.EPTasksParams{}))
	if tasks.Rows[0].Signatures[0].SignatureID != 9 {
		t.Fatalf("tasks = %+v", tasks)
	}

	f.result(`[{"id":1,"name":"ОС ИТМО","is_active":1,"education_level":{"id":1,"name":"Бакалавриат"},"semesters_count":8},{"id":2,"name":"Старый","is_active":false}]`)
	std := must[[]myitmo.EPStandard](t)(c.ConstructorEP.Standards(ctx))
	if !std[0].IsActive || std[1].IsActive || std[0].SemestersCount != 8 {
		t.Fatalf("standards = %+v", std)
	}

	f.result(`{"disciplines":[{"id":101,"name":"Дисциплина","annotation_ru":"","capacity":3,"implementer_name":"Факультет","semesters_count":1,"status_id":6,"status_name":"Подписан","education_levels":[{"id":1,"name":"Бакалавриат"}]}]}`)
	ds := must[[]myitmo.EPDisciplineSearchItem](t)(c.ConstructorEP.Disciplines(ctx, myitmo.EPDisciplineSearchParams{Query: "д"}))
	if len(ds) != 1 || ds[0].EducationLevels[0].ID != 1 {
		t.Fatalf("disciplines = %+v", ds)
	}

	f.result(`true`)
	if !must[bool](t)(c.ConstructorEP.IsAdmin(ctx)) {
		t.Fatal("IsAdmin = false")
	}

	f.result(`17`)
	if n := must[int64](t)(c.ConstructorEP.Create(ctx, myitmo.EPProgramForm{})); n != 17 {
		t.Fatalf("Create = %d", n)
	}

	f.result(`[{"id":1,"part_id":2,"comment":"Текст","fio":"Тестов Тест","user_isu":100001}]`)
	cm := must[[]myitmo.EPComment](t)(c.ConstructorEP.Comments(ctx, 42))
	if cm[0].PartID != myitmo.EPPartPlan || cm[0].UserISU != 100001 {
		t.Fatalf("comments = %+v", cm)
	}

	f.result(`[{"id":2,"name":"Льгота","order":1,"file_allowed":true}]`)
	if it := must[[]myitmo.EPITType](t)(c.ConstructorEP.ITTypes(ctx)); !it[0].FileAllowed {
		t.Fatalf("it types = %+v", it)
	}

	f.result(`[{"id":9,"name":"Теоретическое обучение","color":"#aabbcc","visible":true}]`)
	if at := must[[]myitmo.EPActivityType](t)(c.ConstructorEP.ActivityTypes(ctx)); at[0].ID != myitmo.EPActivityTheory || !at[0].Visible {
		t.Fatalf("activity types = %+v", at)
	}

	f.result(`{"1":12,"3":4}`)
	if raw := must[myitmo.RawJSON](t)(c.ConstructorEP.PlanPracticeCredits(ctx, 42)); string(raw) != `{"1":12,"3":4}` {
		t.Fatalf("raw = %s", raw)
	}
}
