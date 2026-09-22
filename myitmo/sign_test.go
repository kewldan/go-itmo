package myitmo_test

import (
	"context"
	"io"
	"net/http"
	"testing"

	"github.com/kewldan/go-itmo/myitmo"
)

func TestSignTask(t *testing.T) {
	f, c := newFake(t)
	f.result(`{"taskId":"task-1","name":"Consent","type":"pdf","status":"SIGN_IN_PROGRESS",
		"signatures":[{"signatureId":"sig-1","status":"PENDING","type":"nonQualified","rejectable":true},
		{"signatureId":"sig-2","status":"SIGNED","type":"simple","rejectable":false}]}`)
	task := must[*myitmo.SignTask](t)(c.Sign.Task(t.Context(), "task/1"))
	f.expect("GET", "/api/sign/tasks/task%2F1")
	if task.TaskID != "task-1" || task.Type != "pdf" || task.Status != myitmo.SignTaskInProgress || !task.Signable() {
		t.Fatalf("task = %+v", task)
	}
	if s := task.Signatures[0]; s.SignatureID != "sig-1" || s.Type != myitmo.SignTypeNonQualified || !s.Rejectable {
		t.Fatalf("signature = %+v", s)
	}
}

func TestSignFiles(t *testing.T) {
	for _, tc := range []struct {
		name string
		path string
	}{
		{"source", "/api/sign/tasks/task-1/files/source"},
		{"stamped", "/api/sign/tasks/task-1/files/stamped"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f, c := newFake(t)
			f.replyWith(http.StatusOK, http.Header{"Content-Type": {"application/pdf"}}, "%PDF-1.7")
			var fn func(context.Context, string) (*myitmo.File, error)
			switch tc.name {
			case "source":
				fn = c.Sign.SourceFile
			default:
				fn = c.Sign.StampedFile
			}
			file := must[*myitmo.File](t)(fn(t.Context(), "task-1"))
			defer file.Body.Close()
			f.expect("GET", tc.path)
			b, _ := io.ReadAll(file.Body)
			if string(b) != "%PDF-1.7" {
				t.Fatalf("body = %q", b)
			}
		})
	}
}

var signPairs = []myitmo.SignTaskSignature{{TaskID: "task-1", SignatureID: "sig-1"}, {TaskID: "task-2", SignatureID: "sig-2"}}

const signPairsJSON = `[{"taskId":"task-1","signatureId":"sig-1"},{"taskId":"task-2","signatureId":"sig-2"}]`

func TestSignActions(t *testing.T) {
	for _, tc := range []struct {
		name, method, path, body string
		call                     func(*myitmo.Client) error
	}{
		{"submit", "POST", "/api/sign/tasks/task-1/signatures/sig-1/sign", `{"signatureValue":"TUlJ"}`,
			func(c *myitmo.Client) error { return c.Sign.SubmitSignature(t.Context(), "task-1", "sig-1", "TUlJ") }},
		{"simple", "POST", "/api/sign/sign", `{"provider":"default","tasks":` + signPairsJSON + `}`,
			func(c *myitmo.Client) error { return c.Sign.SignSimple(t.Context(), signPairs) }},
		{"reject", "POST", "/api/sign/reject", `{"tasks":` + signPairsJSON + `,"comment":"wrong data"}`,
			func(c *myitmo.Client) error { return c.Sign.Reject(t.Context(), signPairs, "wrong data") }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f, c := newFake(t)
			f.result(`null`)
			if err := tc.call(c); err != nil {
				t.Fatal(err)
			}
			f.expect(tc.method, tc.path).sameJSON(t, tc.body)
		})
	}
}

func TestSignGoskey(t *testing.T) {
	f, c := newFake(t)
	f.result(`12345`)
	order := must[string](t)(c.Sign.GoskeySign(t.Context(), signPairs))
	f.expect("POST", "/api/sign/goskey/sign").sameJSON(t, `{"provider":"goskey","tasks":`+signPairsJSON+`}`)
	if order != "12345" {
		t.Fatalf("order = %q", order)
	}

	f.result(`"ord-1"`)
	if order = must[string](t)(c.Sign.GoskeySign(t.Context(), signPairs)); order != "ord-1" {
		t.Fatalf("order = %q", order)
	}

	f.result(`{"status":"sign_reject"}`)
	st := must[*myitmo.GoskeyOrder](t)(c.Sign.GoskeyOrder(t.Context(), "ord-1"))
	f.expect("GET", "/api/sign/goskey/order/ord-1")
	if st.Status != myitmo.GoskeyOrderRejected {
		t.Fatalf("status = %+v", st)
	}
}

func TestSignKontur(t *testing.T) {
	f, c := newFake(t)
	f.result(`[{"subjectLastname":"Testova","subjectFirstname":"Test","subjectMiddlename":"Testovna","certificateThumbprint":"AB12CD"}]`)
	certs := must[[]myitmo.KonturCertificate](t)(c.Sign.KonturCertificates(t.Context()))
	f.expect("GET", "/api/sign/kontur/certificates/my")
	if len(certs) != 1 || certs[0].CertificateThumbprint != "AB12CD" || certs[0].SubjectFirstname != "Test" {
		t.Fatalf("certs = %+v", certs)
	}

	f.result(`{"operationId":"op-1"}`)
	op := must[*myitmo.KonturOperation](t)(c.Sign.KonturSign(t.Context(), signPairs[:1], "AB12CD"))
	f.expect("POST", "/api/sign/kontur/sign").sameJSON(t,
		`{"provider":"kontur","tasks":[{"taskId":"task-1","signatureId":"sig-1"}],"konturCertificateThumbprint":"AB12CD"}`)
	if op.OperationID != "op-1" {
		t.Fatalf("op = %+v", op)
	}

	f.result(`0`)
	res := must[int](t)(c.Sign.KonturConfirm(t.Context(), "op-1", "123456"))
	r := f.expect("POST", "/api/sign/kontur/sign/op-1/confirm")
	if r.Query.Get("code") != "123456" || res != 0 {
		t.Fatalf("query = %v, result = %d", r.Query, res)
	}
}
