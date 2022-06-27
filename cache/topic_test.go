package cache

import (
	"context"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

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

func TestAddUpdateFunc(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	Convey("Given an update function to update a topic", t, func() {
		mockTopicCache, err := NewTopicCache(ctx, nil)
		So(err, ShouldBeNil)

		topicUpdateFunc := func() (*Topic, error) {
			return &Topic{
				ID:               "test",
				LocaliseKeyName:  "Test",
				SubtopicsIDQuery: "2453,1232",
			}, nil
		}

		Convey("When AddUpdateFunc is called", func() {
			mockTopicCache.AddUpdateFunc("test", topicUpdateFunc)

			Convey("Then the update function is added to the cache", func() {
				So(mockTopicCache.UpdateFuncs["test"], ShouldNotBeEmpty)
			})
		})
	})
}
