package bars

import (
	"context"
	"errors"

	"github.com/kewldan/go-itmo/internal/jsonx"
)

// ReportCurrentControl is the default report.
const ReportCurrentControl = "current_control"

// ReportType is a kind of report with its extra filters.
type ReportType struct {
	// Name is the {reportName} path segment, e.g. ReportCurrentControl.
	Name              string            `json:"name"`
	DisplayName       string            `json:"display_name"`
	AdditionalFilters []ReportFilterDef `json:"additional_filters"`
}

// ReportFilterDef is an extra report filter. The filter "discipline" takes a
// discipline name (ReportFilter.Discipline); the others are free text in
// ReportFilter.Extra.
type ReportFilterDef struct {
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
}

// ReportFilter is the body of a report request. Empty strings are sent as
// they are.
type ReportFilter struct {
	// Year is "2025/2026"; required.
	Year string `json:"year"`
	// Term is required; Spring is 0 and is sent too.
	Term Term `json:"term"`
	// Type is FlowTypeGroup, FlowTypeFlow or "".
	Type string `json:"type"`
	// Identifier is a WhiteListEntry.Identifier or "".
	Identifier string `json:"identifier"`
	// Discipline is a discipline name, not an ID, or "".
	Discipline string `json:"discipline"`
	// Extra holds the other ReportType.AdditionalFilters by name; they are
	// sent as members of the same object.
	Extra map[string]string `json:",embed"`
}

// ReportTable is a report for display. An empty Headers means no data.
type ReportTable struct {
	Headers []string       `json:"headers"`
	Rows    [][]ReportCell `json:"rows"`
	// Filters is not decoded.
	Filters RawJSON `json:"filters,omitzero"`
}

// ReportNested is a table inside a ReportCell.
type ReportNested struct {
	Headers []string       `json:"headers"`
	Rows    [][]ReportCell `json:"rows"`
}

// ReportCell is one value of a report: a string, a number, a bool, null or
// a nested table.
type ReportCell struct {
	Raw RawJSON
}

// UnmarshalJSON keeps the raw value.
func (c *ReportCell) UnmarshalJSON(b []byte) error {
	c.Raw = append(RawJSON(nil), b...)
	return nil
}

// MarshalJSON writes the raw value back; an empty cell is null.
func (c ReportCell) MarshalJSON() ([]byte, error) {
	if len(c.Raw) == 0 {
		return []byte("null"), nil
	}
	return c.Raw, nil
}

// IsNull reports a null cell, usually shown as "--".
func (c ReportCell) IsNull() bool { return len(c.Raw) == 0 || c.Raw.Kind() == 'n' }

// Text returns strings unquoted and numbers and bools as written; null and
// nested tables give "".
func (c ReportCell) Text() string {
	if len(c.Raw) == 0 {
		return ""
	}
	switch c.Raw.Kind() {
	case '"':
		var s string
		if jsonx.Unmarshal(c.Raw, &s) == nil {
			return s
		}
	case '0', 't', 'f':
		return string(c.Raw)
	}
	return ""
}

// Nested decodes a nested table; ok is false for other values.
func (c ReportCell) Nested() (table *ReportNested, ok bool, err error) {
	if len(c.Raw) == 0 || c.Raw.Kind() != '{' {
		return nil, false, nil
	}
	var n ReportNested
	if err := jsonx.Unmarshal(c.Raw, &n); err != nil {
		return nil, true, err
	}
	return &n, true, nil
}

func reportName(name string) string {
	if name == "" {
		return ReportCurrentControl
	}
	return name
}

var errReportYear = errors.New("bars: ReportFilter.Year is required")

// ReportTypes returns the available reports and their extra filters.
//
// GET /report/type
//
// Audience: Teacher, Admin.
func (c *Client) ReportTypes(ctx context.Context) ([]ReportType, error) {
	return fetch[[]ReportType](ctx, c, get("report/type", nil))
}

// Report builds a report for display; name "" means ReportCurrentControl.
//
// POST /report/{reportName}
//
// Audience: Teacher, Admin.
func (c *Client) Report(ctx context.Context, name string, f ReportFilter) (*ReportTable, error) {
	if f.Year == "" {
		return nil, errReportYear
	}
	return fetch[*ReportTable](ctx, c, post(route("report", reportName(name)), f))
}

// ExportReport builds a report as a file (an .xlsx workbook);
// name "" means ReportCurrentControl. The caller must close File.Body.
//
// POST /report/{reportName}/export
//
// Audience: Teacher, Admin.
func (c *Client) ExportReport(ctx context.Context, name string, f ReportFilter) (*File, error) {
	if f.Year == "" {
		return nil, errReportYear
	}
	return c.download(ctx, post(route("report", reportName(name), "export"), f))
}
