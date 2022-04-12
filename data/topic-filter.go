package data

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	errs "github.com/ONSdigital/dp-frontend-search-controller/apperrors"
	"github.com/ONSdigital/log.go/v2/log"
)

// Filter represents information of filters selected by user
type TopicFilter struct {
	Query           []string `json:"query,omitempty"`
	LocaliseKeyName []string `json:"localise_key,omitempty"`
}

// TopicCategory represents all the search categories in search page
type TopicCategory struct {
	LocaliseKeyName string             `json:"localise_key"`
	Count           int                `json:"count"`
	ContentTypes    []TopicContentType `json:"content_types"`
}

// ContentType represents the type of the search results and the number of results for each type
type TopicContentType struct {
	LocaliseKeyName string   `json:"localise_key"`
	Count           int      `json:"count"`
	Group           string   `json:"group"`
	Types           []string `json:"types"`
	ShowInWebUI     bool     `json:"show_in_web_ui"`
}

var defaultTopicContentTypes = "article," +
	"article_download," +
	"bulletin," +
	"compendium_landing_page," +
	"dataset_landing_page," +
	"product_page," +
	"static_adhoc," +
	"static_article," +
	"static_foi," +
	"static_landing_page," +
	"static_methodology," +
	"static_methodology_download," +
	"static_page," +
	"static_qmi," +
	"timeseries," +
	"timeseries_dataset"

var (
	// Categories represent the list of all search categories
	TopicCategories = []TopicCategory{Census}

	// Census - search information on census category
	Census = TopicCategory{
		LocaliseKeyName: "Census",
		ContentTypes:    []TopicContentType{DemographyAndMigration, Education, EthnicGroupNationalIdentityAndReligion, HealthDisabilityAndUnpaidCare, Housing, LabourMarketAndTravelToWork, SexualOrientationAndGenderIdentity, Veterans},
	}

	// Bulletin - Search information specific for statistical bulletins
	DemographyAndMigration = TopicContentType{
		LocaliseKeyName: "DemographyAndMigration",
		Group:           "demography_and_migration",
		Types:           []string{"demography_and_migration"},
		ShowInWebUI:     true,
	}

	// Education - Search information specific for Education
	Education = TopicContentType{
		LocaliseKeyName: "Education",
		Group:           "education",
		Types:           []string{"education"},
		ShowInWebUI:     true,
	}

	// EthnicGroupNationalIdentityAndReligion - Search information specific for ethnic groups, national identity and religion
	EthnicGroupNationalIdentityAndReligion = TopicContentType{
		LocaliseKeyName: "EthnicGroupNationalIdentityAndReligion",
		Group:           "ethnic_group_national_identity_and_religion",
		Types:           []string{"ethnic_group_national_identity_and_religion"},
		ShowInWebUI:     true,
	}

	// HealthDisabilityAndUnpaidCare - Search information specific for health, disabilities and unpaid care
	HealthDisabilityAndUnpaidCare = TopicContentType{
		LocaliseKeyName: "HealthDisabilityAndUnpaidCare",
		Group:           "health_disability_and_unpaid_care",
		Types:           []string{"health_disability_and_unpaid_care"},
		ShowInWebUI:     true,
	}

	// Housing - Search information specific for Housing
	Housing = TopicContentType{
		LocaliseKeyName: "Housing",
		Group:           "housing",
		Types:           []string{"housing"},
		ShowInWebUI:     true,
	}

	// LabourMarketAndTravelToWork - Search information specific for labour market and travel to work
	LabourMarketAndTravelToWork = TopicContentType{
		LocaliseKeyName: "LabourMarketAndTravelToWork",
		Group:           "labour_market_and_travel_to_work",
		Types:           []string{"labour_market_and_travel_to_work"},
		ShowInWebUI:     true,
	}

	// SexualOrientationAndGenderIdentity - Search information specific for sexual orientation and gender identity
	SexualOrientationAndGenderIdentity = TopicContentType{
		LocaliseKeyName: "SexualOrientationAndGenderIdentity",
		Group:           "sexual_orientation_and_gender_identity",
		Types:           []string{"sexual_orientation_and_gender_identity"},
		ShowInWebUI:     true,
	}

	// Veterans - Search information specific for Veterans
	Veterans = TopicContentType{
		LocaliseKeyName: "Veteran",
		Group:           "veterans",
		Types:           []string{"veterans"},
		ShowInWebUI:     true,
	}

	// topicFilterOptions contains all the possible filter available on the search page
	topicFilterOptions = map[string]TopicContentType{
		DemographyAndMigration.Group:                 DemographyAndMigration,
		Education.Group:                              Education,
		EthnicGroupNationalIdentityAndReligion.Group: EthnicGroupNationalIdentityAndReligion,
		HealthDisabilityAndUnpaidCare.Group:          HealthDisabilityAndUnpaidCare,
		Housing.Group:                                Housing,
		LabourMarketAndTravelToWork.Group:            LabourMarketAndTravelToWork,
		SexualOrientationAndGenderIdentity.Group:     SexualOrientationAndGenderIdentity,
		Veterans.Group:                               Veterans,
	}
)

// reviewFilter retrieves filters from query, checks if they are one of the filter options, and updates validatedQueryParams
func reviewTopicFilters(ctx context.Context, urlQuery url.Values, validatedQueryParams *SearchURLParams) error {
	topicFiltersQuery := urlQuery["topic"]

	fmt.Println(string("––––––––––––––––––––––––––––––––––––––––––––topic-filter.go"))
	data, _ := json.Marshal(topicFiltersQuery)
	fmt.Println(string(data))
	fmt.Println(string("––––––––––––––––––––––––––––––––––––––––––––"))

	for _, topicFilterQuery := range topicFiltersQuery {

		topicFilterQuery = strings.ToLower(topicFilterQuery)

		if topicFilterQuery == "" {
			continue
		}

		topicFilter, found := topicFilterOptions[topicFilterQuery]

		if !found {
			err := errs.ErrFilterNotFound
			logData := log.Data{"topic filter not found": topicFilter}
			log.Error(ctx, "failed to find topic filter", err, logData)

			return err
		}

		validatedQueryParams.TopicFilter.Query = append(validatedQueryParams.TopicFilter.Query, topicFilter.Group)
		validatedQueryParams.TopicFilter.LocaliseKeyName = append(validatedQueryParams.TopicFilter.LocaliseKeyName, topicFilter.LocaliseKeyName)
	}

	return nil
}

// GetTopicCategories returns all the categories and its content types where all the count is set to zero
func GetTopicCategories() []TopicCategory {
	var topicCategories []TopicCategory
	topicCategories = append(topicCategories, TopicCategories...)

	// To get a different reference of TopicContentType - deep copy
	for i, topicCategory := range topicCategories {
		topicCategories[i].ContentTypes = []TopicContentType{}
		topicCategories[i].ContentTypes = append(topicCategories[i].ContentTypes, TopicCategories[i].ContentTypes...)

		// To get a different reference of SubTypes - deep copy
		for j := range topicCategory.ContentTypes {
			topicCategories[i].ContentTypes[j].Types = []string{}
			topicCategories[i].ContentTypes[j].Types = append(topicCategories[i].ContentTypes[j].Types, TopicCategories[i].ContentTypes[j].Types...)
		}
	}

	return topicCategories
}

// updateQueryWithAPITopicFilters retrieves and adds all available sub filters which is related to the search filter given by the user
func updateQueryWithAPITopicFilters(apiQuery url.Values) {
	filters := apiQuery["content_type"]

	if len(filters) > 0 {
		subFilters := getTopicSubFilters(filters)

		apiQuery.Set("content_type", strings.Join(subFilters, ","))
	} else {
		apiQuery.Set("content_type", defaultTopicContentTypes)
	}
}

// getTopicSubFilters gets all available sub filters which is related to the search filter given by the user
func getTopicSubFilters(filters []string) []string {
	var subFilters = make([]string, 0)

	for _, filter := range filters {
		subFilter := topicFilterOptions[filter]
		subFilters = append(subFilters, subFilter.Types...)
	}

	return subFilters
}

// GetTopicGroupLocaliseKey gets the localise key of the group type of the search result to be displayed
func GetTopicGroupLocaliseKey(resultType string) string {
	for _, filterOption := range topicFilterOptions {
		for _, optionType := range filterOption.Types {
			if resultType == optionType {
				return filterOption.LocaliseKeyName
			}
		}
	}
	return ""
}
