package handlers

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	zebedeeC "github.com/ONSdigital/dp-api-clients-go/v2/zebedee"
	"github.com/ONSdigital/dp-frontend-search-controller/cache"
	"github.com/ONSdigital/dp-frontend-search-controller/config"
	"github.com/ONSdigital/dp-frontend-search-controller/mapper"
	"github.com/ONSdigital/dp-frontend-search-controller/mocks"
	"github.com/ONSdigital/dp-renderer/v2/helper"
	coreModel "github.com/ONSdigital/dp-renderer/v2/model"
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
			NewBasePageModelFunc: func() coreModel.Page {
				return coreModel.Page{}
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
			readRelatedData(w, req, cfg, mockedZebedeeClient, mockedRendererClient, mockedSearchClient, accessToken, collectionID, englishLang, *mockCacheList)

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
