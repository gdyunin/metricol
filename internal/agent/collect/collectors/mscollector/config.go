package mscollector

import (
	"flag"
	"fmt"

	"github.com/caarlos0/env/v6"
)

const (
	defaultPollInterval = 2 // Default polling interval in seconds.
)

// MemStatsCollectorConfig represents the configuration for the MemStatsCollector.
type MemStatsCollectorConfig struct {
	PollInterval int `env:"POLL_INTERVAL"` // Polling interval in seconds, configurable via environment variables.
}

// ParseConfig parses the configuration for MemStatsCollector.
// It first reads the default configuration, applies command-line flag overrides,
// and then applies environment variables.
// Returns the populated configuration or an error if environment variable parsing fails.
func ParseConfig() (*MemStatsCollectorConfig, error) {
	cfg := MemStatsCollectorConfig{
		PollInterval: defaultPollInterval,
	}

	parseFlagsOrSetDefault(&cfg)

	// Parse environment variables and override any matching values in the configuration.
	if err := env.Parse(&cfg); err != nil {
		return nil, fmt.Errorf("failed to parse environment variables for configuration: %w", err)
	}
	return &cfg, nil
}

// parseFlagsOrSetDefault overrides configuration values with command-line flags
// or retains the default values if flags are not provided.
func parseFlagsOrSetDefault(cfg *MemStatsCollectorConfig) {
	flag.IntVar(&cfg.PollInterval, "p", cfg.PollInterval, "Polling interval for collecting metrics in seconds.")
	flag.Parse()
}
