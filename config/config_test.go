package config

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestConfig(t *testing.T) {
	Convey("Given an environment with no environment variables set", t, func() {
		cfg, err := Get()

		Convey("When the config values are retrieved", func() {

			Convey("Then there should be no error returned", func() {
				So(err, ShouldBeNil)
			})

			Convey("Then the values should be set to the expected defaults", func() {
				cfg, err = Get() // This Get() is only called once, when inside this function
				So(err, ShouldBeNil)

				So(cfg.APIRouterURL, ShouldEqual, "http://localhost:23200/v1")
				So(cfg.BindAddr, ShouldEqual, "localhost:25000")
				So(cfg.CacheTopicsUpdateInterval, ShouldEqual, 30*time.Minute)
				So(cfg.Debug, ShouldBeFalse)
				So(cfg.DefaultLimit, ShouldEqual, 10)
				So(cfg.DefaultMaximumLimit, ShouldEqual, 50)
				So(cfg.DefaultMaximumSearchResults, ShouldEqual, 500)
				So(cfg.DefaultOffset, ShouldEqual, 0)
				So(cfg.DefaultPage, ShouldEqual, 1)
				So(cfg.DefaultSort, ShouldEqual, "relevance")
				So(cfg.EnableCensusTopicFilterOption, ShouldBeFalse)
				So(cfg.GracefulShutdownTimeout, ShouldEqual, 5*time.Second)
				So(cfg.HealthCheckCriticalTimeout, ShouldEqual, 90*time.Second)
				So(cfg.HealthCheckInterval, ShouldEqual, 30*time.Second)
				So(cfg.NoIndexEnabled, ShouldBeFalse)
				So(cfg.SiteDomain, ShouldEqual, "localhost")
				So(cfg.SupportedLanguages, ShouldResemble, []string{"en", "cy"})
			})

			Convey("Then a second call to config should return the same config", func() {
				newCfg, newErr := Get()
				So(newErr, ShouldBeNil)
				So(newCfg, ShouldResemble, cfg)
			})
		})
	})
}
