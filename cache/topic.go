package cache

import (
	"context"
	"errors"
	"fmt"
	"time"

	dpcache "github.com/ONSdigital/dp-cache"
	"github.com/ONSdigital/log.go/v2/log"
)

// CensusTopicID is the id of the Census topic stored in mongodb which is accessible by using dp-topic-api
var CensusTopicID string

// TopicCache is a wrapper to dpcache.Cache which has additional fields and methods specifically for caching topics
type TopicCache struct {
	*dpcache.Cache
}

// Topic represents the data which is cached for a topic to be used by the dp-frontend-search-controller
type Topic struct {
	ID              string
	LocaliseKeyName string
	Slug            string
	ReleaseDate     *time.Time
	// Query is a comma separated string of topic id and its subtopic ids which will be used by the controller to create the query
	Query string
	// List is a map[string]Subtopics which contains the topic id and a list of it's subtopics
	List *Subtopics
	// This is a reference to the parent topic
	ParentID string
}

// NewTopicCache create a topic cache object to be used in the service which will update at every updateInterval
// If updateInterval is nil, this means that the cache will only be updated once at the start of the service
func NewTopicCache(ctx context.Context, updateInterval *time.Duration) (*TopicCache, error) {
	config := dpcache.Config{
		UpdateInterval: updateInterval,
	}

	cache, err := dpcache.NewCache(ctx, config)
	if err != nil {
		logData := log.Data{
			"update_interval": updateInterval,
		}
		log.Error(ctx, "failed to create topic cache from dpcache", err, logData)
		return nil, err
	}

	topicCache := &TopicCache{cache}

	return topicCache, nil
}

func (dc *TopicCache) GetData(ctx context.Context, key string) (*Topic, error) {
	topicCacheInterface, ok := dc.Get(key)
	if !ok {
		err := fmt.Errorf("cached topic data with key %s not found", key)
		log.Error(ctx, "failed to get cached topic data", err)
		return GetEmptyTopic(), err
	}

	topicCacheData, ok := topicCacheInterface.(*Topic)
	if !ok {
		err := errors.New("topicCacheInterface is not type *Topic")
		log.Error(ctx, "failed type assertion on topicCacheInterface", err)
		return GetEmptyTopic(), err
	}

	if topicCacheData == nil {
		err := errors.New("topicCacheData is nil")
		log.Error(ctx, "cached topic data is nil", err)
		return GetEmptyTopic(), err
	}

	return topicCacheData, nil
}

// AddUpdateFunc adds an update function to the topic cache for a topic with the `title` passed to the function
// This update function will then be triggered once or at every fixed interval as per the prior setup of the TopicCache
func (dc *TopicCache) AddUpdateFunc(title string, updateFunc func() *Topic) {
	dc.UpdateFuncs[title] = func() (interface{}, error) {
		// error handling is done within the updateFunc
		return updateFunc(), nil
	}
}

// AddUpdateFuncs adds an update function to the topic cache for a topic with the `title` passed to the function
// This update function will then be triggered once or at every fixed interval as per the prior setup of the TopicCache
func (dc *TopicCache) AddUpdateFuncs(updateFunc func() []*Topic) {
	// Invoke the updateFunc to get the slice of *Topic
	topics := updateFunc()

	// Iterate over each topic in the returned slice
	for _, topic := range topics {
		// Define an update function for the current topic
		// This update function simply returns the current topic as-is
		singleUpdateFunc := func() *Topic {
			return topic
		}

		// Add the update function to the TopicCache for the current topic's title
		dc.AddUpdateFunc(topic.Slug, singleUpdateFunc)
	}
}

func (dc *TopicCache) GetCensusData(ctx context.Context) (*Topic, error) {
	censusTopicCache, err := dc.GetData(ctx, CensusTopicID)
	if err != nil {
		logData := log.Data{
			"key": CensusTopicID,
		}
		log.Error(ctx, "failed to get cached census topic data", err, logData)
		return GetEmptyCensusTopic(), err
	}

	return censusTopicCache, nil
}

// GetEmptyCensusTopic returns an empty census topic cache in the event when updating the cache of the census topic fails
func GetEmptyCensusTopic() *Topic {
	return &Topic{
		ID:   CensusTopicID,
		List: NewSubTopicsMap(),
	}
}

// GetEmptyTopic returns an empty topic cache in the event when updating the cache of the topic fails
func GetEmptyTopic() *Topic {
	return &Topic{
		List: NewSubTopicsMap(),
	}
}
