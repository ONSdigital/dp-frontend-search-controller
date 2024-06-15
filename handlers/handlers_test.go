package handlers

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	zebedeeC "github.com/ONSdigital/dp-api-clients-go/v2/zebedee"
	"github.com/ONSdigital/dp-frontend-search-controller/apperrors"
	"github.com/ONSdigital/dp-frontend-search-controller/cache"
	"github.com/ONSdigital/dp-frontend-search-controller/config"
	"github.com/ONSdigital/dp-frontend-search-controller/data"
	"github.com/ONSdigital/dp-frontend-search-controller/mapper"
	coreModel "github.com/ONSdigital/dp-renderer/v2/model"
	searchModels "github.com/ONSdigital/dp-search-api/models"
	searchSDK "github.com/ONSdigital/dp-search-api/sdk"
	apiError "github.com/ONSdigital/dp-search-api/sdk/errors"
	topicModels "github.com/ONSdigital/dp-topic-api/models"
	"github.com/gorilla/mux"
	. "github.com/smartystreets/goconvey/convey"
)

const englishLang = "en"

type mockClientError struct{}

func (e *mockClientError) Error() string { return "client error" }
func (e *mockClientError) Code() int     { return http.StatusNotFound }

// doTestRequest helper function that creates a router and mocks requests
func doTestRequest(target string, req *http.Request, handlerFunc http.HandlerFunc, w *httptest.ResponseRecorder) *httptest.ResponseRecorder {
	if w == nil {
		w = httptest.NewRecorder()
	}
	router := mux.NewRouter()
	router.Path(target).HandlerFunc(handlerFunc)
	router.ServeHTTP(w, req)
	return w
}

var (
	accessToken  string
	collectionID string

	mockCensusTopic = &cache.Topic{
		ID:              "1234",
		LocaliseKeyName: "Census",
		Query:           "1234",
		List:            &cache.Subtopics{},
	}
)

func TestUnitReadHandlerSuccess(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	mockSearchResponse, err := mapper.GetMockSearchResponse()
	if err != nil {
		t.Errorf("failed to retrieve mock search response for unit tests, failing early: %v", err)
	}

	mockHomepageContent, err := mapper.GetMockHomepageContent()
	if err != nil {
		t.Errorf("failed to retrieve mock homepage content for unit tests, failing early: %v", err)
	}

	Convey("Given a valid request", t, func() {
		req := httptest.NewRequest("GET", "/search?q=housing&filter=bulletin&topics=1234", http.NoBody)
		req.Header.Set("X-Florence-Token", "testuser")

		cfg, err := config.Get()
		So(err, ShouldBeNil)

		mockedRendererClient := &RenderClientMock{
			BuildPageFunc: func(w io.Writer, pageModel interface{}, templateName string) {},
			NewBasePageModelFunc: func() coreModel.Page {
				return coreModel.Page{}
			},
		}

		mockedSearchClient := &SearchClientMock{
			GetSearchFunc: func(ctx context.Context, options searchSDK.Options) (*searchModels.SearchResponse, apiError.Error) {
				return mockSearchResponse, nil
			},
		}

		mockedZebedeeClient := &ZebedeeClientMock{
			GetHomepageContentFunc: func(ctx context.Context, userAuthToken, collectionID, lang, path string) (zebedeeC.HomepageContent, error) {
				return mockHomepageContent, nil
			},
		}

		mockedTopicClient := &TopicClientMock{}

		mockHandlerClients := NewHandlerClients(mockedRendererClient, mockedSearchClient, mockedZebedeeClient, mockedTopicClient)

		mockCacheList, err := cache.GetMockCacheList(ctx, englishLang)
		So(err, ShouldBeNil)

		Convey("When Read is called", func() {
			w := doTestRequest("/search", req, Read(cfg, mockHandlerClients, *mockCacheList, "search"), nil)

			Convey("Then a 200 OK status should be returned", func() {
				So(w.Code, ShouldEqual, http.StatusOK)

				So(mockedRendererClient.BuildPageCalls(), ShouldHaveLength, 1)
				So(mockedZebedeeClient.GetHomepageContentCalls(), ShouldHaveLength, 1)
				So(mockedSearchClient.GetSearchCalls(), ShouldHaveLength, 2)

				if mockedSearchClient.calls.GetSearch[0].Options.Query.Has("topics") {
					So(mockedSearchClient.calls.GetSearch[0].Options.Query.Get("topics"), ShouldEqual, "1234")
					So(mockedSearchClient.calls.GetSearch[1].Options.Query, ShouldNotContainKey, "topics")
				} else {
					So(mockedSearchClient.calls.GetSearch[1].Options.Query, ShouldContainKey, "topics")
					So(mockedSearchClient.calls.GetSearch[1].Options.Query.Get("topics"), ShouldEqual, "1234")
				}

				if mockedSearchClient.calls.GetSearch[0].Options.Query.Has("content_type") {
					So(mockedSearchClient.calls.GetSearch[0].Options.Query.Get("content_type"), ShouldEqual, "bulletin")
					So(mockedSearchClient.calls.GetSearch[1].Options.Query, ShouldNotContainKey, "content_type")
				} else {
					So(mockedSearchClient.calls.GetSearch[1].Options.Query, ShouldContainKey, "content_type")
					So(mockedSearchClient.calls.GetSearch[1].Options.Query.Get("content_type"), ShouldEqual, "bulletin")
				}
			})
		})
	})
}

func TestUnitReadSuccess(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	mockSearchResponse, err := mapper.GetMockSearchResponse()
	if err != nil {
		t.Errorf("failed to retrieve mock search response for unit tests, failing early: %v", err)
	}

	mockHomepageContent, err := mapper.GetMockHomepageContent()
	if err != nil {
		t.Errorf("failed to retrieve mock homepage content for unit tests, failing early: %v", err)
	}

	Convey("Given a valid request", t, func() {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/search?q=housing", http.NoBody)

		cfg, err := config.Get()
		So(err, ShouldBeNil)

		mockedRendererClient := &RenderClientMock{
			BuildPageFunc: func(w io.Writer, pageModel interface{}, templateName string) {},
			NewBasePageModelFunc: func() coreModel.Page {
				return coreModel.Page{}
			},
		}

		mockedSearchClient := &SearchClientMock{
			GetSearchFunc: func(ctx context.Context, options searchSDK.Options) (*searchModels.SearchResponse, apiError.Error) {
				return mockSearchResponse, nil
			},
		}

		mockedZebedeeClient := &ZebedeeClientMock{
			GetHomepageContentFunc: func(ctx context.Context, userAuthToken, collectionID, lang, path string) (zebedeeC.HomepageContent, error) {
				return mockHomepageContent, nil
			},
		}

		mockCacheList, err := cache.GetMockCacheList(ctx, englishLang)
		So(err, ShouldBeNil)

		Convey("When read is called", func() {
			read(w, req, cfg, mockedZebedeeClient, mockedRendererClient, mockedSearchClient, accessToken, collectionID, englishLang, *mockCacheList, "search", false)

			Convey("Then a 200 OK status should be returned", func() {
				So(w.Code, ShouldEqual, http.StatusOK)

				So(mockedRendererClient.BuildPageCalls(), ShouldHaveLength, 1)
				So(mockedSearchClient.GetSearchCalls(), ShouldHaveLength, 2)
				So(mockedZebedeeClient.GetHomepageContentCalls(), ShouldHaveLength, 1)

				So(mockedSearchClient.GetSearchCalls()[0].Options.Query.Get("nlp_weighting"), ShouldEqual, "false")
			})
		})

		Convey("When read is called with NLP switched on", func() {
			read(w, req, cfg, mockedZebedeeClient, mockedRendererClient, mockedSearchClient, accessToken, collectionID, englishLang, *mockCacheList, "search", true)

			Convey("Then a 200 OK status should be returned", func() {
				So(w.Code, ShouldEqual, http.StatusOK)

				So(mockedRendererClient.BuildPageCalls(), ShouldHaveLength, 1)
				So(mockedSearchClient.GetSearchCalls(), ShouldHaveLength, 2)
				So(mockedZebedeeClient.GetHomepageContentCalls(), ShouldHaveLength, 1)

				So(mockedSearchClient.GetSearchCalls()[0].Options.Query.Get("nlp_weighting"), ShouldEqual, "true")
			})
		})
	})
}

func TestUnitReadFailure(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	mockSearchResponse, err := mapper.GetMockSearchResponse()
	if err != nil {
		t.Errorf("failed to retrieve mock search response for unit tests, failing early: %v", err)
	}

	mockHomepageContent, err := mapper.GetMockHomepageContent()
	fmt.Printf("%+v\n", mockHomepageContent)
	if err != nil {
		t.Errorf("failed to retrieve mock homepage content for unit tests, failing early: %v", err)
	}

	Convey("Given an error from failing to review query", t, func() {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/search?q=housing&page=1000000", http.NoBody)

		cfg, err := config.Get()
		So(err, ShouldBeNil)

		mockedRendererClient := &RenderClientMock{
			BuildPageFunc: func(w io.Writer, pageModel interface{}, templateName string) {},
			NewBasePageModelFunc: func() coreModel.Page {
				return coreModel.Page{}
			},
		}

		mockedSearchClient := &SearchClientMock{
			GetSearchFunc: func(ctx context.Context, options searchSDK.Options) (*searchModels.SearchResponse, apiError.Error) {
				return mockSearchResponse, nil
			},
		}

		mockedZebedeeClient := &ZebedeeClientMock{
			GetHomepageContentFunc: func(ctx context.Context, userAuthToken, collectionID, lang, path string) (zebedeeC.HomepageContent, error) {
				return mockHomepageContent, nil
			},
		}

		mockCacheList, err := cache.GetMockCacheList(ctx, englishLang)
		So(err, ShouldBeNil)

		Convey("When read is called", func() {
			read(w, req, cfg, mockedZebedeeClient, mockedRendererClient, mockedSearchClient, accessToken, collectionID, englishLang, *mockCacheList, "search", false)

			Convey("Then a 400 bad request status should be returned", func() {
				So(w.Code, ShouldEqual, http.StatusBadRequest)

				So(mockedRendererClient.BuildPageCalls(), ShouldHaveLength, 0)
				So(mockedSearchClient.GetSearchCalls(), ShouldHaveLength, 0)
				So(mockedZebedeeClient.GetHomepageContentCalls(), ShouldHaveLength, 0)
			})
		})
	})

	Convey("Given an error from failing to get search response from search client", t, func() {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/search?q=housing", http.NoBody)

		cfg, err := config.Get()
		So(err, ShouldBeNil)

		mockedRendererClient := &RenderClientMock{
			BuildPageFunc: func(w io.Writer, pageModel interface{}, templateName string) {},
			NewBasePageModelFunc: func() coreModel.Page {
				return coreModel.Page{}
			},
		}

		mockedSearchClient := &SearchClientMock{
			GetSearchFunc: func(ctx context.Context, options searchSDK.Options) (*searchModels.SearchResponse, apiError.Error) {
				return &searchModels.SearchResponse{}, apiError.StatusError{Code: 500}
			},
		}

		mockedZebedeeClient := &ZebedeeClientMock{
			GetHomepageContentFunc: func(ctx context.Context, userAuthToken, collectionID, lang, path string) (zebedeeC.HomepageContent, error) {
				return mockHomepageContent, nil
			},
		}

		mockCacheList, err := cache.GetMockCacheList(ctx, englishLang)
		So(err, ShouldBeNil)

		Convey("When read is called", func() {
			read(w, req, cfg, mockedZebedeeClient, mockedRendererClient, mockedSearchClient, accessToken, collectionID, englishLang, *mockCacheList, "search", false)

			Convey("Then a 500 internal server error status should be returned", func() {
				So(w.Code, ShouldEqual, http.StatusInternalServerError)

				So(mockedRendererClient.BuildPageCalls(), ShouldHaveLength, 0)
				So(mockedSearchClient.GetSearchCalls(), ShouldHaveLength, 2)
				So(mockedZebedeeClient.GetHomepageContentCalls(), ShouldHaveLength, 1)
			})
		})
	})

	Convey("Given an error as current page exceeds total pages", t, func() {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/search?q=housing&page=2", http.NoBody)

		cfg, err := config.Get()
		So(err, ShouldBeNil)

		mockedRendererClient := &RenderClientMock{
			BuildPageFunc: func(w io.Writer, pageModel interface{}, templateName string) {},
			NewBasePageModelFunc: func() coreModel.Page {
				return coreModel.Page{}
			},
		}

		mockedSearchClient := &SearchClientMock{
			GetSearchFunc: func(ctx context.Context, options searchSDK.Options) (*searchModels.SearchResponse, apiError.Error) {
				return mockSearchResponse, nil
			},
		}

		mockedZebedeeClient := &ZebedeeClientMock{
			GetHomepageContentFunc: func(ctx context.Context, userAuthToken, collectionID, lang, path string) (zebedeeC.HomepageContent, error) {
				return mockHomepageContent, nil
			},
		}

		mockCacheList, err := cache.GetMockCacheList(ctx, englishLang)
		So(err, ShouldBeNil)

		Convey("When read is called", func() {
			read(w, req, cfg, mockedZebedeeClient, mockedRendererClient, mockedSearchClient, accessToken, collectionID, englishLang, *mockCacheList, "search", false)

			Convey("Then a 400 bad request status should be returned", func() {
				So(w.Code, ShouldEqual, http.StatusBadRequest)

				So(mockedRendererClient.BuildPageCalls(), ShouldHaveLength, 0)
				So(mockedSearchClient.GetSearchCalls(), ShouldHaveLength, 2)
				So(mockedZebedeeClient.GetHomepageContentCalls(), ShouldHaveLength, 1)
			})
		})
	})
}

func TestUnitReadDataAggregationSuccess(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	mockSearchResponse, err := mapper.GetMockSearchResponse()
	if err != nil {
		t.Errorf("failed to retrieve mock search response for unit tests, failing early: %v", err)
	}

	mockHomepageContent, err := mapper.GetMockHomepageContent()
	if err != nil {
		t.Errorf("failed to retrieve mock homepage content for unit tests, failing early: %v", err)
	}

	Convey("Given a valid request for a an aggregated data page and a set of mocked services", t, func() {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/publications", http.NoBody)

		cfg, err := config.Get()
		So(err, ShouldBeNil)

		mockedRendererClient := &RenderClientMock{
			BuildPageFunc: func(w io.Writer, pageModel interface{}, templateName string) {},
			NewBasePageModelFunc: func() coreModel.Page {
				return coreModel.Page{}
			},
		}

		mockedSearchClient := &SearchClientMock{
			GetSearchFunc: func(ctx context.Context, options searchSDK.Options) (*searchModels.SearchResponse, apiError.Error) {
				return mockSearchResponse, nil
			},
		}

		mockedZebedeeClient := &ZebedeeClientMock{
			GetHomepageContentFunc: func(ctx context.Context, userAuthToken, collectionID, lang, path string) (zebedeeC.HomepageContent, error) {
				return mockHomepageContent, nil
			},
		}

		mockCacheList, err := cache.GetMockCacheList(ctx, englishLang)
		So(err, ShouldBeNil)

		Convey("When readDataAggregationWithTopics is called", func() {
			readDataAggregation(w, req, cfg, mockedZebedeeClient, mockedRendererClient, mockedSearchClient, accessToken, collectionID, englishLang, *mockCacheList, "home-publications")

			Convey("Then a 200 OK status should be returned", func() {
				So(w.Code, ShouldEqual, http.StatusOK)

				So(mockedRendererClient.BuildPageCalls(), ShouldHaveLength, 1)
				So(mockedSearchClient.GetSearchCalls(), ShouldHaveLength, 2)
				So(mockedZebedeeClient.GetHomepageContentCalls(), ShouldHaveLength, 1)
			})

			Convey("And the Search Client should be called with pre-configured filters", func() {
				_, searchCall := sortSearchCalls(mockedSearchClient.GetSearchCalls()[0], mockedSearchClient.GetSearchCalls()[1], "content_type")

				expectedContentTypes := []string{"bulletin,article,article_download,compendium_landing_page"}

				searchContentTypeParam := searchCall.Options.Query["content_type"]

				So(searchContentTypeParam, ShouldEqual, expectedContentTypes)
			})
		})
	})
}

func TestUnitReadDataAggregationFailure(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	mockSearchResponse, err := mapper.GetMockSearchResponse()
	if err != nil {
		t.Errorf("failed to retrieve mock search response for unit tests, failing early: %v", err)
	}

	mockHomepageContent, err := mapper.GetMockHomepageContent()
	if err != nil {
		t.Errorf("failed to retrieve mock homepage content for unit tests, failing early: %v", err)
	}

	Convey("Given a request for a aggregated data page with an invalid page param", t, func() {
		cfg, err := config.Get()
		So(err, ShouldBeNil)

		invalidPageParam := (cfg.DefaultMaximumSearchResults / cfg.DefaultLimit) + 10

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", fmt.Sprintf("/publications?page=%d", invalidPageParam), http.NoBody)

		mockedRendererClient := &RenderClientMock{
			BuildPageFunc: func(w io.Writer, pageModel interface{}, templateName string) {},
			NewBasePageModelFunc: func() coreModel.Page {
				return coreModel.Page{}
			},
		}

		mockedSearchClient := &SearchClientMock{
			GetSearchFunc: func(ctx context.Context, options searchSDK.Options) (*searchModels.SearchResponse, apiError.Error) {
				return mockSearchResponse, nil
			},
		}

		mockedZebedeeClient := &ZebedeeClientMock{
			GetHomepageContentFunc: func(ctx context.Context, userAuthToken, collectionID, lang, path string) (zebedeeC.HomepageContent, error) {
				return mockHomepageContent, nil
			},
		}

		mockCacheList, err := cache.GetMockCacheList(ctx, englishLang)
		So(err, ShouldBeNil)

		Convey("When readDataAggregation is called", func() {
			readDataAggregation(w, req, cfg, mockedZebedeeClient, mockedRendererClient, mockedSearchClient, accessToken, collectionID, englishLang, *mockCacheList, "home-publications")

			Convey("Then a 200 OK status should be returned", func() {
				So(w.Code, ShouldEqual, http.StatusOK)

				So(mockedRendererClient.BuildPageCalls(), ShouldHaveLength, 1)
			})

			Convey("And no calls should be made to downstream services", func() {
				So(mockedSearchClient.GetSearchCalls(), ShouldHaveLength, 0)
			})
		})
	})

	Convey("Given a request for a aggregated data page with an invalid date param", t, func() {
		cfg, err := config.Get()
		So(err, ShouldBeNil)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/alladhocs?after-month=13&after-year=2024", http.NoBody)

		mockedRendererClient := &RenderClientMock{
			BuildPageFunc: func(w io.Writer, pageModel interface{}, templateName string) {},
			NewBasePageModelFunc: func() coreModel.Page {
				return coreModel.Page{}
			},
		}

		mockedSearchClient := &SearchClientMock{
			GetSearchFunc: func(ctx context.Context, options searchSDK.Options) (*searchModels.SearchResponse, apiError.Error) {
				return mockSearchResponse, nil
			},
		}

		mockedZebedeeClient := &ZebedeeClientMock{
			GetHomepageContentFunc: func(ctx context.Context, userAuthToken, collectionID, lang, path string) (zebedeeC.HomepageContent, error) {
				return mockHomepageContent, nil
			},
		}

		mockCacheList, err := cache.GetMockCacheList(ctx, englishLang)
		So(err, ShouldBeNil)

		Convey("When readDataAggregation is called", func() {
			readDataAggregation(w, req, cfg, mockedZebedeeClient, mockedRendererClient, mockedSearchClient, accessToken, collectionID, englishLang, *mockCacheList, "all-adhocs")

			Convey("Then a 200 OK status should be returned", func() {
				So(w.Code, ShouldEqual, http.StatusOK)

				So(mockedRendererClient.BuildPageCalls(), ShouldHaveLength, 1)
			})

			Convey("And no calls should be made to downstream services", func() {
				So(mockedSearchClient.GetSearchCalls(), ShouldHaveLength, 0)
			})
		})
	})
}

func TestUnitReadDataAggregationWithTopicsSuccess(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	mockSearchResponse, err := mapper.GetMockSearchResponse()
	if err != nil {
		t.Errorf("failed to retrieve mock search response for unit tests, failing early: %v", err)
	}

	mockHomepageContent, err := mapper.GetMockHomepageContent()
	if err != nil {
		t.Errorf("failed to retrieve mock homepage content for unit tests, failing early: %v", err)
	}

	Convey("Given a valid request for a topic filtered page and a set of mocked services", t, func() {
		testTopic := topicModels.Topic{
			ID:    "6734",
			Title: "economy",
		}

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", fmt.Sprintf("/%s/publications", testTopic.Title), http.NoBody)
		req = mux.SetURLVars(req, map[string]string{"topicsPath": testTopic.Title})

		cfg, err := config.Get()
		So(err, ShouldBeNil)

		mockedRendererClient := &RenderClientMock{
			BuildPageFunc: func(w io.Writer, pageModel interface{}, templateName string) {},
			NewBasePageModelFunc: func() coreModel.Page {
				return coreModel.Page{}
			},
		}

		mockedSearchClient := &SearchClientMock{
			GetSearchFunc: func(ctx context.Context, options searchSDK.Options) (*searchModels.SearchResponse, apiError.Error) {
				return mockSearchResponse, nil
			},
		}

		mockedZebedeeClient := &ZebedeeClientMock{
			GetHomepageContentFunc: func(ctx context.Context, userAuthToken, collectionID, lang, path string) (zebedeeC.HomepageContent, error) {
				return mockHomepageContent, nil
			},
		}

		mockCacheList, err := cache.GetMockCacheList(ctx, englishLang)
		So(err, ShouldBeNil)

		Convey("When readDataAggregationWithTopics is called", func() {
			readDataAggregationWithTopics(w, req, cfg, mockedZebedeeClient, mockedRendererClient, mockedSearchClient, accessToken, collectionID, englishLang, *mockCacheList, "publications")

			Convey("Then a 200 OK status should be returned", func() {
				So(w.Code, ShouldEqual, http.StatusOK)

				So(mockedRendererClient.BuildPageCalls(), ShouldHaveLength, 1)
				So(mockedSearchClient.GetSearchCalls(), ShouldHaveLength, 2)
				So(mockedZebedeeClient.GetHomepageContentCalls(), ShouldHaveLength, 1)
			})

			Convey("And the Search Client should be called with pre-configured filters", func() {
				_, searchCall := sortSearchCalls(mockedSearchClient.GetSearchCalls()[0], mockedSearchClient.GetSearchCalls()[1], "topics")

				expectedContentTypes := []string{testTopic.ID}

				searchContentTypeParam := searchCall.Options.Query["topics"]

				So(searchContentTypeParam, ShouldEqual, expectedContentTypes)
			})
		})
	})

	Convey("Given a valid request for a subtopic filtered page and a set of mocked services", t, func() {
		testTopic := topicModels.Topic{
			ID:    "6734",
			Slug:  "economy",
			Title: "Economy",
		}

		testSubtopic := topicModels.Topic{
			ID:    "1834",
			Slug:  "environmentalaccounts",
			Title: "Environmental Accounts",
		}

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", fmt.Sprintf("/%s/%s/publications", testTopic.Slug, testSubtopic.Slug), http.NoBody)
		req = mux.SetURLVars(req, map[string]string{"topicsPath": testTopic.Slug + "/" + testSubtopic.Slug})

		cfg, err := config.Get()
		So(err, ShouldBeNil)

		mockedRendererClient := &RenderClientMock{
			BuildPageFunc: func(w io.Writer, pageModel interface{}, templateName string) {},
			NewBasePageModelFunc: func() coreModel.Page {
				return coreModel.Page{}
			},
		}

		mockedSearchClient := &SearchClientMock{
			GetSearchFunc: func(ctx context.Context, options searchSDK.Options) (*searchModels.SearchResponse, apiError.Error) {
				return mockSearchResponse, nil
			},
		}

		mockedZebedeeClient := &ZebedeeClientMock{
			GetHomepageContentFunc: func(ctx context.Context, userAuthToken, collectionID, lang, path string) (zebedeeC.HomepageContent, error) {
				return mockHomepageContent, nil
			},
		}

		mockCacheList, err := cache.GetMockCacheList(ctx, englishLang)
		So(err, ShouldBeNil)

		Convey("When readDataAggregationWithTopics is called", func() {
			readDataAggregationWithTopics(w, req, cfg, mockedZebedeeClient, mockedRendererClient, mockedSearchClient, accessToken, collectionID, englishLang, *mockCacheList, "publications")

			Convey("Then a 200 OK status should be returned", func() {
				So(w.Code, ShouldEqual, http.StatusOK)

				So(mockedRendererClient.BuildPageCalls(), ShouldHaveLength, 1)
				So(mockedSearchClient.GetSearchCalls(), ShouldHaveLength, 2)
				So(mockedZebedeeClient.GetHomepageContentCalls(), ShouldHaveLength, 1)
			})

			Convey("And the Search Client should be called with the subtopic id from the topic API", func() {
				_, searchCall := sortSearchCalls(mockedSearchClient.GetSearchCalls()[0], mockedSearchClient.GetSearchCalls()[1], "topics")

				expectedContentTypes := []string{testSubtopic.ID}

				searchContentTypeParam := searchCall.Options.Query["topics"]

				So(searchContentTypeParam, ShouldEqual, expectedContentTypes)
			})
		})
	})

	Convey("Given a valid request for a 3rd level subtopic filtered page and a set of mocked services", t, func() {
		testTopic := topicModels.Topic{
			ID:    "6734",
			Slug:  "economy",
			Title: "Economy",
		}

		testSubtopic := topicModels.Topic{
			ID:    "8268",
			Slug:  "governmentpublicsectorandtaxes",
			Title: "Government Public Sector and Taxes",
		}

		testSubSubTopic := topicModels.Topic{
			ID:    "3687",
			Slug:  "publicsectorfinance",
			Title: "Public Sector Finance",
		}

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", fmt.Sprintf("/%s/%s/%s/publications", testTopic.Slug, testSubtopic.Slug, testSubSubTopic.Slug), http.NoBody)
		req = mux.SetURLVars(req, map[string]string{"topicsPath": testTopic.Slug + "/" + testSubtopic.Slug + "/" + testSubSubTopic.Slug})

		cfg, err := config.Get()
		So(err, ShouldBeNil)

		mockedRendererClient := &RenderClientMock{
			BuildPageFunc: func(w io.Writer, pageModel interface{}, templateName string) {},
			NewBasePageModelFunc: func() coreModel.Page {
				return coreModel.Page{}
			},
		}

		mockedSearchClient := &SearchClientMock{
			GetSearchFunc: func(ctx context.Context, options searchSDK.Options) (*searchModels.SearchResponse, apiError.Error) {
				return mockSearchResponse, nil
			},
		}

		mockedZebedeeClient := &ZebedeeClientMock{
			GetHomepageContentFunc: func(ctx context.Context, userAuthToken, collectionID, lang, path string) (zebedeeC.HomepageContent, error) {
				return mockHomepageContent, nil
			},
		}

		mockCacheList, err := cache.GetMockCacheList(ctx, englishLang)
		So(err, ShouldBeNil)

		Convey("When readDataAggregationWithTopics is called", func() {
			readDataAggregationWithTopics(w, req, cfg, mockedZebedeeClient, mockedRendererClient, mockedSearchClient, accessToken, collectionID, englishLang, *mockCacheList, "publications")

			Convey("Then a 200 OK status should be returned", func() {
				So(w.Code, ShouldEqual, http.StatusOK)

				So(mockedRendererClient.BuildPageCalls(), ShouldHaveLength, 1)
				So(mockedSearchClient.GetSearchCalls(), ShouldHaveLength, 2)
				So(mockedZebedeeClient.GetHomepageContentCalls(), ShouldHaveLength, 1)
			})

			Convey("And the Search Client should be called with pre-configured filters", func() {
				_, searchCall := sortSearchCalls(mockedSearchClient.GetSearchCalls()[0], mockedSearchClient.GetSearchCalls()[1], "topics")

				expectedContentTypes := []string{testSubSubTopic.ID}

				searchContentTypeParam := searchCall.Options.Query["topics"]

				So(searchContentTypeParam, ShouldEqual, expectedContentTypes)
			})
		})
	})
}

func TestUnitReadDataAggregationWithTopicsFailure(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	mockSearchResponse, err := mapper.GetMockSearchResponse()
	if err != nil {
		t.Errorf("failed to retrieve mock search response for unit tests, failing early: %v", err)
	}

	mockHomepageContent, err := mapper.GetMockHomepageContent()
	if err != nil {
		t.Errorf("failed to retrieve mock homepage content for unit tests, failing early: %v", err)
	}

	Convey("Given a request for a topic filtered page and a set of mocked services where the topic does not exist", t, func() {
		testTopic := "thisdoesnotexist"

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", fmt.Sprintf("/%s/publications", testTopic), http.NoBody)
		req = mux.SetURLVars(req, map[string]string{"topic": testTopic})

		cfg, err := config.Get()
		So(err, ShouldBeNil)

		mockedRendererClient := &RenderClientMock{
			BuildPageFunc: func(w io.Writer, pageModel interface{}, templateName string) {},
			NewBasePageModelFunc: func() coreModel.Page {
				return coreModel.Page{}
			},
		}

		mockedSearchClient := &SearchClientMock{
			GetSearchFunc: func(ctx context.Context, options searchSDK.Options) (*searchModels.SearchResponse, apiError.Error) {
				return mockSearchResponse, nil
			},
		}

		mockedZebedeeClient := &ZebedeeClientMock{
			GetHomepageContentFunc: func(ctx context.Context, userAuthToken, collectionID, lang, path string) (zebedeeC.HomepageContent, error) {
				return mockHomepageContent, nil
			},
		}

		mockCacheList, err := cache.GetMockCacheList(ctx, englishLang)
		So(err, ShouldBeNil)

		Convey("When readDataAggregationWithTopics is called", func() {
			readDataAggregationWithTopics(w, req, cfg, mockedZebedeeClient, mockedRendererClient, mockedSearchClient, accessToken, collectionID, englishLang, *mockCacheList, "publications")

			Convey("Then a 404 OK status should be returned", func() {
				So(w.Code, ShouldEqual, http.StatusNotFound)

				So(mockedRendererClient.BuildPageCalls(), ShouldHaveLength, 0)
				So(mockedSearchClient.GetSearchCalls(), ShouldHaveLength, 0)
				So(mockedZebedeeClient.GetHomepageContentCalls(), ShouldHaveLength, 0)
			})
		})
	})

	Convey("Given a request for a subtopic filtered page and a set of mocked services where the subtopic does not exist", t, func() {
		testTopic := topicModels.Topic{
			ID:    "6734",
			Title: "economy",
		}
		testSubtopic := "thissubtopicdoesnotexist"

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", fmt.Sprintf("/%s/%s/publications", testTopic.Title, testSubtopic), http.NoBody)
		req = mux.SetURLVars(req, map[string]string{"topicsPath": testTopic.Slug + "/" + testSubtopic})

		cfg, err := config.Get()
		So(err, ShouldBeNil)

		mockedRendererClient := &RenderClientMock{
			BuildPageFunc: func(w io.Writer, pageModel interface{}, templateName string) {},
			NewBasePageModelFunc: func() coreModel.Page {
				return coreModel.Page{}
			},
		}

		mockedSearchClient := &SearchClientMock{
			GetSearchFunc: func(ctx context.Context, options searchSDK.Options) (*searchModels.SearchResponse, apiError.Error) {
				return mockSearchResponse, nil
			},
		}

		mockedZebedeeClient := &ZebedeeClientMock{
			GetHomepageContentFunc: func(ctx context.Context, userAuthToken, collectionID, lang, path string) (zebedeeC.HomepageContent, error) {
				return mockHomepageContent, nil
			},
		}
		mockCacheList, err := cache.GetMockCacheList(ctx, englishLang)
		So(err, ShouldBeNil)

		Convey("When readDataAggregationWithTopics is called", func() {
			readDataAggregationWithTopics(w, req, cfg, mockedZebedeeClient, mockedRendererClient, mockedSearchClient, accessToken, collectionID, englishLang, *mockCacheList, "publications")

			Convey("Then a 404 OK status should be returned", func() {
				So(w.Code, ShouldEqual, http.StatusNotFound)

				So(mockedRendererClient.BuildPageCalls(), ShouldHaveLength, 0)
				So(mockedSearchClient.GetSearchCalls(), ShouldHaveLength, 0)
				So(mockedZebedeeClient.GetHomepageContentCalls(), ShouldHaveLength, 0)
			})
		})
	})

	Convey("Given a request for a 3rd level subtopic filtered page and a set of mocked services where the 3rd level subtopic does not exist", t, func() {
		testTopic := topicModels.Topic{
			ID:    "6734",
			Slug:  "economy",
			Title: "Economy",
		}

		testSubtopic := topicModels.Topic{
			ID:    "8268",
			Slug:  "governmentpublicsectorandtaxes",
			Title: "Government Public Sector and Taxes",
		}

		testSubSubTopic := "thissubtopicdoesnotexist"

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", fmt.Sprintf("/%s/%s/%s/publications", testTopic.Slug, testSubtopic.Slug, testSubSubTopic), http.NoBody)
		req = mux.SetURLVars(req, map[string]string{"topicsPath": testTopic.Slug + "/" + testSubtopic.Slug + "/" + testSubSubTopic})

		cfg, err := config.Get()
		So(err, ShouldBeNil)

		mockedRendererClient := &RenderClientMock{
			BuildPageFunc: func(w io.Writer, pageModel interface{}, templateName string) {},
			NewBasePageModelFunc: func() coreModel.Page {
				return coreModel.Page{}
			},
		}

		mockedSearchClient := &SearchClientMock{
			GetSearchFunc: func(ctx context.Context, options searchSDK.Options) (*searchModels.SearchResponse, apiError.Error) {
				return mockSearchResponse, nil
			},
		}

		mockedZebedeeClient := &ZebedeeClientMock{
			GetHomepageContentFunc: func(ctx context.Context, userAuthToken, collectionID, lang, path string) (zebedeeC.HomepageContent, error) {
				return mockHomepageContent, nil
			},
		}
		mockCacheList, err := cache.GetMockCacheList(ctx, englishLang)
		So(err, ShouldBeNil)

		Convey("When readDataAggregationWithTopics is called", func() {
			readDataAggregationWithTopics(w, req, cfg, mockedZebedeeClient, mockedRendererClient, mockedSearchClient, accessToken, collectionID, englishLang, *mockCacheList, "publications")

			Convey("Then a 404 OK status should be returned", func() {
				So(w.Code, ShouldEqual, http.StatusNotFound)

				So(mockedRendererClient.BuildPageCalls(), ShouldHaveLength, 0)
				So(mockedSearchClient.GetSearchCalls(), ShouldHaveLength, 0)
				So(mockedZebedeeClient.GetHomepageContentCalls(), ShouldHaveLength, 0)
			})
		})
	})
}

func TestUnitValidateCurrentPageSuccess(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	cfg, _ := config.Get()

	Convey("Given number of search results is more than 0", t, func() {
		resultsCount := 10

		Convey("And current page doesn't exceed total pages", func() {
			validatedQueryParams := data.SearchURLParams{
				Limit:       10,
				CurrentPage: 1,
			}

			Convey("When validateCurrentPage is called", func() {
				err := validateCurrentPage(ctx, cfg, validatedQueryParams, resultsCount)

				Convey("Then return no error", func() {
					So(err, ShouldBeNil)
				})
			})
		})
	})

	Convey("Given number of search results is 0", t, func() {
		resultsCount := 0

		Convey("And current page doesn't exceed total pages", func() {
			validatedQueryParams := data.SearchURLParams{
				Limit:       10,
				CurrentPage: 1,
			}

			Convey("When validateCurrentPage is called", func() {
				err := validateCurrentPage(ctx, cfg, validatedQueryParams, resultsCount)

				Convey("Then return no error", func() {
					So(err, ShouldBeNil)
				})
			})
		})
	})
}

func TestUnitValidateCurrentPageFailure(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	cfg, _ := config.Get()

	Convey("Given current page exceeds total pages", t, func() {
		validatedQueryParams := data.SearchURLParams{
			Limit:       10,
			CurrentPage: 10000,
		}

		Convey("And number of search results is more than 0", func() {
			resultsCount := 20

			Convey("When validateCurrentPage is called", func() {
				err := validateCurrentPage(ctx, cfg, validatedQueryParams, resultsCount)

				Convey("Then return no error", func() {
					So(err, ShouldNotBeNil)
					So(err, ShouldResemble, apperrors.ErrPageExceedsTotalPages)
				})
			})
		})
	})
}

func TestUnitGetCategoriesTypesCountSuccess(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	mockSearchResponse, err := mapper.GetMockSearchResponse()
	if err != nil {
		t.Errorf("failed to retrieve mock search response for unit tests, failing early: %v", err)
	}

	Convey("Given api query and search client", t, func() {
		apiQuery := url.Values{
			"q":            []string{"housing"},
			"content_type": []string{"bulletin"},
			"topics":       []string{"1234"},
			"sort":         []string{"relevance"},
			"limit":        []string{"10"},
			"offset":       []string{"0"},
		}

		mockedSearchClient := &SearchClientMock{
			GetSearchFunc: func(ctx context.Context, options searchSDK.Options) (*searchModels.SearchResponse, apiError.Error) {
				return mockSearchResponse, nil
			},
		}

		Convey("When getCategoriesTypesCount is called", func() {
			categories, topicCategories, err := getCategoriesTypesCount(ctx, accessToken, collectionID, apiQuery, mockedSearchClient, mockCensusTopic)

			Convey("Then return all categories and types with its count", func() {
				So(categories[0].Count, ShouldEqual, 1)
				So(categories[0].ContentTypes[1].Count, ShouldEqual, 1)
				So(topicCategories[0].Count, ShouldEqual, 1)
				So(topicCategories[0].LocaliseKeyName, ShouldEqual, mockCensusTopic.LocaliseKeyName)
				So(topicCategories[0].Query, ShouldEqual, mockCensusTopic.Query)
			})

			Convey("And return no error", func() {
				So(err, ShouldBeNil)

				So(mockedSearchClient.GetSearchCalls(), ShouldHaveLength, 1)
			})
		})
	})
}

func TestUnitGetCategoriesTypesCountFailure(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	Convey("Given an error from failing to get search query count from search client", t, func() {
		apiQuery := url.Values{
			"q":            []string{"housing"},
			"content_type": []string{"bulletin"},
			"sort":         []string{"relevance"},
			"limit":        []string{"10"},
			"offset":       []string{"0"},
		}

		mockedSearchClient := &SearchClientMock{
			GetSearchFunc: func(ctx context.Context, options searchSDK.Options) (*searchModels.SearchResponse, apiError.Error) {
				return &searchModels.SearchResponse{}, apiError.StatusError{Code: 500}
			},
		}

		Convey("When getCategoriesTypesCount is called", func() {
			categories, topicCategories, err := getCategoriesTypesCount(ctx, accessToken, collectionID, apiQuery, mockedSearchClient, mockCensusTopic)

			Convey("Then return an error", func() {
				So(err, ShouldNotBeNil)
			})

			Convey("And return nil categories", func() {
				So(categories, ShouldBeNil)
				So(topicCategories, ShouldBeNil)
				So(mockedSearchClient.GetSearchCalls(), ShouldHaveLength, 1)
			})
		})
	})
}

func TestUnitSetCountToCategoriesSuccess(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	Convey("Given list of categories", t, func() {
		categories := data.GetCategories()

		Convey("And count of search results for each categories and types", func() {
			mockCountSearchResponse, err := mapper.GetMockSearchResponse()
			So(err, ShouldBeNil)

			Convey("When setCountToCategories is called", func() {
				setCountToCategories(ctx, mockCountSearchResponse, categories)

				Convey("Then the count should be updated in the list of categories", func() {
					So(categories[0].Count, ShouldEqual, 1)
					So(categories[0].ContentTypes[1].Count, ShouldEqual, 1)
				})
			})
		})
	})

	Convey("Given unrecognised filter type returned from api", t, func() {
		mockCountSearchResponse := searchModels.SearchResponse{
			Count: 1,
			ContentTypes: []searchModels.FilterCount{
				{
					Type:  "article",
					Count: 1,
				},
				{
					Type:  "unknown",
					Count: 1,
				},
			},
		}

		Convey("And list of categories", func() {
			categories := data.GetCategories()

			Convey("When setCountToCategories is called", func() {
				setCountToCategories(ctx, &mockCountSearchResponse, categories)

				Convey("Then the count should be updated in the list of known categories and warning given", func() {
					So(categories[0].Count, ShouldEqual, 1)
					So(categories[0].ContentTypes[1].Count, ShouldEqual, 1)
				})
			})
		})
	})
}

func TestUnitSetStatusCodeSuccess(t *testing.T) {
	t.Parallel()

	Convey("Given a internal server error", t, func() {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/search?q=housing&page=1000000", http.NoBody)

		err := apperrors.ErrInternalServer

		Convey("When setStatusCode is called", func() {
			setStatusCode(w, req, err)

			Convey("Then send a HTTP response header with 500 internal server error", func() {
				So(w.Code, ShouldEqual, http.StatusInternalServerError)
			})
		})
	})

	Convey("Given an client error", t, func() {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/search?q=housing", http.NoBody)

		err := &mockClientError{}

		Convey("When setStatusCode is called", func() {
			setStatusCode(w, req, err)

			Convey("Then send a HTTP response header with 404 status not found", func() {
				So(w.Code, ShouldEqual, http.StatusNotFound)
			})
		})
	})

	Convey("Given a bad request error", t, func() {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/search?q=housing&page=1000000", http.NoBody)

		err := apperrors.ErrInvalidPage

		Convey("When setStatusCode is called", func() {
			setStatusCode(w, req, err)

			Convey("Then send a HTTP response header with 400 bad request status", func() {
				So(w.Code, ShouldEqual, http.StatusBadRequest)
			})
		})
	})
}

func TestSetFlorenceTokenHeader(t *testing.T) {
	t.Parallel()

	Convey("Given a valid access token without 'Bearer' prefix", t, func() {
		headers := make(http.Header)
		accessToken := "accessToken"

		Convey("When setFlorenceTokenHeader is called", func() {
			setFlorenceTokenHeader(headers, accessToken)

			Convey("Then the FlorenceToken header should be set with 'Bearer' prefix", func() {
				So(headers.Get(searchSDK.FlorenceToken), ShouldEqual, "Bearer "+accessToken)
			})
		})
	})

	Convey("Given a valid access token with 'Bearer' prefix", t, func() {
		headers := make(http.Header)
		accessToken := "Bearer accessToken"

		Convey("When setFlorenceTokenHeader is called", func() {
			setFlorenceTokenHeader(headers, accessToken)

			Convey("Then the FlorenceToken header should be set with no additional 'Bearer' prefix", func() {
				So(headers.Get(searchSDK.FlorenceToken), ShouldEqual, accessToken)
			})
		})
	})
}

func TestCreateRSSFeed(t *testing.T) {
	t.Parallel()
	// Prepare test data
	req := httptest.NewRequest("GET", "http://localhost:27700", http.NoBody)
	w := httptest.NewRecorder()
	collectionID := "collection"
	accessToken := "token"
	validatedParams := data.SearchURLParams{}
	template := "all-adhocs"

	// Create a mock SearchClient
	mockSearchClient := &SearchClientMock{}

	Convey("when Search returns success", t, func() {
		// Define expected behavior for the mock SearchClient
		mockSearchResponse := &searchModels.SearchResponse{} // Create a mock response
		mockSearchClient.GetSearchFunc = func(ctx context.Context, options searchSDK.Options) (*searchModels.SearchResponse, apiError.Error) {
			return mockSearchResponse, nil
		}

		// Call the function under test
		err := createRSSFeed(context.Background(), w, req, collectionID, accessToken, mockSearchClient, validatedParams, template)

		Convey("it should not return an error", func() {
			So(err, ShouldBeNil)
		})

		Convey("it should set the Content-Type header to 'application/rss+xml'", func() {
			contentType := w.Header().Get("Content-Type")
			So(contentType, ShouldEqual, "application/rss+xml")
		})
	})

	Convey("when Search returns an error", t, func() {
		// Define expected behavior for the mock SearchClient
		// Create a mock response
		mockSearchClient.GetSearchFunc = func(ctx context.Context, options searchSDK.Options) (*searchModels.SearchResponse, apiError.Error) {
			return nil, apiError.StatusError{Code: 500}
		}

		// Call the function under test
		err := createRSSFeed(context.Background(), w, req, collectionID, accessToken, mockSearchClient, validatedParams, "template")

		Convey("it should return an error", func() {
			So(err, ShouldNotBeNil)
		})
	})
}

// For most handlers, search calls are done in parallel so assessing their
// mocks by order is challenging. This takes two parallel calls and assesses them to
// see which is the category call.
func sortSearchCalls(searchCall1 struct {
	Ctx     context.Context
	Options searchSDK.Options
}, searchCall2 struct {
	Ctx     context.Context
	Options searchSDK.Options
}, filter string,
) (categorySearchCall struct {
	Ctx     context.Context
	Options searchSDK.Options
}, querySearchCall struct {
	Ctx     context.Context
	Options searchSDK.Options
}) {
	if searchCall1.Options.Query.Has(filter) {
		return searchCall2, searchCall1
	}

	return searchCall1, searchCall2
}

func TestUnitReadDataAggregationWithTopicsRSSSuccess(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	mockSearchResponse, err := mapper.GetMockSearchResponse()
	if err != nil {
		t.Errorf("failed to retrieve mock search response for unit tests, failing early: %v", err)
	}

	mockHomepageContent, err := mapper.GetMockHomepageContent()
	if err != nil {
		t.Errorf("failed to retrieve mock homepage content for unit tests, failing early: %v", err)
	}

	Convey("Given a valid request for a subtopic filtered page and a set of mocked services", t, func() {
		testTopic := topicModels.Topic{
			ID:    "6734",
			Slug:  "economy",
			Title: "Economy",
		}

		testSubtopic := topicModels.Topic{
			ID:    "8268",
			Slug:  "governmentpublicsectorandtaxes",
			Title: "Government Public Sector and Taxes",
		}

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", fmt.Sprintf("/%s/%s/publications?rss", testTopic.Slug, testSubtopic.Slug), http.NoBody)
		req = mux.SetURLVars(req, map[string]string{"topicsPath": testTopic.Slug + "/" + testSubtopic.Slug})

		cfg, err := config.Get()
		So(err, ShouldBeNil)

		mockedRendererClient := &RenderClientMock{
			BuildPageFunc: func(w io.Writer, pageModel interface{}, templateName string) {},
			NewBasePageModelFunc: func() coreModel.Page {
				return coreModel.Page{}
			},
		}

		mockedSearchClient := &SearchClientMock{
			GetSearchFunc: func(ctx context.Context, options searchSDK.Options) (*searchModels.SearchResponse, apiError.Error) {
				return mockSearchResponse, nil
			},
		}

		mockedZebedeeClient := &ZebedeeClientMock{
			GetHomepageContentFunc: func(ctx context.Context, userAuthToken, collectionID, lang, path string) (zebedeeC.HomepageContent, error) {
				return mockHomepageContent, nil
			},
		}

		mockCacheList, err := cache.GetMockCacheList(ctx, englishLang)
		So(err, ShouldBeNil)

		Convey("When readDataAggregationWithTopics is called", func() {
			readDataAggregationWithTopics(w, req, cfg, mockedZebedeeClient, mockedRendererClient, mockedSearchClient, accessToken, collectionID, englishLang, *mockCacheList, "publications")

			Convey("Then a 200 OK status should be returned", func() {
				So(w.Code, ShouldEqual, http.StatusOK)
				So(w.Header().Get("Content-Type"), ShouldEqual, "application/rss+xml")
				reqBody, err := io.ReadAll(w.Body)
				if err != nil {
					fmt.Fprintf(w, "Kindly enter data ")
				}
				newBody := strings.Split(strings.ReplaceAll(string(reqBody), "\r\n", "\n"), "\n")
				So(newBody[0], ShouldContainSubstring, "<?xml version")
			})
		})
	})
}

func TestUnitReadDataAggregationWithTopicsRSSFailure(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	mockHomepageContent, err := mapper.GetMockHomepageContent()
	if err != nil {
		t.Errorf("failed to retrieve mock homepage content for unit tests, failing early: %v", err)
	}

	Convey("Given a valid request for a subtopic filtered page and a set of mocked services", t, func() {
		testTopic := topicModels.Topic{
			ID:    "6734",
			Slug:  "economy",
			Title: "Economy",
		}

		testSubtopic := topicModels.Topic{
			ID:    "8268",
			Slug:  "governmentpublicsectorandtaxes",
			Title: "Government Public Sector and Taxes",
		}

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", fmt.Sprintf("/%s/%s/publications?rss", testTopic.Slug, testSubtopic.Slug), http.NoBody)
		req = mux.SetURLVars(req, map[string]string{"topicsPath": testTopic.Slug + "/" + testSubtopic.Slug})

		cfg, err := config.Get()
		So(err, ShouldBeNil)

		mockedRendererClient := &RenderClientMock{
			BuildPageFunc: func(w io.Writer, pageModel interface{}, templateName string) {},
			NewBasePageModelFunc: func() coreModel.Page {
				return coreModel.Page{}
			},
		}

		mockSearchClient := &SearchClientMock{}
		mockSearchClient.GetSearchFunc = func(ctx context.Context, options searchSDK.Options) (*searchModels.SearchResponse, apiError.Error) {
			return nil, apiError.StatusError{Code: 500}
		}

		mockedZebedeeClient := &ZebedeeClientMock{
			GetHomepageContentFunc: func(ctx context.Context, userAuthToken, collectionID, lang, path string) (zebedeeC.HomepageContent, error) {
				return mockHomepageContent, nil
			},
		}

		mockCacheList, err := cache.GetMockCacheList(ctx, englishLang)
		So(err, ShouldBeNil)

		Convey("When readDataAggregationWithTopics is called", func() {
			readDataAggregationWithTopics(w, req, cfg, mockedZebedeeClient, mockedRendererClient, mockSearchClient, accessToken, collectionID, englishLang, *mockCacheList, "publications")

			Convey("Then a 200 OK status should be returned", func() {
				So(w.Code, ShouldEqual, http.StatusInternalServerError)
				So(w.Header().Get("Content-Type"), ShouldBeEmpty)

				reqBody, err := io.ReadAll(w.Body)
				if err != nil {
					fmt.Fprintf(w, "Kindly enter data ")
				}
				newBody := strings.Split(strings.ReplaceAll(string(reqBody), "\r\n", "\n"), "\n")
				So(newBody[0], ShouldBeEmpty)
			})
		})
	})
}

func TestValidateTopicHierarchy(t *testing.T) {
	ctx := context.Background()

	// Set up mock cache list
	cacheList, err := cache.GetMockCacheList(ctx, englishLang)
	if err != nil {
		t.Fatalf("Failed to get mock cache list: %v", err)
	}

	Convey("ValidateTopicHierarchy", t, func() {
		Convey("should return error when there are no segments to validate", func() {
			segments := []string{}
			_, err := ValidateTopicHierarchy(ctx, segments, *cacheList)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "no segments to validate")
		})

		Convey("should return the topic when a valid single segment root topic is provided", func() {
			segments := []string{"economy"}
			topic, err := ValidateTopicHierarchy(ctx, segments, *cacheList)
			So(err, ShouldBeNil)
			So(topic, ShouldNotBeNil)
			So(topic.ID, ShouldEqual, "6734")
		})

		Convey("should return the last topic in a valid topic hierarchy - 2nd level", func() {
			segments := []string{"economy", "governmentpublicsectorandtaxes"}
			topic, err := ValidateTopicHierarchy(ctx, segments, *cacheList)
			So(err, ShouldBeNil)
			So(topic, ShouldNotBeNil)
			So(topic.ID, ShouldEqual, "8268")
		})

		Convey("should return the last topic in a valid topic hierarchy - 3rd level", func() {
			segments := []string{"economy", "governmentpublicsectorandtaxes", "publicsectorfinance"}
			topic, err := ValidateTopicHierarchy(ctx, segments, *cacheList)
			So(err, ShouldBeNil)
			So(topic, ShouldNotBeNil)
			So(topic.ID, ShouldEqual, "3687")
		})

		Convey("should return error for an invalid topic hierarchy", func() {
			segments := []string{"environmentalaccounts", "publicsectorfinance"}
			_, err := ValidateTopicHierarchy(ctx, segments, *cacheList)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "invalid topic hierarchy at segment: publicsectorfinance")
		})

		Convey("should return error when a nonexistent topic is provided", func() {
			segments := []string{"nonexistent"}
			_, err := ValidateTopicHierarchy(ctx, segments, *cacheList)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "topic not found: nonexistent")
		})

		Convey("should return the last topic in another valid topic branch", func() {
			segments := []string{"economy", "environmentalaccounts"}
			topic, err := ValidateTopicHierarchy(ctx, segments, *cacheList)
			So(err, ShouldBeNil)
			So(topic, ShouldNotBeNil)
			So(topic.ID, ShouldEqual, "1834")
		})
	})
}
