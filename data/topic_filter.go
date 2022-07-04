package data

import (
	"context"
	"net/url"
	"strings"

	searchCli "github.com/ONSdigital/dp-api-clients-go/v2/site-search"
	errs "github.com/ONSdigital/dp-frontend-search-controller/apperrors"
	"github.com/ONSdigital/dp-frontend-search-controller/cache"
	"github.com/ONSdigital/log.go/v2/log"
)

// Topic represents a topic filter on the search page
type Topic struct {
	LocaliseKeyName string `json:"localise_key"`
	Count           int    `json:"count"`
	Query           string `json:"query"`
	ShowInWebUI     bool   `json:"show_in_web_ui"`
}

// GetTopicCategories returns the topic filters to be displayed on the search page.
// Please note that only census topic filter is being returned
func GetTopicCategories(censusTopicCache *cache.Topic, countResp searchCli.Response) []Topic {
	censusTopic := Topic{
		LocaliseKeyName: censusTopicCache.LocaliseKeyName,
		Query:           censusTopicCache.Query,
		ShowInWebUI:     true,
	}

	censusTopic.Count = getCensusTopicCount(censusTopicCache, countResp)

	return []Topic{censusTopic}
}

func getCensusTopicCount(censusTopicCache *cache.Topic, countResp searchCli.Response) (count int) {
	for i := range countResp.Topics {
		if censusTopicCache.ID == countResp.Topics[i].Type {
			count = countResp.Topics[i].Count
			return
		}
	}
	return
}

// TODO extend default topics with list of topics
var defaultTopics = ""

// reviewTopicFilters retrieves subtopic ids from query, checks if they are one of the census subtopics, and updates validatedQueryParams
func reviewTopicFilters(ctx context.Context, urlQuery url.Values, validatedQueryParams *SearchURLParams, censusTopicCache *cache.Topic) error {
	topicFilters := urlQuery.Get("topics")
	topicIDs := strings.Split(topicFilters, ",")

	validatedTopicFilters := []string{}

	for i := range topicIDs {

		topicFilterQuery := strings.ToLower(topicIDs[i])

		if topicFilterQuery == "" {
			continue
		}

		// checks if topic id is a subtopic id of the census topic
		found := censusTopicCache.List.Get(topicFilterQuery)
		if !found {
			err := errs.ErrFilterNotFound
			logData := log.Data{"subtopic id not found": topicFilterQuery}
			log.Error(ctx, "failed to find subtopic id in census topic data", err, logData)
			return err
		}

		validatedTopicFilters = append(validatedTopicFilters, topicIDs[i])
	}

	validatedQueryParams.TopicFilter = strings.Join(validatedTopicFilters, ",")

	return nil
}

// updateQueryWithAPITopics retrieves and adds all available subtopics which is related to the search filter given by the user
func updateQueryWithAPITopics(apiQuery url.Values) {
	topics := apiQuery["topics"]

	if len(topics) > 0 {
		// subTopics := getListOfSubTopics(topics)
		apiQuery.Set("topics", strings.Join(topics, ","))
	} else {
		apiQuery.Set("topics", defaultTopics)
	}
}
