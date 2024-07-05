package public

import (
	"context"
	"errors"
	"github.com/ONSdigital/dp-frontend-search-controller/cache"
	"github.com/ONSdigital/dp-topic-api/models"
	topicCli "github.com/ONSdigital/dp-topic-api/sdk"
	"github.com/ONSdigital/log.go/v2/log"
	"net/http"
)

// UpdateCensusTopic is a function to update the census topic cache in web (public) mode.
// This function talks to the dp-topic-api via its public endpoints to retrieve the census topic and its subtopic ids
// The data returned by the dp-topic-api is of type *models.PublicSubtopics which is then transformed to *cache.Topic in this function for the controller
// If an error has occurred, this is captured in log.Error and then an empty census topic is returned
func UpdateCensusTopic(ctx context.Context, topicClient topicCli.Clienter) func() *cache.Topic {
	return func() *cache.Topic {
		// get root topics from dp-topic-api
		rootTopics, err := topicClient.GetRootTopicsPublic(ctx, topicCli.Headers{})
		if err != nil {
			log.Error(ctx, "failed to get public census root topics from topic-api", err)
			return cache.GetEmptyCensusTopic()
		}

		// dereference root topics items to allow ranging through them
		if rootTopics.PublicItems == nil {
			err := errors.New("census root topic public items is nil")
			log.Error(ctx, "failed to dereference public census root topics items pointer", err)
			return cache.GetEmptyCensusTopic()
		}
		rootTopicItems := *rootTopics.PublicItems

		var censusTopicCache *cache.Topic

		// go through each root topic, find census topic and gets its data for caching which includes subtopic ids
		for i := range rootTopicItems {
			if rootTopicItems[i].ID == cache.CensusTopicID {
				censusTopicCache = getRootTopicCachePublic(ctx, topicClient, rootTopicItems[i])
				break
			}
		}

		if censusTopicCache == nil {
			err := errors.New("census root topic not found")
			log.Error(ctx, "failed to get census topic to cache", err)
			return cache.GetEmptyCensusTopic()
		}
		return censusTopicCache
	}
}

// UpdateDataTopic is a function to update the data topic cache in web (public) mode.
// This function talks to the dp-topic-api via its public endpoints to retrieve the root data topic and its subtopic ids
// The data returned by the dp-topic-api is of type *models.Topic which is then transformed to *cache.Topic in this function for the controller
// If an error has occurred, this is captured in log.Error and then an empty data topic is returned
func UpdateDataTopic(ctx context.Context, topicClient topicCli.Clienter) func() *cache.Topic {
	return func() *cache.Topic {
		processedTopics := make(map[string]struct{})

		// get root topics from dp-topic-api
		rootTopics, err := topicClient.GetRootTopicsPublic(ctx, topicCli.Headers{})
		if err != nil {
			log.Error(ctx, "failed to get root data topics from topic-api", err)
			return cache.GetEmptyTopic()
		}

		// dereference root topics items to allow ranging through them
		var rootTopicItems []models.Topic
		if rootTopics.PublicItems != nil {
			rootTopicItems = *rootTopics.PublicItems
		} else {
			err := errors.New("root data topic public items is nil")
			log.Error(ctx, "failed to dereference root data topics items pointer", err)
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
			processTopic(ctx, topicClient, rootTopicItems[i].ID, dataTopicCache, processedTopics, "", 0)
		}

		// Check if any data topics were found
		if len(dataTopicCache.List.GetSubtopics()) == 0 {
			err := errors.New("data root topic found, but no subtopics were returned")
			log.Error(ctx, "No public topics loaded into cache - data root topic found, but no subtopics were returned", err)
			return cache.GetEmptyTopic()
		}
		return dataTopicCache
	}
}

func processTopic(ctx context.Context, topicClient topicCli.Clienter, topicID string, dataTopicCache *cache.Topic, processedTopics map[string]struct{}, parentTopicID string, depth int) {
	log.Info(ctx, "processing public topic", log.Data{
		"topic_id": topicID,
		"depth":    depth,
	})

	// Check if the topic has already been processed
	if _, exists := processedTopics[topicID]; exists {
		err := errors.New("topic already processed")
		log.Error(ctx, "skipping already processed public topic", err, log.Data{
			"topic_id": topicID,
			"depth":    depth,
		})
		return
	}

	// Get the topic details from the topic client
	dataTopic, err := topicClient.GetTopicPublic(ctx, topicCli.Headers{}, topicID)
	if err != nil {
		log.Error(ctx, "failed to get public topic details from topic-api", err, log.Data{
			"topic_id": topicID,
			"depth":    depth,
		})
		return
	}

	if dataTopic != nil {
		// Initialize subtopic list for the current topic if it doesn't exist
		subtopic := mapTopicModelToCache(*dataTopic, parentTopicID)

		// Add the current topic to the dataTopicCache's List
		dataTopicCache.List.AppendSubtopicID(dataTopic.Slug, subtopic)

		// Mark this topic as processed
		processedTopics[topicID] = struct{}{}

		// Process each subtopic recursively
		if dataTopic.SubtopicIds != nil {
			for _, subTopicID := range *dataTopic.SubtopicIds {
				processTopic(ctx, topicClient, subTopicID, dataTopicCache, processedTopics, topicID, depth+1)
			}
		}
	}
}

func mapTopicModelToCache(topic models.Topic, parentID string) cache.Subtopic {
	return cache.Subtopic{
		ID:              topic.ID,
		LocaliseKeyName: topic.Title,
		ParentID:        parentID,
		Slug:            topic.Slug,
	}
}

func getRootTopicCachePublic(ctx context.Context, topicClient topicCli.Clienter, rootTopic models.Topic) *cache.Topic {
	rootTopicCache := &cache.Topic{
		ID:              rootTopic.ID,
		Slug:            rootTopic.Slug,
		LocaliseKeyName: rootTopic.Title,
		ReleaseDate:     rootTopic.ReleaseDate,
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

	processSubtopicsPublic(ctx, subtopicsIDMap, topicClient, rootTopic.ID, processedTopics, 0)

	rootTopicCache.List = subtopicsIDMap
	rootTopicCache.Query = subtopicsIDMap.GetSubtopicsIDsQuery()

	return rootTopicCache
}

func processSubtopicsPublic(ctx context.Context, subtopicsIDMap *cache.Subtopics, topicClient topicCli.Clienter, topLevelTopicID string, processedTopics map[string]struct{}, depth int) {
	log.Info(ctx, "Processing public census sub-topic at depth", log.Data{
		"topic_id": topLevelTopicID,
		"depth":    depth,
	})

	// Check if this topic has already been processed
	if _, exists := processedTopics[topLevelTopicID]; exists {
		err := errors.New("topic already processed")
		log.Error(ctx, "skipping already processed public sub-topic", err, log.Data{
			"topic_id": topLevelTopicID,
			"depth":    depth,
		})
		return
	}

	// Mark this topic as processed
	processedTopics[topLevelTopicID] = struct{}{}

	// Get subtopics from dp-topic-api
	subTopics, err := topicClient.GetSubtopicsPublic(ctx, topicCli.Headers{}, topLevelTopicID)
	if err != nil {
		if err.Error() != http.StatusText(http.StatusNotFound) {
			log.Error(ctx, "failed to get public subtopics from topic-api", err, log.Data{
				"topic_id": topLevelTopicID,
				"depth":    depth,
			})
		}

		// Stop as there are no subtopics items or failed to get subtopics
		return
	}

	// Dereference sub-topics items to allow ranging through them
	var subTopicItems []models.Topic
	if subTopics.PublicItems == nil {
		err := errors.New("items is nil")
		log.Error(ctx, "failed to dereference public sub-topics items pointer", err, log.Data{
			"topic_id": topLevelTopicID,
			"depth":    depth,
		})
		return
	}
	subTopicItems = *subTopics.PublicItems

	// Stop recursion if subTopicItems is empty
	if len(subTopicItems) == 0 {
		log.Info(ctx, "No public subtopics found", log.Data{
			"topic_id": topLevelTopicID,
			"depth":    depth,
		})
		return
	}

	// Process each subtopic item sequentially
	for _, subTopicItem := range subTopicItems {
		subtopic := cache.Subtopic{
			ID:              subTopicItem.ID,
			Slug:            subTopicItem.Slug,
			LocaliseKeyName: subTopicItem.Title,
			ReleaseDate:     subTopicItem.ReleaseDate,
		}

		subtopicsIDMap.AppendSubtopicID(subTopicItem.ID, subtopic)

		if subTopicItem.SubtopicIds == nil || len(*subTopicItem.SubtopicIds) == 0 {
			continue
		}

		// Recursively process subtopics of the subtopic
		processSubtopicsPublic(ctx, subtopicsIDMap, topicClient, subTopicItem.ID, processedTopics, depth+1)
	}
}
