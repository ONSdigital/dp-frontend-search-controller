package public

import (
	"context"
	"errors"
	"net/http"
	"sync"

	"github.com/ONSdigital/dp-frontend-search-controller/cache"
	"github.com/ONSdigital/dp-topic-api/models"
	topicCli "github.com/ONSdigital/dp-topic-api/sdk"
	"github.com/ONSdigital/log.go/v2/log"
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
			log.Error(ctx, "failed to get census root topics from topic-api", err)
			return cache.GetEmptyCensusTopic()
		}

		// dereference root topics items to allow ranging through them
		if rootTopics.PublicItems == nil {
			err := errors.New("census root topic public items is nil")
			log.Error(ctx, "failed to dereference census root topics items pointer", err)
			return cache.GetEmptyCensusTopic()
		}
		rootTopicItems := *rootTopics.PublicItems

		var censusTopicCache *cache.Topic

		// go through each root topic, find census topic and gets its data for caching which includes subtopic ids
		for i := range rootTopicItems {
			if rootTopicItems[i].ID == cache.CensusTopicID {
				subtopicsChan := make(chan models.Topic)

				censusTopicCache = getRootTopicCachePublic(ctx, subtopicsChan, topicClient, rootTopicItems[i])
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

// UpdateDataTopics is a function to update the data topic cache in web (public) mode.
// This function talks to the dp-topic-api via its public endpoints to retrieve the root data topic and its subtopic ids
// The data returned by the dp-topic-api is of type *models.Topic which is then transformed to *cache.Topic in this function for the controller
// If an error has occurred, this is captured in log.Error and then an empty data topic is returned
func UpdateDataTopics(ctx context.Context, topicClient topicCli.Clienter) func() []*cache.Topic {
	return func() []*cache.Topic {
		var topics []*cache.Topic
		processedTopics := make(map[string]bool)

		// get root topics from dp-topic-api
		rootTopics, err := topicClient.GetRootTopicsPublic(ctx, topicCli.Headers{})
		if err != nil {
			log.Error(ctx, "failed to get root data topics from topic-api", err)
			return []*cache.Topic{cache.GetEmptyTopic()}
		}

		// dereference root topics items to allow ranging through them
		var rootTopicItems []models.Topic
		if rootTopics.PublicItems != nil {
			rootTopicItems = *rootTopics.PublicItems
		} else {
			err := errors.New("root data topic public items is nil")
			log.Error(ctx, "failed to dereference root data topics items pointer", err)
			return []*cache.Topic{cache.GetEmptyTopic()}
		}

		// recursively process root topics and their subtopics
		for i := range rootTopicItems {
			processTopic(ctx, topicClient, rootTopicItems[i].ID, &topics, processedTopics, "")
		}

		// Check if any data topics were found
		if len(topics) == 0 {
			err := errors.New("data root topic found, but no subtopics were returned")
			log.Error(ctx, "No topics loaded into cache - data root topic found, but no subtopics were returned", err)
			return []*cache.Topic{cache.GetEmptyTopic()}
		}
		return topics
	}
}

func processTopic(ctx context.Context, topicClient topicCli.Clienter, topicID string, topics *[]*cache.Topic, processedTopics map[string]bool, parentTopicID string) {
	// Check if the topic is already processed
	if processedTopics[topicID] {
		return
	}

	// Get the topic details from the topic client
	dataTopic, err := topicClient.GetTopicPublic(ctx, topicCli.Headers{}, topicID)
	if err != nil {
		log.Error(ctx, "failed to get topic details from topic-api", err, log.Data{
			"Topic ID": topicID,
		})
		return
	}

	if dataTopic != nil {
		// Append the current topic to the list of topics
		*topics = append(*topics, mapTopicModelToCache(*dataTopic, parentTopicID))
		// Mark this topic as processed
		processedTopics[topicID] = true

		// Process each subtopic recursively
		if dataTopic.SubtopicIds != nil {
			for _, subTopicID := range *dataTopic.SubtopicIds {
				processTopic(ctx, topicClient, subTopicID, topics, processedTopics, topicID)
			}
		}
	}
}

func mapTopicModelToCache(topic models.Topic, parentID string) *cache.Topic {
	topicCache := &cache.Topic{
		ID:              topic.ID,
		Slug:            topic.Slug,
		LocaliseKeyName: topic.Title,
		ParentID:        parentID,
		ReleaseDate:     topic.ReleaseDate,
		List:            cache.NewSubTopicsMap(),
	}
	return topicCache
}

func getRootTopicCachePublic(ctx context.Context, subtopicsChan chan models.Topic, topicClient topicCli.Clienter, rootTopic models.Topic) *cache.Topic {
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

	var wg sync.WaitGroup
	wg.Add(2)

	// get subtopics ids
	go func() {
		defer wg.Done()
		getSubtopicsPublic(ctx, subtopicsChan, topicClient, rootTopic.ID)
		close(subtopicsChan)
	}()

	// extract subtopic id from channel to update rootTopicCache
	go func() {
		defer wg.Done()
		for s := range subtopicsChan {
			subtopic := cache.Subtopic{
				ID:              s.ID,
				Slug:            s.Slug,
				LocaliseKeyName: s.Title,
				ReleaseDate:     s.ReleaseDate,
			}

			subtopicsIDMap.AppendSubtopicID(s.ID, subtopic)
		}
	}()

	wg.Wait()

	rootTopicCache.List = subtopicsIDMap
	rootTopicCache.Query = subtopicsIDMap.GetSubtopicsIDsQuery()

	return rootTopicCache
}

func getSubtopicsPublic(ctx context.Context, subtopicsChan chan models.Topic, topicClient topicCli.Clienter, topLevelTopicID string) {
	// get subtopics from dp-topic-api
	subTopics, err := topicClient.GetSubtopicsPublic(ctx, topicCli.Headers{}, topLevelTopicID)
	if err != nil {
		if err.Status() != http.StatusNotFound {
			logData := log.Data{
				"top_level_topic_id": topLevelTopicID,
			}
			log.Error(ctx, "failed to get subtopics from topic-api", err, logData)
		}

		// stop as there are no subtopics items or failed to get subtopics
		return
	}

	// dereference sub-topics items to allow ranging through them
	var subTopicItems []models.Topic
	if subTopics.PublicItems == nil {
		err := errors.New("items is nil")
		log.Error(ctx, "failed to dereference sub-topics items pointer", err)
		return
	}
	subTopicItems = *subTopics.PublicItems

	var wg sync.WaitGroup

	// get subtopics ids of the subtopics items if they exist
	for i := range subTopicItems {
		wg.Add(1)

		// send subtopic id to channel
		subtopicsChan <- subTopicItems[i]

		go func(index int) {
			defer wg.Done()
			getSubtopicsPublic(ctx, subtopicsChan, topicClient, subTopicItems[index].ID)
		}(i)
	}
	wg.Wait()
}
