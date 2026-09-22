package myitmo

import (
	"context"
	"encoding/json/jsontext"
)

// SignService is electronic signature tasks (/api/sign).
//
// A sign task holds one document and one or more signature slots. Its id comes
// from another service, e.g. [AgreementsService.Sign] or a requests v2 task.
// The signing methods take the (task, signature) pairs to sign as
// [SignTaskSignature] values.
type SignService struct{ c *Client }

// Known sign task statuses; other values are possible.
const (
	SignTaskCreatedWithFile = "CREATED_WITH_FILE"
	SignTaskInProgress      = "SIGN_IN_PROGRESS"
	SignTaskSigned          = "SIGNED"
)

// SignSignaturePending is the status of a signature slot that still has to be signed.
const SignSignaturePending = "PENDING"

// Signature kinds a slot requires; they select the providers offered.
const (
	// SignTypeQualified is a qualified signature: CryptoPro ([SignService.SubmitSignature]) or Goskey.
	SignTypeQualified = "qualified"
	// SignTypeNonQualified is a non-qualified signature: Kontur or Goskey.
	SignTypeNonQualified = "nonQualified"
	// SignTypeSimple is a simple electronic signature ([SignService.SignSimple]).
	SignTypeSimple = "simple"
)

// Goskey order statuses that end polling; any other value means the order is still pending.
const (
	GoskeyOrderDone            = "done"
	GoskeyOrderRejected        = "sign_reject"
	GoskeyOrderSNILSNotFound   = "snils_not_found"
	GoskeyOrderExpired         = "expired"
	GoskeyOrderProcessingError = "processing_error"
)

// SignTask is a document to sign with its signature slots.
type SignTask struct {
	TaskID string `json:"taskId"`
	// Name is the document title; may be empty.
	Name string `json:"name"`
	// Type is the document type; "pdf" has a preview.
	Type string `json:"type"`
	// Status is SignTaskCreatedWithFile, SignTaskInProgress (both signable) or SignTaskSigned.
	Status     string          `json:"status"`
	Signatures []SignSignature `json:"signatures"`
}

// Signable reports whether the task accepts signatures and has a pending slot.
func (t *SignTask) Signable() bool {
	if t.Status != SignTaskCreatedWithFile && t.Status != SignTaskInProgress {
		return false
	}
	for _, s := range t.Signatures {
		if s.Status == SignSignaturePending {
			return true
		}
	}
	return false
}

// SignSignature is one signature slot of a task.
type SignSignature struct {
	SignatureID string `json:"signatureId"`
	// Status is SignSignaturePending while the slot is unsigned.
	Status string `json:"status"`
	// Type is SignTypeQualified, SignTypeNonQualified or SignTypeSimple.
	Type string `json:"type"`
	// Rejectable reports whether the signer may refuse via [SignService.Reject].
	Rejectable bool `json:"rejectable"`
}

// SignTaskSignature selects one signature slot of a task.
type SignTaskSignature struct {
	TaskID      string `json:"taskId"`
	SignatureID string `json:"signatureId"`
}

// GoskeyOrder is the state of a Goskey signing order.
type GoskeyOrder struct {
	// Status is one of the GoskeyOrder* constants when final.
	Status string `json:"status"`
}

// KonturCertificate is a Kontur cloud certificate of the user.
type KonturCertificate struct {
	SubjectLastname   string `json:"subjectLastname"`
	SubjectFirstname  string `json:"subjectFirstname"`
	SubjectMiddlename string `json:"subjectMiddlename"`
	// CertificateThumbprint is passed to [SignService.KonturSign].
	CertificateThumbprint string `json:"certificateThumbprint"`
	// The fields below are not confirmed.
	SignatureType           string `json:"signatureType"`
	CertificateSerialNumber string `json:"certificateSerialNumber"`
	CertificateValidFrom    string `json:"certificateValidFrom"`
	CertificateValidTo      string `json:"certificateValidTo"`
}

// KonturOperation is a started Kontur signing operation awaiting an SMS code.
type KonturOperation struct {
	OperationID string `json:"operationId"`
}

type signBody struct {
	Provider                    string              `json:"provider,omitzero"`
	Tasks                       []SignTaskSignature `json:"tasks"`
	KonturCertificateThumbprint string              `json:"konturCertificateThumbprint,omitzero"`
}

func signTasks(t []SignTaskSignature) []SignTaskSignature {
	if t == nil {
		return []SignTaskSignature{}
	}
	return t
}

// Task returns a sign task with its signature slots.
// GET /api/sign/tasks/{taskId}
func (s *SignService) Task(ctx context.Context, taskID string) (*SignTask, error) {
	return call[*SignTask](ctx, s.c, get("api/sign/tasks/"+id(taskID), nil))
}

// SourceFile downloads the original (unsigned) document of a task. The caller closes it.
// GET /api/sign/tasks/{taskId}/files/source
func (s *SignService) SourceFile(ctx context.Context, taskID string) (*File, error) {
	return download(ctx, s.c, get("api/sign/tasks/"+id(taskID)+"/files/source", nil))
}

// StampedFile downloads the signed document with the visual signature stamp
// (task status SignTaskSigned). The caller closes it.
// GET /api/sign/tasks/{taskId}/files/stamped
func (s *SignService) StampedFile(ctx context.Context, taskID string) (*File, error) {
	return download(ctx, s.c, get("api/sign/tasks/"+id(taskID)+"/files/stamped", nil))
}

// SubmitSignature submits a detached CAdES signature (base64, qualified,
// CryptoPro) of the source file for one signature slot.
// POST /api/sign/tasks/{taskId}/signatures/{signatureId}/sign
func (s *SignService) SubmitSignature(ctx context.Context, taskID, signatureID, signatureValue string) error {
	body := struct {
		SignatureValue string `json:"signatureValue"`
	}{signatureValue}
	return exec(ctx, s.c, post("api/sign/tasks/"+id(taskID)+"/signatures/"+id(signatureID)+"/sign", body))
}

// SignSimple signs the given slots with a simple electronic signature.
// POST /api/sign/sign
func (s *SignService) SignSimple(ctx context.Context, tasks []SignTaskSignature) error {
	return exec(ctx, s.c, post("api/sign/sign", signBody{Provider: "default", Tasks: signTasks(tasks)}))
}

// GoskeySign starts a Goskey signing order; the user confirms it in the Goskey
// app. It returns the order id for [SignService.GoskeyOrder], or "" when no
// order was created.
// POST /api/sign/goskey/sign
func (s *SignService) GoskeySign(ctx context.Context, tasks []SignTaskSignature) (string, error) {
	v, err := call[flexString](ctx, s.c, post("api/sign/goskey/sign", signBody{Provider: "goskey", Tasks: signTasks(tasks)}))
	return string(v), err
}

// GoskeyOrder returns the state of a Goskey order; poll it
// (e.g. every 2 s) until the status is final. It returns nil when the server has no state.
// GET /api/sign/goskey/order/{orderId}
func (s *SignService) GoskeyOrder(ctx context.Context, orderID string) (*GoskeyOrder, error) {
	return call[*GoskeyOrder](ctx, s.c, get("api/sign/goskey/order/"+id(orderID), nil))
}

// KonturCertificates lists the user's Kontur cloud certificates.
// GET /api/sign/kontur/certificates/my
func (s *SignService) KonturCertificates(ctx context.Context) ([]KonturCertificate, error) {
	return call[[]KonturCertificate](ctx, s.c, get("api/sign/kontur/certificates/my", nil))
}

// KonturSign starts a Kontur signing operation with the given certificate;
// the server sends an SMS code for [SignService.KonturConfirm].
// POST /api/sign/kontur/sign
func (s *SignService) KonturSign(ctx context.Context, tasks []SignTaskSignature, certificateThumbprint string) (*KonturOperation, error) {
	body := signBody{Provider: "kontur", Tasks: signTasks(tasks), KonturCertificateThumbprint: certificateThumbprint}
	return call[*KonturOperation](ctx, s.c, post("api/sign/kontur/sign", body))
}

// KonturConfirm confirms a Kontur operation with the SMS code. The result is
// 0 on success; a wrong code comes back as an error.
// POST /api/sign/kontur/sign/{operationId}/confirm
func (s *SignService) KonturConfirm(ctx context.Context, operationID, code string) (int, error) {
	r := withQuery(post("api/sign/kontur/sign/"+id(operationID)+"/confirm", nil), q().set("code", code))
	return call[int](ctx, s.c, r)
}

// Reject refuses to sign the given slots with a required reason. Only slots
// with Rejectable set can be rejected; this invalidates the document and
// revokes existing signatures.
// POST /api/sign/reject
func (s *SignService) Reject(ctx context.Context, tasks []SignTaskSignature, comment string) error {
	body := struct {
		Tasks   []SignTaskSignature `json:"tasks"`
		Comment string              `json:"comment"`
	}{signTasks(tasks), comment}
	return exec(ctx, s.c, post("api/sign/reject", body))
}

// flexString decodes a JSON string, number or null into a string.
type flexString string

func (f *flexString) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
	s, err := readFlexString(dec)
	*f = flexString(s)
	return err
}
