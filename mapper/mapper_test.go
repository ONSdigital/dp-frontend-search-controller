package mapper

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http/httptest"
	"testing"

	searchC "github.com/ONSdigital/dp-api-clients-go/site-search"
	. "github.com/smartystreets/goconvey/convey"
)

var respC searchC.Response

func TestUnitMapper(t *testing.T) {
	ctx := context.Background()

	Convey("When search requested with valid query", t, func() {
		req := httptest.NewRequest("GET", "/search?q=housing", nil)
		query := req.URL.Query()

		Convey("convert mock response to client model", func() {
			sampleResponse, err := ioutil.ReadFile("test_data/mock_response.json")
			So(err, ShouldBeNil)

			err = json.Unmarshal(sampleResponse, &respC)
			So(err, ShouldBeNil)

			Convey("successfully map search response from search-query client to page model", func() {
				sp := CreateSearchPage(ctx, query, respC)
				So(sp.Data.Query, ShouldEqual, "housing")
				So(sp.Data.Filter, ShouldBeEmpty)
				So(sp.Data.Sort, ShouldBeEmpty)
				So(sp.Data.Limit, ShouldBeEmpty)
				So(sp.Data.Offset, ShouldBeEmpty)

				So(sp.Data.Response.Count, ShouldEqual, 1)

				So(sp.Data.Response.ContentTypes, ShouldHaveLength, 1)
				So(sp.Data.Response.ContentTypes[0].Type, ShouldEqual, "article")
				So(sp.Data.Response.ContentTypes[0].Count, ShouldEqual, 1)

				So(sp.Data.Response.Items, ShouldHaveLength, 1)

				So(sp.Data.Response.Items[0].Description.Contact.Name, ShouldEqual, "Name")
				So(sp.Data.Response.Items[0].Description.Contact.Telephone, ShouldEqual, "123")
				So(sp.Data.Response.Items[0].Description.Contact.Email, ShouldEqual, "test@ons.gov.uk")
				So(sp.Data.Response.Items[0].Description.Edition, ShouldEqual, "1995 to 2013")
				So(sp.Data.Response.Items[0].Description.Keywords, ShouldHaveLength, 4)
				So(sp.Data.Response.Items[0].Description.LatestRelease, ShouldBeTrue)
				So(sp.Data.Response.Items[0].Description.MetaDescription, ShouldEqual, "Test Meta Description")
				So(sp.Data.Response.Items[0].Description.NationalStatistic, ShouldBeFalse)
				So(sp.Data.Response.Items[0].Description.ReleaseDate, ShouldEqual, "2015-02-17T00:00:00.000Z")
				So(sp.Data.Response.Items[0].Description.Summary, ShouldEqual, "Test Summary")
				So(sp.Data.Response.Items[0].Description.Title, ShouldEqual, "Title Title")

				So(sp.Data.Response.Items[0].Type, ShouldEqual, "article")
				So(sp.Data.Response.Items[0].URI, ShouldEqual, "/uri1/housing/articles/uri2/2015-02-17")

				So(sp.Data.Response.Items[0].Matches.Description.Summary, ShouldBeNil)
				So(sp.Data.Response.Items[0].Matches.Description.Title, ShouldHaveLength, 1)
				So(sp.Data.Response.Items[0].Matches.Description.Keywords, ShouldHaveLength, 3)

			})
		})
	})
}
