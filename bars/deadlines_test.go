package bars_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/kewldan/go-itmo/bars"
)

func TestDeadlines(t *testing.T) {
	f, c := session(t)
	f.reply(`[{"discipline_name":"Предмет","plan_identifier":"a/b","checkpoint":{"id":6,"name":"Работа","type":"Тест"},"deadline_date":1767225600000}]`)
	all := must[[]bars.UserDeadline](t)(c.Deadlines(ctx))
	f.expect(http.MethodGet, "deadline")
	if all[0].Checkpoint.Type != "Тест" || all[0].DeadlineDate.UnixMilli() != 1767225600000 {
		t.Errorf("deadlines = %+v", all)
	}

	f.reply(`[{"id":5,"checkpoint_id":61,"parent_checkpoint_id":6,"deadline_date":1767225600000,"created_by":{"id":1},
"updated_by":null,"created_at":1767000000000,"updated_at":null,"flow_identifier":"a/b"},
{"checkpoint_id":62,"deadline_date":null}]`)
	list := must[[]bars.Deadline](t)(c.PlanDeadlines(ctx, "flow", "a/b", 8))
	f.expect(http.MethodGet, "deadline/flow/a%2Fb/8")
	if *list[0].ParentCheckpointID != 6 || list[0].DeadlineDate.UnixMilli() != 1767225600000 || list[0].CreatedAt.IsZero() {
		t.Errorf("deadline = %+v", list[0])
	}
	if !list[1].DeadlineDate.IsZero() || list[1].ParentCheckpointID != nil {
		t.Errorf("empty deadline = %+v", list[1])
	}

	list[1].DeadlineDate = bars.MillisOf(time.UnixMilli(1767312000000))
	f.reply("")
	ok(t, c.SaveDeadlines(ctx, "flow", "a/b", list))
	f.expect(http.MethodPost, "deadline/flow/a%2Fb").sameJSON(t, `[
{"id":5,"checkpoint_id":61,"parent_checkpoint_id":6,"deadline_date":1767225600000,"flow_identifier":"a/b"},
{"checkpoint_id":62,"deadline_date":1767312000000}]`)
	if list[0].CreatedBy == nil {
		t.Error("SaveDeadlines modified the caller's slice")
	}

	parent := int64(6)
	cp := bars.Checkpoint{ID: 61, Name: "Тест 1", Type: "Тест", ParentCheckpointID: &parent, SubCheckpoints: []bars.Checkpoint{}}
	f.reply("")
	ok(t, c.SetDeadline(ctx, "flow", "a/b", cp, time.UnixMilli(1767225600000)))
	f.expect(http.MethodPost, "deadline/single/flow/a%2Fb").sameJSON(t, `{"checkpoint":{"id":61,"gid":"","name":"Тест 1","type":"Тест",
"type_id":0,"week":null,"group":false,"key":false,"min_grade":0,"max_grade":0,"sub_checkpoints":[],"parent_checkpoint_id":6},
"checkpoint_id":61,"parent_checkpoint_id":6,"deadline_date":1767225600000}`)

	cp.ParentCheckpointID = nil
	f.reply("")
	ok(t, c.SetDeadline(ctx, "flow", "a/b", cp, time.Time{}))
	r := f.expect(http.MethodPost, "deadline/single/flow/a%2Fb")
	var body struct {
		CheckpointID       int64    `json:"checkpoint_id"`
		ParentCheckpointID *int64   `json:"parent_checkpoint_id"`
		DeadlineDate       *float64 `json:"deadline_date"`
	}
	mustJSON(t, r.body, &body)
	if body.CheckpointID != 61 || body.ParentCheckpointID != nil || body.DeadlineDate != nil {
		t.Errorf("clear body = %s", r.body)
	}
}

func TestSessions(t *testing.T) {
	f, c := session(t)
	f.reply(`[{"student":{"lastName":"Фамилия","firstName":"Имя","middleName":"","groupName":"G1","id":77,"isu":"100000"},
"schedule":[{"attempt":1,"from":1767225600000,"to":1767312000000},{"attempt":2,"from":null,"to":null}]}]`)
	sessions := must[[]bars.StudentSession](t)(c.StudentSessions(ctx, []int64{8, 9}, "flow", "a/b", true))
	f.expect(http.MethodGet, "session/8,9/flow/a%2Fb?course=true")
	s := sessions[0]
	if s.Student.GroupName != "G1" || s.Schedule[0].To.UnixMilli() != 1767312000000 || !s.Schedule[1].From.IsZero() {
		t.Errorf("session = %+v", s)
	}

	f.reply(`[]`)
	must[[]bars.StudentSession](t)(c.StudentSessions(ctx, []int64{8}, "group", "G1", false))
	f.expect(http.MethodGet, "session/8/group/G1")
	if _, err := c.StudentSessions(ctx, nil, "group", "G1", false); err == nil {
		t.Error("accepted no plan IDs")
	}

	s.Schedule[1].From = bars.MillisOf(time.UnixMilli(1768000000000))
	f.reply("")
	ok(t, c.SetStudentSession(ctx, []int64{8}, s.Student, s.Schedule, false))
	f.expect(http.MethodPost, "session/8").sameJSON(t, `{"course":null,
"student":{"lastName":"Фамилия","firstName":"Имя","middleName":"","groupName":"G1","id":77,"isu":"100000"},
"schedule":[{"attempt":1,"from":1767225600000,"to":1767312000000},{"attempt":2,"from":1768000000000,"to":null}]}`)

	f.reply("")
	ok(t, c.SetGroupSession(ctx, []int64{8, 9}, "flow", "a/b", s.Schedule[:1], true))
	f.expect(http.MethodPost, "session/8,9/flow/a%2Fb").sameJSON(t, `{"course":true,
"schedule":[{"attempt":1,"from":1767225600000,"to":1767312000000}]}`)
	if err := c.SetGroupSession(ctx, nil, "flow", "a/b", nil, false); err == nil {
		t.Error("accepted no plan IDs")
	}
	if err := c.SetStudentSession(ctx, nil, s.Student, nil, false); err == nil {
		t.Error("accepted no plan IDs")
	}
}
