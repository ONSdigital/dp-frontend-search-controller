package data

import (
	"context"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"

	errs "github.com/ONSdigital/dp-frontend-search-controller/apperrors"
	"github.com/ONSdigital/dp-frontend-search-controller/config"
	. "github.com/smartystreets/goconvey/convey"
)

func createMockCategories() []Category {
	return []Category{Publication, Data, Other}
}

func TestUnitGetAllCategories(t *testing.T) {
	t.Parallel()

	var updatedCategories []Category

	Convey("When setCountZero is called", t, func() {
		mockCategories := createMockCategories()
		updatedCategories = setCountZero(mockCategories)

		for i, category := range updatedCategories {
			So(updatedCategories[i].Count, ShouldEqual, 0)

			for j := range category.ContentTypes {
				So(updatedCategories[i].ContentTypes[j].Count, ShouldEqual, 0)
			}

		}
	})

	Convey("When GetAllCategories is called", t, func() {
		allCategories := GetAllCategories()
		So(allCategories, ShouldResemble, updatedCategories)
	})

}

func TestUnitGetSearchAPIQuery(t *testing.T) {
	t.Parallel()

	Convey("When GetSearchAPIQuery is called", t, func() {
		ctx := context.Background()

		cfg, err := config.Get()
		So(err, ShouldBeNil)

		pagination := &PaginationQuery{
			Limit:       10,
			CurrentPage: 1,
		}

		Convey("successfully get query for search api", func() {

			Convey("when valid filters are given", func() {
				req := httptest.NewRequest("GET", "/search?q=housing&filter=article&filter=bulletin", nil)

				query := req.URL.Query()
				apiQuery, err := GetSearchAPIQuery(ctx, cfg, pagination, query)

				So(err, ShouldBeNil)

				So(apiQuery, ShouldContainKey, "offset")
				So(apiQuery.Get("offset"), ShouldEqual, strconv.Itoa(0))
				So(apiQuery, ShouldNotContainKey, "page")

				So(apiQuery, ShouldContainKey, "content_type")
				So(apiQuery["content_type"], ShouldResemble, []string{"article,article_download,bulletin"})
				So(apiQuery, ShouldNotContainKey, "filter")
			})
		})

		Convey("return error", func() {

			Convey("when failed to update query with offset", func() {
				req := httptest.NewRequest("GET", "/search?q=housing", nil)
				query := req.URL.Query()

				// A large offset value will be calculated which is invalid
				pagination.CurrentPage = 10000

				apiQuery, err := GetSearchAPIQuery(ctx, cfg, pagination, query)

				So(apiQuery, ShouldBeNil)
				So(err, ShouldNotBeNil)
				So(err, ShouldResemble, errs.ErrInvalidPage)
			})

			Convey("when failed to update query with api filters", func() {
				req := httptest.NewRequest("GET", "/search?q=housing&filter=INVALID", nil)
				query := req.URL.Query()
				apiQuery, err := GetSearchAPIQuery(ctx, cfg, pagination, query)

				So(apiQuery, ShouldBeNil)
				So(err, ShouldNotBeNil)
				So(err, ShouldResemble, errs.ErrFilterNotFound)
			})
		})
	})

	Convey("When updateQueryWithAPIFilters is called", t, func() {
		ctx := context.Background()

		mockAPIQuery := url.Values{
			"filter": []string{"bulletin", "article"},
		}

		Convey("successful update query with api filters", func() {

			Convey("when no filters given", func() {
				err := updateQueryWithAPIFilters(ctx, mockAPIQuery)
				So(err, ShouldBeNil)
			})

			Convey("when valid filters given", func() {
				err := updateQueryWithAPIFilters(ctx, mockAPIQuery)
				So(err, ShouldBeNil)
				So(mockAPIQuery, ShouldNotContainKey, "filter")
				So(mockAPIQuery, ShouldContainKey, "content_type")
				So(mockAPIQuery["content_type"], ShouldResemble, []string{"bulletin,article,article_download"})
			})
		})

		Convey("return error", func() {

			Convey("when failed to get sub filters to query", func() {
				mockAPIQuery["filter"] = []string{"invalid"}
				err := updateQueryWithAPIFilters(ctx, mockAPIQuery)
				So(err, ShouldNotBeNil)
			})
		})
	})

	Convey("When getSubFilters is called", t, func() {

		Convey("successful update query with api filters", func() {

			Convey("when no filters given", func() {
				filters := []string{}
				subFilters, err := getSubFilters(filters)

				So(subFilters, ShouldResemble, []string{})
				So(err, ShouldBeNil)
			})

			Convey("when one filter is given", func() {
				filters := []string{"article"}
				subFilters, err := getSubFilters(filters)

				So(subFilters, ShouldResemble, []string{"article", "article_download"})
				So(err, ShouldBeNil)
			})

			Convey("when two or more filters are given", func() {
				filters := []string{"article", "bulletin"}
				subFilters, err := getSubFilters(filters)

				So(subFilters, ShouldResemble, []string{"article", "article_download", "bulletin"})
				So(err, ShouldBeNil)
			})
		})

		Convey("return error", func() {

			Convey("when invalid filter given", func() {
				filters := []string{"invalid"}
				subFilters, err := getSubFilters(filters)

				So(subFilters, ShouldBeNil)
				So(err, ShouldNotBeNil)
				So(err, ShouldResemble, errs.ErrFilterNotFound)
			})
		})
	})
}
