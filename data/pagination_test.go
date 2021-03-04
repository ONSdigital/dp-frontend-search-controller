package data

import (
	"context"
	"net/http/httptest"
	"strconv"

	"testing"

	errs "github.com/ONSdigital/dp-frontend-search-controller/apperrors"
	"github.com/ONSdigital/dp-frontend-search-controller/config"
	. "github.com/smartystreets/goconvey/convey"
)

func TestUnitUpdateQueryWithOffset(t *testing.T) {
	t.Parallel()

	Convey("When updateQueryWithOffset called", t, func() {
		ctx := context.Background()

		cfg, err := config.Get()
		So(err, ShouldBeNil)

		pagination := &PaginationQuery{
			Limit:       10,
			CurrentPage: 1,
		}

		Convey("successfully update query with offset", func() {
			req := httptest.NewRequest("GET", "/search?q=housing&limit=10&page=1", nil)
			query := req.URL.Query()

			updateQueryWithOffset(ctx, cfg, pagination, query)

			encodedQuery := query.Encode()
			So(encodedQuery, ShouldContainSubstring, "offset=0")
			So(encodedQuery, ShouldNotContainSubstring, "page=")
		})

		Convey("return error", func() {

			Convey("when unable to get offset", func() {
				req := httptest.NewRequest("GET", "/search?q=housing&limit=10&page=1", nil)
				query := req.URL.Query()

				updateQueryWithOffset(ctx, cfg, pagination, query)

				encodedQuery := query.Encode()
				So(encodedQuery, ShouldContainSubstring, "offset=0")
				So(encodedQuery, ShouldNotContainSubstring, "page=")
			})
		})

	})

	Convey("When getOffset called", t, func() {
		ctx := context.Background()

		cfg, err := config.Get()
		So(err, ShouldBeNil)

		pagination := &PaginationQuery{
			Limit:       10,
			CurrentPage: 1,
		}

		Convey("successfully get valid offset", func() {

			Convey("when calculated offset is more than 0 and less than maximum search results", func() {
				offset, err := getOffset(ctx, cfg, pagination)

				So(offset, ShouldEqual, 0)
				So(err, ShouldBeNil)
			})

			Convey("when offset is less than 0", func() {
				pagination.CurrentPage = -1

				offset, err := getOffset(ctx, cfg, pagination)

				So(offset, ShouldEqual, cfg.DefaultOffset)
				So(err, ShouldBeNil)
			})
		})

		Convey("return error", func() {

			Convey("when the (offset-limit) exceeds maximum search results", func() {
				pagination.CurrentPage = 1000

				offset, err := getOffset(ctx, cfg, pagination)

				So(offset, ShouldEqual, cfg.DefaultOffset)
				So(err, ShouldNotBeNil)
				So(err, ShouldResemble, errs.ErrInvalidPage)
			})
		})
	})
}

func TestUnitReviewPagination(t *testing.T) {
	t.Parallel()

	Convey("When ReviewPagination called", t, func() {
		ctx := context.Background()

		cfg, err := config.Get()
		So(err, ShouldBeNil)

		req := httptest.NewRequest("GET", "/search?q=housing", nil)
		query := req.URL.Query()

		Convey("successfully review pagination query", func() {
			paginationQuery := ReviewPagination(ctx, cfg, query)
			So(paginationQuery.Limit, ShouldEqual, cfg.DefaultLimit)
			So(paginationQuery.CurrentPage, ShouldEqual, cfg.DefaultPage)
		})
	})

	Convey("When getPage called", t, func() {
		ctx := context.Background()

		cfg, err := config.Get()
		So(err, ShouldBeNil)

		Convey("successfully get page", func() {

			Convey("when valid current page given", func() {
				req := httptest.NewRequest("GET", "/search?q=housing&page=2", nil)
				query := req.URL.Query()

				page := getPage(ctx, cfg, query)

				So(page, ShouldEqual, 2)
				So(query.Get("page"), ShouldEqual, strconv.Itoa(2))
			})

			Convey("when current page is not an integer", func() {
				req := httptest.NewRequest("GET", "/search?q=housing&page=INVALID", nil)
				query := req.URL.Query()

				page := getPage(ctx, cfg, query)

				So(page, ShouldEqual, cfg.DefaultPage)
				So(query.Get("page"), ShouldEqual, strconv.Itoa(cfg.DefaultPage))
			})

			Convey("when current page is less than 1", func() {
				req := httptest.NewRequest("GET", "/search?q=housing&page=0", nil)
				query := req.URL.Query()

				page := getPage(ctx, cfg, query)

				So(page, ShouldEqual, cfg.DefaultPage)
				So(query.Get("page"), ShouldEqual, strconv.Itoa(cfg.DefaultPage))
			})
		})
	})

	Convey("When getLimit called", t, func() {
		ctx := context.Background()

		cfg, err := config.Get()
		So(err, ShouldBeNil)

		Convey("successfully get limit", func() {

			Convey("when valid limit given", func() {
				req := httptest.NewRequest("GET", "/search?q=housing&limit=20", nil)
				query := req.URL.Query()

				limit := getLimit(ctx, cfg, query)

				So(limit, ShouldEqual, 20)
				So(query.Get("limit"), ShouldEqual, strconv.Itoa(20))
			})

			Convey("when limit is not an integer", func() {
				req := httptest.NewRequest("GET", "/search?q=housing&limit=INVALID", nil)
				query := req.URL.Query()

				limit := getLimit(ctx, cfg, query)

				So(limit, ShouldEqual, cfg.DefaultLimit)
				So(query.Get("limit"), ShouldEqual, strconv.Itoa(cfg.DefaultLimit))
			})

			Convey("when limit is less than default limit", func() {
				req := httptest.NewRequest("GET", "/search?q=housing&limit=5", nil)
				query := req.URL.Query()

				limit := getLimit(ctx, cfg, query)

				So(limit, ShouldEqual, cfg.DefaultLimit)
				So(query.Get("limit"), ShouldEqual, strconv.Itoa(cfg.DefaultLimit))
			})

			Convey("when limit is more than default maximum limit", func() {
				req := httptest.NewRequest("GET", "/search?q=housing&limit=100000", nil)
				query := req.URL.Query()

				limit := getLimit(ctx, cfg, query)

				So(limit, ShouldEqual, cfg.DefaultMaximumLimit)
				So(query.Get("limit"), ShouldEqual, strconv.Itoa(cfg.DefaultMaximumLimit))
			})
		})
	})
}
