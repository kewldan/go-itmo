package bars

import (
	"encoding/json/jsontext"
	"errors"
	"strconv"
	"time"

	"github.com/kewldan/go-itmo/internal/jsonx"
	"github.com/kewldan/go-itmo/internal/rest"
)

// RawJSON is a field seen on the wire whose shape is not confirmed yet.
type RawJSON = jsontext.Value

// File is a downloaded document such as a report export. The caller must close Body.
type File = rest.File

// Group or flow kinds that address a journal together with an identifier
// (GroupOrFlow.Type). Only "flow" is confirmed.
const (
	FlowTypeGroup = "group"
	FlowTypeFlow  = "flow"
)

// Role IDs (UserRole.ID) with their labels.
//
// Role ids are not confirmed: student accounts have been seen with id 2 or 3
// («Обучающийся»). Compare UserRole.Name as well before relying on an ID.
const (
	RoleTeacherAdmin       int64 = 1  // Преподаватель-Администратор
	RoleTeacherRealisator  int64 = 2  // Преподаватель-Реализатор
	RoleStudent            int64 = 3  // Ученик
	RoleParent             int64 = 4  // Родитель
	RoleDIDAdmin           int64 = 6  // DID admin, no known label
	RoleDODAdmin           int64 = 7  // ДОД Админ
	RoleSuperuser          int64 = 8  // Суперпользователь
	RoleMentor             int64 = 9  // Ментор
	RoleCurator            int64 = 10 // Куратор
	RoleDepartmentStaff    int64 = 11 // Сотрудник факультета
	RoleViceDepartmentHead int64 = 12 // Замдекана
	RoleOfficeManager      int64 = 13 // Менеджер офиса
)

// Attempts of the intermediate certification (Approval.Attempt).
const (
	AttemptFirst       = 1 // ПА
	AttemptRetry       = 2 // ППА, first retake
	AttemptSecondRetry = 3 // second retake
)

// Millis is a timestamp that BARS both sends and expects as Unix milliseconds.
// It is used in types that are read and written back; the zero value is
// encoded as null. Strings in any jsonx format are accepted on decode.
type Millis struct{ time.Time }

// MillisOf wraps t; the zero time stays zero (null on the wire).
func MillisOf(t time.Time) Millis { return Millis{t} }

// MarshalJSON encodes the time as Unix milliseconds or null.
func (m Millis) MarshalJSON() ([]byte, error) {
	if m.IsZero() {
		return []byte("null"), nil
	}
	return strconv.AppendInt(nil, m.UnixMilli(), 10), nil
}

// UnmarshalJSON decodes Unix milliseconds, a timestamp string or null.
func (m *Millis) UnmarshalJSON(b []byte) error {
	v := jsontext.Value(b)
	switch v.Kind() {
	case 'n':
		m.Time = time.Time{}
		return nil
	case '0':
		ms, err := strconv.ParseInt(string(b), 10, 64)
		if err != nil {
			f, ferr := strconv.ParseFloat(string(b), 64)
			if ferr != nil {
				return err
			}
			ms = int64(f)
		}
		m.Time = time.UnixMilli(ms).In(jsonx.MSK)
		return nil
	case '"':
		var s string
		if err := jsonx.Unmarshal(b, &s); err != nil {
			return err
		}
		if s == "" {
			m.Time = time.Time{}
			return nil
		}
		t, err := jsonx.ParseTime(s)
		if err != nil {
			return err
		}
		m.Time = t
		return nil
	}
	return errors.New("bars: timestamp must be a number, a string or null")
}

// Page is a Spring Data page.: only
// content, totalElements and pageable.pageNumber are read by it; the other
// fields are the standard Spring ones.
type Page[T any] struct {
	Content          []T      `json:"content"`
	TotalElements    int64    `json:"totalElements"`
	TotalPages       int      `json:"totalPages"`
	Size             int      `json:"size"`
	Number           int      `json:"number"`
	NumberOfElements int      `json:"numberOfElements"`
	First            bool     `json:"first"`
	Last             bool     `json:"last"`
	Empty            bool     `json:"empty"`
	Pageable         Pageable `json:"pageable"`
	// Sort is the Spring sort descriptor; its shape is not used.
	Sort RawJSON `json:"sort,omitzero"`
}

// Pageable is the request part of a Spring page.
type Pageable struct {
	PageNumber int     `json:"pageNumber"`
	PageSize   int     `json:"pageSize"`
	Offset     int64   `json:"offset"`
	Paged      bool    `json:"paged"`
	Unpaged    bool    `json:"unpaged"`
	Sort       RawJSON `json:"sort,omitzero"`
}

// LabeledValue is the {label, value} pair some list endpoints use for a
// discipline.
type LabeledValue struct {
	Label string `json:"label"`
	// Value is the discipline ID, sent as a number or a string.
	Value RawJSON `json:"value,omitzero"`
}
