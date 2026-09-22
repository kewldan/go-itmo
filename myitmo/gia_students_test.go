package myitmo_test

import (
	"context"
	"errors"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"strings"
	"testing"

	"github.com/kewldan/go-itmo/myitmo"
)

func TestGIAStudentsSimpleCalls(t *testing.T) {
	ctx := context.Background()
	const d = "/api/gia-students/diplomas/501"
	const a = "/api/gia-students/attachment/501"
	cases := []struct {
		name   string
		call   func(s *myitmo.GIAStudentsService) error
		method string
		path   string
		query  string
		body   string // expected JSON body; "" means no body
	}{
		{"Diplomas", func(s *myitmo.GIAStudentsService) error { _, err := s.Diplomas(ctx); return err }, "GET", "/api/gia-students/diplomas/my", "", ""},
		{"DiplomaInfo", func(s *myitmo.GIAStudentsService) error { _, err := s.DiplomaInfo(ctx, 501); return err }, "GET", d + "/info", "", ""},
		{"AcceptAgreement", func(s *myitmo.GIAStudentsService) error { return s.AcceptAgreement(ctx, 501) }, "POST", d + "/agreement", "", ""},
		{"Application", func(s *myitmo.GIAStudentsService) error { _, err := s.Application(ctx, 501); return err }, "GET", d + "/application", "", ""},
		{"Revoke", func(s *myitmo.GIAStudentsService) error { return s.Revoke(ctx, 501, myitmo.GIAStudentStageTask) }, "DELETE", d + "/revoke", "stage=task", ""},
		{"Task", func(s *myitmo.GIAStudentsService) error { _, err := s.Task(ctx, 501); return err }, "GET", d + "/task", "", ""},
		{"StartTask", func(s *myitmo.GIAStudentsService) error { return s.StartTask(ctx, 501) }, "PUT", d + "/task", "", `{"status_id":7}`},
		{"Annotation", func(s *myitmo.GIAStudentsService) error { _, err := s.Annotation(ctx, 501); return err }, "GET", d + "/annotation", "", ""},
		{"StartAnnotation", func(s *myitmo.GIAStudentsService) error { return s.StartAnnotation(ctx, 501) }, "PUT", d + "/annotation", "", `{"status_id":7}`},
		{"Grants", func(s *myitmo.GIAStudentsService) error { _, err := s.Grants(ctx); return err }, "GET", "/api/gia-students/me/grants", "", ""},
		{"Publications", func(s *myitmo.GIAStudentsService) error { _, err := s.Publications(ctx); return err }, "GET", "/api/gia-students/me/publications", "", ""},
		{"Speeches", func(s *myitmo.GIAStudentsService) error { _, err := s.Speeches(ctx); return err }, "GET", "/api/gia-students/me/speeches", "", ""},
		{"File", func(s *myitmo.GIAStudentsService) error { _, err := s.File(ctx, 501); return err }, "GET", d + "/file", "", ""},
		{"StartFileStage", func(s *myitmo.GIAStudentsService) error { return s.StartFileStage(ctx, 501) }, "PUT", d + "/file", "status_id=7", ""},
		{"ViewSupervisorReview", func(s *myitmo.GIAStudentsService) error { return s.ViewSupervisorReview(ctx, 501) }, "POST", d + "/review/supervisor/view", "", ""},
		{"ViewReviewerReview", func(s *myitmo.GIAStudentsService) error { return s.ViewReviewerReview(ctx, 501, 77) }, "POST", d + "/review/reviewer/77/view", "", ""},
		{"PossiblePresentations", func(s *myitmo.GIAStudentsService) error { _, err := s.PossiblePresentations(ctx, 501); return err }, "GET", d + "/presentations/possible", "", ""},
		{"SignUpPresentation", func(s *myitmo.GIAStudentsService) error { return s.SignUpPresentation(ctx, 501, 12) }, "POST", d + "/presentations/12/signup", "", ""},
		{"CancelPresentationSignUp", func(s *myitmo.GIAStudentsService) error { return s.CancelPresentationSignUp(ctx, 501, 12) }, "DELETE", d + "/presentations/12/signup", "", ""},
		{"Presentation", func(s *myitmo.GIAStudentsService) error { _, err := s.Presentation(ctx, 501); return err }, "GET", d + "/presentation", "", ""},
		{"DeletePresentationExtraFile", func(s *myitmo.GIAStudentsService) error { return s.DeletePresentationExtraFile(ctx, 501, "a/b c") }, "DELETE", d + "/presentation/extra-files/a%2Fb%20c", "", ""},
		{"Supplement", func(s *myitmo.GIAStudentsService) error { _, err := s.Supplement(ctx, 501); return err }, "GET", a, "", ""},
		{"ApproveSupplementPersInfo", func(s *myitmo.GIAStudentsService) error { return s.ApproveSupplementPersInfo(ctx, 501) }, "POST", "/api/gia-students/attachment/approve/pers_info/501", "", ""},
		{"ApproveSupplementDisciplines", func(s *myitmo.GIAStudentsService) error { return s.ApproveSupplementDisciplines(ctx, 501) }, "POST", "/api/gia-students/attachment/approve/disciplines/501", "", ""},
		{"ApproveSupplementElectives", func(s *myitmo.GIAStudentsService) error {
			return s.ApproveSupplementElectives(ctx, 501, []myitmo.GIAStudentElectiveChoice{{ObjectID: 31, Choice: true}, {ObjectID: 32}})
		}, "POST", "/api/gia-students/attachment/approve/faculties/501", "", `[{"object_id":31,"choice":true},{"object_id":32,"choice":false}]`},
		{"ApproveSupplementAdditionalInfo", func(s *myitmo.GIAStudentsService) error {
			return s.ApproveSupplementAdditionalInfo(ctx, 501, []myitmo.GIAStudentAddInfoChoice{{ObjectID: "Olympiad winner", Choice: true}})
		}, "POST", "/api/gia-students/attachment/approve/additional_info/501", "", `[{"object_id":"Olympiad winner","choice":true}]`},
		{"EUSupplement", func(s *myitmo.GIAStudentsService) error { _, err := s.EUSupplement(ctx, 501); return err }, "GET", a + "/eu", "", ""},
		{"SubmitEUSupplement", func(s *myitmo.GIAStudentsService) error {
			_, err := s.SubmitEUSupplement(ctx, 501, myitmo.GIAStudentEUSupplementRequest{
				Name: "Ivan", Surname: "Testov", BirthPlace: "Saint Petersburg", Theme: "Graph search", PrintAttachment: true, Email: "student@example.com",
			})
			return err
		}, "POST", a + "/eu", "", `{"name":"Ivan","surname":"Testov","birth_place":"Saint Petersburg","theme":"Graph search","print_attachment":true,"email":"student@example.com"}`},
		{"DeleteEUSupplement", func(s *myitmo.GIAStudentsService) error { return s.DeleteEUSupplement(ctx, 501) }, "DELETE", a + "/eu", "", ""},
		{"ApproveEUSupplement", func(s *myitmo.GIAStudentsService) error { return s.ApproveEUSupplement(ctx, 501) }, "POST", a + "/eu/approve", "", ""},
		{"RollbackEUSupplement", func(s *myitmo.GIAStudentsService) error { return s.RollbackEUSupplement(ctx, 501, "Wrong birth place") }, "POST", a + "/eu/rollback", "", `{"student_comment":"Wrong birth place"}`},
		{"People", func(s *myitmo.GIAStudentsService) error { _, err := s.People(ctx, " "); return err }, "GET", "/api/gia-students/people", "query=+", ""},
		{"DiplomaTypes", func(s *myitmo.GIAStudentsService) error { _, err := s.DiplomaTypes(ctx); return err }, "GET", "/api/gia-students/references/diploma-types", "", ""},
		{"Statuses", func(s *myitmo.GIAStudentsService) error {
			_, err := s.Statuses(ctx, myitmo.GIAStudentStageFile)
			return err
		}, "GET", "/api/gia-students/references/statuses", "entity=file", ""},
		{"AcademicTitles", func(s *myitmo.GIAStudentsService) error { _, err := s.AcademicTitles(ctx); return err }, "GET", "/api/gia-students/references/academic-titles", "", ""},
		{"DegreeTypes", func(s *myitmo.GIAStudentsService) error { _, err := s.DegreeTypes(ctx); return err }, "GET", "/api/gia-students/references/degree-types", "", ""},
		{"OrganizationStatuses", func(s *myitmo.GIAStudentsService) error { _, err := s.OrganizationStatuses(ctx); return err }, "GET", "/api/gia-students/references/organization-statuses", "", ""},
		{"AgreementText", func(s *myitmo.GIAStudentsService) error { _, err := s.AgreementText(ctx); return err }, "GET", "/api/gia-students/references/agreement", "", ""},
		{"UserStatus", func(s *myitmo.GIAStudentsService) error { _, err := s.UserStatus(ctx); return err }, "GET", "/api/gia-students/users/status", "", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			f, c := newFake(t)
			if err := tc.call(c.GIAStudents); err != nil {
				t.Fatal(err)
			}
			r := f.expect(tc.method, tc.path)
			if got := r.Query.Encode(); got != tc.query {
				t.Errorf("query = %q, want %q", got, tc.query)
			}
			if tc.body == "" {
				if len(r.Body) != 0 {
					t.Errorf("unexpected body %s", r.Body)
				}
				return
			}
			r.sameJSON(t, tc.body)
		})
	}
}

func TestGIAStudentsUpdateBodies(t *testing.T) {
	ctx := context.Background()
	f, c := newFake(t)
	s := c.GIAStudents
	noErr := func(err error) {
		t.Helper()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}

	noErr(s.UpdateApplication(ctx, 501, myitmo.GIAStudentApplicationUpdate{
		StatusID:           myitmo.GIAStudentStatusSentToApproval,
		DiplomaTypeID:      2,
		DiplomaTheme:       "Graph search methods",
		DiplomaFundamental: true,
		SupervisorISU:      myitmo.Ptr[int64](100001),
		CoSupervisors: []myitmo.GIAStudentAdviser{
			{ISU: myitmo.Ptr[int64](100002)},
			{Surname: "Petrov", Name: "Petr", WorkPlace: "Example LLC", JobTitle: "Engineer", DegreeTypeID: myitmo.Ptr[int64](3)},
		},
	}))
	f.expect("PUT", "/api/gia-students/diplomas/501/application").sameJSON(t, `{
		"status_id":8,"diploma_type_id":2,"diploma_type_name":null,"diploma_theme":"Graph search methods",
		"diploma_justification":null,"diploma_external_partner":false,"diploma_external_partner_name":null,
		"diploma_fundamental":true,"supervisor_isu":100001,"consultants":null,
		"co_supervisors":[{"isu":100002},{"surname":"Petrov","name":"Petr","work_place":"Example LLC","job_title":"Engineer","degree_type_id":3}]}`)

	doc := &myitmo.EditorJS{Time: 1700000000000, Blocks: []myitmo.EditorJSBlock{{ID: "b1", Type: "paragraph", Data: myitmo.RawJSON(`{"text":"Study X"}`)}}, Version: "2.28.0"}
	noErr(s.UpdateTask(ctx, 501, myitmo.GIAStudentTaskUpdate{StatusID: 7, MainQuestions: doc, Format: "Report"}))
	f.expect("PUT", "/api/gia-students/diplomas/501/task").sameJSON(t, `{"status_id":7,
		"main_questions":{"time":1700000000000,"blocks":[{"id":"b1","type":"paragraph","data":{"text":"Study X"}}],"version":"2.28.0"},
		"format":"Report","extra_info":null}`)

	noErr(s.UpdateAnnotation(ctx, 501, myitmo.GIAStudentAnnotationUpdate{
		StatusID: 8, Goal: doc, Tasks: doc, Results: doc, GrantIDs: []int64{1}, PublicationIDs: []int64{2, 3},
	}))
	f.expect("PUT", "/api/gia-students/diplomas/501/annotation").sameJSON(t, `{"status_id":8,
		"goal":{"time":1700000000000,"blocks":[{"id":"b1","type":"paragraph","data":{"text":"Study X"}}],"version":"2.28.0"},
		"tasks":{"time":1700000000000,"blocks":[{"id":"b1","type":"paragraph","data":{"text":"Study X"}}],"version":"2.28.0"},
		"results":{"time":1700000000000,"blocks":[{"id":"b1","type":"paragraph","data":{"text":"Study X"}}],"version":"2.28.0"},
		"grant_ids":[1],"publication_ids":[2,3],"speech_ids":[],"extra_info":null}`)

	f.result(`"https://pay.example.com/form/1"`)
	u := must[string](t)(s.InitEUSupplementPayment(ctx, 501, "https://my.example/ok", "https://my.example/fail"))
	f.expect("POST", "/api/gia-students/attachment/501/eu/init_payment").sameJSON(t, `{"SuccessUrl":"https://my.example/ok","FailureUrl":"https://my.example/fail"}`)
	if u != "https://pay.example.com/form/1" {
		t.Errorf("payment url = %q", u)
	}
}

// multipartParts reads the fields and file names of a multipart request.
func multipartParts(t *testing.T, r recorded) (fields map[string]string, files map[string]string) {
	t.Helper()
	_, params, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil {
		t.Fatalf("content type: %v", err)
	}
	fields, files = map[string]string{}, map[string]string{}
	mr := multipart.NewReader(strings.NewReader(string(r.Body)), params["boundary"])
	for {
		p, err := mr.NextPart()
		if errors.Is(err, io.EOF) {
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

func TestGIAStudentsUploads(t *testing.T) {
	ctx := context.Background()
	f, c := newFake(t)
	s := c.GIAStudents

	err := s.UploadFile(ctx, 501, myitmo.GIAStudentStatusFileSentToApproval, true,
		myitmo.Upload{Name: "work.pdf", ContentType: "application/pdf", Content: strings.NewReader("%PDF")})
	if err != nil {
		t.Fatal(err)
	}
	r := f.expect("PUT", "/api/gia-students/diplomas/501/file")
	if got := r.Query.Encode(); got != "publish=true&status_id=10" {
		t.Errorf("query = %q", got)
	}
	fields, files := multipartParts(t, r)
	if fields["status_id"] != "10" || fields["publish"] != "true" || files["file"] != "work.pdf:%PDF" {
		t.Errorf("fields = %v, files = %v", fields, files)
	}

	if err := s.UploadPresentation(ctx, 501, myitmo.Upload{Name: "slides.pdf", Content: strings.NewReader("S")}); err != nil {
		t.Fatal(err)
	}
	_, files = multipartParts(t, f.expect("PUT", "/api/gia-students/diplomas/501/presentation/file"))
	if files["file"] != "slides.pdf:S" {
		t.Errorf("files = %v", files)
	}

	if err := s.AddPresentationExtraFile(ctx, 501, myitmo.Upload{Name: "extra.zip", Content: strings.NewReader("Z")}); err != nil {
		t.Fatal(err)
	}
	_, files = multipartParts(t, f.expect("POST", "/api/gia-students/diplomas/501/presentation/extra-files"))
	if files["file"] != "extra.zip:Z" {
		t.Errorf("files = %v", files)
	}
}

func TestGIAStudentsDownloads(t *testing.T) {
	ctx := context.Background()
	cases := []struct {
		name string
		call func(s *myitmo.GIAStudentsService) (*myitmo.File, error)
		path string
	}{
		{"DownloadFile", func(s *myitmo.GIAStudentsService) (*myitmo.File, error) { return s.DownloadFile(ctx, 501, "key 1") }, "/api/gia-students/diplomas/501/file/download/key%201"},
		{"AntiplagiatReport", func(s *myitmo.GIAStudentsService) (*myitmo.File, error) { return s.AntiplagiatReport(ctx, 501) }, "/api/gia-students/diplomas/501/file/antiplagiat/report"},
		{"SupervisorReviewFile", func(s *myitmo.GIAStudentsService) (*myitmo.File, error) { return s.SupervisorReviewFile(ctx, 501) }, "/api/gia-students/diplomas/501/review/supervisor/file"},
		{"ReviewerReviewFile", func(s *myitmo.GIAStudentsService) (*myitmo.File, error) { return s.ReviewerReviewFile(ctx, 501, 77) }, "/api/gia-students/diplomas/501/review/reviewer/77/file"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			f, c := newFake(t)
			f.replyWith(http.StatusOK, http.Header{"Content-Type": {"application/pdf"}}, "%PDF-1.7")
			file := must[*myitmo.File](t)(tc.call(c.GIAStudents))
			defer file.Body.Close()
			f.expect("GET", tc.path)
			b, _ := io.ReadAll(file.Body)
			if string(b) != "%PDF-1.7" || file.ContentType != "application/pdf" {
				t.Errorf("file = %q %q", file.ContentType, b)
			}
		})
	}
}

func TestGIAStudentsDecodeWorkflow(t *testing.T) {
	ctx := context.Background()
	f, c := newFake(t)
	s := c.GIAStudents

	f.result(`[{"diploma_id":501,"education_level_name":"Бакалавриат","ep_name":"Программная инженерия","ep_enrollment_year":2022,
		"faculty_name":"ФПИиКТ","ep_direction_code":"09.03.04","ep_direction_name":"Программная инженерия","ep_language":"ru"}]`)
	ds := must[[]myitmo.GIAStudentDiploma](t)(s.Diplomas(ctx))
	if len(ds) != 1 || ds[0].DiplomaID != 501 || string(ds[0].EPEnrollmentYear) != "2022" {
		t.Errorf("diplomas = %+v", ds)
	}

	f.result(`{"diploma_id":501,"stages":[{"status_id":4,"deadline":"01.12.2025"},{"status_id":3,"deadline":null}],
		"contacts":[{"role":"Руководитель","fio":"Тестов Тест Тестович","isu":100001,"email":"t@example.com","phone":null,"photo_url":"https://example.com/p/"}],
		"rpd_gia_link":"https://example.com/rpd.pdf"}`)
	info := must[*myitmo.GIAStudentDiplomaInfo](t)(s.DiplomaInfo(ctx, 501))
	if len(info.Stages) != 2 || info.Stages[1].StatusID != myitmo.GIAStudentStatusRejected || info.Stages[0].Deadline != "01.12.2025" ||
		info.Contacts[0].ISU == nil || *info.Contacts[0].ISU != 100001 || info.Contacts[0].Phone != "" {
		t.Errorf("info = %+v", info)
	}

	f.result(`{"diploma_type_id":6,"diploma_type_name":"Проект","diploma_theme":"Тема","diploma_theme_en":null,"diploma_justification":null,
		"diploma_external_partner":false,"diploma_external_partner_name":null,"diploma_fundamental":false,"supervisor_isu":null,"supervisor_fio":"  ",
		"supervisor_job_title":null,"consultants":[{"isu":100002}],"co_supervisors":[{"surname":"Петров","name":"Пётр","second_name":null,
		"work_place":"ООО Пример","job_title":"Инженер","degree_type_id":null,"academic_title_id":2}],
		"comment":"Уточните тему","comment_author_role":"Руководитель ОП","comment_author_fio":"Иванов И. И."}`)
	app := must[*myitmo.GIAStudentApplication](t)(s.Application(ctx, 501))
	if app.SupervisorISU != nil || app.Comment != "Уточните тему" || *app.Consultants[0].ISU != 100002 ||
		*app.CoSupervisors[0].AcademicTitleID != 2 || app.CoSupervisors[0].WorkPlace != "ООО Пример" {
		t.Errorf("application = %+v", app)
	}

	f.result(`{"main_questions":{"time":1700000000000,"blocks":[{"id":"q1","type":"list","data":{"style":"ordered","items":["A","B"]}}],"version":"2.28.0"},
		"format":"Отчёт","extra_info":null,"comment":null,"comment_author_role":null,"comment_author_fio":null}`)
	task := must[*myitmo.GIAStudentTask](t)(s.Task(ctx, 501))
	if task.MainQuestions.IsEmpty() || task.MainQuestions.Blocks[0].Type != "list" || !task.ExtraInfo.IsEmpty() {
		t.Errorf("task = %+v", task)
	}

	f.result(`{"goal":{"blocks":[]},"tasks":null,"results":null,"extra_info":null,"grants":[{"id":1,"name":"Грант","year":2024}],
		"publications":[{"id":2,"name":"Статья","year":"2023"}],"speeches":null,"comment":null}`)
	ann := must[*myitmo.GIAStudentAnnotation](t)(s.Annotation(ctx, 501))
	if !ann.Goal.IsEmpty() || ann.Grants[0].ID != 1 || string(ann.Publications[0].Year) != `"2023"` {
		t.Errorf("annotation = %+v", ann)
	}

	f.result(`[{"id":1,"name":"Грант","year":2024}]`)
	if g := must[[]myitmo.GIAStudentAchievement](t)(s.Grants(ctx)); len(g) != 1 || g[0].Name != "Грант" {
		t.Errorf("grants = %+v", g)
	}
}

func TestGIAStudentsDecodeFileAndDefense(t *testing.T) {
	ctx := context.Background()
	f, c := newFake(t)
	s := c.GIAStudents

	f.result(`{"student_file_key":"k1","student_file_name":"work.pdf","publish":true,"result_file_key":null,
		"antiplagiat_report_web_url":"https://example.com/r","originality":87.5,"citations":5,"similarity":7.5,"self_citations":0,
		"comment":null,"supervisor_review":{"student_signature":null,"initiative_goals_name":"Высокий","advantages":"Плюсы",
		"assessment_name":"Отлично","thesis_complete":true,"awarding_qualifications":true},
		"reviewer_reviews":[{"reviewer_id":77,"student_signature":"2025-06-01","value_name":"Средний","awarding_qualifications":"да"}],
		"reviewer_review_questions":[{"id":1,"reviewer_id":77,"question":"Почему?"}],"status_id":4,"name":"ВКР"}`)
	file := must[*myitmo.GIAStudentDiplomaFile](t)(s.File(ctx, 501))
	if file.StudentFileKey != "k1" || *file.Originality != 87.5 || file.SupervisorReview.StudentSignature != nil ||
		file.SupervisorReview.InitiativeGoalsName != "Высокий" || file.ReviewerReviews[0].ReviewerID != 77 ||
		*file.ReviewerReviews[0].StudentSignature != "2025-06-01" || file.ReviewerReviewQuestions[0].Question != "Почему?" {
		t.Errorf("file = %+v", file)
	}

	f.result(`{"possible_presentations":[{"presentation_id":12,"presentation_defense_type_id":1,"presentation_defense_type_name":"Онлайн",
		"defense_date":"2025-06-20","start_time":"10:00:00","link":"https://example.com/meet","address":null,"room_name":null,
		"can_choose_defense_date":true}],"student_sign_up":{"presentation_id":null,"approved":false}}`)
	pp := must[*myitmo.GIAStudentPossiblePresentations](t)(s.PossiblePresentations(ctx, 501))
	if p := pp.PossiblePresentations[0]; p.DefenseDate != myitmo.NewDate(2025, 6, 20) || p.PresentationDefenseTypeID != myitmo.GIAStudentDefenseTypeOnline ||
		!p.CanChooseDefenseDate || pp.StudentSignUp.PresentationID != nil {
		t.Errorf("possible = %+v", pp)
	}

	f.result(`{"main_info":{"defense_date":"2025-06-20T00:00:00","start_time":"10:00:00","presentation_defense_type_name":"Очно",
		"address":"Кронверкский пр., 49","room_name":"1404","link":null,"group":"P3400","secretary":{"surname":"Секретарев","name":"С","second_name":"С"},
		"supervisor":{"surname":"Руков","name":"Р","second_name":null},"theme_ru":"Тема","assessment":5,"assessment_id":30,
		"assessment_name":"Отлично","average_mark":4.8,"red_diploma":true,"has_supervisor_review":true,"has_reviewer_review":false},
		"presentation":{"file_key":"p1","file_name":"slides.pdf"},"extra_files":{"files":[{"file_key":"e1","file_name":"extra.zip"}]},
		"documents":null,"student_list":{"students":[{"isu":100003,"surname":"Студентов","name":"С","second_name":"С"}]}}`)
	def := must[*myitmo.GIAStudentDefense](t)(s.Presentation(ctx, 501))
	if def.MainInfo.DefenseDate != myitmo.NewDate(2025, 6, 20) || def.MainInfo.Secretary.Surname != "Секретарев" ||
		string(def.MainInfo.Assessment) != "5" || *def.MainInfo.AverageMark != 4.8 || def.Presentation.FileKey != "p1" ||
		def.ExtraFiles.Files[0].FileKey != "e1" || def.Documents != nil || def.StudentList.Students[0].ISU != 100003 ||
		def.StudentList.Students[0].Surname != "Студентов" {
		t.Errorf("defense = %+v", def)
	}
}

func TestGIAStudentsDecodeSupplements(t *testing.T) {
	ctx := context.Background()
	f, c := newFake(t)
	s := c.GIAStudents

	f.result(`{"pers_info":{"status":{"status_id":7,"status_name":"Редактируется"},"surname":"Тестов","name":"Тест","second_name":"Тестович",
		"birth_date":"01.01.2003","snils":null,"is_foreigner":false,"dir_code":"09.03.04","dir_name":"Программная инженерия",
		"edu_level":"Бакалавриат","edu_period":4,"doc_obr":"Аттестат","theme_en":null},
		"disciplines":{"status":{"status_id":4,"status_name":"Подтверждено"},"blocks":[{"block_id":1,"block_name":"Блок 1","total_credits":180,
		"disc_info":[{"disc_id":31,"disc_name":"Математика","credit_units":6,"hours_amount":216,"avg_grade":"5","avg_grade_letter":"A"}]}],
		"avg_grade":4.7,"ungraded_count":0,"total_credits":240,"ep_volume":240,"changes":[]},
		"elective_disciplines":{"status":{"status_id":7,"status_name":"Редактируется"},"blocks":[],"avg_grade":0,"ungraded_count":0,"total_credits":0,"ep_volume":0},
		"additional_info":{"status":{"status_id":7,"status_name":"Редактируется"},"add_info":[{"description":"Победитель олимпиады","id":5,"choice":true}]},
		"deadline":"01.06.2025"}`)
	sup := must[*myitmo.GIAStudentSupplement](t)(s.Supplement(ctx, 501))
	if sup.PersInfo.Status.StatusID != myitmo.GIAStudentSupplementEditable || sup.PersInfo.SNILS != "" ||
		sup.Disciplines.Status.StatusID != myitmo.GIAStudentSupplementConfirmed || *sup.Disciplines.Blocks[0].TotalCredits != 180 ||
		sup.Disciplines.Blocks[0].DiscInfo[0].DiscID != 31 || !sup.AdditionalInfo.AddInfo[0].Choice || sup.Deadline != "01.06.2025" {
		t.Errorf("supplement = %+v", sup)
	}

	f.result(`{"status":{"status_id":28,"status_name":"Проверка студентом"},"payment_required":true,"payment_done":false,"reject_reason":null,
		"print_ready":false,"pers_info":{"isu":100004,"surname":"Testov","name":"Test","second_name":null,"birth_date":"2003-01-01",
		"birth_place":"Saint Petersburg","email":"student@example.com","foreigner":false,"print_attachment":true},
		"diploma_info":{"qualification":"Bachelor","ep_name":"Software Engineering","dir_name":"Software Engineering","theme":"Graph search",
		"surname":"Testov","name":"Test","edu_period":"4 years","total_credits":240,"total_volume":240,"edu_level":"Бакалавриат",
		"edu_level_en":"Bachelor","diploma_number":"000000","defence_date":"2025-06-20"},
		"disciplines":{"blocks":[{"block_id":1,"block_name":"Block 1","total_credits":180,"disc_info":[{"disc_id":31,"disc_name":"Mathematics",
		"credit_units":6,"avg_grade":5,"avg_grade_letter":"A"}]}]},"elective_disciplines":{"blocks":[]}}`)
	eu := must[*myitmo.GIAStudentEUSupplement](t)(s.EUSupplement(ctx, 501))
	if eu.Status.StatusID != myitmo.GIAStudentEUStudentReview || !eu.PaymentRequired || !eu.PersInfo.PrintAttachment ||
		eu.DiplomaInfo.EduLevelEN != "Bachelor" || eu.Disciplines.Blocks[0].DiscInfo[0].DiscName != "Mathematics" {
		t.Errorf("eu = %+v", eu)
	}

	f.reply(http.StatusNotFound, `{"error_code":404,"error_message":"not found"}`)
	_, err := s.EUSupplement(ctx, 501)
	if e, ok := errors.AsType[*myitmo.Error](err); !ok || !e.IsNotFound() {
		t.Errorf("err = %v, want not found", err)
	}
}

func TestGIAStudentsDecodeReferences(t *testing.T) {
	ctx := context.Background()
	f, c := newFake(t)
	s := c.GIAStudents

	f.result(`[{"isu":100001,"surname":"Тестов","name":"Тест","second_name":null,"job_title":"Доцент"}]`)
	if w := must[[]myitmo.GIAStudentWorker](t)(s.People(ctx, "Тес")); len(w) != 1 || w[0].ISU != 100001 || w[0].JobTitle != "Доцент" {
		t.Errorf("people = %+v", w)
	}
	f.expect("GET", "/api/gia-students/people")

	f.result(`[{"diploma_type_id":6,"diploma_type_name":"Другое"}]`)
	if dt := must[[]myitmo.GIAStudentDiplomaType](t)(s.DiplomaTypes(ctx)); dt[0].DiplomaTypeID != myitmo.GIAStudentDiplomaTypeOther {
		t.Errorf("types = %+v", dt)
	}

	f.result(`[{"id":8,"name":"На согласовании","color":"#FEF8E5"}]`)
	if st := must[[]myitmo.GIAStudentStatus](t)(s.Statuses(ctx, myitmo.GIAStudentStageApplication)); st[0].ID != 8 || st[0].Color != "#FEF8E5" {
		t.Errorf("statuses = %+v", st)
	}

	f.result(`[{"id":2,"name_ru":"Доцент"}]`)
	if at := must[[]myitmo.GIAStudentTitle](t)(s.AcademicTitles(ctx)); at[0].NameRU != "Доцент" {
		t.Errorf("titles = %+v", at)
	}

	f.result(`{"is_student":true}`)
	if raw := must[myitmo.RawJSON](t)(s.UserStatus(ctx)); string(raw) != `{"is_student":true}` {
		t.Errorf("status = %s", raw)
	}
}
