// Package jsonx holds the JSON options shared by every client in the module.
//
// ITMO services are not consistent about timestamps: most send RFC 3339 with an
// offset, some drop the seconds, some drop the offset, and a few send an empty
// string instead of null. Decoding goes through [Unmarshal] so that model
// structs can use plain [time.Time] fields.
package jsonx

import (
	"encoding/json/jsontext"
	"encoding/json/v2"
	"errors"
	"io"
	"time"
)

// MSK is the zone of every ITMO service; timestamps without an offset are read in it.
var MSK = time.FixedZone("MSK", 3*60*60)

var layouts = []string{
	time.RFC3339Nano,
	"2006-01-02T15:04Z07:00",
	"2006-01-02T15:04:05.999999999",
	"2006-01-02T15:04:05",
	"2006-01-02T15:04",
	"2006-01-02 15:04:05.999999999Z07:00",
	"2006-01-02 15:04:05.999999999",
	"2006-01-02 15:04:05",
	"2006-01-02",
}

// ParseTime parses the timestamp formats used across ITMO services.
func ParseTime(s string) (time.Time, error) {
	for _, layout := range layouts {
		if t, err := time.ParseInLocation(layout, s, MSK); err == nil {
			return t, nil
		}
	}
	return time.Time{}, &time.ParseError{Layout: time.RFC3339, Value: s, Message: ": unsupported timestamp format"}
}

func unmarshalTime(dec *jsontext.Decoder, t *time.Time) error {
	tok, err := dec.ReadToken()
	if err != nil {
		return err
	}
	switch tok.Kind() {
	case 'n':
		*t = time.Time{}
		return nil
	case '"':
		s := tok.String()
		if s == "" {
			*t = time.Time{}
			return nil
		}
		parsed, err := ParseTime(s)
		if err != nil {
			return err
		}
		*t = parsed
		return nil
	case '0':
		// Unix milliseconds, used by a few staff services.
		ms, err := tok.Int()
		if err != nil {
			return err
		}
		*t = time.UnixMilli(ms).In(MSK)
		return nil
	}
	return errors.New("jsonx: timestamp must be a string, a number or null")
}

var decodeOptions = json.JoinOptions(
	json.WithUnmarshalers(json.UnmarshalFromFunc(unmarshalTime)),
)

// Unmarshal decodes data into v with the module-wide options.
func Unmarshal(data []byte, v any) error {
	return json.Unmarshal(data, v, decodeOptions)
}

// UnmarshalRead decodes r into v with the module-wide options.
func UnmarshalRead(r io.Reader, v any) error {
	return json.UnmarshalRead(r, v, decodeOptions)
}

// Marshal encodes v. Request bodies never contain time.Time directly; dates
// use their own MarshalText.
func Marshal(v any) ([]byte, error) {
	return json.Marshal(v)
}
