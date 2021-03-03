package data

import (
	"context"
	"net/http/httptest"
	"strconv"

	"testing"

	"github.com/ONSdigital/dp-frontend-search-controller/config"
	. "github.com/smartystreets/goconvey/convey"
)

func TestUnitQuery(t *testing.T) {
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

		Convey("successfully update query with offset if offset less than 0", func() {
			req := httptest.NewRequest("GET", "/search?q=housing&limit=10&page=1", nil)
			query := req.URL.Query()
			pagination.CurrentPage = -2
			updateQueryWithOffset(ctx, cfg, pagination, query)
			encodedQuery := query.Encode()
			So(encodedQuery, ShouldContainSubstring, "offset=0")
			So(encodedQuery, ShouldNotContainSubstring, "page=")
		})

		Convey("successfully update query with offset if offset more than maximum search results", func() {
			req := httptest.NewRequest("GET", "/search?q=housing&limit=10&page=1", nil)
			query := req.URL.Query()
			pagination.CurrentPage = 500
			updateQueryWithOffset(ctx, cfg, pagination, query)
			encodedQuery := query.Encode()
			So(encodedQuery, ShouldContainSubstring, "offset=489")
			So(encodedQuery, ShouldNotContainSubstring, "page=")
		})

	})

	Convey("When ReviewQuery called", t, func() {
		ctx := context.Background()
		cfg, err := config.Get()
		So(err, ShouldBeNil)

		Convey("successfully set default page to query if page not given", func() {
			req := httptest.NewRequest("GET", "/search?q=housing&limit=10&sort=relevance", nil)
			updatedURL, paginationQuery, err := ReviewQuery(ctx, cfg, req.URL)
			So(updatedURL.Query().Get("limit"), ShouldEqual, "10")
			So(updatedURL.Query().Get("page"), ShouldEqual, strconv.Itoa(cfg.DefaultPage))
			So(updatedURL.Query().Get("sort"), ShouldEqual, "relevance")
			So(paginationQuery.CurrentPage, ShouldEqual, cfg.DefaultPage)
			So(paginationQuery.Limit, ShouldEqual, 10)
			So(err, ShouldBeNil)
		})
		Convey("successfully set default page to query if invalid page given", func() {
			req := httptest.NewRequest("GET", "/search?q=housing&limit=10&sort=relevance&page=invalid", nil)
			updatedURL, paginationQuery, err := ReviewQuery(ctx, cfg, req.URL)
			So(updatedURL.Query().Get("limit"), ShouldEqual, "10")
			So(updatedURL.Query().Get("page"), ShouldEqual, strconv.Itoa(cfg.DefaultPage))
			So(updatedURL.Query().Get("sort"), ShouldEqual, "relevance")
			So(paginationQuery.CurrentPage, ShouldEqual, cfg.DefaultPage)
			So(paginationQuery.Limit, ShouldEqual, 10)
			So(err, ShouldBeNil)
		})
		Convey("successfully set default page to query if page less than 1", func() {
			req := httptest.NewRequest("GET", "/search?q=housing&limit=10&sort=relevance&page=0", nil)
			updatedURL, paginationQuery, err := ReviewQuery(ctx, cfg, req.URL)
			So(updatedURL.Query().Get("limit"), ShouldEqual, "10")
			So(updatedURL.Query().Get("page"), ShouldEqual, strconv.Itoa(cfg.DefaultPage))
			So(updatedURL.Query().Get("sort"), ShouldEqual, "relevance")
			So(paginationQuery.CurrentPage, ShouldEqual, cfg.DefaultPage)
			So(paginationQuery.Limit, ShouldEqual, 10)
			So(err, ShouldBeNil)
		})
		Convey("successfully set default limit to query if limit not given", func() {
			req := httptest.NewRequest("GET", "/search?q=housing&page=1&sort=relevance", nil)
			updatedURL, paginationQuery, err := ReviewQuery(ctx, cfg, req.URL)
			So(updatedURL.Query().Get("limit"), ShouldEqual, strconv.Itoa(cfg.DefaultLimit))
			So(updatedURL.Query().Get("page"), ShouldEqual, "1")
			So(updatedURL.Query().Get("sort"), ShouldEqual, "relevance")
			So(paginationQuery.CurrentPage, ShouldEqual, 1)
			So(paginationQuery.Limit, ShouldEqual, cfg.DefaultLimit)
			So(err, ShouldBeNil)
		})
		Convey("successfully set default limit to query if invalid limit given", func() {
			req := httptest.NewRequest("GET", "/search?q=housing&page=1&sort=relevance&limit=invalid", nil)
			updatedURL, paginationQuery, err := ReviewQuery(ctx, cfg, req.URL)
			So(updatedURL.Query().Get("limit"), ShouldEqual, strconv.Itoa(cfg.DefaultLimit))
			So(updatedURL.Query().Get("page"), ShouldEqual, "1")
			So(updatedURL.Query().Get("sort"), ShouldEqual, "relevance")
			So(paginationQuery.CurrentPage, ShouldEqual, 1)
			So(paginationQuery.Limit, ShouldEqual, cfg.DefaultLimit)
			So(err, ShouldBeNil)
		})
		Convey("successfully set default limit to query if limit less than default limit", func() {
			req := httptest.NewRequest("GET", "/search?q=housing&page=1&sort=relevance&limit=-1", nil)
			updatedURL, paginationQuery, err := ReviewQuery(ctx, cfg, req.URL)
			So(updatedURL.Query().Get("limit"), ShouldEqual, strconv.Itoa(cfg.DefaultLimit))
			So(updatedURL.Query().Get("page"), ShouldEqual, "1")
			So(updatedURL.Query().Get("sort"), ShouldEqual, "relevance")
			So(paginationQuery.CurrentPage, ShouldEqual, 1)
			So(paginationQuery.Limit, ShouldEqual, cfg.DefaultLimit)
			So(err, ShouldBeNil)
		})
		Convey("successfully set default limit to query if limit is between default limit and maximum limit", func() {
			req := httptest.NewRequest("GET", "/search?q=housing&page=1&sort=relevance&limit=20", nil)
			updatedURL, paginationQuery, err := ReviewQuery(ctx, cfg, req.URL)
			So(updatedURL.Query().Get("limit"), ShouldEqual, "20")
			So(updatedURL.Query().Get("page"), ShouldEqual, "1")
			So(updatedURL.Query().Get("sort"), ShouldEqual, "relevance")
			So(paginationQuery.CurrentPage, ShouldEqual, 1)
			So(paginationQuery.Limit, ShouldEqual, 20)
			So(err, ShouldBeNil)
		})
		Convey("successfully set default limit to query if limit more than maximum limit", func() {
			req := httptest.NewRequest("GET", "/search?q=housing&page=1&sort=relevance&limit=100000000000", nil)
			updatedURL, paginationQuery, err := ReviewQuery(ctx, cfg, req.URL)
			So(updatedURL.Query().Get("limit"), ShouldEqual, strconv.Itoa(cfg.DefaultMaximumLimit))
			So(updatedURL.Query().Get("page"), ShouldEqual, "1")
			So(updatedURL.Query().Get("sort"), ShouldEqual, "relevance")
			So(paginationQuery.CurrentPage, ShouldEqual, 1)
			So(paginationQuery.Limit, ShouldEqual, cfg.DefaultMaximumLimit)
			So(err, ShouldBeNil)
		})
		Convey("successfully set default sort to query if sort not given", func() {
			req := httptest.NewRequest("GET", "/search?q=housing&limit=10&page=1", nil)
			updatedURL, paginationQuery, err := ReviewQuery(ctx, cfg, req.URL)
			So(updatedURL.Query().Get("limit"), ShouldEqual, "10")
			So(updatedURL.Query().Get("page"), ShouldEqual, "1")
			So(updatedURL.Query().Get("sort"), ShouldEqual, cfg.DefaultSort)
			So(paginationQuery.CurrentPage, ShouldEqual, 1)
			So(paginationQuery.Limit, ShouldEqual, 10)
			So(err, ShouldBeNil)
		})
		Convey("successfully set default sort to query if invalid sort given", func() {
			req := httptest.NewRequest("GET", "/search?q=housing&limit=10&page=1&sort=invalid", nil)
			updatedURL, paginationQuery, err := ReviewQuery(ctx, cfg, req.URL)
			So(updatedURL.Query().Get("limit"), ShouldEqual, "10")
			So(updatedURL.Query().Get("page"), ShouldEqual, "1")
			So(updatedURL.Query().Get("sort"), ShouldEqual, cfg.DefaultSort)
			So(paginationQuery.CurrentPage, ShouldEqual, 1)
			So(paginationQuery.Limit, ShouldEqual, 10)
			So(err, ShouldBeNil)
		})
	})
}
