package public

import (
	"context"
	"errors"
	"strings"
	"sync"

	"github.com/ONSdigital/dp-frontend-search-controller/cache"
	"github.com/ONSdigital/dp-topic-api/models"
	topicCli "github.com/ONSdigital/dp-topic-api/sdk"
	"github.com/ONSdigital/log.go/v2/log"
)

// UpdateCensusTopic is a function to update the census topic cache in web (public) mode.
// This function talks to the dp-topic-api via its public endpoints to retrieve the census topic and its subtopic ids
// The data returned by the dp-topic-api is of type *models.PublicSubtopics which is then transformed to *cache.Topic in this function for the controller
// If an error has occurred, this is captured in log.Error and then a nil *cache.Topic is returned
func UpdateCensusTopic(ctx context.Context, topicClient topicCli.Clienter) func() *cache.Topic {
	return func() *cache.Topic {
		// get root topics from dp-topic-api
		rootTopics, err := topicClient.GetRootTopicsPublic(ctx, topicCli.Headers{})
		if err != nil {
			logData := log.Data{
				"req_headers": topicCli.Headers{},
			}
			log.Error(ctx, "failed to get root topics from topic-api", err, logData)
			return nil
		}

		//deference root topics items to allow ranging through them
		if rootTopics.PublicItems == nil {
			err := errors.New("root topic public items is nil")
			log.Error(ctx, "failed to deference root topics items pointer", err)
			return nil
		}
		rootTopicItems := *rootTopics.PublicItems

		var censusTopicCache *cache.Topic

		// go through each root topic, find census topic and gets its data for caching which includes subtopic ids
		for i := range rootTopicItems {
			if rootTopicItems[i].Title == cache.CensusTopicTitle {
				subtopicsIDChan := make(chan string)

				censusTopicCache = getRootTopicCachePublic(ctx, subtopicsIDChan, topicClient, rootTopicItems[i])
				break
			}
		}

		if censusTopicCache == nil {
			err := errors.New("census root topic not found")
			log.Error(ctx, "failed to get census topic to cache", err)
			return nil
		}

		return censusTopicCache
	}
}

func getRootTopicCachePublic(ctx context.Context, subtopicsIDChan chan string, topicClient topicCli.Clienter, rootTopic models.Topic) *cache.Topic {
	rootTopicCache := &cache.Topic{
		ID:              rootTopic.ID,
		LocaliseKeyName: rootTopic.Title,
	}

	subtopicsIDMap := cache.NewSubTopicsMap()
	subtopicsIDMap.AppendSubtopicID(rootTopic.ID)

	var wg sync.WaitGroup
	wg.Add(2)

	// get subtopics ids
	go func() {
		defer wg.Done()
		getSubtopicsIDsPublic(ctx, subtopicsIDChan, topicClient, rootTopic.ID)
		close(subtopicsIDChan)
	}()

	// extract subtopic id from channel to update rootTopicCache
	go func() {
		defer wg.Done()
		for subtopicID := range subtopicsIDChan {
			subtopicsIDMap.AppendSubtopicID(subtopicID)
		}
	}()

	wg.Wait()

	rootTopicCache.List = subtopicsIDMap
	rootTopicCache.Query = subtopicsIDMap.GetSubtopicsIDsQuery()

	return rootTopicCache
}

func getSubtopicsIDsPublic(ctx context.Context, subtopicsIDChan chan string, topicClient topicCli.Clienter, topLevelTopicID string) {
	// get subtopics from dp-topic-api
	subTopics, err := topicClient.GetSubtopicsPublic(ctx, topicCli.Headers{}, topLevelTopicID)
	if err != nil {
		if !strings.Contains(err.Error(), "404") {
			logData := log.Data{
				"req_headers":        topicCli.Headers{},
				"top_level_topic_id": topLevelTopicID,
			}
			log.Error(ctx, "failed to get subtopics from topic-api", err, logData)
		}

		// stop as there are no subtopics items or failed to get subtopics
		return
	}

	//deference sub topics items to allow ranging through them
	var subTopicItems []models.Topic
	if subTopics.PublicItems != nil {
		subTopicItems = *subTopics.PublicItems
	} else {
		err := errors.New("sub topics public items is nil")
		log.Error(ctx, "failed to deference sub topics items pointer", err)
		return
	}

	var wg sync.WaitGroup

	// get subtopics ids of the subtopics items if they exist
	for i := range subTopicItems {
		wg.Add(1)

		// send subtopic id to channel
		subtopicsIDChan <- subTopicItems[i].ID

		go func(index int) {
			defer wg.Done()
			getSubtopicsIDsPublic(ctx, subtopicsIDChan, topicClient, subTopicItems[index].ID)
		}(i)
	}
	wg.Wait()
}
