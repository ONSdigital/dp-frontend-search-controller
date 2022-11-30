package cache

import (
	"context"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

const englishLang string = "en"

func TestNewTopicCache(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	Convey("Given a valid cache update interval which is greater than 0", t, func() {
		updateCacheInterval := 1 * time.Millisecond

		Convey("When NewTopicCache is called", func() {
			testCache, err := NewTopicCache(ctx, &updateCacheInterval)

			Convey("Then a topic cache object should be successfully returned", func() {
				So(testCache, ShouldNotBeEmpty)

				Convey("And no error should be returned", func() {
					So(err, ShouldBeNil)
				})
			})
		})
	})

	Convey("Given no cache update interval (nil)", t, func() {
		Convey("When NewTopicCache is called", func() {
			testCache, err := NewTopicCache(ctx, nil)

			Convey("Then a cache object should be successfully returned", func() {
				So(testCache, ShouldNotBeEmpty)

				Convey("And no error should be returned", func() {
					So(err, ShouldBeNil)
				})
			})
		})
	})

	Convey("Given an invalid cache update interval which is less than or equal to 0", t, func() {
		updateCacheInterval := 0 * time.Second

		Convey("When NewTopicCache is called", func() {
			testCache, err := NewTopicCache(ctx, &updateCacheInterval)

			Convey("Then an error should be returned", func() {
				So(err, ShouldNotBeNil)

				Convey("And a nil cache object should be returned", func() {
					So(testCache, ShouldBeNil)
				})
			})
		})
	})
}

func TestGetData(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	mockCacheList, err := GetMockCacheList(ctx, englishLang)
	if err != nil {
		t.Error("failed to get mock census topic cache list")
	}

	Convey("Given a valid topic id which exists", t, func() {
		id := CensusTopicID

		Convey("When GetData is called", func() {
			testCacheData, err := mockCacheList.CensusTopic.GetData(ctx, id)

			Convey("Then a topic cache data should be successfully returned", func() {
				So(testCacheData, ShouldNotBeNil)

				Convey("And no error should be returned", func() {
					So(err, ShouldBeNil)
				})
			})
		})
	})

	Convey("Given an invalid topic id which also does not exist", t, func() {
		id := "invalid"

		Convey("When GetData is called", func() {
			testCacheData, err := mockCacheList.CensusTopic.GetData(ctx, id)

			Convey("Then an error should be returned", func() {
				So(err, ShouldNotBeNil)

				Convey("And the cache data returned should be empty", func() {
					So(testCacheData, ShouldResemble, getEmptyTopic())
				})
			})
		})
	})

	Convey("Given cached data is not type of *Topic", t, func() {
		id := "1234"

		mockCacheTopic, err := NewTopicCache(ctx, nil)
		So(err, ShouldBeNil)
		mockCacheTopic.Set(id, "test")

		Convey("When GetData is called", func() {
			testCacheData, err := mockCacheTopic.GetData(ctx, id)

			Convey("Then an error should be returned", func() {
				So(err, ShouldNotBeNil)

				Convey("And the cache data returned should be empty", func() {
					So(testCacheData, ShouldResemble, getEmptyTopic())
				})
			})
		})
	})

	Convey("Given cached data is not type of *Topic", t, func() {
		id := "1234"

		mockCacheTopic, err := NewTopicCache(ctx, nil)
		So(err, ShouldBeNil)
		mockCacheTopic.Set(id, nil)

		Convey("When GetData is called", func() {
			testCacheData, err := mockCacheTopic.GetData(ctx, id)

			Convey("Then an error should be returned", func() {
				So(err, ShouldNotBeNil)

				Convey("And the cache data returned should be empty", func() {
					So(testCacheData, ShouldResemble, getEmptyTopic())
				})
			})
		})
	})
}

func TestGetCensusData(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	mockCacheList, err := GetMockCacheList(ctx, englishLang)
	if err != nil {
		t.Error("failed to get mock census topic cache list")
	}

	Convey("Given the census data exists in cache", t, func() {
		Convey("When GetCensusData is called", func() {
			censusData, err := mockCacheList.CensusTopic.GetCensusData(ctx)

			Convey("Then census topic cache data should be successfully returned", func() {
				So(censusData, ShouldNotBeNil)

				Convey("And no error should be returned", func() {
					So(err, ShouldBeNil)
				})
			})
		})
	})

	Convey("Given the census data does not exist in cache", t, func() {
		mockCacheTopic, err := NewTopicCache(ctx, nil)
		So(err, ShouldBeNil)

		Convey("When GetCensusData is called", func() {
			censusData, err := mockCacheTopic.GetCensusData(ctx)

			Convey("Then an error should be returned", func() {
				So(err, ShouldNotBeNil)

				Convey("And the census cache data returned should be empty", func() {
					So(censusData, ShouldResemble, GetEmptyCensusTopic())
				})
			})
		})
	})
}

func TestAddUpdateFunc(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	Convey("Given an update function to update a topic", t, func() {
		mockTopicCache, err := NewTopicCache(ctx, nil)
		So(err, ShouldBeNil)

		topicUpdateFunc := func() *Topic {
			return &Topic{
				ID:              "test",
				LocaliseKeyName: "Test",
				Query:           "2453,1232",
			}
		}

		Convey("When AddUpdateFunc is called", func() {
			mockTopicCache.AddUpdateFunc("test", topicUpdateFunc)

			Convey("Then the update function is added to the cache", func() {
				So(mockTopicCache.UpdateFuncs["test"], ShouldNotBeEmpty)
			})
		})
	})
}

func TestGetEmptyCensusTopic(t *testing.T) {
	t.Parallel()

	Convey("When GetEmptyCensusTopic is called", t, func() {
		emptyCensusTopic := GetEmptyCensusTopic()

		Convey("Then an empty census topic should be returned", func() {
			So(emptyCensusTopic.ID, ShouldEqual, CensusTopicID)
			So(emptyCensusTopic.LocaliseKeyName, ShouldEqual, "")
			So(emptyCensusTopic.Query, ShouldEqual, "")
			So(emptyCensusTopic.List, ShouldNotBeNil)
		})
	})
}
