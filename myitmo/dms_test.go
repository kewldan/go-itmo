package myitmo_test

import (
	"bytes"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"strings"
	"testing"

	"github.com/kewldan/go-itmo/myitmo"
)

func TestDMSRoutes(t *testing.T) {
	tests := []struct {
		name   string
		call   func(c *myitmo.Client) error
		method string
		path   string
		query  string
		body   string
	}{
		{"Access", func(c *myitmo.Client) error { _, err := c.DMS.Access(ctx); return err }, "GET", "/api/dms/me/access", "", ""},
		{"MyRequest", func(c *myitmo.Client) error { _, err := c.DMS.MyRequest(ctx); return err }, "GET", "/api/dms/me/request", "", ""},
		{"Contacts", func(c *myitmo.Client) error { _, err := c.DMS.Contacts(ctx); return err }, "GET", "/api/dms/me/request/contacts", "", ""},
		{"Sign", func(c *myitmo.Client) error {
			_, err := c.DMS.Sign(ctx, 12, "Тестов", "Тест", "")
			return err
		},
			"POST", "/api/dms/me/request/12/sign", "", `{"last_name":"Тестов","first_name":"Тест","parental_name":""}`},
		{"Reject", func(c *myitmo.Client) error { return c.DMS.Reject(ctx, 12) }, "POST", "/api/dms/me/request/12/reject", "", ""},
		{"CreateCorpMail", func(c *myitmo.Client) error { return c.DMS.CreateCorpMail(ctx, "test.user", "Example-pass1!") },
			"POST", "/api/dms/me/corp_mail", "", `{"username":"test.user","password":"Example-pass1!"}`},
		{"Citizenships", func(c *myitmo.Client) error { _, err := c.DMS.Citizenships(ctx); return err }, "GET", "/api/dms/citizenships", "", ""},
		{"StatusFilters", func(c *myitmo.Client) error { _, err := c.DMS.StatusFilters(ctx); return err }, "GET", "/api/dms/filters/statuses", "", ""},
		{"ListDefault", func(c *myitmo.Client) error {
			_, err := c.DMS.List(ctx, myitmo.DMSListParams{Limit: 20})
			return err
		}, "GET", "/api/dms/request/list", "approval_needed=false&limit=20&offset=0", ""},
		{"ListFiltered", func(c *myitmo.Client) error {
			_, err := c.DMS.List(ctx, myitmo.DMSListParams{Limit: 20, Offset: 40, StatusID: myitmo.Ptr[int64](2),
				SortBy: "full_name,last_update", SortOrder: "asc,desc", ApprovalNeeded: true, Query: "Тестов"})
			return err
		}, "GET", "/api/dms/request/list",
			"approval_needed=true&limit=20&offset=40&query=%D0%A2%D0%B5%D1%81%D1%82%D0%BE%D0%B2&sort_by=full_name%2Clast_update&sort_order=asc%2Cdesc&status_id=2", ""},
		{"Details", func(c *myitmo.Client) error { _, err := c.DMS.Details(ctx, 12); return err }, "GET", "/api/dms/request/list/12", "", ""},
		{"Approve", func(c *myitmo.Client) error { return c.DMS.Decide(ctx, 12, myitmo.DMSStatusApproved, "") },
			"POST", "/api/dms/request/me/12/decision", "", `{"request_id":12,"status_id":3}`},
		{"Return", func(c *myitmo.Client) error {
			return c.DMS.Decide(ctx, 12, myitmo.DMSStatusReturned, "Нет скана")
		},
			"POST", "/api/dms/request/me/12/decision", "", `{"request_id":12,"status_id":4,"comment":"Нет скана"}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, c := newFake(t)
			if err := tt.call(c); err != nil {
				t.Fatal(err)
			}
			r := f.expect(tt.method, tt.path)
			if got := r.Query.Encode(); got != tt.query {
				t.Errorf("query = %q, want %q", got, tt.query)
			}
			if tt.body != "" {
				r.sameJSON(t, tt.body)
			} else if len(r.Body) != 0 {
				t.Errorf("unexpected body %s", r.Body)
			}
		})
	}
}

func TestDMSSubmitPersonalData(t *testing.T) {
	f, c := newFake(t)
	err := c.DMS.SubmitPersonalData(ctx, 12, myitmo.DMSPersonalData{
		LastName: "Тестов", FirstName: "Тест",
		BirthDate: myitmo.NewDate(1990, 1, 2), Gender: myitmo.DMSGenderMale, CitizenshipID: myitmo.DMSCitizenshipRussia,
		PassportSeries: "0000", PassportNumber: "000000", PassportIssueDate: myitmo.NewDate(2010, 3, 4),
		PassportIssueDep: "000-000", PassportIssuedBy: "Отдел", ResidenceFactAddress: "ул. Примерная, 1",
		MobileNumber: "+70000000000", CorpMail: "test@itmo.ru",
	},
		myitmo.Upload{Field: "ignored", Name: "a.pdf", ContentType: "application/pdf", Content: strings.NewReader("%PDF-a")},
		myitmo.Upload{Name: "b.pdf", ContentType: "application/pdf", Content: strings.NewReader("%PDF-b")},
	)
	if err != nil {
		t.Fatal(err)
	}
	r := f.expect(http.MethodPost, "/api/dms/me/request/12")
	_, params, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil {
		t.Fatal(err)
	}
	mr := multipart.NewReader(bytes.NewReader(r.Body), params["boundary"])
	fields := map[string]string{}
	var files []string
	for {
		p, err := mr.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		data, _ := io.ReadAll(p)
		if p.FileName() != "" {
			if p.FormName() != "files" {
				t.Errorf("file part name = %q", p.FormName())
			}
			files = append(files, p.FileName()+"="+string(data))
			continue
		}
		fields[p.FormName()] = string(data)
	}
	want := map[string]string{
		"last_name": "Тестов", "first_name": "Тест", "birth_date": "1990-01-02", "gender": "M", "citizenship_id": "1",
		"passport_series": "0000", "passport_number": "000000", "passport_issue_date": "2010-03-04",
		"passport_issue_dep": "000-000", "passport_issued_by": "Отдел", "residence_fact_address": "ул. Примерная, 1",
		"residence_apartment": "1", "mobile_number": "+70000000000", "corp_mail": "test@itmo.ru",
	}
	if len(fields) != len(want) {
		t.Errorf("fields = %v", fields)
	}
	for k, v := range want {
		if fields[k] != v {
			t.Errorf("field %s = %q, want %q", k, fields[k], v)
		}
	}
	if len(files) != 2 || files[0] != "a.pdf=%PDF-a" || files[1] != "b.pdf=%PDF-b" {
		t.Errorf("files = %v", files)
	}
}

func TestDMSFile(t *testing.T) {
	f, c := newFake(t)
	f.replyWith(http.StatusOK, http.Header{"Content-Type": {"application/pdf"}}, "%PDF-1.4")
	file := must[*myitmo.File](t)(c.DMS.File(ctx, 12, "scan 1.pdf"))
	defer file.Body.Close()
	f.expect(http.MethodGet, "/api/dms/request/12/files/scan%201.pdf")
	if data, _ := io.ReadAll(file.Body); string(data) != "%PDF-1.4" {
		t.Errorf("body = %q", data)
	}
}

func TestDMSDecode(t *testing.T) {
	f, c := newFake(t)
	f.result(`{"request_id":12,"req_status":4,"last_name":"Тестов","first_name":"Тест","parental_name":null,"birth_date":"1990-01-02",
		"gender":"W","citizenship":"Россия","citizenship_id":1,"passport_series":"0000","passport_number":"000000","passport_issue_date":"2010-03-04",
		"passport_issue_dep":"000-000","passport_issued_by":"Отдел","residence_fact_address":"ул. Примерная, 1","mobile_number":"+70000000000",
		"corp_mail":null,"comment":"Нет скана","comment_author_isu":100001,"comment_author_surname":"Кадров","comment_author_name":"Кадр",
		"comment_author_second_name":null,"comment_created_at":"2026-05-01T12:00:00+03:00"}`)
	req := must[*myitmo.DMSRequest](t)(c.DMS.MyRequest(ctx))
	if req.ReqStatus != myitmo.DMSStatusReturned || req.Gender != myitmo.DMSGenderFemale || req.BirthDate != myitmo.NewDate(1990, 1, 2) ||
		req.CorpMail != "" || *req.CommentAuthorISU != 100001 || req.CommentCreatedAt == nil {
		t.Errorf("request = %+v", req)
	}
	f.result(`null`)
	if req := must[*myitmo.DMSRequest](t)(c.DMS.MyRequest(ctx)); req != nil {
		t.Errorf("null request = %+v", req)
	}

	f.result(`true`)
	if !must[bool](t)(c.DMS.Access(ctx)) {
		t.Error("access = false")
	}

	f.result(`{"university_support":{"contact_link":"https://example.test/u","contact_support_email":"u@example.test","contact_phone_number":"+70000000001"},
		"DMS_support":{"contact_link":"https://example.test/d","contact_email":"d@example.test"},
		"insurance_support":{"insurance_company":"Страховая","insurance_company_contact_number":"8-800-000-00-00","contact_support_email":"doc@example.test"},
		"information_link":"https://example.test/info",
		"insurance_data":{"program_name":"Базовая","insurance_company":"Страховая","insurance_company_link":"https://example.test/i",
		"insurance_company_contact_number":"8-800-000-00-00","qr_codes":[{"qr_code":"iVBORw0KGgo=","qr_code_desc":"Приложение"}]}}`)
	ct := must[*myitmo.DMSContacts](t)(c.DMS.Contacts(ctx))
	if ct.DMSSupport.ContactEmail != "d@example.test" || ct.InsuranceSupport.ContactSupportEmail != "doc@example.test" ||
		len(ct.InsuranceData.QRCodes) != 1 || ct.UniversitySupport.ContactPhoneNumber == "" {
		t.Errorf("contacts = %+v", ct)
	}

	f.result(`[{"signature_id":901,"task_id":"t-1"}]`)
	tasks := must[[]myitmo.DMSSignTask](t)(c.DMS.Sign(ctx, 12, "Тестов", "Тест", ""))
	if len(tasks) != 1 || tasks[0].SignatureID != "901" || tasks[0].TaskID != "t-1" {
		t.Errorf("tasks = %+v", tasks)
	}

	f.result(`[{"id":1,"name":"Россия"},{"id":2,"name":"Беларусь"}]`)
	if cs := must[[]myitmo.DMSDictItem](t)(c.DMS.Citizenships(ctx)); len(cs) != 2 || cs[0].ID != myitmo.DMSCitizenshipRussia {
		t.Errorf("citizenships = %+v", cs)
	}
	f.result(`[{"id":2,"name":"На согласовании"}]`)
	if st := must[[]myitmo.DMSDictItem](t)(c.DMS.StatusFilters(ctx)); len(st) != 1 || st[0].ID != myitmo.DMSStatusOnApproval {
		t.Errorf("statuses = %+v", st)
	}

	f.result(`{"count":41,"on_approval_count":3,"requests":[{"request_id":12,"pers_id":100002,"full_name":"Тестов Тест","corp_mail":null,
		"last_update":"01.05.2026","status_id":2,"status_name":"На согласовании"}]}`)
	list := must[*myitmo.DMSRequestList](t)(c.DMS.List(ctx, myitmo.DMSListParams{Limit: 20}))
	if list.Count != 41 || list.OnApprovalCount != 3 || len(list.Requests) != 1 || list.Requests[0].LastUpdate != "01.05.2026" {
		t.Errorf("list = %+v", list)
	}

	f.result(`{"request_id":12,"pers_id":100002,"last_name":"Тестов","first_name":"Тест","parental_name":"","birth_date":"02.01.1990","gender":"M",
		"citizenship":"Россия","passport_series":"0000","passport_number":"000000","passport_issue_date":"04.03.2010","passport_issue_dep":"000-000",
		"passport_issued_by":"Отдел","residence_fact_address":"ул. Примерная, 1","mobile_number":"+70000000000","corp_mail":"test@itmo.ru",
		"user_file_name":"scan.pdf","status_id":2,"status_name":"На согласовании",
		"rejections":[{"comment":"Нет скана","commenter_pers_id":100001,"commenter_full_name":"Кадров Кадр","commenter_date":"30.04.2026","commenter_time":null}]}`)
	d := must[*myitmo.DMSRequestDetails](t)(c.DMS.Details(ctx, 12))
	if d.StatusID != myitmo.DMSStatusOnApproval || d.UserFileName != "scan.pdf" || len(d.Rejections) != 1 || d.Rejections[0].CommenterPersID != 100001 {
		t.Errorf("details = %+v", d)
	}
}
