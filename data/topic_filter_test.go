package data

import (
	"context"
	"net/url"
	"testing"

	searchCli "github.com/ONSdigital/dp-api-clients-go/v2/site-search"
	errs "github.com/ONSdigital/dp-frontend-search-controller/apperrors"
	"github.com/ONSdigital/dp-frontend-search-controller/cache"
	. "github.com/smartystreets/goconvey/convey"
)

func TestGetTopicCategories(t *testing.T) {
	t.Parallel()

	Convey("Given the count response has results with topics", t, func() {
		mockSearchCliResponse := searchCli.Response{
			Topics: []searchCli.FilterCount{
				{
					Type:  "1234",
					Count: 1,
				},
			},
		}

		mockCensusTopic := cache.GetMockCensusTopic()

		Convey("When GetTopicCategories is called", func() {
			topicCategories := GetTopicCategories(mockCensusTopic, mockSearchCliResponse)

			Convey("Then a list of topic categories with count should be returned", func() {
				So(topicCategories, ShouldNotBeEmpty)
				So(topicCategories[0].Count, ShouldEqual, 1)
			})
		})
	})

	Convey("Given the count response has results with no topics", t, func() {
		mockSearchCliResponse := searchCli.Response{
			Topics: []searchCli.FilterCount{},
		}

		mockCensusTopic := cache.GetMockCensusTopic()

		Convey("When GetTopicCategories is called", func() {
			topicCategories := GetTopicCategories(mockCensusTopic, mockSearchCliResponse)

			Convey("Then a list of topic categories with 0 count should be returned", func() {
				So(topicCategories, ShouldNotBeEmpty)
				So(topicCategories[0].Count, ShouldEqual, 0)
			})
		})
	})
}

func TestReviewTopicFilters(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	mockCensusTopic := cache.GetMockCensusTopic()

	Convey("Given no topics is selected", t, func() {
		urlQuery := url.Values{}
		validatedQueryParams := &SearchURLParams{}

		Convey("When reviewTopicFilters is called", func() {
			err := reviewTopicFilters(ctx, urlQuery, validatedQueryParams, mockCensusTopic)

			Convey("Then return no errors", func() {
				So(err, ShouldBeNil)
			})

			Convey("And update validatedQueryParams for topics", func() {
				So(validatedQueryParams.TopicFilter, ShouldBeEmpty)
			})
		})
	})

	Convey("Given empty topic is provided", t, func() {
		urlQuery := url.Values{
			"topics": []string{""},
		}
		validatedQueryParams := &SearchURLParams{}

		Convey("When reviewTopicFilters is called", func() {
			err := reviewTopicFilters(ctx, urlQuery, validatedQueryParams, mockCensusTopic)

			Convey("Then return no errors", func() {
				So(err, ShouldBeNil)
			})

			Convey("And update validatedQueryParams for topics", func() {
				So(validatedQueryParams.TopicFilter, ShouldBeEmpty)
			})
		})
	})

	Convey("Given multiple empty topics is provided", t, func() {
		urlQuery := url.Values{
			"topics": []string{"", "", ""},
		}
		validatedQueryParams := &SearchURLParams{}

		Convey("When reviewTopicFilters is called", func() {
			err := reviewTopicFilters(ctx, urlQuery, validatedQueryParams, mockCensusTopic)

			Convey("Then return no errors", func() {
				So(err, ShouldBeNil)
			})

			Convey("And update validatedQueryParams for topics", func() {
				So(validatedQueryParams.TopicFilter, ShouldBeEmpty)
			})
		})
	})

	Convey("Given one topic is selected", t, func() {
		urlQuery := url.Values{
			"topics": []string{"1234"},
		}

		validatedQueryParams := &SearchURLParams{}

		Convey("When reviewTopicFilters is called", func() {
			err := reviewTopicFilters(ctx, urlQuery, validatedQueryParams, mockCensusTopic)

			Convey("Then return no error", func() {
				So(err, ShouldBeNil)
			})

			Convey("And update validatedQueryParams for topics", func() {
				So(validatedQueryParams.TopicFilter, ShouldEqual, "1234")
			})
		})
	})

	Convey("Given more than one valid topics is selected", t, func() {
		urlQuery := url.Values{
			"topics": []string{"1234,5678"},
		}

		validatedQueryParams := &SearchURLParams{}

		Convey("When reviewTopicFilters is called", func() {
			err := reviewTopicFilters(ctx, urlQuery, validatedQueryParams, mockCensusTopic)

			Convey("Then return no error", func() {
				So(err, ShouldBeNil)
			})

			Convey("And update validatedQueryParams for topics", func() {
				So(validatedQueryParams.TopicFilter, ShouldEqual, "1234,5678")
			})
		})
	})

	Convey("Given more than one valid topics is selected and given separately", t, func() {
		urlQuery := url.Values{
			"topics": []string{"1234", "5678"},
		}

		validatedQueryParams := &SearchURLParams{}

		Convey("When reviewTopicFilters is called", func() {
			err := reviewTopicFilters(ctx, urlQuery, validatedQueryParams, mockCensusTopic)

			Convey("Then return no error", func() {
				So(err, ShouldBeNil)
			})

			Convey("And update validatedQueryParams for the first given topic", func() {
				So(validatedQueryParams.TopicFilter, ShouldEqual, "1234")
			})
		})
	})

	Convey("Given a mix of empty and valid topics", t, func() {
		urlQuery := url.Values{
			"topics": []string{"1234", ""},
		}

		validatedQueryParams := &SearchURLParams{}

		Convey("When reviewTopicFilters is called", func() {
			err := reviewTopicFilters(ctx, urlQuery, validatedQueryParams, mockCensusTopic)

			Convey("Then return no error", func() {
				So(err, ShouldBeNil)
			})

			Convey("And update validatedQueryParams for topics", func() {
				So(validatedQueryParams.TopicFilter, ShouldEqual, "1234")
			})
		})
	})

	Convey("Given an invalid topic", t, func() {
		urlQuery := url.Values{
			"topics": []string{"invalid"},
		}

		validatedQueryParams := &SearchURLParams{}

		Convey("When reviewTopicFilters is called", func() {
			err := reviewTopicFilters(ctx, urlQuery, validatedQueryParams, mockCensusTopic)

			Convey("Then return no error", func() {
				So(err, ShouldNotBeNil)
				So(err, ShouldResemble, errs.ErrFilterNotFound)
			})
		})
	})

	Convey("Given a mix of valid and invalid topics", t, func() {
		urlQuery := url.Values{
			"topics": []string{"1234,invalid"},
		}

		validatedQueryParams := &SearchURLParams{}

		Convey("When reviewTopicFilters is called", func() {
			err := reviewTopicFilters(ctx, urlQuery, validatedQueryParams, mockCensusTopic)

			Convey("Then return an error", func() {
				So(err, ShouldNotBeNil)
				So(err, ShouldResemble, errs.ErrFilterNotFound)
			})
		})
	})

	Convey("Given a mix of empty, valid and invalid topics", t, func() {
		urlQuery := url.Values{
			"topics": []string{"1234,invalid,"},
		}

		validatedQueryParams := &SearchURLParams{}

		Convey("When reviewTopicFilters is called", func() {
			err := reviewTopicFilters(ctx, urlQuery, validatedQueryParams, mockCensusTopic)

			Convey("Then return an error", func() {
				So(err, ShouldNotBeNil)
				So(err, ShouldResemble, errs.ErrFilterNotFound)
			})
		})
	})
}
