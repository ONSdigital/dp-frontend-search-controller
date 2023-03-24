package data

import (
	"testing"

	searchModels "github.com/ONSdigital/dp-search-api/models"
	. "github.com/smartystreets/goconvey/convey"
)

func TestUnitGetPoputationTypes(t *testing.T) {
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
			{Type: "Ethenticity", Count: 1},
		},
		PopulationType: []searchModels.FilterCount{
			{Type: "Usual Residents", Count: 1},
		},
	}
	populationTypes := GetPopulationTypes(searchResponseMock)
	Convey("Given search result includes Popultation Types ", t, func() {
		Convey("Check we can map them", func() {
			So(populationTypes[0].Type, ShouldEqual, searchResponseMock.Items[0].PopulationType)
		})
	})
}
