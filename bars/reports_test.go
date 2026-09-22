package bars_test

import (
	"errors"
	"io"
	"net/http"
	"testing"

	"github.com/kewldan/go-itmo/bars"
)

func TestReports(t *testing.T) {
	f, c := session(t)
	f.reply(`[{"name":"current_control","display_name":"Текущий контроль","additional_filters":[{"name":"discipline","display_name":"Дисциплина"},{"name":"teacher","display_name":"Преподаватель"}]}]`)
	types := must[[]bars.ReportType](t)(c.ReportTypes(ctx))
	f.expect(http.MethodGet, "report/type")
	if types[0].Name != bars.ReportCurrentControl || types[0].AdditionalFilters[1].Name != "teacher" {
		t.Errorf("types = %+v", types)
	}

	filter := bars.ReportFilter{Year: "2026/2027", Term: bars.Spring, Type: bars.FlowTypeFlow, Identifier: "a/b", Extra: map[string]string{"teacher": "Иванов"}}
	f.reply(`{"headers":["ФИО","Балл","Сдал","Детали",""],"rows":[["Студент",75.5,true,{"headers":["КТ"],"rows":[["Тест",null]]},null]],"filters":{}}`)
	table := must[*bars.ReportTable](t)(c.Report(ctx, "", filter))
	f.expect(http.MethodPost, "report/current_control").sameJSON(t, `{"year":"2026/2027","term":0,"type":"flow","identifier":"a/b","discipline":"","teacher":"Иванов"}`)
	row := table.Rows[0]
	if row[0].Text() != "Студент" || row[1].Text() != "75.5" || row[2].Text() != "true" || !row[4].IsNull() || row[0].IsNull() {
		t.Errorf("row = %+v", row)
	}
	nested, isNested, err := row[3].Nested()
	if err != nil || !isNested || nested.Headers[0] != "КТ" || nested.Rows[0][0].Text() != "Тест" || !nested.Rows[0][1].IsNull() {
		t.Errorf("nested = %+v %v %v", nested, isNested, err)
	}
	if _, isNested, _ := row[0].Nested(); isNested || row[3].Text() != "" {
		t.Error("scalar reported as nested")
	}

	f.reply(`{"headers":[],"rows":[]}`)
	must[*bars.ReportTable](t)(c.Report(ctx, "by teacher", bars.ReportFilter{Year: "2026/2027", Term: bars.Autumn}))
	f.expect(http.MethodPost, "report/by%20teacher").sameJSON(t, `{"year":"2026/2027","term":1,"type":"","identifier":"","discipline":""}`)

	if _, err := c.Report(ctx, "", bars.ReportFilter{}); err == nil {
		t.Error("accepted a filter without year")
	}
	if _, err := c.ExportReport(ctx, "", bars.ReportFilter{}); err == nil {
		t.Error("accepted a filter without year")
	}
}

func TestExportReport(t *testing.T) {
	f, c := session(t)
	f.queue(step{header: map[string]string{
		"Content-Type":        "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		"Content-Disposition": `attachment; filename="report.xlsx"`,
	}, body: "PK\x03\x04synthetic"})
	file := must[*bars.File](t)(c.ExportReport(ctx, "current_control", bars.ReportFilter{Year: "2026/2027", Discipline: "Предмет"}))
	defer func() { _ = file.Body.Close() }()
	f.expect(http.MethodPost, "report/current_control/export").sameJSON(t, `{"year":"2026/2027","term":0,"type":"","identifier":"","discipline":"Предмет"}`)
	data, err := io.ReadAll(file.Body)
	if err != nil || string(data) != "PK\x03\x04synthetic" || file.Name != "report.xlsx" {
		t.Errorf("file = %+v %q %v", file, data, err)
	}

	f.queue(step{status: http.StatusForbidden})
	var e *bars.Error
	if _, err := c.ExportReport(ctx, "x", bars.ReportFilter{Year: "2026/2027"}); !errors.As(err, &e) || !e.IsForbidden() {
		t.Fatalf("err = %v", err)
	}
}

func TestGoogleSheets(t *testing.T) {
	f, c := session(t)
	sheet := "https://docs.google.com/spreadsheets/d/abc/edit?usp=sharing&x=1#gid=0"
	f.reply("")
	list, err := c.ImportGoogleSheet(ctx, sheet, false)
	if err != nil || list != nil {
		t.Fatalf("import = %+v, %v", list, err)
	}
	f.expect(http.MethodGet, "google/import/sheet?spreadSheetUrlWithSheetGid=https%3A%2F%2Fdocs.google.com%2Fspreadsheets%2Fd%2Fabc%2Fedit%3Fusp%3Dsharing%26x%3D1%23gid%3D0")

	f.queue(step{status: http.StatusInternalServerError, body: `{"accessErrors":[],"namingErrors":["Лист не найден"],"warnings":["w"]}`})
	list, err = c.ImportGoogleSheet(ctx, "s", true)
	f.expect(http.MethodGet, "google/import/sheet?spreadSheetUrlWithSheetGid=s&validateOnly=true")
	var e *bars.Error
	if !errors.As(err, &e) || e.StatusCode != http.StatusInternalServerError || list == nil || list.NamingErrors[0] != "Лист не найден" {
		t.Fatalf("import error = %+v, %v", list, err)
	}

	f.queue(step{status: http.StatusBadRequest, body: "not json"})
	if list, err = c.ImportGoogleSheet(ctx, "s", false); list != nil || !errors.As(err, &e) {
		t.Fatalf("import error = %+v, %v", list, err)
	}

	f.reply(`"https://docs.google.com/spreadsheets/d/new"`)
	link := must[string](t)(c.ExportGoogleSheet(ctx, bars.GoogleExportParams{CheckpointPlanID: 8, SheetName: "Лист 1", SheetTitle: "T&T", PlanIdentifierName: "a/b"}))
	f.expect(http.MethodGet, "google/export/sheet?checkpointPlanId=8&planIdentifierName=a%2Fb&sheetName=%D0%9B%D0%B8%D1%81%D1%82+1&sheetTitle=T%26T")
	if link != "https://docs.google.com/spreadsheets/d/new" {
		t.Errorf("link = %q", link)
	}
	f.queue(step{header: map[string]string{"Content-Type": "text/plain"}, body: "https://docs.google.com/x\n"})
	if link = must[string](t)(c.ExportGoogleSheet(ctx, bars.GoogleExportParams{})); link != "https://docs.google.com/x" {
		t.Errorf("plain link = %q", link)
	}
	f.expect(http.MethodGet, "google/export/sheet")
}
