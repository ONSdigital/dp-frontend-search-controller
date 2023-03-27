package data

import (
	"testing"

	searchModels "github.com/ONSdigital/dp-search-api/models"
	. "github.com/smartystreets/goconvey/convey"
)

func TestUnitGetDimensions(t *testing.T) {
	t.Parallel()

	searchResponseMock := &searchModels.SearchResponse{
		Count:              0,
		Took:               0,
		DistinctItemsCount: 0,
		Topics:             []searchModels.FilterCount{},
		ContentTypes:       []searchModels.FilterCount{},
		Items: []searchModels.Item{
			{
				Dimensions: []searchModels.ESDimensions{
					{
						Name:     "Ethnicity",
						Label:    "Ethnicity",
						RawLabel: "Ethnicity",
					},
				},
				PopulationType: "Usual Residents",
			},
		},
		Suggestions:         []string{},
		AdditionSuggestions: []string{},
		Dimensions: []searchModels.FilterCount{
			{Type: "ethnicity", Label: "Ethnicity", Count: 1},
		},
		PopulationType: []searchModels.FilterCount{
			{Type: "UR", Label: "Usual Residents", Count: 1},
		},
	}
	dimensions := GetDimensions(searchResponseMock)
	Convey("Given search result includes dimensions ", t, func() {
		Convey("Check we can map them", func() {
			So(dimensions[0].LocaliseKeyName, ShouldEqual, searchResponseMock.Items[0].Dimensions[0].RawLabel)
		})
	})
}
