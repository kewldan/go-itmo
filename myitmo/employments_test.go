package myitmo_test

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/kewldan/go-itmo/myitmo"
)

func TestEmploymentsRoutes(t *testing.T) {
	ctx := context.Background()
	st := myitmo.EmploymentDocumentCheck
	cases := []struct {
		name   string
		call   func(*myitmo.Client) error
		method string
		path   string
		query  string
		body   string
	}{
		{"ListDefault", func(c *myitmo.Client) error {
			_, err := c.Employments.List(ctx, myitmo.EmploymentListParams{})
			return err
		}, "GET", "/api/employments/employments/", "limit=20&offset=0", ""},
		{"ListFiltered", func(c *myitmo.Client) error {
			_, err := c.Employments.List(ctx, myitmo.EmploymentListParams{Limit: 10, Offset: 30, Department: 5, Status: &st, Query: "ivan"})
			return err
		}, "GET", "/api/employments/employments/", "department=5&limit=10&offset=30&query=ivan&status=3", ""},
		{"ListStatusZero", func(c *myitmo.Client) error {
			_, err := c.Employments.List(ctx, myitmo.EmploymentListParams{Status: myitmo.Ptr(myitmo.EmploymentAwaitingApproval)})
			return err
		}, "GET", "/api/employments/employments/", "limit=20&offset=0&status=0", ""},
		{"Get", func(c *myitmo.Client) error { _, err := c.Employments.Get(ctx, 42); return err }, "GET", "/api/employments/employments/42", "", ""},
		{"Rights", func(c *myitmo.Client) error { _, err := c.Employments.Rights(ctx, 42); return err }, "GET", "/api/employments/employments/42/rights", "", ""},
		{"Approve", func(c *myitmo.Client) error { _, err := c.Employments.Approve(ctx, 42); return err }, "POST", "/api/employments/employments/42/approve", "", ""},
		{"Reject", func(c *myitmo.Client) error { _, err := c.Employments.Reject(ctx, 42, "нет"); return err }, "POST", "/api/employments/employments/42/reject", "", `{"comment":"нет"}`},
		{"RejectEmpty", func(c *myitmo.Client) error { _, err := c.Employments.Reject(ctx, 42, ""); return err }, "POST", "/api/employments/employments/42/reject", "", `{"comment":""}`},
		{"Stop", func(c *myitmo.Client) error { _, err := c.Employments.Stop(ctx, 42, 2, "отказ"); return err }, "POST", "/api/employments/employments/42/stop", "", `{"reasonId":2,"comment":"отказ"}`},
		{"Documents", func(c *myitmo.Client) error { _, err := c.Employments.Documents(ctx, 42); return err }, "GET", "/api/employments/employments/42/documents", "", ""},
		{"ApproveDocument", func(c *myitmo.Client) error {
			return c.Employments.ApproveDocument(ctx, 42, myitmo.EmploymentDocPassport)
		},
			"POST", "/api/employments/employments/42/documents/passport/approve", "", ""},
		{"RejectDocument", func(c *myitmo.Client) error {
			return c.Employments.RejectDocument(ctx, 42, myitmo.EmploymentDocCriminalRecord, "нечитаемо")
		}, "POST", "/api/employments/employments/42/documents/criminal_record/reject", "", `{"comment":"нечитаемо"}`},
		{"ApproveDocuments", func(c *myitmo.Client) error { _, err := c.Employments.ApproveDocuments(ctx, 42); return err },
			"POST", "/api/employments/employments/42/documents/approve", "", ""},
		{"ReturnDocuments", func(c *myitmo.Client) error {
			_, err := c.Employments.ReturnDocuments(ctx, 42, "исправить")
			return err
		},
			"POST", "/api/employments/employments/42/documents/reject", "", `{"comment":"исправить"}`},
		{"RejectDocuments", func(c *myitmo.Client) error { _, err := c.Employments.RejectDocuments(ctx, 42); return err },
			"POST", "/api/employments/employments/42/documents/reject", "", ""},
		{"SendDocuments", func(c *myitmo.Client) error { return c.Employments.SendDocuments(ctx, 42) },
			"POST", "/api/employments/employments/42/documents/send", "", ""},
		{"SignedDocuments", func(c *myitmo.Client) error { _, err := c.Employments.SignedDocuments(ctx, 42); return err },
			"GET", "/api/employments/employments/42/documents/signed", "", ""},
		{"DeleteSignedDocument", func(c *myitmo.Client) error {
			return c.Employments.DeleteSignedDocument(ctx, 42, myitmo.EmploymentSignedCivilServiceNotice)
		}, "DELETE", "/api/employments/employments/42/documents/signed/civil_service_notice", "", ""},
		{"Signers", func(c *myitmo.Client) error { _, err := c.Employments.Signers(ctx, 42); return err }, "GET", "/api/employments/employments/42/signers", "", ""},
		{"Briefing", func(c *myitmo.Client) error { _, err := c.Employments.Briefing(ctx, 42); return err }, "GET", "/api/employments/employments/42/briefing", "", ""},
		{"ApproveBriefing", func(c *myitmo.Client) error { return c.Employments.ApproveBriefing(ctx, 42) }, "POST", "/api/employments/employments/42/briefing/approve", "", ""},
		{"PaperWorkBook", func(c *myitmo.Client) error { _, err := c.Employments.PaperWorkBook(ctx, 42); return err }, "GET", "/api/employments/employments/42/paper_work_book", "", ""},
		{"ApprovePaperWorkBook", func(c *myitmo.Client) error { return c.Employments.ApprovePaperWorkBook(ctx, 42) },
			"POST", "/api/employments/employments/42/paper_work_book/approve", "", ""},
		{"Departments", func(c *myitmo.Client) error { _, err := c.Employments.Departments(ctx, "физ"); return err }, "GET", "/api/employments/references/departments", "query=%D1%84%D0%B8%D0%B7", ""},
		{"DepartmentsAll", func(c *myitmo.Client) error { _, err := c.Employments.Departments(ctx, ""); return err }, "GET", "/api/employments/references/departments", "", ""},
		{"StaffUnitsDefault", func(c *myitmo.Client) error { _, err := c.Employments.StaffUnits(ctx, 0, 0, 0); return err },
			"GET", "/api/employments/references/staff_units", "limit=100&offset=0", ""},
		{"StaffUnits", func(c *myitmo.Client) error { _, err := c.Employments.StaffUnits(ctx, 5, 10, 20); return err },
			"GET", "/api/employments/references/staff_units", "department=5&limit=10&offset=20", ""},
		{"StopReasons", func(c *myitmo.Client) error { _, err := c.Employments.StopReasons(ctx); return err }, "GET", "/api/employments/references/employment_stop_reasons", "", ""},
		{"FixedTermReasons", func(c *myitmo.Client) error { _, err := c.Employments.FixedTermReasons(ctx); return err }, "GET", "/api/employments/references/fixed_term_reasons", "", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			f, c := newFake(t)
			if err := tc.call(c); err != nil {
				t.Fatal(err)
			}
			r := f.expect(tc.method, tc.path)
			if got := r.Query.Encode(); got != tc.query {
				t.Errorf("query = %q, want %q", got, tc.query)
			}
			if tc.body != "" {
				r.sameJSON(t, tc.body)
			} else if len(r.Body) != 0 {
				t.Errorf("unexpected body %q", r.Body)
			}
		})
	}
}

func TestEmploymentsCreate(t *testing.T) {
	f, c := newFake(t)
	f.result(`{"employmentId":42,"status":0}`)
	e := must[*myitmo.Employment](t)(c.Employments.Create(context.Background(), myitmo.EmploymentCreateRequest{
		CandidateInfo: myitmo.EmploymentCandidateInput{
			LastName: "Петров", FirstName: "Пётр", Patronymic: "Петрович",
			CitizenshipCode: myitmo.EmploymentCitizenshipRussia, Phone: "+70000000000", Email: "p@example.com",
		},
		PositionInfo:        myitmo.EmploymentPositionInput{StaffUnit: 0.5, StaffUnitID: myitmo.Ptr[int64](900)},
		ProbationDurationID: myitmo.EmploymentProbation1Month,
		Format:              myitmo.EmploymentFormatInput{ID: myitmo.EmploymentFormatOnsite, FromRussia: true},
		EmploymentTypeID:    myitmo.EmploymentTypeMain,
		Schedule:            "Пн, Вт, Ср",
		WorkingHoursTypeID:  myitmo.EmploymentHoursStandard,
		ContractDateStart:   myitmo.NewDate(2026, 10, 1),
		ContractDateEnd:     myitmo.NewDate(2027, 6, 30),
		FixedTermReasonID:   4,
	}))
	if e.Key() != 42 || e.Status != myitmo.EmploymentAwaitingApproval {
		t.Fatalf("employment = %+v", e)
	}
	f.expect("POST", "/api/employments/employments/").sameJSON(t, `{
		"candidateInfo":{"lastName":"Петров","firstName":"Пётр","patronymic":"Петрович","citizenshipCode":643,"phone":"+70000000000","email":"p@example.com"},
		"positionInfo":{"staffUnit":0.5,"staffUnitId":900},
		"probationDurationId":2,"format":{"id":1,"fromRussia":true},"employmentTypeId":1,"schedule":"Пн, Вт, Ср",
		"workingHoursTypeId":1,"contractDateStart":"2026-10-01","contractDateEnd":"2027-06-30","fixedTermReasonId":4}`)

	f.result(`{"employmentID":43}`)
	e = must[*myitmo.Employment](t)(c.Employments.Create(context.Background(), myitmo.EmploymentCreateRequest{}))
	if e.Key() != 43 {
		t.Fatalf("key = %d", e.Key())
	}
	f.last().sameJSON(t, `{"candidateInfo":{"lastName":"","firstName":"","patronymic":"","citizenshipCode":0,"phone":"","email":""},
		"positionInfo":{"staffUnit":0,"staffUnitId":null},"probationDurationId":0,"format":{"id":0,"fromRussia":false},
		"employmentTypeId":0,"schedule":"","workingHoursTypeId":0,"contractDateStart":""}`)
}

func TestEmploymentsUploadSigned(t *testing.T) {
	f, c := newFake(t)
	err := c.Employments.UploadSignedDocument(context.Background(), 42, myitmo.EmploymentSignedContract,
		myitmo.Upload{Name: "contract.pdf", ContentType: "application/pdf", Content: strings.NewReader("PDF")})
	if err != nil {
		t.Fatal(err)
	}
	fields, files := formOf(t, f.expect("POST", "/api/employments/employments/42/documents/signed/upload"))
	if fields["documentType"] != "contract" || files["file"] != "contract.pdf:PDF" {
		t.Fatalf("fields = %v, files = %v", fields, files)
	}
}

func TestEmploymentsDownloads(t *testing.T) {
	ctx := context.Background()
	cases := []struct {
		name   string
		call   func(*myitmo.Client) (*myitmo.File, error)
		method string
		path   string
		query  string
		body   string
	}{
		{"Contract", func(c *myitmo.Client) (*myitmo.File, error) { return c.Employments.Contract(ctx, 42, "kc-1") },
			"POST", "/api/employments/employments/42/contract", "", `{"signerKeycloakId":"kc-1"}`},
		{"ContractNoSigner", func(c *myitmo.Client) (*myitmo.File, error) { return c.Employments.Contract(ctx, 42, "") },
			"POST", "/api/employments/employments/42/contract", "", ""},
		{"ContractPreview", func(c *myitmo.Client) (*myitmo.File, error) { return c.Employments.ContractPreview(ctx, 42, "kc-1") },
			"POST", "/api/employments/employments/42/contract/file", "", `{"signerKeycloakId":"kc-1"}`},
		{"StaticDocument", func(c *myitmo.Client) (*myitmo.File, error) {
			return c.Employments.StaticDocument(ctx, "/abc/scan 1.jpg?sig=x%2By&exp=1")
		}, "GET", "/api/employments/static/documents/abc/scan%201.jpg", "exp=1&sig=x%2By", ""},
		{"V1File", func(c *myitmo.Client) (*myitmo.File, error) { return c.Employments.V1File(ctx, "files/7/book.xml") },
			"GET", "/api/employments/api/v1/files/7/book.xml", "", ""},
		{"DocumentFileStatic", func(c *myitmo.Client) (*myitmo.File, error) {
			return c.Employments.DocumentFile(ctx, "https://backend.example.com/employment/api/v1/static/documents/u1/pass.png?t=1")
		}, "GET", "/api/employments/static/documents/u1/pass.png", "t=1", ""},
		{"DocumentFileV1", func(c *myitmo.Client) (*myitmo.File, error) {
			return c.Employments.DocumentFile(ctx, "https://backend.example.com/x/api/v1/work-books/9/pdf")
		}, "GET", "/api/employments/api/v1/work-books/9/pdf", "", ""},
		{"DocumentFileProxyPath", func(c *myitmo.Client) (*myitmo.File, error) {
			return c.Employments.DocumentFile(ctx, "/api/employments/static/documents/u1/a.pdf")
		}, "GET", "/api/employments/static/documents/u1/a.pdf", "", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			f, c := newFake(t)
			f.replyWith(http.StatusOK, http.Header{"Content-Type": {"application/pdf"}}, "BYTES")
			file := must[*myitmo.File](t)(tc.call(c))
			defer file.Body.Close()
			if b, _ := io.ReadAll(file.Body); string(b) != "BYTES" || file.ContentType != "application/pdf" {
				t.Fatalf("file = %q %q", file.ContentType, b)
			}
			r := f.expect(tc.method, tc.path)
			if got := r.Query.Encode(); got != tc.query {
				t.Errorf("query = %q, want %q", got, tc.query)
			}
			if tc.body != "" {
				r.sameJSON(t, tc.body)
			} else if len(r.Body) != 0 {
				t.Errorf("unexpected body %q", r.Body)
			}
		})
	}
}

func TestEmploymentsFilePathRejected(t *testing.T) {
	ctx := context.Background()
	_, c := newFake(t)
	for _, p := range []string{"", "../../intro/json", "a/../../b", "https://evil.example.com/x"} {
		if _, err := c.Employments.StaticDocument(ctx, p); err == nil {
			t.Errorf("StaticDocument(%q) succeeded", p)
		}
	}
	for _, u := range []string{"relative/path", "https://example.com/other/file"} {
		if _, err := c.Employments.DocumentFile(ctx, u); err == nil {
			t.Errorf("DocumentFile(%q) succeeded", u)
		}
	}
}

func TestEmploymentsDecode(t *testing.T) {
	ctx := context.Background()
	f, c := newFake(t)

	f.result(`{"employments":[{"employmentId":42,"status":3,"briefingStatus":null,"stopReason":null,"rejectComment":null,
		"candidateInfo":{"lastName":"Петров","firstName":"Пётр","patronymic":"Петрович","personId":null,"phone":"+70000000000","email":"p@example.com","citizenship":{"code":643}},
		"contact":{"personId":100003,"name":"Контакт","phone":"","email":"c@example.com"},
		"positionInfo":{"position":{"name":"Инженер"},"department":{"shortName":"ФПИ","name":"Факультет","departmentCode":"F1"},"staffUnit":0.25,"category":"ППС"},
		"format":{"id":2,"name":"Дистанционно","fromRussia":false},"employmentType":{"name":"Основное"},
		"probationDuration":null,"fixedTermReason":null,"contractDateStart":"2026-10-01","contractDateEnd":null,
		"schedule":"Пн, Вт","isInoagent":false,"isDisqualified":true}],"count":1}`)
	list := must[*myitmo.EmploymentList](t)(c.Employments.List(ctx, myitmo.EmploymentListParams{}))
	e := list.Employments[0]
	if list.Count != 1 || e.Status != myitmo.EmploymentDocumentCheck || e.BriefingStatus != nil || e.CandidateInfo.PersonID != nil ||
		e.PositionInfo.StaffUnit != 0.25 || e.PositionInfo.Department.ShortName != "ФПИ" || e.Format.ID != myitmo.EmploymentFormatRemote ||
		e.ProbationDuration != nil || !e.ContractDateEnd.IsZero() || e.ContractDateStart != myitmo.NewDate(2026, 10, 1) || !e.IsDisqualified {
		t.Fatalf("employment = %+v", e)
	}

	f.result(`{"status":-2,"stopReason":{"name":"Отказ кандидата","comment":""},"briefingStatus":3}`)
	e2 := must[*myitmo.Employment](t)(c.Employments.Stop(ctx, 42, 1, ""))
	if e2.Status != myitmo.EmploymentStopped || e2.StopReason.Name == "" || *e2.BriefingStatus != myitmo.EmploymentDocumentApproved {
		t.Fatalf("stopped = %+v", e2)
	}

	f.result(`{"isLead":true,"isHR":false}`)
	if r := must[*myitmo.EmploymentRights](t)(c.Employments.Rights(ctx, 42)); !r.IsLead || r.IsHR {
		t.Fatalf("rights = %+v", r)
	}

	f.result(`[{"type":"passport","status":2,"uploadedAt":"2026-09-01T10:00:00+03:00","comment":null,
		"data":{"series":"0000","number":"000000","firstPageScanUrl":"https://backend.example.com/api/v1/static/documents/x.png"}},
		{"type":"snils","status":0,"uploadedAt":null,"comment":null,"data":null}]`)
	docs := must[[]myitmo.EmploymentDocument](t)(c.Employments.Documents(ctx, 42))
	if docs[0].Type != myitmo.EmploymentDocPassport || docs[0].Status != myitmo.EmploymentDocumentOnCheck || docs[0].UploadedAt == nil ||
		!strings.Contains(string(docs[0].Data), "firstPageScanUrl") || docs[1].UploadedAt != nil {
		t.Fatalf("documents = %+v", docs)
	}

	f.result(`{"documents":[{"type":"contract","status":"signed","fileUrl":"https://backend.example.com/api/v1/f/1","fileName":"c.pdf","fileSize":1024},
		{"type":"civil_service_notice","status":null,"fileUrl":null,"fileName":null}],"employmentStatus":6}`)
	signed := must[*myitmo.EmploymentSignedDocuments](t)(c.Employments.SignedDocuments(ctx, 42))
	if signed.EmploymentStatus != myitmo.EmploymentOPSSigned || signed.Documents[0].FileSize != 1024 || signed.Documents[1].Type != myitmo.EmploymentSignedCivilServiceNotice {
		t.Fatalf("signed = %+v", signed)
	}

	f.result(`[{"keycloakId":"kc-1","fullName":"Подписант","proxyDateTo":"2027-01-31"},{"keycloakId":"kc-2","fullName":"Другой","proxyDateTo":null}]`)
	signers := must[[]myitmo.EmploymentSigner](t)(c.Employments.Signers(ctx, 42))
	if signers[0].ProxyDateTo != myitmo.NewDate(2027, 1, 31) || !signers[1].ProxyDateTo.IsZero() {
		t.Fatalf("signers = %+v", signers)
	}

	f.result(`{"status":2}`)
	if b := must[*myitmo.EmploymentStepStatus](t)(c.Employments.PaperWorkBook(ctx, 42)); b.Status != myitmo.EmploymentDocumentOnCheck {
		t.Fatalf("paper work book = %+v", b)
	}

	f.result(`{"departments":[{"id":5,"name":"Факультет","shortName":null}]}`)
	if d := must[[]myitmo.EmploymentDepartment](t)(c.Employments.Departments(ctx, "")); len(d) != 1 || d[0].ID != 5 {
		t.Fatalf("departments = %+v", d)
	}

	f.result(`{"staffUnits":[{"staffUnitId":900,"position":{"name":"Инженер"},"department":{"name":"Факультет","shortName":"ФПИ"},
		"category":"АУП","unitsFree":1.5,"staffUnitAvailableBefore":null}]}`)
	if u := must[[]myitmo.EmploymentStaffUnit](t)(c.Employments.StaffUnits(ctx, 5, 0, 0)); u[0].StaffUnitID != 900 || u[0].UnitsFree != 1.5 {
		t.Fatalf("staff units = %+v", u)
	}

	f.result(`[{"id":1,"name":"Отказ кандидата"}]`)
	if r := must[[]myitmo.EmploymentStopReason](t)(c.Employments.StopReasons(ctx)); r[0].ID != 1 {
		t.Fatalf("stop reasons = %+v", r)
	}

	f.result(`[{"id":4,"name":"Замещение","requiresAbsentEmployee":true}]`)
	if r := must[[]myitmo.EmploymentFixedTermReason](t)(c.Employments.FixedTermReasons(ctx)); !r[0].RequiresAbsentEmployee {
		t.Fatalf("fixed-term reasons = %+v", r)
	}
}
