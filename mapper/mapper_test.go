package mapper

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http/httptest"
	"testing"

	searchC "github.com/ONSdigital/dp-api-clients-go/site-search"
	"github.com/ONSdigital/dp-frontend-search-controller/data"
	. "github.com/smartystreets/goconvey/convey"
)

var respC searchC.Response

func TestUnitMapper(t *testing.T) {
	ctx := context.Background()
	categories := []data.Category{data.Publication, data.Data, data.Other}

	Convey("When search requested with valid query", t, func() {
		req := httptest.NewRequest("GET", "/search?q=housing&limit=3&offset=20&filter=article&filter=filter2&sort=relevance", nil)
		query := req.URL.Query()
		categories[0].Count = 1
		categories[0].ContentTypes[1].Count = 1

		Convey("convert mock response to client model", func() {
			sampleResponse, err := ioutil.ReadFile("test_data/mock_response.json")
			So(err, ShouldBeNil)

			err = json.Unmarshal(sampleResponse, &respC)
			So(err, ShouldBeNil)

			Convey("successfully map search response from search-query client to page model", func() {
				sp := CreateSearchPage(ctx, query, respC, categories)
				So(sp.Data.Query, ShouldEqual, "housing")
				So(sp.Data.Filter.Query, ShouldHaveLength, 2)
				So(sp.Data.Filter.Query[0], ShouldEqual, "article")
				So(sp.Data.Filter.Query[1], ShouldEqual, "filter2")
				So(sp.Data.Filter.Options, ShouldResemble, []string{"Publication", "Data", "Other"})
				So(sp.Data.Sort.Query, ShouldEqual, "relevance")
				So(sp.Data.Sort.LocaliseFilterKeys, ShouldResemble, []string{"Article"})
				So(sp.Data.Sort.FilterText, ShouldEqual, "relevance")
				So(sp.Data.Limit, ShouldEqual, 3)
				So(sp.Data.Offset, ShouldEqual, 20)

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
		query := req.URL.Query()

		Convey("mapping search response fails from search-query client to page model", func() {
			sp := CreateSearchPage(ctx, query, respC, categories)
			So(sp.Data.Limit, ShouldEqual, 10)
			So(sp.Data.Offset, ShouldEqual, 0)
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
}
