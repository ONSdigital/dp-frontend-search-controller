package config

import (
	"time"

	"github.com/kelseyhightower/envconfig"
)

// Config represents service configuration for dp-frontend-search-controller
type Config struct {
	BindAddr                      string        `envconfig:"BIND_ADDR"`
	APIRouterURL                  string        `envconfig:"API_ROUTER_URL"`
	GracefulShutdownTimeout       time.Duration `envconfig:"GRACEFUL_SHUTDOWN_TIMEOUT"`
	HealthCheckInterval           time.Duration `envconfig:"HEALTHCHECK_INTERVAL"`
	HealthCheckCriticalTimeout    time.Duration `envconfig:"HEALTHCHECK_CRITICAL_TIMEOUT"`
	DefaultSort                   string        `envconfig:"DEFAULT_SORT"`
	DefaultOffset                 int           `envconfig:"DEFAULT_OFFSET"`
	DefaultPage                   int           `envconfig:"DEFAULT_PAGE"`
	DefaultLimit                  int           `envconfig:"DEFAULT_LIMIT"`
	DefaultMaximumLimit           int           `envconfig:"DEFAULT_MAXIMUM_LIMIT"`
	DefaultMaximumSearchResults   int           `envconfig:"DEFAULT_MAXIMUM_SEARCH_RESULTS"`
	PatternLibraryAssetsPath      string        `envconfig:"PATTERN_LIBRARY_ASSETS_PATH"`
	SiteDomain                    string        `envconfig:"SITE_DOMAIN"`
	Debug                         bool          `envconfig:"DEBUG"`
	SupportedLanguages            []string      `envconfig:"SUPPORTED_LANGUAGES"`
	EnableCensusTopicFilterOption bool          `envconfig:"ENABLE_CENSUS_TOPIC_FILTER_OPTION"`
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
		cfg.PatternLibraryAssetsPath = "//cdn.ons.gov.uk/dp-design-system/f1f6886"
	}
	return cfg, nil
}

func get() (*Config, error) {
	if cfg != nil {
		return cfg, nil
	}

	cfg := &Config{
		BindAddr:                      "localhost:25000",
		APIRouterURL:                  "http://localhost:23200/v1",
		GracefulShutdownTimeout:       5 * time.Second,
		HealthCheckInterval:           30 * time.Second,
		HealthCheckCriticalTimeout:    90 * time.Second,
		DefaultSort:                   "relevance",
		DefaultOffset:                 0,
		DefaultPage:                   1,
		DefaultLimit:                  10,
		DefaultMaximumLimit:           50,
		DefaultMaximumSearchResults:   500,
		SiteDomain:                    "localhost",
		Debug:                         false,
		SupportedLanguages:            []string{"en", "cy"},
		EnableCensusTopicFilterOption: false,
	}

	return cfg, envconfig.Process("", cfg)
}
