package cache

import (
	"context"
	"fmt"

	"github.com/ONSdigital/dp-topic-api/models"
)

// GetMockCensusTopicCacheList returns a mocked list of cache which contains the census topic cache and the census topic cache itself
// should have census topic data
func GetMockCensusTopicCacheList(ctx context.Context) (*CacheList, error) {
	testCensusTopicCache, err := NewTopicCache(ctx, nil)
	if err != nil {
		return nil, err
	}

	testCensusTopicCache.Set(CensusTopicID, GetMockCensusTopic())

	cacheList := CacheList{
		CensusTopic: testCensusTopicCache,
	}

	return &cacheList, nil
}

// GetMockCensusTopic returns a mocked Cenus topic which contains all the information for the mock census topic
func GetMockCensusTopic() *Topic {
	mockCensusTopic := &Topic{
		ID:              CensusTopicID,
		LocaliseKeyName: "Census",
		Query:           fmt.Sprintf("1234,5678,%s", CensusTopicID),
	}

	mockCensusTopic.List = NewSubTopicsMap()
	mockCensusTopic.List.AppendSubtopicID("1234")
	mockCensusTopic.List.AppendSubtopicID("5678")
	mockCensusTopic.List.AppendSubtopicID(CensusTopicID)

	return mockCensusTopic
}

// GetMockNavigationCacheList returns a mocked list of cache which contains the navigation cache and the navigation cache itself
// should have navigation data
func GetMockNavigationCacheList(ctx context.Context, lang string) (*CacheList, error) {
	testNavigationCache, err := NewNavigationCache(ctx, nil)
	if err != nil {
		return nil, err
	}

	mockNavigationData := &models.Navigation{
		Description: "this is a test description",
	}

	navigationlangKey := testNavigationCache.GetCachingKeyForNavigationLanguage(lang)

	testNavigationCache.Set(navigationlangKey, mockNavigationData)

	cacheList := CacheList{
		Navigation: testNavigationCache,
	}

	return &cacheList, nil
}
