package cache

import (
	"context"
	"fmt"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestGetMockCacheList(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	Convey("When GetMockCacheList is called", t, func() {
		lang := "en"
		cacheList, err := GetMockCacheList(ctx, lang)

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
			So(mockCensusTopic.List.Get("1234"), ShouldBeTrue)
			So(mockCensusTopic.List.Get("5678"), ShouldBeTrue)
			So(mockCensusTopic.List.Get(CensusTopicID), ShouldBeTrue)
		})
	})
}
