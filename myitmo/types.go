package myitmo

import "encoding/json/jsontext"

// RawJSON is a field whose shape has been seen on the wire but not pinned
// down yet. Decode it yourself; it becomes a typed field once confirmed.
type RawJSON = jsontext.Value

// IDValue is a generic dictionary entry.
type IDValue struct {
	ID    int64  `json:"id"`
	Value string `json:"value"`
}

// Page is a page of results with the total number of matches.
type Page[T any] struct {
	Count int `json:"count"`
	Data  []T `json:"data"`
}

// ptr returns a pointer to v; handy for optional parameters.
func ptr[T any](v T) *T { return &v }

// Ptr returns a pointer to v, for optional fields of parameter structs.
func Ptr[T any](v T) *T { return ptr(v) }
