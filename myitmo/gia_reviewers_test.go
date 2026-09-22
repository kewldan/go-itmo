package myitmo_test

import (
	"net/url"
	"testing"

	"github.com/kewldan/go-itmo/myitmo"
)

func TestGIAReviewerRequests(t *testing.T) {
	yes := true
	in := myitmo.GIAReviewerInput{ReviewerName: "Тест", ReviewerSurname: "Рецензентов", ReviewerEmail: "reviewer@example.com",
		DegreeTypeID: myitmo.Ptr[int64](1), Workplace: myitmo.Ptr("ООО Тест"), INN: myitmo.Ptr[int64](7700000000)}
	inJSON := `{"reviewer_name":"Тест","reviewer_surname":"Рецензентов","reviewer_second_name":"","reviewer_email":"reviewer@example.com",
		"degree_type_id":1,"academic_title_id":null,"workplace":"ООО Тест","inn":7700000000,"job_title":null,"organization_status_id":null}`
	rev := myitmo.GIAReviewerReview{Approved: true, RelevantContent: 1, JustificationRelevance: 2, RelevantEdu: 3, CorrectMethods: 4,
		QualityLogic: 5, JustificationAssertions: 6, Value: 7, Integration: 8, Advantages: "плюсы", Disadvantages: "минусы",
		Comment: "ок", Assessment: 21, ThesisComplete: true, AwardingQualifications: false}
	revJSON := `{"approved":true,"relevant_content":1,"justification_relevance":2,"relevant_edu":3,"correct_methods":4,"quality_logic":5,
		"justification_assertions":6,"value":7,"integration":8,"advantages":"плюсы","disadvantages":"минусы","comment":"ок",
		"assessment":21,"thesis_complete":true,"awarding_qualifications":false}`
	runGIACases(t, []giaCase{
		{"CreateReviewer", func(c *myitmo.Client) error { _, err := c.GIA.CreateReviewer(ctx, in); return err },
			"POST", "/api/gia/reviewers", nil, inJSON},
		{"Reviewer", func(c *myitmo.Client) error { _, err := c.GIA.Reviewer(ctx, 5); return err },
			"GET", "/api/gia/reviewers/5", nil, ""},
		{"UpdateReviewer", func(c *myitmo.Client) error { return c.GIA.UpdateReviewer(ctx, 5, in) },
			"PATCH", "/api/gia/reviewers/5", nil, inJSON},
		{"ReviewerStudents", func(c *myitmo.Client) error { _, err := c.GIA.ReviewerStudents(ctx, 5); return err },
			"GET", "/api/gia/reviewers/5/students", nil, ""},
		{"Reviewers", func(c *myitmo.Client) error {
			_, err := c.GIA.Reviewers(ctx, myitmo.GIAReviewersParams{Limit: 25, Offset: 25, Query: "тест", ShowAs: myitmo.GIARoleEPManager})
			return err
		}, "GET", "/api/gia/reviewers/reviewers", url.Values{"limit": {"25"}, "offset": {"25"}, "query": {"тест"}, "show_as": {"ep_manager"}}, ""},
		{"ReviewedStudents", func(c *myitmo.Client) error {
			_, err := c.GIA.ReviewedStudents(ctx, myitmo.GIAReviewedStudentsParams{Limit: 25, Year: 2025, GroupID: "Z3400", EduDirection: 8, EduProgram: 7, Query: "тест"})
			return err
		}, "GET", "/api/gia/reviewers/students", url.Values{"limit": {"25"}, "offset": {"0"}, "year": {"2025"}, "group_id": {"Z3400"},
			"edu_direction": {"8"}, "edu_program": {"7"}, "query": {"тест"}}, ""},
		{"ReviewedStudentFilters", func(c *myitmo.Client) error { _, err := c.GIA.ReviewedStudentFilters(ctx); return err },
			"GET", "/api/gia/reviewers/students/filters", nil, ""},
		{"SearchReviewerStudents", func(c *myitmo.Client) error {
			_, err := c.GIA.SearchReviewerStudents(ctx, "100001", 1, myitmo.GIARoleOSOP)
			return err
		}, "GET", "/api/gia/reviewers/students/search", url.Values{"query": {"100001"}, "limit": {"1"}, "show_as": {"osop"}}, ""},
		{"Appointments", func(c *myitmo.Client) error {
			_, err := c.GIA.Appointments(ctx, myitmo.GIAAppointmentsParams{Limit: 25, Year: 2025, GroupID: "Z3400", EduDirection: 8,
				EduProgram: 7, Reviewer: 5, EduLevel: 1, HasReview: &yes, ShowAs: myitmo.GIARoleOSOP, Query: "тест"})
			return err
		}, "GET", "/api/gia/reviewers/appointments", url.Values{"limit": {"25"}, "offset": {"0"}, "year": {"2025"}, "group_id": {"Z3400"},
			"edu_direction": {"8"}, "edu_program": {"7"}, "reviewer": {"5"}, "edu_level": {"1"}, "has_review": {"true"},
			"show_as": {"osop"}, "query": {"тест"}}, ""},
		{"AppointmentFilters", func(c *myitmo.Client) error { _, err := c.GIA.AppointmentFilters(ctx, ""); return err },
			"GET", "/api/gia/reviewers/appointments/filters", nil, ""},
		{"AppointReviewer", func(c *myitmo.Client) error { return c.GIA.AppointReviewer(ctx, 5, 11, 12) },
			"POST", "/api/gia/reviewers/appoint/students", nil, `{"reviewer_id":5,"student_ids":[11,12]}`},
		{"NotifyReviewers", func(c *myitmo.Client) error { return c.GIA.NotifyReviewers(ctx, 5) },
			"POST", "/api/gia/reviewers/notify", nil, `{"reviewer_ids":[5]}`},
		{"RestrictAppointments", func(c *myitmo.Client) error { return c.GIA.RestrictAppointments(ctx) },
			"POST", "/api/gia/reviewers/edit-restriction", nil, `{}`},
		{"RegisterReviewer", func(c *myitmo.Client) error { return c.GIA.RegisterReviewer(ctx, "invite-token") },
			"POST", "/api/gia/reviewers/register", nil, `{"token":"invite-token"}`},
		{"ReviewerDiploma", func(c *myitmo.Client) error { _, err := c.GIA.ReviewerDiploma(ctx, 31); return err },
			"GET", "/api/gia/reviewers/diplomas/31/student", nil, ""},
		{"SubmitReviewerReview", func(c *myitmo.Client) error { return c.GIA.SubmitReviewerReview(ctx, 31, rev) },
			"PUT", "/api/gia/reviewers/diplomas/31/review/reviewer", nil, revJSON},
		{"AddReviewerQuestion", func(c *myitmo.Client) error { return c.GIA.AddReviewerQuestion(ctx, 31, "Почему?") },
			"POST", "/api/gia/reviewers/diplomas/31/review/reviewer/questions", nil, `{"question":"Почему?"}`},
		{"UpdateReviewerQuestion", func(c *myitmo.Client) error { return c.GIA.UpdateReviewerQuestion(ctx, 31, 9, "Зачем?") },
			"PATCH", "/api/gia/reviewers/diplomas/31/review/reviewer/questions/9", nil, `{"question":"Зачем?"}`},
		{"DeleteReviewerQuestion", func(c *myitmo.Client) error { return c.GIA.DeleteReviewerQuestion(ctx, 31, 9) },
			"DELETE", "/api/gia/reviewers/diplomas/31/review/reviewer/questions/9", nil, ""},
	})
}

func TestGIAReviewerDecodes(t *testing.T) {
	f, c := newFake(t)
	f.result(`{"reviewer_id":5}`)
	if id := must[int64](t)(c.GIA.CreateReviewer(ctx, myitmo.GIAReviewerInput{})); id != 5 {
		t.Errorf("id = %d", id)
	}

	f.result(`{"reviewers":[{"reviewer_id":5,"reviewer_surname":"Рецензентов","reviewer_email":"reviewer@example.com",
		"academic_title":"доцент","degree_type":"к.т.н.","job_title":"инженер","workplace":"ООО Тест","inn":null,"registered":true}],"total":1}`)
	list := must[*myitmo.GIAReviewerList](t)(c.GIA.Reviewers(ctx, myitmo.GIAReviewersParams{}))
	if list.Total != 1 || !list.Reviewers[0].Registered || list.Reviewers[0].INN != nil {
		t.Errorf("reviewers = %+v", list)
	}

	f.result(`{"appointments":[{"id":11,"student_isu":100001,"reviewer_id":5,"reviewers":[{"reviewer_id":5,"reviewer_surname":"Рецензентов"}],
		"review":null}],"total":1,"is_appointments_restricted":true}`)
	ap := must[*myitmo.GIAAppointmentList](t)(c.GIA.Appointments(ctx, myitmo.GIAAppointmentsParams{}))
	if !ap.IsAppointmentsRestricted || ap.Appointments[0].Reviewers[0].ReviewerID != 5 {
		t.Errorf("appointments = %+v", ap)
	}

	f.result(`{"students":[{"diploma_id":31,"student_isu":100001,"group_id":"Z3400","review":{"approved":true}}],"total":1}`)
	rs := must[*myitmo.GIAReviewedStudentList](t)(c.GIA.ReviewedStudents(ctx, myitmo.GIAReviewedStudentsParams{}))
	if rs.Total != 1 || rs.Students[0].DiplomaID != 31 || len(rs.Students[0].Review) == 0 {
		t.Errorf("students = %+v", rs)
	}

	f.result(`[{"student_id":11,"student_isu":100001,"student_surname":"Студентов","ep_name":"Программа","ep_enrollment_year":2021}]`)
	ss := must[[]myitmo.GIAReviewerStudent](t)(c.GIA.SearchReviewerStudents(ctx, "100001", 1, ""))
	if ss[0].StudentID != 11 || ss[0].EPEnrollmentYear != 2021 {
		t.Errorf("search = %+v", ss)
	}

	f.result(`{"edu_program_filters":[{"id":7,"name":"Программа"}],"edu_directions_filters":[],"educational_level_filters":[{"id":1,"name":"Бакалавриат"}],
		"group_filters":[{"group_id":"Z3400"}]}`)
	fl := must[*myitmo.GIAAppointmentFilters](t)(c.GIA.AppointmentFilters(ctx, ""))
	if fl.GroupFilters[0].GroupID != "Z3400" || fl.EducationalLevelFilters[0].Name != "Бакалавриат" {
		t.Errorf("filters = %+v", fl)
	}
}
