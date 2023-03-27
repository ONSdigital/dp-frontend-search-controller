package data

import (
	"context"
	"net/url"
	"strings"

	searchModels "github.com/ONSdigital/dp-search-api/models"
)

type Dimensions struct {
	LocaliseKeyName string `json:"localise_key"`
	Count           int    `json:"count"`
	Type            string `json:"string"`
	Query           string `json:"query"`
	ShowInWebUI     bool   `json:"show_in_web_ui"`
}

func reviewDimensionsFilters(ctx context.Context, urlQuery url.Values, validatedQueryParams *SearchURLParams) error {
	dimensionFilters := urlQuery.Get("dimensions")
	dimensions := strings.Split(dimensionFilters, ",")
	validatedDimensionFilters := []string{}

	for i := range dimensions {
		dimensionsQuery := dimensions[i]
		if dimensionsQuery == "" {
			continue
		}

		validatedDimensionFilters = append(validatedDimensionFilters, dimensions[i])
	}
	validatedQueryParams.DimensionsFilter = strings.Join(validatedDimensionFilters, ",")
	return nil
}

func GetDimensions(countResp *searchModels.SearchResponse) (dimensions []Dimensions) {
	for _, dimension := range countResp.Dimensions {
		if len(dimension.Label) > 0 && len(dimension.Type) > 0 {
			dimensions = append(dimensions, Dimensions{
				/*
				* TODO - Get translations
				 */
				LocaliseKeyName: dimension.Label,
				Count:           dimension.Count,
				Type:            dimension.Type,
				ShowInWebUI:     true,
			})
		}
	}
	return
}
