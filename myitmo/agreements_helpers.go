package myitmo

import (
	"crypto/rand"
	"encoding/json/jsontext"
	"fmt"
	"strconv"
)

// newUUID returns a random RFC 4122 version 4 UUID.
func newUUID() string {
	var b [16]byte
	_, _ = rand.Read(b[:])
	b[6] = b[6]&0x0f | 0x40
	b[8] = b[8]&0x3f | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

// readFlexString reads a JSON string, number or null as a string; for ids that
// arrive either as a number or as a string.
func readFlexString(dec *jsontext.Decoder) (string, error) {
	tok, err := dec.ReadToken()
	if err != nil {
		return "", err
	}
	switch tok.Kind() {
	case 'n':
		return "", nil
	case '"':
		return tok.String(), nil
	case '0':
		return tok.String(), nil
	}
	return "", fmt.Errorf("myitmo: expected a string or a number, got %v", tok.Kind())
}

// readFlexInt reads a JSON number, numeric string or null as an int.
func readFlexInt(dec *jsontext.Decoder) (int, error) {
	s, err := readFlexString(dec)
	if err != nil || s == "" {
		return 0, err
	}
	return strconv.Atoi(s)
}
