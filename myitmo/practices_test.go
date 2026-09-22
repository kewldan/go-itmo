package myitmo_test

import (
	"context"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"strings"
	"testing"

	"github.com/kewldan/go-itmo/myitmo"
)

// formOf parses a multipart request body into fields and file contents.
func formOf(t *testing.T, r recorded) (fields map[string]string, files map[string]string) {
	t.Helper()
	mt, params, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || mt != "multipart/form-data" {
		t.Fatalf("content type = %q, want multipart/form-data", r.Header.Get("Content-Type"))
	}
	fields, files = map[string]string{}, map[string]string{}
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

func TestPracticesRoutes(t *testing.T) {
	ctx := context.Background()
	stage := myitmo.PracticeStageInput{PracticeID: 7, StageName: "Анализ", StageTask: "Изучить", Duration: 3}
	cases := []struct {
		name   string
		call   func(*myitmo.Client) error
		method string
		path   string
		query  string
		body   string
	}{
		{"Plans", func(c *myitmo.Client) error { _, err := c.Practices.Plans(ctx); return err }, "GET", "/api/practices/practices/op", "", ""},
		{"ListAll", func(c *myitmo.Client) error { _, err := c.Practices.List(ctx, nil); return err }, "GET", "/api/practices/practices/all", "", ""},
		{"ListPlan", func(c *myitmo.Client) error {
			_, err := c.Practices.List(ctx, &myitmo.PracticePlan{ID: 12, IsCurrentPlan: true})
			return err
		}, "GET", "/api/practices/practices/all", "current=true&op_id=12", ""},
		{"Get", func(c *myitmo.Client) error { _, err := c.Practices.Get(ctx, 7); return err }, "GET", "/api/practices/practices", "id=7", ""},
		{"Task", func(c *myitmo.Client) error { _, err := c.Practices.Task(ctx, 7); return err }, "GET", "/api/practices/ind_task", "id=7", ""},
		{"CreateTask", func(c *myitmo.Client) error { return c.Practices.CreateTask(ctx, 7, "Тема") }, "POST", "/api/practices/ind_task/create", "", `{"practiceId":7,"topic":"Тема"}`},
		{"UpdateTopic", func(c *myitmo.Client) error { return c.Practices.UpdateTopic(ctx, 7, "Новая") }, "PUT", "/api/practices/ind_task/create", "", `{"practiceId":7,"topic":"Новая"}`},
		{"AddStage", func(c *myitmo.Client) error { return c.Practices.AddStage(ctx, stage) }, "POST", "/api/practices/ind_task/stage/add", "",
			`{"practiceId":7,"stageName":"Анализ","stageTask":"Изучить","duration":3}`},
		{"UpdateStageDates", func(c *myitmo.Client) error {
			s := myitmo.PracticeStageInput{PracticeID: 7, StageName: "A", StageTask: "B", DateFrom: myitmo.NewDate(2026, 7, 1), DateTo: myitmo.NewDate(2026, 7, 10)}
			return c.Practices.UpdateStage(ctx, 3, s)
		}, "PUT", "/api/practices/ind_task/stage/update", "",
			`{"stageId":3,"practiceId":7,"stageName":"A","stageTask":"B","dateFrom":"2026-07-01","dateTo":"2026-07-10"}`},
		{"DeleteStage", func(c *myitmo.Client) error { return c.Practices.DeleteStage(ctx, 7, 3) }, "DELETE", "/api/practices/ind_task/stage/delete", "", `{"practiceId":7,"stageId":3}`},
		{"SubmitTask", func(c *myitmo.Client) error { return c.Practices.SubmitTask(ctx, 7) }, "POST", "/api/practices/ind_task/approve", "", `{"practiceId":7}`},
		{"CancelTask", func(c *myitmo.Client) error { return c.Practices.CancelTask(ctx, 7) }, "DELETE", "/api/practices/ind_task/cancel", "", `{"practiceId":7}`},
		{"Report", func(c *myitmo.Client) error { _, err := c.Practices.Report(ctx, 7); return err }, "GET", "/api/practices/report", "id=7", ""},
		{"ReportTypes", func(c *myitmo.Client) error { _, err := c.Practices.ReportTypes(ctx); return err }, "GET", "/api/practices/report/types", "", ""},
		{"PlaceGrades", func(c *myitmo.Client) error { _, err := c.Practices.PlaceGrades(ctx); return err }, "GET", "/api/practices/report/grades", "", ""},
		{"EmploymentStats", func(c *myitmo.Client) error { _, err := c.Practices.EmploymentStats(ctx); return err }, "GET", "/api/practices/report/emp_stats", "", ""},
		{"SubmitReport", func(c *myitmo.Client) error { return c.Practices.SubmitReport(ctx, 7) }, "POST", "/api/practices/report/approve", "", `{"practiceId":7}`},
		{"Feedback", func(c *myitmo.Client) error { _, err := c.Practices.Feedback(ctx, 7); return err }, "GET", "/api/practices/feedback/head", "id=7", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			f, c := newFake(t)
			if err := tc.call(c); err != nil {
				t.Fatal(err)
			}
			r := f.expect(tc.method, tc.path)
			if got := r.Query.Encode(); got != tc.query {
				t.Errorf("query = %q, want %q", got, tc.query)
			}
			if tc.body != "" {
				r.sameJSON(t, tc.body)
			} else if len(r.Body) != 0 {
				t.Errorf("unexpected body %q", r.Body)
			}
		})
	}
}

func TestPracticesDownloads(t *testing.T) {
	ctx := context.Background()
	cases := []struct {
		name   string
		call   func(*myitmo.Client) (*myitmo.File, error)
		method string
		path   string
		query  string
	}{
		{"Application", func(c *myitmo.Client) (*myitmo.File, error) { return c.Practices.Application(ctx, 7) }, "POST", "/api/practices/practices/7/application", ""},
		{"ReportFile", func(c *myitmo.Client) (*myitmo.File, error) { return c.Practices.ReportFile(ctx, 7) }, "GET", "/api/practices/report/file", "id=7"},
		{"FeedbackFile", func(c *myitmo.Client) (*myitmo.File, error) { return c.Practices.FeedbackFile(ctx, 7) }, "GET", "/api/practices/feedback/file", "id=7"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			f, c := newFake(t)
			f.replyWith(http.StatusOK, http.Header{
				"Content-Type":        {"application/octet-stream"},
				"Content-Disposition": {`attachment; filename="doc.docx"`},
			}, "DATA")
			file := must[*myitmo.File](t)(tc.call(c))
			defer file.Body.Close()
			b, _ := io.ReadAll(file.Body)
			if string(b) != "DATA" || file.Name != "doc.docx" {
				t.Fatalf("file = %q %q", file.Name, b)
			}
			r := f.expect(tc.method, tc.path)
			if got := r.Query.Encode(); got != tc.query {
				t.Errorf("query = %q, want %q", got, tc.query)
			}
		})
	}
}

func TestPracticesApplicationError(t *testing.T) {
	f, c := newFake(t)
	f.reply(http.StatusBadRequest, `{"error_code":1,"error_message":"Нет места практики"}`)
	_, err := c.Practices.Application(context.Background(), 7)
	if err == nil || !strings.Contains(err.Error(), "Нет места практики") {
		t.Fatalf("err = %v", err)
	}
}

func TestPracticesCreateReport(t *testing.T) {
	f, c := newFake(t)
	err := c.Practices.CreateReport(context.Background(), myitmo.PracticeReportUpload{
		PracticeID:       7,
		File:             &myitmo.Upload{Name: "report.zip", Content: strings.NewReader("ZIP")},
		ReportTypes:      []myitmo.IDValue{{ID: 1, Value: "Отчёт"}, {ID: 2, Value: "Презентация"}},
		EmploymentStatID: myitmo.Ptr[int64](3),
		PlaceGradeID:     myitmo.Ptr[int64](5),
	})
	if err != nil {
		t.Fatal(err)
	}
	fields, files := formOf(t, f.expect("POST", "/api/practices/report/create"))
	want := map[string]string{
		"practiceId":       "7",
		"fileName":         "report.zip",
		"reportTypes":      `[{"id":1,"value":"Отчёт"},{"id":2,"value":"Презентация"}]`,
		"employmentStatId": "3",
		"placeGradeId":     "5",
	}
	for k, v := range want {
		if fields[k] != v {
			t.Errorf("field %s = %q, want %q", k, fields[k], v)
		}
	}
	if len(fields) != len(want) {
		t.Errorf("fields = %v", fields)
	}
	if files["file"] != "report.zip:ZIP" {
		t.Errorf("files = %v", files)
	}
}

func TestPracticesUpdateReportKeepFile(t *testing.T) {
	f, c := newFake(t)
	err := c.Practices.UpdateReport(context.Background(), myitmo.PracticeReportUpload{
		PracticeID: 7, FileID: "55", FileName: "report.zip",
	})
	if err != nil {
		t.Fatal(err)
	}
	fields, files := formOf(t, f.expect("PUT", "/api/practices/report/update"))
	if fields["fileId"] != "55" || fields["fileName"] != "report.zip" || fields["reportTypes"] != "[]" || fields["practiceId"] != "7" {
		t.Errorf("fields = %v", fields)
	}
	if _, ok := fields["placeGradeId"]; ok || len(files) != 0 {
		t.Errorf("fields = %v, files = %v", fields, files)
	}
}

func TestPracticesDecode(t *testing.T) {
	ctx := context.Background()
	f, c := newFake(t)

	f.result(`[{"id":12,"opName":"Программная инженерия","isCurrentPlan":true}]`)
	plans := must[[]myitmo.PracticePlan](t)(c.Practices.Plans(ctx))
	if len(plans) != 1 || plans[0].ID != 12 || !plans[0].IsCurrentPlan {
		t.Fatalf("plans = %+v", plans)
	}

	f.result(`[{"year":2026,"practices":[{"id":7,"name":"Производственная практика","isCurrent":true,"mark":null,
		"baigColor":"Yellow","headName":"Иванов И. И.","dateFrom":"2026-07-01","dateTo":"2026-07-28T00:00:00"}]}]`)
	years := must[[]myitmo.PracticeYear](t)(c.Practices.List(ctx, nil))
	p := years[0].Practices[0]
	if string(years[0].Year) != "2026" || p.Mark != nil || p.BaigColor != "Yellow" || p.DateTo != myitmo.NewDate(2026, 7, 28) {
		t.Fatalf("years = %+v", years)
	}

	f.result(`{"id":7,"name":"Практика","status":{"statusOrder":4},"dateFrom":"2026-07-01","dateTo":"2026-07-28",
		"placeId":null,"placeName":"ООО Пример","departmentName":"Отдел","position":"Стажёр","format":null,
		"curatorId":100001,"curatorName":"Куратор","curatorEmail":"c@example.com","curatorPhoto":null,
		"headItmoName":"Руководитель","HeadItmoIsu":100002,"headItmoEmail":"h@example.com","headItmoPhoto":"https://example.com/p/",
		"headExternalName":"Внешний","headExternalEmail":"e@example.com"}`)
	pr := must[*myitmo.Practice](t)(c.Practices.Get(ctx, 7))
	if pr.HeadItmoIsu != 100002 || pr.Status.StatusOrder < myitmo.PracticeOrderReport || pr.PlaceID != nil {
		t.Fatalf("practice = %+v", pr)
	}

	f.result(`{"topic":"Тема","status":"Черновик","statusId":108,"comment":null,"stages":[
		{"stageId":3,"stageNumber":1,"stageName":"A","stageTask":"B","dateFrom":null,"dateTo":null,"duration":5},
		{"stageId":4,"stageNumber":2,"stageName":"C","stageTask":"D","dateFrom":"2026-07-05","dateTo":"2026-07-09","duration":null}]}`)
	task := must[*myitmo.PracticeTask](t)(c.Practices.Task(ctx, 7))
	if task.StatusID != myitmo.PracticeTaskDraft || *task.Stages[0].Duration != 5 || task.Stages[1].Duration != nil ||
		task.Stages[1].DateFrom != myitmo.NewDate(2026, 7, 5) {
		t.Fatalf("task = %+v", task)
	}

	f.result(`null`)
	if task := must[*myitmo.PracticeTask](t)(c.Practices.Task(ctx, 7)); task != nil {
		t.Fatalf("task = %+v, want nil", task)
	}

	f.result(`{"report":{"fileId":55,"fileName":"report.zip","reportTypes":[{"id":1,"value":"Отчёт"}],
		"feedBack":{"placeGradeId":5,"employmentStatId":3,"placeGrade":"5","employmentStat":"Трудоустроен"}},"showFeedBack":true}`)
	rep := must[*myitmo.PracticeReportState](t)(c.Practices.Report(ctx, 7))
	if !rep.ShowFeedBack || string(rep.Report.FileID) != "55" || rep.Report.FeedBack.PlaceGradeID != 5 || rep.Report.ReportTypes[0].Value != "Отчёт" {
		t.Fatalf("report = %+v", rep)
	}

	f.result(`[{"id":1,"value":"Отлично"}]`)
	if g := must[[]myitmo.IDValue](t)(c.Practices.PlaceGrades(ctx)); g[0].ID != 1 {
		t.Fatalf("grades = %+v", g)
	}

	f.result(`{"internalHead":{"feedBack":"Хорошо","feedBackGrade":"5","feedBackFileName":null,
		"stages":[{"id":1,"stageNumber":1,"stageName":"A","stageResult":"Выполнено","stageComment":""}]},
		"externalHead":{"feedBack":null,"feedBackGrade":4,"feedBackFileName":"review.pdf","stages":[]}}`)
	fb := must[*myitmo.PracticeFeedback](t)(c.Practices.Feedback(ctx, 7))
	if fb.InternalHead.Stages[0].StageResult != "Выполнено" || fb.ExternalHead.FeedBackFileName != "review.pdf" || string(fb.ExternalHead.FeedBackGrade) != "4" {
		t.Fatalf("feedback = %+v", fb)
	}
}
