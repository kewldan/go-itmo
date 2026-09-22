package myitmo_test

import (
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/kewldan/go-itmo/myitmo"
)

func TestGIAOrderRequests(t *testing.T) {
	in := myitmo.GIAOrderInput{FormDate: myitmo.NewDate(2025, time.May, 15), OrderTypeID: 2, StudyYear: 2025,
		EducationLevelID: 1, ImplementerID: 3, Students: []int64{11, 12}}
	runGIACases(t, []giaCase{
		{"Orders", func(c *myitmo.Client) error {
			_, err := c.GIA.Orders(ctx, myitmo.GIAOrdersParams{Limit: 15, Query: "тест", OrderTypeID: 2, ImplementerID: 3,
				EducationLevelID: 1, StatusID: 20, NeedApprove: true})
			return err
		}, "GET", "/api/gia/orders/list", url.Values{"limit": {"15"}, "offset": {"0"}, "query": {"тест"}, "order_type_id": {"2"},
			"implementer_id": {"3"}, "education_level_id": {"1"}, "status_id": {"20"}, "need_approve": {"1"}}, ""},
		{"CreateOrder", func(c *myitmo.Client) error { _, err := c.GIA.CreateOrder(ctx, in); return err },
			"POST", "/api/gia/orders/", nil, `{"form_date":"2025-05-15","order_type_id":2,"study_year":2025,"education_level_id":1,
			"implementer_id":3,"students":[11,12]}`},
		{"Order", func(c *myitmo.Client) error { _, err := c.GIA.Order(ctx, 51); return err },
			"GET", "/api/gia/orders/51", nil, ""},
		{"UpdateOrderStudents", func(c *myitmo.Client) error { return c.GIA.UpdateOrderStudents(ctx, 51, 11) },
			"PATCH", "/api/gia/orders/51", nil, `{"students":[11]}`},
		{"ApproveOrder", func(c *myitmo.Client) error { return c.GIA.ApproveOrder(ctx, 51) },
			"POST", "/api/gia/orders/51/approve", nil, `{}`},
		{"DeclineOrder", func(c *myitmo.Client) error { return c.GIA.DeclineOrder(ctx, 51, "ошибка") },
			"POST", "/api/gia/orders/51/decline", nil, `{"comment":"ошибка"}`},
		{"DeclineOrderEmpty", func(c *myitmo.Client) error { return c.GIA.DeclineOrder(ctx, 51, "") },
			"POST", "/api/gia/orders/51/decline", nil, `{}`},
		{"OrderCandidates", func(c *myitmo.Client) error {
			_, err := c.GIA.OrderCandidates(ctx, myitmo.GIAOrderCandidatesParams{Year: 2025, ImplementerID: 3, EducationLevelID: 1, TypeID: 2})
			return err
		}, "GET", "/api/gia/orders/students/list", url.Values{"year": {"2025"}, "implementer_id": {"3"}, "education_level_id": {"1"}, "type_id": {"2"}}, ""},
	})
	runGIADownloads(t, []giaDownload{
		{"OrderFile", func(c *myitmo.Client) (*myitmo.File, error) { return c.GIA.OrderFile(ctx, 51) },
			"GET", "/api/gia/orders/51/file", url.Values{"format": {"pdf"}}, ""},
	})
}

func TestGIAOrderDecodes(t *testing.T) {
	f, c := newFake(t)
	f.result(`{"result":[{"order_id":51,"number":"12-ГИА","order_type":{"id":2,"name":"О рецензентах"},"education_level":{"name":"Бакалавриат"},
		"implementer":{"faculty_short_name":"ФИТ"},"contingent":null,"form_date":"2025-05-15","approve_date":null,
		"created_by":{"fio":"Автор"},"status":{"status_id":20,"status_name":"Ожидает согласования"},"scan_file_link":""}],"count":7}`)
	list := must[*myitmo.GIAOrderList](t)(c.GIA.Orders(ctx, myitmo.GIAOrdersParams{}))
	if list.Count != 7 || list.Result[0].Status.StatusID != myitmo.GIAOrderStatusAwaitingApproval || list.Result[0].OrderType.ID != 2 ||
		!list.Result[0].ApproveDate.IsZero() {
		t.Errorf("orders = %+v", list)
	}

	detail := `{"id":51,"order_type":{"name":"О рецензентах"},"form_date":"2025-05-15","updated_at":"2025-05-16T10:00:00",
		"created_by":{"fio":"Автор","isu":100004},"status":{"status_id":20,"status_name":"Ожидает"},
		"signers":[{"fio":"Подписант","isu":100005,"job_title":"декан","status":{"status_id":1,"status_name":"Ожидает"},"created_at":null}],
		"comments":[{"comment":"проверить","created_at":"2025-05-16T11:00:00","created_by":{"fio":"Подписант","isu":100005}}],
		"attachment":%s}`
	block := `{"education_level":{"name":"Бакалавриат"},"implementer":{"faculty_short_name":"ФИТ","faculty":"Факультет"},
		"direction":{"dir_code":"09.03.01","dir_name":"Информатика"},"ep":{"ep_name":"Программа"},"group_id":"Z3400",
		"students":[{"student_id":11,"fio":"Студентов С.С.","isu":100001,"theme":"Тема","supervisor":"Руководителев Р.Р.","consultants":[],
		"reviewers":[{"fio":"Рецензентов Р.Р.","academic_title":"доцент","job_title":"инженер","work_place":"ООО Тест","degree":"к.т.н."}]}]}`
	for _, att := range []string{block, "[" + block + "]"} {
		f.result(giaf(detail, att))
		o := must[*myitmo.GIAOrder](t)(c.GIA.Order(ctx, 51))
		blocks, err := o.Attachments()
		if err != nil || len(blocks) != 1 || blocks[0].Implementer.Faculty != "Факультет" || blocks[0].Students[0].Reviewers[0].Degree != "к.т.н." ||
			o.Signers[0].ISU != 100005 || o.Comments[0].CreatedBy.ISU != 100005 {
			t.Errorf("order = %+v, blocks = %+v, err = %v", o, blocks, err)
		}
	}

	f.result(`{"order_id":52}`)
	res := must[myitmo.RawJSON](t)(c.GIA.CreateOrder(ctx, myitmo.GIAOrderInput{}))
	if string(res) != `{"order_id":52}` {
		t.Errorf("create = %s", res)
	}

	f.reply(http.StatusOK, `{"error_code":3,"error_message":"Студенты уже в приказе","result":null}`)
	if _, err := c.GIA.CreateOrder(ctx, myitmo.GIAOrderInput{}); myitmo.ErrorCode(err) != 3 {
		t.Errorf("err = %v", err)
	}

	f.result(`[{"student_id":11,"fio":"Студентов С.С.","isu":100001,"group_id":"Z3400","theme":"Тема","supervisor":"Руководителев",
		"consultant":"","reviewers":[],"direction":{"dir_code":"09.03.01","dir_name":"Информатика"}}]`)
	cand := must[[]myitmo.GIAOrderCandidate](t)(c.GIA.OrderCandidates(ctx, myitmo.GIAOrderCandidatesParams{}))
	if cand[0].Direction.DirCode != "09.03.01" {
		t.Errorf("candidates = %+v", cand)
	}
}
