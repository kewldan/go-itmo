package myitmo

import (
	"fmt"
	"time"

	"github.com/kewldan/go-itmo/internal/jsonx"
)

// MSK is the time zone of every ITMO service.
var MSK = jsonx.MSK

// Date is a calendar date without a time zone, sent as "2006-01-02".
type Date struct {
	Year  int
	Month time.Month
	Day   int
}

// NewDate returns the date y-m-d.
func NewDate(y int, m time.Month, d int) Date { return Date{y, m, d} }

// DateOf returns the calendar date of t in its own location.
func DateOf(t time.Time) Date {
	y, m, d := t.Date()
	return Date{y, m, d}
}

// Today returns the current date in Moscow.
func Today() Date { return DateOf(time.Now().In(MSK)) }

// ParseDate parses "2006-01-02"; a timestamp prefix like "2006-01-02T..." is accepted too.
func ParseDate(s string) (Date, error) {
	if len(s) > 10 && (s[10] == 'T' || s[10] == ' ') {
		s = s[:10]
	}
	t, err := time.Parse(time.DateOnly, s)
	if err != nil {
		return Date{}, fmt.Errorf("myitmo: invalid date %q", s)
	}
	return DateOf(t), nil
}

// String formats the date as "2006-01-02".
func (d Date) String() string { return fmt.Sprintf("%04d-%02d-%02d", d.Year, d.Month, d.Day) }

// IsZero reports whether d is the zero date.
func (d Date) IsZero() bool { return d == Date{} }

// In returns midnight of d in loc.
func (d Date) In(loc *time.Location) time.Time {
	return time.Date(d.Year, d.Month, d.Day, 0, 0, 0, 0, loc)
}

// AddDays returns d shifted by n days.
func (d Date) AddDays(n int) Date { return DateOf(d.In(time.UTC).AddDate(0, 0, n)) }

// Before reports whether d is before other.
func (d Date) Before(other Date) bool { return d.In(time.UTC).Before(other.In(time.UTC)) }

// Weekday returns the day of the week.
func (d Date) Weekday() time.Weekday { return d.In(time.UTC).Weekday() }

// MarshalText implements encoding.TextMarshaler.
func (d Date) MarshalText() ([]byte, error) {
	if d.IsZero() {
		return []byte{}, nil
	}
	return []byte(d.String()), nil
}

// UnmarshalText implements encoding.TextUnmarshaler; an empty string is the zero date.
func (d *Date) UnmarshalText(b []byte) error {
	if len(b) == 0 {
		*d = Date{}
		return nil
	}
	parsed, err := ParseDate(string(b))
	if err != nil {
		return err
	}
	*d = parsed
	return nil
}
