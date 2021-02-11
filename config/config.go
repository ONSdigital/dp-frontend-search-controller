package config

import (
	"time"

	"github.com/kelseyhightower/envconfig"
)

// Config represents service configuration for dp-frontend-search-controller
type Config struct {
	BindAddr                   string        `envconfig:"BIND_ADDR"`
	RendererURL                string        `envconfig:"RENDERER_URL"`
	SearchAPIURL               string        `envconfig:"SEARCH_API_URL"`
	GracefulShutdownTimeout    time.Duration `envconfig:"GRACEFUL_SHUTDOWN_TIMEOUT"`
	HealthCheckInterval        time.Duration `envconfig:"HEALTHCHECK_INTERVAL"`
	HealthCheckCriticalTimeout time.Duration `envconfig:"HEALTHCHECK_CRITICAL_TIMEOUT"`
}

var cfg *Config

// Get returns the default config with any modifications through environment
// variables
func Get() (*Config, error) {
	if cfg != nil {
		return cfg, nil
	}

	cfg := &Config{
		BindAddr:                   ":25000",
		RendererURL:                "http://localhost:20010",
		SearchAPIURL:               "http://localhost:23900",
		GracefulShutdownTimeout:    5 * time.Second,
		HealthCheckInterval:        30 * time.Second,
		HealthCheckCriticalTimeout: 90 * time.Second,
	}

	return cfg, envconfig.Process("", cfg)
}
