package data

import (
	"net/url"
	"sort"
	"strings"

	searchModels "github.com/ONSdigital/dp-search-api/models"
)

type PopulationTypes struct {
	LocaliseKeyName string `json:"localise_key"`
	Count           int    `json:"count"`
	Type            string `json:"string"`
	Query           string `json:"query"`
	ShowInWebUI     bool   `json:"show_in_web_ui"`
}

func reviewPopulationTypeFilters(urlQuery url.Values, validatedQueryParams *SearchURLParams) error {
	populationTypeFilters := urlQuery.Get("population_types")
	populationTypes := strings.Split(populationTypeFilters, ",")
	validatedPopulationTypeFilters := []string{}

	for i := range populationTypes {
		populationTypesQuery := populationTypes[i]
		if populationTypesQuery == "" {
			continue
		}

		/*
		 * TODO - Check population type is valid
		 */

		validatedPopulationTypeFilters = append(validatedPopulationTypeFilters, populationTypes[i])
	}
	validatedQueryParams.PopulationTypeFilter = strings.Join(validatedPopulationTypeFilters, ",")
	return nil
}

func GetPopulationTypes(countResp *searchModels.SearchResponse) (populationTypes []PopulationTypes) {
	for _, populationType := range countResp.PopulationType {
		if len(populationType.Label) > 0 && len(populationType.Type) > 0 {
			populationTypes = append(populationTypes, PopulationTypes{
				/*
				* TODO - Get translations
				 */
				LocaliseKeyName: populationType.Label,
				Count:           populationType.Count,
				Type:            populationType.Type,
				ShowInWebUI:     true,
			})
		}
	}
	sort.Slice(populationTypes, func(i, j int) bool {
		return populationTypes[i].LocaliseKeyName < populationTypes[j].LocaliseKeyName
	})
	return
}
