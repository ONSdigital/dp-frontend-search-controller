package config

import (
	"time"

	"github.com/kelseyhightower/envconfig"
)

// Config represents service configuration for dp-frontend-search-controller
type Config struct {
	*ABTest
	APIRouterURL                   string        `envconfig:"API_ROUTER_URL"`
	BindAddr                       string        `envconfig:"BIND_ADDR"`
	CacheCensusTopicUpdateInterval time.Duration `envconfig:"CACHE_CENSUS_TOPICS_UPDATE_INTERVAL"`
	CacheDataTopicUpdateInterval   time.Duration `envconfig:"CACHE_DATA_TOPICS_UPDATE_INTERVAL"`
	CacheNavigationUpdateInterval  time.Duration `envconfig:"CACHE_NAVIGATION_UPDATE_INTERVAL"`
	CensusTopicID                  string        `envconfig:"CENSUS_TOPIC_ID"`
	Debug                          bool          `envconfig:"DEBUG"`
	DefaultLimit                   int           `envconfig:"DEFAULT_LIMIT"`
	DefaultMaximumLimit            int           `envconfig:"DEFAULT_MAXIMUM_LIMIT"`
	DefaultMaximumSearchResults    int           `envconfig:"DEFAULT_MAXIMUM_SEARCH_RESULTS"`
	DefaultOffset                  int           `envconfig:"DEFAULT_OFFSET"`
	DefaultPage                    int           `envconfig:"DEFAULT_PAGE"`
	*DefaultSort
	EnableAggregationPages                  bool          `envconfig:"ENABLE_AGGREGATION_PAGES"`
	EnableTopicAggregationPages             bool          `envconfig:"ENABLE_TOPIC_AGGREGATION_PAGES"`
	FeedbackAPIURL                          string        `envconfig:"FEEDBACK_API_URL"`
	EnableCensusDimensionsFilterOption      bool          `envconfig:"ENABLE_CENSUS_DIMENSIONS_FILTER_OPTION"`
	EnableCensusPopulationTypesFilterOption bool          `envconfig:"ENABLE_CENSUS_POPULATION_TYPE_FILTER_OPTION"`
	EnableCensusTopicFilterOption           bool          `envconfig:"ENABLE_CENSUS_TOPIC_FILTER_OPTION"`
	EnableNewNavBar                         bool          `envconfig:"ENABLE_NEW_NAV_BAR"`
	GracefulShutdownTimeout                 time.Duration `envconfig:"GRACEFUL_SHUTDOWN_TIMEOUT"`
	HealthCheckCriticalTimeout              time.Duration `envconfig:"HEALTHCHECK_CRITICAL_TIMEOUT"`
	HealthCheckInterval                     time.Duration `envconfig:"HEALTHCHECK_INTERVAL"`
	OTBatchTimeout                          time.Duration `encconfig:"OTEL_BATCH_TIMEOUT"`
	OTExporterOTLPEndpoint                  string        `envconfig:"OTEL_EXPORTER_OTLP_ENDPOINT"`
	OTServiceName                           string        `envconfig:"OTEL_SERVICE_NAME"`
	OtelEnabled                             bool          `envconfig:"OTEL_ENABLED"`
	IsPublishing                            bool          `envconfig:"IS_PUBLISHING"`
	PatternLibraryAssetsPath                string        `envconfig:"PATTERN_LIBRARY_ASSETS_PATH"`
	ServiceAuthToken                        string        `envconfig:"SERVICE_AUTH_TOKEN"   json:"-"`
	SiteDomain                              string        `envconfig:"SITE_DOMAIN"`
	SupportedLanguages                      []string      `envconfig:"SUPPORTED_LANGUAGES"`
}

type ABTest struct {
	AspectID   string `envconfig:"AB_TEST_ASPECT_ID"`
	Enabled    bool   `envconfig:"AB_TEST_ENABLED"`
	Percentage int    `envconfig:"AB_TEST_PERCENTAGE"`
	Exit       string `envconfig:"AB_TEST_EXIT"`
}

type DefaultSort struct {
	Aggregation      string `envconfig:"DEFAULT_AGGREGATION_SORT"`
	Dataset          string `envconfig:"DEFAULT_DATASET_SORT"`
	Other            string `envconfig:"DEFAULT_SORT"`
	PreviousReleases string `envconfig:"DEFAULT_PREVIOUS_RELEASES_SORT"`
	RelatedData      string `envconfig:"DEFAULT_RELATED_DATA_SORT"`
}

var cfg *Config

// Get returns the default config with any modifications through environment
// variables
func Get() (*Config, error) {
	newCfg, err := get()
	if err != nil {
		return nil, err
	}

	if newCfg.Debug {
		newCfg.PatternLibraryAssetsPath = "http://localhost:9002/dist/assets"
	} else {
		newCfg.PatternLibraryAssetsPath = "//cdn.ons.gov.uk/dp-design-system/27f731a"
	}
	return newCfg, nil
}

func get() (*Config, error) {
	if cfg != nil {
		return cfg, nil
	}

	cfg := &Config{
		APIRouterURL: "http://localhost:23200/v1",
		ABTest: &ABTest{
			AspectID:   "dp-frontend-search-controller",
			Enabled:    true,
			Exit:       "search-ab-exit",
			Percentage: 0,
		},
		BindAddr:                       ":25000",
		CacheCensusTopicUpdateInterval: 30 * time.Minute,
		CacheDataTopicUpdateInterval:   30 * time.Minute,
		CacheNavigationUpdateInterval:  30 * time.Minute,
		CensusTopicID:                  "4445",
		Debug:                          false,
		DefaultLimit:                   10,
		DefaultMaximumLimit:            50,
		DefaultMaximumSearchResults:    500,
		DefaultOffset:                  0,
		DefaultPage:                    1,
		DefaultSort: &DefaultSort{
			Aggregation:      "release_date",
			Dataset:          "release_date",
			Other:            "relevance",
			PreviousReleases: "release_date",
			RelatedData:      "title",
		},
		FeedbackAPIURL:                          "http://localhost:23200/v1/feedback",
		EnableCensusTopicFilterOption:           false,
		EnableCensusPopulationTypesFilterOption: false,
		EnableCensusDimensionsFilterOption:      false,
		EnableAggregationPages:                  false,
		EnableTopicAggregationPages:             false,
		EnableNewNavBar:                         false,
		GracefulShutdownTimeout:                 5 * time.Second,
		HealthCheckCriticalTimeout:              90 * time.Second,
		HealthCheckInterval:                     30 * time.Second,
		OTBatchTimeout:                          5 * time.Second,
		OTExporterOTLPEndpoint:                  "localhost:4317",
		OTServiceName:                           "dp-frontend-search-controller",
		OtelEnabled:                             false,
		IsPublishing:                            false,
		ServiceAuthToken:                        "",
		SiteDomain:                              "localhost",
		SupportedLanguages:                      []string{"en", "cy"},
	}

	return cfg, envconfig.Process("", cfg)
}
