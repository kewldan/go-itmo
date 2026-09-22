package bars

import (
	"context"
	"fmt"
	"strings"

	"github.com/kewldan/go-itmo/internal/jsonx"
	"github.com/kewldan/go-itmo/internal/rest"
)

// ImportErrors is the error body of a failed Google Sheets import; each list
// holds messages in Russian.
type ImportErrors struct {
	AccessErrors       []string `json:"accessErrors"`
	GtAddressErrors    []string `json:"gtAddressErrors"`
	MarksError         []string `json:"marksError"`
	NamingErrors       []string `json:"namingErrors"`
	StudentsListErrors []string `json:"studentsListErrors"`
	Warnings           []string `json:"warnings"`
}

// ImportGoogleSheet imports marks from a Google Sheet into the journal;
// sheetURL is the sheet link including "#gid=...". validateOnly checks the
// sheet without importing. On failure it returns the decoded error lists,
// when the body has them, together with *Error. The link is sent fully
// query-encoded.
//
// Warning: this GET writes marks on the server.
//
// GET /google/import/sheet
//
// Audience: Teacher.
func (c *Client) ImportGoogleSheet(ctx context.Context, sheetURL string, validateOnly bool) (*ImportErrors, error) {
	r := get("google/import/sheet", q().str("spreadSheetUrlWithSheetGid", sheetURL).flag("validateOnly", validateOnly).values())
	resp, sent, err := c.exchange(ctx, r) //nolint:bodyclose // closed by rest.ReadAll
	if err != nil {
		return nil, err
	}
	body, err := rest.ReadAll(resp, 4<<20)
	if err != nil {
		return nil, fmt.Errorf("bars: read %s %s: %w", r.Method, r.Path, err)
	}
	if resp.StatusCode/100 == 2 {
		return nil, c.rotate(ctx, resp, sent)
	}
	callErr := &Error{Method: r.Method, Path: r.Path, StatusCode: resp.StatusCode}
	var list ImportErrors
	if len(body) > 0 && jsonx.Unmarshal(body, &list) == nil {
		return &list, callErr
	}
	return nil, callErr
}

// GoogleExportParams selects the journal to export; empty values are not sent.
type GoogleExportParams struct {
	CheckpointPlanID int64
	SheetName        string
	SheetTitle       string
	// PlanIdentifierName is a GroupOrFlow.Identifier of the plan.
	PlanIdentifierName string
}

// ExportGoogleSheet exports a journal to a new Google Sheet and returns its
// URL. The server answers 404 for a wrong or missing sheet.
//
// GET /google/export/sheet
//
// Audience: Teacher.
func (c *Client) ExportGoogleSheet(ctx context.Context, p GoogleExportParams) (string, error) {
	qv := q().num("checkpointPlanId", p.CheckpointPlanID).
		str("sheetName", p.SheetName).
		str("sheetTitle", p.SheetTitle).
		str("planIdentifierName", p.PlanIdentifierName).
		values()
	r := get("google/export/sheet", qv)
	resp, err := c.send(ctx, r) //nolint:bodyclose // closed by rest.ReadAll
	if err != nil {
		return "", err
	}
	body, err := rest.ReadAll(resp, 1<<20)
	if err != nil {
		return "", fmt.Errorf("bars: read %s %s: %w", r.Method, r.Path, err)
	}
	// The content type is unknown: accept a JSON string or plain text.
	text := strings.TrimSpace(string(body))
	if strings.HasPrefix(text, `"`) {
		var s string
		if err := jsonx.Unmarshal([]byte(text), &s); err != nil {
			return "", fmt.Errorf("bars: decode %s %s: %w", r.Method, r.Path, err)
		}
		return s, nil
	}
	return text, nil
}
