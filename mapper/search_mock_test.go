package mapper

import (
	"testing"

	searchC "github.com/ONSdigital/dp-api-clients-go/v2/site-search"
	. "github.com/smartystreets/goconvey/convey"
)

var (
	trueValue   = true
	truePointer = &trueValue

	falseValue   = false
	falsePointer = &falseValue
)

func TestUnitGetMockSearchResponseSuccess(t *testing.T) {
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

			mockSearchItems := []searchC.ContentItem{
				{
					Description: searchC.Description{
						CanonicalTopic: "1234",
						Contact: &searchC.Contact{
							Name:      "Name",
							Telephone: "123",
							Email:     "test@ons.gov.uk",
						},
						Edition:           "1995 to 2013",
						Keywords:          &[]string{"regional house prices", "property prices", "area with cheapest houses", "area with most expensive houses"},
						LatestRelease:     truePointer,
						MetaDescription:   "Test Meta Description",
						NationalStatistic: falsePointer,
						ReleaseDate:       "2015-02-17T00:00:00.000Z",
						Source:            "",
						Summary:           "Test Summary",
						Title:             "Title Title",
						Unit:              "",
						Highlight: searchC.Highlight{
							Summary:  "Test Summary",
							Title:    "Title Title",
							Keywords: &[]string{"regional house prices", "property prices", "area with cheapest houses", "area with most expensive houses"},
							Edition:  "1995 to 2013",
						},
					},
					Type: "article",
					URI:  "/uri1/housing/articles/uri2/2015-02-17",

					Matches: &searchC.Matches{},
				},
			}

			mockSearchItems[0].Matches.Description.Summary = &[]searchC.MatchDetails{
				{
					Value: "summary",
					Start: 1,
					End:   5,
				},
			}

			mockSearchItems[0].Matches.Description.Title = &[]searchC.MatchDetails{
				{
					Value: "title",
					Start: 6,
					End:   10,
				},
			}

			mockSearchItems[0].Matches.Description.Edition = &[]searchC.MatchDetails{
				{
					Value: "edition",
					Start: 11,
					End:   15,
				},
			}

			mockSearchItems[0].Matches.Description.MetaDescription = &[]searchC.MatchDetails{
				{
					Value: "meta_description",
					Start: 16,
					End:   20,
				},
			}

			mockSearchItems[0].Matches.Description.Keywords = &[]searchC.MatchDetails{
				{
					Value: "keywords",
					Start: 21,
					End:   25,
				},
			}

			mockSearchItems[0].Matches.Description.DatasetID = &[]searchC.MatchDetails{
				{
					Value: "dataset_id",
					Start: 26,
					End:   30,
				},
			}

			So(mockSearchResponse, ShouldResemble, searchC.Response{
				Count:        1,
				ContentTypes: mockSearchContentTypes,
				Items:        mockSearchItems,
				Topics:       mockSearchTopics,
			})

		})

		Convey("And return no error", func() {
			So(err, ShouldBeNil)
		})
	})
}
