package cache

import (
	"sync"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

var (
	subtopic1 = Subtopic{
		LocaliseKeyName: "International Migration",
		ReleaseDate:     timeHelper("2022-10-10T08:30:00Z"),
	}
	subtopic2 = Subtopic{
		LocaliseKeyName: "Age",
		ReleaseDate:     timeHelper("2022-11-09T09:30:00Z"),
	}
	subtopic3 = Subtopic{
		LocaliseKeyName: "Health",
		ReleaseDate:     timeHelper("2023-01-07T09:30:00Z"),
	}
)

func TestAppendSubtopicID(t *testing.T) {
	t.Parallel()

	Convey("Given an empty SubtopicsIDs object", t, func() {
		subtopicIDsStore := NewSubTopicsMap()

		Convey("When AppendSubtopicID is called", func() {
			subtopicIDsStore.AppendSubtopicID("1234", subtopic1)

			Convey("Then the new subtopic id should be added to the map", func() {
				So(subtopicIDsStore.Get("1234"), ShouldResemble, subtopic1)
			})
		})
	})

	Convey("Given a nil map in the SubtopicsIDs object", t, func() {
		subtopicIDsStore := NewSubTopicsMap()
		subtopicIDsStore.subtopicsMap = nil

		Convey("When AppendSubtopicID is called", func() {
			subtopicIDsStore.AppendSubtopicID("1234", subtopic1)

			Convey("Then the new subtopic id should be added to the map", func() {
				So(subtopicIDsStore.Get("1234"), ShouldResemble, subtopic1)
			})
		})
	})

	Convey("Given an existing SubtopicsIDs object with data", t, func() {
		subtopicIDsStore := NewSubTopicsMap()
		subtopicIDsStore.subtopicsMap = map[string]Subtopic{
			"1234": subtopic1,
		}

		Convey("When AppendSubtopicID is called", func() {
			subtopicIDsStore.AppendSubtopicID("5678", subtopic2)

			Convey("Then the new subtopic id should be added to the map", func() {
				So(subtopicIDsStore.Get("5678"), ShouldResemble, subtopic2)
			})

			Convey("And the existing subtopic `1234` should still exist in map", func() {
				So(subtopicIDsStore.Get("1234"), ShouldResemble, subtopic1)
			})
		})
	})

	Convey("Given AppendSubtopicID is called synchronously", t, func() {
		subtopicIDsStore := NewSubTopicsMap()

		Convey("When AppendSubtopicID is called", func() {
			go subtopicIDsStore.AppendSubtopicID("5678", subtopic2)
			go subtopicIDsStore.AppendSubtopicID("9123", subtopic3)

			Convey("Then the new subtopic ids should be added", func() {
				// Wait for the goroutines to finish
				time.Sleep(300 * time.Millisecond)

				So(subtopicIDsStore.Get("5678"), ShouldResemble, subtopic2)
				So(subtopicIDsStore.Get("9123"), ShouldResemble, subtopic3)
			})
		})
	})
}

func TestGetSubtopicsIDsQuery(t *testing.T) {
	t.Parallel()

	Convey("Given an empty list of subtopics", t, func() {
		subtopicIDsStore := NewSubTopicsMap()

		Convey("When GetSubtopicsIDsQuery is called", func() {
			subtopicsIDQuery := subtopicIDsStore.GetSubtopicsIDsQuery()

			Convey("Then subtopic ids query should be empty", func() {
				So(subtopicsIDQuery, ShouldBeEmpty)
			})
		})
	})

	Convey("Given a list of subtopics", t, func() {
		subtopicIDsStore := Subtopics{
			mutex: &sync.RWMutex{},
			subtopicsMap: map[string]Subtopic{
				"1234": subtopic1,
				"5678": subtopic2,
			},
		}

		Convey("When GetSubtopicsIDsQuery is called", func() {
			subtopicsIDQuery := subtopicIDsStore.GetSubtopicsIDsQuery()

			Convey("Then subtopic ids query should be returned successfully", func() {
				So(subtopicsIDQuery, ShouldContainSubstring, "1234")
				So(subtopicsIDQuery, ShouldContainSubstring, "5678")
			})
		})
	})
}
