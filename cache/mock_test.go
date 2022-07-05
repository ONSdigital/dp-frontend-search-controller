package cache

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestGetMockCensusTopicCacheList(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	Convey("When GetMockCensusTopicCacheList is called", t, func() {
		cacheList, err := GetMockCensusTopicCacheList(ctx)

		Convey("Then the list of cache should be returned", func() {
			So(cacheList, ShouldNotBeNil)
			So(err, ShouldBeNil)

			So(cacheList.CensusTopic, ShouldNotBeNil)

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
			So(mockCensusTopic.Query, ShouldEqual, "1234,5678")
			So(mockCensusTopic.List.Get("1234"), ShouldBeTrue)
			So(mockCensusTopic.List.Get("5678"), ShouldBeTrue)
		})
	})
}
