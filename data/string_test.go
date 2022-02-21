package data

import (
	"context"
	"net/url"
	"testing"

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
			validationProblem := reviewQueryString(ctx, urlQuery)

			Convey("Then return false with a valid query string", func() {
				So(validationProblem, ShouldBeFalse)
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
			validationProblem := reviewQueryString(ctx, urlQuery)

			Convey("Then return true with a valid query string", func() {
				So(validationProblem, ShouldBeTrue)
			})

		})
	})

	Convey("Given an invalid query string (too short)", t, func() {

		urlQuery := url.Values{
			"q": []string{"ab"},
		}

		Convey("When reviewQueryString is called", func() {
			validationProblem := reviewQueryString(ctx, urlQuery)

			Convey("Then return true with a valid query string", func() {
				So(validationProblem, ShouldBeTrue)
			})

		})
	})
}
