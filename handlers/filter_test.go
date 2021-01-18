package handlers

import (
	"context"
	"errors"
	"net/http/httptest"
	"net/url"

	"testing"

	searchC "github.com/ONSdigital/dp-api-clients-go/site-search"
	. "github.com/smartystreets/goconvey/convey"
)

func TestUnitFilterMaps(t *testing.T) {
	t.Parallel()

	Convey("When mapFilterTypes is called", t, func() {
		ctx := context.Background()

		Convey("successfully map one filter given", func() {
			req := httptest.NewRequest("GET", "/search?q=housing&filter=article", nil)
			query := req.URL.Query()
			apiQuery, err := mapFilterTypes(ctx, query)
			So(apiQuery["content_type"], ShouldResemble, []string{"article,article_download"})
			So(err, ShouldBeNil)
		})

		Convey("successfully map two or more filters given", func() {
			req := httptest.NewRequest("GET", "/search?q=housing&filter=article&filter=compendia", nil)
			query := req.URL.Query()
			apiQuery, err := mapFilterTypes(ctx, query)
			So(apiQuery["content_type"], ShouldResemble, []string{"article,article_download,compendium_landing_page"})
			So(err, ShouldBeNil)
		})

		Convey("successfully map no filters given", func() {
			req := httptest.NewRequest("GET", "/search?q=housing", nil)
			query := req.URL.Query()
			apiQuery, err := mapFilterTypes(ctx, query)
			So(apiQuery["content_type"], ShouldBeNil)
			So(err, ShouldBeNil)
		})

		Convey("return error when mapping bad filters", func() {
			req := httptest.NewRequest("GET", "/search?q=housing&filter=INVALID", nil)
			query := req.URL.Query()
			apiQuery, err := mapFilterTypes(ctx, query)
			So(apiQuery["content_type"], ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err, ShouldResemble, errors.New("invalid filter type given"))
		})
	})

	Convey("When mapCountFilterTypes is called", t, func() {
		ctx := context.Background()
		mockedAPIQuery := url.Values{
			"content_type": []string{"bulletin,article,article_download"},
			"q":            []string{"housing"},
		}
		countResp := searchC.Response{
			ContentTypes: []searchC.ContentType{
				{
					Count: 3,
					Type:  "bulletin",
				},
				{
					Count: 4,
					Type:  "article",
				},
				{
					Count: 1,
					Type:  "article_download",
				},
			},
		}
		mockedSearchClient := &SearchClientMock{
			GetSearchFunc: func(ctx context.Context, query url.Values) (searchC.Response, error) {
				return countResp, nil
			},
		}

		Convey("return error as unable to retrieve count response from search client", func() {
			mockedSearchClient = &SearchClientMock{
				GetSearchFunc: func(ctx context.Context, query url.Values) (searchC.Response, error) {
					return searchC.Response{}, errors.New("internal server error")
				},
			}
			mappedContentType, err := mapCountFilterTypes(ctx, mockedAPIQuery, mockedSearchClient)
			So(mappedContentType, ShouldBeNil)
			So(err, ShouldNotBeNil)
		})

		Convey("return error when filter given by client not available in map", func() {
			invalidFilterResponse := searchC.Response{
				ContentTypes: []searchC.ContentType{
					{
						Count: 3,
						Type:  "invalid",
					},
				},
			}
			mockedSearchClient = &SearchClientMock{
				GetSearchFunc: func(ctx context.Context, query url.Values) (searchC.Response, error) {
					return invalidFilterResponse, nil
				},
			}
			mappedContentType, err := mapCountFilterTypes(ctx, mockedAPIQuery, mockedSearchClient)
			So(mappedContentType, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err, ShouldResemble, errors.New("filter type from client not available in filterTypes map"))
		})

		Convey("successfully retrieve the count of filter mapping to single filter type", func() {
			mockedAPIQuery = url.Values{
				"content_type": []string{"bulletin"},
				"q":            []string{"housing"},
			}
			singleFilterResponse := searchC.Response{
				ContentTypes: []searchC.ContentType{
					{
						Count: 3,
						Type:  "bulletin",
					},
				},
			}
			mockedSearchClient = &SearchClientMock{
				GetSearchFunc: func(ctx context.Context, query url.Values) (searchC.Response, error) {
					return singleFilterResponse, nil
				},
			}
			mappedContentType, err := mapCountFilterTypes(ctx, mockedAPIQuery, mockedSearchClient)
			So(mappedContentType, ShouldNotBeNil)
			So(mappedContentType, ShouldResemble, singleFilterResponse.ContentTypes)
			So(err, ShouldBeNil)
		})

		Convey("successfully retrieve the count of filter types mapping to multiple filter types", func() {
			mockedAPIQuery = url.Values{
				"content_type": []string{"bulletin,article,article_download,static_article"},
				"q":            []string{"housing"},
			}
			mappedContentType, err := mapCountFilterTypes(ctx, mockedAPIQuery, mockedSearchClient)
			multipleFilterContentType := []searchC.ContentType{
				{
					Count: 3,
					Type:  "bulletin",
				},
				{
					Count: 5,
					Type:  "article",
				},
			}
			So(mappedContentType, ShouldNotBeNil)
			So(mappedContentType, ShouldResemble, multipleFilterContentType)
			So(err, ShouldBeNil)
		})

	})
}
