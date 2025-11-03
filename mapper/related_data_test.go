package mapper

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ONSdigital/dis-design-system-go/helper"
	core "github.com/ONSdigital/dis-design-system-go/model"
	"github.com/ONSdigital/dp-frontend-search-controller/cache"
	"github.com/ONSdigital/dp-frontend-search-controller/config"
	"github.com/ONSdigital/dp-frontend-search-controller/data"
	"github.com/ONSdigital/dp-frontend-search-controller/mocks"
	topicModels "github.com/ONSdigital/dp-topic-api/models"
	. "github.com/smartystreets/goconvey/convey"
)

func TestCreateRelatedDataPage(t *testing.T) {
	helper.InitialiseLocalisationsHelper(mocks.MockAssetFunction)

	Convey("Given validated query and response from search-api", t, func() {
		cfg, err := config.Get()
		So(err, ShouldBeNil)
		req := httptest.NewRequest("", "/foo/bar/relateddata", http.NoBody)
		mdl := core.Page{}

		validatedQueryParams := data.SearchURLParams{
			Limit:       10,
			CurrentPage: 1,
		}

		respZ, err := GetMockZebedeePageDataResponse()
		So(err, ShouldBeNil)

		respH, err := GetMockHomepageContent()
		So(err, ShouldBeNil)

		respC, err := GetMockSearchResponse()
		So(err, ShouldBeNil)

		respBc, err := GetMockBreadcrumbResponse()
		So(err, ShouldBeNil)

		Convey("When CreateRelatedDataPage is called", func() {
			sp := CreateRelatedDataPage(cfg, req, mdl, validatedQueryParams, respC, englishLang, respH, "", &topicModels.Navigation{}, "", cache.Topic{}, nil, respZ, respBc)

			Convey("Then successfully map search response from search-query client to page model", func() {
				So(sp.Data.Pagination.CurrentPage, ShouldEqual, 1)
				So(sp.Data.Pagination.TotalPages, ShouldEqual, 1)
				So(sp.Data.Pagination.PagesToDisplay, ShouldHaveLength, 1)
				So(sp.Data.Pagination.PagesToDisplay[0].PageNumber, ShouldEqual, 1)
				So(sp.Data.Pagination.PagesToDisplay[0].URL, ShouldStartWith, "/foo/bar/relateddata")
				So(sp.Data.Pagination.Limit, ShouldEqual, 10)
				So(sp.Data.Pagination.LimitOptions, ShouldResemble, []int{10, 25, 50})

				So(sp.Data.Response.Count, ShouldEqual, 1)
				So(sp.Data.Response.Items, ShouldHaveLength, 1)

				So(sp.Data.Response.Items[0].Description.Keywords, ShouldHaveLength, 4)
				So(sp.Data.Response.Items[0].Description.MetaDescription, ShouldEqual, "Test Meta Description")
				So(sp.Data.Response.Items[0].Description.ReleaseDate, ShouldEqual, "2015-02-17T00:00:00.000Z")
				So(sp.Data.Response.Items[0].Description.Summary, ShouldEqual, "Test Summary")
				So(sp.Data.Response.Items[0].Description.Title, ShouldEqual, "Title Title")

				So(sp.Data.Response.Items[0].Type.Type, ShouldEqual, "article")
				So(sp.Data.Response.Items[0].Type.LocaliseKeyName, ShouldEqual, "Article")
				So(sp.Data.Response.Items[0].URI, ShouldEqual, "/uri1/housing/articles/uri2/2015-02-17")

				So(sp.ServiceMessage, ShouldEqual, respH.ServiceMessage)

				So(len(sp.Breadcrumb), ShouldEqual, 4)
				So(sp.Breadcrumb[0], ShouldResemble, expectedMappedBreadcrumb[0])
				So(sp.Breadcrumb[1], ShouldResemble, expectedMappedBreadcrumb[1])
				So(sp.Breadcrumb[2], ShouldResemble, expectedMappedBreadcrumb[2])
				So(sp.Breadcrumb[3].Title, ShouldEqual, "Foo bar bulletin")
				So(sp.Breadcrumb[3].URI, ShouldEqual, "foo/bar/1/2/3")

				So(sp.EmergencyBanner.Type, ShouldEqual, strings.Replace(respH.EmergencyBanner.Type, "_", "-", -1))
				So(sp.EmergencyBanner.Title, ShouldEqual, respH.EmergencyBanner.Title)
				So(sp.EmergencyBanner.Description, ShouldEqual, respH.EmergencyBanner.Description)
				So(sp.EmergencyBanner.URI, ShouldEqual, respH.EmergencyBanner.URI)
				So(sp.EmergencyBanner.LinkText, ShouldEqual, respH.EmergencyBanner.LinkText)
			})
		})

		Convey("When CreateRelatedDataPage is called with validation errors", func() {
			validationErrs := []core.ErrorItem{
				{
					Description: core.Localisation{
						Text: "This is a current page error",
					},
					ID:  "currentPage-error",
					URL: "#currentPage-error",
				},
			}

			sp := CreateRelatedDataPage(cfg, req, mdl, validatedQueryParams, respC, englishLang, respH, "", &topicModels.Navigation{}, "", cache.Topic{}, validationErrs, respZ, respBc)

			Convey("Then validation errors are successfully mapped to the page model", func() {
				So(sp.Error.ErrorItems, ShouldResemble, validationErrs)
			})
		})
	})
}
