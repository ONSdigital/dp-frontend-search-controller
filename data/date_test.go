package data

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestParseDate(t *testing.T) {
	Convey("Given a valid date as string", t, func() {
		dateString := "2022-11-08"
		Convey("Then ParseDate() returns a valid date", func() {
			date, err := ParseDate(dateString)
			So(err, ShouldBeNil)
			So(date.DayString(), ShouldEqual, "8")
			So(date.MonthString(), ShouldEqual, "11")
			So(date.YearString(), ShouldEqual, "2022")
			So(date.String(), ShouldEqual, dateString)
		})
	})

	Convey("Given an empty string", t, func() {
		dateString := ""
		Convey("Then ParseDate() returns an empty date", func() {
			date, err := ParseDate(dateString)
			So(err, ShouldBeNil)
			So(date, ShouldResemble, Date{})
		})
	})

	Convey("Given an invalid date as string", t, func() {
		dateString := "2022"
		Convey("Then ParseDate() returns an error", func() {
			date, err := ParseDate(dateString)
			So(err, ShouldNotBeNil)
			So(date, ShouldResemble, Date{})
		})
	})
}

func TestMustParseDate(t *testing.T) {
	Convey("Given a valid date as string", t, func() {
		dateString := "2012-03-28"
		Convey("Then ParseDate() returns a valid date", func() {
			date := MustParseDate(dateString)
			So(date.DayString(), ShouldEqual, "28")
			So(date.MonthString(), ShouldEqual, "3")
			So(date.YearString(), ShouldEqual, "2012")
			So(date.String(), ShouldEqual, dateString)
		})
	})

	Convey("Given an empty string", t, func() {
		dateString := ""
		Convey("Then MustParseDate() returns an empty date", func() {
			date := MustParseDate(dateString)
			So(date, ShouldResemble, Date{})
		})
	})

	Convey("Given an invalid date as string", t, func() {
		dateString := "2022"
		Convey("Then MustParseDate() should panic", func() {
			So(func() { MustParseDate(dateString) }, ShouldPanic)
		})
	})
}

func TestDateFromTime(t *testing.T) {
	Convey("Given a valid time", t, func() {
		t := time.Date(2010, time.Month(6), 26, 16, 47, 51, 1, time.UTC)
		Convey("Then DateFromTime() returns a valid date", func() {
			date := DateFromTime(t)
			So(date.DayString(), ShouldEqual, "26")
			So(date.MonthString(), ShouldEqual, "6")
			So(date.YearString(), ShouldEqual, "2010")
			So(date.String(), ShouldEqual, "2010-06-26")
		})
	})

	Convey("Given an empty time", t, func() {
		t := time.Time{}
		Convey("Then DateFromTime() returns an empty date", func() {
			date := DateFromTime(t)
			So(date, ShouldResemble, Date{})
		})
	})
}
