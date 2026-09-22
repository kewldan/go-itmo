package bars_test

import (
	"net/http"
	"testing"

	"github.com/kewldan/go-itmo/bars"
)

func TestDisciplineSearches(t *testing.T) {
	f, c := session(t)
	f.reply(`[{"id":90,"name":"Предмет","terms":[2],"checkpoint_plan_ids":[8],"departments":[{"id":1}]}]`)
	list := must[[]bars.Discipline](t)(c.SearchDisciplines(ctx, bars.DisciplineQuery{
		Name: "Пр", Type: bars.FlowTypeFlow, Identifier: "a/b", WithCheckpointPlansOnly: true, DistinctByCheckpointPlans: true, Size: 20,
	}))
	f.expect(http.MethodGet, "journal/disciplines?distinctByCheckpointPlans=true&identifier=a%2Fb&name=%D0%9F%D1%80&size=20&type=flow&withCheckpointPlansOnly=true")
	if list[0].CheckpointPlanIDs[0] != 8 || string(list[0].Departments) != `[{"id":1}]` {
		t.Errorf("disciplines = %+v", list)
	}

	f.reply(`[]`)
	must[[]bars.Discipline](t)(c.SearchDisciplines(ctx, bars.DisciplineQuery{}))
	f.expect(http.MethodGet, "journal/disciplines")

	f.reply(`[{"id":1}]`)
	must[bars.RawJSON](t)(c.LegacyDisciplines(ctx, bars.DisciplineQuery{Name: "x"}))
	f.expect(http.MethodGet, "disciplines?name=x")

	f.reply(`[{"id":90,"name":"Предмет","terms":[2],"checkpoint_plan_ids":[]}]`)
	must[[]bars.Discipline](t)(c.RealizerDisciplines(ctx, "teacher1", bars.DisciplineFilter{Name: "Пр", ExcludedIDs: []int64{1, 2}}))
	f.expect(http.MethodGet, "disciplines/realizer?excludedIds=1%2C2&name=%D0%9F%D1%80&realizerPeopleId=teacher1")

	f.reply(`[]`)
	must[[]bars.Discipline](t)(c.AllDisciplines(ctx, bars.DisciplineFilter{Name: "x", Type: "group", Identifier: "G1"}))
	f.expect(http.MethodGet, "disciplines/all?filter=x&identifier=G1&type=group")

	f.reply(`[]`)
	must[[]bars.Discipline](t)(c.AllDisciplines(ctx, bars.DisciplineFilter{}))
	f.expect(http.MethodGet, "disciplines/all")

	f.reply(`{"id":90}`)
	must[bars.RawJSON](t)(c.DisciplineByID(ctx, 90))
	f.expect(http.MethodGet, "disciplines/90")

	f.reply(`[{"id":90,"name":"Предмет","term":3,"course_project":true}]`)
	options := must[[]bars.PlanDiscipline](t)(c.PlanDisciplines(ctx, 0))
	f.expect(http.MethodGet, "disciplines_for_plan/")
	if *options[0].Term != 3 || !options[0].CourseProject {
		t.Errorf("options = %+v", options)
	}
	f.reply(`[]`)
	must[[]bars.PlanDiscipline](t)(c.PlanDisciplines(ctx, 44))
	f.expect(http.MethodGet, "disciplines_for_plan/?flowId=44")
}

func TestShortCatalogueForms(t *testing.T) {
	f, c := session(t)
	f.reply(`[{"name":"current_year","value":"2026/2027"}]`)
	cfg := must[[]bars.Setting](t)(c.Config(ctx))
	f.expect(http.MethodGet, "config/")
	if cfg[0].Name != bars.SettingCurrentYear {
		t.Errorf("config = %+v", cfg)
	}

	f.reply(`[]`)
	must[[]bars.Discipline](t)(c.Disciplines(ctx, false))
	f.expect(http.MethodGet, "journal/disciplines?withCheckpointPlansOnly=false")

	f.reply(`[]`)
	must[[]bars.GroupOrFlow](t)(c.GroupsAndFlows(ctx, 90))
	f.expect(http.MethodGet, "journal/groups-and-flows?disciplineId=90")
}

func TestDisciplineAccessAndGroups(t *testing.T) {
	f, c := session(t)
	f.reply(`{"open":true}`)
	must[bars.RawJSON](t)(c.DisciplineAccess(ctx, 90))
	f.expect(http.MethodGet, "discipline/90/access")

	f.reply("")
	must[bars.RawJSON](t)(c.SetDisciplineAccess(ctx, 90, bars.RawJSON(`{"open":false}`)))
	f.expect(http.MethodPost, "discipline/90/access").sameJSON(t, `{"open":false}`)

	f.reply(`[{"type":"flow","name":"Поток","identifier":"a/b","checkpoint_plan_ids":[8],"term":1,"year":"2026/2027"}]`)
	groups := must[[]bars.GroupOrFlow](t)(c.SearchGroupsAndFlows(ctx, bars.GroupsAndFlowsQuery{DisciplineID: 90, CheckpointPlanID: 8, Name: "По"}))
	f.expect(http.MethodGet, "journal/groups-and-flows?checkpointPlanId=8&disciplineId=90&name=%D0%9F%D0%BE")
	if *groups[0].Term != 1 || groups[0].Year != "2026/2027" {
		t.Errorf("groups = %+v", groups)
	}

	f.reply(`[]`)
	must[bars.RawJSON](t)(c.Groups(ctx))
	f.expect(http.MethodGet, "groups")

	f.reply(`[]`)
	must[bars.RawJSON](t)(c.DisciplineGroups(ctx, 90))
	f.expect(http.MethodGet, "disciplines/90/group")

	f.reply(`["G1","Поток"]`)
	names := must[bars.RawJSON](t)(c.PlanGroupAndFlowNames(ctx, 8))
	f.expect(http.MethodGet, "checkpoint_plans/8/group_and_flow_names")
	if len(names) == 0 {
		t.Error("empty names")
	}
}

func TestFlows(t *testing.T) {
	f, c := session(t)
	f.reply(`[{"flow":{"id":44,"name":"Поток"},"checkpointPlan":{"id":8,"discipline":{"label":"Предмет","value":90},"terms":[1,2]}}]`)
	links := must[[]bars.FlowPlanLink](t)(c.FlowPlanLinks(ctx, "По"))
	f.expect(http.MethodGet, "flows/plans?name=%D0%9F%D0%BE")
	if links[0].Flow.ID != 44 || links[0].CheckpointPlan.Discipline.Label != "Предмет" || string(links[0].CheckpointPlan.Discipline.Value) != "90" {
		t.Errorf("links = %+v", links)
	}

	f.reply(`[{"id":44,"name":"Поток"}]`)
	flows := must[[]bars.Flow](t)(c.Flows(ctx, "x", true))
	f.expect(http.MethodGet, "flows?all=true&name=x")
	if flows[0].Name != "Поток" {
		t.Errorf("flows = %+v", flows)
	}
	f.reply(`[]`)
	must[[]bars.Flow](t)(c.Flows(ctx, "", false))
	f.expect(http.MethodGet, "flows")

	f.reply("")
	ok(t, c.LinkFlowPlan(ctx, 44, 8))
	f.expect(http.MethodPost, "flows/44/plans/8").noBody(t)

	f.reply("")
	ok(t, c.UnlinkFlowPlan(ctx, 44, 8))
	f.expect(http.MethodDelete, "flows/44/plans/8")
}
