package myitmo_test

import (
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/kewldan/go-itmo/myitmo"
)

func electiveServices(c *myitmo.Client) map[string]*myitmo.ElectiveCourseService {
	return map[string]*myitmo.ElectiveCourseService{"intro": c.Intro, "facultative": c.Facultative}
}

func TestElectiveRoutes(t *testing.T) {
	for _, prefix := range []string{"intro", "facultative"} {
		cases := []struct {
			name   string
			method string
			path   string
			body   string
			do     func(*myitmo.ElectiveCourseService) error
		}{
			{"Tree", http.MethodGet, "/json/", "", func(s *myitmo.ElectiveCourseService) error { _, err := s.Tree(ctx); return err }},
			{"TreeOf", http.MethodGet, "/json/100001", "", func(s *myitmo.ElectiveCourseService) error { _, err := s.TreeOf(ctx, 100001); return err }},
			{"Current", http.MethodGet, "/current/", "", func(s *myitmo.ElectiveCourseService) error { _, err := s.Current(ctx); return err }},
			{"CurrentOf", http.MethodGet, "/current/100001", "", func(s *myitmo.ElectiveCourseService) error { _, err := s.CurrentOf(ctx, 100001); return err }},
			{"Book", http.MethodPost, "/booking/", `{"selected_flows":[11,12],"canceled_flows":[]}`, func(s *myitmo.ElectiveCourseService) error {
				_, err := s.Book(ctx, []int64{11, 12}, nil)
				return err
			}},
			{"Commit", http.MethodPost, "/commit/", `{"status":2,"flow_id":[11,12]}`, func(s *myitmo.ElectiveCourseService) error {
				return s.Commit(ctx, myitmo.ElectiveCommitConfirm, []int64{11, 12})
			}},
			{"CommitReset", http.MethodPost, "/commit/", `{"status":0,"flow_id":[]}`, func(s *myitmo.ElectiveCourseService) error {
				return s.Commit(ctx, myitmo.ElectiveCommitReset, nil)
			}},
			{"FlowSchedule", http.MethodGet, "/schedule/flow/42/", "", func(s *myitmo.ElectiveCourseService) error { _, err := s.FlowSchedule(ctx, 42); return err }},
			{"UserSchedule", http.MethodGet, "/schedule/user/", "", func(s *myitmo.ElectiveCourseService) error { _, err := s.UserSchedule(ctx); return err }},
			{"Descriptions", http.MethodGet, "/description/", "", func(s *myitmo.ElectiveCourseService) error { _, err := s.Descriptions(ctx); return err }},
			{"Catalog", http.MethodGet, "/description/start/", "", func(s *myitmo.ElectiveCourseService) error { _, err := s.Catalog(ctx); return err }},
			{"Limits", http.MethodGet, "/limits/", "", func(s *myitmo.ElectiveCourseService) error { _, err := s.Limits(ctx); return err }},
			{"LimitsOf", http.MethodGet, "/limits/100001", "", func(s *myitmo.ElectiveCourseService) error { _, err := s.LimitsOf(ctx, 100001); return err }},
			{"Status", http.MethodGet, "/status/", "", func(s *myitmo.ElectiveCourseService) error { _, err := s.Status(ctx); return err }},
		}
		for _, tc := range cases {
			t.Run(prefix+"/"+tc.name, func(t *testing.T) {
				f, c := newFake(t)
				if err := tc.do(electiveServices(c)[prefix]); err != nil {
					t.Fatal(err)
				}
				r := f.expect(tc.method, "/api/"+prefix+tc.path)
				if len(r.Query) != 0 {
					t.Errorf("query = %v, want none", r.Query)
				}
				if tc.body != "" {
					r.sameJSON(t, tc.body)
				}
			})
		}
	}
}

func TestElectiveTreeDecode(t *testing.T) {
	f, c := newFake(t)
	f.result(`{"json":[{"id":1,"name":"Group","required":true,"selections":[1],"color_index":null,"variants":[
		{"id":2,"name":"Module","required":false,"selections":[2,3],"color_index":4,"variants":[
			{"id":3,"name":"Discipline","language":"en","flow_info":1,"selectable":true,"available_selections":[[10,11]],"variants":[
				{"id":10,"name":"Flow","work_type":1,"teachers":["Teacher A"],"variants":[]}]}]}]}]}`)
	tree := must[[]myitmo.ElectiveNode](t)(c.Facultative.Tree(ctx))
	if len(tree) != 1 || tree[0].ColorIndex != nil || !tree[0].Required {
		t.Fatalf("tree = %+v", tree)
	}
	mod := tree[0].Variants[0]
	if *mod.ColorIndex != 4 || len(mod.Selections) != 2 {
		t.Errorf("module = %+v", mod)
	}
	disc := mod.Variants[0]
	if disc.Language != "en" || *disc.FlowInfo != 1 || !disc.Selectable || disc.AvailableSelections[0][1] != 11 {
		t.Errorf("discipline = %+v", disc)
	}
	flow := disc.Variants[0]
	if flow.ID != 10 || *flow.WorkType != 1 || flow.Teachers[0] != "Teacher A" {
		t.Errorf("flow = %+v", flow)
	}
}

func TestElectiveCurrentDecode(t *testing.T) {
	f, c := newFake(t)
	f.result(`{"status":1,"flow_id":[10,20],"intersections":[{"flow1":{"flow_id":10,"date":"2026-09-07T10:00:00+03:00","date_start":"2026-09-07T08:20:00+03:00"},"flow2":{"flow_id":20,"date":"2026-09-07T10:00:00+03:00","date_start":"2026-09-07T08:20:00+03:00"},"intersection_type":2}]}`)
	cur := must[*myitmo.ElectiveCurrent](t)(c.Intro.Current(ctx))
	if cur.Status != myitmo.ElectiveStatusBooked || len(cur.FlowID) != 2 {
		t.Fatalf("current = %+v", cur)
	}
	x := cur.Intersections[0]
	if x.IntersectionType != myitmo.ElectiveClashHard || x.Flow2.FlowID != 20 || x.Flow1.DateStart.Hour() != 8 {
		t.Errorf("intersection = %+v", x)
	}
}

func TestElectiveBookClashes(t *testing.T) {
	f, c := newFake(t)
	f.reply(http.StatusOK, `{"error_code":109,"error_message":"clash","result":[{"flow1":{"flow_id":10,"date":"2026-09-07T10:00:00+03:00","date_start":"2026-09-07T08:20:00+03:00"},"flow2":{"flow_id":20,"date":"2026-09-07T10:00:00+03:00","date_start":"2026-09-07T08:20:00+03:00"},"intersection_type":1}]}`)
	b, err := c.Intro.Book(ctx, []int64{20}, []int64{30})
	f.expect(http.MethodPost, "/api/intro/booking/").sameJSON(t, `{"selected_flows":[20],"canceled_flows":[30]}`)
	if myitmo.ErrorCode(err) != myitmo.ElectiveCodeClashes {
		t.Fatalf("err = %v", err)
	}
	if b == nil || len(b.Clashes) != 1 || b.Clashes[0].IntersectionType != myitmo.ElectiveClashSoft {
		t.Fatalf("booking = %+v", b)
	}

	f.reply(http.StatusOK, `{"error_code":0,"result":null}`)
	b = must[*myitmo.ElectiveBooking](t)(c.Intro.Book(ctx, []int64{20}, nil))
	if len(b.Clashes) != 0 {
		t.Errorf("clashes = %+v", b.Clashes)
	}

	f.reply(http.StatusOK, `{"error_code":110,"error_message":"no seats","result":null}`)
	b, err = c.Facultative.Book(ctx, []int64{20}, nil)
	var e *myitmo.Error
	if b != nil || !errors.As(err, &e) || e.Code != 110 || e.Message != "no seats" {
		t.Errorf("booking = %+v, err = %v", b, err)
	}
}

func TestElectiveScheduleAndDescriptionsDecode(t *testing.T) {
	f, c := newFake(t)
	f.result(`{"json":[{"date":"2026-09-07","lessons":[{"time_start":"08:20","time_end":"09:50","teacher_name":"Teacher A","room":"101","building":"Main","work_type_id":1,"subject":"Course"}]}]}`)
	days := must[[]myitmo.ScheduleDay](t)(c.Intro.FlowSchedule(ctx, 42))
	if len(days) != 1 || days[0].Date != myitmo.NewDate(2026, time.September, 7) || days[0].Lessons[0].TimeStart != "08:20" || days[0].Lessons[0].WorkTypeID != 1 {
		t.Errorf("days = %+v", days)
	}

	f.result(`{"description":[{"id":3,"description":"About"}]}`)
	d := must[[]myitmo.ElectiveDescription](t)(c.Intro.Descriptions(ctx))
	if len(d) != 1 || d[0].ID != 3 || d[0].Description != "About" {
		t.Errorf("descriptions = %+v", d)
	}

	f.result(`{"description":[{"id":1,"name":"Group","variants":[{"name":"Module","variants":[{"id":3,"name":"Discipline","description":"About","url":null}]}]}]}`)
	cat := must[[]myitmo.ElectiveCatalogGroup](t)(c.Facultative.Catalog(ctx))
	if len(cat) != 1 || cat[0].Variants[0].Variants[0].ID != 3 || cat[0].Variants[0].Name != "Module" {
		t.Errorf("catalog = %+v", cat)
	}
}

func TestElectiveLimitsAndStatusDecode(t *testing.T) {
	f, c := newFake(t)
	f.result(`{"limits":{"10":{"limit":5,"limit_max":30},"20":{"limit":null,"limit_max":25}},"start_time":"2026-09-01T10:00:00+03:00","end_time":"2026-09-10T23:59:00+03:00"}`)
	l := must[*myitmo.ElectiveLimits](t)(c.Intro.Limits(ctx))
	if *l.Limits[10].Limit != 5 || l.Limits[10].LimitMax != 30 || l.Limits[20].Limit != nil || l.EndTime.Day() != 10 {
		t.Errorf("limits = %+v", l)
	}

	f.result(`{"limits":[{"flow_id":10,"limit":0,"limit_max":30}]}`)
	lo := must[[]myitmo.ElectiveFlowLimitEntry](t)(c.Facultative.LimitsOf(ctx, 100001))
	if len(lo) != 1 || lo[0].FlowID != 10 || *lo[0].Limit != 0 {
		t.Errorf("limits of = %+v", lo)
	}

	f.result(`{"semester":{"date_start":"2026-09-01","date_end":"2027-01-31","choice_status":97,"semester":1,"study_year":"2026/2027"}}`)
	st := must[*myitmo.ElectiveSemester](t)(c.Intro.Status(ctx))
	if st.ChoiceStatus != myitmo.ElectiveChoiceNotOpened || st.Semester != myitmo.ElectiveSemesterAutumn || st.DateEnd != myitmo.NewDate(2027, time.January, 31) || st.StudyYear != "2026/2027" {
		t.Errorf("status = %+v", st)
	}
}
