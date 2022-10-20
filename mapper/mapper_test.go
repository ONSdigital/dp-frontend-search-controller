package mapper

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ONSdigital/dp-frontend-search-controller/config"
	"github.com/ONSdigital/dp-frontend-search-controller/data"
	"github.com/ONSdigital/dp-renderer/model"
	"github.com/ONSdigital/dp-topic-api/models"
	. "github.com/smartystreets/goconvey/convey"
)

var (
	mockTopicCategories = []data.Topic{
		{
			LocaliseKeyName: "Census",
			Count:           1,
			Query:           "1234",
			ShowInWebUI:     true,
		},
	}
)

const englishLang string = "en"

func TestUnitCreateSearchPage(t *testing.T) {
	t.Parallel()

	Convey("Given validated query and response from search-api", t, func() {
		cfg, err := config.Get()
		So(err, ShouldBeNil)
		req := httptest.NewRequest("", "/", nil)
		mdl := model.Page{}

		validatedQueryParams := data.SearchURLParams{
			Query: "housing",

			Filter: data.Filter{
				Query:           []string{"article", "filter2"},
				LocaliseKeyName: []string{"Article"},
			},

			Sort: data.Sort{
				Query:           "relevance",
				LocaliseKeyName: "Relevance",
			},

			Limit:       10,
			CurrentPage: 1,
			TopicFilter: "1234",
		}

		categories := data.GetCategories()
		categories[0].Count = 1
		categories[0].ContentTypes[1].Count = 1

		topicCategories := mockTopicCategories

		respH, err := GetMockHomepageContent()
		So(err, ShouldBeNil)

		respC, err := GetMockSearchResponse()
		So(err, ShouldBeNil)

		Convey("When CreateSearchPage is called", func() {
			// NOTE: temporary measure until topic filter feature flag is removed
			cfg.EnableCensusTopicFilterOption = true

			sp := CreateSearchPage(cfg, req, mdl, validatedQueryParams, categories, topicCategories, respC, englishLang, respH, "", &models.Navigation{})

			Convey("Then successfully map search response from search-query client to page model", func() {
				So(sp.Data.Query, ShouldEqual, "housing")
				So(sp.Data.Filter, ShouldHaveLength, 2)
				So(sp.Data.Filter[0], ShouldEqual, "article")
				So(sp.Data.Filter[1], ShouldEqual, "filter2")

				So(sp.Data.Sort.Query, ShouldEqual, "relevance")
				So(sp.Data.Sort.LocaliseFilterKeys, ShouldResemble, []string{"Article"})
				So(sp.Data.Sort.LocaliseSortKey, ShouldEqual, "Relevance")
				So(sp.Data.Sort.Options[0].Query, ShouldEqual, "relevance")
				So(sp.Data.Sort.Options[0].LocaliseKeyName, ShouldEqual, "Relevance")

				So(sp.Data.Pagination.CurrentPage, ShouldEqual, 1)
				So(sp.Data.Pagination.TotalPages, ShouldEqual, 1)
				So(sp.Data.Pagination.PagesToDisplay, ShouldHaveLength, 1)
				So(sp.Data.Pagination.PagesToDisplay[0].PageNumber, ShouldEqual, 1)
				So(sp.Data.Pagination.PagesToDisplay[0].URL, ShouldEqual, "/search?q=housing&filter=article&filter=filter2&limit=10&sort=relevance&page=1")
				So(sp.Data.Pagination.Limit, ShouldEqual, 10)
				So(sp.Data.Pagination.LimitOptions, ShouldResemble, []int{10, 25, 50})

				So(sp.Data.Response.Count, ShouldEqual, 1)

				So(sp.Data.Response.Categories, ShouldHaveLength, 3)
				So(sp.Data.Response.Categories[0].Count, ShouldEqual, 1)
				So(sp.Data.Response.Categories[0].LocaliseKeyName, ShouldEqual, "Publication")
				So(sp.Data.Response.Categories[0].ContentTypes, ShouldHaveLength, 3)

				So(sp.Data.Response.Categories[0].ContentTypes[0].Group, ShouldEqual, "bulletin")
				So(sp.Data.Response.Categories[0].ContentTypes[0].Count, ShouldEqual, 0)
				So(sp.Data.Response.Categories[0].ContentTypes[0].LocaliseKeyName, ShouldEqual, "StatisticalBulletin")

				So(sp.Data.Response.Categories[0].ContentTypes[1].Group, ShouldEqual, "article")
				So(sp.Data.Response.Categories[0].ContentTypes[1].Count, ShouldEqual, 1)
				So(sp.Data.Response.Categories[0].ContentTypes[1].LocaliseKeyName, ShouldEqual, "Article")

				So(sp.Data.Response.Items, ShouldHaveLength, 1)

				So(sp.Data.Response.Items[0].Description.Keywords, ShouldHaveLength, 4)
				So(sp.Data.Response.Items[0].Description.MetaDescription, ShouldEqual, "Test Meta Description")
				So(sp.Data.Response.Items[0].Description.ReleaseDate, ShouldEqual, "2015-02-17T00:00:00.000Z")
				So(sp.Data.Response.Items[0].Description.Summary, ShouldEqual, "Test Summary")
				So(sp.Data.Response.Items[0].Description.Title, ShouldEqual, "Title Title")

				So(sp.Data.Response.Items[0].Type.Type, ShouldEqual, "article")
				So(sp.Data.Response.Items[0].Type.LocaliseKeyName, ShouldEqual, "Article")
				So(sp.Data.Response.Items[0].URI, ShouldEqual, "/uri1/housing/articles/uri2/2015-02-17")

				So(len(sp.Data.Filters[0].FilterKey), ShouldEqual, 3)
				So(sp.Data.Filters[0].LocaliseKeyName, ShouldEqual, "Publication")
				So(sp.Data.Filters[0].IsChecked, ShouldBeTrue)
				So(sp.Data.Filters[0].NumberOfResults, ShouldEqual, 1)
				So(len(sp.Data.Filters[0].Types[0].FilterKey), ShouldEqual, 1)
				So(sp.Data.Filters[0].Types[0].LocaliseKeyName, ShouldEqual, "StatisticalBulletin")
				So(sp.Data.Filters[0].Types[0].IsChecked, ShouldBeFalse)
				So(sp.Data.Filters[0].Types[0].NumberOfResults, ShouldEqual, 0)
				So(len(sp.Data.Filters[0].Types[1].FilterKey), ShouldEqual, 1)
				So(sp.Data.Filters[0].Types[1].LocaliseKeyName, ShouldEqual, "Article")
				So(sp.Data.Filters[0].Types[1].IsChecked, ShouldBeTrue)
				So(sp.Data.Filters[0].Types[1].NumberOfResults, ShouldEqual, 1)
				So(len(sp.Data.Filters[0].Types[2].FilterKey), ShouldEqual, 1)
				So(sp.Data.Filters[0].Types[2].LocaliseKeyName, ShouldEqual, "Compendium")
				So(sp.Data.Filters[0].Types[2].IsChecked, ShouldBeFalse)
				So(sp.Data.Filters[0].Types[2].NumberOfResults, ShouldEqual, 0)
				So(len(sp.Data.Filters[2].Types), ShouldEqual, 3)

				So(sp.Data.TopicFilters, ShouldNotBeEmpty)
				So(sp.Data.TopicFilters[0], ShouldNotBeEmpty)
				So(sp.Data.TopicFilters[0].IsChecked, ShouldBeTrue)
				So(sp.Data.TopicFilters[0].LocaliseKeyName, ShouldEqual, "Census")
				So(sp.Data.TopicFilters[0].NumberOfResults, ShouldEqual, 1)
				So(sp.Data.TopicFilters[0].Query, ShouldEqual, "1234")

				So(sp.ServiceMessage, ShouldEqual, respH.ServiceMessage)

				So(sp.EmergencyBanner.Type, ShouldEqual, strings.Replace(respH.EmergencyBanner.Type, "_", "-", -1))
				So(sp.EmergencyBanner.Title, ShouldEqual, respH.EmergencyBanner.Title)
				So(sp.EmergencyBanner.Description, ShouldEqual, respH.EmergencyBanner.Description)
				So(sp.EmergencyBanner.URI, ShouldEqual, respH.EmergencyBanner.URI)
				So(sp.EmergencyBanner.LinkText, ShouldEqual, respH.EmergencyBanner.LinkText)

				So(sp.SearchNoIndexEnabled, ShouldEqual, false)
			})
		})
	})
}
