package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	searchC "github.com/ONSdigital/dp-api-clients-go/site-search"
	errs "github.com/ONSdigital/dp-frontend-search-controller/apperrors"
	"github.com/ONSdigital/dp-frontend-search-controller/config"
	"github.com/ONSdigital/dp-frontend-search-controller/data"
	"github.com/ONSdigital/dp-frontend-search-controller/mapper"
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

	Convey("Given a valid request", t, func() {
		req := httptest.NewRequest("GET", "/search?q=housing", nil)

		cfg, err := config.Get()
		So(err, ShouldBeNil)

		mockedRendererClient := &RenderClientMock{
			DoFunc: func(in1 string, in2 []byte) ([]byte, error) {
				return []byte(`<html><body><h1>Some HTML from renderer!</h1></body></html>`), nil
			},
		}

		mockedSearchClient := &SearchClientMock{
			GetSearchFunc: func(ctx context.Context, query url.Values) (searchC.Response, error) {
				return mockSearchResponse, nil
			},
			GetDepartmentsFunc: func(ctx context.Context, query url.Values) (searchC.Department, error) {
				return mockDepartmentResponse, nil
			},
		}

		Convey("When Read is called", func() {
			w := doTestRequest("/search", req, Read(cfg, mockedRendererClient, mockedSearchClient), nil)

			Convey("Then a 200 OK status should be returned", func() {
				So(w.Code, ShouldEqual, http.StatusOK)

				So(len(mockedRendererClient.DoCalls()), ShouldEqual, 1)
				So(len(mockedSearchClient.GetSearchCalls()), ShouldEqual, 2)
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

	Convey("Given a valid request", t, func() {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/search?q=housing", nil)

		cfg, err := config.Get()
		So(err, ShouldBeNil)

		mockedRendererClient := &RenderClientMock{
			DoFunc: func(in1 string, in2 []byte) ([]byte, error) {
				return []byte(`<html><body><h1>Some HTML from renderer!</h1></body></html>`), nil
			},
		}

		mockedSearchClient := &SearchClientMock{
			GetSearchFunc: func(ctx context.Context, query url.Values) (searchC.Response, error) {
				return mockSearchResponse, nil
			},
			GetDepartmentsFunc: func(ctx context.Context, query url.Values) (searchC.Department, error) {
				return mockDepartmentResponse, nil
			},
		}

		Convey("When read is called", func() {
			read(w, req, cfg, mockedRendererClient, mockedSearchClient, accessToken, collectionID, lang)

			Convey("Then a 200 OK status should be returned", func() {
				So(w.Code, ShouldEqual, http.StatusOK)

				So(len(mockedRendererClient.DoCalls()), ShouldEqual, 1)
				So(len(mockedSearchClient.GetSearchCalls()), ShouldEqual, 2)
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

	Convey("Given an error from failing to review query", t, func() {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/search?q=housing&page=1000000", nil)

		cfg, err := config.Get()
		So(err, ShouldBeNil)

		mockedRendererClient := &RenderClientMock{
			DoFunc: func(in1 string, in2 []byte) ([]byte, error) {
				return []byte(`<html><body><h1>Some HTML from renderer!</h1></body></html>`), nil
			},
		}

		mockedSearchClient := &SearchClientMock{
			GetSearchFunc: func(ctx context.Context, query url.Values) (searchC.Response, error) {
				return mockSearchResponse, nil
			},
			GetDepartmentsFunc: func(ctx context.Context, query url.Values) (searchC.Department, error) {
				return mockDepartmentResponse, nil
			},
		}

		Convey("When read is called", func() {
			read(w, req, cfg, mockedRendererClient, mockedSearchClient, accessToken, collectionID, lang)

			Convey("Then a 400 bad request status should be returned", func() {
				So(w.Code, ShouldEqual, http.StatusBadRequest)

				So(len(mockedRendererClient.DoCalls()), ShouldEqual, 0)
				So(len(mockedSearchClient.GetSearchCalls()), ShouldEqual, 0)
			})
		})
	})

	Convey("Given an error from failing to get search response from search client", t, func() {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/search?q=housing", nil)

		cfg, err := config.Get()
		So(err, ShouldBeNil)

		mockedRendererClient := &RenderClientMock{
			DoFunc: func(in1 string, in2 []byte) ([]byte, error) {
				return []byte(`<html><body><h1>Some HTML from renderer!</h1></body></html>`), nil
			},
		}

		mockedSearchClient := &SearchClientMock{
			GetSearchFunc: func(ctx context.Context, query url.Values) (searchC.Response, error) {
				return searchC.Response{}, errs.ErrInternalServer
			},
			GetDepartmentsFunc: func(ctx context.Context, query url.Values) (searchC.Department, error) {
				return mockDepartmentResponse, nil
			},
		}

		Convey("When read is called", func() {
			read(w, req, cfg, mockedRendererClient, mockedSearchClient, accessToken, collectionID, lang)

			Convey("Then a 500 internal server error status should be returned", func() {
				So(w.Code, ShouldEqual, http.StatusInternalServerError)

				So(len(mockedRendererClient.DoCalls()), ShouldEqual, 0)
				So(len(mockedSearchClient.GetSearchCalls()), ShouldEqual, 1)
				So(len(mockedSearchClient.GetDepartmentsCalls()), ShouldEqual, 1)
			})
		})
	})

	Convey("Given an error as current page exceeds total pages", t, func() {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/search?q=housing&page=2", nil)

		cfg, err := config.Get()
		So(err, ShouldBeNil)

		mockedRendererClient := &RenderClientMock{
			DoFunc: func(in1 string, in2 []byte) ([]byte, error) {
				return []byte(`<html><body><h1>Some HTML from renderer!</h1></body></html>`), nil
			},
		}

		mockedSearchClient := &SearchClientMock{
			GetSearchFunc: func(ctx context.Context, query url.Values) (searchC.Response, error) {
				return mockSearchResponse, nil
			},
			GetDepartmentsFunc: func(ctx context.Context, query url.Values) (searchC.Department, error) {
				return mockDepartmentResponse, nil
			},
		}

		Convey("When read is called", func() {
			read(w, req, cfg, mockedRendererClient, mockedSearchClient, accessToken, collectionID, lang)

			Convey("Then a 400 bad request status should be returned", func() {
				So(w.Code, ShouldEqual, http.StatusBadRequest)

				So(len(mockedRendererClient.DoCalls()), ShouldEqual, 0)
				So(len(mockedSearchClient.GetSearchCalls()), ShouldEqual, 1)
				So(len(mockedSearchClient.GetDepartmentsCalls()), ShouldEqual, 1)
			})
		})
	})

	Convey("Given an error from failing to get search page from renderer", t, func() {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/search?q=housing", nil)

		cfg, err := config.Get()
		So(err, ShouldBeNil)

		mockedRendererClient := &RenderClientMock{
			DoFunc: func(in1 string, in2 []byte) ([]byte, error) {
				return []byte{}, errs.ErrInternalServer
			},
		}

		mockedSearchClient := &SearchClientMock{
			GetSearchFunc: func(ctx context.Context, query url.Values) (searchC.Response, error) {
				return mockSearchResponse, nil
			},
			GetDepartmentsFunc: func(ctx context.Context, query url.Values) (searchC.Department, error) {
				return mockDepartmentResponse, nil
			},
		}

		Convey("When read is called", func() {
			read(w, req, cfg, mockedRendererClient, mockedSearchClient, accessToken, collectionID, lang)

			Convey("Then a 500 internal server error status should be returned", func() {
				So(w.Code, ShouldEqual, http.StatusInternalServerError)

				So(len(mockedRendererClient.DoCalls()), ShouldEqual, 1)
				So(len(mockedSearchClient.GetSearchCalls()), ShouldEqual, 2)
				So(len(mockedSearchClient.GetDepartmentsCalls()), ShouldEqual, 1)
			})
		})
	})
}

func TestUnitValidateCurrentPageSuccess(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	Convey("Given number of search results is more than 0", t, func() {
		resultsCount := 10

		Convey("And current page doesn't exceed total pages", func() {
			validatedQueryParams := data.SearchURLParams{
				Limit:       10,
				CurrentPage: 1,
			}

			Convey("When validateCurrentPage is called", func() {
				err := validateCurrentPage(ctx, validatedQueryParams, resultsCount)

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
				err := validateCurrentPage(ctx, validatedQueryParams, resultsCount)

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

	Convey("Given current page exceeds total pages", t, func() {
		validatedQueryParams := data.SearchURLParams{
			Limit:       10,
			CurrentPage: 10000,
		}

		Convey("And number of search results is more than 0", func() {
			resultsCount := 20

			Convey("When validateCurrentPage is called", func() {
				err := validateCurrentPage(ctx, validatedQueryParams, resultsCount)

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
			GetSearchFunc: func(ctx context.Context, query url.Values) (searchC.Response, error) {
				return mockSearchResponse, nil
			},
		}

		Convey("When getCategoriesTypesCount is called", func() {
			categories, err := getCategoriesTypesCount(ctx, apiQuery, mockedSearchClient)

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
			GetSearchFunc: func(ctx context.Context, query url.Values) (searchC.Response, error) {
				return searchC.Response{}, errs.ErrInternalServer
			},
		}

		Convey("When getCategoriesTypesCount is called", func() {
			categories, err := getCategoriesTypesCount(ctx, apiQuery, mockedSearchClient)

			Convey("Then return an error", func() {
				So(err, ShouldNotBeNil)
			})

			Convey("And return nil categories", func() {
				So(categories, ShouldBeNil)

				So(len(mockedSearchClient.GetSearchCalls()), ShouldEqual, 1)
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
			ContentTypes: []searchC.ContentType{
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

func TestUnitGetSearchPageSuccess(t *testing.T) {
	t.Parallel()

	Convey("Given valid search data such as query parameters, categories and response", t, func() {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/search?q=housing", nil)

		cfg, err := config.Get()
		So(err, ShouldBeNil)

		mockedRendererClient := &RenderClientMock{
			DoFunc: func(in1 string, in2 []byte) ([]byte, error) {
				return []byte(`<html><body><h1>Some HTML from renderer!</h1></body></html>`), nil
			},
		}

		validatedQueryParams := data.SearchURLParams{
			Query: "housing",
			Sort: data.Sort{
				Query:           "relevance",
				LocaliseKeyName: "Relevance",
			},
			Limit:       10,
			CurrentPage: 1,
		}

		categories := data.GetCategories()
		categories[0].Count = 1
		categories[0].ContentTypes[1].Count = 1

		mockCountSearchResponse, err := mapper.GetMockSearchResponse()
		mockDeptResponse, err := mapper.GetMockDepartmentResponse()
		So(err, ShouldBeNil)

		Convey("When getSearchPage is called", func() {
			err := getSearchPage(w, req, cfg, mockedRendererClient, validatedQueryParams, categories, mockCountSearchResponse, mockDeptResponse, lang)

			Convey("Then return no error and successfully get search page", func() {
				So(err, ShouldBeNil)

				So(len(mockedRendererClient.DoCalls()), ShouldEqual, 1)
			})
		})
	})
}

func TestUnitGetSearchPageFailure(t *testing.T) {
	t.Parallel()

	Convey("Given an error from failing to get template from renderer", t, func() {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/search?q=housing", nil)

		cfg, err := config.Get()
		So(err, ShouldBeNil)

		mockedRendererClient := &RenderClientMock{
			DoFunc: func(in1 string, in2 []byte) ([]byte, error) {
				return []byte{}, errs.ErrInternalServer
			},
		}

		validatedQueryParams := data.SearchURLParams{
			Query: "housing",
			Sort: data.Sort{
				Query:           "relevance",
				LocaliseKeyName: "Relevance",
			},
			Limit:       10,
			CurrentPage: 1,
		}

		categories := data.GetCategories()
		categories[0].Count = 1
		categories[0].ContentTypes[1].Count = 1

		mockCountSearchResponse, err := mapper.GetMockSearchResponse()
		mockDeptResponse, err := mapper.GetMockDepartmentResponse()
		So(err, ShouldBeNil)

		Convey("When getSearchPage is called", func() {
			err := getSearchPage(w, req, cfg, mockedRendererClient, validatedQueryParams, categories, mockCountSearchResponse, mockDeptResponse, lang)

			Convey("Then return error", func() {
				So(err, ShouldNotBeNil)

				So(len(mockedRendererClient.DoCalls()), ShouldEqual, 1)
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
