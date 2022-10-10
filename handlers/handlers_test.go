package handlers

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync"
	"testing"

	searchC "github.com/ONSdigital/dp-api-clients-go/v2/site-search"
	zebedeeC "github.com/ONSdigital/dp-api-clients-go/v2/zebedee"
	errs "github.com/ONSdigital/dp-frontend-search-controller/apperrors"
	"github.com/ONSdigital/dp-frontend-search-controller/cache"
	"github.com/ONSdigital/dp-frontend-search-controller/config"
	"github.com/ONSdigital/dp-frontend-search-controller/data"
	"github.com/ONSdigital/dp-frontend-search-controller/mapper"
	coreModel "github.com/ONSdigital/dp-renderer/model"
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
		req := httptest.NewRequest("GET", "/search?q=housing&filter=bulletin&topics=1234", nil)
		req.Header.Set("X-Florence-Token", "testuser")

		cfg, err := config.Get()
		So(err, ShouldBeNil)

		mockedRendererClient := &RenderClientMock{
			BuildPageFunc: func(w io.Writer, pageModel interface{}, templateName string) {},
			NewBasePageModelFunc: func() coreModel.Page {
				return coreModel.Page{}
			},
		}

		expectedQueries := map[int]url.Values{
			0: {
				"limit":  []string{"10"},
				"offset": []string{"0"},
				"q":      []string{"housing"},
				"sort":   []string{"relevance"},
			},
			1: {
				"limit":  []string{"10"},
				"offset": []string{"0"},
				"q":      []string{"housing"},
				"sort":   []string{"relevance"},
			},
		}

		contentTypes := "bulletin"
		topics := "1234"

		mockedSearchClient := GetSearchClient(mockSearchResponse, expectedQueries, contentTypes, topics)

		mockedZebedeeClient := &ZebedeeClientMock{
			GetHomepageContentFunc: func(ctx context.Context, userAuthToken, collectionID, lang, path string) (zebedeeC.HomepageContent, error) {
				return mockHomepageContent, nil
			}}

		mockCacheList, err := cache.GetMockCacheList(ctx, englishLang)
		So(err, ShouldBeNil)

		Convey("When Read is called", func() {
			w := doTestRequest("/search", req, Read(cfg, mockedZebedeeClient, mockedRendererClient, mockedSearchClient, *mockCacheList), nil)

			Convey("Then a 200 OK status should be returned", func() {
				So(w.Code, ShouldEqual, http.StatusOK)

				So(len(mockedRendererClient.BuildPageCalls()), ShouldEqual, 1)
				So(len(mockedSearchClient.GetSearchCalls()), ShouldEqual, 2)
				So(mockedSearchClient.calls.GetSearch[0].Query, ShouldResemble, expectedQueries[0])
				So(mockedSearchClient.calls.GetSearch[1].Query, ShouldResemble, expectedQueries[1])
				So(len(mockedZebedeeClient.GetHomepageContentCalls()), ShouldEqual, 1)
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
		req := httptest.NewRequest("GET", "/search?q=housing", nil)

		cfg, err := config.Get()
		So(err, ShouldBeNil)

		mockedRendererClient := &RenderClientMock{
			BuildPageFunc: func(w io.Writer, pageModel interface{}, templateName string) {},
			NewBasePageModelFunc: func() coreModel.Page {
				return coreModel.Page{}
			},
		}

		mockedSearchClient := &SearchClientMock{
			GetSearchFunc: func(ctx context.Context, userAuthToken, serviceAuthToken, collectionID string, query url.Values) (searchC.Response, error) {
				return mockSearchResponse, nil
			},
		}

		mockedZebedeeClient := &ZebedeeClientMock{
			GetHomepageContentFunc: func(ctx context.Context, userAuthToken, collectionID, lang, path string) (zebedeeC.HomepageContent, error) {
				return mockHomepageContent, nil
			}}

		mockCacheList, err := cache.GetMockCacheList(ctx, englishLang)
		So(err, ShouldBeNil)

		Convey("When read is called", func() {
			read(w, req, cfg, mockedZebedeeClient, mockedRendererClient, mockedSearchClient, accessToken, collectionID, englishLang, *mockCacheList)

			Convey("Then a 200 OK status should be returned", func() {
				So(w.Code, ShouldEqual, http.StatusOK)

				So(len(mockedRendererClient.BuildPageCalls()), ShouldEqual, 1)
				So(len(mockedSearchClient.GetSearchCalls()), ShouldEqual, 2)
				So(mockedSearchClient.GetSearchCalls()[0].UserAuthToken, ShouldEqual, accessToken)
				So(mockedSearchClient.GetSearchCalls()[0].CollectionID, ShouldEqual, collectionID)
				So(len(mockedZebedeeClient.GetHomepageContentCalls()), ShouldEqual, 1)
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
		req := httptest.NewRequest("GET", "/search?q=housing&page=1000000", nil)

		cfg, err := config.Get()
		So(err, ShouldBeNil)

		mockedRendererClient := &RenderClientMock{
			BuildPageFunc: func(w io.Writer, pageModel interface{}, templateName string) {},
			NewBasePageModelFunc: func() coreModel.Page {
				return coreModel.Page{}
			},
		}

		mockedSearchClient := &SearchClientMock{
			GetSearchFunc: func(ctx context.Context, userAuthToken, serviceAuthToken, collectionID string, query url.Values) (searchC.Response, error) {
				return mockSearchResponse, nil
			},
		}

		mockedZebedeeClient := &ZebedeeClientMock{
			GetHomepageContentFunc: func(ctx context.Context, userAuthToken, collectionID, lang, path string) (zebedeeC.HomepageContent, error) {
				return mockHomepageContent, nil
			}}

		mockCacheList, err := cache.GetMockCacheList(ctx, englishLang)
		So(err, ShouldBeNil)

		Convey("When read is called", func() {
			read(w, req, cfg, mockedZebedeeClient, mockedRendererClient, mockedSearchClient, accessToken, collectionID, englishLang, *mockCacheList)

			Convey("Then a 400 bad request status should be returned", func() {
				So(w.Code, ShouldEqual, http.StatusBadRequest)

				So(len(mockedRendererClient.BuildPageCalls()), ShouldEqual, 0)
				So(len(mockedSearchClient.GetSearchCalls()), ShouldEqual, 0)
				So(len(mockedZebedeeClient.GetHomepageContentCalls()), ShouldEqual, 0)
			})
		})
	})

	Convey("Given an error from failing to get search response from search client", t, func() {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/search?q=housing", nil)

		cfg, err := config.Get()
		So(err, ShouldBeNil)

		mockedRendererClient := &RenderClientMock{
			BuildPageFunc: func(w io.Writer, pageModel interface{}, templateName string) {},
			NewBasePageModelFunc: func() coreModel.Page {
				return coreModel.Page{}
			},
		}

		mockedSearchClient := &SearchClientMock{
			GetSearchFunc: func(ctx context.Context, userAuthToken, serviceAuthToken, collectionID string, query url.Values) (searchC.Response, error) {
				return searchC.Response{}, errs.ErrInternalServer
			},
		}

		mockedZebedeeClient := &ZebedeeClientMock{
			GetHomepageContentFunc: func(ctx context.Context, userAuthToken, collectionID, lang, path string) (zebedeeC.HomepageContent, error) {
				return mockHomepageContent, nil
			}}

		mockCacheList, err := cache.GetMockCacheList(ctx, englishLang)
		So(err, ShouldBeNil)

		Convey("When read is called", func() {
			read(w, req, cfg, mockedZebedeeClient, mockedRendererClient, mockedSearchClient, accessToken, collectionID, englishLang, *mockCacheList)

			Convey("Then a 500 internal server error status should be returned", func() {
				So(w.Code, ShouldEqual, http.StatusInternalServerError)

				So(len(mockedRendererClient.BuildPageCalls()), ShouldEqual, 0)
				So(len(mockedSearchClient.GetSearchCalls()), ShouldEqual, 2)
				So(len(mockedZebedeeClient.GetHomepageContentCalls()), ShouldEqual, 1)
			})
		})
	})

	Convey("Given an error as current page exceeds total pages", t, func() {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/search?q=housing&page=2", nil)

		cfg, err := config.Get()
		So(err, ShouldBeNil)

		mockedRendererClient := &RenderClientMock{
			BuildPageFunc: func(w io.Writer, pageModel interface{}, templateName string) {},
			NewBasePageModelFunc: func() coreModel.Page {
				return coreModel.Page{}
			},
		}

		mockedSearchClient := &SearchClientMock{
			GetSearchFunc: func(ctx context.Context, userAuthToken, serviceAuthToken, collectionID string, query url.Values) (searchC.Response, error) {
				return mockSearchResponse, nil
			},
		}

		mockedZebedeeClient := &ZebedeeClientMock{
			GetHomepageContentFunc: func(ctx context.Context, userAuthToken, collectionID, lang, path string) (zebedeeC.HomepageContent, error) {
				return mockHomepageContent, nil
			}}

		mockCacheList, err := cache.GetMockCacheList(ctx, englishLang)
		So(err, ShouldBeNil)

		Convey("When read is called", func() {
			read(w, req, cfg, mockedZebedeeClient, mockedRendererClient, mockedSearchClient, accessToken, collectionID, englishLang, *mockCacheList)

			Convey("Then a 400 bad request status should be returned", func() {
				So(w.Code, ShouldEqual, http.StatusBadRequest)

				So(len(mockedRendererClient.BuildPageCalls()), ShouldEqual, 0)
				So(len(mockedSearchClient.GetSearchCalls()), ShouldEqual, 2)
				So(len(mockedZebedeeClient.GetHomepageContentCalls()), ShouldEqual, 1)
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
					So(err, ShouldResemble, errs.ErrPageExceedsTotalPages)
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
			GetSearchFunc: func(ctx context.Context, userAuthToken, serviceAuthToken, collectionID string, query url.Values) (searchC.Response, error) {
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

				So(len(mockedSearchClient.GetSearchCalls()), ShouldEqual, 1)
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
			GetSearchFunc: func(ctx context.Context, userAuthToken, serviceAuthToken, collectionID string, query url.Values) (searchC.Response, error) {
				return searchC.Response{}, errs.ErrInternalServer
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
				So(len(mockedSearchClient.GetSearchCalls()), ShouldEqual, 1)
				So(mockedSearchClient.GetSearchCalls()[0].UserAuthToken, ShouldEqual, accessToken)
				So(mockedSearchClient.GetSearchCalls()[0].CollectionID, ShouldEqual, collectionID)
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
		mockCountSearchResponse := searchC.Response{
			Count: 1,
			ContentTypes: []searchC.FilterCount{
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
				setCountToCategories(ctx, mockCountSearchResponse, categories)

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
		req := httptest.NewRequest("GET", "/search?q=housing&page=1000000", nil)

		err := errs.ErrInternalServer

		Convey("When setStatusCode is called", func() {
			setStatusCode(w, req, err)

			Convey("Then send a HTTP response header with 500 internal server error", func() {
				So(w.Code, ShouldEqual, http.StatusInternalServerError)
			})
		})
	})

	Convey("Given an client error", t, func() {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/search?q=housing", nil)

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
		req := httptest.NewRequest("GET", "/search?q=housing&page=1000000", nil)

		err := errs.ErrInvalidPage

		Convey("When setStatusCode is called", func() {
			setStatusCode(w, req, err)

			Convey("Then send a HTTP response header with 400 bad request status", func() {
				So(w.Code, ShouldEqual, http.StatusBadRequest)
			})
		})
	})
}

type container struct {
	counter int
	mu      sync.Mutex
}

func GetSearchClient(mockSearchResponse searchC.Response, expectedQueries map[int]url.Values, contentTypes, topics string) *SearchClientMock {
	var searchRequestContainer container

	mockedSearchClient := &SearchClientMock{
		GetSearchFunc: func(ctx context.Context, userAuthToken, serviceAuthToken, collectionID string, query url.Values) (searchC.Response, error) {
			searchRequestContainer.mu.Lock()
			defer searchRequestContainer.mu.Unlock()

			if query.Get("content_type") != "" {
				expectedQueries[searchRequestContainer.counter].Add("content_type", contentTypes)
			}

			if query.Get("topics") != "" {
				expectedQueries[searchRequestContainer.counter].Add("topics", topics)
			}

			searchRequestContainer.counter++

			return mockSearchResponse, nil
		},
	}

	return mockedSearchClient
}
