package cache

import (
	"context"
	"fmt"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

const testLang = "en"

func TestGetMockCacheList(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	Convey("When GetMockCacheList is called", t, func() {
		cacheList, err := GetMockCacheList(ctx, testLang)

		Convey("Then the list of cache should be returned", func() {
			So(cacheList, ShouldNotBeNil)
			So(err, ShouldBeNil)

			So(cacheList.CensusTopic, ShouldNotBeNil)
			So(cacheList.Navigation, ShouldNotBeNil)

			censusTopic, err := cacheList.CensusTopic.GetCensusData(ctx)
			So(censusTopic, ShouldNotBeNil)
			So(err, ShouldBeNil)
		})
	})
}

func TestGetMockCensusTopic(t *testing.T) {
	t.Parallel()

	Convey("When GetMockCensusTopic is called", t, func() {
		mockCensusTopic := GetMockCensusTopic()

		Convey("Then the mock census topic is returned", func() {
			So(mockCensusTopic, ShouldNotBeNil)
			So(mockCensusTopic.ID, ShouldEqual, CensusTopicID)
			So(mockCensusTopic.LocaliseKeyName, ShouldEqual, "Census")
			So(mockCensusTopic.Query, ShouldEqual, fmt.Sprintf("1234,5678,%s", CensusTopicID))

			subtopic, exists := mockCensusTopic.List.Get("1234")
			So(exists, ShouldBeTrue)
			So(subtopic, ShouldResemble, Subtopic{ID: "1234", LocaliseKeyName: "International Migration", ReleaseDate: timeHelper("2022-10-10T08:30:00Z")})

			subtopic, exists = mockCensusTopic.List.Get("5678")
			So(exists, ShouldBeTrue)
			So(subtopic, ShouldResemble, Subtopic{ID: "5678", LocaliseKeyName: "Age", ReleaseDate: timeHelper("2022-11-09T09:30:00Z")})

			subtopic, exists = mockCensusTopic.List.Get(CensusTopicID)
			So(exists, ShouldBeTrue)
			So(subtopic, ShouldResemble, Subtopic{ID: CensusTopicID, LocaliseKeyName: "Census", ReleaseDate: timeHelper("2022-10-10T09:30:00Z")})
		})
	})
}

func TestGetMockRootTopic(t *testing.T) {
	t.Parallel()

	rootTopicID := "root-topic-cache"
	slug := "root"

	Convey("When GetMockRootTopic is called", t, func() {
		mockTopic := GetMockRootTopic(rootTopicID)

		Convey("Then the mocked root topic is returned", func() {
			So(mockTopic, ShouldNotBeNil)
			So(mockTopic.ID, ShouldEqual, rootTopicID)
			So(mockTopic.Slug, ShouldEqual, slug)

			subtopic, exists := mockTopic.List.Get("economy")
			So(exists, ShouldBeTrue)
			So(subtopic, ShouldResemble, Subtopic{ID: "6734", Slug: "economy", LocaliseKeyName: "Economy", ReleaseDate: timeHelper("2022-10-10T08:30:00Z"), ParentID: ""})
		})
	})
}
