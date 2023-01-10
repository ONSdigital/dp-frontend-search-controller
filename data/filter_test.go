package data

import (
	"context"
	"net/url"
	"testing"

	errs "github.com/ONSdigital/dp-frontend-search-controller/apperrors"
	. "github.com/smartystreets/goconvey/convey"
)

func TestUnitReviewFiltersSuccess(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	Convey("Given no filter is selected", t, func() {
		urlQuery := url.Values{}
		validatedQueryParams := &SearchURLParams{}

		Convey("When reviewFilter is called", func() {
			err := reviewFilters(ctx, urlQuery, validatedQueryParams)

			Convey("Then return no errors", func() {
				So(err, ShouldBeNil)
			})

			Convey("And update validatedQueryParams for filter", func() {
				So(validatedQueryParams.Filter.Query, ShouldBeEmpty)
				So(validatedQueryParams.Filter.LocaliseKeyName, ShouldBeEmpty)
			})
		})
	})

	Convey("Given empty filter is provided", t, func() {
		urlQuery := url.Values{
			"filter": []string{""},
		}
		validatedQueryParams := &SearchURLParams{}

		Convey("When reviewFilter is called", func() {
			err := reviewFilters(ctx, urlQuery, validatedQueryParams)

			Convey("Then return no errors", func() {
				So(err, ShouldBeNil)
			})

			Convey("And update validatedQueryParams for filter", func() {
				So(validatedQueryParams.Filter.Query, ShouldBeEmpty)
				So(validatedQueryParams.Filter.LocaliseKeyName, ShouldBeEmpty)
			})
		})
	})

	Convey("Given multiple empty filter is provided", t, func() {
		urlQuery := url.Values{
			"filter": []string{"", "", ""},
		}
		validatedQueryParams := &SearchURLParams{}

		Convey("When reviewFilter is called", func() {
			err := reviewFilters(ctx, urlQuery, validatedQueryParams)

			Convey("Then return no errors", func() {
				So(err, ShouldBeNil)
			})

			Convey("And update validatedQueryParams for filter", func() {
				So(validatedQueryParams.Filter.Query, ShouldBeEmpty)
				So(validatedQueryParams.Filter.LocaliseKeyName, ShouldBeEmpty)
			})
		})
	})

	Convey("Given one valid filter is selected", t, func() {
		urlQuery := url.Values{
			"filter": []string{"article"},
		}

		validatedQueryParams := &SearchURLParams{}

		Convey("When reviewFilter is called", func() {
			err := reviewFilters(ctx, urlQuery, validatedQueryParams)

			Convey("Then return no error", func() {
				So(err, ShouldBeNil)
			})

			Convey("And update validatedQueryParams for filter", func() {
				So(validatedQueryParams, ShouldNotBeNil)
				So(validatedQueryParams.Filter.Query, ShouldResemble, []string{"article"})
				So(validatedQueryParams.Filter.LocaliseKeyName, ShouldResemble, []string{"Article"})
			})
		})
	})

	Convey("Given more than one valid filter is selected", t, func() {
		urlQuery := url.Values{
			"filter": []string{"article", "bulletin"},
		}

		validatedQueryParams := &SearchURLParams{}

		Convey("When reviewFilter is called", func() {
			err := reviewFilters(ctx, urlQuery, validatedQueryParams)

			Convey("Then return no error", func() {
				So(err, ShouldBeNil)
			})

			Convey("And update validatedQueryParams for filter", func() {
				So(validatedQueryParams, ShouldNotBeNil)
				So(validatedQueryParams.Filter.Query, ShouldResemble, []string{"article", "bulletin"})
				So(validatedQueryParams.Filter.LocaliseKeyName, ShouldResemble, []string{"Article", "StatisticalBulletin"})
			})
		})
	})

	Convey("Given filter with mixed case", t, func() {
		urlQuery := url.Values{
			"filter": []string{"ArTiClE"},
		}

		validatedQueryParams := &SearchURLParams{}

		Convey("When reviewFilter is called", func() {
			err := reviewFilters(ctx, urlQuery, validatedQueryParams)

			Convey("Then return no error", func() {
				So(err, ShouldBeNil)
			})

			Convey("And update validatedQueryParams for filter", func() {
				So(validatedQueryParams, ShouldNotBeNil)
				So(validatedQueryParams.Filter.Query, ShouldResemble, []string{"article"})
				So(validatedQueryParams.Filter.LocaliseKeyName, ShouldResemble, []string{"Article"})
			})
		})
	})

	Convey("Given a mix of empty and valid filters", t, func() {
		urlQuery := url.Values{
			"filter": []string{"", "article", ""},
		}

		validatedQueryParams := &SearchURLParams{}

		Convey("When reviewFilter is called", func() {
			err := reviewFilters(ctx, urlQuery, validatedQueryParams)

			Convey("Then return no error", func() {
				So(err, ShouldBeNil)
			})

			Convey("And update validatedQueryParams for filter", func() {
				So(validatedQueryParams, ShouldNotBeNil)
				So(validatedQueryParams.Filter.Query, ShouldResemble, []string{"article"})
				So(validatedQueryParams.Filter.LocaliseKeyName, ShouldResemble, []string{"Article"})
			})
		})
	})
}

func TestUnitReviewFiltersFailure(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	Convey("Given an invalid filter is provided", t, func() {
		urlQuery := url.Values{
			"filter": []string{"INVALID"},
		}

		validatedQueryParams := &SearchURLParams{}

		Convey("When reviewFilter is called", func() {
			err := reviewFilters(ctx, urlQuery, validatedQueryParams)

			Convey("Then return an error", func() {
				So(err, ShouldNotBeNil)
				So(err, ShouldResemble, errs.ErrContentTypeNotFound)
			})
		})
	})

	Convey("Given a mix of valid and invalid filters", t, func() {
		urlQuery := url.Values{
			"filter": []string{"BORK", "article", "bark bark bark"},
		}

		validatedQueryParams := &SearchURLParams{}

		Convey("When reviewFilter is called", func() {
			err := reviewFilters(ctx, urlQuery, validatedQueryParams)

			Convey("Then return an error", func() {
				So(err, ShouldNotBeNil)
				So(err, ShouldResemble, errs.ErrContentTypeNotFound)
			})
		})
	})

	Convey("Given a mix of empty, valid and invalid filters", t, func() {
		urlQuery := url.Values{
			"filter": []string{"BORK", "article", ""},
		}

		validatedQueryParams := &SearchURLParams{}

		Convey("When reviewFilter is called", func() {
			err := reviewFilters(ctx, urlQuery, validatedQueryParams)

			Convey("Then return an error", func() {
				So(err, ShouldNotBeNil)
				So(err, ShouldResemble, errs.ErrContentTypeNotFound)
			})
		})
	})
}

func TestUnitGetCategoriesSuccess(t *testing.T) {
	t.Parallel()

	Convey("When GetCategories is called", t, func() {
		categories := GetCategories()

		Convey("Then return list of categories which includes its filter types", func() {
			So(categories, ShouldNotBeNil)
			So(categories, ShouldResemble, Categories)
		})

		Convey("And all count should be set to zero", func() {
			for i := range categories {
				So(categories[i].Count, ShouldEqual, 0)

				for j := range categories[i].ContentTypes {
					So(categories[i].ContentTypes[j].Count, ShouldEqual, 0)
				}
			}
		})
	})
}

func TestUnitUpdateQueryWithAPIFiltersSuccess(t *testing.T) {
	t.Parallel()

	Convey("Given no filter is selected", t, func() {
		apiQuery := url.Values{}
		expected := url.Values{"content_type": []string{defaultContentTypes}}

		Convey("When updateQueryWithAPIFilters is called", func() {
			updateQueryWithAPIFilters(apiQuery)

			Convey("Use default content type list", func() {
				So(apiQuery, ShouldResemble, expected)
			})
		})
	})

	Convey("Given one or more filters are selected", t, func() {
		apiQuery := url.Values{
			"content_type": []string{"article", "bulletin"},
		}

		Convey("When updateQueryWithAPIFilters is called", func() {
			updateQueryWithAPIFilters(apiQuery)

			Convey("Then set apiQuery's content_type with the respective sub-filters", func() {
				So(apiQuery, ShouldNotBeEmpty)
				So(apiQuery.Get("content_type"), ShouldEqual, "article,article_download,bulletin")
			})
		})
	})
}

func TestUnitGetSubFiltersSuccess(t *testing.T) {
	t.Parallel()

	Convey("Given one or more filters are provided", t, func() {
		filters := []string{"article", "bulletin"}

		Convey("When getSubFilters is called", func() {
			subFilters := getSubFilters(filters)

			Convey("Then get the respective sub filters for the filters given", func() {
				So(subFilters, ShouldResemble, []string{"article", "article_download", "bulletin"})
			})
		})
	})
}

func TestUnitGetGroupLocaliseKey(t *testing.T) {
	t.Parallel()

	Convey("Given the type of the search result", t, func() {
		searchType := "static_methodology"

		Convey("When getSubFilters is called", func() {
			groupLocaliseKey := GetGroupLocaliseKey(searchType)

			Convey("Then the localise key of the group type should be returned", func() {
				So(groupLocaliseKey, ShouldEqual, "Methodology")
			})
		})
	})

	Convey("Given an unknown type of the search result", t, func() {
		searchType := "unknown"

		Convey("When getSubFilters is called", func() {
			groupLocaliseKey := GetGroupLocaliseKey(searchType)

			Convey("Then an empty localise key should be returned", func() {
				So(groupLocaliseKey, ShouldEqual, "")
			})
		})
	})
}
