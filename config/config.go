package config

import (
	"time"

	"github.com/kelseyhightower/envconfig"
)

// Config represents service configuration for dp-frontend-search-controller
type Config struct {
	BindAddr                    string        `envconfig:"BIND_ADDR"`
	RendererURL                 string        `envconfig:"RENDERER_URL"`
	SearchAPIURL                string        `envconfig:"SEARCH_API_URL"`
	GracefulShutdownTimeout     time.Duration `envconfig:"GRACEFUL_SHUTDOWN_TIMEOUT"`
	HealthCheckInterval         time.Duration `envconfig:"HEALTHCHECK_INTERVAL"`
	HealthCheckCriticalTimeout  time.Duration `envconfig:"HEALTHCHECK_CRITICAL_TIMEOUT"`
	DefaultOffset               int           `envconfig:"DEFAULT_OFFSET"`
	DefaultSort                 string        `envconfig:"DEFAULT_SORT"`
	DefaultPage                 int           `envconfig:"DEFAULT_PAGE"`
	DefaultLimit                int           `envconfig:"DEFAULT_LIMIT"`
	DefaultMaximumLimit         int           `envconfig:"DEFAULT_MAXIMUM_LIMIT"`
	DefaultMaximumSearchResults int           `envconfig:"DEFAULT_MAXIMUM_SEARCH_RESULTS"`
}

var cfg *Config

// Get returns the default config with any modifications through environment
// variables
func Get() (*Config, error) {
	if cfg != nil {
		return cfg, nil
	}

	cfg := &Config{
		BindAddr:                    "localhost:25000",
		RendererURL:                 "http://localhost:20010",
		SearchAPIURL:                "http://localhost:23900",
		GracefulShutdownTimeout:     5 * time.Second,
		HealthCheckInterval:         30 * time.Second,
		HealthCheckCriticalTimeout:  90 * time.Second,
		DefaultOffset:               0,
		DefaultSort:                 "relevance",
		DefaultPage:                 1,
		DefaultLimit:                10,
		DefaultMaximumLimit:         50,
		DefaultMaximumSearchResults: 500,
	}

	return cfg, envconfig.Process("", cfg)
}
