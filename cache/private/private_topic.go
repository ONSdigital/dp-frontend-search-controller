package private

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

// UpdateCensusTopic is a function to update the census topic cache in publishing (private) mode.
// This function talks to the dp-topic-api via its private endpoints to retrieve the census topic and its subtopic ids
// The data returned by the dp-topic-api is of type *models.PrivateSubtopics which is then transformed in this function for the controller
// If an error has occurred, this is captured in log.Error and then an empty census topic is returned
func UpdateCensusTopic(ctx context.Context, serviceAuthToken string, topicClient topicCli.Clienter) func() *cache.Topic {
	return func() *cache.Topic {
		// get root topics from dp-topic-api
		rootTopics, err := topicClient.GetRootTopicsPrivate(ctx, topicCli.Headers{ServiceAuthToken: serviceAuthToken})
		if err != nil {
			logData := log.Data{
				"req_headers": topicCli.Headers{},
			}
			log.Error(ctx, "failed to get root topics from topic-api", err, logData)
			return cache.GetEmptyCensusTopic()
		}

		// deference root topics items to allow ranging through them
		var rootTopicItems []models.TopicResponse
		if rootTopics.PrivateItems != nil {
			rootTopicItems = *rootTopics.PrivateItems
		} else {
			err := errors.New("root topic private items is nil")
			log.Error(ctx, "failed to deference root topics items pointer", err)
			return cache.GetEmptyCensusTopic()
		}

		var censusTopicCache *cache.Topic

		// go through each root topic, find census topic and gets its data for caching which includes subtopic ids
		for i := range rootTopicItems {
			if rootTopicItems[i].Current.ID == cache.CensusTopicID {
				subtopicsChan := make(chan models.TopicResponse)

				censusTopicCache = getRootTopicCachePrivate(ctx, serviceAuthToken, subtopicsChan, topicClient, *rootTopicItems[i].Current)
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

// UpdateDataTopics is a function to update the data topic cache in publishing (private) mode.
// This function talks to the dp-topic-api via its private endpoints to retrieve the root topic and its subtopic ids
// The data returned by the dp-topic-api is of type *models.PrivateSubtopics which is then transformed in this function for the controller
// If an error has occurred, this is captured in log.Error and then an empty topic is returned
func UpdateDataTopics(ctx context.Context, serviceAuthToken string, topicClient topicCli.Clienter) func() []*cache.Topic {
	return func() []*cache.Topic {
		var topics []*cache.Topic
		processedTopics := make(map[string]bool)

		// get root topic from dp-topic-api
		rootTopic, err := topicClient.GetTopicPrivate(ctx, topicCli.Headers{ServiceAuthToken: serviceAuthToken}, cache.RootTopicID)
		if err != nil {
			logData := log.Data{
				"req_headers": topicCli.Headers{},
			}
			log.Error(ctx, "failed to get root topic from topic-api", err, logData)
			return []*cache.Topic{cache.GetEmptyTopic()}
		}

		// deference rootTopic's subTopicIDs to allow ranging through them
		var rootSubTopicIds []string
		if rootTopic.Current.SubtopicIds != nil {
			rootSubTopicIds = *rootTopic.Current.SubtopicIds
		} else {
			err := errors.New("root topic subtopic IDs is nil")
			log.Error(ctx, "failed to deference rootTopic subtopic IDs pointer", err)
			return []*cache.Topic{cache.GetEmptyTopic()}
		}

		// recursively process topics and their subtopics
		for i := range rootSubTopicIds {
			processTopic(ctx, serviceAuthToken, topicClient, rootSubTopicIds[i], &topics, processedTopics)
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

func processTopic(ctx context.Context, serviceAuthToken string, topicClient topicCli.Clienter, topicID string, topics *[]*cache.Topic, processedTopics map[string]bool) {
	// Check if the topic is already processed
	if processedTopics[topicID] {
		return
	}

	// Get the topic details from the topic client
	dataTopic, err := topicClient.GetTopicPrivate(ctx, topicCli.Headers{ServiceAuthToken: serviceAuthToken}, topicID)
	if err != nil {
		log.Error(ctx, "failed to get topic details from topic-api", err)
		return
	}

	if dataTopic != nil {
		// Append the current topic to the list of topics
		// subtopicsChan := make(chan models.Topic)
		*topics = append(*topics, mapTopicModelToCache(*dataTopic.Current))
		// Mark this topic as processed
		processedTopics[topicID] = true

		// Process each subtopic recursively
		if dataTopic.Current.SubtopicIds != nil {
			for _, subTopicID := range *dataTopic.Current.SubtopicIds {
				processTopic(ctx, serviceAuthToken, topicClient, subTopicID, topics, processedTopics)
			}
		}
	}
}

func mapTopicModelToCache(rootTopic models.Topic) *cache.Topic {
	rootTopicCache := &cache.Topic{
		ID:              rootTopic.ID,
		Slug:            rootTopic.Slug,
		LocaliseKeyName: rootTopic.Title,
		ReleaseDate:     rootTopic.ReleaseDate,
		List:            cache.NewSubTopicsMap(),
	}
	return rootTopicCache
}

func getRootTopicCachePrivate(ctx context.Context, serviceAuthToken string, subtopicsChan chan models.TopicResponse, topicClient topicCli.Clienter, rootTopic models.Topic) *cache.Topic {
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

	var wg sync.WaitGroup
	wg.Add(2)

	// get subtopics ids
	go func() {
		defer wg.Done()
		getSubtopicsIDsPrivate(ctx, serviceAuthToken, subtopicsChan, topicClient, rootTopic.ID)
		close(subtopicsChan)
	}()

	// extract subtopic id from channel to update rootTopicCache
	go func() {
		defer wg.Done()

		for s := range subtopicsChan {
			subtopic := cache.Subtopic{
				ID:              s.ID,
				Slug:            s.Next.Slug,
				LocaliseKeyName: s.Next.Title,
				ReleaseDate:     s.Next.ReleaseDate,
			}

			subtopicsIDMap.AppendSubtopicID(s.ID, subtopic)
		}
	}()

	wg.Wait()

	rootTopicCache.List = subtopicsIDMap
	rootTopicCache.Query = subtopicsIDMap.GetSubtopicsIDsQuery()

	return rootTopicCache
}

func getSubtopicsIDsPrivate(ctx context.Context, serviceAuthToken string, subtopicsChan chan models.TopicResponse, topicClient topicCli.Clienter, topLevelTopicID string) {
	topicCliReqHeaders := topicCli.Headers{ServiceAuthToken: serviceAuthToken}

	// get subtopics from dp-topic-api
	subTopics, err := topicClient.GetSubtopicsPrivate(ctx, topicCliReqHeaders, topLevelTopicID)
	if err != nil {
		if err.Status() != http.StatusNotFound {
			logData := log.Data{
				"req_headers":        topicCliReqHeaders,
				"top_level_topic_id": topLevelTopicID,
			}
			log.Error(ctx, "failed to get subtopics from topic-api", err, logData)
		}

		// stop as there are no subtopics items or failed to get subtopics
		return
	}

	// deference sub topics items to allow ranging through them
	if subTopics.PrivateItems == nil {
		err := errors.New("sub topics private items is nil")
		log.Error(ctx, "failed to deference sub topics items pointer", err)
		return
	}
	subTopicItems := *subTopics.PrivateItems

	var wg sync.WaitGroup

	// get subtopics ids of the subtopics items if they exist
	for i := range subTopicItems {
		wg.Add(1)

		// send subtopic id to channel
		subtopicsChan <- subTopicItems[i]

		go func(index int) {
			defer wg.Done()
			getSubtopicsIDsPrivate(ctx, serviceAuthToken, subtopicsChan, topicClient, subTopicItems[index].ID)
		}(i)
	}
	wg.Wait()
}
