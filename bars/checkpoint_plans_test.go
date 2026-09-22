package bars_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/kewldan/go-itmo/bars"
)

func TestCheckpointPlanReads(t *testing.T) {
	f, c := session(t)
	f.reply(`{"id":8,"gid":"p_1","year":"2026/2027","terms":[1],"term":1,"discipline":{"id":90,"name":"Предмет","course_project":false},
"regular_checkpoints":[],"final_checkpoint":{"id":9,"name":null,"type":"Экзамен","type_id":1,"min_grade":12,"max_grade":20,"sub_checkpoints":[]},
"point_distribution":80,"additional_points":true,"alternate_methods":true,"has_course_project":false,"status":"SAVED"}`)
	p := must[*bars.CheckpointPlan](t)(c.CheckpointPlan(ctx, 8))
	f.expect(http.MethodGet, "checkpoint_plans/8")
	if p.PointDistribution != 80 || *p.Term != 1 || !*p.AlternateMethods || string(p.Status) != `"SAVED"` {
		t.Errorf("plan = %+v", p)
	}

	f.reply(`[{"id":8,"discipline":{"id":90,"name":"Предмет"},"terms":[1,2]}]`)
	plans := must[[]bars.CheckpointPlan](t)(c.DisciplineCheckpointPlans(ctx, 90, []int64{1, 2}))
	f.expect(http.MethodGet, "checkpoint_plans/discipline/90?programIds=1%2C2")
	if plans[0].Discipline.Name != "Предмет" {
		t.Errorf("plans = %+v", plans)
	}
	f.reply(`[]`)
	must[[]bars.CheckpointPlan](t)(c.DisciplineCheckpointPlans(ctx, 90, []int64{}))
	f.expect(http.MethodGet, "checkpoint_plans/discipline/90?programIds=")
	f.reply(`[]`)
	must[[]bars.CheckpointPlan](t)(c.DisciplineCheckpointPlans(ctx, 90, nil))
	f.expect(http.MethodGet, "checkpoint_plans/discipline/90")

	f.reply(`{"content":[{"id":8,"discipline":{"label":"Предмет","value":"90"},"programs":[{"code":"09.03.01","name":"Информатика"}],
"terms":[1],"created_at":1767225600000,"updated_at":1767312000000}],"totalElements":11,"totalPages":2,"size":10,"number":1,
"numberOfElements":1,"first":false,"last":true,"empty":false,"pageable":{"pageNumber":1,"pageSize":10,"offset":10,"paged":true,"unpaged":false,"sort":{}}}`)
	page := must[*bars.Page[bars.CheckpointPlanSummary]](t)(c.CheckpointPlans(ctx, bars.CheckpointPlanQuery{
		Name: "Пр", Page: 1, Size: 10, Sort: "updatedAt,desc",
		StartDate: time.UnixMilli(1767225600000), EndDate: time.UnixMilli(1767312000000),
	}))
	f.expect(http.MethodGet, "checkpoint_plans?end_date=1767312000000&name=%D0%9F%D1%80&page=1&size=10&sort=updatedAt%2Cdesc&start_date=1767225600000")
	if page.TotalElements != 11 || page.Pageable.PageNumber != 1 || page.Content[0].UpdatedAt.UnixMilli() != 1767312000000 ||
		string(page.Content[0].Discipline.Value) != `"90"` {
		t.Errorf("page = %+v", page)
	}
	f.reply(`{"content":[],"totalElements":0}`)
	must[*bars.Page[bars.CheckpointPlanSummary]](t)(c.CheckpointPlans(ctx, bars.CheckpointPlanQuery{}))
	f.expect(http.MethodGet, "checkpoint_plans?page=0")
}

func TestCheckpointPlanWrites(t *testing.T) {
	f, c := session(t)
	term := 1
	week := 3
	plan := bars.CheckpointPlanWrite{
		GID:        "p_1",
		Discipline: bars.PlanDiscipline{ID: 90, Name: "Предмет", Term: &term},
		Terms:      []int{1},
		Programs:   []bars.EducationalProgram{{ID: 1, Code: "09.03.01", Name: "Информатика"}},
		RegularCheckpoints: []bars.Checkpoint{{GID: "chk_1", Name: "Работа", Type: "Тест", TypeID: 2, Week: &week,
			MinGrade: 1, MaxGrade: 10, SubCheckpoints: []bars.Checkpoint{}}},
		FinalCheckpoint:   &bars.Checkpoint{GID: "f_1", Type: "Экзамен", TypeID: 1, MinGrade: 12, MaxGrade: 20},
		PointDistribution: 80,
		Status:            bars.PlanStatusSaved,
		UpdatedAt:         bars.MillisOf(time.UnixMilli(1767225600000)),
	}
	f.reply(`{"id":8,"gid":"p_1"}`)
	created := must[*bars.CheckpointPlan](t)(c.CreateCheckpointPlan(ctx, plan))
	f.expect(http.MethodPost, "checkpoint_plans").sameJSON(t, `{"gid":"p_1",
"discipline":{"id":90,"name":"Предмет","course_project":false,"term":1},"terms":[1],
"programs":[{"id":1,"code":"09.03.01","name":"Информатика"}],
"regular_checkpoints":[{"gid":"chk_1","name":"Работа","type":"Тест","type_id":2,"week":3,"group":false,"key":false,
"min_grade":1,"max_grade":10,"sub_checkpoints":[],"parent_checkpoint_id":null}],
"final_checkpoint":{"gid":"f_1","name":"","type":"Экзамен","type_id":1,"week":null,"group":false,"key":false,
"min_grade":12,"max_grade":20,"sub_checkpoints":[],"parent_checkpoint_id":null},
"point_distribution":80,"additional_points":false,"alternate_methods":false,"has_course_project":false,
"course_project_checkpoint":null,"status":"SAVED","updated_at":1767225600000}`)
	if created.ID != 8 {
		t.Errorf("created = %+v", created)
	}

	plan.ID = 8
	f.queue(step{status: http.StatusNoContent})
	ok(t, c.UpdateCheckpointPlan(ctx, plan))
	r := f.expect(http.MethodPut, "checkpoint_plans/8")
	var body struct {
		ID  int64  `json:"id"`
		GID string `json:"gid"`
	}
	mustJSON(t, r.body, &body)
	if body.ID != 8 || body.GID != "p_1" {
		t.Errorf("update body = %s", r.body)
	}
	if err := c.UpdateCheckpointPlan(ctx, bars.CheckpointPlanWrite{}); err == nil {
		t.Error("accepted a plan without ID")
	}

	cpw := bars.CourseProjectPlanWrite{CourseProject: true, CourseProjectCheckPoint: bars.CourseProjectRange{MinGrade: 60, MaxGrade: 100},
		Status: bars.PlanStatusSent, Discipline: bars.PlanDiscipline{ID: 91, Name: "Курсовая", CourseProject: true}}
	f.reply(`{"id":10}`)
	must[*bars.CheckpointPlan](t)(c.CreateCourseProjectPlan(ctx, cpw))
	f.expect(http.MethodPost, "checkpoint_plans").sameJSON(t, `{"course_project":true,"course_project_checkPoint":{"min_grade":60,"max_grade":100},
"status":"SENT","discipline":{"id":91,"name":"Курсовая","course_project":true}}`)

	cpw.ID = 10
	f.queue(step{status: http.StatusNoContent})
	ok(t, c.UpdateCourseProjectPlan(ctx, cpw))
	f.expect(http.MethodPut, "checkpoint_plans/10").sameJSON(t, `{"id":10,"course_project":true,"course_project_checkPoint":{"min_grade":60,"max_grade":100},
"status":"SENT","discipline":{"id":91,"name":"Курсовая","course_project":true}}`)
	if err := c.UpdateCourseProjectPlan(ctx, bars.CourseProjectPlanWrite{}); err == nil {
		t.Error("accepted a plan without ID")
	}

	f.reply("")
	ok(t, c.DeleteCheckpointPlan(ctx, 8))
	f.expect(http.MethodDelete, "checkpoint_plans/8")

	f.reply(`[90,91]`)
	imported := must[[]int64](t)(c.ImportRPD(ctx, "123", "456"))
	f.expect(http.MethodGet, "checkpoint_plans/rpd/constructor/import?ids=123%2C456")
	if len(imported) != 2 || imported[1] != 91 {
		t.Errorf("imported = %v", imported)
	}
	if _, err := c.ImportRPD(ctx); err == nil {
		t.Error("accepted no RPD IDs")
	}
}

func TestPersonalPlans(t *testing.T) {
	f, c := session(t)
	f.reply(`{"id":3}`)
	must[bars.RawJSON](t)(c.PersonalPlan(ctx, 3))
	f.expect(http.MethodGet, "personal_plans/3")

	f.reply(`{"id":3}`)
	must[bars.RawJSON](t)(c.GroupPersonalPlan(ctx, 90, "G 1/2"))
	f.expect(http.MethodGet, "personal_plans/90/G%201%2F2")

	f.reply(`{"id":4}`)
	v := must[bars.RawJSON](t)(c.SavePersonalPlan(ctx, bars.RawJSON(`{"discipline_id":90}`)))
	f.expect(http.MethodPost, "personal_plans").sameJSON(t, `{"discipline_id":90}`)
	if string(v) != `{"id":4}` {
		t.Errorf("saved = %s", v)
	}
}
