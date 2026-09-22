package myitmo

import (
	"encoding/json/v2"
	"errors"
	"strconv"
	"strings"
)

// FlexID is an identifier that the server sends either as a JSON number or as
// a string; both mean the same. It encodes back as a JSON
// number when it holds an integer and as a string otherwise.
type FlexID string

// UnmarshalJSON accepts a number, a string or null.
func (v *FlexID) UnmarshalJSON(b []byte) error {
	s := string(b)
	switch {
	case s == "null":
		*v = ""
	case strings.HasPrefix(s, `"`):
		var str string
		if err := json.Unmarshal(b, &str); err != nil {
			return err
		}
		*v = FlexID(str)
	default:
		if _, err := strconv.ParseFloat(s, 64); err != nil {
			return errors.New("myitmo: identifier must be a number or a string")
		}
		*v = FlexID(s)
	}
	return nil
}

// MarshalJSON writes an integer identifier as a number, anything else as a string.
func (v FlexID) MarshalJSON() ([]byte, error) {
	if isJSONInteger(string(v)) {
		return []byte(v), nil
	}
	return json.Marshal(string(v))
}

func isJSONInteger(s string) bool {
	s = strings.TrimPrefix(s, "-")
	if s == "" || (len(s) > 1 && s[0] == '0') {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

// FlexInt is an integer that the server sends either as a JSON number or as a
// numeric string. An empty string or null decodes to 0. It encodes as a JSON number.
type FlexInt int64

// UnmarshalJSON accepts 2024, "2024", "" and null.
func (v *FlexInt) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), `"`)
	if s == "" || s == "null" {
		*v = 0
		return nil
	}
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return errors.New("myitmo: integer must be a number or a numeric string")
	}
	*v = FlexInt(n)
	return nil
}

// FlexNumber is an amount that the server sends either as a JSON number or as
// a numeric string. An empty string or
// null decodes to 0. It encodes as a JSON number.
type FlexNumber float64

// UnmarshalJSON accepts a number, a numeric string (a decimal comma is allowed) or null.
func (v *FlexNumber) UnmarshalJSON(b []byte) error {
	s := string(b)
	if strings.HasPrefix(s, `"`) {
		if err := json.Unmarshal(b, &s); err != nil {
			return err
		}
		s = strings.ReplaceAll(strings.TrimSpace(s), ",", ".")
	}
	if s == "" || s == "null" {
		*v = 0
		return nil
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return errors.New("myitmo: amount must be a number or a numeric string")
	}
	*v = FlexNumber(f)
	return nil
}

// FlexBool is a flag that the server sends as a JSON boolean or as 0/1
// (1 means true). It encodes as a JSON boolean.
type FlexBool bool

// UnmarshalJSON accepts true/false, 0/1 (also quoted) and null.
func (v *FlexBool) UnmarshalJSON(b []byte) error {
	switch strings.Trim(string(b), `"`) {
	case "true", "1":
		*v = true
	case "false", "0", "", "null":
		*v = false
	default:
		return errors.New("myitmo: flag must be a boolean or 0/1")
	}
	return nil
}
