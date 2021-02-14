package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"

	"testing"

	searchC "github.com/ONSdigital/dp-api-clients-go/site-search"
	"github.com/ONSdigital/dp-frontend-search-controller/data"
	"github.com/gorilla/mux"
	. "github.com/smartystreets/goconvey/convey"
)

var respC searchC.Response

type testCliError struct{}

func (e *testCliError) Error() string { return "client error" }
func (e *testCliError) Code() int     { return http.StatusNotFound }

func createMockCategories() []data.Category {
	return []data.Category{data.Publication, data.Data, data.Other}
}

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

func TestUnitHandlers(t *testing.T) {
	t.Parallel()

	Convey("When setStatusCode called", t, func() {

		Convey("handles 404 response from client", func() {
			req := httptest.NewRequest("GET", "http://localhost:", nil)
			w := httptest.NewRecorder()
			err := &testCliError{}

			setStatusCode(req, w, err)

			So(w.Code, ShouldEqual, http.StatusNotFound)
		})

		Convey("handles internal server error", func() {
			req := httptest.NewRequest("GET", "http://localhost:", nil)
			w := httptest.NewRecorder()
			err := errors.New("internal server error")

			setStatusCode(req, w, err)

			So(w.Code, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("handles bad request error", func() {
			req := httptest.NewRequest("GET", "/search?q=housing&filter=INVALID", nil)
			w := httptest.NewRecorder()
			err := errors.New("invalid filter type given")

			setStatusCode(req, w, err)

			So(w.Code, ShouldEqual, http.StatusBadRequest)
		})
	})

	Convey("When getSearchPage called", t, func() {
		req := httptest.NewRequest("GET", "/search?q=housing&limit=1&offset=10&filter=article,filter2&sortBy=relevance", nil)
		url := req.URL
		w := httptest.NewRecorder()
		mockedRenderClient := &RenderClientMock{
			DoFunc: func(in1 string, in2 []byte) ([]byte, error) {
				return []byte{}, nil
			},
		}
		categories := data.Categories

		Convey("convert mock response to client model", func() {
			sampleResponse, err := ioutil.ReadFile("../mapper/test_data/mock_response.json")
			So(err, ShouldBeNil)

			err = json.Unmarshal(sampleResponse, &respC)
			So(err, ShouldBeNil)

			Convey("successfully gets the search page", func() {
				err := getSearchPage(w, req, mockedRenderClient, url, respC, categories)
				So(err, ShouldBeNil)
				So(len(mockedRenderClient.DoCalls()), ShouldEqual, 1)
			})

			Convey("returns err as unable to marshal search response", func() {
				defaultM := marshal
				marshal = func(v interface{}) ([]byte, error) {
					return []byte{}, errors.New("internal server error")
				}
				err := getSearchPage(w, req, mockedRenderClient, url, respC, categories)
				So(err, ShouldNotBeNil)
				So(len(mockedRenderClient.DoCalls()), ShouldEqual, 0)
				marshal = defaultM
			})

			Convey("returns err as getting template from renderer fails", func() {
				mockedRenderClient := &RenderClientMock{
					DoFunc: func(in1 string, in2 []byte) ([]byte, error) {
						return []byte{}, errors.New("internal server error")
					},
				}
				err := getSearchPage(w, req, mockedRenderClient, url, respC, categories)
				So(err, ShouldNotBeNil)
				So(len(mockedRenderClient.DoCalls()), ShouldEqual, 1)
			})

			Convey("returns err as unable to write of search template", func() {
				defaultW := writeResponse
				writeResponse = func(w http.ResponseWriter, templateHTML []byte) (int, error) {
					return 0, errors.New("internal server error")
				}
				err = getSearchPage(w, req, mockedRenderClient, url, respC, categories)
				So(err, ShouldNotBeNil)
				So(len(mockedRenderClient.DoCalls()), ShouldEqual, 1)
				writeResponse = defaultW
			})

		})
	})

	Convey("When getCategoriesTypesCount is called", t, func() {
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
			categories, err := getCategoriesTypesCount(ctx, mockedAPIQuery, mockedSearchClient)
			So(categories, ShouldBeNil)
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
			categories, err := getCategoriesTypesCount(ctx, mockedAPIQuery, mockedSearchClient)
			So(categories, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err, ShouldResemble, errors.New("filter type from client not available in data.go"))
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
			mockCategories := createMockCategories()
			mockCategories[0].Count = 3
			mockCategories[0].ContentTypes[0].Count = 3
			categories, err := getCategoriesTypesCount(ctx, mockedAPIQuery, mockedSearchClient)
			So(categories, ShouldNotBeNil)
			So(categories, ShouldResemble, mockCategories)
			So(err, ShouldBeNil)
		})

		Convey("successfully retrieve the count of filter types mapping to multiple filter types", func() {
			mockedAPIQuery = url.Values{
				"content_type": []string{"bulletin,article,article_download,static_article"},
				"q":            []string{"housing"},
			}
			mockCategories := createMockCategories()
			mockCategories[0].Count = 8
			mockCategories[0].ContentTypes[0].Count = 3
			mockCategories[0].ContentTypes[1].Count = 5
			categories, err := getCategoriesTypesCount(ctx, mockedAPIQuery, mockedSearchClient)
			So(categories, ShouldNotBeNil)
			So(categories, ShouldResemble, mockCategories)
			So(err, ShouldBeNil)
		})
	})

	Convey("When read is called", t, func() {
		req := httptest.NewRequest("GET", "/search?q=housing", nil)
		mockedRenderClient := &RenderClientMock{
			DoFunc: func(in1 string, in2 []byte) ([]byte, error) {
				return []byte(`<html><body><h1>Some HTML from renderer!</h1></body></html>`), nil
			},
		}

		Convey("convert mock response to client model", func() {
			sampleResponse, err := ioutil.ReadFile("../mapper/test_data/mock_response.json")
			So(err, ShouldBeNil)

			err = json.Unmarshal(sampleResponse, &respC)
			So(err, ShouldBeNil)

			mockedSearchClient := &SearchClientMock{
				GetSearchFunc: func(ctx context.Context, query url.Values) (searchC.Response, error) {
					return respC, nil
				},
			}

			Convey("return error as mapping filter types failed", func() {
				req = httptest.NewRequest("GET", "/search?q=housing&filter=INVALID", nil)
				w := doTestRequest("/search", req, Read(mockedRenderClient, mockedSearchClient), nil)
				So(w.Code, ShouldEqual, http.StatusBadRequest)
				So(len(mockedRenderClient.DoCalls()), ShouldEqual, 0)
				So(len(mockedSearchClient.GetSearchCalls()), ShouldEqual, 0)
			})

			Convey("successfully talks to the renderer to get the search page", func() {
				w := doTestRequest("/search", req, Read(mockedRenderClient, mockedSearchClient), nil)
				So(w.Code, ShouldEqual, http.StatusOK)
				So(w.Body.String(), ShouldEqual, "<html><body><h1>Some HTML from renderer!</h1></body></html>")
				So(len(mockedRenderClient.DoCalls()), ShouldEqual, 1)
				So(len(mockedSearchClient.GetSearchCalls()), ShouldEqual, 2)
			})

			Convey("return error as getting search response from client failed", func() {
				mockedSearchClient = &SearchClientMock{
					GetSearchFunc: func(ctx context.Context, query url.Values) (searchC.Response, error) {
						return searchC.Response{}, errors.New("internal server error")
					},
				}
				w := doTestRequest("/search", req, Read(mockedRenderClient, mockedSearchClient), nil)
				So(w.Code, ShouldEqual, http.StatusInternalServerError)
				So(len(mockedRenderClient.DoCalls()), ShouldEqual, 0)
				So(len(mockedSearchClient.GetSearchCalls()), ShouldEqual, 1)
			})

			Convey("return error as getting search page failed", func() {
				mockedRenderClient = &RenderClientMock{
					DoFunc: func(in1 string, in2 []byte) ([]byte, error) {
						return []byte{}, errors.New("internal server error")
					},
				}
				w := doTestRequest("/search", req, Read(mockedRenderClient, mockedSearchClient), nil)
				So(w.Code, ShouldEqual, http.StatusInternalServerError)
				So(len(mockedRenderClient.DoCalls()), ShouldEqual, 1)
				So(len(mockedSearchClient.GetSearchCalls()), ShouldEqual, 2)
			})
		})
	})
}
