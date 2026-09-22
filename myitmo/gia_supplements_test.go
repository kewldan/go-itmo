package myitmo_test

import (
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/kewldan/go-itmo/myitmo"
)

func TestGIASupplementRequests(t *testing.T) {
	status := 0
	sp := myitmo.GIASupplementParams{Limit: 25, Year: 2025, EPID: 7, FacultyID: 2, GroupID: "Z3400", Credits: 240, ZECount: 60.5,
		AttachmentStatus: &status, DiplomaStatus: "approved", ValidCredits: true, Qualification: "бакалавр", AvgGrade: "4.5",
		ShowAs: myitmo.GIARoleOSOP, Query: "тест"}
	spq := url.Values{"year": {"2025"}, "ep_id": {"7"}, "faculty_id": {"2"}, "group_id": {"Z3400"}, "credits": {"240"},
		"ze_count": {"60.5"}, "attachment_status": {"0"}, "diploma_status": {"approved"}, "valid_credits": {"1"},
		"qualification": {"бакалавр"}, "avg_grade": {"4.5"}, "show_as": {"osop"}, "query": {"тест"}}
	paged := url.Values{"limit": {"25"}, "offset": {"0"}}
	for k, v := range spq {
		paged[k] = v
	}
	runGIACases(t, []giaCase{
		{"SupplementStudents", func(c *myitmo.Client) error { _, err := c.GIA.SupplementStudents(ctx, sp); return err },
			"GET", "/api/gia/attachment/students/list", paged, ""},
		{"SupplementStudent", func(c *myitmo.Client) error { _, err := c.GIA.SupplementStudent(ctx, 11); return err },
			"GET", "/api/gia/attachment/students/11", nil, ""},
		{"SyncConfirmedDisciplines", func(c *myitmo.Client) error { return c.GIA.SyncConfirmedDisciplines(ctx, 11) },
			"POST", "/api/gia/attachment/students/11/sync_confirmed_discs", nil, ""},
		{"RollbackSupplementBlock", func(c *myitmo.Client) error {
			return c.GIA.RollbackSupplementBlock(ctx, myitmo.GIABlockDisciplines, 11)
		},
			"POST", "/api/gia/attachment/rollback/disciplines/11", nil, ""},
		{"EUSupplements", func(c *myitmo.Client) error {
			_, err := c.GIA.EUSupplements(ctx, myitmo.GIAEUSupplementsParams{Limit: 15, Year: 2025, EPID: 7, GroupID: "Z3400",
				StatusID: -1, EPLanguage: "eng", Query: "тест"})
			return err
		}, "GET", "/api/gia/attachment/eu/list", url.Values{"limit": {"15"}, "offset": {"0"}, "year": {"2025"}, "ep_id": {"7"},
			"group_id": {"Z3400"}, "status_id": {"-1"}, "ep_language": {"eng"}, "query": {"тест"}}, ""},
		{"EUSupplement", func(c *myitmo.Client) error { _, err := c.GIA.EUSupplement(ctx, 41); return err },
			"GET", "/api/gia/attachment/eu/41", nil, ""},
		{"SendEUSupplement", func(c *myitmo.Client) error { return c.GIA.SendEUSupplement(ctx, 41) },
			"POST", "/api/gia/attachment/eu/41/send", nil, ""},
		{"RevokeEUSupplement", func(c *myitmo.Client) error { return c.GIA.RevokeEUSupplement(ctx, 41) },
			"POST", "/api/gia/attachment/eu/41/revoke", nil, `{}`},
		{"RollbackEUSupplement", func(c *myitmo.Client) error { return c.GIA.RollbackEUSupplement(ctx, 41) },
			"POST", "/api/gia/attachment/eu/41/rollback", nil, `{}`},
		{"ConfirmEUSupplement", func(c *myitmo.Client) error { return c.GIA.ConfirmEUSupplement(ctx, 41) },
			"POST", "/api/gia/attachment/eu/41/confirm", nil, `{}`},
		{"NotifyEUSupplementPrinted", func(c *myitmo.Client) error { return c.GIA.NotifyEUSupplementPrinted(ctx, 41) },
			"POST", "/api/gia/attachment/eu/41/print_notification", nil, ""},
		{"UpdateEUSupplementDiploma", func(c *myitmo.Client) error {
			return c.GIA.UpdateEUSupplementDiploma(ctx, 41, myitmo.GIAEUDiplomaPatch{Surname: "Testov", Name: "Test",
				BirthPlace: "Saint Petersburg", Direction: "Informatics", EPName: "Programme", Theme: "Topic"})
		}, "PATCH", "/api/gia/attachment/eu/41/modify/diploma", nil,
			`{"surname":"Testov","name":"Test","birth_place":"Saint Petersburg","direction":"Informatics","ep_name":"Programme","theme":"Topic"}`},
		{"RenameEUSupplementDisciplines", func(c *myitmo.Client) error {
			return c.GIA.RenameEUSupplementDisciplines(ctx, 41, myitmo.GIAEUDisciplinesPatch{
				DisciplinesToPatch: []myitmo.GIAEURename{{DiscID: 101, NewName: "Mathematics"}}})
		}, "PATCH", "/api/gia/attachment/eu/41/modify/disciplines", nil,
			`{"disciplines_to_patch":[{"disc_id":101,"new_name":"Mathematics"}],"faculties_to_patch":[]}`},
	})
	runGIADownloads(t, []giaDownload{
		{"ExportSupplementStudents", func(c *myitmo.Client) (*myitmo.File, error) { return c.GIA.ExportSupplementStudents(ctx, sp) },
			"GET", "/api/gia/attachment/students/file", spq, ""},
		{"EUSupplementFile", func(c *myitmo.Client) (*myitmo.File, error) { return c.GIA.EUSupplementFile(ctx, 41) },
			"GET", "/api/gia/attachment/eu/41/file", nil, ""},
	})
}

func TestGIAEUSupplementUploadAndForm(t *testing.T) {
	f, c := newFake(t)
	if err := c.GIA.UploadEUSupplement(ctx, 41, myitmo.Upload{Field: "file", Name: "eu.pdf", Content: strings.NewReader("EU")}); err != nil {
		t.Fatal(err)
	}
	if _, files := giaMultipart(t, f.expect(http.MethodPost, "/api/gia/attachment/eu/41/file")); files["scan"] != "eu.pdf:EU" {
		t.Errorf("files = %v", files)
	}

	f.replyWith(http.StatusOK, http.Header{"Content-Type": {"application/pdf"}, "Content-Disposition": {`attachment; filename="eu.pdf"`}}, "%PDF")
	file := must[*myitmo.File](t)(c.GIA.EUSupplementAttachment(ctx, 41))
	file.Body.Close()
	r := f.expect(http.MethodPost, "/api/gia/attachment/eu/41/attachment")
	if ct := r.Header.Get("Content-Type"); ct != "application/x-www-form-urlencoded" || len(r.Body) != 0 || file.Name != "eu.pdf" {
		t.Errorf("content type = %q, body = %q, file = %q", ct, r.Body, file.Name)
	}
}

func TestGIASupplementDecodes(t *testing.T) {
	f, c := newFake(t)
	f.result(`{"attachment_info":[{"student_info":{"student_id":11,"student_isu":100001,"student_surname":"Студентов"},
		"ep_info":{"ep_name":"Программа","ep_enrollment_year":2021},"faculty":{"faculty_short_name":"ФИТ"},"megafaculty":{"short_name":"МФ"},
		"group_id":"Z3400","secretary_info":null,"attachment_status":4,"diploma_status":{"status_id":4,"status_name":"Утверждено"},
		"avg_grade":4.8,"credit_units":240,"defence_date":"2025-06-20","defence_grade":"отлично","protocol_number":"1/1",
		"has_threes":false,"perfect_diploma":true}],"total_count":1}`)
	list := must[*myitmo.GIASupplementStudentList](t)(c.GIA.SupplementStudents(ctx, myitmo.GIASupplementParams{}))
	row := list.AttachmentInfo[0]
	if list.TotalCount != 1 || row.StudentInfo.StudentID != 11 || row.StudentInfo.StudentISU != 100001 || row.Megafaculty.ShortName != "МФ" ||
		row.Faculty.FacultyShortName != "ФИТ" || !row.PerfectDiploma || row.DefenceDate.Day() != 20 {
		t.Errorf("list = %+v", list)
	}

	f.result(`{"pers_info":{"surname":"Студентов","birth_date":"2003-01-02","snils":"000-000-000 00","status":{"status_id":4,"status_name":"Подтвержден"}},
		"disciplines":{"blocks":[{"title":"Блок 1","total":200,"rows":[{"disc_id":101,"disc_name":"Математика","credit_units":6,"avg_grade":5}]}],
			"avg_grade":4.8,"ungraded_count":0,"status":{"status_id":7,"status_name":"Не заполнено"}},
		"elective_disciplines":{"rows":[],"status":null},
		"additional_info":{"add_info":"нет","rows":[{"description":"язык","value":"англ","choice":true}],"status":{"status_id":4}}}`)
	st := must[*myitmo.GIASupplementStudent](t)(c.GIA.SupplementStudent(ctx, 11))
	if st.PersInfo.BirthDate != myitmo.NewDate(2003, 1, 2) || st.PersInfo.Status.StatusID != myitmo.GIASupplementBlockConfirmed ||
		st.Disciplines.Status.StatusID != myitmo.GIASupplementBlockNotFilled || !strings.Contains(string(st.Disciplines.Rest), "Математика") ||
		st.ElectiveDisciplines.Status != nil || st.AdditionalInfo.AddInfo != "нет" {
		t.Errorf("student = %+v", st)
	}

	f.result(`{"attachment_info":[{"eu_attachment_id":41,"diploma_id":31,"surname":"Студентов","isu":100001,"language_info":"eng",
		"payment_required":true,"payment_done":false,"created_at":"2025-06-25T10:00:00","status":{"status_id":28,"status_name":"В работе"}},
		{"id":42}],"total_count":2}`)
	eu := must[*myitmo.GIAEUSupplementList](t)(c.GIA.EUSupplements(ctx, myitmo.GIAEUSupplementsParams{}))
	if eu.TotalCount != 2 || eu.AttachmentInfo[0].Key() != 41 || eu.AttachmentInfo[1].Key() != 42 || !eu.AttachmentInfo[0].PaymentRequired {
		t.Errorf("eu list = %+v", eu)
	}

	f.result(`{"status":{"status_id":3,"status_name":"Отклонено","comment":"исправить"},"comment":"","reject_reason":"ошибка",
		"payment_required":false,"payment_done":false,"student_data":{"student_surname":"Студентов"},
		"pers_info":{"surname":"Studentov","birth_date":"2003-01-02","birth_place":"Saint Petersburg","isu":100001,"foreigner":false},
		"diploma_info":{"diploma_number":"100000 0000001","defence_date":"2025-06-20","edu_level_en":"Bachelor","total_credits":240,"file":null},
		"disciplines":{"blocks":[{"disc_info":[{"disc_id":101,"disc_name":"Mathematics","credit_units":6,"avg_grade":5,"avg_grade_letter":"A"}]}]},
		"elective_disciplines":null,"file":null,"attachment_file":{"file_name":"eu.pdf"}}`)
	d := must[*myitmo.GIAEUSupplement](t)(c.GIA.EUSupplement(ctx, 41))
	if d.Status.Comment != "исправить" || d.RejectReason != "ошибка" || d.StudentData.StudentSurname != "Студентов" ||
		d.PersInfo.BirthPlace != "Saint Petersburg" || d.DiplomaInfo.TotalCredits != 240 || d.AttachmentFile.FileName != "eu.pdf" ||
		!strings.Contains(string(d.Disciplines), "avg_grade_letter") {
		t.Errorf("eu = %+v", d)
	}
}
