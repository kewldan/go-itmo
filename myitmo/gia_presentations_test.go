package myitmo_test

import (
	"errors"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/kewldan/go-itmo/myitmo"
)

func TestGIAPresentationRequests(t *testing.T) {
	day := time.Date(2025, 6, 20, 0, 0, 0, 0, time.UTC)
	in := myitmo.GIAPresentationInput{
		EduProgramID: 7, DefenseDate: day, PresentationDefenseTypeID: 1, Building: "Корпус", Auditorium: "101",
		GeneralStateDate: day.Add(9 * time.Hour), StartTime: day.Add(10 * time.Hour), EndTime: day.Add(14 * time.Hour),
		IsScratch: true, Link: "https://meet.example.com/x", IsGeneralState: true,
		BookingInfo: &myitmo.GIABookingInfo{
			StartDatetime: time.Date(2025, 6, 20, 10, 0, 0, 0, myitmo.MSK), EndDatetime: time.Date(2025, 6, 20, 14, 30, 0, 0, myitmo.MSK),
			RoomID: 3, GroupID: 4, CategoryID: 5, RoomName: "101", Address: "ул. Тестовая, 1",
		},
	}
	bodyFmt := `{"edu_program_id":7,"defense_date":"2025-06-20T00:00:00Z","presentation_defense_type_id":1,
		"can_choose_defense_date":false,"building":"Корпус","auditorium":"101","general_state_date":"2025-06-20T09:00:00Z",
		"start_time":"2025-06-20T10:00:00Z","end_time":"2025-06-20T14:00:00Z","is_scratch":true,"link":"https://meet.example.com/x",%s
		"booking_info":{"co_bookers":[],"start_datetime":"2025-06-20 10:00","end_datetime":"2025-06-20 14:30","event_id":null,
		"room_id":3,"group_id":4,"category_id":5,"room_name":"101","address":"ул. Тестовая, 1"}}`
	p := myitmo.GIAPresentationsParams{Limit: 15, Year: 2025, StatusID: 2, FacultyID: 3, EduProgramID: 7, DirID: 8,
		EduLevelID: 1, Date: "2025-06-20", ShowAs: myitmo.GIARoleAdmin, Query: "тест"}
	pq := url.Values{"limit": {"15"}, "offset": {"0"}, "year": {"2025"}, "status_id": {"2"}, "faculty_id": {"3"},
		"edu_program_id": {"7"}, "dir_id": {"8"}, "edu_level_id": {"1"}, "date": {"2025-06-20"}, "show_as": {"admin"}, "query": {"тест"}}
	mark := int64(5)
	runGIACases(t, []giaCase{
		{"Presentations", func(c *myitmo.Client) error { _, err := c.GIA.Presentations(ctx, p); return err },
			"GET", "/api/gia/presentations", pq, ""},
		{"CreatePresentation", func(c *myitmo.Client) error { return c.GIA.CreatePresentation(ctx, in) },
			"POST", "/api/gia/presentations", nil, giaf(bodyFmt, `"is_general_state":true,`)},
		{"Presentation", func(c *myitmo.Client) error { _, err := c.GIA.Presentation(ctx, 21); return err },
			"GET", "/api/gia/presentations/21", nil, ""},
		{"UpdatePresentation", func(c *myitmo.Client) error { return c.GIA.UpdatePresentation(ctx, 21, in) },
			"PATCH", "/api/gia/presentations/21", nil, giaf(bodyFmt, "")},
		{"ApprovePresentation", func(c *myitmo.Client) error { return c.GIA.ApprovePresentation(ctx, 21) },
			"PATCH", "/api/gia/presentations/21/status", nil, `{}`},
		{"RejectPresentation", func(c *myitmo.Client) error { return c.GIA.RejectPresentation(ctx, 21, "занято") },
			"PATCH", "/api/gia/presentations/21/status", nil, `{"comment":"занято"}`},
		{"RevokePresentation", func(c *myitmo.Client) error { return c.GIA.RevokePresentation(ctx, 21) },
			"POST", "/api/gia/presentations/21/revoke", nil, ""},
		{"SetPresentationLink", func(c *myitmo.Client) error { return c.GIA.SetPresentationLink(ctx, 21, "https://meet.example.com/y") },
			"POST", "/api/gia/presentations/21/links", nil, `{"link":"https://meet.example.com/y"}`},
		{"AttachCommittee", func(c *myitmo.Client) error { return c.GIA.AttachCommittee(ctx, 21, 3) },
			"POST", "/api/gia/presentations/21/committees", nil, `{"committee_id":3}`},
		{"ApproveMarks", func(c *myitmo.Client) error { return c.GIA.ApproveMarks(ctx, 21) },
			"POST", "/api/gia/presentations/21/marks/approve", nil, ""},
		{"SetPresentationStudents", func(c *myitmo.Client) error { return c.GIA.SetPresentationStudents(ctx, 21, 11, 12) },
			"PATCH", "/api/gia/presentations/21/students", nil, `[11,12]`},
		{"OrderPresentationStudents", func(c *myitmo.Client) error {
			return c.GIA.OrderPresentationStudents(ctx, 21, []myitmo.GIAQueuePosition{{StudentID: 12, OrderInQueue: 1}, {StudentID: 11, OrderInQueue: 2}})
		}, "PATCH", "/api/gia/presentations/21/students/order", nil, `[{"student_id":12,"order_in_queue":1},{"student_id":11,"order_in_queue":2}]`},
		{"SetPresentationMarks", func(c *myitmo.Client) error {
			return c.GIA.SetPresentationMarks(ctx, 21, []myitmo.GIAMarkEntry{{StudentID: 11, Mark: &mark}, {StudentID: 12}})
		}, "PATCH", "/api/gia/presentations/21/students/marks", nil, `[{"student_id":11,"mark":5},{"student_id":12,"mark":null}]`},
		{"ApprovePresentationStudent", func(c *myitmo.Client) error { return c.GIA.ApprovePresentationStudent(ctx, 21, 11) },
			"PATCH", "/api/gia/presentations/21/students/11/approve", nil, ""},
		{"RejectPresentationStudent", func(c *myitmo.Client) error { return c.GIA.RejectPresentationStudent(ctx, 21, 11) },
			"PATCH", "/api/gia/presentations/21/students/11/reject", nil, ""},
		{"DefenseQuestions", func(c *myitmo.Client) error { _, err := c.GIA.DefenseQuestions(ctx, 21, 11); return err },
			"GET", "/api/gia/presentations/21/students/11/questions", nil, ""},
		{"SetDefenseQuestions", func(c *myitmo.Client) error {
			return c.GIA.SetDefenseQuestions(ctx, 21, 11, true, []myitmo.GIADefenseQuestion{{ID: 9, Question: "Почему?", AnswerQuality: "полный"}})
		}, "PATCH", "/api/gia/presentations/21/students/11/questions", nil,
			`{"approved":true,"questions":[{"question":"Почему?","answer_quality":"полный"}]}`},
		{"DefenseStudent", func(c *myitmo.Client) error { _, err := c.GIA.DefenseStudent(ctx, 11); return err },
			"GET", "/api/gia/presentations/students/11/main-info", nil, ""},
		{"PresentationCandidates", func(c *myitmo.Client) error { _, err := c.GIA.PresentationCandidates(ctx, 7); return err },
			"GET", "/api/gia/presentations/students", url.Values{"edu_program_id": {"7"}}, ""},
		{"SecretaryPresentations", func(c *myitmo.Client) error { _, err := c.GIA.SecretaryPresentations(ctx, p); return err },
			"GET", "/api/gia/presentations/secretary", pq, ""},
		{"GeneralStatePresentations", func(c *myitmo.Client) error { _, err := c.GIA.GeneralStatePresentations(ctx, p); return err },
			"GET", "/api/gia/presentations/general-state", pq, ""},
		{"PresentationFilters", func(c *myitmo.Client) error {
			_, err := c.GIA.PresentationFilters(ctx, myitmo.GIARoleAdmin)
			return err
		},
			"GET", "/api/gia/presentations/filters", url.Values{"show_as": {"admin"}}, ""},
		{"GeneralStatePresentationFilters", func(c *myitmo.Client) error { _, err := c.GIA.GeneralStatePresentationFilters(ctx, ""); return err },
			"GET", "/api/gia/presentations/general-state/filters", nil, ""},
		{"PresentationCommittees", func(c *myitmo.Client) error { _, err := c.GIA.PresentationCommittees(ctx, 7); return err },
			"GET", "/api/gia/presentations/committees", url.Values{"edu_program_id": {"7"}}, ""},
		{"PresentationPrograms", func(c *myitmo.Client) error { _, err := c.GIA.PresentationPrograms(ctx, "тест", 15); return err },
			"GET", "/api/gia/presentations/edu-programs", url.Values{"query": {"тест"}, "limit": {"15"}}, ""},
		{"DefenseTypes", func(c *myitmo.Client) error { _, err := c.GIA.DefenseTypes(ctx); return err },
			"GET", "/api/gia/presentations/defense-types", nil, ""},
	})
	docx := url.Values{"format": {"docx"}}
	pdf := url.Values{"format": {"pdf"}}
	runGIADownloads(t, []giaDownload{
		{"IndividualProtocols", func(c *myitmo.Client) (*myitmo.File, error) {
			return c.GIA.IndividualProtocols(ctx, 21, myitmo.GIAFormatDOCX)
		}, "GET", "/api/gia/presentations/21/indv-protocols-zip", docx, ""},
		{"AttendanceSheet", func(c *myitmo.Client) (*myitmo.File, error) {
			return c.GIA.AttendanceSheet(ctx, 21, myitmo.GIAFormatPDF)
		}, "GET", "/api/gia/presentations/21/attendance-sheet", pdf, ""},
		{"DefenseOrder", func(c *myitmo.Client) (*myitmo.File, error) {
			return c.GIA.DefenseOrder(ctx, 21, myitmo.GIAFormatDOCX)
		}, "GET", "/api/gia/presentations/21/defense-order", docx, ""},
		{"DefenseAct", func(c *myitmo.Client) (*myitmo.File, error) {
			return c.GIA.DefenseAct(ctx, 21, myitmo.GIAFormatDOCX)
		}, "GET", "/api/gia/presentations/21/defense-act", docx, ""},
		{"ChairmanConclusion", func(c *myitmo.Client) (*myitmo.File, error) {
			return c.GIA.ChairmanConclusion(ctx, 21, myitmo.GIAFormatDOCX)
		}, "GET", "/api/gia/presentations/21/chairman-conclusion", docx, ""},
		{"MarkTemplate", func(c *myitmo.Client) (*myitmo.File, error) {
			return c.GIA.MarkTemplate(ctx, 21, myitmo.GIAFormatDOCX)
		}, "GET", "/api/gia/presentations/21/mark-template", docx, ""},
		{"PresentationFiles", func(c *myitmo.Client) (*myitmo.File, error) {
			return c.GIA.PresentationFiles(ctx, 21, myitmo.GIAFormatPDF)
		}, "GET", "/api/gia/presentations/21/all-files", pdf, ""},
		{"PresentationMaterials", func(c *myitmo.Client) (*myitmo.File, error) { return c.GIA.PresentationMaterials(ctx, 21) },
			"GET", "/api/gia/presentations/21/materials-zip", nil, ""},
		{"IndividualProtocol", func(c *myitmo.Client) (*myitmo.File, error) { return c.GIA.IndividualProtocol(ctx, 11, "") },
			"GET", "/api/gia/presentations/students/11/indv-protocol", nil, ""},
	})
}

func TestGIAApproveMarksErrors(t *testing.T) {
	f, c := newFake(t)
	f.reply(http.StatusBadRequest, `{"error_code":"presentation_not_defense_passed","error_message":"Защита еще не прошла","result":null}`)
	err := c.GIA.ApproveMarks(ctx, 21)
	e, ok := errors.AsType[*myitmo.Error](err)
	if !ok || e.StatusCode != http.StatusBadRequest || e.Code != 0 ||
		!strings.HasPrefix(e.Message, myitmo.GIAErrPresentationNotDefensePassed+": ") {
		t.Fatalf("err = %#v", err)
	}

	f.reply(http.StatusOK, `{"error_code":"presentation_not_defense_passed","error_message":""}`)
	e, ok = errors.AsType[*myitmo.Error](c.GIA.ApproveMarks(ctx, 21))
	if !ok || e.StatusCode != http.StatusOK || e.Message != myitmo.GIAErrPresentationNotDefensePassed {
		t.Fatalf("err = %#v", e)
	}

	f.reply(http.StatusOK, `{"error_code":17,"error_message":"нельзя"}`)
	if code := myitmo.ErrorCode(c.GIA.ApproveMarks(ctx, 21)); code != 17 {
		t.Errorf("code = %d", code)
	}

	f.reply(http.StatusOK, `{"error_code":0,"error_message":null,"result":null}`)
	if err := c.GIA.ApproveMarks(ctx, 21); err != nil {
		t.Errorf("err = %v", err)
	}
	f.reply(http.StatusOK, `{"error_code":"0"}`)
	if err := c.GIA.ApproveMarks(ctx, 21); err != nil {
		t.Errorf("err = %v", err)
	}
	f.reply(http.StatusOK, ``)
	if err := c.GIA.ApproveMarks(ctx, 21); err != nil {
		t.Errorf("err = %v", err)
	}
	f.reply(http.StatusInternalServerError, `oops`)
	if e, ok := errors.AsType[*myitmo.Error](c.GIA.ApproveMarks(ctx, 21)); !ok || e.StatusCode != http.StatusInternalServerError {
		t.Errorf("err = %v", e)
	}
}

func TestGIAPresentationDecodes(t *testing.T) {
	f, c := newFake(t)
	f.result(`{"id":21,"ep_id":7,"ep_name":"Программа","ep_enrollment_year":2021,"language":"ru","defense_date":"2025-06-20T00:00:00Z",
		"start_time":"2025-06-20T10:00:00Z","end_time":"2025-06-20T14:00:00Z","general_state_date":"2025-06-20T09:00:00Z",
		"link":"","presentation_defense_type_id":1,"presentation_defense_type_name":"Очно","can_choose_defense_date":true,
		"status_id":13,"status_name":"Оценки переданы","reservation_status_id":14,"room_id":3,"room_name":"101",
		"committee":{"chairman_surname":"Тестов","internal_experts":[],"external_experts":[]},
		"students":[{"student_id":11,"student_isu":100001,"student_surname":"Студентов","approved":true,"average_mark":4.5,
		"red_diploma":false,"has_threes":false,"diploma_ready":true,"supervisor_mark":"отлично","reviewer_mark":5,"general_state_mark":3}]}`)
	day := must[*myitmo.GIAPresentation](t)(c.GIA.Presentation(ctx, 21))
	if day.StatusID != myitmo.GIAPresentationStatusMarksSent || day.ReservationStatusID != myitmo.GIAReservationBooked ||
		day.Committee.ChairmanSurname != "Тестов" || day.Students[0].AverageMark != 4.5 || string(day.Students[0].ReviewerMark) != "5" ||
		day.StartTime.Hour() != 10 {
		t.Errorf("day = %+v", day)
	}

	f.result(`{"presentations":[{"presentation_id":21,"defense_date":"2025-06-20","place":"101","chairman":"Тестов Т.Т.",
		"create_at":"2025-05-01T12:00:00","status_id":2}],"total":1}`)
	list := must[*myitmo.GIAPresentationList](t)(c.GIA.Presentations(ctx, myitmo.GIAPresentationsParams{}))
	if list.Total != 1 || list.Presentations[0].CreateAt.Month() != time.May || list.Presentations[0].DefenseDate.Day() != 20 {
		t.Errorf("list = %+v", list)
	}

	f.result(`{"Z3400":[{"student_isu":100001,"student_surname":"Студентов","date":"","appointed":true}],"Z3401":[]}`)
	cand := must[map[string][]myitmo.GIAPresentationCandidate](t)(c.GIA.PresentationCandidates(ctx, 7))
	if len(cand) != 2 || !cand["Z3400"][0].Appointed {
		t.Errorf("candidates = %+v", cand)
	}

	f.result(`{"student_id":11,"student_isu":100001,"diploma_id":31,"theme_ru":"Тема","mark":"отлично","mark_id":1,
		"result_file":{"key":"k1","name":"vkr","extension":"pdf"},"presentation":null,"extra_files":[{"key":"k2","name":"code","extension":"zip"}],
		"questions":[{"id":9,"question":"Почему?","answer_quality":"полный"}]}`)
	st := must[*myitmo.GIADefenseStudent](t)(c.GIA.DefenseStudent(ctx, 11))
	if st.ResultFile.Key != "k1" || st.Presentation != nil || st.ExtraFiles[0].Extension != "zip" || st.Questions[0].ID != 9 {
		t.Errorf("student = %+v", st)
	}

	f.result(`{"year_filters":[{"year":2025}],"defense_date_filters":[{"value":"2025-06-20","name":"20.06.2025"}],
		"role_filters":[{"id":"admin","name":"Администратор"}]}`)
	fl := must[*myitmo.GIAPresentationFilters](t)(c.GIA.PresentationFilters(ctx, ""))
	if fl.DefenseDateFilters[0].Value != "2025-06-20" || string(fl.RoleFilters[0].ID) != `"admin"` {
		t.Errorf("filters = %+v", fl)
	}
}
