package private

import (
	"context"
	"errors"
	"net/http"

	"github.com/ONSdigital/dp-frontend-search-controller/cache"
	"github.com/ONSdigital/dp-topic-api/models"
	topicCli "github.com/ONSdigital/dp-topic-api/sdk"
	"github.com/ONSdigital/log.go/v2/log"
)

// UpdateCensusTopic is a function to update the census topic cache in publishing (private) mode.
// This function talks to the dp-topic-api via its private endpoints to retrieve the census topic and its subtopic ids
// The data returned by the dp-topic-api is of type *models.PrivateSubtopics which is then transformed in this function for the controller
// If an error has occurred, this is captured in log.Error and then an empty census topic is returned
func UpdateCensusTopic(ctx context.Context, serviceAuthToken string, topicClient topicCli.Clienter) func() *cache.Topic {
	return func() *cache.Topic {
		// get root topics from dp-topic-api
		rootTopics, err := topicClient.GetRootTopicsPrivate(ctx, topicCli.Headers{ServiceAuthToken: serviceAuthToken})
		if err != nil {
			log.Error(ctx, "failed to get private census root topics from topic-api", err)
			return cache.GetEmptyCensusTopic()
		}

		// dereference root topics items to allow ranging through them
		var rootTopicItems []models.TopicResponse
		if rootTopics.PrivateItems != nil {
			rootTopicItems = *rootTopics.PrivateItems
		} else {
			err := errors.New("census root topic private items is nil")
			log.Error(ctx, "failed to dereference private census root topics items pointer", err)
			return cache.GetEmptyCensusTopic()
		}

		var censusTopicCache *cache.Topic

		// go through each root topic, find census topic and gets its data for caching which includes subtopic ids
		for i := range rootTopicItems {
			if rootTopicItems[i].Current.ID == cache.CensusTopicID {
				censusTopicCache = getRootTopicCachePrivate(ctx, serviceAuthToken, topicClient, *rootTopicItems[i].Current)
				break
			}
		}

		if censusTopicCache == nil {
			err := errors.New("not found")
			log.Error(ctx, "failed to get census root topics to cache", err)
			return cache.GetEmptyCensusTopic()
		}

		return censusTopicCache
	}
}

// UpdateDataTopicCache is a function to update the data topic cache in publishing (private) mode.
// This function talks to the dp-topic-api via its private endpoints to retrieve the root topic and its subtopic ids
// The data returned by the dp-topic-api is of type *models.PrivateSubtopics which is then transformed in this function for the controller
// If an error has occurred, this is captured in log.Error and then an empty topic is returned
func UpdateDataTopicCache(ctx context.Context, serviceAuthToken string, topicClient topicCli.Clienter) func() *cache.Topic {
	return func() *cache.Topic {
		processedTopics := make(map[string]struct{})

		// get root topics from dp-topic-api
		rootTopics, err := topicClient.GetRootTopicsPrivate(ctx, topicCli.Headers{ServiceAuthToken: serviceAuthToken})
		if err != nil {
			log.Error(ctx, "failed to get data root topics from topic-api", err)
			return cache.GetEmptyTopic()
		}

		// deference root topics items to allow ranging through them
		var rootTopicItems []models.TopicResponse
		if rootTopics.PrivateItems != nil {
			rootTopicItems = *rootTopics.PrivateItems
		} else {
			err := errors.New("data root topic private items is nil")
			log.Error(ctx, "failed to dereference data root topics items pointer", err)
			return cache.GetEmptyTopic()
		}

		// Initialize dataTopicCache
		dataTopicCache := &cache.Topic{
			ID:              cache.DataTopicCacheKey,
			LocaliseKeyName: "Root",
			List:            cache.NewSubTopicsMap(),
		}

		// recursively process root topics and their subtopics
		for i := range rootTopicItems {
			processTopic(ctx, serviceAuthToken, topicClient, rootTopicItems[i].ID, dataTopicCache, processedTopics, "", 0)
		}

		// Check if any data topics were found
		if len(dataTopicCache.List.GetSubtopics()) == 0 {
			err := errors.New("data root topic found, but no subtopics were returned")
			log.Error(ctx, "No private topics loaded into cache - data root topic found, but no subtopics were returned", err)
			return cache.GetEmptyTopic()
		}
		return dataTopicCache
	}
}

func processTopic(ctx context.Context, serviceAuthToken string, topicClient topicCli.Clienter, topicID string, dataTopicCache *cache.Topic, processedTopics map[string]struct{}, parentTopicID string, depth int) {
	log.Info(ctx, "processing private topic", log.Data{
		"topic_id": topicID,
		"depth":    depth,
	})

	// Check if the topic has already been processed
	if _, exists := processedTopics[topicID]; exists {
		err := errors.New("topic already processed")
		log.Error(ctx, "skipping already processed private topic", err, log.Data{
			"topic_id": topicID,
			"depth":    depth,
		})
		return
	}

	// Get the topic details from the topic client
	dataTopic, err := topicClient.GetTopicPrivate(ctx, topicCli.Headers{ServiceAuthToken: serviceAuthToken}, topicID)
	if err != nil {
		log.Error(ctx, "failed to get private topic details from topic-api", err, log.Data{
			"topic_id": topicID,
			"depth":    depth,
		})
		return
	}

	if dataTopic != nil {
		// Initialize subtopic list for the current topic if it doesn't exist
		subtopic := mapTopicModelToCache(*dataTopic.Current, parentTopicID)

		// Add the current topic to the dataTopicCache's List
		dataTopicCache.List.AppendSubtopicID(dataTopic.Current.Slug, subtopic)

		// Mark this topic as processed
		processedTopics[topicID] = struct{}{}

		// Process each subtopic recursively
		if dataTopic.Current.SubtopicIds != nil {
			for _, subTopicID := range *dataTopic.Current.SubtopicIds {
				processTopic(ctx, serviceAuthToken, topicClient, subTopicID, dataTopicCache, processedTopics, topicID, depth+1)
			}
		}
	}
}

func mapTopicModelToCache(topic models.Topic, parentID string) cache.Subtopic {
	return cache.Subtopic{
		ID:              topic.ID,
		Slug:            topic.Slug,
		LocaliseKeyName: topic.Title,
		ReleaseDate:     topic.ReleaseDate,
		ParentID:        parentID,
	}
}

func getRootTopicCachePrivate(ctx context.Context, serviceAuthToken string, topicClient topicCli.Clienter, rootTopic models.Topic) *cache.Topic {
	rootTopicCache := &cache.Topic{
		ID:              rootTopic.ID,
		Slug:            rootTopic.Slug,
		LocaliseKeyName: rootTopic.Title,
	}

	subtopic := cache.Subtopic{
		ID:              rootTopic.ID,
		Slug:            rootTopic.Slug,
		LocaliseKeyName: rootTopic.Title,
		ReleaseDate:     rootTopic.ReleaseDate,
	}

	subtopicsIDMap := cache.NewSubTopicsMap()
	subtopicsIDMap.AppendSubtopicID(rootTopic.ID, subtopic)

	processedTopics := make(map[string]struct{})

	processSubtopicsPrivate(ctx, serviceAuthToken, subtopicsIDMap, topicClient, rootTopic.ID, processedTopics, 0)

	rootTopicCache.List = subtopicsIDMap
	rootTopicCache.Query = subtopicsIDMap.GetSubtopicsIDsQuery()

	return rootTopicCache
}

func processSubtopicsPrivate(ctx context.Context, serviceAuthToken string, subtopicsIDMap *cache.Subtopics, topicClient topicCli.Clienter, topLevelTopicID string, processedTopics map[string]struct{}, depth int) {
	log.Info(ctx, "Processing private census sub-topic at depth", log.Data{
		"topic_id": topLevelTopicID,
		"depth":    depth,
	})

	// Check if this topic has already been processed
	if _, exists := processedTopics[topLevelTopicID]; exists {
		err := errors.New("topic already processed")
		log.Error(ctx, "skipping already processed private sub-topic", err, log.Data{
			"topic_id": topLevelTopicID,
			"depth":    depth,
		})
		return
	}

	// Mark this topic as processed
	processedTopics[topLevelTopicID] = struct{}{}

	// Get subtopics from dp-topic-api
	subTopics, err := topicClient.GetSubtopicsPrivate(ctx, topicCli.Headers{ServiceAuthToken: serviceAuthToken}, topLevelTopicID)
	if err != nil {
		if err.Error() != http.StatusText(http.StatusNotFound) {
			log.Error(ctx, "failed to get private subtopics from topic-api", err, log.Data{
				"topic_id": topLevelTopicID,
				"depth":    depth,
			})
		}

		// Stop as there are no subtopics items or failed to get subtopics
		return
	}

	// Dereference sub-topics items to allow ranging through them
	var subTopicItems []models.TopicResponse
	if subTopics.PrivateItems == nil {
		err := errors.New("items is nil")
		log.Error(ctx, "failed to dereference private sub-topics items pointer", err, log.Data{
			"topic_id": topLevelTopicID,
			"depth":    depth,
		})
		return
	}
	subTopicItems = *subTopics.PrivateItems

	// Stop recursion if subTopicItems is empty
	if len(subTopicItems) == 0 {
		log.Info(ctx, "No private subtopics found", log.Data{
			"topic_id": topLevelTopicID,
			"depth":    depth,
		})
		return
	}

	// Process each subtopic item sequentially
	for _, subTopicItem := range subTopicItems {
		subtopic := cache.Subtopic{
			ID:              subTopicItem.ID,
			Slug:            subTopicItem.Current.Slug,
			LocaliseKeyName: subTopicItem.Current.Title,
			ReleaseDate:     subTopicItem.Current.ReleaseDate,
		}

		subtopicsIDMap.AppendSubtopicID(subTopicItem.ID, subtopic)

		if subTopicItem.Current.SubtopicIds == nil || len(*subTopicItem.Current.SubtopicIds) == 0 {
			continue
		}

		// Recursively process subtopics of the subtopic
		processSubtopicsPrivate(ctx, serviceAuthToken, subtopicsIDMap, topicClient, subTopicItem.ID, processedTopics, depth+1)
	}
}
