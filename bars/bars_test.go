package bars_test

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/kewldan/go-itmo/bars"
)

var ctx = context.Background()

const user = `{"id":1,"login":"100000","first_name":"","middle_name":"","last_name":"",
"user_roles":[{"id":2,"name":"Обучающийся"}],"selected_role":{"id":2,"name":"Обучающийся","locked":false,"selected":true},
"selected_year":"2026/2027","selected_term":1,"personal_config":[],"can_change_user":false,"restricted_to_have_read_only_access":false}`

func TestLoginStoresHeaderAndRequestsCarryIt(t *testing.T) {
	f := newFakeBARS(t)
	c := f.client()
	f.queue(step{header: map[string]string{"Authorization": "Bearer synthetic-token"}}, step{body: user})

	if err := c.Login(ctx, "synthetic&code"); err != nil {
		t.Fatal(err)
	}
	u, err := c.CurrentUser(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if u.Login != "100000" || u.SelectedTerm != bars.Autumn || u.SelectedRole == nil || *u.SelectedRole.Selected != true {
		t.Errorf("user = %+v", u)
	}
	reqs := f.requests()
	if want := "/backend/rest/login?code=synthetic%26code&customRedirectUri=https%3A%2F%2Fbars.itmo.ru%2Frest%2Flogin"; reqs[0].path != want {
		t.Errorf("login path = %s", reqs[0].path)
	}
	if reqs[0].auth != "" || reqs[1].auth != "Bearer synthetic-token" {
		t.Errorf("auth headers = %q, %q", reqs[0].auth, reqs[1].auth)
	}
}

func TestUnauthorizedWithoutCodeSource(t *testing.T) {
	f := newFakeBARS(t)
	f.queue(step{status: http.StatusUnauthorized})
	_, err := f.client(withSession("Bearer synthetic-token")).CurrentUser(ctx)
	var e *bars.Error
	if !errors.As(err, &e) || !e.IsUnauthorized() {
		t.Fatalf("err = %v", err)
	}
	if _, err := f.client().CurrentUser(ctx); !errors.Is(err, bars.ErrNoSession) {
		t.Fatalf("without session: %v", err)
	}
}

func TestExpiredSessionIsRenewedOnce(t *testing.T) {
	f := newFakeBARS(t)
	calls := 0
	c := f.client(withSession("Bearer synthetic-expired"), bars.WithCodeSource(func(context.Context) (string, error) {
		calls++
		return "fresh-code", nil
	}))
	f.queue(
		step{status: http.StatusUnauthorized},
		step{header: map[string]string{"Authorization": "Bearer synthetic-fresh"}},
		step{body: user},
	)
	if _, err := c.CurrentUser(ctx); err != nil {
		t.Fatal(err)
	}
	reqs := f.requests()
	if calls != 1 || reqs[0].auth != "Bearer synthetic-expired" || !strings.Contains(reqs[1].path, "code=fresh-code") || reqs[2].auth != "Bearer synthetic-fresh" {
		t.Errorf("calls=%d requests=%+v", calls, reqs)
	}

	// A second 401 after renewal is final.
	f.queue(
		step{status: http.StatusUnauthorized},
		step{header: map[string]string{"Authorization": "Bearer synthetic-fresh-2"}},
		step{status: http.StatusUnauthorized},
	)
	var e *bars.Error
	if _, err := c.CurrentUser(ctx); !errors.As(err, &e) || !e.IsUnauthorized() {
		t.Fatalf("err = %v", err)
	}
}

func TestSelectPeriodWritesOnlyDifferences(t *testing.T) {
	f := newFakeBARS(t)
	c := f.client(withSession("Bearer synthetic-token"))

	f.queue(step{body: user})
	if _, err := c.SelectPeriod(ctx, "2026/2027", bars.Autumn); err != nil {
		t.Fatal(err)
	}
	if n := len(f.requests()); n != 1 {
		t.Fatalf("unchanged period made %d requests", n)
	}

	spring := strings.NewReplacer("2026/2027", "2025/2026", `"selected_term":1`, `"selected_term":0`).Replace(user)
	f.queue(step{body: user}, step{body: `{"id":9,"name":"current_year","value":"2025/2026"}`}, step{body: `{"id":9,"name":"current_term","value":"0"}`}, step{body: spring})
	if _, err := c.SelectPeriod(ctx, "2025/2026", bars.Spring); err != nil {
		t.Fatal(err)
	}
	reqs := f.requests()
	if reqs[2].method != http.MethodPost || reqs[2].path != "/backend/rest/config/personal" || reqs[2].body != `{"name":"current_year","value":"2025/2026"}` {
		t.Errorf("year write = %+v", reqs[2])
	}
	if reqs[3].body != `{"name":"current_term","value":"0"}` {
		t.Errorf("term write = %+v", reqs[3])
	}

	f.queue(step{body: user}, step{body: `{"name":"current_term","value":"0"}`}, step{body: user})
	if _, err := c.SelectPeriod(ctx, "2026/2027", bars.Spring); !errors.Is(err, bars.ErrPeriodNotApplied) {
		t.Fatalf("err = %v", err)
	}
	if _, err := c.SelectPeriod(ctx, "2026", bars.Spring); err == nil {
		t.Fatal("accepted malformed year")
	}
}

func TestStudentJournal(t *testing.T) {
	f := newFakeBARS(t)
	c := f.client(withSession("Bearer synthetic-token"))
	f.queue(step{body: `{"students":[{"student_id":1,"student_login":"100000","student_name":"",
"marks":{"regular":[{"id":5,"checkpoint_id":6,"checkpoint_plan_id":8,"mark":7.5,"type":"current","is_absent":false,"created_at":1767225600000}],
"additional":{"id":7,"checkpoint_id":null,"checkpoint_plan_id":8,"mark":2.0,"type":"current","is_absent":false},
"regularSum":7.5,"total":9.5,"active_approvals":[{"id":1,"attempt":2,"checkpoint_plan_id":8,"student_login":"100000",
"marks_sum":9.5,"mark_string":"Удвл., E","is_active":true,"is_invalid":false,"is_absent":false,"course":false}]}}],
"headers":{"plan":{"id":8,"year":"2025/2026","terms":[2],"discipline":{"id":90,"name":"Предмет"},
"regular_checkpoints":[{"id":6,"name":"Работа","type":"Тест","min_grade":1.0,"max_grade":10.0,"key":true,"sub_checkpoints":[]}],
"final_checkpoint":{"id":9,"name":null,"type":"Экзамен","min_grade":12.0,"max_grade":20.0,"key":false,"sub_checkpoints":[]},
"has_course_project":false},"type":"flow","identifier":"a/b","name":"Поток"}}`})

	j, err := c.StudentJournal(ctx, 8, "flow", "a/b")
	if err != nil {
		t.Fatal(err)
	}
	if p := f.requests()[0].path; p != "/backend/rest/marks/8/flow/a%2Fb/student" {
		t.Errorf("path = %s", p)
	}
	m := j.Students[0].Marks
	if m.Additional.CheckpointID != nil || !m.HasAnyMark() || j.Headers.Plan.FinalCheckpoint.Name != "" {
		t.Errorf("journal = %+v", j)
	}
	if m.Regular[0].CreatedAt.UnixMilli() != 1767225600000 {
		t.Errorf("created_at = %v", m.Regular[0].CreatedAt)
	}
	if got := m.ActiveApprovals[0].GradeCode(); got != "3/E" {
		t.Errorf("grade = %s", got)
	}
	var empty bars.StudentMarks
	if empty.HasAnyMark() {
		t.Error("empty marks report a mark")
	}
}

func TestGradeCode(t *testing.T) {
	cases := map[string]string{
		"Отл., A": "5/A", "Хор., C": "4/C", "Удвл., E": "3/E", "Неуд., FX": "2/FX",
		"Зачет": "Зачет", "Незачет": "Незачет", "Удвл.": "Удвл.", " Отл., B ": "5/B", "": "",
	}
	for in, want := range cases {
		if got := (&bars.Approval{MarkString: in}).GradeCode(); got != want {
			t.Errorf("GradeCode(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestRotatedHeaderReplacesSession(t *testing.T) {
	f := newFakeBARS(t)
	store := &bars.MemoryStore{}
	_ = store.Save(ctx, "Bearer synthetic-token")
	c := f.client(bars.WithStore(store))
	f.queue(step{body: user, header: map[string]string{"Authorization": "Bearer synthetic-rotated"}})
	if _, err := c.CurrentUser(ctx); err != nil {
		t.Fatal(err)
	}
	if v, _ := store.Load(ctx); v != "Bearer synthetic-rotated" {
		t.Errorf("session = %q", v)
	}
}
