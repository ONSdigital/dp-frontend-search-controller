package handlers

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	searchC "github.com/ONSdigital/dp-api-clients-go/v2/site-search"
	zebedeeC "github.com/ONSdigital/dp-api-clients-go/v2/zebedee"
	errs "github.com/ONSdigital/dp-frontend-search-controller/apperrors"
	"github.com/ONSdigital/dp-frontend-search-controller/config"
	"github.com/ONSdigital/dp-frontend-search-controller/data"
	"github.com/ONSdigital/dp-frontend-search-controller/mapper"
	coreModel "github.com/ONSdigital/dp-renderer/model"
	"github.com/gorilla/mux"
	. "github.com/smartystreets/goconvey/convey"
)

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

var accessToken string
var collectionID string
var lang string

func TestUnitReadHandlerSuccess(t *testing.T) {
	t.Parallel()

	mockSearchResponse, err := mapper.GetMockSearchResponse()
	if err != nil {
		t.Errorf("failed to retrieve mock search response for unit tests, failing early: %v", err)
	}
	mockDepartmentResponse, err := mapper.GetMockDepartmentResponse()
	if err != nil {
		t.Errorf("failed to retrieve mock department response for unit tests, failing early: %v", err)
	}

	mockHomepageContent, err := mapper.GetMockHomepageContent()
	if err != nil {
		t.Errorf("failed to retrieve mock homepage content for unit tests, failing early: %v", err)
	}

	Convey("Given a valid request", t, func() {
		req := httptest.NewRequest("GET", "/search?q=housing", nil)
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
			GetSearchFunc: func(ctx context.Context, userAuthToken, serviceAuthToken, collectionID string, query url.Values) (searchC.Response, error) {
				return mockSearchResponse, nil
			},
			GetDepartmentsFunc: func(ctx context.Context, userAuthToken, serviceAuthToken, collectionID string, query url.Values) (searchC.Department, error) {
				return mockDepartmentResponse, nil
			},
		}

		mockedZebedeeClient := &ZebedeeClientMock{
			GetHomepageContentFunc: func(ctx context.Context, userAuthToken, collectionID, lang, path string) (zebedeeC.HomepageContent, error) {
				fmt.Printf("%+v\n", mockHomepageContent)
				return mockHomepageContent, nil
			}}

		Convey("When Read is called", func() {
			w := doTestRequest("/search", req, Read(cfg, mockedZebedeeClient, mockedRendererClient, mockedSearchClient), nil)

			Convey("Then a 200 OK status should be returned", func() {
				So(w.Code, ShouldEqual, http.StatusOK)

				So(len(mockedRendererClient.BuildPageCalls()), ShouldEqual, 1)
				So(len(mockedSearchClient.GetSearchCalls()), ShouldEqual, 2)
				So(len(mockedZebedeeClient.GetHomepageContentCalls()), ShouldEqual, 1)
			})
		})
	})
}

func TestUnitReadSuccess(t *testing.T) {
	t.Parallel()

	mockSearchResponse, err := mapper.GetMockSearchResponse()
	if err != nil {
		t.Errorf("failed to retrieve mock search response for unit tests, failing early: %v", err)
	}

	mockDepartmentResponse, err := mapper.GetMockDepartmentResponse()
	if err != nil {
		t.Errorf("failed to retrieve mock department response for unit tests, failing early: %v", err)
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
			GetDepartmentsFunc: func(ctx context.Context, userAuthToken, serviceAuthToken, collectionID string, query url.Values) (searchC.Department, error) {
				return mockDepartmentResponse, nil
			},
		}

		mockedZebedeeClient := &ZebedeeClientMock{
			GetHomepageContentFunc: func(ctx context.Context, userAuthToken, collectionID, lang, path string) (zebedeeC.HomepageContent, error) {
				return mockHomepageContent, nil
			}}

		Convey("When read is called", func() {
			read(w, req, cfg, mockedZebedeeClient, mockedRendererClient, mockedSearchClient, accessToken, collectionID, lang)

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

	mockSearchResponse, err := mapper.GetMockSearchResponse()
	if err != nil {
		t.Errorf("failed to retrieve mock search response for unit tests, failing early: %v", err)
	}
	mockDepartmentResponse, err := mapper.GetMockDepartmentResponse()
	if err != nil {
		t.Errorf("failed to retrieve mock department response for unit tests, failing early: %v", err)
	}

	mockHomepageContent, err := mapper.GetMockHomepageContent()
	fmt.Printf("%+v\n", mockDepartmentResponse)
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
			GetDepartmentsFunc: func(ctx context.Context, userAuthToken, serviceAuthToken, collectionID string, query url.Values) (searchC.Department, error) {
				return mockDepartmentResponse, nil
			},
		}

		mockedZebedeeClient := &ZebedeeClientMock{
			GetHomepageContentFunc: func(ctx context.Context, userAuthToken, collectionID, lang, path string) (zebedeeC.HomepageContent, error) {
				return mockHomepageContent, nil
			}}

		Convey("When read is called", func() {
			read(w, req, cfg, mockedZebedeeClient, mockedRendererClient, mockedSearchClient, accessToken, collectionID, lang)

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
			GetDepartmentsFunc: func(ctx context.Context, userAuthToken, serviceAuthToken, collectionID string, query url.Values) (searchC.Department, error) {
				return mockDepartmentResponse, nil
			},
		}

		mockedZebedeeClient := &ZebedeeClientMock{
			GetHomepageContentFunc: func(ctx context.Context, userAuthToken, collectionID, lang, path string) (zebedeeC.HomepageContent, error) {
				return mockHomepageContent, nil
			}}

		Convey("When read is called", func() {
			read(w, req, cfg, mockedZebedeeClient, mockedRendererClient, mockedSearchClient, accessToken, collectionID, lang)

			Convey("Then a 500 internal server error status should be returned", func() {
				So(w.Code, ShouldEqual, http.StatusInternalServerError)

				So(len(mockedRendererClient.BuildPageCalls()), ShouldEqual, 0)
				So(len(mockedSearchClient.GetSearchCalls()), ShouldEqual, 1)
				So(len(mockedSearchClient.GetDepartmentsCalls()), ShouldEqual, 1)
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
			GetDepartmentsFunc: func(ctx context.Context, userAuthToken, serviceAuthToken, collectionID string, query url.Values) (searchC.Department, error) {
				return mockDepartmentResponse, nil
			},
		}

		mockedZebedeeClient := &ZebedeeClientMock{
			GetHomepageContentFunc: func(ctx context.Context, userAuthToken, collectionID, lang, path string) (zebedeeC.HomepageContent, error) {
				return mockHomepageContent, nil
			}}

		Convey("When read is called", func() {
			read(w, req, cfg, mockedZebedeeClient, mockedRendererClient, mockedSearchClient, accessToken, collectionID, lang)

			Convey("Then a 400 bad request status should be returned", func() {
				So(w.Code, ShouldEqual, http.StatusBadRequest)

				So(len(mockedRendererClient.BuildPageCalls()), ShouldEqual, 0)
				So(len(mockedSearchClient.GetSearchCalls()), ShouldEqual, 1)
				So(len(mockedSearchClient.GetDepartmentsCalls()), ShouldEqual, 1)
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
			categories, err := getCategoriesTypesCount(ctx, accessToken, collectionID, apiQuery, mockedSearchClient)

			Convey("Then return all categories and types with its count", func() {
				So(categories[0].Count, ShouldEqual, 1)
				So(categories[0].ContentTypes[1].Count, ShouldEqual, 1)
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
			categories, err := getCategoriesTypesCount(ctx, accessToken, collectionID, apiQuery, mockedSearchClient)

			Convey("Then return an error", func() {
				So(err, ShouldNotBeNil)
			})

			Convey("And return nil categories", func() {
				So(categories, ShouldBeNil)
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
