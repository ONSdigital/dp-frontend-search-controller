package cache

import "context"

// GetMockCensusTopicCacheList returns a mocked list of cache which contains the census topic cache and the census topic cache itself
// should have census topic data
func GetMockCensusTopicCacheList(ctx context.Context) (*CacheList, error) {
	testCensusTopicCache, err := NewTopicCache(ctx, nil)
	if err != nil {
		return nil, err
	}

	testCensusTopicCache.Set("Census", GetMockCensusTopic())

	cacheList := CacheList{
		CensusTopic: testCensusTopicCache,
	}

	return &cacheList, nil
}

// GetMockCensusTopic returns a mocked Cenus topic which contains all the information for the mock census topic
func GetMockCensusTopic() *Topic {
	mockCensusTopic := &Topic{
		ID:              "1234",
		LocaliseKeyName: CensusTopicTitle,
		Query:           "1234,5678",
	}

	mockCensusTopic.List = NewSubTopicsMap()
	mockCensusTopic.List.AppendSubtopicID("1234")
	mockCensusTopic.List.AppendSubtopicID("5678")

	return mockCensusTopic
}
