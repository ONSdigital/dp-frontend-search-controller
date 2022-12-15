package data

import (
	"context"
	"net/url"
	"testing"

	"github.com/ONSdigital/dp-frontend-search-controller/cache"
	"github.com/ONSdigital/dp-frontend-search-controller/config"
	. "github.com/smartystreets/goconvey/convey"
)

func TestUnitReviewQuerySuccess(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	Convey("Given valid url query", t, func() {
		cfg, err := config.Get()
		So(err, ShouldBeNil)

		urlQuery := url.Values{
			"q":      []string{"housing"},
			"filter": []string{"article"},
			"topics": []string{"1234,5678"},
			"sort":   []string{"relevance"},
			"limit":  []string{"10"},
			"page":   []string{"1"},
		}

		Convey("When ReviewQuery is called", func() {
			validatedQueryParams, err := ReviewQuery(ctx, cfg, urlQuery, cache.GetMockCensusTopic())

			Convey("Then successfully review and return validated query parameters", func() {
				So(validatedQueryParams, ShouldResemble, SearchURLParams{
					Query: "housing",
					Filter: Filter{
						Query:           []string{"article"},
						LocaliseKeyName: []string{"Article"},
					},
					Sort: Sort{
						Query:           "relevance",
						LocaliseKeyName: "Relevance",
					},
					Limit:       10,
					CurrentPage: 1,
					TopicFilter: "1234,5678",
				})
			})

			Convey("And return no errors", func() {
				So(err, ShouldBeNil)
			})

			Convey("And have a valid query string", func() {
				So(err, ShouldBeNil)
			})
		})
	})

	Convey("Given process query when both valid content type and invalid topic filters provided", t, func() {
		cfg, err := config.Get()
		So(err, ShouldBeNil)

		urlQuery := url.Values{
			"q":      []string{"housing"},
			"filter": []string{"article"},
			"topics": []string{"INVALID"},
			"sort":   []string{"relevance"},
			"limit":  []string{"10"},
			"page":   []string{"1"},
		}

		Convey("When ReviewQuery is called", func() {
			_, err := ReviewQuery(ctx, cfg, urlQuery, cache.GetMockCensusTopic())

			Convey("Then return an error", func() {
				So(err, ShouldBeNil)
			})
		})
	})

	Convey("Given process query when both invalid content type and valid topic filters provided", t, func() {
		cfg, err := config.Get()
		So(err, ShouldBeNil)

		urlQuery := url.Values{
			"q":      []string{"housing"},
			"filter": []string{"INVALID"},
			"topics": []string{"1234,5678"},
			"sort":   []string{"relevance"},
			"limit":  []string{"10"},
			"page":   []string{"1"},
		}

		Convey("When ReviewQuery is called", func() {
			_, err := ReviewQuery(ctx, cfg, urlQuery, cache.GetMockCensusTopic())

			Convey("Then return an error", func() {
				So(err, ShouldBeNil)
			})
		})
	})

	Convey("Given process query when both are valid filters but the query is less than minimum char length", t, func() {
		cfg, err := config.Get()
		So(err, ShouldBeNil)

		urlQuery := url.Values{
			"q":      []string{"h"},
			"filter": []string{"article"},
			"topics": []string{"1234,5678"},
			"sort":   []string{"relevance"},
			"limit":  []string{"10"},
			"page":   []string{"1"},
		}

		Convey("When ReviewQuery is called", func() {
			_, err := ReviewQuery(ctx, cfg, urlQuery, cache.GetMockCensusTopic())

			Convey("Then return an error", func() {
				So(err, ShouldBeNil)
			})
		})
	})
}

func TestUnitReviewQueryFailure(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	Convey("Given failure to review pagination due to invalid pagination parameters provided", t, func() {
		cfg, err := config.Get()
		So(err, ShouldBeNil)

		urlQuery := url.Values{
			"q":      []string{"housing"},
			"filter": []string{"article"},
			"topics": []string{"1234,5678"},
			"sort":   []string{"relevance"},
			"limit":  []string{"10"},
			"page":   []string{"10000000"},
		}

		Convey("When ReviewQuery is called", func() {
			_, err := ReviewQuery(ctx, cfg, urlQuery, cache.GetMockCensusTopic())

			Convey("Then return an error", func() {
				So(err, ShouldNotBeNil)
			})
		})
	})

	Convey("Given failure to review filter due to both invalid content type and topic filters provided", t, func() {
		cfg, err := config.Get()
		So(err, ShouldBeNil)

		urlQuery := url.Values{
			"q":      []string{"housing"},
			"filter": []string{"INVALID"},
			"topics": []string{"INVALID"},
			"sort":   []string{"relevance"},
			"limit":  []string{"10"},
			"page":   []string{"1"},
		}

		Convey("When ReviewQuery is called", func() {
			_, err := ReviewQuery(ctx, cfg, urlQuery, cache.GetMockCensusTopic())

			Convey("Then return an error", func() {
				So(err, ShouldNotBeNil)
			})
		})
	})
}

func TestUnitGetSearchAPIQuerySuccess(t *testing.T) {
	t.Parallel()

	mockCensusTopic := cache.GetMockCensusTopic()

	Convey("Given validated query parameters", t, func() {
		validatedQueryParams := SearchURLParams{
			Query: "housing",
			Filter: Filter{
				Query:           []string{"article"},
				LocaliseKeyName: []string{"Article"},
			},
			Sort: Sort{
				Query:           "relevance",
				LocaliseKeyName: "Relevance",
			},
			Limit:       10,
			Offset:      0,
			TopicFilter: "1234,5678",
		}

		Convey("When GetSearchAPIQuery is called", func() {
			apiQuery := GetSearchAPIQuery(validatedQueryParams, mockCensusTopic)

			Convey("Then successfully return apiQuery for dp-search-api", func() {
				So(apiQuery["q"], ShouldResemble, []string{"housing"})
				So(apiQuery["sort"], ShouldResemble, []string{"relevance"})
				So(apiQuery["limit"], ShouldResemble, []string{"10"})
				So(apiQuery["offset"], ShouldResemble, []string{"0"})
				So(apiQuery["topics"], ShouldResemble, []string{"1234,5678"})
			})

			Convey("And update content_type (filter) query with sub filters", func() {
				So(apiQuery["content_type"], ShouldResemble, []string{"article,article_download"})
			})
		})
	})
}

func TestUnitCreateSearchAPIQuerySuccess(t *testing.T) {
	t.Parallel()

	Convey("Given validated query parameters provided", t, func() {
		validatedQueryParams := SearchURLParams{
			Query: "housing",
			Filter: Filter{
				Query:           []string{"article"},
				LocaliseKeyName: []string{"Article"},
			},
			Sort: Sort{
				Query:           "relevance",
				LocaliseKeyName: "Relevance",
			},
			Limit:       10,
			Offset:      0,
			TopicFilter: "1234,5678",
		}

		Convey("When createSearchAPIQuery is called", func() {
			apiQuery := createSearchAPIQuery(validatedQueryParams)

			Convey("Then successfully return api query for dp-search-api", func() {
				So(apiQuery, ShouldResemble, url.Values{
					"q":            []string{"housing"},
					"content_type": []string{"article"},
					"sort":         []string{"relevance"},
					"limit":        []string{"10"},
					"offset":       []string{"0"},
					"topics":       []string{"1234,5678"},
				})
			})
		})
	})
}

func TestUnitCreateSearchControllerQuerySuccess(t *testing.T) {
	t.Parallel()

	Convey("Given validated query parameters provided", t, func() {
		validatedQueryParams := SearchURLParams{
			Query: "housing",
			Filter: Filter{
				Query:           []string{"article"},
				LocaliseKeyName: []string{"Article"},
			},
			Sort: Sort{
				Query:           "relevance",
				LocaliseKeyName: "Relevance",
			},
			Limit:       10,
			CurrentPage: 1,
		}

		Convey("When createSearchControllerQuery is called", func() {
			controllerQuery := createSearchControllerQuery(validatedQueryParams)

			Convey("Then successfully return controller query", func() {
				So(controllerQuery, ShouldResemble, url.Values{
					"q":      []string{"housing"},
					"filter": []string{"article"},
					"sort":   []string{"relevance"},
					"limit":  []string{"10"},
					"page":   []string{"1"},
				})
			})
		})
	})
}
