package myitmo_test

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/kewldan/go-itmo/myitmo"
)

// requestsRoute is one route assertion: run the call and compare method, path, query and body.
type requestsRoute struct {
	name   string
	result string
	run    func(c *myitmo.Client) error
	method string
	path   string
	query  string
	body   string
}

func runRequestsRoutes(t *testing.T, cases []requestsRoute) {
	t.Helper()
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			f, c := newFake(t)
			res := tc.result
			if res == "" {
				res = "null"
			}
			f.result(res)
			if err := tc.run(c); err != nil {
				t.Fatalf("call: %v", err)
			}
			r := f.expect(tc.method, tc.path)
			if got := r.Query.Encode(); got != tc.query {
				t.Errorf("query = %q, want %q", got, tc.query)
			}
			if tc.body != "" {
				r.sameJSON(t, tc.body)
			} else if len(r.Body) != 0 && tc.method != http.MethodPost {
				t.Errorf("unexpected body %q", r.Body)
			}
		})
	}
}

func requestsDiscard[T any](_ T, err error) error { return err }

func TestRequestsRoutes(t *testing.T) {
	values := []myitmo.RequestFieldValue{
		{FieldID: "101", FieldType: myitmo.RequestFieldDictionary, Value: "1,2"},
		{FieldID: "102", FieldType: myitmo.RequestFieldText, Value: "comment"},
	}
	runRequestsRoutes(t, []requestsRoute{
		{name: "Catalog", result: `[]`, method: "GET", path: "/api/requests/all",
			run: func(c *myitmo.Client) error { return requestsDiscard(c.Requests.Catalog(ctx)) }},
		{name: "Template", result: `{}`, method: "GET", path: "/api/requests/42",
			run: func(c *myitmo.Client) error { return requestsDiscard(c.Requests.Template(ctx, 42)) }},
		{name: "Dictionary", result: `[]`, method: "GET", path: "/api/requests/dict/7", query: "dep=3&field=101&q=",
			run: func(c *myitmo.Client) error { return requestsDiscard(c.Requests.Dictionary(ctx, "7", 101, "", "3")) }},
		{name: "FormUpdate", result: `{"show":[],"hide":[]}`, method: "POST", path: "/api/requests/form_update",
			body: `{"changed_field":101,"current_values":[{"field_id":"101","value":"1"}]}`,
			run: func(c *myitmo.Client) error {
				return requestsDiscard(c.Requests.FormUpdate(ctx, 101, []myitmo.RequestFieldValue{{FieldID: "101", Value: "1"}}))
			}},
		{name: "Send", result: `{"reqId":900}`, method: "POST", path: "/api/requests/send",
			body: `{"request_id":55,"values":[{"field_id":"101","field_type":"dictionary","value":"1,2"},{"field_id":"102","field_type":"text","value":"comment"}]}`,
			run:  func(c *myitmo.Client) error { return requestsDiscard(c.Requests.Send(ctx, 55, values)) }},
		{name: "List", result: `[]`, method: "GET", path: "/api/requests/my",
			run: func(c *myitmo.Client) error { return requestsDiscard(c.Requests.List(ctx)) }},
		{name: "Get", result: `{}`, method: "GET", path: "/api/requests/my/900",
			run: func(c *myitmo.Client) error { return requestsDiscard(c.Requests.Get(ctx, 900)) }},
		{name: "Delete", method: "DELETE", path: "/api/requests/my/900",
			run: func(c *myitmo.Client) error { return c.Requests.Delete(ctx, 900) }},
	})
}

func TestRequestsDownloads(t *testing.T) {
	cases := []struct {
		name string
		path string
		run  func(c *myitmo.Client) (*myitmo.File, error)
	}{
		{"TemplateFile", "/api/requests/file/15", func(c *myitmo.Client) (*myitmo.File, error) { return c.Requests.TemplateFile(ctx, "15") }},
		{"Download", "/api/requests/my/900/download", func(c *myitmo.Client) (*myitmo.File, error) { return c.Requests.Download(ctx, 900) }},
		{"AttachedFile", "/api/requests/files/my/abc%2F1", func(c *myitmo.Client) (*myitmo.File, error) {
			return c.Requests.AttachedFile(ctx, "abc/1")
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			f, c := newFake(t)
			f.replyWith(http.StatusOK, http.Header{"Content-Type": {"application/pdf"}, "Content-Disposition": {`attachment; filename="doc.pdf"`}}, "%PDF-1.4")
			file := must[*myitmo.File](t)(tc.run(c))
			defer file.Body.Close()
			f.expect(http.MethodGet, tc.path)
			b, _ := io.ReadAll(file.Body)
			if string(b) != "%PDF-1.4" || file.Name != "doc.pdf" {
				t.Errorf("file = %q %q", file.Name, b)
			}
		})
	}
}

func TestRequestsUpload(t *testing.T) {
	f, c := newFake(t)
	f.result(`{"name":"upl_123.pdf"}`)
	name := must[string](t)(c.Requests.Upload(ctx, myitmo.Upload{Field: "ignored", Name: "a.pdf", Content: strings.NewReader("data")}))
	r := f.expect(http.MethodPost, "/api/requests/upload")
	if name != "upl_123.pdf" {
		t.Errorf("name = %q", name)
	}
	body := string(r.Body)
	if !strings.Contains(r.Header.Get("Content-Type"), "multipart/form-data") || !strings.Contains(body, `name="file"; filename="a.pdf"`) || !strings.Contains(body, "data") {
		t.Errorf("multipart body = %q", body)
	}
}

func TestRequestsDecode(t *testing.T) {
	f, c := newFake(t)
	f.result(`[{"id":3,"name":"Учёба","color":5,"icon":"icon icon-other","requests":[{"id":42,"name":"Справка"}]},{"id":"x","name":"Другое","color":"default","icon":"","requests":[]}]`)
	cats := must[[]myitmo.RequestCategory](t)(c.Requests.Catalog(ctx))
	if len(cats) != 2 || cats[0].ID != "3" || cats[0].Color != "5" || cats[1].Color != "default" || cats[0].Requests[0].ID != 42 {
		t.Errorf("catalog = %+v", cats)
	}

	f.result(`{"template_description":"<p>x</p>","template_files":[{"file_id":15,"file_name":"blank.docx"}],"user_can_apply_now":true,"user_cannot_apply_now_reason":"",
		"fields_data":[{"field_id":101,"field_name":"Тип","field_note":"","field_type":"dictionary","field_size":"S","required_field_flag":true,"default_value":null,"multiple_choice":true,"dictionary_id":7,"dependent_field_id":null,"init_dictionary":{"id":1,"text":"Один"},"show_condition_flag":true},
		{"field_id":102,"field_name":"Таблица","field_type":"multiple_field","default_value":"","multiple_field_max_block":3,"multiple_field_data":[{"name":"Дата","type":"date","size":"s","required":false,"notice":"","dictionary":null,"dependsId":null}]}]}`)
	form := must[*myitmo.RequestForm](t)(c.Requests.Template(ctx, 42))
	if form.TemplateFiles[0].FileID != "15" || !form.UserCanApplyNow || len(form.Fields) != 2 {
		t.Fatalf("form = %+v", form)
	}
	fd := form.Fields[0]
	if fd.FieldType != myitmo.RequestFieldDictionary || fd.DictionaryID != "7" || fd.DependentFieldID != "" || fd.InitDictionary.ID != "1" || !fd.MultipleChoice {
		t.Errorf("field = %+v", fd)
	}
	if sf := form.Fields[1].MultipleFieldData[0]; sf.Type != myitmo.RequestFieldDate || form.Fields[1].MultipleFieldMaxBlock != 3 {
		t.Errorf("sub field = %+v", sf)
	}

	f.result(`{"show":[101,"102"],"hide":[]}`)
	upd := must[*myitmo.RequestFormUpdate](t)(c.Requests.FormUpdate(ctx, 101, nil))
	if len(upd.Show) != 2 || upd.Show[0] != "101" || upd.Show[1] != "102" {
		t.Errorf("update = %+v", upd)
	}

	f.result(`[{"id":1,"text":"Один"}]`)
	items := must[[]myitmo.RequestDictItem](t)(c.Requests.Dictionary(ctx, "7", 101, "од", ""))
	if len(items) != 1 || items[0].Text != "Один" {
		t.Errorf("items = %+v", items)
	}
	f.expect(http.MethodGet, "/api/requests/dict/7")

	f.result(`[{"id":900,"name":"Справка","notice":"","status":2,"status_name":"Отклонена","created_at":"2026-03-01T10:00:00+03:00","updated_at":"2026-03-02T11:30:00+03:00"}]`)
	list := must[[]myitmo.RequestSummary](t)(c.Requests.List(ctx))
	if list[0].Status != myitmo.RequestStatusRejected || list[0].UpdatedAt.Day() != 2 || list[0].CreatedAt.IsZero() {
		t.Errorf("list = %+v", list)
	}

	f.result(`{"id":900,"request_name":"Справка","request_status":"В работе","request_state":"Ожидает","request_status_tag":"processed","request_create_date":"2026-03-01 10:00:00",
		"agreement_list":[{"fio":"Тестов Т. Т.","context_name":"Деканат","decision":"Согласовано","decision_date":"02.03.2026"}],
		"entered_fields":[{"name":"Файл","type":"file","data":"55","file_name":"a.pdf"},{"name":"Таблица","type":"multiple_field","data":"","multiple_data":{"headers":["A","B"],"values":[["1","2"]]}}]}`)
	det := must[*myitmo.RequestDetail](t)(c.Requests.Get(ctx, 900))
	if det.RequestStatusTag != myitmo.RequestStatusTagProcessed || det.RequestCreateDate.Year() != 2026 || det.Approvals[0].Decision != "Согласовано" ||
		det.EnteredFields[0].Data != "55" || det.EnteredFields[1].MultipleData.Values[0][1] != "2" {
		t.Errorf("detail = %+v", det)
	}
}

func TestRequestsSendValidationError(t *testing.T) {
	f, c := newFake(t)
	f.reply(http.StatusBadRequest, `{"error_code":400,"error_message":"Проверьте поля","result":{"error_list":[
		{"id":"101","error_info":{"text":"Обязательное поле","obj_list":null}},
		{"id":102,"error_info":{"text":"","obj_list":[{"field_number":2,"part_number":1,"error_text":"Неверная дата"}]}}]}}`)
	_, err := c.Requests.Send(ctx, 55, nil)
	var ve *myitmo.RequestValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("err = %v, want *RequestValidationError", err)
	}
	if len(ve.Fields) != 2 || ve.Fields[0].ID != "101" || ve.Fields[0].Info.Text != "Обязательное поле" ||
		ve.Fields[1].ID != "102" || ve.Fields[1].Info.Cells[0].FieldNumber != 2 || ve.Fields[1].Info.Cells[0].ErrorText != "Неверная дата" {
		t.Errorf("fields = %+v", ve.Fields)
	}
	var e *myitmo.Error
	if !errors.As(err, &e) || e.StatusCode != http.StatusBadRequest || e.Message != "Проверьте поля" || myitmo.ErrorCode(err) != 400 {
		t.Errorf("error = %+v", e)
	}
	f.expect(http.MethodPost, "/api/requests/send")
}

func TestRequestsSendOutcomes(t *testing.T) {
	f, c := newFake(t)
	f.result(`{"reqId":901}`)
	if id := must[int64](t)(c.Requests.Send(ctx, 55, nil)); id != 901 {
		t.Errorf("id = %d", id)
	}

	f.result(`{"reqId":null}`)
	if _, err := c.Requests.Send(ctx, 55, nil); err == nil {
		t.Error("null reqId must fail")
	}

	f.reply(http.StatusOK, `{"error_code":5,"error_message":"Нельзя подать","result":null}`)
	_, err := c.Requests.Send(ctx, 55, nil)
	var ve *myitmo.RequestValidationError
	if errors.As(err, &ve) || myitmo.ErrorCode(err) != 5 {
		t.Errorf("err = %v", err)
	}
}

func TestRequestsDeleteTolerantBody(t *testing.T) {
	f, c := newFake(t)
	f.reply(http.StatusOK, `OK`)
	if err := c.Requests.Delete(ctx, 900); err != nil {
		t.Errorf("plain body: %v", err)
	}
	f.reply(http.StatusOK, `{"error_code":3,"error_message":"Нельзя удалить","result":null}`)
	if err := c.Requests.Delete(ctx, 900); myitmo.ErrorCode(err) != 3 {
		t.Errorf("envelope error: %v", err)
	}
	f.reply(http.StatusInternalServerError, `oops`)
	if err := c.Requests.Delete(ctx, 900); err == nil {
		t.Error("HTTP 500 must fail")
	}
}
