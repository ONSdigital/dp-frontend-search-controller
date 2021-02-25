package data

import (
	"context"
	"net/http/httptest"

	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestUnitQuery(t *testing.T) {
	t.Parallel()

	Convey("When GetLimitOptions is called", t, func() {
		Convey("successfully get limit options", func() {
			limitOptions := GetLimitOptions()
			So(limitOptions, ShouldResemble, []int{10, 25, 50})
		})
	})

	Convey("When updateQueryWithOffset called", t, func() {
		ctx := context.Background()
		Convey("successfully update query with offset", func() {
			req := httptest.NewRequest("GET", "/search?q=housing&limit=10&page=1", nil)
			query := req.URL.Query()
			updatedQuery := updateQueryWithOffset(ctx, query).Encode()
			So(updatedQuery, ShouldContainSubstring, "offset=0")
			So(updatedQuery, ShouldNotContainSubstring, "page=")
		})
		Convey("successfully update query with offset with invalid limit", func() {
			req := httptest.NewRequest("GET", "/search?q=housing&limit=invalid&page=2", nil)
			query := req.URL.Query()
			updatedQuery := updateQueryWithOffset(ctx, query).Encode()
			So(updatedQuery, ShouldContainSubstring, "offset=10")
			So(updatedQuery, ShouldNotContainSubstring, "page=")
		})
		Convey("successfully update query with offset with invalid page", func() {
			req := httptest.NewRequest("GET", "/search?q=housing&limit=10&page=invalid", nil)
			query := req.URL.Query()
			updatedQuery := updateQueryWithOffset(ctx, query).Encode()
			So(updatedQuery, ShouldContainSubstring, "offset=0")
			So(updatedQuery, ShouldNotContainSubstring, "page=")
		})
	})

	Convey("When SetDefaultQueries called", t, func() {
		ctx := context.Background()
		Convey("successfully set default page to query if page not given", func() {
			req := httptest.NewRequest("GET", "/search?q=housing&limit=10&sort=relevance", nil)
			updatedURL, paginationQuery := SetDefaultQueries(ctx, req.URL)
			So(updatedURL.Query().Get("limit"), ShouldEqual, "10")
			So(updatedURL.Query().Get("page"), ShouldEqual, DefaultPageStr)
			So(updatedURL.Query().Get("sort"), ShouldEqual, "relevance")
			So(paginationQuery.CurrentPage, ShouldEqual, DefaultPage)
			So(paginationQuery.Limit, ShouldEqual, 10)
		})
		Convey("successfully set default page to query if invalid page given", func() {
			req := httptest.NewRequest("GET", "/search?q=housing&limit=10&sort=relevance&page=invalid", nil)
			updatedURL, paginationQuery := SetDefaultQueries(ctx, req.URL)
			So(updatedURL.Query().Get("limit"), ShouldEqual, "10")
			So(updatedURL.Query().Get("page"), ShouldEqual, DefaultPageStr)
			So(updatedURL.Query().Get("sort"), ShouldEqual, "relevance")
			So(paginationQuery.CurrentPage, ShouldEqual, DefaultPage)
			So(paginationQuery.Limit, ShouldEqual, 10)
		})
		Convey("successfully set default page to query if page less than 1", func() {
			req := httptest.NewRequest("GET", "/search?q=housing&limit=10&sort=relevance&page=0", nil)
			updatedURL, paginationQuery := SetDefaultQueries(ctx, req.URL)
			So(updatedURL.Query().Get("limit"), ShouldEqual, "10")
			So(updatedURL.Query().Get("page"), ShouldEqual, DefaultPageStr)
			So(updatedURL.Query().Get("sort"), ShouldEqual, "relevance")
			So(paginationQuery.CurrentPage, ShouldEqual, DefaultPage)
			So(paginationQuery.Limit, ShouldEqual, 10)
		})
		Convey("successfully set default limit to query if limit not given", func() {
			req := httptest.NewRequest("GET", "/search?q=housing&page=1&sort=relevance", nil)
			updatedURL, paginationQuery := SetDefaultQueries(ctx, req.URL)
			So(updatedURL.Query().Get("limit"), ShouldEqual, DefaultLimitStr)
			So(updatedURL.Query().Get("page"), ShouldEqual, "1")
			So(updatedURL.Query().Get("sort"), ShouldEqual, "relevance")
			So(paginationQuery.CurrentPage, ShouldEqual, 1)
			So(paginationQuery.Limit, ShouldEqual, DefaultLimit)
		})
		Convey("successfully set default limit to query if invalid limit given", func() {
			req := httptest.NewRequest("GET", "/search?q=housing&page=1&sort=relevance&limit=invalid", nil)
			updatedURL, paginationQuery := SetDefaultQueries(ctx, req.URL)
			So(updatedURL.Query().Get("limit"), ShouldEqual, DefaultLimitStr)
			So(updatedURL.Query().Get("page"), ShouldEqual, "1")
			So(updatedURL.Query().Get("sort"), ShouldEqual, "relevance")
			So(paginationQuery.CurrentPage, ShouldEqual, 1)
			So(paginationQuery.Limit, ShouldEqual, DefaultLimit)
		})
		Convey("successfully set default limit to query if limit does not exist in limit options", func() {
			req := httptest.NewRequest("GET", "/search?q=housing&page=1&sort=relevance&limit=2", nil)
			updatedURL, paginationQuery := SetDefaultQueries(ctx, req.URL)
			So(updatedURL.Query().Get("limit"), ShouldEqual, DefaultLimitStr)
			So(updatedURL.Query().Get("page"), ShouldEqual, "1")
			So(updatedURL.Query().Get("sort"), ShouldEqual, "relevance")
			So(paginationQuery.CurrentPage, ShouldEqual, 1)
			So(paginationQuery.Limit, ShouldEqual, DefaultLimit)
		})
		Convey("successfully set default sort to query if sort not given", func() {
			req := httptest.NewRequest("GET", "/search?q=housing&limit=10&page=1", nil)
			updatedURL, paginationQuery := SetDefaultQueries(ctx, req.URL)
			So(updatedURL.Query().Get("limit"), ShouldEqual, "10")
			So(updatedURL.Query().Get("page"), ShouldEqual, "1")
			So(updatedURL.Query().Get("sort"), ShouldEqual, DefaultSort)
			So(paginationQuery.CurrentPage, ShouldEqual, 1)
			So(paginationQuery.Limit, ShouldEqual, 10)
		})
		Convey("successfully set default sort to query if invalid sort given", func() {
			req := httptest.NewRequest("GET", "/search?q=housing&limit=10&page=1&sort=invalid", nil)
			updatedURL, paginationQuery := SetDefaultQueries(ctx, req.URL)
			So(updatedURL.Query().Get("limit"), ShouldEqual, "10")
			So(updatedURL.Query().Get("page"), ShouldEqual, "1")
			So(updatedURL.Query().Get("sort"), ShouldEqual, DefaultSort)
			So(paginationQuery.CurrentPage, ShouldEqual, 1)
			So(paginationQuery.Limit, ShouldEqual, 10)
		})
	})
}
