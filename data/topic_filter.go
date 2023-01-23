package data

import (
	"context"
	"net/url"
	"sort"
	"strings"

	errs "github.com/ONSdigital/dp-frontend-search-controller/apperrors"
	"github.com/ONSdigital/dp-frontend-search-controller/cache"
	searchModels "github.com/ONSdigital/dp-search-api/models"
	"github.com/ONSdigital/log.go/v2/log"
)

// Topic represents a topic filter on the search page
type Topic struct {
	Count              int        `json:"count"`
	DistinctTopicCount int        `json:"distinct_topics_count"`
	LocaliseKeyName    string     `json:"localise_key"`
	Query              string     `json:"query"`
	ShowInWebUI        bool       `json:"show_in_web_ui"`
	Subtopics          []Subtopic `json:"subtopics"`
}

// Subtopic represents a subtopic filter on the search page
type Subtopic struct {
	Count           int    `json:"count"`
	LocaliseKeyName string `json:"localise_key"`
	Query           string `json:"query"`
	ShowInWebUI     bool   `json:"show_in_web_ui"`
}

// GetTopicCategories returns the topic filters to be displayed on the search page.
// Please note that only census topic filter is being returned
func GetTopics(censusTopicCache *cache.Topic, countResp *searchModels.SearchResponse) []Topic {
	var cachedSubtopics []cache.Subtopic
	if censusTopicCache != nil || censusTopicCache.List != nil {
		cachedSubtopics = censusTopicCache.List.GetSubtopics(censusTopicCache.LocaliseKeyName)
	}

	subtopics := make([]Subtopic, 0, len(cachedSubtopics))
	for i := range cachedSubtopics {
		// Do not add census topic to subtopics
		if cachedSubtopics[i].ID == censusTopicCache.ID {
			continue
		}

		subtopics = append(subtopics, Subtopic{
			LocaliseKeyName: cachedSubtopics[i].LocaliseKeyName,
			Query:           cachedSubtopics[i].ID,
			ShowInWebUI:     true,
		})
	}

	// Order subtopics alphabetically
	sort.Slice(subtopics, func(i, j int) bool {
		return subtopics[i].LocaliseKeyName < subtopics[j].LocaliseKeyName
	})

	censusTopic := Topic{
		LocaliseKeyName: censusTopicCache.LocaliseKeyName,
		Query:           censusTopicCache.ID,
		ShowInWebUI:     true,
		Subtopics:       subtopics,
	}

	// if censusTopicCache has not been updated, don't show census topic filter in web UI
	if censusTopicCache.LocaliseKeyName != "" {
		censusTopic = addTopicCounts(censusTopic, countResp)
	} else {
		censusTopic.ShowInWebUI = false
	}

	return []Topic{censusTopic}
}

func addTopicCounts(censusTopic Topic, countResp *searchModels.SearchResponse) Topic {
	for i := range countResp.Topics {
		if censusTopic.Query == countResp.Topics[i].Type {
			censusTopic.Count = countResp.Topics[i].Count
			continue
		}

		for j := range censusTopic.Subtopics {
			if censusTopic.Subtopics[j].Query == countResp.Topics[i].Type {
				censusTopic.Subtopics[j].Count = countResp.Topics[i].Count
				continue
			}
		}
	}

	censusTopic.DistinctTopicCount = countResp.DistinctTopicCount

	return censusTopic
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

		if ok := censusTopicCache.List.CheckTopicIDExists(topicFilterQuery); !ok {
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
