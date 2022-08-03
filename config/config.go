package config

import (
	"time"

	"github.com/kelseyhightower/envconfig"
)

// Config represents service configuration for dp-frontend-search-controller
type Config struct {
	APIRouterURL                   string        `envconfig:"API_ROUTER_URL"`
	BindAddr                       string        `envconfig:"BIND_ADDR"`
	CacheCensusTopicUpdateInterval time.Duration `envconfig:"CACHE_CENSUS_TOPICS_UPDATE_INTERVAL"`
	CacheNavigationUpdateInterval  time.Duration `envconfig:"CACHE_NAVIGATION_UPDATE_INTERVAL"`
	CensusTopicID                  string        `envconfig:"CENSUS_TOPIC_ID"`
	Debug                          bool          `envconfig:"DEBUG"`
	DefaultLimit                   int           `envconfig:"DEFAULT_LIMIT"`
	DefaultMaximumLimit            int           `envconfig:"DEFAULT_MAXIMUM_LIMIT"`
	DefaultMaximumSearchResults    int           `envconfig:"DEFAULT_MAXIMUM_SEARCH_RESULTS"`
	DefaultOffset                  int           `envconfig:"DEFAULT_OFFSET"`
	DefaultPage                    int           `envconfig:"DEFAULT_PAGE"`
	DefaultSort                    string        `envconfig:"DEFAULT_SORT"`
	EnableCensusTopicFilterOption  bool          `envconfig:"ENABLE_CENSUS_TOPIC_FILTER_OPTION"`
	GracefulShutdownTimeout        time.Duration `envconfig:"GRACEFUL_SHUTDOWN_TIMEOUT"`
	HealthCheckCriticalTimeout     time.Duration `envconfig:"HEALTHCHECK_CRITICAL_TIMEOUT"`
	HealthCheckInterval            time.Duration `envconfig:"HEALTHCHECK_INTERVAL"`
	IsPublishing                   bool          `envconfig:"IS_PUBLISHING"`
	NoIndexEnabled                 bool          `envconfig:"NO_INDEX_ENABLED"`
	PatternLibraryAssetsPath       string        `envconfig:"PATTERN_LIBRARY_ASSETS_PATH"`
	ServiceAuthToken               string        `envconfig:"SERVICE_AUTH_TOKEN"   json:"-"`
	SiteDomain                     string        `envconfig:"SITE_DOMAIN"`
	SupportedLanguages             []string      `envconfig:"SUPPORTED_LANGUAGES"`
}

var cfg *Config

// Get returns the default config with any modifications through environment
// variables
func Get() (*Config, error) {
	cfg, err := get()
	if err != nil {
		return nil, err
	}

	if cfg.Debug {
		cfg.PatternLibraryAssetsPath = "http://localhost:9002/dist/assets"
	} else {
		cfg.PatternLibraryAssetsPath = "//cdn.ons.gov.uk/dp-design-system/80d766d"
	}
	return cfg, nil
}

func get() (*Config, error) {
	if cfg != nil {
		return cfg, nil
	}

	cfg := &Config{
		APIRouterURL:                   "http://localhost:23200/v1",
		BindAddr:                       "localhost:25000",
		CacheCensusTopicUpdateInterval: 30 * time.Minute,
		CacheNavigationUpdateInterval:  10 * time.Second,
		CensusTopicID:                  "4445",
		Debug:                          false,
		DefaultLimit:                   10,
		DefaultMaximumLimit:            50,
		DefaultMaximumSearchResults:    500,
		DefaultOffset:                  0,
		DefaultPage:                    1,
		DefaultSort:                    "relevance",
		EnableCensusTopicFilterOption:  false,
		GracefulShutdownTimeout:        5 * time.Second,
		HealthCheckCriticalTimeout:     90 * time.Second,
		HealthCheckInterval:            30 * time.Second,
		IsPublishing:                   false,
		NoIndexEnabled:                 false,
		ServiceAuthToken:               "",
		SiteDomain:                     "localhost",
		SupportedLanguages:             []string{"en", "cy"},
	}

	return cfg, envconfig.Process("", cfg)
}
