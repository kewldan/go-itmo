package myitmo_test

import (
	"io"
	"net/http"
	"net/url"
	"reflect"
	"testing"
	"time"

	"github.com/kewldan/go-itmo/myitmo"
)

func TestSportRoutes(t *testing.T) {
	sportType := int64(3)
	tests := []struct {
		name   string
		call   func(c *myitmo.Client) error
		method string
		path   string
		query  url.Values
		body   string
	}{
		{"TimeSlots", func(c *myitmo.Client) error { _, err := c.Sport.TimeSlots(ctx); return err },
			http.MethodGet, "/api/sport/time_slots", nil, ""},
		{"SportTypes", func(c *myitmo.Client) error { _, err := c.Sport.SportTypes(ctx); return err },
			http.MethodGet, "/api/sport/sport_types", nil, ""},
		{"Semesters", func(c *myitmo.Client) error { _, err := c.Sport.Semesters(ctx); return err },
			http.MethodGet, "/api/sport/semesters/list", nil, ""},
		{"CurrentSemester", func(c *myitmo.Client) error { _, err := c.Sport.CurrentSemester(ctx); return err },
			http.MethodGet, "/api/sport/semesters/current", nil, ""},
		{"Filters", func(c *myitmo.Client) error { _, err := c.Sport.Filters(ctx); return err },
			http.MethodGet, "/api/sport/sign/schedule/filters", nil, ""},
		{"Schedule", func(c *myitmo.Client) error {
			_, err := c.Sport.Schedule(ctx, myitmo.SportScheduleParams{
				DateStart:    myitmo.NewDate(2026, 9, 21),
				DateEnd:      myitmo.NewDate(2026, 9, 28),
				BuildingID:   myitmo.SportBuildingLomonosova,
				SportTypeIDs: []int64{1, 2},
				TeacherISUs:  []int64{100001},
			})
			return err
		}, http.MethodGet, "/api/sport/sign/schedule", url.Values{
			"date_start": {"2026-09-21"}, "date_end": {"2026-09-28"}, "building_id": {"273"},
			"sport_type_id": {"1", "2"}, "teacher_isu": {"100001"},
		}, ""},
		{"ScheduleOnline", func(c *myitmo.Client) error {
			_, err := c.Sport.Schedule(ctx, myitmo.SportScheduleParams{
				DateStart: myitmo.NewDate(2026, 9, 21), DateEnd: myitmo.NewDate(2026, 9, 22), BuildingID: myitmo.SportBuildingOnline,
			})
			return err
		}, http.MethodGet, "/api/sport/sign/schedule", url.Values{
			"date_start": {"2026-09-21"}, "date_end": {"2026-09-22"}, "building_id": {"-1"},
		}, ""},
		{"Limits", func(c *myitmo.Client) error { _, err := c.Sport.Limits(ctx); return err },
			http.MethodGet, "/api/sport/sign/schedule/limits", nil, ""},
		{"OtherLessons", func(c *myitmo.Client) error { _, err := c.Sport.OtherLessons(ctx, 42); return err },
			http.MethodGet, "/api/sport/sign/schedule/lessons/42/other", nil, ""},
		{"SignIn", func(c *myitmo.Client) error { _, err := c.Sport.SignIn(ctx, 1, 2); return err },
			http.MethodPost, "/api/sport/sign/schedule/lessons", nil, `[1,2]`},
		{"SignOut", func(c *myitmo.Client) error { _, err := c.Sport.SignOut(ctx, 3); return err },
			http.MethodDelete, "/api/sport/sign/schedule/lessons", nil, `[3]`},
		{"SignInGroup", func(c *myitmo.Client) error { return c.Sport.SignInGroup(ctx, 7, nil) },
			http.MethodPost, "/api/sport/sign/schedule/lesson_groups/7", nil, ""},
		{"SignInGroupOpenForm", func(c *myitmo.Client) error {
			return c.Sport.SignInGroup(ctx, 7, &myitmo.SportOpenFormSubmit{
				SchoolName: myitmo.Ptr("School 1"), RankID: myitmo.Ptr[int64](2), SectionID: 9, ISU: 100001,
			})
		}, http.MethodPost, "/api/sport/sign/schedule/lesson_groups/7", nil,
			`{"school_name":"School 1","rank_id":2,"comment":null,"achievements":null,"section_id":9,"isu":100001}`},
		{"SignOutGroup", func(c *myitmo.Client) error { return c.Sport.SignOutGroup(ctx, 7) },
			http.MethodDelete, "/api/sport/sign/schedule/lesson_groups/7", nil, ""},
		{"Chosen", func(c *myitmo.Client) error { _, err := c.Sport.Chosen(ctx); return err },
			http.MethodGet, "/api/sport/sign/chosen", nil, ""},
		{"Score", func(c *myitmo.Client) error { _, err := c.Sport.Score(ctx, 11); return err },
			http.MethodGet, "/api/sport/personal/score", url.Values{"semester_id": {"11"}}, ""},
		{"Calendar", func(c *myitmo.Client) error {
			_, err := c.Sport.Calendar(ctx, myitmo.NewDate(2026, 9, 1), myitmo.NewDate(2026, 9, 30))
			return err
		}, http.MethodGet, "/api/sport/personal/calendar", url.Values{"date_start": {"2026-09-01"}, "date_end": {"2026-09-30"}}, ""},
		{"SignAttempts", func(c *myitmo.Client) error { _, err := c.Sport.SignAttempts(ctx); return err },
			http.MethodGet, "/api/sport/personal/sign_attempts", nil, ""},
		{"Attempts", func(c *myitmo.Client) error { _, err := c.Sport.Attempts(ctx); return err },
			http.MethodGet, "/api/sport/personal/have_attempts", nil, ""},
		{"Debt", func(c *myitmo.Client) error { _, err := c.Sport.Debt(ctx); return err },
			http.MethodGet, "/api/sport/personal/debt", nil, ""},
		{"Externat", func(c *myitmo.Client) error { _, err := c.Sport.Externat(ctx); return err },
			http.MethodGet, "/api/sport/personal/externat", nil, ""},
		{"HealthLevel", func(c *myitmo.Client) error { _, err := c.Sport.HealthLevel(ctx); return err },
			http.MethodGet, "/api/sport/personal/health_level", nil, ""},
		{"Selections", func(c *myitmo.Client) error { _, err := c.Sport.Selections(ctx); return err },
			http.MethodGet, "/api/sport/personal/selections", nil, ""},
		{"OpenForm", func(c *myitmo.Client) error { _, err := c.Sport.OpenForm(ctx, 9); return err },
			http.MethodGet, "/api/sport/personal/open_form", url.Values{"section": {"9"}}, ""},
		{"OpenFormRanks", func(c *myitmo.Client) error { _, err := c.Sport.OpenFormRanks(ctx); return err },
			http.MethodGet, "/api/sport/personal/open_form/ranks", nil, ""},
		{"Briefings", func(c *myitmo.Client) error { _, err := c.Sport.Briefings(ctx); return err },
			http.MethodGet, "/api/sport/personal/briefing/list", nil, ""},
		{"CreateBriefing", func(c *myitmo.Client) error { _, err := c.Sport.CreateBriefing(ctx); return err },
			http.MethodPost, "/api/sport/personal/briefing", nil, ""},
		{"SignBriefing", func(c *myitmo.Client) error { _, err := c.Sport.SignBriefing(ctx, 5); return err },
			http.MethodPost, "/api/sport/briefing/5/sign", nil, ""},
		{"Projects", func(c *myitmo.Client) error { _, err := c.Sport.Projects(ctx); return err },
			http.MethodGet, "/api/sport/projects/list", nil, ""},
		{"SignInProject", func(c *myitmo.Client) error { return c.Sport.SignInProject(ctx, 4, "https://example.com/x") },
			http.MethodPost, "/api/sport/sign/projects/4", nil, `{"link":"https://example.com/x"}`},
		{"SignInProjectNoLink", func(c *myitmo.Client) error {
			return c.Sport.SignInProject(ctx, myitmo.SportProjectKronbarsRunning, "")
		}, http.MethodPost, "/api/sport/sign/projects/1", nil, `{"link":null}`},
		{"SignOutProject", func(c *myitmo.Client) error { return c.Sport.SignOutProject(ctx, 4) },
			http.MethodDelete, "/api/sport/sign/projects/4", nil, ""},
		{"Competitions", func(c *myitmo.Client) error { _, err := c.Sport.Competitions(ctx, &sportType); return err },
			http.MethodGet, "/api/sport/competitions/list", url.Values{"sport_type_id": {"3"}}, ""},
		{"CompetitionsAll", func(c *myitmo.Client) error { _, err := c.Sport.Competitions(ctx, nil); return err },
			http.MethodGet, "/api/sport/competitions/list", url.Values{}, ""},
		{"CompetitionLimits", func(c *myitmo.Client) error { _, err := c.Sport.CompetitionLimits(ctx); return err },
			http.MethodGet, "/api/sport/competitions/list/limits", nil, ""},
		{"SignInCompetition", func(c *myitmo.Client) error { return c.Sport.SignInCompetition(ctx, 8, 21, 22) },
			http.MethodPost, "/api/sport/sign/competitions/8", nil, `[21,22]`},
		{"SignOutCompetition", func(c *myitmo.Client) error { return c.Sport.SignOutCompetition(ctx, 8) },
			http.MethodDelete, "/api/sport/sign/competitions/8", nil, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, c := newFake(t)
			if err := tt.call(c); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			r := f.expect(tt.method, tt.path)
			if tt.query != nil && !reflect.DeepEqual(r.Query, tt.query) {
				t.Errorf("query = %v, want %v", r.Query, tt.query)
			}
			if tt.body == "" {
				if len(r.Body) != 0 {
					t.Errorf("body = %s, want none", r.Body)
				}
			} else {
				r.sameJSON(t, tt.body)
			}
		})
	}
}

func TestSportDownloadBriefing(t *testing.T) {
	f, c := newFake(t)
	f.replyWith(http.StatusOK, http.Header{
		"Content-Type":        {"application/pdf"},
		"Content-Disposition": {`attachment; filename="signed.pdf"`},
	}, "%PDF-1.4")
	file := must[*myitmo.File](t)(c.Sport.DownloadBriefing(ctx, 5))
	defer file.Body.Close()
	f.expect(http.MethodGet, "/api/sport/personal/briefing/5/signed")
	data, _ := io.ReadAll(file.Body)
	if string(data) != "%PDF-1.4" {
		t.Errorf("body = %q", data)
	}
}

func TestSportScheduleDecode(t *testing.T) {
	f, c := newFake(t)
	f.result(`[{"date":"2026-09-21","lessons":[{
		"id":1001,"date":"2026-09-21T10:00:00+03:00","date_end":"2026-09-21T11:30:00+03:00",
		"section_id":9,"section_name":"Volleyball","section_level":2,"lesson_group_id":77,"lesson_level":1,"type_id":2,
		"building_id":null,"room_id":-1,"room_name":"Online","limit":20,"available":-1,"comment":null,
		"time_slot_id":3,"time_slot_start":"10:00","time_slot_end":"11:30","intersection":false,
		"can_sign_in":{"can_sign_in":false,"unavailable_reasons":["No places"]},
		"other_lessons":[{"id":1002,"date_start":"2026-09-28T10:00:00+03:00","weekday":0,"room_id":5,"room_name":"Hall",
			"time_slot_id":3,"time_start":null,"time_end":null,"type_id":5,"teacher_isu":100001,"teacher_fio":"Teacher A",
			"can_sign_in":{"can_sign_in":true,"unavailable_reasons":[]},"intersection":true,"repeatable":true}],
		"signed":false,"teacher_isu":100001,"teacher_fio":"Teacher A"}]},{"date":"2026-09-22","lessons":null}]`)
	days := must[[]myitmo.SportScheduleDay](t)(c.Sport.Schedule(ctx, myitmo.SportScheduleParams{
		DateStart: myitmo.NewDate(2026, 9, 21), DateEnd: myitmo.NewDate(2026, 9, 22),
	}))
	if len(days) != 2 || days[0].Date != myitmo.NewDate(2026, 9, 21) || days[1].Lessons != nil {
		t.Fatalf("days = %+v", days)
	}
	l := days[0].Lessons[0]
	if l.BuildingID != nil || l.RoomID != myitmo.SportRoomOnline || *l.Limit != 20 || *l.Available != -1 ||
		l.CanSignIn.CanSignIn || l.CanSignIn.UnavailableReasons[0] != "No places" || *l.Signed ||
		l.LessonLevel != myitmo.SportLessonLevelOpen || l.TypeID != myitmo.SportLessonTypeFreeVisit {
		t.Errorf("lesson = %+v", l)
	}
	if !l.Date.Equal(time.Date(2026, 9, 21, 10, 0, 0, 0, myitmo.MSK)) {
		t.Errorf("date = %v", l.Date)
	}
	o := l.OtherLessons[0]
	if time.Weekday(o.Weekday) != time.Sunday || o.TypeID != myitmo.SportLessonTypeDebt || !o.CanSignIn.CanSignIn ||
		!o.Intersection || o.TimeStart != "" || o.DateStart.Day() != 28 {
		t.Errorf("other lesson = %+v", o)
	}
}

func TestSportCalendarDecode(t *testing.T) {
	f, c := newFake(t)
	f.result(`[{"date":"2026-09-21","lessons":[{"id":5,"date":"2026-09-21T18:00:00+03:00","date_end":"2026-09-21T19:30:00+03:00",
		"section_name":"Yoga","section_level":1,"lesson_level":1,"lesson_group_id":6,"room_name":"Online","teacher_fio":"Teacher B",
		"comment":"bring a mat","link_url":"https://example.com/call"}]}]`)
	days := must[[]myitmo.SportScheduleDay](t)(c.Sport.Calendar(ctx, myitmo.NewDate(2026, 9, 21), myitmo.NewDate(2026, 9, 21)))
	l := days[0].Lessons[0]
	if l.LinkURL != "https://example.com/call" || l.Comment != "bring a mat" || l.Limit != nil || l.LessonGroupID != 6 {
		t.Errorf("lesson = %+v", l)
	}
}

func TestSportChosenDecode(t *testing.T) {
	f, c := newFake(t)
	f.result(`[{"id":9,"section_name":"Volleyball","level":2,"lesson_groups":[{"id":77,"level":2,"level_name":"Basic",
		"has_future_lessons":true,
		"lessons":[{"id":1001,"date_start":"2026-09-21T10:00:00+03:00","date_end":"2026-09-21T11:30:00+03:00","time_slot_id":3,
			"time_start":"10:00","time_end":"11:30","room_id":5,"room_name":"Hall","teacher_isu":100001,"teacher_fio":"Teacher A",
			"type_id":2,"link_url":null,"comment":null,"intersection":false}],
		"weekdays":[{"weekday":1,"date_start":"2026-09-21T10:00:00+03:00","room_name":"Hall","teacher_fio":"Teacher A",
			"time_start":null,"time_end":null,"time_slot_id":3}]}]}]`)
	sections := must[[]myitmo.ChosenSportSection](t)(c.Sport.Chosen(ctx))
	g := sections[0].LessonGroups[0]
	if sections[0].Level != 2 || g.LevelName != "Basic" || !g.HasFutureLessons || g.Lessons[0].TeacherISU != 100001 {
		t.Errorf("group = %+v", g)
	}
	w := g.Weekdays[0]
	if time.Weekday(w.Weekday) != time.Monday || w.RoomName != "Hall" || w.TimeSlotID != 3 || w.DateStart.IsZero() {
		t.Errorf("weekday = %+v", w)
	}
}

func TestSportLimitsDecode(t *testing.T) {
	f, c := newFake(t)
	f.result(`{"77":{"1001":{"limit":20,"available":3},"1002":{"limit":20,"available":-2}}}`)
	limits := must[map[int64]map[int64]myitmo.SportSignLimit](t)(c.Sport.Limits(ctx))
	if limits[77][1001].Available != 3 || limits[77][1002].Available != -2 || limits[77][1001].Limit != 20 {
		t.Errorf("limits = %+v", limits)
	}

	f.result(`[]`)
	empty := must[map[int64]map[int64]myitmo.SportSignLimit](t)(c.Sport.Limits(ctx))
	if empty == nil || len(empty) != 0 {
		t.Errorf("empty limits = %#v", empty)
	}

	f.result(`{"8":{"limit":100,"available":40}}`)
	comp := must[map[int64]myitmo.SportSignLimit](t)(c.Sport.CompetitionLimits(ctx))
	if comp[8].Available != 40 {
		t.Errorf("competition limits = %+v", comp)
	}
}

func TestSportOtherLessonsShapes(t *testing.T) {
	f, c := newFake(t)
	f.result(`[1002,1003]`)
	o := must[*myitmo.SportOtherSelection](t)(c.Sport.OtherLessons(ctx, 1001))
	if !reflect.DeepEqual(o.LessonIDs, []int64{1002, 1003}) || o.Signed {
		t.Errorf("free visit = %+v", o)
	}
	f.result(`{"signed":true}`)
	o = must[*myitmo.SportOtherSelection](t)(c.Sport.OtherLessons(ctx, 1001))
	if o.LessonIDs != nil || !o.Signed {
		t.Errorf("section = %+v", o)
	}
}

func TestSportSignInReturnsIDsAndErrors(t *testing.T) {
	f, c := newFake(t)
	f.result(`[1001]`)
	ids := must[[]int64](t)(c.Sport.SignIn(ctx, 1001))
	if !reflect.DeepEqual(ids, []int64{1001}) {
		t.Errorf("ids = %v", ids)
	}
	f.reply(http.StatusOK, `{"error_code":137,"error_message":"нельзя записать студента: [Выбрано 1 занятие в этот день]","result":null}`)
	if _, err := c.Sport.SignIn(ctx, 1001); myitmo.ErrorCode(err) != myitmo.SportErrorCannotSignIn {
		t.Errorf("err = %v", err)
	}
	f.reply(http.StatusOK, `{"error_code":130,"error_message":"нельзя отписать студента: [Вы не записаны на это занятие]","result":null}`)
	if _, err := c.Sport.SignOut(ctx, 1001); myitmo.ErrorCode(err) != myitmo.SportErrorCannotSignOut {
		t.Errorf("err = %v", err)
	}
	f.reply(http.StatusOK, `{"error_code":135,"error_message":"closed","result":null}`)
	if _, err := c.Sport.Filters(ctx); myitmo.ErrorCode(err) != myitmo.SportErrorChoiceClosed {
		t.Errorf("err = %v", err)
	}
}

func TestSportPersonalDecode(t *testing.T) {
	f, c := newFake(t)

	f.result(`{"id":12,"study_year":"2026/2027","semester":0,"date_start":"2026-09-01T00:00:00+03:00","date_end":"2026-12-20T00:00:00+03:00",
		"hard_date_end":"2026-12-31T00:00:00+03:00","current":true,"choice_start":"2026-08-25T10:00:00+03:00","bachelor_bound":null,
		"ppa1_start":"2026-10-01T00:00:00+03:00","ppa1_end":"2026-10-15T00:00:00+03:00","ppa2_start":null,"ppa2_end":null,"sign_duration":14}`)
	sem := must[*myitmo.SportSemester](t)(c.Sport.CurrentSemester(ctx))
	if sem.ID != 12 || !sem.Current || sem.SignDuration != 14 || sem.ChoiceStart.Day() != 25 || !sem.BachelorBound.IsZero() {
		t.Errorf("semester = %+v", sem)
	}

	f.result(`[{"id":12,"value":"Осень 2026/2027","comment":null}]`)
	if opts := must[[]myitmo.SportSemesterOption](t)(c.Sport.Semesters(ctx)); opts[0].Value != "Осень 2026/2027" {
		t.Errorf("options = %+v", opts)
	}

	f.result(`{"building_id":[{"id":-1,"value":"Online"},{"id":0,"value":"Other"},{"id":273,"value":"Building"}],
		"section_id":[{"id":9,"value":"Volleyball"}],"sport_type_id":[{"id":1,"value":"Swimming"}],"teacher_isu":[{"id":100001,"value":"Teacher A"}]}`)
	fl := must[*myitmo.SportFilters](t)(c.Sport.Filters(ctx))
	if fl.BuildingID[0].ID != myitmo.SportBuildingOnline || fl.TeacherISU[0].ID != 100001 || fl.SectionID[0].Value != "Volleyball" {
		t.Errorf("filters = %+v", fl)
	}

	f.result(`[{"id":1,"time_start":"08:20","time_end":"09:50"}]`)
	if slots := must[[]myitmo.SportTimeSlot](t)(c.Sport.TimeSlots(ctx)); slots[0].TimeEnd != "09:50" {
		t.Errorf("slots = %+v", slots)
	}

	f.result(`{"sum":{"attendances":42,"other":10},"attendances":[{"type":"competition","name":"Cup","evaluation_id":3,
		"evaluation_name":"Participation","section_level":1,"score":10,"date":"2026-10-01T12:00:00+03:00","is_competition":true,
		"discipline_name":"Swimming","competition_name":"Cup","place":"Участие"}]}`)
	score := must[*myitmo.SportScore](t)(c.Sport.Score(ctx, 12))
	if score.Sum.Attendances != 42 || score.Sum.Other != 10 || score.Attendances[0].Type != myitmo.SportAttendanceCompetition ||
		!score.Attendances[0].IsCompetition || score.Attendances[0].Score != 10 {
		t.Errorf("score = %+v", score)
	}
	f.result(`{"sum":{"attendances":0,"other":0},"attendances":null}`)
	if score := must[*myitmo.SportScore](t)(c.Sport.Score(ctx, 12)); score.Attendances != nil {
		t.Errorf("empty score = %+v", score)
	}

	f.result(`3`)
	if n := must[int](t)(c.Sport.SignAttempts(ctx)); n != 3 {
		t.Errorf("sign attempts = %d", n)
	}

	f.result(`{"total_attempts":2,"used_attempts":1,"free_attempts":1,"can_sign_in":true}`)
	if a := must[*myitmo.SportAttempts](t)(c.Sport.Attempts(ctx)); a.TotalAttempts != 2 || a.FreeAttempts != 1 || !a.CanSignIn {
		t.Errorf("attempts = %+v", a)
	}

	f.result(`{"is_having_debt":true,"needed_score":12.5,"free_attempts":2}`)
	if d := must[*myitmo.SportDebt](t)(c.Sport.Debt(ctx)); !d.IsHavingDebt || *d.NeededScore != 12.5 || *d.FreeAttempts != 2 {
		t.Errorf("debt = %+v", d)
	}
	f.result(`{"is_having_debt":false}`)
	if d := must[*myitmo.SportDebt](t)(c.Sport.Debt(ctx)); d.NeededScore != nil || d.FreeAttempts != nil {
		t.Errorf("no debt = %+v", d)
	}

	f.result(`{"signed":false,"externat_status_id":2,"decline_reason":"Late"}`)
	if e := must[*myitmo.SportExternat](t)(c.Sport.Externat(ctx)); *e.ExternatStatusID != myitmo.SportExternatDeclined || e.DeclineReason != "Late" {
		t.Errorf("externat = %+v", e)
	}

	f.result(`{"health_level":{"id":1,"isu":100001,"name":"Основная группа здоровья"}}`)
	if h := must[*myitmo.SportHealthLevel](t)(c.Sport.HealthLevel(ctx)); h == nil || h.Name != "Основная группа здоровья" {
		t.Errorf("health = %+v", h)
	}
	f.result(`{"health_level":null}`)
	if h := must[*myitmo.SportHealthLevel](t)(c.Sport.HealthLevel(ctx)); h != nil {
		t.Errorf("health = %+v, want nil", h)
	}

	f.result(`[{"id":1,"name":"Swimming","requisites":[{"id":2,"level_name":"Сборная команда","name":"Swimming, team"}]}]`)
	if s := must[[]myitmo.SportSelection](t)(c.Sport.Selections(ctx)); s[0].Requisites[0].LevelName != "Сборная команда" {
		t.Errorf("selections = %+v", s)
	}
}

func TestSportFormsAndProjectsDecode(t *testing.T) {
	f, c := newFake(t)

	f.result(`{"id":4,"school_name":"School 1","rank_id":2,"comment":null,"achievements":"Cup"}`)
	if o := must[*myitmo.SportOpenForm](t)(c.Sport.OpenForm(ctx, 9)); o.ID != 4 || *o.RankID != 2 || o.Achievements != "Cup" {
		t.Errorf("open form = %+v", o)
	}
	f.result(`null`)
	if o := must[*myitmo.SportOpenForm](t)(c.Sport.OpenForm(ctx, 9)); o != nil {
		t.Errorf("open form = %+v, want nil", o)
	}

	f.result(`[{"id":1,"value":"КМС"}]`)
	if r := must[[]myitmo.IDValue](t)(c.Sport.OpenFormRanks(ctx)); r[0].Value != "КМС" {
		t.Errorf("ranks = %+v", r)
	}

	f.result(`{"signed":true,"briefs":[{"id":5,"url":"https://example.com/brief.pdf"}]}`)
	if b := must[*myitmo.SportBriefingList](t)(c.Sport.Briefings(ctx)); !b.Signed || b.Briefs[0].URL != "https://example.com/brief.pdf" {
		t.Errorf("briefings = %+v", b)
	}

	f.result(`{"signature_id":"abc","sign_task_id":17}`)
	if s := must[*myitmo.SportBriefingSignTask](t)(c.Sport.CreateBriefing(ctx)); string(s.SignatureID) != `"abc"` || string(s.SignTaskID) != `17` {
		t.Errorf("sign task = %+v", s)
	}

	f.result(`[{"id":1,"name":"Kronbars Running","description":"d","signed":false,"limit":100,"available":50,
		"instruction_link":null,"instruction_description":"I agree","requisite_available":true,"link":null,"health_level_id":1}]`)
	p := must[[]myitmo.SportProject](t)(c.Sport.Projects(ctx))
	if p[0].ID != myitmo.SportProjectKronbarsRunning || !p[0].RequisiteAvailable || *p[0].HealthLevelID != myitmo.SportHealthLevelSpecial {
		t.Errorf("projects = %+v", p)
	}

	f.result(`[{"id":8,"name":"Cup","sport_type_name":"Swimming","date_start":"2026-11-01T10:00:00+03:00","building_name":"Pool",
		"disciplines":[{"id":21,"name":"50 m","signed":true}],"registration_link":null,"free_registration":true,"available":40}]`)
	comp := must[[]myitmo.SportCompetition](t)(c.Sport.Competitions(ctx, nil))
	if comp[0].Disciplines[0].ID != 21 || !comp[0].Disciplines[0].Signed || !comp[0].FreeRegistration || comp[0].DateStart.Month() != time.November {
		t.Errorf("competitions = %+v", comp)
	}
}
