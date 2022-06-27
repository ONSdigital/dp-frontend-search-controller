package dpcache

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func getTestCache(updateInterval *time.Duration) *Cache {
	testCache := Cache{
		data:           sync.Map{},
		updateInterval: updateInterval,
		close:          make(chan struct{}),
		UpdateFuncs:    make(map[string]func() (interface{}, error)),
	}

	testCache.data.Store("string", "test")
	testCache.data.Store("int", 1)
	testCache.data.Store("bool", false)
	testCache.data.Store("float", 1.1)

	testCache.UpdateFuncs["string"] = func() (interface{}, error) {
		val, ok := testCache.Get("string")

		// the first update
		if ok && val == "test" {
			return "test2", nil
		}

		// the second update or more
		return "test3", nil
	}
	testCache.UpdateFuncs["int"] = func() (interface{}, error) {
		val, ok := testCache.Get("int")
		if ok && val == 1 {
			return 2, nil
		}
		return 3, nil
	}
	testCache.UpdateFuncs["bool"] = func() (interface{}, error) {
		val, ok := testCache.Get("bool")
		if ok && val == false {
			return true, nil
		}
		return false, nil
	}
	testCache.UpdateFuncs["float"] = func() (interface{}, error) {
		val, ok := testCache.Get("float")
		if ok && val == 1.1 {
			return 2.2, nil
		}
		return 3.3, nil
	}

	return &testCache
}

func TestNewCache(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	Convey("Given a valid cache update interval which is greater than 0", t, func() {
		updateCacheInterval := 1 * time.Millisecond

		Convey("When NewCache is called", func() {
			testCache, err := NewCache(ctx, &updateCacheInterval)

			Convey("Then a cache object should be successfully returned", func() {
				So(testCache, ShouldNotBeEmpty)

				Convey("And no error should be returned", func() {
					So(err, ShouldBeNil)
				})
			})
		})
	})

	Convey("Given no cache update interval (nil)", t, func() {
		Convey("When NewCache is called", func() {
			testCache, err := NewCache(ctx, nil)

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

		Convey("When NewCache is called", func() {
			testCache, err := NewCache(ctx, &updateCacheInterval)

			Convey("Then an error should be returned", func() {
				So(err, ShouldNotBeNil)
				So(err, ShouldResemble, errors.New("cache update interval duration is less than or equal to 0"))

				Convey("And a nil cache object should be returned", func() {
					So(testCache, ShouldBeNil)
				})
			})
		})
	})
}

func TestClose(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	errorChan := make(chan error, 1)

	Convey("Given cache is already updating", t, func() {
		updateCacheInterval := 1 * time.Millisecond
		testCache := getTestCache(&updateCacheInterval)

		go testCache.StartUpdates(ctx, errorChan)

		Convey("When Close is called", func() {
			testCache.Close()

			Convey("Then all the values of the cache data should be emptied", func() {
				cacheStringValue, ok := testCache.Get("string")
				So(cacheStringValue, ShouldBeEmpty)
				So(ok, ShouldBeTrue)

				cacheIntValue, ok := testCache.Get("int")
				So(cacheIntValue, ShouldBeEmpty)
				So(ok, ShouldBeTrue)

				cacheBoolValue, ok := testCache.Get("bool")
				So(cacheBoolValue, ShouldBeEmpty)
				So(ok, ShouldBeTrue)

				cacheFloatValue, ok := testCache.Get("float")
				So(cacheFloatValue, ShouldBeEmpty)
				So(ok, ShouldBeTrue)

				Convey("And update functions in cache should be emptied", func() {
					So(testCache.UpdateFuncs, ShouldBeEmpty)
				})
			})
		})
	})

	Convey("Given cache is not set to update in intervals", t, func() {
		testCache := getTestCache(nil)

		go testCache.StartUpdates(ctx, errorChan)

		Convey("When Close is called", func() {
			testCache.Close()

			Convey("Then this function does nothing to the cache as StartUpdates go-routine was stopped beforehand", func() {})
		})
	})
}

func TestAddUpdateFunc(t *testing.T) {
	t.Parallel()

	Convey("Given an update function", t, func() {
		testCache := getTestCache(nil)
		updateFunc := func() (interface{}, error) {
			return "test", nil
		}

		Convey("When AddUpdateFunc is called", func() {
			testCache.AddUpdateFunc("test", updateFunc)

			Convey("Then the update function is added to the cache", func() {
				So(testCache.UpdateFuncs["test"], ShouldNotBeEmpty)
				So(testCache.UpdateFuncs["test"], ShouldEqual, updateFunc)
			})
		})
	})
}

func TestUpdateContent(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	Convey("Given a cache", t, func() {
		updateCacheInterval := 1 * time.Millisecond
		testCache := getTestCache(&updateCacheInterval)

		Convey("When UpdateContent is called", func() {
			err := testCache.UpdateContent(ctx)

			Convey("Then no error should be returned", func() {
				So(err, ShouldBeNil)

				Convey("And all the values of the cache data should be updated", func() {
					cacheStringValue, ok := testCache.Get("string")
					So(cacheStringValue, ShouldEqual, "test2")
					So(ok, ShouldBeTrue)

					cacheIntValue, ok := testCache.Get("int")
					So(cacheIntValue, ShouldEqual, 2)
					So(ok, ShouldBeTrue)

					cacheBoolValue, ok := testCache.Get("bool")
					So(cacheBoolValue, ShouldBeTrue)
					So(ok, ShouldBeTrue)

					cacheFloatValue, ok := testCache.Get("float")
					So(cacheFloatValue, ShouldEqual, 2.2)
					So(ok, ShouldBeTrue)
				})
			})
		})
	})

	Convey("Given an update function which causes an error for cache", t, func() {
		updateCacheInterval := 1 * time.Millisecond
		testCache := getTestCache(&updateCacheInterval)

		testCache.UpdateFuncs = make(map[string]func() (interface{}, error))
		testCache.UpdateFuncs["error_update_func"] = func() (interface{}, error) {
			return nil, errors.New("unexpected error")
		}

		Convey("When UpdateContent is called", func() {
			err := testCache.UpdateContent(ctx)

			Convey("Then an error should be returned", func() {
				So(err, ShouldNotBeNil)
				So(err, ShouldResemble, errors.New("failed to update search cache for error_update_func. error: unexpected error"))
			})
		})
	})
}

func TestStartUpdates(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	errorChan := make(chan error, 1)

	Convey("Given at initial cache setup with update interval set", t, func() {
		updateCacheInterval := 20 * time.Millisecond
		testCache := getTestCache(&updateCacheInterval)

		Convey("When StartUpdates is called", func() {
			go testCache.StartUpdates(ctx, errorChan)

			Convey("Then cache data should be updated immediately", func() {
				// give time for go-routine to update in test but this time is less than the update interval
				time.Sleep(1 * time.Millisecond)

				cacheStringValue, ok := testCache.Get("string")
				So(cacheStringValue, ShouldEqual, "test2")
				So(ok, ShouldBeTrue)

				cacheIntValue, ok := testCache.Get("int")
				So(cacheIntValue, ShouldEqual, 2)
				So(ok, ShouldBeTrue)

				cacheBoolValue, ok := testCache.Get("bool")
				So(cacheBoolValue, ShouldBeTrue)
				So(ok, ShouldBeTrue)

				cacheFloatValue, ok := testCache.Get("float")
				So(cacheFloatValue, ShouldEqual, 2.2)
				So(ok, ShouldBeTrue)

				Convey("And close cache to stop go-routine", func() {
					testCache.Close()
				})
			})
		})
	})

	Convey("Given cache is already set up with update interval set", t, func() {
		updateCacheInterval := 20 * time.Millisecond
		testCache := getTestCache(&updateCacheInterval)

		go testCache.StartUpdates(ctx, errorChan)

		Convey("When the updateInterval time has passed", func() {
			time.Sleep(updateCacheInterval)
			// give extra time for go-routine in test to update and this ensures the expected values to match
			time.Sleep(1 * time.Millisecond)

			Convey("Then cache data should be updated for the second time or more", func() {
				cacheStringValue, ok := testCache.Get("string")
				So(cacheStringValue, ShouldEqual, "test3")
				So(ok, ShouldBeTrue)

				cacheIntValue, ok := testCache.Get("int")
				So(cacheIntValue, ShouldEqual, 3)
				So(ok, ShouldBeTrue)

				cacheBoolValue, ok := testCache.Get("bool")
				So(cacheBoolValue, ShouldBeFalse)
				So(ok, ShouldBeTrue)

				cacheFloatValue, ok := testCache.Get("float")
				So(cacheFloatValue, ShouldEqual, 3.3)
				So(ok, ShouldBeTrue)

				Convey("And close cache to stop go-routine", func() {
					testCache.Close()
				})
			})
		})
	})

	Convey("Given no update functions for cache", t, func() {
		updateCacheInterval := 1 * time.Millisecond
		testCache := getTestCache(&updateCacheInterval)

		testCache.UpdateFuncs = make(map[string]func() (interface{}, error), 0)

		Convey("When StartUpdates is called", func() {
			testCache.StartUpdates(ctx, errorChan)

			Convey("Then cache data should not be updated", func() {
				cacheStringValue, ok := testCache.Get("string")
				So(cacheStringValue, ShouldEqual, "test")
				So(ok, ShouldBeTrue)

				cacheIntValue, ok := testCache.Get("int")
				So(cacheIntValue, ShouldEqual, 1)
				So(ok, ShouldBeTrue)

				cacheBoolValue, ok := testCache.Get("bool")
				So(cacheBoolValue, ShouldBeFalse)
				So(ok, ShouldBeTrue)

				cacheFloatValue, ok := testCache.Get("float")
				So(cacheFloatValue, ShouldEqual, 1.1)
				So(ok, ShouldBeTrue)
			})
		})
	})

	Convey("Given a cache with no update interval but has update functions", t, func() {
		testCache := getTestCache(nil)

		Convey("When StartUpdates is called", func() {
			testCache.StartUpdates(ctx, errorChan)

			Convey("Then cache data should be updated/loaded once", func() {
				cacheStringValue, ok := testCache.Get("string")
				So(cacheStringValue, ShouldEqual, "test2")
				So(ok, ShouldBeTrue)

				cacheIntValue, ok := testCache.Get("int")
				So(cacheIntValue, ShouldEqual, 2)
				So(ok, ShouldBeTrue)

				cacheBoolValue, ok := testCache.Get("bool")
				So(cacheBoolValue, ShouldBeTrue)
				So(ok, ShouldBeTrue)

				cacheFloatValue, ok := testCache.Get("float")
				So(cacheFloatValue, ShouldEqual, 2.2)
				So(ok, ShouldBeTrue)
			})
		})
	})
}
