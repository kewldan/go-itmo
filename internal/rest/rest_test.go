package rest

import (
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"strings"
	"testing"
)

func TestMultipartStreamsFieldsAndFiles(t *testing.T) {
	body, ct := Multipart(map[string]string{"a": "1"}, Upload{Field: "file", Name: `re"port.pdf`, ContentType: "application/pdf", Content: strings.NewReader("PDF")})
	_, params, err := mime.ParseMediaType(ct)
	if err != nil {
		t.Fatal(err)
	}
	r := multipart.NewReader(body, params["boundary"])
	form, err := r.ReadForm(1 << 20)
	if err != nil {
		t.Fatal(err)
	}
	if form.Value["a"][0] != "1" {
		t.Errorf("fields = %v", form.Value)
	}
	fh := form.File["file"][0]
	f, _ := fh.Open()
	data, _ := io.ReadAll(f)
	if fh.Filename != `re"port.pdf` || string(data) != "PDF" || fh.Header.Get("Content-Type") != "application/pdf" {
		t.Errorf("file = %q %q %v", fh.Filename, data, fh.Header)
	}
}

func TestEvents(t *testing.T) {
	stream := ": comment\nevent: delta\ndata: {\"a\":1}\n\nid: 7\ndata: line1\ndata: line2\n\ndata: tail"
	var got []Event
	for ev, err := range Events(io.NopCloser(strings.NewReader(stream))) {
		if err != nil {
			t.Fatal(err)
		}
		got = append(got, ev)
	}
	if len(got) != 3 || got[0].Event != "delta" || got[0].Data != `{"a":1}` || got[1].ID != "7" || got[1].Data != "line1\nline2" || got[2].Data != "tail" {
		t.Errorf("events = %+v", got)
	}
}

func TestFileFrom(t *testing.T) {
	resp := &http.Response{Header: http.Header{
		"Content-Type":        {"application/pdf"},
		"Content-Disposition": {`attachment; filename="order.pdf"`},
	}, Body: io.NopCloser(strings.NewReader("x")), ContentLength: 1}
	f := FileFrom(resp)
	if f.Name != "order.pdf" || f.ContentType != "application/pdf" || f.Size != 1 {
		t.Errorf("file = %+v", f)
	}
}
