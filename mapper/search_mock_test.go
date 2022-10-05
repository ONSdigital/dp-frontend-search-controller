package mapper

import (
	"testing"

	searchC "github.com/ONSdigital/dp-api-clients-go/v2/site-search"
	. "github.com/smartystreets/goconvey/convey"
)

func TestUnitGetMockSearchResponse(t *testing.T) {
	t.Parallel()

	Convey("When GetMockSearchResponse is called", t, func() {
		mockSearchResponse, err := GetMockSearchResponse()

		Convey("Then successfully get mock search response", func() {

			mockSearchContentTypes := []searchC.FilterCount{
				{
					Type:  "article",
					Count: 1,
				},
			}

			mockSearchTopics := []searchC.FilterCount{
				{
					Type:  "1234",
					Count: 1,
				},
			}

			mockSearchDescription := searchC.Description{
				Keywords:        []string{"regional house prices", "property prices", "area with cheapest houses", "area with most expensive houses"},
				MetaDescription: "Test Meta Description",
				ReleaseDate:     "2015-02-17T00:00:00.000Z",
				Summary:         "Test Summary",
				Title:           "Title Title",
				Highlight: &searchC.Highlight{
					Summary:  "Test Summary",
					Title:    "Title Title",
					Keywords: &[]string{"regional house prices", "property prices", "area with cheapest houses", "area with most expensive houses"},
				},
			}

			mockSearchItems := []searchC.ContentItem{
				{
					Description: mockSearchDescription,
					Type:        "article",
					URI:         "/uri1/housing/articles/uri2/2015-02-17",
				},
			}

			So(mockSearchResponse, ShouldResemble, searchC.Response{
				Count:        1,
				ContentTypes: mockSearchContentTypes,
				Topics:       mockSearchTopics,
				Items:        mockSearchItems,
			})

		})

		Convey("And return no error", func() {
			So(err, ShouldBeNil)
		})
	})
}
