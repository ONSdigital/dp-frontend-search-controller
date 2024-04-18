package data

import (
	"strconv"
	"time"
)

type Date struct {
	date       time.Time
	ys, ms, ds string
	fieldsetErrID, fieldsetStr                                       string
	assumedDay, assumedMonth                                         bool
	hasDayValidationErr, hasMonthValidationErr, hasYearValidationErr bool
}

const DateFormat = "2006-01-02"

// MustParseDate checks if the date format is correct and parsable
func MustParseDate(dateAsString string) Date {
	d, err := ParseDate(dateAsString)
	if err != nil {
		panic("invalid date string: " + dateAsString)
	}

	return d
}

func MustSetFieldsetErrID(id string) (d Date) {
	d.fieldsetErrID = id
	return d
}

// ParseDate parses the date into the default DateFormat
func ParseDate(dateAsString string) (Date, error) {
	if dateAsString == "" {
		return Date{}, nil
	}
	t, err := time.Parse(DateFormat, dateAsString)
	if err != nil {
		return Date{}, err
	}

	return DateFromTime(t), nil
}

// DateFromTime extracts a date from a time object
func DateFromTime(t time.Time) Date {
	if t.IsZero() {
		return Date{}
	}
	date := Date{date: t}
	y, m, d := t.Date()
	date.ys, date.ms, date.ds = strconv.Itoa(y), strconv.Itoa(int(m)), strconv.Itoa(d)
	return date
}

func (d Date) String() string {
	if d.date.IsZero() {
		return ""
	}

	return d.date.UTC().Format(DateFormat)
}

func (d Date) YearString() string {
	return d.ys
}

func (d Date) MonthString() string {
	return d.ms
}

func (d Date) DayString() string {
	return d.ds
}

func (d Date) GetFieldsetErrID() string {
	return d.fieldsetErrID
}

func (d Date) HasDayValidationErr() bool {
	return d.hasDayValidationErr
}

func (d Date) HasMonthValidationErr() bool {
	return d.hasMonthValidationErr
}

func (d Date) HasYearValidationErr() bool {
	return d.hasYearValidationErr
}