package mapper

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http/httptest"
	"testing"

	searchC "github.com/ONSdigital/dp-api-clients-go/site-search"
	"github.com/ONSdigital/dp-frontend-models/model"
	"github.com/ONSdigital/dp-frontend-search-controller/data"
	. "github.com/smartystreets/goconvey/convey"
)

var respC searchC.Response

func TestUnitMapper(t *testing.T) {
	ctx := context.Background()
	categories := []data.Category{data.Publication, data.Data, data.Other}

	Convey("When search requested with valid query", t, func() {
		req := httptest.NewRequest("GET", "/search?q=housing&filter=article&filter=filter2&page=2", nil)
		url := req.URL
		categories[0].Count = 1
		categories[0].ContentTypes[1].Count = 1

		Convey("convert mock response to client model", func() {
			sampleResponse, err := ioutil.ReadFile("test_data/mock_response.json")
			So(err, ShouldBeNil)

			err = json.Unmarshal(sampleResponse, &respC)
			So(err, ShouldBeNil)

			Convey("successfully map search response from search-query client to page model", func() {
				sp := CreateSearchPage(ctx, url, respC, categories)
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
				So(sp.Data.Pagination.PagesToDisplay[0].URL, ShouldEqual, "/search?q=housing&filter=article&filter=filter2&page=1")
				So(sp.Data.Pagination.Limit, ShouldEqual, 10)
				So(sp.Data.Pagination.LimitOptions, ShouldResemble, []int{10, 25, 50})

				So(sp.Data.Response.Count, ShouldEqual, 1)

				So(sp.Data.Response.Categories, ShouldHaveLength, 3)
				So(sp.Data.Response.Categories[0].Count, ShouldEqual, 1)
				So(sp.Data.Response.Categories[0].LocaliseKeyName, ShouldEqual, "Publication")
				So(sp.Data.Response.Categories[0].ContentTypes, ShouldHaveLength, 3)

				So(sp.Data.Response.Categories[0].ContentTypes[0].Type, ShouldEqual, "bulletin")
				So(sp.Data.Response.Categories[0].ContentTypes[0].Count, ShouldEqual, 0)
				So(sp.Data.Response.Categories[0].ContentTypes[0].LocaliseKeyName, ShouldEqual, "StatisticalBulletin")

				So(sp.Data.Response.Categories[0].ContentTypes[1].Type, ShouldEqual, "article")
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

				So(sp.Data.Response.Items[0].Type, ShouldEqual, "article")
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
			})
		})
	})

	Convey("When search requested with invalid query", t, func() {
		req := httptest.NewRequest("GET", "/search?q=housing&limit=invalid&offset=invalid", nil)
		url := req.URL

		Convey("mapping search response fails from search-query client to page model", func() {
			sp := CreateSearchPage(ctx, url, respC, categories)
			So(sp.Data.Pagination.Limit, ShouldEqual, 10)
			So(sp.Data.Pagination.CurrentPage, ShouldEqual, 1)
		})
	})

	Convey("When getFilterSortText is called", t, func() {
		Convey("successfully add one filter given", func() {
			req := httptest.NewRequest("GET", "/search?q=housing&filter=article", nil)
			query := req.URL.Query()
			filterSortText := getFilterSortKeyList(query, categories)
			So(filterSortText, ShouldResemble, []string{"Article"})
		})

		Convey("successfully add two filters given", func() {
			req := httptest.NewRequest("GET", "/search?q=housing&filter=article&filter=compendia", nil)
			query := req.URL.Query()
			filterSortText := getFilterSortKeyList(query, categories)
			So(filterSortText, ShouldResemble, []string{"Article", "Compendium"})
		})

		Convey("successfully add three or more filters given", func() {
			req := httptest.NewRequest("GET", "/search?q=housing&filter=article&filter=compendia&filter=methodology", nil)
			query := req.URL.Query()
			filterSortText := getFilterSortKeyList(query, categories)
			So(filterSortText, ShouldResemble, []string{"Article", "Compendium", "Methodology"})
		})

		Convey("successfully add no filters given", func() {
			req := httptest.NewRequest("GET", "/search?q=housing", nil)
			query := req.URL.Query()
			filterSortText := getFilterSortKeyList(query, categories)
			So(filterSortText, ShouldResemble, []string{})
		})
	})

	Convey("When getSortLocaliseKey is called", t, func() {
		Convey("successfully get localisation key for sort query", func() {
			req := httptest.NewRequest("GET", "/search?q=housing&sort=relevance", nil)
			query := req.URL.Query()
			sortNameKey := getSortLocaliseKey(query)
			So(sortNameKey, ShouldEqual, "Relevance")
		})

		Convey("successfully get no localisation key for invalid sort query", func() {
			req := httptest.NewRequest("GET", "/search?q=housing&sort=invalid", nil)
			query := req.URL.Query()
			sortNameKey := getSortLocaliseKey(query)
			So(sortNameKey, ShouldEqual, "")
		})

		Convey("successfully get no localisation key when no sort query given", func() {
			req := httptest.NewRequest("GET", "/search?q=housing", nil)
			query := req.URL.Query()
			sortNameKey := getSortLocaliseKey(query)
			So(sortNameKey, ShouldEqual, "")
		})
	})

	Convey("When getLimitOptions is called", t, func() {
		Convey("successfully get limit options", func() {
			limitOptions := getLimitOptions()
			So(limitOptions, ShouldResemble, []int{10, 25, 50})
		})
	})

	Convey("When getPagesToDisplay is called", t, func() {
		Convey("successfully get pages to display when current page is less than or equal to 2", func() {
			req := httptest.NewRequest("GET", "/search?q=housing&sort=relevance", nil)
			pagesToDisplay := getPagesToDisplay(1, 200, req.URL)
			So(pagesToDisplay, ShouldResemble, []model.PageToDisplay{
				{
					PageNumber: 1,
					URL:        "/search?q=housing&sort=relevance&page=1",
				},
				{
					PageNumber: 2,
					URL:        "/search?q=housing&sort=relevance&page=2",
				},
				{
					PageNumber: 3,
					URL:        "/search?q=housing&sort=relevance&page=3",
				},
				{
					PageNumber: 4,
					URL:        "/search?q=housing&sort=relevance&page=4",
				},
				{
					PageNumber: 5,
					URL:        "/search?q=housing&sort=relevance&page=5",
				},
			})
		})

		Convey("successfully get pages to display when current page is more than or equal to totalPages-1", func() {
			req := httptest.NewRequest("GET", "/search?q=housing&sort=relevance", nil)
			pagesToDisplay := getPagesToDisplay(199, 200, req.URL)
			So(pagesToDisplay, ShouldResemble, []model.PageToDisplay{
				{
					PageNumber: 196,
					URL:        "/search?q=housing&sort=relevance&page=196",
				},
				{
					PageNumber: 197,
					URL:        "/search?q=housing&sort=relevance&page=197",
				},
				{
					PageNumber: 198,
					URL:        "/search?q=housing&sort=relevance&page=198",
				},
				{
					PageNumber: 199,
					URL:        "/search?q=housing&sort=relevance&page=199",
				},
				{
					PageNumber: 200,
					URL:        "/search?q=housing&sort=relevance&page=200",
				},
			})
		})

		Convey("successfully get pages to display when current page is between 3 and totalPages-2 ", func() {
			req := httptest.NewRequest("GET", "/search?q=housing&sort=relevance", nil)
			pagesToDisplay := getPagesToDisplay(150, 200, req.URL)
			So(pagesToDisplay, ShouldResemble, []model.PageToDisplay{
				{
					PageNumber: 148,
					URL:        "/search?q=housing&sort=relevance&page=148",
				},
				{
					PageNumber: 149,
					URL:        "/search?q=housing&sort=relevance&page=149",
				},
				{
					PageNumber: 150,
					URL:        "/search?q=housing&sort=relevance&page=150",
				},
				{
					PageNumber: 151,
					URL:        "/search?q=housing&sort=relevance&page=151",
				},
				{
					PageNumber: 152,
					URL:        "/search?q=housing&sort=relevance&page=152",
				},
			})
		})
	})
}
