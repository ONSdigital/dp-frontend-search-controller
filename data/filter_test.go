package data

import (
	"context"
	"errors"
	"net/http/httptest"
	"testing"

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

	Convey("When MapSubFilterTypes is called", t, func() {
		ctx := context.Background()

		Convey("successfully map one filter given", func() {
			req := httptest.NewRequest("GET", "/search?q=housing&filter=article", nil)
			query := req.URL.Query()
			apiQuery, err := MapSubFilterTypes(ctx, query)
			So(apiQuery["content_type"], ShouldResemble, []string{"article,article_download"})
			So(err, ShouldBeNil)
		})

		Convey("successfully map two or more filters given", func() {
			req := httptest.NewRequest("GET", "/search?q=housing&filter=article&filter=compendia", nil)
			query := req.URL.Query()
			apiQuery, err := MapSubFilterTypes(ctx, query)
			So(apiQuery["content_type"], ShouldResemble, []string{"article,article_download,compendium_landing_page"})
			So(err, ShouldBeNil)
		})

		Convey("successfully map no filters given", func() {
			req := httptest.NewRequest("GET", "/search?q=housing", nil)
			query := req.URL.Query()
			apiQuery, err := MapSubFilterTypes(ctx, query)
			So(apiQuery["content_type"], ShouldBeNil)
			So(err, ShouldBeNil)
		})

		Convey("return error when mapping bad filters", func() {
			req := httptest.NewRequest("GET", "/search?q=housing&filter=INVALID", nil)
			query := req.URL.Query()
			apiQuery, err := MapSubFilterTypes(ctx, query)
			So(apiQuery["content_type"], ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err, ShouldResemble, errors.New("invalid filter type given"))
		})
	})
}
