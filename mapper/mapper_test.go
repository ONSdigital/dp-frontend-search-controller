package mapper

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ONSdigital/dp-frontend-search-controller/config"
	"github.com/ONSdigital/dp-frontend-search-controller/data"
	"github.com/ONSdigital/dp-renderer/model"
	. "github.com/smartystreets/goconvey/convey"
)

func TestUnitCreateSearchPageSuccess(t *testing.T) {
	t.Parallel()

	lang := "en"

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
		}

		categories := data.GetCategories()
		categories[0].Count = 1
		categories[0].ContentTypes[1].Count = 1

		topicCategories := data.GetTopicCategories()
		categories[0].Count = 1
		categories[0].ContentTypes[1].Count = 1

		respH, err := GetMockHomepageContent()
		So(err, ShouldBeNil)

		respC, err := GetMockSearchResponse()
		So(err, ShouldBeNil)

		respD, err := GetMockDepartmentResponse()
		So(err, ShouldBeNil)

		Convey("When CreateSearchPage is called", func() {
			sp := CreateSearchPage(cfg, req, mdl, validatedQueryParams, categories, topicCategories, respC, respD, lang, respH, "")

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

				So(sp.Data.Response.Items[0].Description.Contact.Name, ShouldEqual, "Name")
				So(sp.Data.Response.Items[0].Description.Contact.Telephone, ShouldEqual, "123")
				So(sp.Data.Response.Items[0].Description.Contact.Email, ShouldEqual, "test@ons.gov.uk")
				So(sp.Data.Response.Items[0].Description.Edition, ShouldEqual, "1995 to 2013")
				So(sp.Data.Response.Items[0].Description.Keywords, ShouldHaveLength, 4)
				So(*sp.Data.Response.Items[0].Description.LatestRelease, ShouldBeTrue)
				So(sp.Data.Response.Items[0].Description.MetaDescription, ShouldEqual, "Test Meta Description")
				So(*sp.Data.Response.Items[0].Description.NationalStatistic, ShouldBeFalse)
				So(sp.Data.Response.Items[0].Description.ReleaseDate, ShouldEqual, "2015-02-17T00:00:00.000Z")
				So(sp.Data.Response.Items[0].Description.Summary, ShouldEqual, "Test Summary")
				So(sp.Data.Response.Items[0].Description.Title, ShouldEqual, "Title Title")

				So(sp.Data.Response.Items[0].Type.Type, ShouldEqual, "article")
				So(sp.Data.Response.Items[0].Type.LocaliseKeyName, ShouldEqual, "Article")
				So(sp.Data.Response.Items[0].URI, ShouldEqual, "/uri1/housing/articles/uri2/2015-02-17")

				testMatchesDescSummary := *sp.Data.Response.Items[0].Matches.Description.Summary
				So(testMatchesDescSummary, ShouldHaveLength, 1)
				So(testMatchesDescSummary[0].Value, ShouldEqual, "summary")
				So(testMatchesDescSummary[0].Start, ShouldEqual, 1)
				So(testMatchesDescSummary[0].End, ShouldEqual, 5)

				testMatchesDescTitle := *sp.Data.Response.Items[0].Matches.Description.Title
				So(testMatchesDescTitle, ShouldHaveLength, 1)
				So(testMatchesDescTitle[0].Value, ShouldEqual, "title")
				So(testMatchesDescTitle[0].Start, ShouldEqual, 6)
				So(testMatchesDescTitle[0].End, ShouldEqual, 10)

				testMatchesDescEdition := *sp.Data.Response.Items[0].Matches.Description.Edition
				So(testMatchesDescEdition, ShouldHaveLength, 1)
				So(testMatchesDescEdition[0].Value, ShouldEqual, "edition")
				So(testMatchesDescEdition[0].Start, ShouldEqual, 11)
				So(testMatchesDescEdition[0].End, ShouldEqual, 15)

				testMatchesDescMetaDesc := *sp.Data.Response.Items[0].Matches.Description.MetaDescription
				So(testMatchesDescMetaDesc, ShouldHaveLength, 1)
				So(testMatchesDescMetaDesc[0].Value, ShouldEqual, "meta_description")
				So(testMatchesDescMetaDesc[0].Start, ShouldEqual, 16)
				So(testMatchesDescMetaDesc[0].End, ShouldEqual, 20)

				testMatchesDescKeywords := *sp.Data.Response.Items[0].Matches.Description.Keywords
				So(testMatchesDescKeywords, ShouldHaveLength, 1)
				So(testMatchesDescKeywords[0].Value, ShouldEqual, "keywords")
				So(testMatchesDescKeywords[0].Start, ShouldEqual, 21)
				So(testMatchesDescKeywords[0].End, ShouldEqual, 25)

				testMatchesDescDatasetID := *sp.Data.Response.Items[0].Matches.Description.DatasetID
				So(testMatchesDescDatasetID, ShouldHaveLength, 1)
				So(testMatchesDescDatasetID[0].Value, ShouldEqual, "dataset_id")
				So(testMatchesDescDatasetID[0].Start, ShouldEqual, 26)
				So(testMatchesDescDatasetID[0].End, ShouldEqual, 30)

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

				// NOTE until the API is built for topics, it seems silly to mock the data, some of the following lines have been removed until a later stage.
				So(len(sp.Data.TopicFilters[0].FilterKey), ShouldEqual, 8)
				So(sp.Data.TopicFilters[0].LocaliseKeyName, ShouldEqual, "Census")
				//So(sp.Data.TopicFilters[0].IsChecked, ShouldBeTrue)
				//So(sp.Data.TopicFilters[0].NumberOfResults, ShouldEqual, 1)
				So(len(sp.Data.TopicFilters[0].Types[0].FilterKey), ShouldEqual, 1)
				So(sp.Data.TopicFilters[0].Types[0].LocaliseKeyName, ShouldEqual, "DemographyAndMigration")
				So(sp.Data.TopicFilters[0].Types[0].IsChecked, ShouldBeFalse)
				So(sp.Data.TopicFilters[0].Types[0].NumberOfResults, ShouldEqual, 0)
				So(len(sp.Data.TopicFilters[0].Types[1].FilterKey), ShouldEqual, 1)
				So(sp.Data.TopicFilters[0].Types[1].LocaliseKeyName, ShouldEqual, "Education")
				//So(sp.Data.TopicFilters[0].Types[1].IsChecked, ShouldBeTrue)
				//So(sp.Data.TopicFilters[0].Types[1].NumberOfResults, ShouldEqual, 1)
				So(len(sp.Data.TopicFilters[0].Types[2].FilterKey), ShouldEqual, 1)
				So(sp.Data.TopicFilters[0].Types[2].LocaliseKeyName, ShouldEqual, "EthnicGroupNationalIdentityAndReligion")
				//So(sp.Data.TopicFilters[0].Types[2].IsChecked, ShouldBeFalse)
				So(sp.Data.TopicFilters[0].Types[2].NumberOfResults, ShouldEqual, 0)

				So(sp.Department.Code, ShouldEqual, "dept-code")
				So(sp.Department.URL, ShouldEqual, "www.dept.com")
				So(sp.Department.Name, ShouldEqual, "dept-name")
				So(sp.Department.Match, ShouldEqual, "dept-match")

				So(sp.ServiceMessage, ShouldEqual, respH.ServiceMessage)

				So(sp.EmergencyBanner.Type, ShouldEqual, strings.Replace(respH.EmergencyBanner.Type, "_", "-", -1))
				So(sp.EmergencyBanner.Title, ShouldEqual, respH.EmergencyBanner.Title)
				So(sp.EmergencyBanner.Description, ShouldEqual, respH.EmergencyBanner.Description)
				So(sp.EmergencyBanner.URI, ShouldEqual, respH.EmergencyBanner.URI)
				So(sp.EmergencyBanner.LinkText, ShouldEqual, respH.EmergencyBanner.LinkText)
			})
		})
	})
}
