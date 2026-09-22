package myitmo_test

import (
	"net/http"
	"testing"

	"github.com/kewldan/go-itmo/myitmo"
)

func TestStudyPlanRoutes(t *testing.T) {
	contents := []myitmo.StudyPlanSignContent{{DcID: 11, Semester: 5}, {DcID: 12, Semester: 6}}
	runElRoutes(t, []elRoute{
		{"Programs", func(c *myitmo.Client) error { return elDiscard(c.StudyPlan.Programs(ctx)) },
			http.MethodGet, "/api/eduPlanNew/programs", "", ""},
		{"Plan", func(c *myitmo.Client) error { return elDiscard(c.StudyPlan.Plan(ctx, 500, myitmo.Ptr[int64](7))) },
			http.MethodGet, "/api/eduPlanNew/study_plan/500", "spec_id=7", ""},
		{"PlanNoSpec", func(c *myitmo.Client) error { return elDiscard(c.StudyPlan.Plan(ctx, 500, nil)) },
			http.MethodGet, "/api/eduPlanNew/study_plan/500", "", ""},
		{"Choice", func(c *myitmo.Client) error { return elDiscard(c.StudyPlan.Choice(ctx, 500)) },
			http.MethodGet, "/api/eduPlan/choice/500", "", ""},
		{"SignDiscipline", func(c *myitmo.Client) error { return c.StudyPlan.SignDiscipline(ctx, 500, 3, 100, contents) },
			http.MethodPost, "/api/eduPlan/500/modules/3/disciplines/100/sign", "", `[{"dc_id":11,"semester":5},{"dc_id":12,"semester":6}]`},
		{"UnsignDiscipline", func(c *myitmo.Client) error { return c.StudyPlan.UnsignDiscipline(ctx, 500, 3, 100, contents[:1]) },
			http.MethodDelete, "/api/eduPlan/500/modules/3/disciplines/100/sign", "", `[{"dc_id":11,"semester":5}]`},
		{"Replaceable", func(c *myitmo.Client) error { return elDiscard(c.StudyPlan.Replaceable(ctx, 500, 3, 100)) },
			http.MethodGet, "/api/eduPlan/500/modules/3/disciplines/100/replaceable", "", ""},
		{"Replacements", func(c *myitmo.Client) error { return elDiscard(c.StudyPlan.Replacements(ctx, 500)) },
			http.MethodGet, "/api/eduPlan/500/replaceable_discs", "", ""},
		{"RequestReplacement", func(c *myitmo.Client) error { return elDiscard(c.StudyPlan.RequestReplacement(ctx, 500, 3, 100, 200)) },
			http.MethodPost, "/api/eduPlan/500/modules/3/disciplines/100/replace/200", "", ""},
		{"CancelReplacement", func(c *myitmo.Client) error { return c.StudyPlan.CancelReplacement(ctx, 500, 42) },
			http.MethodDelete, "/api/eduPlan/500/replace/42", "", ""},
	})
}

func TestStudyPlanDecode(t *testing.T) {
	f, c := newFake(t)
	f.result(`{"isu":100000,"programs":[{"planId":500,"specializationId":null,"name":"Programme","isActive":true}]}`)
	p := must[*myitmo.StudyPlanPrograms](t)(c.StudyPlan.Programs(ctx))
	if len(p.Programs) != 1 || p.Programs[0].PlanID != 500 || p.Programs[0].SpecializationID != nil || !p.Programs[0].IsActive {
		t.Errorf("programs = %+v", p)
	}

	f.result(`{"id":500,"currentSemester":5,"currentSemesterId":9,"semestersCount":8,
		"planInfo":{"directionCode":"01.03.02","directionName":"Direction","levelQualification":"bachelor","planType":"base","programName":"Programme","startYear":2024},
		"semesters":[{"semester":5,"semesterId":9,"semesterParity":1,"studyYear":"2026/2027"}],
		"structure":[{"id":3,"name":"Module","type":"module","moduleId":3,"choiceParameterId":1,"choiceAvailable":true,"rules":[1],
			"children":[{"id":100,"name":"Discipline A","type":"discipline","moduleId":3,"replaceable":true,"creditPoints":3,
				"department":{"id":1,"name":"Department","shortName":"DEP"},
				"contents":{"5":[{"id":11,"moduleId":3,"disciplineId":100,"order":1,"semester":5,"creditPoints":3,
					"activities":[{"id":1,"contentId":11,"name":"Lectures","volume":32,"workTypeId":1},{"id":2,"contentId":11,"name":"Exam","volume":null,"workTypeId":5}]}]}}]}]}`)
	plan := must[*myitmo.StudyPlan](t)(c.StudyPlan.Plan(ctx, 500, nil))
	m := plan.Structure[0]
	if *m.ChoiceParameterID != myitmo.StudyPlanRuleChooseN || len(m.Children) != 1 {
		t.Fatalf("module = %+v", m)
	}
	d := m.Children[0]
	acts := d.Contents["5"][0].Activities
	if !*d.Replaceable || d.Department.ShortName != "DEP" || *acts[0].Volume != 32 || acts[1].Volume != nil {
		t.Errorf("discipline = %+v", d)
	}
}

func TestStudyPlanLegacyDecode(t *testing.T) {
	f, c := newFake(t)
	f.result(`{"chosenContents":[{"moduleId":3,"disciplineId":100,"semester":5,"locked":true}]}`)
	ch := must[*myitmo.StudyPlanChoice](t)(c.StudyPlan.Choice(ctx, 500))
	if len(ch.ChosenContents) != 1 || ch.ChosenContents[0].DisciplineID != 100 || !ch.ChosenContents[0].Locked {
		t.Errorf("choice = %+v", ch)
	}

	f.result(`[{"id":200,"name":"Discipline B","abstract":"About"}]`)
	rs := must[[]myitmo.StudyPlanReplaceable](t)(c.StudyPlan.Replaceable(ctx, 500, 3, 100))
	if len(rs) != 1 || rs[0].ID != 200 || rs[0].Abstract != "About" {
		t.Errorf("replaceable = %+v", rs)
	}

	f.result(`[{"id":42,"moduleId":3,"disciplineIdFrom":100,"name":"Discipline B"}]`)
	ps := must[[]myitmo.StudyPlanReplacement](t)(c.StudyPlan.Replacements(ctx, 500))
	if len(ps) != 1 || ps[0].ID != 42 || ps[0].DisciplineIDFrom != 100 {
		t.Errorf("replacements = %+v", ps)
	}

	f.result(`42`)
	if id := must[int64](t)(c.StudyPlan.RequestReplacement(ctx, 500, 3, 100, 200)); id != 42 {
		t.Errorf("request id = %d", id)
	}
}
