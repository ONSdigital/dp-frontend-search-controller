package data

import (
	"context"
	"net/http/httptest"
	"testing"

	errs "github.com/ONSdigital/dp-frontend-search-controller/apperrors"
	"github.com/ONSdigital/dp-frontend-search-controller/config"
	. "github.com/smartystreets/goconvey/convey"
)

func createMockCategories() []Category {
	return []Category{Publication, Data, Other}
}

func TestUnitFilter(t *testing.T) {
	t.Parallel()

	Convey("When setCountZero is called", t, func() {
		mockCategories := createMockCategories()
		updatedCategories := setCountZero(mockCategories)
		for i, category := range updatedCategories {
			So(updatedCategories[i].Count, ShouldEqual, 0)
			for j := range category.ContentTypes {
				So(updatedCategories[i].ContentTypes[j].Count, ShouldEqual, 0)
			}
		}

		Convey("When GetAllCategories is called", func() {
			allCategories := GetAllCategories()
			So(allCategories, ShouldResemble, updatedCategories)
		})
	})

	Convey("When GetSearchAPIQuery is called", t, func() {
		ctx := context.Background()
		cfg, err := config.Get()
		So(err, ShouldBeNil)
		pagination := &PaginationQuery{
			Limit:       10,
			CurrentPage: 1,
		}

		Convey("successfully map one filter given", func() {
			req := httptest.NewRequest("GET", "/search?q=housing&filter=article", nil)
			query := req.URL.Query()
			apiQuery, err := GetSearchAPIQuery(ctx, cfg, pagination, query)
			So(apiQuery["content_type"], ShouldResemble, []string{"article,article_download"})
			So(err, ShouldBeNil)
		})

		Convey("successfully map two or more filters given", func() {
			req := httptest.NewRequest("GET", "/search?q=housing&filter=article&filter=compendia", nil)
			query := req.URL.Query()
			apiQuery, err := GetSearchAPIQuery(ctx, cfg, pagination, query)
			So(apiQuery["content_type"], ShouldResemble, []string{"article,article_download,compendium_landing_page"})
			So(err, ShouldBeNil)
		})

		Convey("successfully map no filters given", func() {
			req := httptest.NewRequest("GET", "/search?q=housing", nil)
			query := req.URL.Query()
			apiQuery, err := GetSearchAPIQuery(ctx, cfg, pagination, query)
			So(apiQuery["content_type"], ShouldBeNil)
			So(err, ShouldBeNil)
		})

		Convey("return error when mapping bad filters", func() {
			req := httptest.NewRequest("GET", "/search?q=housing&filter=INVALID", nil)
			query := req.URL.Query()
			apiQuery, err := GetSearchAPIQuery(ctx, cfg, pagination, query)
			So(apiQuery["content_type"], ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err, ShouldResemble, errs.ErrFilterNotFound)
		})
	})
}
