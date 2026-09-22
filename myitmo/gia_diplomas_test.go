package myitmo_test

import (
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/kewldan/go-itmo/myitmo"
)

func TestGIADiplomaRequests(t *testing.T) {
	yes := true
	no := false
	sup := myitmo.GIASupervisorReview{Approved: true, InitiativeGoals: 1, InitiativeMethods: 2, QualityLogic: 3,
		JustificationRelevance: 4, Consistency: 5, Involvement: 6, Preparedness: 7, Approbation: 8, Publication: 9,
		PersonalInvolvement: 10, Integration: 11, Advantages: "плюсы", Disadvantages: "минусы", Comment: "ок",
		Assessment: 21, ThesisComplete: true, AwardingQualifications: true}
	supJSON := `{"approved":true,"initiative_goals":1,"initiative_methods":2,"quality_logic":3,"justification_relevance":4,
		"consistency":5,"involvement":6,"preparedness":7,"approbation":8,"publication":9,"personal_involvement":10,"integration":11,
		"advantages":"плюсы","disadvantages":"минусы","comment":"ок","assessment":21,"thesis_complete":true,"awarding_qualifications":true}`
	runGIACases(t, []giaCase{
		{"Monitoring", func(c *myitmo.Client) error {
			_, err := c.GIA.Monitoring(ctx, myitmo.GIAMonitoringParams{Limit: 15, Year: 2025, ShowAs: myitmo.GIARoleOSOP, Faculty: 2,
				GroupID: "Z3400", EduDirection: 8, EduProgram: 7, ApprovalNeeded: &yes, Approved: &no, DefenseDate: "2025-06-20",
				SortBy: []string{"student_surname", "group_id"}, SortOrder: []string{"asc", "desc"}, Query: "тест"})
			return err
		}, "GET", "/api/gia/diplomas/monitoring", url.Values{"limit": {"15"}, "offset": {"0"}, "year": {"2025"},
			"hide_deducted": {"false"}, "show_as": {"osop"}, "faculty": {"2"}, "group_id": {"Z3400"}, "edu_direction": {"8"},
			"edu_program": {"7"}, "approval_needed": {"true"}, "approved": {"false"}, "defense_date": {"2025-06-20"},
			"sort_by": {"student_surname,group_id"}, "sort_order": {"asc,desc"}, "type": {"monitoring"}, "query": {"тест"}}, ""},
		{"MonitoringFilters", func(c *myitmo.Client) error {
			_, err := c.GIA.MonitoringFilters(ctx, myitmo.GIARoleGeneralStateCoordinator)
			return err
		}, "GET", "/api/gia/diplomas/monitoring/filters", url.Values{"show_as": {"general_state_coordinator"}}, ""},
		{"StageMonitoring", func(c *myitmo.Client) error {
			_, err := c.GIA.StageMonitoring(ctx, myitmo.GIAStageMonitoringParams{Stage: myitmo.GIAStageTask, Limit: 15, Offset: 15,
				StatusID: 2, HideDeducted: true, EduProgram: 7, ApprovalNeeded: &no})
			return err
		}, "GET", "/api/gia/diplomas/monitoring/stage", url.Values{"stage": {"task"}, "limit": {"15"}, "offset": {"15"},
			"status_id": {"2"}, "hide_deducted": {"true"}, "edu_program": {"7"}, "approval_needed": {"false"}}, ""},
		{"StageFilters", func(c *myitmo.Client) error {
			_, err := c.GIA.StageFilters(ctx, myitmo.GIAStageAnnotation, myitmo.GIARoleOSOP)
			return err
		}, "GET", "/api/gia/diplomas/monitoring/stage/filters", url.Values{"stage": {"annotation"}, "show_as": {"osop"}}, ""},
		{"PendingApprovalCounts", func(c *myitmo.Client) error { _, err := c.GIA.PendingApprovalCounts(ctx); return err },
			"GET", "/api/gia/diplomas/stage/onApproval/count", nil, ""},
		{"SupervisorReviewCount", func(c *myitmo.Client) error { _, err := c.GIA.SupervisorReviewCount(ctx); return err },
			"GET", "/api/gia/diplomas/review/supervisor/count", nil, ""},
		{"ApproveStages", func(c *myitmo.Client) error { return c.GIA.ApproveStages(ctx, myitmo.GIAStageApplication, 31, 32) },
			"PUT", "/api/gia/diplomas/statuses", nil, `{"stage":"application","diploma_ids":[31,32]}`},
		{"Diploma", func(c *myitmo.Client) error { _, err := c.GIA.Diploma(ctx, 31); return err },
			"GET", "/api/gia/diplomas/31/student", nil, ""},
		{"DiplomaHistory", func(c *myitmo.Client) error { _, err := c.GIA.DiplomaHistory(ctx, 31, myitmo.GIAStageFile); return err },
			"GET", "/api/gia/diplomas/31/history", url.Values{"stage": {"file"}}, ""},
		{"AntiplagiatStatus", func(c *myitmo.Client) error { _, err := c.GIA.AntiplagiatStatus(ctx, 31); return err },
			"GET", "/api/gia/diplomas/31/antiplagiat/status", nil, ""},
		{"SendToAntiplagiat", func(c *myitmo.Client) error { return c.GIA.SendToAntiplagiat(ctx, 31) },
			"POST", "/api/gia/diplomas/31/antiplagiat/send", nil, ""},
		{"ApproveStage", func(c *myitmo.Client) error {
			return c.GIA.ApproveStage(ctx, 31, myitmo.GIAStageApplication, myitmo.GIARoleEPManager)
		}, "PUT", "/api/gia/diplomas/31/application/status", url.Values{"status": {"ep_manager"}}, `{}`},
		{"RejectStage", func(c *myitmo.Client) error {
			return c.GIA.RejectStage(ctx, 31, myitmo.GIAStageTask, "доработать")
		},
			"PUT", "/api/gia/diplomas/31/task/status", url.Values{"status": {"reject"}}, `{"message":"доработать"}`},
		{"SubmitSupervisorReview", func(c *myitmo.Client) error { return c.GIA.SubmitSupervisorReview(ctx, 31, sup) },
			"PUT", "/api/gia/diplomas/31/review/supervisor", nil, supJSON},
		{"PrintList", func(c *myitmo.Client) error {
			_, err := c.GIA.PrintList(ctx, myitmo.GIAPrintParams{DocType: myitmo.GIADocTypeDiploma, Limit: 20, Year: 2025,
				MegafacultyID: 1, ImplementerID: 2, EducationLevelID: 3, GroupID: "Z3400", WithHonors: &yes, StatusID: 4, Query: "тест"})
			return err
		}, "GET", "/api/gia/diplomas/print/list", url.Values{"doctype": {"diploma"}, "limit": {"20"}, "offset": {"0"},
			"year": {"2025"}, "megafaculty_id": {"1"}, "implementer_id": {"2"}, "education_level_id": {"3"}, "group_id": {"Z3400"},
			"with_honors": {"true"}, "status_id": {"4"}, "query": {"тест"}}, ""},
		{"SetBlankNumber", func(c *myitmo.Client) error {
			return c.GIA.SetBlankNumber(ctx, 31, myitmo.GIADocTypeAttachment, 123456)
		},
			"PATCH", "/api/gia/diplomas/print/31/number", url.Values{"doctype": {"attachment"}}, `{"number":123456}`},
	})
	runGIADownloads(t, []giaDownload{
		{"AntiplagiatReport", func(c *myitmo.Client) (*myitmo.File, error) { return c.GIA.AntiplagiatReport(ctx, 31) },
			"GET", "/api/gia/diplomas/31/antiplagiat/verification-report", nil, ""},
		{"ApplicationFile", func(c *myitmo.Client) (*myitmo.File, error) { return c.GIA.ApplicationFile(ctx, 31) },
			"GET", "/api/gia/diplomas/31/application/file", nil, ""},
		{"DiplomaFile", func(c *myitmo.Client) (*myitmo.File, error) { return c.GIA.DiplomaFile(ctx, 31, "a b/c") },
			"GET", "/api/gia/diplomas/31/download/a%20b%2Fc", nil, ""},
		{"ReviewFile", func(c *myitmo.Client) (*myitmo.File, error) {
			return c.GIA.ReviewFile(ctx, 31, myitmo.GIAReviewReviewer)
		},
			"GET", "/api/gia/diplomas/31/review/reviewer/file", nil, ""},
		{"PrintPreview", func(c *myitmo.Client) (*myitmo.File, error) {
			return c.GIA.PrintPreview(ctx, 31, myitmo.GIAPrintPreviewParams{DocType: myitmo.GIADocTypeAttachment, FontSize: 10.5, AttachmentSecondFontSize: 9})
		}, "POST", "/api/gia/diplomas/print/31", url.Values{"font_size": {"10.5"}, "attachment_second_font_size": {"9"}, "doctype": {"attachment"}}, `{}`},
		{"PrintPreviewDiploma", func(c *myitmo.Client) (*myitmo.File, error) {
			return c.GIA.PrintPreview(ctx, 31, myitmo.GIAPrintPreviewParams{DocType: myitmo.GIADocTypeDiploma, FontSize: 12, AttachmentSecondFontSize: 9})
		}, "POST", "/api/gia/diplomas/print/31", url.Values{"font_size": {"12"}, "doctype": {"diploma"}}, `{}`},
		{"EUSupplementTemplate", func(c *myitmo.Client) (*myitmo.File, error) { return c.GIA.EUSupplementTemplate(ctx, 41) },
			"POST", "/api/gia/diplomas/print/41/eu_attachment", nil, ""},
	})
}

// giaMultipart parses a multipart request body into fields and file parts.
func giaMultipart(t *testing.T, r recorded) (map[string]string, map[string]string) {
	t.Helper()
	mt, params, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || mt != "multipart/form-data" {
		t.Fatalf("content type = %q", r.Header.Get("Content-Type"))
	}
	fields, files := map[string]string{}, map[string]string{}
	mr := multipart.NewReader(strings.NewReader(string(r.Body)), params["boundary"])
	for {
		p, err := mr.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		b, _ := io.ReadAll(p)
		if p.FileName() != "" {
			files[p.FormName()] = p.FileName() + ":" + string(b)
		} else {
			fields[p.FormName()] = string(b)
		}
	}
	return fields, files
}

func TestGIADiplomaUploads(t *testing.T) {
	f, c := newFake(t)
	err := c.GIA.SetAntiplagiatInfo(ctx, 31, myitmo.GIAAntiplagiatInfo{Originality: 85.5, Citations: 10, Similarity: 4.5, SelfCitations: 0,
		AntiplagiatReportWebURL: "https://example.com/report"}, myitmo.Upload{Name: "report.pdf", ContentType: "application/pdf", Content: strings.NewReader("PDF")})
	if err != nil {
		t.Fatal(err)
	}
	fields, files := giaMultipart(t, f.expect(http.MethodPut, "/api/gia/diplomas/31/antiplagiat/info"))
	if fields["originality"] != "85.5" || fields["citations"] != "10" || fields["similarity"] != "4.5" || fields["self_citations"] != "0" ||
		fields["antiplagiat_report_web_url"] != "https://example.com/report" || files["report"] != "report.pdf:PDF" {
		t.Errorf("fields = %v, files = %v", fields, files)
	}

	if err := c.GIA.UploadPrintScan(ctx, 31, myitmo.GIADocTypeDiploma, myitmo.Upload{Name: "scan.pdf", Content: strings.NewReader("SCAN")}); err != nil {
		t.Fatal(err)
	}
	r := f.expect(http.MethodPost, "/api/gia/diplomas/print/31/scan")
	if r.Query.Get("doctype") != "diploma" {
		t.Errorf("query = %v", r.Query)
	}
	if _, files := giaMultipart(t, r); files["scan"] != "scan.pdf:SCAN" {
		t.Errorf("files = %v", files)
	}
}

func TestGIADiplomaDecodes(t *testing.T) {
	f, c := newFake(t)
	f.result(`{"diploma_id":31,"requester_reviewer_id":null,
		"main_info":{"student_isu":100001,"student_surname":"Студентов","theme_ru":"Тема","ep_diploma_year":2025,"supervisor_isu":100002},
		"application":{"status_id":8,"status_name":"На согласовании","theme_en":"Topic","co_supervisors":[{"surname":"Со","job_title":"доцент"}],
			"consultants":[],"comment":"","commenter_fio":"","commenter_role":""},
		"task":{"status_id":4,"issued_at":"2025-02-01","main_questions":"вопросы"},
		"annotation":{"status_id":4,"grants":[{"name":"грант","year":2024}],"publications":[{"name":"статья","status":"опубликована","type":"ВАК"}],
			"speeches":[{"name":"конференция"}]},
		"file":{"status_id":10,"student_file_key":"sf","result_file_key":"rf","originality":87.2,"comment":"ок","commenter_role":"secretary"},
		"supervisor_review":{"approved":true,"initiative_goals":3,"initiative_goals_name":"хорошо","assessment":21,"assessment_name":"отлично",
			"supervisor_signature":null},
		"reviewer_reviews":[{"approved":false,"reviewer_id":5,"value":2}],
		"reviewer_review_questions":[{"id":1,"question":"Почему?","reviewer_id":5}]}`)
	d := must[*myitmo.GIADiploma](t)(c.GIA.Diploma(ctx, 31))
	if d.RequesterReviewerID != nil || d.MainInfo.EPDiplomaYear != 2025 || d.Application.CoSupervisors[0].JobTitle != "доцент" ||
		d.Task.IssuedAt.Month() != 2 || string(d.Annotation.Grants[0].Year) != "2024" || d.File.Originality != 87.2 ||
		d.File.Comment != "ок" || d.File.CommenterRole != "secretary" || d.SupervisorReview.AssessmentName != "отлично" ||
		!strings.Contains(string(d.SupervisorReview.Criteria), `"initiative_goals_name":"хорошо"`) ||
		d.ReviewerReviews[0].ReviewerID != 5 || d.ReviewerReviewQuestions[0].Question != "Почему?" {
		t.Errorf("diploma = %+v", d)
	}

	f.result(`{"result":[{"diploma_id":31,"student_isu":100001,"student_status":"академ","defense_date":null,
		"reviewers":[{"reviewer_id":5,"reviewer_surname":"Рецензентов"}],
		"stage_statuses":[{"stage":"application","status_id":4,"status_name":"Утверждено","color":"#E9F7E8"}],
		"review_statuses":[{"reviewer":"supervisor","status_id":1}]}],"count":1}`)
	m := must[*myitmo.GIAMonitoringList](t)(c.GIA.Monitoring(ctx, myitmo.GIAMonitoringParams{}))
	if m.Count != 1 || m.Result[0].StageStatuses[0].Stage != myitmo.GIAStageApplication ||
		!strings.Contains(string(m.Result[0].StageStatuses[0].Extra), "#E9F7E8") || m.Result[0].ReviewStatuses[0].Reviewer != "supervisor" ||
		!m.Result[0].DefenseDate.IsZero() {
		t.Errorf("monitoring = %+v", m)
	}

	f.result(`{"result":[{"diploma_id":31,"diploma_theme":"Тема","status_id":10,"originality":90,"result_file_key":"rf",
		"supervisor_review_status":{"status_name":"Подписан"},"reviewer_review_status":null}],"count":1}`)
	sl := must[*myitmo.GIAStageList](t)(c.GIA.StageMonitoring(ctx, myitmo.GIAStageMonitoringParams{Stage: myitmo.GIAStageFile}))
	if sl.Result[0].SupervisorReviewStatus.StatusName != "Подписан" || sl.Result[0].ReviewerReviewStatus != nil {
		t.Errorf("stage = %+v", sl)
	}

	f.result(`[{"stage":"application","count":3},{"stage":"file","count":0}]`)
	counts := must[[]myitmo.GIAStageCount](t)(c.GIA.PendingApprovalCounts(ctx))
	if counts[0].Count != 3 {
		t.Errorf("counts = %+v", counts)
	}

	f.result(`{"count":4}`)
	if n := must[int](t)(c.GIA.SupervisorReviewCount(ctx)); n != 4 {
		t.Errorf("count = %d", n)
	}

	f.result(`[{"status_id":3,"comment":"исправить","commenter_fio":"Проверяющий","commenter_isu":100003,"commenter_role":"osop",
		"created_at":"2025-03-01T10:00:00+03:00"}]`)
	h := must[[]myitmo.GIADiplomaHistoryItem](t)(c.GIA.DiplomaHistory(ctx, 31, myitmo.GIAStageTask))
	if h[0].CommenterISU != 100003 || h[0].CreatedAt.IsZero() {
		t.Errorf("history = %+v", h)
	}

	f.result(`{"status":"done","is_ready":true,"is_failed":false,"fail_details":"","originality":90.1,"similarity":5,
		"citations":4.9,"self_citations":0,"report_file_name":"report.pdf","report_web_url":"https://example.com/r"}`)
	ap := must[*myitmo.GIAAntiplagiatStatus](t)(c.GIA.AntiplagiatStatus(ctx, 31))
	if !ap.IsReady || ap.Originality != 90.1 {
		t.Errorf("antiplagiat = %+v", ap)
	}

	f.result(`{"array":[{"id":1,"diploma_id":31,"student_fio":"Студентов С.С.","student_isu":100001,"defense_date":"2025-06-20",
		"doc_series":"100000","doc_number":"0000001","with_honors":true,"status_id":2,"status_name":"Напечатан"}],"count":1,
		"filters":{"megafaculties":[{"id":1,"name":"Мегафакультет","short_name":"МФ"}],"implementers":[],"education_levels":[{"id":3,"name":"Бакалавриат"}],
		"groups":["Z3400"],"statuses":[{"id":2,"name":"Напечатан"}]}}`)
	pl := must[*myitmo.GIAPrintList](t)(c.GIA.PrintList(ctx, myitmo.GIAPrintParams{}))
	if pl.Count != 1 || !pl.Array[0].WithHonors || pl.Filters.Megafaculties[0].ShortName != "МФ" || pl.Filters.EducationLevels[0].ID != 3 {
		t.Errorf("print = %+v", pl)
	}
}
