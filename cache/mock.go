package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/ONSdigital/dp-topic-api/models"
)

// GetMockCacheList returns a mocked list of cache which contains the census topic cache and navigation cache
func GetMockCacheList(ctx context.Context, lang string) (*List, error) {
	testCensusTopicCache, err := getMockCensusTopicCache(ctx)
	if err != nil {
		return nil, err
	}

	testDataTopicCache, err := getMockDataTopicCache(ctx)
	if err != nil {
		return nil, err
	}

	testNavigationCache, err := getMockNavigationCache(ctx, lang)
	if err != nil {
		return nil, err
	}

	cacheList := List{
		CensusTopic: testCensusTopicCache,
		DataTopic:   testDataTopicCache,
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
	mockCensusTopic.List.AppendSubtopicID("1234", Subtopic{ID: "1234", LocaliseKeyName: "International Migration", ReleaseDate: timeHelper("2022-10-10T08:30:00Z")})
	mockCensusTopic.List.AppendSubtopicID("5678", Subtopic{ID: "5678", LocaliseKeyName: "Age", ReleaseDate: timeHelper("2022-11-09T09:30:00Z")})
	mockCensusTopic.List.AppendSubtopicID(CensusTopicID, Subtopic{ID: CensusTopicID, LocaliseKeyName: "Census", ReleaseDate: timeHelper("2022-10-10T09:30:00Z")})

	return mockCensusTopic
}

// getMockDataTopicCache returns a mocked data topic which contains all the information for the mock data topic
func getMockDataTopicCache(ctx context.Context) (*TopicCache, error) {
	testDataTopicCache, err := NewTopicCache(ctx, nil)
	if err != nil {
		return nil, err
	}

	testDataTopicCache.Set("economy", GetMockDataTopic("6734", "economy", "Economy", ""))
	testDataTopicCache.Set("environmentalaccounts", GetMockDataTopic("1834", "environmentalaccounts", "Environmental Accounts", "6734"))
	testDataTopicCache.Set("governmentpublicsectorandtaxes", GetMockDataTopic("8268", "governmentpublicsectorandtaxes", "Government Public Sector and Taxes", "6734"))
	testDataTopicCache.Set("publicsectorfinance", GetMockDataTopic("3687", "publicsectorfinance", "Public Sector Finance", "8268"))

	return testDataTopicCache, nil
}

// GetMockDataTopic returns a mocked data topic which contains all the information for the mock data topic
func GetMockDataTopic(id, slug, title, parentID string) *Topic {
	mockDataTopic := &Topic{
		ID:              id,
		Slug:            slug,
		LocaliseKeyName: title,
		ParentID:        parentID,
	}

	mockDataTopic.List = NewSubTopicsMap()
	mockDataTopic.List.AppendSubtopicID("1234", Subtopic{ID: "1234", LocaliseKeyName: "International Migration", ReleaseDate: timeHelper("2022-10-10T08:30:00Z")})
	mockDataTopic.List.AppendSubtopicID("1834", Subtopic{ID: "1834", LocaliseKeyName: "Environmental Accounts", ReleaseDate: timeHelper("2022-11-09T09:30:00Z")})

	return mockDataTopic
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

// timeHelper is a helper function given a time returns a Time pointer
func timeHelper(timeFormat string) *time.Time {
	t, _ := time.Parse(time.RFC3339, timeFormat)
	return &t
}
