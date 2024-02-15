package cache

// CacheList is a list of caches for the dp-frontend-search-controller
type List struct {
	CensusTopic *TopicCache
	DataTopic   *TopicCache
	Navigation  *NavigationCache
}
