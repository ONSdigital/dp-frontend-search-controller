package data

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/ONSdigital/dp-frontend-search-controller/config"
	. "github.com/smartystreets/goconvey/convey"
)

func TestUnitReviewSort(t *testing.T) {
	t.Parallel()

	Convey("When ReviewSort called", t, func() {
		ctx := context.Background()

		cfg, err := config.Get()
		So(err, ShouldBeNil)

		Convey("successfully review sort", func() {

			Convey("when valid sort given", func() {
				req := httptest.NewRequest("GET", "/search?q=housing&sort=release_date", nil)
				query := req.URL.Query()

				ReviewSort(ctx, cfg, query)

				So(query.Get("sort"), ShouldEqual, "release_date")
			})

			Convey("when invalid sort given", func() {
				req := httptest.NewRequest("GET", "/search?q=housing&sort=INVALID", nil)
				query := req.URL.Query()

				ReviewSort(ctx, cfg, query)

				So(query.Get("sort"), ShouldEqual, cfg.DefaultSort)
			})
		})
	})
}
