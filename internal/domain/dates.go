package domain

import (
	"errors"
	"fmt"
	"time"
)

// Date is a civil date in the user's zone, with no time of day. The history
// filter and the receipt range are stated in dates (FR-021, FR-040).
type Date struct {
	Year  int
	Month time.Month
	Day   int
}

// dateLayout is the ISO form a date crosses the window boundary in.
const dateLayout = "2006-01-02"

// localLayout is the form an occurrence time crosses the window boundary in:
// what an HTML datetime-local input produces, to the minute.
const localLayout = "2006-01-02T15:04"

// displayLayout is how the application shows an instant outside the receipt:
// in the history and in every message naming an event.
const displayLayout = "02 Jan 2006 15:04"

// FormatDisplay writes an instant as the history shows it.
func FormatDisplay(instant time.Time, zone *time.Location) string {
	return instant.In(zone).Format(displayLayout)
}

// ErrBadDate refuses a date or a time that cannot be read.
var ErrBadDate = errors.New("not a date")

// ErrReversedRange refuses a range whose end comes before its start.
var ErrReversedRange = errors.New("the range ends before it starts")

// DateOf answers the civil date of an instant in the given zone.
func DateOf(instant time.Time, zone *time.Location) Date {
	year, month, day := instant.In(zone).Date()
	return Date{Year: year, Month: month, Day: day}
}

// ParseDate reads an ISO date such as 2026-09-22.
func ParseDate(text string) (Date, error) {
	parsed, err := time.Parse(dateLayout, text)
	if err != nil {
		return Date{}, fmt.Errorf("%w: %q", ErrBadDate, text)
	}
	return DateOf(parsed, time.UTC), nil
}

// ParseLocal reads an occurrence time such as 2026-09-22T12:30 in the given zone.
func ParseLocal(text string, zone *time.Location) (time.Time, error) {
	parsed, err := time.ParseInLocation(localLayout, text, zone)
	if err != nil {
		return time.Time{}, fmt.Errorf("%w: %q", ErrBadDate, text)
	}
	return parsed, nil
}

// FormatLocal writes an instant in the form ParseLocal reads.
func FormatLocal(instant time.Time, zone *time.Location) string {
	return instant.In(zone).Format(localLayout)
}

// IsZero reports whether the date is unset, which leaves that end of a range
// open.
func (d Date) IsZero() bool { return d == Date{} }

// String writes the date in ISO form.
func (d Date) String() string { return d.start(time.UTC).Format(dateLayout) }

// Compare answers -1, 0 or 1 as d is before, equal to or after other.
func (d Date) Compare(other Date) int {
	return d.start(time.UTC).Compare(other.start(time.UTC))
}

// start answers the first instant of the date in a zone.
func (d Date) start(zone *time.Location) time.Time {
	return time.Date(d.Year, d.Month, d.Day, 0, 0, 0, 0, zone)
}

// CheckRange refuses a range whose two set ends are the wrong way round.
func CheckRange(from, to Date) error {
	if !from.IsZero() && !to.IsZero() && from.Compare(to) > 0 {
		return ErrReversedRange
	}
	return nil
}
