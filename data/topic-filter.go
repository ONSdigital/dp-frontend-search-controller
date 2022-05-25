package data

import (
	"context"
	"net/url"
	"strings"

	errs "github.com/ONSdigital/dp-frontend-search-controller/apperrors"
	"github.com/ONSdigital/log.go/v2/log"
)

// Filter represents information of filters selected by user
type TopicFilter struct {
	Filter
}

// TopicCategory represents all the search categories in search page
type TopicCategory struct {
	LocaliseKeyName string  `json:"localise_key"`
	Count           int     `json:"count"`
	Topics          []Topic `json:"topics"`
}

// Topic represents the type of the search results and the number of results for each type
type Topic struct {
	LocaliseKeyName string   `json:"localise_key"`
	Count           int      `json:"count"`
	Group           string   `json:"group"`
	SubTopics       []string `json:"subtopics"`
	ShowInWebUI     bool     `json:"show_in_web_ui"`
}

// TODO extend default topics with list of topics
var defaultTopics = ""

var (
	// Categories represent the list of all search categories
	TopicCategories = []TopicCategory{Census}

	// Census - search information on census category
	Census = TopicCategory{
		LocaliseKeyName: "Census",
		Topics:          []Topic{DemographyAndMigration, Education, EthnicGroupNationalIdentityAndReligion, HealthDisabilityAndUnpaidCare, Housing, LabourMarketAndTravelToWork, SexualOrientationAndGenderIdentity, Veterans},
	}

	// Bulletin - Search information specific for statistical bulletins
	DemographyAndMigration = Topic{
		LocaliseKeyName: "DemographyAndMigration",
		Group:           "demography_and_migration",
		SubTopics:       []string{"demography_and_migration"},
		ShowInWebUI:     true,
	}

	// Education - Search information specific for Education
	Education = Topic{
		LocaliseKeyName: "Education",
		Group:           "education",
		SubTopics:       []string{"education"},
		ShowInWebUI:     true,
	}

	// EthnicGroupNationalIdentityAndReligion - Search information specific for ethnic groups, national identity and religion
	EthnicGroupNationalIdentityAndReligion = Topic{
		LocaliseKeyName: "EthnicGroupNationalIdentityAndReligion",
		Group:           "ethnic_group_national_identity_and_religion",
		SubTopics:       []string{"ethnic_group_national_identity_and_religion"},
		ShowInWebUI:     true,
	}

	// HealthDisabilityAndUnpaidCare - Search information specific for health, disabilities and unpaid care
	HealthDisabilityAndUnpaidCare = Topic{
		LocaliseKeyName: "HealthDisabilityAndUnpaidCare",
		Group:           "health_disability_and_unpaid_care",
		SubTopics:       []string{"health_disability_and_unpaid_care"},
		ShowInWebUI:     true,
	}

	// Housing - Search information specific for Housing
	Housing = Topic{
		LocaliseKeyName: "Housing",
		Group:           "housing",
		SubTopics:       []string{"housing"},
		ShowInWebUI:     true,
	}

	// LabourMarketAndTravelToWork - Search information specific for labour market and travel to work
	LabourMarketAndTravelToWork = Topic{
		LocaliseKeyName: "LabourMarketAndTravelToWork",
		Group:           "labour_market_and_travel_to_work",
		SubTopics:       []string{"labour_market_and_travel_to_work"},
		ShowInWebUI:     true,
	}

	// SexualOrientationAndGenderIdentity - Search information specific for sexual orientation and gender identity
	SexualOrientationAndGenderIdentity = Topic{
		LocaliseKeyName: "SexualOrientationAndGenderIdentity",
		Group:           "sexual_orientation_and_gender_identity",
		SubTopics:       []string{"sexual_orientation_and_gender_identity"},
		ShowInWebUI:     true,
	}

	// Veterans - Search information specific for Veterans
	Veterans = Topic{
		LocaliseKeyName: "Veteran",
		Group:           "veterans",
		SubTopics:       []string{"veterans"},
		ShowInWebUI:     true,
	}

	// topicFilterOptions contains all the possible filter available on the search page
	topicFilterOptions = map[string]Topic{
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
	topicFilters := urlQuery.Get("topics")

	topics := strings.Split(topicFilters, ",")

	for i := range topics {

		topicFilterQuery := strings.ToLower(topics[i])

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

	// To get a different reference of Topic - deep copy
	for i, topicCategory := range topicCategories {
		topicCategories[i].Topics = []Topic{}
		topicCategories[i].Topics = append(topicCategories[i].Topics, TopicCategories[i].Topics...)

		// To get a different reference of SubTypes - deep copy
		for j := range topicCategory.Topics {
			topicCategories[i].Topics[j].SubTopics = []string{}
			topicCategories[i].Topics[j].SubTopics = append(topicCategories[i].Topics[j].SubTopics, TopicCategories[i].Topics[j].SubTopics...)
		}
	}

	return topicCategories
}

// updateQueryWithAPITopics retrieves and adds all available subtopics which is related to the search filter given by the user
func updateQueryWithAPITopics(apiQuery url.Values) {
	topics := apiQuery["topics"]

	if len(topics) > 0 {
		subTopics := getListOfSubTopics(topics)

		apiQuery.Set("topics", strings.Join(subTopics, ","))
	} else {
		apiQuery.Set("topics", defaultTopics)
	}
}

// updateQueryWithAPITopicFilters adds all available subtopic filters to the API query
func getListOfSubTopics(topics []string) []string {
	var subTopics []string

	for i := range topics {
		topicFilterOption := topicFilterOptions[topics[i]]
		subTopics = append(subTopics, topicFilterOption.SubTopics...)
	}

	return subTopics
}

// GetTopicGroupLocaliseKey gets the localise key of the group type of the search result to be displayed
func GetTopicGroupLocaliseKey(resultType string) string {
	for i := range topicFilterOptions {
		for j := range topicFilterOptions[i].SubTopics {
			if resultType == topicFilterOptions[i].SubTopics[j] {
				return topicFilterOptions[i].LocaliseKeyName
			}
		}
	}
	return ""
}