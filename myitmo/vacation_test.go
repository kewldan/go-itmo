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

func TestVacationRequests(t *testing.T) {
	d1, d2 := myitmo.NewDate(2026, time.July, 1), myitmo.NewDate(2026, time.July, 14)
	draft := myitmo.VacationStatusDraft
	tests := []struct {
		name   string
		call   func(c *myitmo.Client) error
		method string
		path   string
		query  url.Values
		body   string
	}{
		{"Access", func(c *myitmo.Client) error { _, err := c.Vacation.Access(ctx); return err },
			"GET", "/api/vacation/me/access", url.Values{}, ""},
		{"Filters", func(c *myitmo.Client) error { _, err := c.Vacation.Filters(ctx); return err },
			"GET", "/api/vacation/filters", url.Values{}, ""},
		{"PersonalFilters", func(c *myitmo.Client) error { _, err := c.Vacation.PersonalFilters(ctx); return err },
			"GET", "/api/vacation/filters/personal", url.Values{}, ""},
		{"Types", func(c *myitmo.Client) error { _, err := c.Vacation.Types(ctx); return err },
			"GET", "/api/vacation/filters/types", url.Values{}, ""},
		{"UnpaidReasons", func(c *myitmo.Client) error { _, err := c.Vacation.UnpaidReasons(ctx); return err },
			"GET", "/api/vacation/filters/uto_reasons", url.Values{}, ""},
		{"Positions", func(c *myitmo.Client) error { _, err := c.Vacation.Positions(ctx); return err },
			"GET", "/api/vacation/filters/positions", url.Values{}, ""},
		{"People", func(c *myitmo.Client) error { _, err := c.Vacation.People(ctx, "Иванов"); return err },
			"GET", "/api/vacation/filters/people", url.Values{"query": {"Иванов"}}, ""},
		{"Assignees", func(c *myitmo.Client) error { _, err := c.Vacation.Assignees(ctx); return err },
			"GET", "/api/vacation/filters/assigned", url.Values{}, ""},
		{"Calendar", func(c *myitmo.Client) error { _, err := c.Vacation.Calendar(ctx, 2026); return err },
			"GET", "/api/vacation/calendar", url.Values{"year": {"2026"}}, ""},
		{"MyCalendar", func(c *myitmo.Client) error { _, err := c.Vacation.MyCalendar(ctx, d1, d2); return err },
			"GET", "/api/vacation/me/calendar", url.Values{"date_start": {"2026-07-01"}, "date_end": {"2026-07-14"}}, ""},
		{"Balance", func(c *myitmo.Client) error { _, err := c.Vacation.Balance(ctx, d1); return err },
			"GET", "/api/vacation/me/balance", url.Values{"date_start": {"2026-07-01"}}, ""},
		{"History", func(c *myitmo.Client) error {
			_, err := c.Vacation.History(ctx, myitmo.VacationHistoryParams{From: d1, DepartmentID: 5, Type: myitmo.VacationTypePaid, Status: &draft, Project: "P-1"})
			return err
		}, "GET", "/api/vacation/me/vacations", url.Values{"date_start": {"2026-07-01"}, "depId": {"5"}, "type": {"1"}, "stat": {"0"}, "proj": {"P-1"}}, ""},
		{"Create", func(c *myitmo.Client) error {
			_, err := c.Vacation.Create(ctx, []myitmo.VacationApplication{{EmployeeID: 11, TypeID: myitmo.VacationTypePaid, DateStart: d1, DateEnd: d2, Duration: 14, Digital: true}})
			return err
		}, "POST", "/api/vacation/me/vacations", url.Values{},
			`[{"employee_id":11,"type_id":1,"date_start":"2026-07-01","date_end":"2026-07-14","duration":14,"digital":true,"additional_data":null,"extra_fields":null}]`},
		{"Delete", func(c *myitmo.Client) error { return c.Vacation.Delete(ctx, 42) },
			"DELETE", "/api/vacation/me/vacations/42", url.Values{}, ""},
		{"Upcoming", func(c *myitmo.Client) error { _, err := c.Vacation.Upcoming(ctx); return err },
			"GET", "/api/vacation/me/vacations/upcoming", url.Values{}, ""},
		{"Conflicts", func(c *myitmo.Client) error { _, err := c.Vacation.Conflicts(ctx, d1, d2, 11, 12); return err },
			"GET", "/api/vacation/me/vacation_conflict_with_days", url.Values{"date_start": {"2026-07-01"}, "date_end": {"2026-07-14"}, "employee_id": {"11", "12"}}, ""},
		{"List", func(c *myitmo.Client) error {
			_, err := c.Vacation.List(ctx, myitmo.VacationListParams{Limit: 20, PositionID: 3, ISU: 100001, Category: "ППС", From: d1, To: d2,
				Statuses: []myitmo.VacationStatus{myitmo.VacationStatusPending, myitmo.VacationStatusSigned}})
			return err
		}, "GET", "/api/vacation/vacations", url.Values{"limit": {"20"}, "offset": {"0"}, "posId": {"3"}, "query": {"100001"}, "category": {"ППС"},
			"date_start": {"2026-07-01"}, "date_end": {"2026-07-14"}, "stat": {"1", "3"}}, ""},
		{"Employees", func(c *myitmo.Client) error {
			_, err := c.Vacation.Employees(ctx, myitmo.VacationEmployeesParams{DepartmentID: 7, Project: "9", From: d1})
			return err
		}, "GET", "/api/vacation/employees", url.Values{"depId": {"7"}, "proj": {"9"}, "date_start": {"2026-07-01"}}, ""},
		{"PendingApproval", func(c *myitmo.Client) error {
			_, err := c.Vacation.PendingApproval(ctx, myitmo.VacationAgreementParams{Type: myitmo.VacationTypeUnpaid, Assigned: 100002, Limit: 10, Offset: 20})
			return err
		}, "GET", "/api/vacation/pending-approval", url.Values{"type": {"2"}, "assigned": {"100002"}, "limit": {"10"}, "offset": {"20"}}, ""},
		{"PendingApprovalCount", func(c *myitmo.Client) error { _, err := c.Vacation.PendingApprovalCount(ctx); return err },
			"GET", "/api/vacation/pending-approval/count", url.Values{}, ""},
		{"PendingApprovalTask", func(c *myitmo.Client) error { _, err := c.Vacation.PendingApprovalTask(ctx, 42); return err },
			"GET", "/api/vacation/pending-approval/42", url.Values{}, ""},
		{"Unapproved", func(c *myitmo.Client) error {
			_, err := c.Vacation.Unapproved(ctx, myitmo.VacationAgreementParams{PositionID: 4, ISU: 100003})
			return err
		}, "GET", "/api/vacation/unapproved-vacations", url.Values{"posId": {"4"}, "query": {"100003"}}, ""},
		{"Approve", func(c *myitmo.Client) error { return c.Vacation.Approve(ctx, 1, 2, 3) },
			"PATCH", "/api/vacation/vacations/approve", url.Values{}, `[1,2,3]`},
		{"Reject", func(c *myitmo.Client) error { return c.Vacation.Reject(ctx, 42, "") },
			"PATCH", "/api/vacation/vacation/42/reject", url.Values{}, `{"id":42,"rejection_reason":""}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, c := newFake(t)
			if err := tt.call(c); err != nil {
				t.Fatal(err)
			}
			r := f.expect(tt.method, tt.path)
			if !reflect.DeepEqual(r.Query, tt.query) {
				t.Errorf("query = %v, want %v", r.Query, tt.query)
			}
			if tt.body != "" {
				r.sameJSON(t, tt.body)
			} else if len(r.Body) != 0 {
				t.Errorf("unexpected body %s", r.Body)
			}
		})
	}
}

func TestVacationDownloads(t *testing.T) {
	tests := []struct {
		name  string
		call  func(c *myitmo.Client) (*myitmo.File, error)
		path  string
		query url.Values
	}{
		{"PDF", func(c *myitmo.Client) (*myitmo.File, error) { return c.Vacation.PDF(ctx, 11, 42) },
			"/api/vacation/static/vacations/11/42/pdf", url.Values{}},
		{"Signed", func(c *myitmo.Client) (*myitmo.File, error) { return c.Vacation.Signed(ctx, 42) },
			"/api/vacation/vacations/42/signed", url.Values{}},
		{"Original", func(c *myitmo.Client) (*myitmo.File, error) { return c.Vacation.Original(ctx, 42) },
			"/api/vacation/vacations/42/signed", url.Values{"type": {"original"}}},
		{"Archive", func(c *myitmo.Client) (*myitmo.File, error) { return c.Vacation.Archive(ctx, 42) },
			"/api/vacation/vacations/42/file-archive", url.Values{}},
		{"Download", func(c *myitmo.Client) (*myitmo.File, error) { return c.Vacation.Download(ctx, 42, 43) },
			"/api/vacation/vacations/download", url.Values{"ids": {"42", "43"}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, c := newFake(t)
			f.replyWith(http.StatusOK, http.Header{
				"Content-Type":        {"application/pdf"},
				"Content-Disposition": {`attachment; filename="app.pdf"`},
			}, "%PDF-1.4")
			file := must[*myitmo.File](t)(tt.call(c))
			defer file.Body.Close()
			r := f.expect(http.MethodGet, tt.path)
			if !reflect.DeepEqual(r.Query, tt.query) {
				t.Errorf("query = %v, want %v", r.Query, tt.query)
			}
			body, _ := io.ReadAll(file.Body)
			if file.Name != "app.pdf" || string(body) != "%PDF-1.4" {
				t.Errorf("file = %q %q", file.Name, body)
			}
		})
	}
}

func TestVacationCreateUnpaid(t *testing.T) {
	f, c := newFake(t)
	f.result(`[{"id":77,"signature_id":"501","task_id":601}]`)
	apps := []myitmo.VacationApplication{{
		EmployeeID: 11, TypeID: myitmo.VacationTypeUnpaid,
		DateStart: myitmo.NewDate(2026, time.March, 2), DateEnd: myitmo.NewDate(2026, time.March, 3), Duration: 2,
		AdditionalData: &myitmo.VacationReason{ReasonID: myitmo.VacationReasonPersonal, PersonalReason: "семейные обстоятельства"},
	}}
	got := must[[]myitmo.VacationCreated](t)(c.Vacation.CreateUnpaid(ctx, apps,
		myitmo.Upload{Name: "note.pdf", ContentType: "application/pdf", Content: strings.NewReader("doc")}))
	if len(got) != 1 || got[0].ID != 77 || got[0].SignatureID != 501 || got[0].TaskID != 601 {
		t.Errorf("created = %+v", got)
	}
	r := f.expect(http.MethodPost, "/api/vacation/me/vacations/uto")
	_, params, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil {
		t.Fatal(err)
	}
	mr := multipart.NewReader(strings.NewReader(string(r.Body)), params["boundary"])
	parts := map[string]string{}
	var fileName string
	for {
		p, err := mr.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		b, _ := io.ReadAll(p)
		parts[p.FormName()] = string(b)
		if p.FormName() == "files" {
			fileName = p.FileName()
		}
	}
	if parts["files"] != "doc" || fileName != "note.pdf" {
		t.Errorf("files part = %q %q", fileName, parts["files"])
	}
	recorded{Body: []byte(parts["vacations"])}.sameJSON(t, `[{"employee_id":11,"type_id":2,"date_start":"2026-03-02","date_end":"2026-03-03","duration":2,"digital":false,
		"additional_data":{"reason_id":7,"personal_reason":"семейные обстоятельства"},"file_names":["note.pdf"],"extra_fields":null}]`)
}

func TestVacationDecode(t *testing.T) {
	f, c := newFake(t)

	f.result(`{"access":true,"e_vacation":true,"admin":false,"supervisor":true,"hr":null,"employer":true}`)
	acc := must[*myitmo.VacationAccess](t)(c.Vacation.Access(ctx))
	if !acc.Access || !acc.EVacation || !acc.Supervisor || acc.HR || !acc.Employer {
		t.Errorf("access = %+v", acc)
	}

	f.result(`{"positions":[{"id":1,"value":"Инженер"}],"departments":[{"id":2,"value":"Отдел"}],"statuses":[{"id":-1,"value":"Отклонено"}],
		"projects":[{"id":"P-1","value":"Проект"}],"category":[{"id":3,"value":"АУП"}]}`)
	fl := must[*myitmo.VacationFilters](t)(c.Vacation.Filters(ctx))
	if fl.Statuses[0].ID != -1 || string(fl.Projects[0].ID) != `"P-1"` || fl.Category[0].Value != "АУП" {
		t.Errorf("filters = %+v", fl)
	}

	f.result(`[{"date":"2026-01-01","type":2},{"date":"2026-01-03","type":1}]`)
	days := must[[]myitmo.VacationCalendarDay](t)(c.Vacation.Calendar(ctx, 2026))
	if len(days) != 2 || days[0].Type != myitmo.VacationDayHoliday || days[0].Date != myitmo.NewDate(2026, time.January, 1) {
		t.Errorf("calendar = %+v", days)
	}

	f.result(`[{"employee_id":11,"remaining":12.5,"normative":28}]`)
	bal := must[[]myitmo.VacationBalance](t)(c.Vacation.Balance(ctx, myitmo.Today()))
	if bal[0].Remaining != 12.5 || bal[0].Normative != 28 {
		t.Errorf("balance = %+v", bal)
	}

	f.result(`[{"id":42,"vacation_id":42,"vacation_type":"Оплачиваемый отпуск","date_start":"2026-07-01","date_end":"2026-07-14",
		"duration":14,"status_id":1,"task_id":"900","signature_id":null,"needs_signature":true,"org_representer":"Петров П. П.",
		"employee":{"employee_id":11,"position_name":"Инженер","department_name":"Отдел","project_name":"","project_id":null}}]`)
	up := must[[]myitmo.Vacation](t)(c.Vacation.Upcoming(ctx))
	if v := up[0]; v.StatusID != myitmo.VacationStatusPending || v.TaskID != 900 || v.SignatureID != 0 || !v.NeedsSignature || v.Employee.EmployeeID != 11 || v.DateEnd.Day != 14 {
		t.Errorf("upcoming = %+v", up)
	}

	f.result(`{"conflict":"Пересечение с отпуском","num_of_days":14}`)
	cf := must[*myitmo.VacationConflict](t)(c.Vacation.Conflicts(ctx, myitmo.Today(), myitmo.Today(), 11))
	if !cf.HasConflict() || cf.NumOfDays != 14 {
		t.Errorf("conflict = %+v", cf)
	}
	f.result(`{"conflict":false,"num_of_days":3}`)
	cf = must[*myitmo.VacationConflict](t)(c.Vacation.Conflicts(ctx, myitmo.Today(), myitmo.Today(), 11))
	if cf.HasConflict() {
		t.Errorf("conflict = %+v", cf)
	}

	f.result(`{"count":1,"departments":[{"department_id":2,"department_name":"Отдел","people":[{"name":"Сидоров С. С.","isu":100001,"is_upper":true,
		"positions":[{"employee_id":11,"position_name":"Инженер","current_remaining_days":20.33,"project_code":"","project_name":"","category":"АУП",
		"vacations":[{"vacation_id":42,"employee":{"employee_id":11},"date_start":"2026-07-01","date_end":"2026-07-14","duration":14,
		"remaining_days_on_vac_start":28,"vacation_type":"Оплачиваемый отпуск","status_id":3,"status_1c":true,"task_id":900,"signature_id":901,
		"needs_signature":false,"read_only":true,"approved_by":"Петров П. П.","approved_at":"2026-06-01T10:00:00+03:00"}]}]}]}]}`)
	ov := must[*myitmo.VacationOverview](t)(c.Vacation.List(ctx, myitmo.VacationListParams{Limit: 10}))
	vac := ov.Departments[0].People[0].Positions[0].Vacations[0]
	if ov.Count != 1 || !ov.Departments[0].People[0].IsUpper || vac.Employee.EmployeeID != 11 || *vac.RemainingDaysOnVacStart != 28 ||
		!vac.Status1C || vac.SignatureID != 901 || vac.ApprovedAt == nil || vac.ApprovedAt.Hour() != 10 {
		t.Errorf("overview = %+v", ov)
	}

	f.result(`[{"department_id":2,"department_name":"Отдел","employees":[{"employee_id":11,"pers_name":"Сидоров С. С.","pers_id":100001,
		"position_name":"Инженер","category":"АУП","project_code":null,"project_name":null,"remaining":7}]}]`)
	emp := must[[]myitmo.VacationEmployeeGroup](t)(c.Vacation.Employees(ctx, myitmo.VacationEmployeesParams{}))
	if e := emp[0].Employees[0]; e.PersID != 100001 || e.Remaining != 7 {
		t.Errorf("employees = %+v", emp)
	}

	f.result(`{"count":1,"tasks":[{"vacation_id":42,"emp_id":11,"emp_name":"Сидоров С. С.","emp_isu":100001,"date_start":"2026-03-02","date_end":"2026-03-03",
		"duration":2,"vacation_type":"Отпуск за свой счет","position_name":"Инженер","project_id":"P-1","department_name":"Отдел","status":1,
		"approval_only":true,"refusable":false,"signature_id":"501","sign_task_id":"601",
		"files":[{"file_link":"https://example.test/f/1","file_name":"note.pdf"}]}]}`)
	tasks := must[*myitmo.VacationAgreementTasks](t)(c.Vacation.PendingApproval(ctx, myitmo.VacationAgreementParams{}))
	if tk := tasks.Tasks[0]; tasks.Count != 1 || tk.EmpISU != 100001 || tk.Refusable == nil || *tk.Refusable || tk.SignTaskID != 601 ||
		len(tk.Files) != 1 || tk.Files[0].FileName != "note.pdf" || tk.Status != myitmo.VacationStatusPending {
		t.Errorf("tasks = %+v", tasks)
	}

	f.result(`{"count":5}`)
	if n := must[int](t)(c.Vacation.PendingApprovalCount(ctx)); n != 5 {
		t.Errorf("count = %d", n)
	}
}
