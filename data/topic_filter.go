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
		Query:           censusTopicCache.ID,
		ShowInWebUI:     true,
	}

	// if censusTopicCache has not been updated, don't show census topic filter in web UI
	if censusTopicCache.LocaliseKeyName != "" {
		censusTopic.Count = getCensusTopicCount(censusTopicCache, countResp)
	} else {
		censusTopic.ShowInWebUI = false
	}

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

		found := censusTopicCache.List.Get(topicFilterQuery)
		if !found {
			err := errs.ErrTopicNotFound
			logData := log.Data{"subtopic id not found": topicFilterQuery}
			log.Error(ctx, "failed to find subtopic id in census topic data", err, logData)
			return err
		}

		validatedTopicFilters = append(validatedTopicFilters, topicIDs[i])
	}

	validatedQueryParams.TopicFilter = strings.Join(validatedTopicFilters, ",")

	return nil
}

// updateTopicsQueryForSearchAPI updates the topics query with subtopic ids if one of the topic is a root id
func updateTopicsQueryForSearchAPI(apiQuery url.Values, censusTopicCache *cache.Topic) {
	topicFilters := apiQuery.Get("topics")
	topicIDs := strings.Split(topicFilters, ",")

	rootAndSubtopics := []string{}

	for i := range topicIDs {
		// if topic id is root id of the census topic
		if topicIDs[i] == censusTopicCache.ID {
			// append topic root id and its subtopic ids
			rootAndSubtopics = append(rootAndSubtopics, censusTopicCache.Query)
			continue
		}

		rootAndSubtopics = append(rootAndSubtopics, topicIDs[i])
	}

	apiQuery.Set("topics", strings.Join(rootAndSubtopics, ","))
}
