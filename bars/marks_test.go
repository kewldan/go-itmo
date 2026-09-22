package bars_test

import (
	"net/http"
	"testing"

	"github.com/kewldan/go-itmo/bars"
)

const teacherJournal = `{"students":[
{"student_id":1,"student_login":"100000","student_name":"Студент Один","wants_to_increase_marks":false,
 "marks":{"regular":[],"final":null,"additional":null,"course":{"id":3,"checkpoint_id":70,"checkpoint_plan_id":8,"mark":15,"type":"course","is_absent":false},
 "regularSum":0,"total":15,"active_approvals":[{"id":2,"attempt":1,"course":true,"mark_string":"Хор., C","updated_by_name":"Преподаватель"}]},
 "accessibility":{"can_edit_current_marks":true,"can_approve_marks":true,"can_approve_course_project":true,"can_approve_retry_course_project":false}},
{"student_id":2,"student_login":"100001","student_name":"Студент Два","marks":{"regular":[]}}],
"headers":{"plan":{"id":8,"year":"2026/2027","terms":[1],"discipline":{"id":90,"name":"Предмет","course_project":true,"term":null},
"regular_checkpoints":[{"id":6,"name":"Группа","type":"Тест","type_id":21,"group":true,"max_sub_checkpoints_fillable":2,"test_id":5,"test_name":"ЦДО",
"sub_checkpoints":[{"id":61,"name":"Тест 1","week":3,"parent_checkpoint_id":6}]}],
"final_checkpoint":null,"course_project_checkpoint":{"id":70,"min_grade":0,"max_grade":100},"alternate_methods":null,
"programs":[{"id":1,"code":"09.03.01","name":"Информатика"}],"components":[],"status":"SENT",
"created_at":1767225600000,"updated_at":"2026-01-02T10:00:00+03:00"},
"type":"group","identifier":"G 1/2","name":"G1"}}`

func TestTeacherJournals(t *testing.T) {
	f, c := session(t)
	f.reply(teacherJournal)
	j := must[*bars.TeacherJournal](t)(c.TeacherJournal(ctx, 8, bars.FlowTypeGroup, "G 1/2"))
	f.expect(http.MethodGet, "marks/8/group/G%201%2F2").noBody(t)
	if len(j.Students) != 2 {
		t.Fatalf("students = %d", len(j.Students))
	}
	s := j.Students[0]
	if s.Marks.Course == nil || *s.Marks.Course.Mark != 15 || !s.Accessibility.CanApproveCourseProject {
		t.Errorf("student = %+v", s)
	}
	if s.Marks.ActiveApprovals[0].UpdatedByName != "Преподаватель" {
		t.Errorf("approval = %+v", s.Marks.ActiveApprovals[0])
	}
	p := j.Headers.Plan
	if p.CourseProjectCheckpoint.MaxGrade != 100 || p.AlternateMethods != nil || p.Programs[0].Code != "09.03.01" {
		t.Errorf("plan = %+v", p)
	}
	if p.CreatedAt.UnixMilli() != 1767225600000 || p.UpdatedAt.IsZero() || p.Discipline.Term != nil {
		t.Errorf("plan times = %v %v", p.CreatedAt, p.UpdatedAt)
	}
	cp := p.RegularCheckpoints[0]
	if *cp.MaxSubCheckpointsFillable != 2 || *cp.TestID != 5 || cp.TestName != "ЦДО" || *cp.SubCheckpoints[0].ParentCheckpointID != 6 {
		t.Errorf("checkpoint = %+v", cp)
	}

	f.reply(teacherJournal)
	must[*bars.TeacherJournal](t)(c.TeacherJournalAuto(ctx, 90, bars.FlowTypeFlow, "a/b"))
	f.expect(http.MethodGet, "marks/90/flow/a%2Fb/auto")

	f.reply(`{"student_id":2,"student_login":"100001","student_name":"","marks":{"regular":[],"regularSum":5,"total":5}}`)
	row := must[*bars.StudentRecord](t)(c.JournalStudent(ctx, 8, "flow", "a/b", 2))
	f.expect(http.MethodGet, "marks/8/flow/a%2Fb/student/2")
	if row.StudentID != 2 || *row.Marks.Total != 5 {
		t.Errorf("row = %+v", row)
	}
}

func TestMarkHistory(t *testing.T) {
	f, c := session(t)
	f.reply(`[{"id":1,"set_at":1767225600000,"set_by_name":"Преподаватель","mark":7.5,"checkpoint_id":6}]`)
	h := must[[]bars.MarkHistoryEntry](t)(c.MarkHistory(ctx, 8, "flow", "a/b", 2, 6))
	f.expect(http.MethodGet, "marks/8/flow/a%2Fb/student/2/history?checkpointId=6")
	if h[0].SetAt.UnixMilli() != 1767225600000 || *h[0].Mark != 7.5 || h[0].MarksSum != nil {
		t.Errorf("history = %+v", h)
	}
	f.reply(`[]`)
	must[[]bars.MarkHistoryEntry](t)(c.MarkHistory(ctx, 8, "flow", "a/b", 2, 0))
	f.expect(http.MethodGet, "marks/8/flow/a%2Fb/student/2/history")

	f.reply(`[{"id":4,"set_at":1767225600000,"set_by_name":"П","marks_sum":61,"mark_string":"Удвл., E"}]`)
	h = must[[]bars.MarkHistoryEntry](t)(c.ApprovalHistory(ctx, "flow", "a/b", 4))
	f.expect(http.MethodGet, "marks/flow/a%2Fb/approval/4/history")
	if *h[0].MarksSum != 61 || h[0].MarkString != "Удвл., E" {
		t.Errorf("approval history = %+v", h)
	}
}

func TestSetMarks(t *testing.T) {
	f, c := session(t)
	mark := 7.5
	f.reply("")
	ok(t, c.SetMark(ctx, "flow", "a/b", 2, bars.MarkUpdate{Mark: &mark, CheckpointID: 6, Type: bars.MarkTypeCurrent}))
	f.expect(http.MethodPost, "marks/flow/a%2Fb/current/2").sameJSON(t, `{"mark":7.5,"checkpoint_id":6,"type":"current"}`)

	f.reply("")
	ok(t, c.SetMark(ctx, "flow", "a/b", 2, bars.MarkUpdate{CheckpointID: 8, CheckpointPlanID: 8, Type: bars.MarkTypeAdditional}))
	f.expect(http.MethodPost, "marks/flow/a%2Fb/additional/2").sameJSON(t, `{"mark":null,"checkpoint_id":8,"checkpoint_plan_id":8,"type":"additional"}`)

	zero, absent := 0.0, true
	f.reply("")
	ok(t, c.SetMark(ctx, "group", "G1", 2, bars.MarkUpdate{Mark: &zero, CheckpointID: 9, Type: bars.MarkTypeFinal, IsAbsent: &absent}))
	f.expect(http.MethodPost, "marks/group/G1/final/2").sameJSON(t, `{"mark":0,"checkpoint_id":9,"type":"final","is_absent":true}`)

	if err := c.SetMark(ctx, "group", "G1", 2, bars.MarkUpdate{}); err == nil {
		t.Error("accepted a mark without type")
	}

	f.reply("")
	ok(t, c.SetAdditionalMarkLegacy(ctx, 30, 2, bars.MarkUpdate{Mark: &mark, Type: bars.MarkTypeAdditional}))
	f.expect(http.MethodPost, "marks/additional/30/2").sameJSON(t, `{"mark":7.5,"checkpoint_id":0,"type":"additional"}`)

	f.reply("")
	ok(t, c.SetFinalMarkLegacy(ctx, 31, 2, bars.MarkUpdate{Mark: &mark, CheckpointID: 9, Type: bars.MarkTypeFinal}))
	f.expect(http.MethodPost, "marks/final/31/2").sameJSON(t, `{"mark":7.5,"checkpoint_id":9,"type":"final"}`)

	f.reply(`{"any":1}`)
	v := must[bars.RawJSON](t)(c.DisciplineMarks(ctx, 90))
	f.expect(http.MethodGet, "marks/90/")
	if string(v) != `{"any":1}` {
		t.Errorf("marks = %s", v)
	}
}

func TestApprovals(t *testing.T) {
	f, c := session(t)
	f.reply("")
	ok(t, c.ApproveStudent(ctx, 8, "flow", "a/b", 2, bars.AttemptFirst, false))
	f.expect(http.MethodPost, "marks/8/flow/a%2Fb/2/approval/1").noBody(t)

	f.reply("")
	ok(t, c.ApproveStudent(ctx, 8, "flow", "a/b", 2, bars.AttemptRetry, true))
	f.expect(http.MethodPost, "marks/8/flow/a%2Fb/2/approval/2?course=true")

	f.reply(`[{"id":1,"student_id":2,"student_login":"100001","checkpoint_plan_id":8,"attempt":1,"marks_sum":75,"mark_string":"Хор., C","is_active":true}]`)
	approvals := must[[]bars.Approval](t)(c.ApproveStudents(ctx, 8, "flow", "a/b", 1, []int64{2, 3}, true))
	f.expect(http.MethodPost, "marks/8/flow/a%2Fb/approval/1?course=true&studentIds=2%2C3")
	if approvals[0].StudentID != 2 || approvals[0].GradeCode() != "4/C" {
		t.Errorf("approvals = %+v", approvals)
	}

	f.reply(`[]`)
	must[[]bars.Approval](t)(c.ApproveStudents(ctx, 8, "flow", "a/b", 3, []int64{}, false))
	f.expect(http.MethodPost, "marks/8/flow/a%2Fb/approval/3?studentIds=")

	f.reply(`[]`)
	must[[]bars.Approval](t)(c.ApproveStudents(ctx, 8, "flow", "a/b", 3, nil, false))
	f.expect(http.MethodPost, "marks/8/flow/a%2Fb/approval/3")

	f.reply("")
	ok(t, c.AcceptJournalAgreement(ctx, 8))
	f.expect(http.MethodPost, "marks/8/student/agreement").noBody(t)
}

func TestJournalAdmin(t *testing.T) {
	f, c := session(t)
	f.reply(`[{"id":8,"name":"Предмет","terms":[1],"programs":[{"code":"09.03.01","name":"Информатика"}]}]`)
	plans := must[[]bars.JournalPlan](t)(c.JournalPlans(ctx, "Пред мет"))
	f.expect(http.MethodGet, "journal/checkpoint-plans?name=%D0%9F%D1%80%D0%B5%D0%B4+%D0%BC%D0%B5%D1%82")
	if plans[0].ID != 8 || plans[0].Programs[0].ID != 0 {
		t.Errorf("plans = %+v", plans)
	}
	f.reply(`[]`)
	must[[]bars.JournalPlan](t)(c.JournalPlans(ctx, ""))
	f.expect(http.MethodGet, "journal/checkpoint-plans")

	f.reply("")
	ok(t, c.ClearJournalCache(ctx, "2026/2027", bars.Spring))
	f.expect(http.MethodDelete, "journal/cache?term=0&year=2026%2F2027")
}
