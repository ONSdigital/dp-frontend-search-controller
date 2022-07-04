package cache

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestAppendSubtopicID(t *testing.T) {
	t.Parallel()

	Convey("Given an empty SubtopicsIDs object", t, func() {
		subtopicIDsStore := NewSubTopicsMap()

		Convey("When AppendSubtopicID is called", func() {
			subtopicIDsStore.AppendSubtopicID("1234")

			Convey("Then the new subtopic id should be added to the map", func() {
				So(subtopicIDsStore.Get("1234"), ShouldBeTrue)
			})
		})
	})

	Convey("Given a nil map in the SubtopicsIDs object", t, func() {
		subtopicIDsStore := NewSubTopicsMap()
		subtopicIDsStore.subtopicsMap = nil

		Convey("When AppendSubtopicID is called", func() {
			subtopicIDsStore.AppendSubtopicID("1234")

			Convey("Then the new subtopic id should be added to the map", func() {
				So(subtopicIDsStore.Get("1234"), ShouldBeTrue)
			})
		})
	})

	Convey("Given an existing SubtopicsIDs object with data", t, func() {
		subtopicIDsStore := NewSubTopicsMap()
		subtopicIDsStore.subtopicsMap = map[string]bool{
			"1234": true,
		}

		Convey("When AppendSubtopicID is called", func() {
			subtopicIDsStore.AppendSubtopicID("5678")

			Convey("Then the new subtopic id should be added to the map", func() {
				So(subtopicIDsStore.Get("5678"), ShouldBeTrue)
			})
		})
	})

	Convey("Given AppendSubtopicID is called synchronously", t, func() {
		subtopicIDsStore := NewSubTopicsMap()

		Convey("When AppendSubtopicID is called", func() {
			go subtopicIDsStore.AppendSubtopicID("5678")
			go subtopicIDsStore.AppendSubtopicID("9123")

			Convey("Then the new subtopic ids should be added", func() {
				// Wait for the goroutines to finish
				time.Sleep(300 * time.Millisecond)

				So(subtopicIDsStore.Get("5678"), ShouldBeTrue)
				So(subtopicIDsStore.Get("9123"), ShouldBeTrue)
			})
		})
	})
}

func TestGetSubtopicsIDsQuery(t *testing.T) {
	t.Parallel()

	Convey("Given an empty list of subtopics", t, func() {
		subtopicIDsStore := SubtopicsIDs{}

		Convey("When GetSubtopicsIDsQuery is called", func() {
			subtopicsIDQuery := subtopicIDsStore.GetSubtopicsIDsQuery()

			Convey("Then subtopic ids query should be empty", func() {
				So(subtopicsIDQuery, ShouldBeEmpty)
			})
		})
	})

	Convey("Given a list of subtopics", t, func() {
		subtopicIDsStore := SubtopicsIDs{
			subtopicsMap: map[string]bool{
				"1234": true,
				"5678": true,
			},
		}

		Convey("When GetSubtopicsIDsQuery is called", func() {
			subtopicsIDQuery := subtopicIDsStore.GetSubtopicsIDsQuery()

			Convey("Then subtopic ids query should be returned successfully in increasing order", func() {
				So(subtopicsIDQuery, ShouldEqual, "1234,5678")
			})
		})
	})
}
