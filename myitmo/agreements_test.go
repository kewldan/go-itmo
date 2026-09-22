package myitmo_test

import (
	"regexp"
	"testing"

	"github.com/kewldan/go-itmo/myitmo"
)

func TestAgreementsList(t *testing.T) {
	f, c := newFake(t)
	f.result(`[{"id":12,"title":"Consent to data processing","versions":[
		{"id":"101","date":"2025-01-10T10:00:00+03:00","status_id":6,"status_name":"Signed","date_signed":"2025-01-11T09:00:00+03:00","date_revoked":null,"reject_reason":null,"task_id":null,"signature_id":null},
		{"id":102,"date":"2026-02-01T12:00:00+03:00","status_id":"11","status_name":"Signing","date_signed":null,"date_revoked":null,"reject_reason":null,"task_id":"task-1","signature_id":"sig-1"}]},
		{"id":13,"title":"Terms","versions":null}]`)
	list := must[[]myitmo.Agreement](t)(c.Agreements.List(t.Context()))
	f.expect("GET", "/api/agreements/agreements")
	if len(list) != 2 || list[0].ID != "12" || list[0].Title != "Consent to data processing" || len(list[1].Versions) != 0 {
		t.Fatalf("list = %+v", list)
	}
	v := list[0].Versions
	if v[0].ID != "101" || v[0].StatusID != myitmo.AgreementStatusSigned || v[0].DateSigned == nil || v[0].DateRevoked != nil {
		t.Fatalf("version 0 = %+v", v[0])
	}
	if v[1].ID != "102" || v[1].StatusID != myitmo.AgreementStatusSigning || v[1].TaskID != "task-1" || v[1].SignatureID != "sig-1" {
		t.Fatalf("version 1 = %+v", v[1])
	}
	latest := list[0].Latest()
	if latest == nil || latest.ID != "102" || latest.Signable() {
		t.Fatalf("latest = %+v", latest)
	}
	if list[1].Latest() != nil {
		t.Fatal("latest of empty history must be nil")
	}
}

func TestAgreementsDownloadLink(t *testing.T) {
	for _, tc := range []struct{ name, result string }{
		{"string", `"https://files.example/doc.pdf"`},
		{"url", `{"url":"https://files.example/doc.pdf"}`},
		{"signed_link", `{"signed_link":"https://files.example/doc.pdf"}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f, c := newFake(t)
			f.result(tc.result)
			link := must[string](t)(c.Agreements.DownloadLink(t.Context(), "12", "101"))
			f.expect("GET", "/api/agreements/agreements/12/101/download")
			if link != "https://files.example/doc.pdf" {
				t.Fatalf("link = %q", link)
			}
		})
	}
	f, c := newFake(t)
	f.result(`{}`)
	if _, err := c.Agreements.DownloadLink(t.Context(), "12", "101"); err == nil {
		t.Fatal("want an error when no link is returned")
	}
}

var uuidV4 = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)

func TestAgreementsSign(t *testing.T) {
	f, c := newFake(t)
	f.result(`{"task_id":"task-1","signature_id":"sig-1"}`)
	ref := must[*myitmo.AgreementSignTask](t)(c.Agreements.Sign(t.Context(), "12", "key-1"))
	r := f.expect("POST", "/api/agreements/agreements/12/sign")
	r.sameJSON(t, `{}`)
	if r.Header.Get("Idempotency-Key") != "key-1" {
		t.Fatalf("Idempotency-Key = %q", r.Header.Get("Idempotency-Key"))
	}
	if ref.TaskID != "task-1" || ref.SignatureID != "sig-1" {
		t.Fatalf("ref = %+v", ref)
	}

	f.result(`{"task_id":"task-2","signature_id":"sig-2"}`)
	must[*myitmo.AgreementSignTask](t)(c.Agreements.Sign(t.Context(), "12", ""))
	if k := f.last().Header.Get("Idempotency-Key"); !uuidV4.MatchString(k) {
		t.Fatalf("generated Idempotency-Key = %q, want a UUID v4", k)
	}
}

func TestAgreementsSignBatch(t *testing.T) {
	f, c := newFake(t)
	f.result(`[{"agreement_id":"12","task_id":"task-1","signature_id":"sig-1","error":null},
		{"agreement_id":13,"task_id":null,"signature_id":null,"error":{"message":"already signed"}}]`)
	items := must[[]myitmo.AgreementSignBatchItem](t)(c.Agreements.SignBatch(t.Context(), []myitmo.AgreementID{"12", "13"}, ""))
	r := f.expect("POST", "/api/agreements/agreements/sign-batch")
	r.sameJSON(t, `{"agreement_ids":["12","13"]}`)
	if !uuidV4.MatchString(r.Header.Get("Idempotency-Key")) {
		t.Fatalf("Idempotency-Key = %q", r.Header.Get("Idempotency-Key"))
	}
	if len(items) != 2 || items[0].Failed() || !items[1].Failed() || items[1].AgreementID != "13" {
		t.Fatalf("items = %+v", items)
	}
}
