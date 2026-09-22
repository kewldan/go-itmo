package myitmo

import (
	"context"
	"encoding/json/jsontext"
	"errors"
	"time"

	"github.com/kewldan/go-itmo/internal/jsonx"
)

// AgreementsService is documents to acknowledge and sign (/api/agreements).
type AgreementsService struct{ c *Client }

// AgreementID is an agreement or agreement version id. The server may send it
// as a number or a string; both decode here.
type AgreementID string

// UnmarshalJSONFrom accepts a JSON string, number or null.
func (a *AgreementID) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
	s, err := readFlexString(dec)
	*a = AgreementID(s)
	return err
}

// AgreementStatus is the status_id of an agreement version. The server may
// send it as a number or a numeric string; both decode here.
type AgreementStatus int

// Agreement version statuses.
const (
	// AgreementStatusSigned is a signed version; only these can be downloaded.
	AgreementStatusSigned AgreementStatus = 6
	// AgreementStatusRejected is a rejected version; see RejectReason.
	AgreementStatusRejected AgreementStatus = 8
	// AgreementStatusRevoked is a revoked version; see DateSigned and DateRevoked.
	AgreementStatusRevoked AgreementStatus = 9
	// AgreementStatusSigning is a version whose signing is in progress.
	AgreementStatusSigning AgreementStatus = 11
)

// UnmarshalJSONFrom accepts a JSON number, numeric string or null.
func (s *AgreementStatus) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
	n, err := readFlexInt(dec)
	*s = AgreementStatus(n)
	return err
}

// Agreement is a consent (e.g. personal-data processing) with its version history.
type Agreement struct {
	ID    AgreementID `json:"id"`
	Title string      `json:"title"`
	// Versions may be empty; the newest by Date is the current one.
	Versions []AgreementVersion `json:"versions"`
}

// Latest returns the newest version by Date, or nil when there is none.
func (a *Agreement) Latest() *AgreementVersion {
	var latest *AgreementVersion
	for i := range a.Versions {
		if latest == nil || a.Versions[i].Date.After(latest.Date) {
			latest = &a.Versions[i]
		}
	}
	return latest
}

// AgreementVersion is one version of an agreement.
type AgreementVersion struct {
	ID   AgreementID `json:"id"`
	Date time.Time   `json:"date"`
	// StatusID is one of the AgreementStatus constants; other values are possible.
	StatusID AgreementStatus `json:"status_id"`
	// StatusName is the localized status label.
	StatusName  string     `json:"status_name"`
	DateSigned  *time.Time `json:"date_signed"`
	DateRevoked *time.Time `json:"date_revoked"`
	// RejectReason is set for rejected versions.
	RejectReason string `json:"reject_reason"`
	// TaskID and SignatureID reference an existing sign task (see [SignService.Task]);
	// when both are set there is no need to call [AgreementsService.Sign].
	TaskID      string `json:"task_id"`
	SignatureID string `json:"signature_id"`
}

// Signable reports whether the version can be signed (neither signed nor being signed).
func (v *AgreementVersion) Signable() bool {
	return v.StatusID != AgreementStatusSigned && v.StatusID != AgreementStatusSigning
}

// AgreementSignTask references the sign task created for an agreement; pass
// it to [SignService.Task] and the signing methods of [SignService].
type AgreementSignTask struct {
	TaskID      string `json:"task_id"`
	SignatureID string `json:"signature_id"`
}

// AgreementSignBatchItem is the outcome for one agreement of a batch sign.
type AgreementSignBatchItem struct {
	// AgreementID may be absent.
	AgreementID AgreementID `json:"agreement_id"`
	TaskID      string      `json:"task_id"`
	SignatureID string      `json:"signature_id"`
	// Error is set (non-null) when the task could not be created for this item.
	Error RawJSON `json:"error"`
}

// Failed reports whether the item carries an error or lacks the task ids.
func (i *AgreementSignBatchItem) Failed() bool {
	e := string(i.Error)
	return (e != "" && e != "null" && e != "false" && e != `""`) || i.TaskID == "" || i.SignatureID == ""
}

// List returns the user's agreements with their versions.
// GET /api/agreements/agreements
func (s *AgreementsService) List(ctx context.Context) ([]Agreement, error) {
	return call[[]Agreement](ctx, s.c, get("api/agreements/agreements", nil))
}

// DownloadLink returns a link to the signed document of an agreement version
// (status [AgreementStatusSigned]). The response carries a link, not the file.
// GET /api/agreements/agreements/{agreementId}/{versionId}/download
func (s *AgreementsService) DownloadLink(ctx context.Context, agreementID, versionID AgreementID) (string, error) {
	path := "api/agreements/agreements/" + id(string(agreementID)) + "/" + id(string(versionID)) + "/download"
	raw, err := call[RawJSON](ctx, s.c, get(path, nil))
	if err != nil {
		return "", err
	}
	if link := linkFrom(raw); link != "" {
		return link, nil
	}
	return "", &DecodeError{Method: "GET", Path: path, Err: errors.New("no download link in the response")}
}

// linkFrom extracts the link from a string result or from its url, link or signed_link field.
func linkFrom(raw RawJSON) string {
	var s string
	if jsonx.Unmarshal(raw, &s) == nil {
		return s
	}
	var obj struct {
		URL        string `json:"url"`
		Link       string `json:"link"`
		SignedLink string `json:"signed_link"`
	}
	if jsonx.Unmarshal(raw, &obj) != nil {
		return ""
	}
	for _, v := range []string{obj.URL, obj.Link, obj.SignedLink} {
		if v != "" {
			return v
		}
	}
	return ""
}

// Sign creates (or reuses) a sign task for the latest version of an agreement.
// idempotencyKey is sent as the Idempotency-Key header; reuse the same key to
// retry safely. When empty, a random UUID v4 is generated.
// POST /api/agreements/agreements/{agreementId}/sign
func (s *AgreementsService) Sign(ctx context.Context, agreementID AgreementID, idempotencyKey string) (*AgreementSignTask, error) {
	r := post("api/agreements/agreements/"+id(string(agreementID))+"/sign", struct{}{})
	return call[*AgreementSignTask](ctx, s.c, withHeader(r, "Idempotency-Key", orUUID(idempotencyKey)))
}

// SignBatch creates sign tasks for several agreements at once (at most 10
// per call). Items with [AgreementSignBatchItem.Failed] should be skipped.
// idempotencyKey works as in [AgreementsService.Sign].
// POST /api/agreements/agreements/sign-batch
func (s *AgreementsService) SignBatch(ctx context.Context, agreementIDs []AgreementID, idempotencyKey string) ([]AgreementSignBatchItem, error) {
	body := struct {
		AgreementIDs []AgreementID `json:"agreement_ids"`
	}{agreementIDs}
	if body.AgreementIDs == nil {
		body.AgreementIDs = []AgreementID{}
	}
	r := post("api/agreements/agreements/sign-batch", body)
	return call[[]AgreementSignBatchItem](ctx, s.c, withHeader(r, "Idempotency-Key", orUUID(idempotencyKey)))
}

func orUUID(key string) string {
	if key == "" {
		return newUUID()
	}
	return key
}
