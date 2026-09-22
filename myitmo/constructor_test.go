package myitmo_test

import (
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/kewldan/go-itmo/myitmo"
)

type constructorCase struct {
	name   string
	call   func(c *myitmo.Client) error
	method string
	path   string
	query  url.Values
	body   string // expected JSON body; empty means no body is checked
}

func constructorIgnore[T any](_ T, err error) error { return err }

func TestConstructorRoutes(t *testing.T) {
	cs := func(c *myitmo.Client) *myitmo.ConstructorService { return c.Constructor }
	editor := myitmo.ConstructorEditorJS{Blocks: []myitmo.ConstructorEditorJSBlock{{Type: "math", Data: []byte(`{"math":"x^2"}`)}}}
	form := myitmo.ConstructorDisciplineForm{NameRU: "Дисциплина", NameEN: "Discipline", Abbreviation: "D", FinAttrID: 1, LanguageID: 2, FormatID: 1, ImplementerID: 5, BarsRealization: true, EducationLevels: []int64{3}}
	formJSON := `"name_ru":"Дисциплина","name_en":"Discipline","abbreviation":"D","fin_attr_id":1,"language_id":2,"format_id":1,"implementer_id":5,"bars_realization":true,"education_levels":[3]`
	pform := myitmo.ConstructorPracticeForm{NameRU: "Практика", NameEN: "Practice", FinAttrID: 1, Abbreviation: "P", LanguageID: 2, FormatID: 1, ImplementerID: 5, EducationLevels: []int64{3}}
	pformJSON := `"name_ru":"Практика","name_en":"Practice","fin_attr_id":1,"abbreviation":"P","language_id":2,"format_id":1,"implementer_id":5,"education_levels":[3]`
	cert := myitmo.ConstructorCertificationForm{Abbreviation: "G", LanguageID: 2, FinAttrID: 1, ImplementerID: 5, EducationLevels: []int64{3}, FormatID: 1}
	certJSON := `{"abbreviation":"G","language_id":2,"fin_attr_id":1,"implementer_id":5,"education_levels":[3],"format_id":1}`
	up := func() myitmo.Upload {
		return myitmo.Upload{Name: "a.pdf", ContentType: "application/pdf", Content: strings.NewReader("%PDF")}
	}
	p := "/api/constructor/programs/10"
	cn := p + "/contents/20"
	ref := func(p string) string { return "/api/constructor/references/" + p }

	cases := []constructorCase{
		{"Programs", func(c *myitmo.Client) error {
			return constructorIgnore(cs(c).Programs(ctx, myitmo.ConstructorProgramListParams{Limit: 15, Query: "math", ProgramTypeID: myitmo.Ptr[int64](1), StatusID: myitmo.Ptr[int64](4), Archive: true, LanguageID: myitmo.Ptr[int64](2)}))
		}, "GET", "/api/constructor/programs/list", url.Values{"limit": {"15"}, "offset": {"0"}, "query": {"math"}, "program_type_id": {"1"}, "my_disciplines": {"0"}, "archive": {"1"}, "status_id": {"4"}, "language_id": {"2"}, "dpo": {"0"}}, ""},
		{"ExpertisePrograms", func(c *myitmo.Client) error {
			return constructorIgnore(cs(c).ExpertisePrograms(ctx, myitmo.ConstructorExpertiseListParams{Limit: 10, Offset: 10, MyDisciplines: true, DPO: true}))
		}, "GET", "/api/constructor/programs/expertise/list", url.Values{"limit": {"10"}, "offset": {"10"}, "my_disciplines": {"1"}, "dpo": {"1"}}, ""},
		{"SignTasks", func(c *myitmo.Client) error {
			return constructorIgnore(cs(c).SignTasks(ctx, myitmo.ConstructorSignTaskParams{Query: "x", ImplementerID: myitmo.Ptr[int64](7)}))
		}, "GET", "/api/constructor/programs/tasks_list", url.Values{"query": {"x"}, "implementer_id": {"7"}}, ""},
		{"Changes", func(c *myitmo.Client) error {
			return constructorIgnore(cs(c).Changes(ctx, 15, 30, myitmo.Ptr[int64](2), ""))
		},
			"GET", "/api/constructor/programs/changes", url.Values{"limit": {"15"}, "offset": {"30"}, "program_type_id": {"2"}}, ""},
		{"ProgramChanges", func(c *myitmo.Client) error {
			return constructorIgnore(cs(c).ProgramChanges(ctx, 10, 15, 0, myitmo.Ptr[int64](3)))
		},
			"GET", p + "/changes", url.Values{"limit": {"15"}, "offset": {"0"}, "part": {"3"}}, ""},
		{"Status", func(c *myitmo.Client) error { return constructorIgnore(cs(c).Status(ctx, 10)) }, "GET", p + "/status", nil, ""},
		{"SetStatus", func(c *myitmo.Client) error { return cs(c).SetStatus(ctx, 10, myitmo.ConstructorToExamination) }, "POST", p + "/status/to_examination", nil, ""},
		{"SetStatusArchive", func(c *myitmo.Client) error { return cs(c).SetStatus(ctx, 10, myitmo.ConstructorArchive) }, "POST", p + "/status/archive", nil, ""},
		{"Copy", func(c *myitmo.Client) error { return constructorIgnore(cs(c).Copy(ctx, 10)) }, "POST", p + "/copy", nil, ""},
		{"ProgramJSON", func(c *myitmo.Client) error { return constructorIgnore(cs(c).ProgramJSON(ctx, 10)) }, "GET", p + "/json", nil, ""},
		{"UploadFile", func(c *myitmo.Client) error { return cs(c).UploadFile(ctx, 10, up()) }, "POST", p + "/files", nil, ""},
		{"AddDeveloper", func(c *myitmo.Client) error {
			return cs(c).AddDeveloper(ctx, 10, myitmo.ConstructorPersonRef{ISU: myitmo.Ptr[int64](100001), FIO: "Иванов И. И."})
		}, "POST", p + "/developers", nil, `{"isu":100001,"fio":"Иванов И. И."}`},
		{"RemoveDeveloper", func(c *myitmo.Client) error { return cs(c).RemoveDeveloper(ctx, 10, 5) }, "DELETE", p + "/developers/5", nil, ""},
		{"AddAuthor", func(c *myitmo.Client) error {
			return cs(c).AddAuthor(ctx, 10, myitmo.ConstructorPersonRef{FIO: "Внешний автор"})
		}, "POST", p + "/authors", nil, `{"isu":null,"fio":"Внешний автор"}`},
		{"RemoveAuthor", func(c *myitmo.Client) error { return cs(c).RemoveAuthor(ctx, 10, 6) }, "DELETE", p + "/authors/6", nil, ""},
		{"AddExpert", func(c *myitmo.Client) error { return cs(c).AddExpert(ctx, 10, 100002) }, "POST", p + "/experts", nil, `{"isu":100002}`},
		{"SetExperts", func(c *myitmo.Client) error { return cs(c).SetExperts(ctx, 10, []int64{100002, 100003}) }, "PATCH", p + "/experts", nil, `[100002,100003]`},
		{"RemoveExpert", func(c *myitmo.Client) error { return cs(c).RemoveExpert(ctx, 10, 7) }, "DELETE", p + "/experts/7", nil, ""},
		{"Comments", func(c *myitmo.Client) error { return constructorIgnore(cs(c).Comments(ctx, 10)) }, "GET", p + "/comments", nil, ""},
		{"AddComment", func(c *myitmo.Client) error { return cs(c).AddComment(ctx, 10, 3, "Исправить") }, "POST", p + "/comments", nil, `{"program_part_id":3,"comment":"Исправить"}`},
		{"UpdateComment", func(c *myitmo.Client) error { return cs(c).UpdateComment(ctx, 10, 8, "ok") }, "PATCH", p + "/comments/8", nil, `{"comment":"ok"}`},
		{"DeleteComment", func(c *myitmo.Client) error { return cs(c).DeleteComment(ctx, 10, 8) }, "DELETE", p + "/comments/8", nil, ""},
		{"SearchPeople", func(c *myitmo.Client) error { return constructorIgnore(cs(c).SearchPeople(ctx, "Иван ов")) },
			"GET", "/api/constructor/people/search/%D0%98%D0%B2%D0%B0%D0%BD%20%D0%BE%D0%B2", nil, ""},
		{"Role", func(c *myitmo.Client) error { return constructorIgnore(cs(c).Role(ctx)) }, "GET", "/api/constructor/people/roles", nil, ""},

		{"CreateDiscipline", func(c *myitmo.Client) error {
			return constructorIgnore(cs(c).CreateDiscipline(ctx, myitmo.ConstructorDisciplineCreate{ConstructorDisciplineForm: form, Contents: []myitmo.ConstructorPartCapacity{{Capacity: 3}}}))
		}, "POST", "/api/constructor/disciplines/create", nil, `{` + formJSON + `,"contents":[{"capacity":3}]}`},
		{"CreateDisciplineDPO", func(c *myitmo.Client) error {
			f := form
			f.ISUPublish = true
			return constructorIgnore(cs(c).CreateDiscipline(ctx, myitmo.ConstructorDisciplineCreate{ConstructorDisciplineForm: f}))
		}, "POST", "/api/constructor/disciplines/create", nil, `{` + formJSON + `,"isu_publish":true,"contents":null}`},
		{"CreateDPO", func(c *myitmo.Client) error { return constructorIgnore(cs(c).CreateDPO(ctx, 10)) }, "POST", "/api/constructor/disciplines/10/create_dpo", nil, ""},
		{"Discipline", func(c *myitmo.Client) error { return constructorIgnore(cs(c).Discipline(ctx, 10)) }, "GET", "/api/constructor/disciplines/10/info", nil, ""},
		{"UpdateDiscipline", func(c *myitmo.Client) error { return cs(c).UpdateDiscipline(ctx, 10, form) }, "PATCH", "/api/constructor/disciplines/10/info", nil, `{` + formJSON + `}`},
		{"UpdateDisciplineDescription", func(c *myitmo.Client) error {
			return cs(c).UpdateDisciplineDescription(ctx, 10, myitmo.ConstructorDisciplineDescription{AnnotationRU: "а", AnnotationEN: "a", AdditionalInfo: "i"})
		}, "PATCH", "/api/constructor/disciplines/10/description", nil, `{"annotation_ru":"а","annotation_en":"a","additional_info":"i"}`},
		{"SetDisciplineCapacity", func(c *myitmo.Client) error {
			return cs(c).SetDisciplineCapacity(ctx, 10, []myitmo.ConstructorCapacityPart{{Capacity: 3, ContentOrder: 1,
				IntermediateAssessments: []myitmo.ConstructorWorkTypeRef{{WorkTypeID: 4}},
				WorkTypes:               []myitmo.ConstructorWorkTypeHours{{WorkTypeID: 1, Hours: 16}}}})
		}, "PATCH", "/api/constructor/disciplines/10/capacity", nil, `[{"capacity":3,"content_order":1,"intermediate_assessments":[{"work_type_id":4,"hours":null}],"work_types":[{"work_type_id":1,"hours":16}]}]`},
		{"ImportDiscipline", func(c *myitmo.Client) error { return cs(c).ImportDiscipline(ctx, 9, 10) }, "POST", "/api/constructor/disciplines/9/10", nil, ""},
		{"DisciplineUsage", func(c *myitmo.Client) error { return constructorIgnore(cs(c).DisciplineUsage(ctx, 10)) }, "GET", "/api/constructor/disciplines/10/use", nil, ""},
		{"SetIncomingTags", func(c *myitmo.Client) error { return cs(c).SetIncomingTags(ctx, 10, nil) }, "PATCH", p + "/incoming_tags", nil, `[]`},
		{"SetOutcomingTags", func(c *myitmo.Client) error {
			return cs(c).SetOutcomingTags(ctx, 10, []myitmo.ConstructorOutcomeRef{{ID: 1, TypeID: 2}})
		}, "PATCH", p + "/outcoming_tags", nil, `[{"id":1,"type_id":2}]`},
		{"TagTypes", func(c *myitmo.Client) error { return constructorIgnore(cs(c).TagTypes(ctx)) }, "GET", "/api/constructor/tags/types", nil, ""},
		{"SearchTags", func(c *myitmo.Client) error { return constructorIgnore(cs(c).SearchTags(ctx, "алг")) }, "GET", "/api/constructor/tags/search", url.Values{"query": {"алг"}}, ""},
		{"CreateTag", func(c *myitmo.Client) error { return constructorIgnore(cs(c).CreateTag(ctx, "графы")) }, "POST", "/api/constructor/tags/create", nil, `{"name":"графы"}`},
		{"ITTypes", func(c *myitmo.Client) error { return constructorIgnore(cs(c).ITTypes(ctx)) }, "GET", "/api/constructor/it_types", nil, ""},
		{"SearchCompanies", func(c *myitmo.Client) error { return constructorIgnore(cs(c).SearchCompanies(ctx, "ООО")) }, "GET", "/api/constructor/companies", url.Values{"query": {"ООО"}}, ""},
		{"AddCompany", func(c *myitmo.Client) error {
			return cs(c).AddCompany(ctx, 10, myitmo.ConstructorCompanyRef{INN: "7700000000", NameShortWithOPF: "ООО Ромашка", NameFullWithOPF: "Общество Ромашка"}, 2)
		}, "POST", p + "/companies", nil, `{"inn":"7700000000","name_short_with_opf":"ООО Ромашка","name_full_with_opf":"Общество Ромашка","type_id":2}`},
		{"RemoveCompany", func(c *myitmo.Client) error { return cs(c).RemoveCompany(ctx, 10, 4) }, "DELETE", p + "/companies/4", nil, ""},
		{"UploadCompanyFile", func(c *myitmo.Client) error { return cs(c).UploadCompanyFile(ctx, 4, up()) }, "POST", "/api/constructor/companies/4/file", nil, ""},
		{"DeleteCompanyFile", func(c *myitmo.Client) error { return cs(c).DeleteCompanyFile(ctx, 4) }, "DELETE", "/api/constructor/companies/4/file", nil, ""},

		{"Chapters", func(c *myitmo.Client) error { return constructorIgnore(cs(c).Chapters(ctx, 10, 20)) }, "GET", cn + "/chapters", nil, ""},
		{"CreateChapter", func(c *myitmo.Client) error {
			return cs(c).CreateChapter(ctx, 10, 20, myitmo.ConstructorChapterWrite{Name: "Введение", Order: 1,
				Themes:           []myitmo.ConstructorTheme{{Name: "Тема", Order: 1, Resources: []myitmo.ConstructorResource{{Link: "https://example.org"}}}},
				ProgramWorkTypes: []myitmo.ConstructorProgramWorkTypeHours{{ProgramWorkTypeID: 30, Hours: 4}}})
		}, "POST", cn + "/chapters/create", nil, `{"name":"Введение","order":1,"themes":[{"name":"Тема","order":1,"resources":[{"name":null,"link":"https://example.org"}]}],"program_work_types":[{"program_work_type_id":30,"hours":4}]}`},
		{"UpdateChapter", func(c *myitmo.Client) error {
			return cs(c).UpdateChapter(ctx, 10, 20, 31, myitmo.ConstructorChapterWrite{Name: "X", Order: 2, Themes: []myitmo.ConstructorTheme{}, ProgramWorkTypes: []myitmo.ConstructorProgramWorkTypeHours{}})
		}, "PATCH", cn + "/chapters/31", nil, `{"name":"X","order":2,"themes":[],"program_work_types":[]}`},
		{"DeleteChapter", func(c *myitmo.Client) error { return cs(c).DeleteChapter(ctx, 10, 20, 31) }, "DELETE", cn + "/chapters/31", nil, ""},
		{"SetContentCapacity", func(c *myitmo.Client) error {
			return cs(c).SetContentCapacity(ctx, 10, 20, myitmo.ConstructorContentCapacity{Capacity: 4, WorkTypes: []myitmo.ConstructorWorkTypeHours{{WorkTypeID: 22, Hours: 100}}})
		}, "PATCH", cn + "/capacity", nil, `{"capacity":4,"work_types":[{"work_type_id":22,"hours":100}]}`},
		{"SplitIndependentWork", func(c *myitmo.Client) error { return cs(c).SplitIndependentWork(ctx, 10, 20) }, "POST", cn + "/capacity/split", nil, ""},
		{"Assessments", func(c *myitmo.Client) error { return constructorIgnore(cs(c).Assessments(ctx, 10, 20)) }, "GET", cn + "/work_types", nil, ""},
		{"UpdateIntermediateAssessment", func(c *myitmo.Client) error {
			return cs(c).UpdateIntermediateAssessment(ctx, 10, 20, 40, myitmo.ConstructorIntermediateAssessmentWrite{Description: editor, MinScore: 20, MaxScore: 40})
		}, "PATCH", cn + "/work_types/40", nil, `{"description":{"blocks":[{"type":"math","data":{"math":"x^2"}}]},"min_score":20,"max_score":40}`},
		{"DeleteIntermediateAssessment", func(c *myitmo.Client) error { return cs(c).DeleteIntermediateAssessment(ctx, 10, 20, 40) }, "DELETE", cn + "/work_types/40", nil, ""},
		{"ToggleAdditionalScore", func(c *myitmo.Client) error { return cs(c).ToggleAdditionalScore(ctx, 10, 20) }, "PATCH", cn + "/work_types/extra", nil, ""},
		{"CreateCurrentAssessment", func(c *myitmo.Client) error {
			return cs(c).CreateCurrentAssessment(ctx, 10, 20, myitmo.ConstructorCurrentAssessmentWrite{Description: editor, Name: "Тест", Order: 1, ControlPoint: true, MinScore: 5, MaxScore: 10, CurrentAssessmentID: 2, ChaptersID: []int64{31}})
		}, "POST", cn + "/assessments/current/create", nil, `{"description":{"blocks":[{"type":"math","data":{"math":"x^2"}}]},"name":"Тест","order":1,"control_point":true,"min_score":5,"max_score":10,"current_assessment_id":2,"chapters_id":[31]}`},
		{"UpdateCurrentAssessment", func(c *myitmo.Client) error {
			return cs(c).UpdateCurrentAssessment(ctx, 10, 20, 41, myitmo.ConstructorCurrentAssessmentWrite{Description: editor, Name: "Т", ChaptersID: []int64{}})
		}, "PATCH", cn + "/assessments/current/41", nil, `{"description":{"blocks":[{"type":"math","data":{"math":"x^2"}}]},"name":"Т","order":0,"control_point":false,"min_score":0,"max_score":0,"current_assessment_id":0,"chapters_id":[]}`},
		{"DeleteCurrentAssessment", func(c *myitmo.Client) error { return cs(c).DeleteCurrentAssessment(ctx, 10, 20, 41) }, "DELETE", cn + "/assessments/current/41", nil, ""},

		{"Sources", func(c *myitmo.Client) error { return constructorIgnore(cs(c).Sources(ctx, 10)) }, "GET", p + "/sources", nil, ""},
		{"AddLibrarySources", func(c *myitmo.Client) error { return cs(c).AddLibrarySources(ctx, 10, []int64{55}) }, "POST", p + "/sources", nil, `[55]`},
		{"RemoveLibrarySource", func(c *myitmo.Client) error { return cs(c).RemoveLibrarySource(ctx, 10, 55) }, "DELETE", p + "/sources/55", nil, ""},
		{"AddOtherSource", func(c *myitmo.Client) error { return cs(c).AddOtherSource(ctx, 10, "Курс") }, "POST", p + "/sources/other", nil, `{"name":"Курс"}`},
		{"RemoveOtherSource", func(c *myitmo.Client) error { return cs(c).RemoveOtherSource(ctx, 10, 56) }, "DELETE", p + "/sources/other/56", nil, ""},
		{"SearchLibrary", func(c *myitmo.Client) error { return constructorIgnore(cs(c).SearchLibrary(ctx, "a/b")) }, "GET", "/api/constructor/sources/library/search/a%2Fb", nil, ""},

		{"CreatePractice", func(c *myitmo.Client) error {
			return constructorIgnore(cs(c).CreatePractice(ctx, myitmo.ConstructorPracticeCreate{ConstructorPracticeForm: pform, Contents: []myitmo.ConstructorPartCapacity{{Capacity: 6}}}))
		}, "POST", "/api/constructor/practices/create", nil, `{` + pformJSON + `,"contents":[{"capacity":6}]}`},
		{"UpdatePractice", func(c *myitmo.Client) error { return cs(c).UpdatePractice(ctx, 11, pform) }, "PATCH", "/api/constructor/disciplines/11/info", nil, `{` + pformJSON + `}`},
		{"Practice", func(c *myitmo.Client) error { return constructorIgnore(cs(c).Practice(ctx, 11)) }, "GET", "/api/constructor/practices/11/info", nil, ""},
		{"SetPracticeCapacity", func(c *myitmo.Client) error {
			return cs(c).SetPracticeCapacity(ctx, 11, []myitmo.ConstructorPracticePart{{Capacity: 6, ContentOrder: 1, IntermediateAssessments: []myitmo.ConstructorWorkTypeRef{{WorkTypeID: 5}}}})
		}, "PATCH", "/api/constructor/practices/11/capacity", nil, `[{"capacity":6,"content_order":1,"intermediate_assessments":[{"work_type_id":5,"hours":null}]}]`},
		{"UpdatePracticeAbout", func(c *myitmo.Client) error {
			return cs(c).UpdatePracticeAbout(ctx, 11, myitmo.ConstructorPracticeAbout{FormID: 1, TypeID: 2, PassTypeID: 3, PracticeFormatID: 4})
		}, "PATCH", "/api/constructor/practices/11/about", nil, `{"form_id":1,"type_id":2,"pass_type_id":3,"practice_format_id":4}`},
		{"UpdatePracticeDescription", func(c *myitmo.Client) error {
			return cs(c).UpdatePracticeDescription(ctx, 11, "особенности")
		},
			"PATCH", "/api/constructor/practices/11/description", nil, `{"additional_description":"особенности"}`},
		{"PracticeTypes", func(c *myitmo.Client) error { return constructorIgnore(cs(c).PracticeTypes(ctx, 11)) }, "GET", "/api/constructor/practices/11/types", nil, ""},
		{"PracticeContent", func(c *myitmo.Client) error { return constructorIgnore(cs(c).PracticeContent(ctx, 11)) }, "GET", "/api/constructor/practices/11/content", nil, ""},
		{"UpdatePracticeContent", func(c *myitmo.Client) error {
			return cs(c).UpdatePracticeContent(ctx, 11, myitmo.ConstructorPracticeContentWrite{Content: editor})
		}, "PATCH", "/api/constructor/practices/11/content", nil, `{"additional_content":null,"content":{"blocks":[{"type":"math","data":{"math":"x^2"}}]}}`},
		{"UpdatePracticeReport", func(c *myitmo.Client) error { return cs(c).UpdatePracticeReport(ctx, 11, "отчёт") }, "PATCH", "/api/constructor/practices/11/report", nil, `{"report":"отчёт"}`},
		{"PracticeAssessments", func(c *myitmo.Client) error { return constructorIgnore(cs(c).PracticeAssessments(ctx, 11)) }, "GET", "/api/constructor/practices/11/assessments", nil, ""},
		{"UpdatePracticeGrades", func(c *myitmo.Client) error {
			return cs(c).UpdatePracticeGrades(ctx, 11, []myitmo.ConstructorGradeDescription{{ID: 1, Description: "d"}})
		}, "PATCH", "/api/constructor/practices/11/grades", nil, `[{"id":1,"description":"d"}]`},

		{"CreateCertification", func(c *myitmo.Client) error { return constructorIgnore(cs(c).CreateCertification(ctx, cert)) }, "POST", "/api/constructor/certifications/create", nil, certJSON},
		{"Certification", func(c *myitmo.Client) error { return constructorIgnore(cs(c).Certification(ctx, 12)) }, "GET", "/api/constructor/certifications/12/info", nil, ""},
		{"UpdateCertification", func(c *myitmo.Client) error { return cs(c).UpdateCertification(ctx, 12, cert) }, "PATCH", "/api/constructor/certifications/12/info", nil, certJSON},
		{"PreparationStages", func(c *myitmo.Client) error { return constructorIgnore(cs(c).PreparationStages(ctx, 12)) }, "GET", "/api/constructor/certifications/12/preparation_stages", nil, ""},
		{"UpdatePreparationStages", func(c *myitmo.Client) error {
			return cs(c).UpdatePreparationStages(ctx, 12, []myitmo.ConstructorPreparationStageUpdate{{ID: 1, AdditionalDescription: myitmo.Ptr("a<br>b")}, {ID: 2}})
		}, "PATCH", "/api/constructor/certifications/12/preparation_stages", nil, `[{"id":1,"additional_description":"a<br>b"},{"id":2,"additional_description":null}]`},
		{"MainStages", func(c *myitmo.Client) error { return constructorIgnore(cs(c).MainStages(ctx, 12)) }, "GET", "/api/constructor/certifications/12/main_stages", nil, ""},
		{"UpdateMainStages", func(c *myitmo.Client) error {
			return cs(c).UpdateMainStages(ctx, 12, []myitmo.ConstructorMainStageUpdate{{ID: 1, Deadline: "до 1 июня"}})
		}, "PATCH", "/api/constructor/certifications/12/main_stages", nil, `[{"id":1,"deadline":"до 1 июня"}]`},
		{"CertificationAssessments", func(c *myitmo.Client) error { return constructorIgnore(cs(c).CertificationAssessments(ctx, 12, 1)) },
			"GET", "/api/constructor/certifications/12/assessments/types/1", nil, ""},
		{"UpdateCertificationAssessment", func(c *myitmo.Client) error {
			return cs(c).UpdateCertificationAssessment(ctx, 12, 50, []myitmo.ConstructorCertGrade{{ID: 1, Name: "отлично", Description: "d"}})
		}, "PATCH", "/api/constructor/certifications/12/assessments/50", nil, `[{"id":1,"name":"отлично","description":"d"}]`},

		{"Statuses", func(c *myitmo.Client) error { return constructorIgnore(cs(c).Statuses(ctx)) }, "GET", ref("statuses"), nil, ""},
		{"ProgramTypes", func(c *myitmo.Client) error { return constructorIgnore(cs(c).ProgramTypes(ctx)) }, "GET", ref("program_types"), nil, ""},
		{"Languages", func(c *myitmo.Client) error { return constructorIgnore(cs(c).Languages(ctx)) }, "GET", ref("languages"), nil, ""},
		{"Implementers", func(c *myitmo.Client) error { return constructorIgnore(cs(c).Implementers(ctx)) }, "GET", ref("implementers"), nil, ""},
		{"Formats", func(c *myitmo.Client) error { return constructorIgnore(cs(c).Formats(ctx)) }, "GET", ref("formats"), nil, ""},
		{"IntermediateAssessmentTypes", func(c *myitmo.Client) error { return constructorIgnore(cs(c).IntermediateAssessmentTypes(ctx)) }, "GET", ref("intermediate_assessments"), nil, ""},
		{"WorkTypes", func(c *myitmo.Client) error { return constructorIgnore(cs(c).WorkTypes(ctx)) }, "GET", ref("work_types"), nil, ""},
		{"EducationLevels", func(c *myitmo.Client) error { return constructorIgnore(cs(c).EducationLevels(ctx)) }, "GET", ref("education_levels"), nil, ""},
		{"CurrentAssessmentTypes", func(c *myitmo.Client) error { return constructorIgnore(cs(c).CurrentAssessmentTypes(ctx)) }, "GET", ref("current_assessments"), nil, ""},
		{"MaxContents", func(c *myitmo.Client) error { return constructorIgnore(cs(c).MaxContents(ctx)) }, "GET", ref("max_contents"), nil, ""},
		{"FinAttrs", func(c *myitmo.Client) error { return constructorIgnore(cs(c).FinAttrs(ctx)) }, "GET", ref("fin_attrs"), nil, ""},
		{"FinAttrImplementers", func(c *myitmo.Client) error { return constructorIgnore(cs(c).FinAttrImplementers(ctx, 3)) }, "GET", ref("fin_attrs/3/implementers"), nil, ""},
		{"CertificationAssessmentTypes", func(c *myitmo.Client) error { return constructorIgnore(cs(c).CertificationAssessmentTypes(ctx)) }, "GET", ref("certification_assessment_types"), nil, ""},
		{"PracticeForms", func(c *myitmo.Client) error { return constructorIgnore(cs(c).PracticeForms(ctx)) }, "GET", ref("practice_forms"), nil, ""},
		{"PracticePassTypes", func(c *myitmo.Client) error { return constructorIgnore(cs(c).PracticePassTypes(ctx)) }, "GET", ref("practice_pass_types"), nil, ""},
		{"PracticeFormats", func(c *myitmo.Client) error { return constructorIgnore(cs(c).PracticeFormats(ctx)) }, "GET", ref("practice_formats"), nil, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			f, c := newFake(t)
			if err := tc.call(c); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			r := f.expect(tc.method, tc.path)
			if tc.query != nil {
				if got, want := r.Query.Encode(), tc.query.Encode(); got != want {
					t.Errorf("query = %s, want %s", got, want)
				}
			} else if len(r.Query) != 0 {
				t.Errorf("unexpected query %v", r.Query)
			}
			if tc.body != "" {
				r.sameJSON(t, tc.body)
			}
		})
	}
}

func TestConstructorUploads(t *testing.T) {
	f, c := newFake(t)
	must[any](t)(nil, c.Constructor.UploadFile(ctx, 10, myitmo.Upload{Field: "ignored", Name: "scan.pdf", Content: strings.NewReader("%PDF-1.7")}))
	r := f.expect(http.MethodPost, "/api/constructor/programs/10/files")
	if !strings.Contains(string(r.Body), `name="file"; filename="scan.pdf"`) || !strings.Contains(string(r.Body), "%PDF-1.7") {
		t.Errorf("multipart body = %q", r.Body)
	}
	if !strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data") {
		t.Errorf("Content-Type = %q", r.Header.Get("Content-Type"))
	}

	f.result(`"https://example.org/static/img.png"`)
	u := must[string](t)(c.Constructor.UploadImage(ctx, myitmo.Upload{Name: "img.png", ContentType: "image/png", Content: strings.NewReader("PNG")}))
	r = f.expect(http.MethodPost, "/api/constructor/static/image")
	if u != "https://example.org/static/img.png" || !strings.Contains(string(r.Body), `name="image"; filename="img.png"`) {
		t.Errorf("url = %q, body = %q", u, r.Body)
	}

	must[any](t)(nil, c.Constructor.UploadCompanyFile(ctx, 4, myitmo.Upload{Name: "letter.pdf", Content: strings.NewReader("x")}))
	r = f.expect(http.MethodPost, "/api/constructor/companies/4/file")
	if !strings.Contains(string(r.Body), `name="file"; filename="letter.pdf"`) {
		t.Errorf("multipart body = %q", r.Body)
	}
}

func TestConstructorDownloads(t *testing.T) {
	f, c := newFake(t)
	cases := []struct {
		path string
		get  func() (*myitmo.File, error)
	}{
		{"/api/constructor/programs/10/document/docx", func() (*myitmo.File, error) {
			return c.Constructor.Document(ctx, 10, myitmo.ConstructorDOCX)
		}},
		{"/api/constructor/programs/10/document/pdf", func() (*myitmo.File, error) {
			return c.Constructor.Document(ctx, 10, myitmo.ConstructorPDF)
		}},
		{"/api/constructor/programs/10/signed", func() (*myitmo.File, error) { return c.Constructor.SignedDocument(ctx, 10) }},
		{"/api/constructor/companies/4/file", func() (*myitmo.File, error) { return c.Constructor.CompanyFile(ctx, 4) }},
	}
	for _, tc := range cases {
		f.replyWith(http.StatusOK, http.Header{"Content-Type": {"application/pdf"}, "Content-Disposition": {`attachment; filename="doc.pdf"`}}, "%PDF")
		file := must[*myitmo.File](t)(tc.get())
		body, _ := io.ReadAll(file.Body)
		file.Body.Close()
		f.expect(http.MethodGet, tc.path)
		if string(body) != "%PDF" || file.Name != "doc.pdf" || file.ContentType != "application/pdf" {
			t.Errorf("%s: file = %+v, body = %q", tc.path, file, body)
		}
	}
	f.reply(http.StatusNotFound, `{"error_code":404,"error_message":"Нет файла","result":null}`)
	if _, err := c.Constructor.SignedDocument(ctx, 10); err == nil {
		t.Error("want error for 404")
	}
}

func TestConstructorDecodeLists(t *testing.T) {
	f, c := newFake(t)
	f.result(`{"count":31,"can_sign":true,"is_dpo_expert":false,"programs":[{"id":10,"name":"Алгоритмы","capacity":4.5,
		"implementer_name":"Факультет","fin_attr_name":"ФА","language_code":"RU","education_levels":[{"id":1,"name":"Бакалавриат"}],
		"developers_count":1,"developers":[{"id":3,"isu":100001,"fio":"Иванов И. И."}],"status_id":5,"status_name":"На экспертизе",
		"experts":[{"id":4,"isu":null,"fio":"Эксперт"}]}]}`)
	list := must[*myitmo.ConstructorProgramList](t)(c.Constructor.Programs(ctx, myitmo.ConstructorProgramListParams{Limit: 15}))
	it := list.Programs[0]
	if list.Count != 31 || !list.CanSign || it.Capacity != 4.5 || it.StatusID != myitmo.ConstructorStatusOnExpertise ||
		*it.Developers[0].ISU != 100001 || it.Experts[0].ISU != nil || it.EducationLevels[0].Name != "Бакалавриат" {
		t.Errorf("list = %+v", list)
	}

	f.result(`{"rows":[{"task_id":123,"name":null,"type":"pdf","status":"SIGNED","signatures":[{"signature_id":"a-b"}]}]}`)
	tasks := must[[]myitmo.ConstructorSignTask](t)(c.Constructor.SignTasks(ctx, myitmo.ConstructorSignTaskParams{}))
	if len(tasks) != 1 || tasks[0].TaskID != "123" || tasks[0].Signatures[0].SignatureID != "a-b" || tasks[0].Status != "SIGNED" {
		t.Errorf("tasks = %+v", tasks)
	}

	f.result(`{"count":1,"rows":[{"timestamp":"2026-03-01T12:30:00+03:00","program_type_name":"РПД","program_id":10,"name":"Статус",
		"actor_isu":100001,"actor_fio":"Иванов И. И.","prior_state":"Черновик","resulting_state":"Опубликована"}]}`)
	ch := must[*myitmo.ConstructorChanges](t)(c.Constructor.Changes(ctx, 15, 0, nil, ""))
	if ch.Rows[0].ProgramID != 10 || !ch.Rows[0].Timestamp.Equal(time.Date(2026, 3, 1, 9, 30, 0, 0, time.UTC)) {
		t.Errorf("changes = %+v", ch)
	}

	f.result(`[{"id":1,"program_part_id":3,"comment":"Исправить","fio":"Эксперт","created_at":"2026-03-02 10:00:00"}]`)
	cm := must[[]myitmo.ConstructorComment](t)(c.Constructor.Comments(ctx, 10))
	if cm[0].ProgramPartID != 3 || cm[0].CreatedAt.Hour() != 10 {
		t.Errorf("comments = %+v", cm)
	}

	f.result(`[{"isu":100001,"fio":"Иванов И. И."}]`)
	people := must[[]myitmo.ConstructorPersonRef](t)(c.Constructor.SearchPeople(ctx, "Иван"))
	if *people[0].ISU != 100001 {
		t.Errorf("people = %+v", people)
	}

	f.result(`2`)
	if role := must[int](t)(c.Constructor.Role(ctx)); role != 2 {
		t.Errorf("role = %d", role)
	}
	f.result(`77`)
	if id := must[int64](t)(c.Constructor.Copy(ctx, 10)); id != 77 {
		t.Errorf("copy = %d", id)
	}
	f.result(`[{"id":1,"name":"Черновик","style":"secondary"}]`)
	if st := must[[]myitmo.ConstructorStatus](t)(c.Constructor.Statuses(ctx)); st[0].Style != "secondary" {
		t.Errorf("statuses = %+v", st)
	}
	f.result(`[{"id":1,"name":"Очный","description":"В аудитории"}]`)
	if fm := must[[]myitmo.ConstructorFormat](t)(c.Constructor.Formats(ctx)); fm[0].Description != "В аудитории" {
		t.Errorf("formats = %+v", fm)
	}
	f.result(`[{"id":7,"name":"Курсовая работа","is_course_work":true}]`)
	if ia := must[[]myitmo.ConstructorIntermediateAssessmentType](t)(c.Constructor.IntermediateAssessmentTypes(ctx)); !ia[0].IsCourseWork {
		t.Errorf("intermediate = %+v", ia)
	}
	f.result(`8`)
	if n := must[int](t)(c.Constructor.MaxContents(ctx)); n != 8 {
		t.Errorf("max contents = %d", n)
	}
	f.result(`[{"id":1,"name":"знания","third_person":"знает"}]`)
	if tt := must[[]myitmo.ConstructorTagType](t)(c.Constructor.TagTypes(ctx)); tt[0].ThirdPerson != "знает" {
		t.Errorf("tag types = %+v", tt)
	}
	f.result(`[{"id":1,"name":"Аккредитованная","order":1,"file_allowed":true}]`)
	if it := must[[]myitmo.ConstructorITType](t)(c.Constructor.ITTypes(ctx)); !it[0].FileAllowed {
		t.Errorf("it types = %+v", it)
	}
	f.result(`[{"inn":"7700000000","name_short_with_opf":"ООО Ромашка","name_full_with_opf":"Общество Ромашка"}]`)
	if co := must[[]myitmo.ConstructorCompanyRef](t)(c.Constructor.SearchCompanies(ctx, "Ром")); co[0].INN != "7700000000" {
		t.Errorf("companies = %+v", co)
	}
	f.result(`[{"ep_id":1,"ep_name":"Программа","link":"https://example.org/ep/1","module_id":2,"module_name":"Модуль","semesters":[1,2]}]`)
	if use := must[[]myitmo.ConstructorDisciplineUse](t)(c.Constructor.DisciplineUsage(ctx, 10)); use[0].Semesters[1] != 2 {
		t.Errorf("use = %+v", use)
	}
	f.result(`{"library_sources":[{"id":55,"name":"Кнут Д. Искусство программирования"}],"other_sources":[]}`)
	if src := must[*myitmo.ConstructorSources](t)(c.Constructor.Sources(ctx, 10)); src.LibrarySources[0].ID != 55 {
		t.Errorf("sources = %+v", src)
	}
}

func TestConstructorDecodeProgram(t *testing.T) {
	f, c := newFake(t)
	f.result(`{"menu":[{"id":3,"name":"Основное","query":"mainRPD","order":1}],"status_id":7,"status_name":"На подписании",
		"name":"Алгоритмы","type_id":1,"can_edit":true,"is_admin":false,"is_expert":true,"can_sign":true,
		"errors":[{"text":"Не заполнено","objects":["Раздел 1"]}],"contents_count":1,
		"signature_task":{"task_id":"t-1","name":"РПД","type":"pdf","status":"NEW","signatures":[{"signature_id":9}]},"can_edit_companies":true}`)
	st := must[*myitmo.ConstructorProgramStatus](t)(c.Constructor.Status(ctx, 10))
	if st.Menu[0].Query != "mainRPD" || st.TypeID != myitmo.ConstructorTypeDiscipline || st.Errors[0].Objects[0] != "Раздел 1" ||
		st.SignatureTask == nil || st.SignatureTask.Signatures[0].SignatureID != "9" || !st.CanEditCompanies {
		t.Errorf("status = %+v", st)
	}

	f.result(`{"discipline_fields":{"annotation_ru":"Курс об алгоритмах","hours":72},"program_id":10}`)
	pj := must[*myitmo.ConstructorProgramJSON](t)(c.Constructor.ProgramJSON(ctx, 10))
	if pj.DisciplineFields.AnnotationRU != "Курс об алгоритмах" || !strings.Contains(string(pj.Unknown), "program_id") {
		t.Errorf("json = %+v", pj)
	}

	f.result(`{"id":10,"name_ru":"Алгоритмы","name_en":"Algorithms","abbreviation":"Алг","language_id":1,"language_name":"Русский",
		"implementer_id":5,"implementer_name":"Факультет","fin_attr_id":1,"fin_attr_name":"ФА","format_id":1,"format_name":"Очный",
		"education_levels":[{"id":1,"name":"Бакалавриат"}],"bars_realization":true,"is_dpo":false,"isu_publish":false,"capacity":6,
		"contents":[{"id":20,"content_order":1,"capacity":3,"intermediate_assessments":[{"work_type_id":4,"name":"Экзамен","hours":null}],
			"work_types":[{"work_type_id":1,"name":"Лекции","hours":16}]}],
		"annotation_ru":"Аннотация","annotation_en":null,"additional_info":null,
		"in_tags":[{"id":1,"name":"математика"}],"out_tags":[{"id":2,"name":"графы","type_id":1,"program_out_tag_id":9,"is_attached":true}],
		"companies":[{"id":4,"type_id":2,"inn":"7700000000","name_short_with_opf":"ООО Ромашка","name_full_with_opf":"Общество Ромашка","file_attached":true,"file_name":"letter","file_type":"pdf"}],
		"developers":[],"authors":[{"id":5,"isu":null,"fio":"Внешний автор"}],"experts":[]}`)
	d := must[*myitmo.ConstructorDisciplineInfo](t)(c.Constructor.Discipline(ctx, 10))
	ct := d.Contents[0]
	if d.Capacity != 6 || ct.ID != 20 || ct.IntermediateAssessments[0].Hours != nil || *ct.WorkTypes[0].Hours != 16 ||
		!d.OutTags[0].IsAttached || !d.Companies[0].FileAttached || d.AnnotationEN != "" || d.Authors[0].ISU != nil {
		t.Errorf("discipline = %+v", d)
	}

	f.result(`{"chapters":[{"id":31,"name":"Введение","order":1,"work_types":[{"work_type_id":1,"hours":4}],
			"themes":[{"id":1,"name":"Тема","order":1,"resources":[{"name":null,"link":"https://example.org"}]}]}],
		"work_type_limits":[{"work_type_id":1,"total":16,"program_work_type_id":30}],"capacity":3,
		"distributed_total":[{"work_type_id":1,"total":4}],"errors":[]}`)
	chs := must[*myitmo.ConstructorChapters](t)(c.Constructor.Chapters(ctx, 10, 20))
	if chs.Chapters[0].Themes[0].Resources[0].Name != nil || chs.WorkTypeLimits[0].ProgramWorkTypeID != 30 || chs.DistributedTotal[0].Total != 4 {
		t.Errorf("chapters = %+v", chs)
	}

	f.result(`{"intermediate_assessments":{"list":[{"id":40,"name":"Экзамен","min_score":20,"max_score":40,"is_course_work":false,"description":null}],"additional_score":3},
		"current_assessments":[{"id":41,"name":"Тест","order":1,"min_score":5,"max_score":10,"control_point":true,"current_assessment_id":2,
			"current_assessment_name":"Тестирование","chapters":[{"chapter_id":31,"chapter_name":"Введение"}],
			"description":{"time":1700000000000,"blocks":[{"id":"b1","type":"paragraph","data":{"text":"Описание"}}],"version":"2.28.0"}}]}`)
	as := must[*myitmo.ConstructorAssessments](t)(c.Constructor.Assessments(ctx, 10, 20))
	cur := as.CurrentAssessments[0]
	if as.IntermediateAssessments.AdditionalScore != 3 || as.IntermediateAssessments.List[0].Description != nil ||
		cur.Chapters[0].ChapterID != 31 || cur.Description.Blocks[0].Type != "paragraph" || cur.Description.Version != "2.28.0" {
		t.Errorf("assessments = %+v", as)
	}
}

func TestConstructorDecodePracticeAndCertification(t *testing.T) {
	f, c := newFake(t)
	f.result(`{"id":11,"name_ru":"Практика","name_en":"Practice","abbreviation":"П","fin_attr_id":1,"fin_attr_name":"ФА",
		"language_id":1,"language_name":"Русский","implementer_id":5,"implementer_name":"Факультет","format_id":1,
		"education_levels":[{"id":1,"name":"Бакалавриат"}],"form_id":null,"type_id":2,"pass_type_id":null,"practice_format_id":null,
		"description":"<p>Положения</p>","additional_description":null,"out_tags":[],
		"contents":[{"id":21,"content_order":1,"capacity":6,"intermediate_assessments":[],"work_types":[{"work_type_id":128,"name":"Контактная","hours":10}]}],
		"developers":[],"authors":[]}`)
	pr := must[*myitmo.ConstructorPracticeInfo](t)(c.Constructor.Practice(ctx, 11))
	if pr.FormID != nil || *pr.TypeID != 2 || pr.Contents[0].WorkTypes[0].WorkTypeID != myitmo.ConstructorWorkContactTotal {
		t.Errorf("practice = %+v", pr)
	}

	f.result(`{"content":null,"additional_content":null,"report":"Отчёт","report_description":"<p>r</p>","disabled_description":"Особые условия"}`)
	pc := must[*myitmo.ConstructorPracticeContent](t)(c.Constructor.PracticeContent(ctx, 11))
	if pc.Content != nil || pc.Report != "Отчёт" || pc.DisabledDescription != "Особые условия" {
		t.Errorf("content = %+v", pc)
	}

	f.result(`{"assessments":[{"id":1,"name":"Отлично","description":"a\nb"}]}`)
	if ga := must[[]myitmo.ConstructorGradeCriterion](t)(c.Constructor.PracticeAssessments(ctx, 11)); ga[0].Description != "a\nb" {
		t.Errorf("grades = %+v", ga)
	}

	f.result(`{"id":12,"abbreviation":"ГИА","education_levels":[{"id":1,"name":"Бакалавриат"}],"fin_attr_id":1,"fin_attr_name":"ФА",
		"language_id":1,"language_name":"Русский","implementer_id":5,"implementer_name":"Факультет","format_id":1,
		"disabled_description":"строка 1\nстрока 2","developers":[{"id":1,"isu":100001,"fio":"Иванов И. И."}],"authors":[]}`)
	ci := must[*myitmo.ConstructorCertificationInfo](t)(c.Constructor.Certification(ctx, 12))
	if ci.Abbreviation != "ГИА" || ci.Developers[0].FIO == "" {
		t.Errorf("certification = %+v", ci)
	}

	f.result(`{"stages":[{"id":1,"name":"Подготовка","description":"<p>d</p>","additional_description":null}]}`)
	if ps := must[[]myitmo.ConstructorPreparationStage](t)(c.Constructor.PreparationStages(ctx, 12)); ps[0].Name != "Подготовка" {
		t.Errorf("preparation = %+v", ps)
	}
	f.result(`{"stages":[{"id":1,"name":"Защита","deadline":"до 20 июня","is_editable":true}]}`)
	if ms := must[[]myitmo.ConstructorMainStage](t)(c.Constructor.MainStages(ctx, 12)); !ms[0].IsEditable || ms[0].Deadline != "до 20 июня" {
		t.Errorf("main stages = %+v", ms)
	}
	f.result(`{"assessments":[{"id":50,"name":"ВКР","grades":[{"id":1,"name":"отлично","description":"d"}]}]}`)
	if ca := must[[]myitmo.ConstructorCertificationAssessment](t)(c.Constructor.CertificationAssessments(ctx, 12, 1)); ca[0].Grades[0].Name != "отлично" {
		t.Errorf("cert assessments = %+v", ca)
	}
}
