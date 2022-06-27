package cache

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ONSdigital/dp-frontend-search-controller/dpcache"
	"github.com/ONSdigital/log.go/v2/log"
)

const (
	// CensusTopicTitle is the title of the Census topic stored in mongodb which is accessible by using dp-topic-api
	CensusTopicTitle = "Census"
)

// TopicCache is a wrapper to dpcache.Cache which has additional fields and methods specifically for caching topics
type TopicCache struct {
	*dpcache.Cache
}

// Topic represents the data which is cached for a topic to be used by the dp-frontend-search-controller
type Topic struct {
	ID              string
	LocaliseKeyName string
	// Query is a comma separated string of topic id and its subtopic ids which will be used by the controller to create the query
	Query string
	// List is a map[string]bool which contains the topic id and its subtopic ids which will be used to check if the topic id given by the user exists
	List *SubtopicsIDs
}

// NewTopicCache create a topic cache object to be used in the service which will update at every updateInterval
// If updateInterval is nil, this means that the cache will only be updated once at the start of the service
func NewTopicCache(ctx context.Context, updateInterval *time.Duration) (*TopicCache, error) {
	cache, err := dpcache.NewCache(ctx, updateInterval)
	if err != nil {
		logData := log.Data{
			"update_interval": updateInterval,
		}
		log.Error(ctx, "failed to create cache from dpcache", err, logData)
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
		return nil, err
	}

	topicCacheData, ok := topicCacheInterface.(*Topic)
	if !ok {
		err := errors.New("topicCacheInterface is not type *Topic")
		log.Error(ctx, "failed type assertion on topicCacheInterface", err)
		return nil, err
	}

	if topicCacheData == nil {
		err := errors.New("topicCacheData is nil")
		log.Error(ctx, "cached topic data is nil", err)
		return nil, err
	}

	return topicCacheData, nil
}

func (dc *TopicCache) GetCensusData(ctx context.Context) (*Topic, error) {
	censusTopicCache, err := dc.GetData(ctx, CensusTopicTitle)
	if err != nil {
		logData := log.Data{
			"key": CensusTopicTitle,
		}
		log.Error(ctx, "failed to get cached census topic data", err, logData)
		return nil, err
	}

	return censusTopicCache, nil
}

// AddUpdateFunc adds an update function to the topic cache for a topic with the `title` passed to the function
// This update function will then be triggered once or at every fixed interval as per the prior setup of the TopicCache
func (dc *TopicCache) AddUpdateFunc(title string, updateFunc func() (*Topic, error)) {
	dc.UpdateFuncs[title] = func() (interface{}, error) {
		return updateFunc()
	}
}
