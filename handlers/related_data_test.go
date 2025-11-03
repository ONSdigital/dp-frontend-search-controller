package handlers

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ONSdigital/dis-design-system-go/helper"
	core "github.com/ONSdigital/dis-design-system-go/model"
	zebedeeC "github.com/ONSdigital/dp-api-clients-go/v2/zebedee"
	"github.com/ONSdigital/dp-frontend-search-controller/cache"
	"github.com/ONSdigital/dp-frontend-search-controller/config"
	"github.com/ONSdigital/dp-frontend-search-controller/mapper"
	"github.com/ONSdigital/dp-frontend-search-controller/mocks"
	searchAPI "github.com/ONSdigital/dp-search-api/api"
	searchModels "github.com/ONSdigital/dp-search-api/models"
	searchSDK "github.com/ONSdigital/dp-search-api/sdk"
	searchError "github.com/ONSdigital/dp-search-api/sdk/errors"
	. "github.com/smartystreets/goconvey/convey"
)

func TestUnitReadRelatedData(t *testing.T) {
	helper.InitialiseLocalisationsHelper(mocks.MockAssetFunction)

	ctx := context.Background()

	mockZebedeePageContent, err := mapper.GetMockZebedeePageDataResponse()
	if err != nil {
		t.Errorf("failed to retrieve mock zebedee page data content for unit tests, failing early: %v", err)
	}

	mockHomepageContent, err := mapper.GetMockHomepageContent()
	if err != nil {
		t.Errorf("failed to retrieve mock homepage content for unit tests, failing early: %v", err)
	}

	mockSearchResponse, err := mapper.GetMockSearchResponse()
	if err != nil {
		t.Errorf("failed to retrieve mock search response for unit tests, failing early: %v", err)
	}

	mockBreadcrumbContent, err := mapper.GetMockBreadcrumbResponse()
	if err != nil {
		t.Errorf("failed to retrieve mock breadcrumb content for unit tests, failing early: %v", err)
	}

	Convey("Given a valid request", t, func() {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/foo/bar/relateddata", http.NoBody)

		cfg, err := config.Get()
		So(err, ShouldBeNil)

		mockedRendererClient := &RenderClientMock{
			BuildPageFunc: func(w io.Writer, pageModel interface{}, templateName string) {},
			NewBasePageModelFunc: func() core.Page {
				return core.Page{}
			},
		}

		mockedSearchClient := &SearchClientMock{
			PostSearchURIsFunc: func(ctx context.Context, options searchSDK.Options, urisRequest searchAPI.URIsRequest) (*searchModels.SearchResponse, searchError.Error) {
				return mockSearchResponse, nil
			},
		}

		mockedZebedeeClient := &ZebedeeClientMock{
			GetHomepageContentFunc: func(ctx context.Context, userAuthToken, collectionID, lang, path string) (zebedeeC.HomepageContent, error) {
				return mockHomepageContent, nil
			},
			GetPageDataFunc: func(ctx context.Context, userAuthToken, collectionID, lang, path string) (zebedeeC.PageData, error) {
				return mockZebedeePageContent, nil
			},
			GetBreadcrumbFunc: func(ctx context.Context, userAuthToken, collectionID, lang, path string) ([]zebedeeC.Breadcrumb, error) {
				return mockBreadcrumbContent, nil
			},
		}

		mockCacheList, err := cache.GetMockCacheList(ctx, englishLang)
		So(err, ShouldBeNil)

		Convey("When readRelatedData is called", func() {
			relatedDataConfig := NewRelatedDataConfig(*req)
			handleReadRequest(w, req, cfg, mockedZebedeeClient, mockedRendererClient, mockedSearchClient, accessToken, collectionID, englishLang, *mockCacheList, relatedDataConfig)
			Convey("Then a 200 OK status should be returned", func() {
				So(w.Code, ShouldEqual, http.StatusOK)

				So(mockedZebedeeClient.GetPageDataCalls(), ShouldHaveLength, 1)
				So(mockedZebedeeClient.GetHomepageContentCalls(), ShouldHaveLength, 1)
				So(mockedSearchClient.PostSearchURIsCalls(), ShouldHaveLength, 1)
				So(mockedZebedeeClient.GetBreadcrumbCalls(), ShouldHaveLength, 1)
				So(mockedRendererClient.BuildPageCalls(), ShouldHaveLength, 1)
			})
		})
	})
}

func TestUnitReadRelatedDataWithMigrationLink(t *testing.T) {
	Convey("Given a search handler and zebedee client with a migration link", t, func() {
		cfg, err := config.Get()
		So(err, ShouldBeNil)

		mockedZebedeeClient := &ZebedeeClientMock{
			GetPageDataFunc: func(ctx context.Context, userAuthToken, collectionID, lang, path string) (zebedeeC.PageData, error) {
				return zebedeeC.PageData{
					Type: "bulletin",
					Description: zebedeeC.Description{
						Title:         "My test bulletin",
						Edition:       "March 2024",
						MigrationLink: "/new-average-earnings",
					},
				}, nil
			},
		}

		mockSearchHandler := NewSearchHandler(&RenderClientMock{}, &SearchClientMock{}, &TopicClientMock{}, mockedZebedeeClient, cfg, cache.List{})

		Convey("When /relateddata is called", func() {
			req := httptest.NewRequest("GET", "/foo/latest/relateddata", http.NoBody)

			Convey("Then a 308 redirect should be returned", func() {
				w := doTestRequest("/{uri:.*}/relateddata", req, mockSearchHandler.RelatedData(cfg), nil)
				location := w.Header().Get("Location")
				expectedLocation := "/new-average-earnings/related-data"

				So(w.Code, ShouldEqual, http.StatusPermanentRedirect)
				So(location, ShouldEqual, expectedLocation)
			})
		})
	})
}
