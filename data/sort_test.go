package data

import (
	"context"
	"net/url"
	"testing"

	"github.com/ONSdigital/dp-frontend-search-controller/config"
	. "github.com/smartystreets/goconvey/convey"
)

func TestUnitReviewSortSuccess(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	Convey("Given an empty sort", t, func() {
		cfg, err := config.Get()
		So(err, ShouldBeNil)

		urlQuery := url.Values{
			"sort": []string{""},
		}

		validatedQueryParams := &SearchURLParams{}

		Convey("When reviewSort is called", func() {
			reviewSort(ctx, urlQuery, validatedQueryParams, cfg.DefaultSort.Default)

			Convey("Then set default sort and localisation key for default to validatedQueryParams", func() {
				So(validatedQueryParams.Sort.Query, ShouldEqual, cfg.DefaultSort.Default)
				So(validatedQueryParams.Sort.LocaliseKeyName, ShouldEqual, sortOptions[cfg.DefaultSort.Default].LocaliseKeyName)
			})
		})
	})

	Convey("Given a valid sort which is available in the sort options", t, func() {
		cfg, err := config.Get()
		So(err, ShouldBeNil)

		urlQuery := url.Values{
			"sort": []string{"relevance"},
		}

		validatedQueryParams := &SearchURLParams{}

		Convey("When reviewSort is called", func() {
			reviewSort(ctx, urlQuery, validatedQueryParams, cfg.DefaultSort.Default)

			Convey("Then set sort query and localisation key for sort value to validatedQueryParams", func() {
				So(validatedQueryParams.Sort.Query, ShouldEqual, "relevance")
				So(validatedQueryParams.Sort.LocaliseKeyName, ShouldEqual, "Relevance")
			})
		})
	})

	Convey("Given an invalid sort which is not available in the sort options", t, func() {
		cfg, err := config.Get()
		So(err, ShouldBeNil)

		urlQuery := url.Values{
			"sort": []string{"invalid"},
		}

		validatedQueryParams := &SearchURLParams{}

		Convey("When reviewSort is called", func() {
			reviewSort(ctx, urlQuery, validatedQueryParams, cfg.DefaultSort.Default)

			Convey("Then set default sort and localisation key for default to validatedQueryParams", func() {
				So(validatedQueryParams.Sort.Query, ShouldEqual, cfg.DefaultSort.Default)
				So(validatedQueryParams.Sort.LocaliseKeyName, ShouldEqual, sortOptions[cfg.DefaultSort.Default].LocaliseKeyName)
			})
		})
	})
}
