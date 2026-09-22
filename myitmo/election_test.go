package myitmo_test

import (
	"net/http"
	"testing"

	"github.com/kewldan/go-itmo/myitmo"
)

// elRoute is one request assertion: the call, method, escaped path, encoded
// query and JSON body ("" means no body is checked to be empty).
type elRoute struct {
	name   string
	call   func(c *myitmo.Client) error
	method string
	path   string
	query  string
	body   string
}

func runElRoutes(t *testing.T, cases []elRoute) {
	t.Helper()
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			f, c := newFake(t)
			if err := tc.call(c); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			r := f.expect(tc.method, tc.path)
			if got := r.Query.Encode(); got != tc.query {
				t.Errorf("query = %q, want %q", got, tc.query)
			}
			if tc.body == "" {
				if len(r.Body) != 0 {
					t.Errorf("body = %s, want none", r.Body)
				}
				return
			}
			r.sameJSON(t, tc.body)
		})
	}
}

func elDiscard[T any](_ T, err error) error { return err }

func TestElectionRoutes(t *testing.T) {
	from, to := myitmo.NewDate(2026, 9, 1), myitmo.NewDate(2026, 12, 31)
	runElRoutes(t, []elRoute{
		{"Availability", func(c *myitmo.Client) error { return elDiscard(c.Election.Availability(ctx)) },
			http.MethodGet, "/api/election/students/availability", "", ""},
		{"SelectedFlowChains", func(c *myitmo.Client) error { return elDiscard(c.Election.SelectedFlowChains(ctx)) },
			http.MethodGet, "/api/election/students/selected_flow_chains", "", ""},
		{"AvailableDisciplines", func(c *myitmo.Client) error { return elDiscard(c.Election.AvailableDisciplines(ctx)) },
			http.MethodGet, "/api/election/students/available_disciplines", "", ""},
		{"ValidateDisciplines", func(c *myitmo.Client) error {
			return elDiscard(c.Election.ValidateDisciplines(ctx, []string{"g1", "g2"}))
		}, http.MethodPost, "/api/election/students/group_flow_available_disciplines", "", `["g1","g2"]`},
		{"ValidateDisciplinesNil", func(c *myitmo.Client) error { return elDiscard(c.Election.ValidateDisciplines(ctx, nil)) },
			http.MethodPost, "/api/election/students/group_flow_available_disciplines", "", `[]`},
		{"FlowLimits", func(c *myitmo.Client) error { return elDiscard(c.Election.FlowLimits(ctx)) },
			http.MethodGet, "/api/election/students/limits/flows", "", ""},
		{"FlowGroupLimits", func(c *myitmo.Client) error { return elDiscard(c.Election.FlowGroupLimits(ctx)) },
			http.MethodGet, "/api/election/students/limits/flow_groups", "", ""},
		{"SelectDisciplines", func(c *myitmo.Client) error {
			return elDiscard(c.Election.SelectDisciplines(ctx, []string{"g1"}))
		}, http.MethodPost, "/api/election/students/order/", "", `["g1"]`},
		{"ClearSelection", func(c *myitmo.Client) error { return c.Election.ClearSelection(ctx) },
			http.MethodPost, "/api/election/students/order/clear", "", ""},
		{"ChangeFlows", func(c *myitmo.Client) error { return elDiscard(c.Election.ChangeFlows(ctx, []int64{10, 20})) },
			http.MethodPost, "/api/election/students/order/change", "", `[10,20]`},
		{"ChosenFlows", func(c *myitmo.Client) error { return elDiscard(c.Election.ChosenFlows(ctx)) },
			http.MethodGet, "/api/election/students/chosen_flows", "", ""},
		{"OrderedFlowChains", func(c *myitmo.Client) error { return elDiscard(c.Election.OrderedFlowChains(ctx)) },
			http.MethodGet, "/api/election/students/ordered_flow_chains", "", ""},
		{"FlowIntersections", func(c *myitmo.Client) error {
			return elDiscard(c.Election.FlowIntersections(ctx, []int64{10, 20}))
		}, http.MethodPost, "/api/election/students/schedule/flows/intersections", "", `[10,20]`},
		{"BaseTimeline", func(c *myitmo.Client) error { return elDiscard(c.Election.BaseTimeline(ctx, from, to)) },
			http.MethodGet, "/api/election/students/schedule/base/timeline", "date_end=2026-12-31&date_start=2026-09-01", ""},
		{"BaseTimelineNoDates", func(c *myitmo.Client) error {
			return elDiscard(c.Election.BaseTimeline(ctx, myitmo.Date{}, myitmo.Date{}))
		}, http.MethodGet, "/api/election/students/schedule/base/timeline", "", ""},
		{"FlowTimeline", func(c *myitmo.Client) error { return elDiscard(c.Election.FlowTimeline(ctx, 10, from, to)) },
			http.MethodGet, "/api/election/students/schedule/flows/10/timeline", "date_end=2026-12-31&date_start=2026-09-01", ""},
		{"CombinedSchedule", func(c *myitmo.Client) error {
			return elDiscard(c.Election.CombinedSchedule(ctx, from, to, []int64{10}))
		}, http.MethodPost, "/api/election/students/schedule/combined", "date_end=2026-12-31&date_start=2026-09-01", `[10]`},
	})
}

func TestElectionAvailabilityDecode(t *testing.T) {
	f, c := newFake(t)
	f.result(`{"id":1,"status":"open","semesterStart":"2026-09-01T00:00:00+03:00","semesterEnd":"2027-01-31T00:00:00+03:00",
		"dateStart":"2026-08-20T10:00:00+03:00","dateEnd":"2026-08-27T18:00:00+03:00","timeStart":"10:00","timeEnd":"18:00",
		"studyYear":"2026/2027","semesterId":202620271,"semester":1}`)
	a := must[*myitmo.ElectionAvailability](t)(c.Election.Availability(ctx))
	if a.ID != myitmo.ElectionOpen || a.SemesterID != 202620271 || a.DateEnd.Hour() != 18 || a.TimeStart != "10:00" || a.Semester != 1 {
		t.Errorf("availability = %+v", a)
	}
}

func TestElectionAvailableDisciplinesDecode(t *testing.T) {
	f, c := newFake(t)
	f.result(`[{"dcId":1,"discId":100,"langId":2,"depName":"Department","depNameShort":"DEP","discName":"Discipline A",
		"langCode":"EN","required":false,"description":null,"notCompatibleWith":[101],
		"semesters":[{"semester":5,"statusId":1,"statusName":"","groupFlow":"gf-100-5",
			"themes":[{"children":[{"id":7,"ruleId":1,"ruleParams":[1,2],"themes":[{"groupFlow":"gf-t1","name":"Theme 1"}]},
				{"id":8,"ruleId":2,"ruleParams":null,"themes":[]}]}]},
			{"semester":6,"statusId":3,"statusName":"closed","groupFlow":"gf-100-6","themes":null}]}]`)
	ds := must[[]myitmo.ElectionAvailableDiscipline](t)(c.Election.AvailableDisciplines(ctx))
	if len(ds) != 1 || ds[0].DiscID != 100 || ds[0].NotCompatibleWith[0] != 101 || len(ds[0].Semesters) != 2 {
		t.Fatalf("disciplines = %+v", ds)
	}
	s := ds[0].Semesters[0]
	if s.StatusID != myitmo.ElectionDisciplineAvailable || s.GroupFlow != "gf-100-5" {
		t.Errorf("semester = %+v", s)
	}
	b := s.Themes[0].Children
	if len(b) != 2 || b[0].RuleID != myitmo.ElectionRuleChooseN || b[0].RuleParams[1] != 2 || b[0].Themes[0].GroupFlow != "gf-t1" || b[1].RuleParams != nil {
		t.Errorf("theme blocks = %+v", b)
	}
	if ds[0].Semesters[1].Themes != nil {
		t.Errorf("themes of second semester = %+v", ds[0].Semesters[1].Themes)
	}
}

func TestElectionValidationAndLimitsDecode(t *testing.T) {
	f, c := newFake(t)
	f.result(`{"disciplines":[100,200],"validVariantSelection":true,"validRequiredSelection":false,"needSelectRequired":1,
		"needSelectVariants":2,"needSelectVariantsMax":3,"scheduleAvailable":true}`)
	v := must[*myitmo.ElectionSelectionValidation](t)(c.Election.ValidateDisciplines(ctx, []string{"g"}))
	if len(v.Disciplines) != 2 || !v.ValidVariantSelection || v.ValidRequiredSelection || v.NeedSelectVariantsMax != 3 || !v.ScheduleAvailable {
		t.Errorf("validation = %+v", v)
	}

	f.result(`{"10":{"limitMax":30,"occupied":31,"free":-1},"20":{"limitMax":25,"occupied":5,"free":20}}`)
	l := must[map[int64]myitmo.ElectionFlowLimit](t)(c.Election.FlowLimits(ctx))
	if l[10].Free != -1 || l[20].LimitMax != 25 || l[20].Occupied != 5 {
		t.Errorf("limits = %+v", l)
	}

	f.result(`{"gf-100-5":{"limitMax":30,"free":4}}`)
	g := must[map[string]myitmo.ElectionFlowLimit](t)(c.Election.FlowGroupLimits(ctx))
	if g["gf-100-5"].Free != 4 {
		t.Errorf("group limits = %+v", g)
	}
}

func TestElectionFlowChainsDecode(t *testing.T) {
	f, c := newFake(t)
	f.result(`[{"groupFlow":"gf-100-5","disciplineId":100,"disciplineName":"Discipline A","flows":[
		{"id":1,"name":null,"limitMax":null,"teachers":[],"workType":0,"available":true,"selections":[],"variants":[
			{"id":10,"name":"Lecture 1","limitMax":30,"teachers":["Teacher A"],"workType":1,"available":true,"selections":[1],"variants":[]}]}]}]`)
	chains := must[[]myitmo.ElectionFlowChain](t)(c.Election.OrderedFlowChains(ctx))
	if len(chains) != 1 || chains[0].DisciplineID != 100 || len(chains[0].Flows) != 1 {
		t.Fatalf("chains = %+v", chains)
	}
	root := chains[0].Flows[0]
	if root.LimitMax != nil || len(root.Variants) != 1 {
		t.Fatalf("root = %+v", root)
	}
	leaf := root.Variants[0]
	if leaf.WorkType != myitmo.ElectionWorkTypeLecture || *leaf.LimitMax != 30 || leaf.Teachers[0] != "Teacher A" || leaf.Selections[0] != 1 {
		t.Errorf("leaf = %+v", leaf)
	}

	f.result(`{"status":1,"name":null,"flows":[10,20]}`)
	r := must[*myitmo.ElectionChangeResult](t)(c.Election.ChangeFlows(ctx, []int64{10, 20}))
	if *r.Status != 1 || len(r.Flows) != 2 {
		t.Errorf("change result = %+v", r)
	}

	f.result(`[10,20]`)
	if ids := must[[]int64](t)(c.Election.ChosenFlows(ctx)); len(ids) != 2 || ids[1] != 20 {
		t.Errorf("chosen = %v", ids)
	}
}

func TestElectionSelectedFlowChainsDecode(t *testing.T) {
	f, c := newFake(t)
	f.result(`{"selectedBy":2,"flowChains":[{"disciplineId":100,"discName":"Discipline A","flows":[{"flowId":10,"flowName":"Lecture 1"}]}]}`)
	s := must[*myitmo.ElectionSelectedFlowChains](t)(c.Election.SelectedFlowChains(ctx))
	if *s.SelectedBy != myitmo.ElectionSelectedByIndividualPlan || s.FlowChains[0].DisciplineID != 100 || s.FlowChains[0].Flows[0].FlowName != "Lecture 1" {
		t.Errorf("selected = %+v", s)
	}

	f.result(`null`)
	if s := must[*myitmo.ElectionSelectedFlowChains](t)(c.Election.SelectedFlowChains(ctx)); s != nil {
		t.Errorf("selected = %+v, want nil", s)
	}
}

func TestElectionScheduleDecode(t *testing.T) {
	f, c := newFake(t)
	f.result(`[{"flow1":{"id":10,"dateStart":"2026-09-07T08:20:00+03:00"},"flow2":{"id":20,"dateStart":"2026-09-07T08:20:00+03:00"},"intersectionType":3}]`)
	in := must[[]myitmo.ElectionFlowIntersection](t)(c.Election.FlowIntersections(ctx, []int64{10, 20}))
	if len(in) != 1 || in[0].Flow2.ID != 20 || in[0].IntersectionType != 3 || in[0].Flow1.DateStart.Minute() != 20 {
		t.Errorf("intersections = %+v", in)
	}

	f.result(`["2026-09-07","2026-09-14"]`)
	if w := must[[]string](t)(c.Election.BaseTimeline(ctx, myitmo.Date{}, myitmo.Date{})); len(w) != 2 {
		t.Errorf("timeline = %v", w)
	}

	f.result(`[{"timeStart":"2026-09-07T08:20:00+03:00","timeEnd":"2026-09-07T09:50:00+03:00","dateStart":"2026-09-07T00:00:00+03:00",
		"disciplineId":100,"disciplineName":"Discipline A","workTypeId":3,"teacherFio":"Teacher A","roomName":"101","buildingName":"Main","flowId":10}]`)
	ls := must[[]myitmo.ElectionLesson](t)(c.Election.CombinedSchedule(ctx, myitmo.NewDate(2026, 9, 7), myitmo.NewDate(2026, 9, 13), nil))
	if len(ls) != 1 || ls[0].WorkTypeID != myitmo.ElectionWorkTypePractice || ls[0].TimeEnd.Sub(ls[0].TimeStart).Minutes() != 90 || ls[0].FlowID != 10 {
		t.Errorf("lessons = %+v", ls)
	}
}

func TestElectionOrderErrorCode(t *testing.T) {
	f, c := newFake(t)
	f.reply(http.StatusBadRequest, `{"error_code":109,"error_message":"no places","result":null}`)
	_, err := c.Election.SelectDisciplines(ctx, []string{"g"})
	if myitmo.ErrorCode(err) != myitmo.ElectionErrNoPlaces {
		t.Errorf("err = %v", err)
	}
}
