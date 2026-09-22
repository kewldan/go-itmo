package jsonx

import (
	"testing"
	"time"
)

func TestParseTime(t *testing.T) {
	want := time.Date(2026, 9, 1, 10, 30, 0, 0, MSK)
	for _, in := range []string{
		"2026-09-01T10:30:00+03:00",
		"2026-09-01T07:30:00Z",
		"2026-09-01T10:30+03:00",
		"2026-09-01T10:30:00",
		"2026-09-01T10:30",
		"2026-09-01 10:30:00",
		"2026-09-01T10:30:00.000+03:00",
	} {
		got, err := ParseTime(in)
		if err != nil || !got.Equal(want) {
			t.Errorf("ParseTime(%q) = %v, %v", in, got, err)
		}
	}
	if _, err := ParseTime("yesterday"); err == nil {
		t.Error("accepted garbage")
	}
}

func TestUnmarshalTimeVariants(t *testing.T) {
	var v struct {
		A time.Time  `json:"a"`
		B *time.Time `json:"b"`
		C time.Time  `json:"c"`
		D *time.Time `json:"d"`
		E time.Time  `json:"e"`
	}
	err := Unmarshal([]byte(`{"a":"2026-09-01","b":"2026-09-01T10:30","c":1767225600000,"d":null,"e":""}`), &v)
	if err != nil {
		t.Fatal(err)
	}
	if v.A.Day() != 1 || v.B == nil || v.B.Hour() != 10 || v.C.UnixMilli() != 1767225600000 || v.D != nil || !v.E.IsZero() {
		t.Errorf("decoded %+v", v)
	}
	if err := Unmarshal([]byte(`{"a":true}`), &v); err == nil {
		t.Error("accepted a bool timestamp")
	}
}
