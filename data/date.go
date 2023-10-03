package data

import (
	"strconv"
	"time"
)

type Date struct {
	date       time.Time
	ys, ms, ds string
}

const DateFormat = "2006-01-02"

func MustParseDate(dateAsString string) Date {
	d, err := ParseDate(dateAsString)
	if err != nil {
		panic("invalid date string: " + dateAsString)
	}

	return d
}

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
