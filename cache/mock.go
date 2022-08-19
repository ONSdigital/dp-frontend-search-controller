package cache

import (
	"context"
	"fmt"

	"github.com/ONSdigital/dp-topic-api/models"
)

// GetMockCacheList returns a mocked list of cache which contains the census topic cache and navigation cache
func GetMockCacheList(ctx context.Context, lang string) (*CacheList, error) {
	testCensusTopicCache, err := getMockCensusTopicCache(ctx)
	if err != nil {
		return nil, err
	}

	testNavigationCache, err := getMockNavigationCache(ctx, lang)
	if err != nil {
		return nil, err
	}

	cacheList := CacheList{
		CensusTopic: testCensusTopicCache,
		Navigation:  testNavigationCache,
	}

	return &cacheList, nil
}

// getMockCensusTopicCache returns a mocked Cenus topic which contains all the information for the mock census topic
func getMockCensusTopicCache(ctx context.Context) (*TopicCache, error) {
	testCensusTopicCache, err := NewTopicCache(ctx, nil)
	if err != nil {
		return nil, err
	}

	testCensusTopicCache.Set(CensusTopicID, GetMockCensusTopic())

	return testCensusTopicCache, nil
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

// getMockNavigationCache returns a mocked navigation cache which should have navigation data
func getMockNavigationCache(ctx context.Context, lang string) (*NavigationCache, error) {
	testNavigationCache, err := NewNavigationCache(ctx, nil)
	if err != nil {
		return nil, err
	}

	mockNavigationData := &models.Navigation{
		Description: "this is a test description",
	}

	navigationlangKey := testNavigationCache.GetCachingKeyForNavigationLanguage(lang)

	testNavigationCache.Set(navigationlangKey, mockNavigationData)

	return testNavigationCache, nil
}
