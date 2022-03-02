package data

import (
	"context"
	"net/url"
	"testing"

	errs "github.com/ONSdigital/dp-frontend-search-controller/apperrors"
	. "github.com/smartystreets/goconvey/convey"
)

func TestUnitReviewQueryStringSuccess(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	Convey("Given valid query string", t, func() {

		urlQuery := url.Values{
			"q": []string{"housing"},
		}

		Convey("When reviewQueryString is called", func() {
			err := reviewQueryString(ctx, urlQuery)

			Convey("Then return false with a valid query string", func() {
				So(err, ShouldBeNil)
			})

		})
	})
}

func TestUnitReviewQueryStringFailure(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	Convey("Given an invalid query string (only whitespace)", t, func() {

		urlQuery := url.Values{
			"q": []string{"        "},
		}

		Convey("When reviewQueryString is called", func() {
			err := reviewQueryString(ctx, urlQuery)

			Convey("Then return true with a valid query string", func() {
				So(err, ShouldResemble, errs.ErrInvalidQueryString)
			})

		})
	})

	Convey("Given an invalid query string (too short)", t, func() {

		urlQuery := url.Values{
			"q": []string{"ab"},
		}

		Convey("When reviewQueryString is called", func() {
			err := reviewQueryString(ctx, urlQuery)

			Convey("Then return true with a valid query string", func() {
				So(err, ShouldResemble, errs.ErrInvalidQueryString)
			})

		})
	})
}
