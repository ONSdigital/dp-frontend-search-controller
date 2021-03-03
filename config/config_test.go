package config

import (
	"os"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestConfig(t *testing.T) {
	os.Clearenv()
	var err error
	var config *Config

	Convey("Given an environment with no environment variables set", t, func() {
		Convey("Then cfg should be nil", func() {
			So(cfg, ShouldBeNil)
		})

		Convey("When the config values are retrieved", func() {

			Convey("Then there should be no error returned, and values are as expected", func() {
				config, err = Get() // This Get() is only called once, when inside this function
				So(err, ShouldBeNil)

				So(config.BindAddr, ShouldEqual, ":25000")
				So(config.RendererURL, ShouldEqual, "http://localhost:20010")
				So(config.SearchAPIURL, ShouldEqual, "http://localhost:23900")
				So(config.GracefulShutdownTimeout, ShouldEqual, 5*time.Second)
				So(config.HealthCheckInterval, ShouldEqual, 30*time.Second)
				So(config.HealthCheckCriticalTimeout, ShouldEqual, 90*time.Second)
				So(config.DefaultOffset, ShouldEqual, 0)
				So(config.DefaultSort, ShouldEqual, "relevance")
				So(config.DefaultPage, ShouldEqual, 1)
				So(config.DefaultLimit, ShouldEqual, 10)
				So(config.DefaultMaximumLimit, ShouldEqual, 50)
				So(config.DefaultMaximumSearchResults, ShouldEqual, 500)
			})

			Convey("Then a second call to config should return the same config", func() {
				newCfg, newErr := Get()
				So(newErr, ShouldBeNil)
				So(newCfg, ShouldResemble, config)
			})
		})
	})
}
