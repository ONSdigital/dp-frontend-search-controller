package mapper

import (
	"testing"

	searchModels "github.com/ONSdigital/dp-search-api/models"
	. "github.com/smartystreets/goconvey/convey"
)

func TestUnitGetMockSearchResponse(t *testing.T) {
	t.Parallel()

	Convey("When GetMockSearchResponse is called", t, func() {
		mockSearchResponse, err := GetMockSearchResponse()

		Convey("Then successfully get mock search response", func() {
			mockSearchContentTypes := []searchModels.FilterCount{
				{
					Type:  "article",
					Count: 1,
				},
			}

			mockSearchTopics := []searchModels.FilterCount{
				{
					Type:  "1234",
					Count: 1,
				},
			}
			testString1 := "regional house prices"
			testString2 := "property prices"
			testString3 := "area with cheapest houses"
			testString4 := "area with most expensive houses"
			mockSearchItems := []searchModels.ESSourceDocument{
				{
					CanonicalTopic:  "1234",
					Keywords:        []string{"regional house prices", "property prices", "area with cheapest houses", "area with most expensive houses"},
					MetaDescription: "Test Meta Description",
					ReleaseDate:     "2015-02-17T00:00:00.000Z",
					Summary:         "Test Summary",
					Title:           "Title Title",
					Highlight: &searchModels.HighlightObj{
						Summary:  "Test Summary",
						Title:    "Title Title",
						Keywords: []*string{&testString1, &testString2, &testString3, &testString4}, // "regional house prices", "property prices", "area with cheapest houses", "area with most expensive houses"},
					},
					DataType: "article",
					URI:      "/uri1/housing/articles/uri2/2015-02-17",
				},
			}

			So(mockSearchResponse, ShouldResemble, &searchModels.SearchResponse{
				Count:        1,
				Took:         96,
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
