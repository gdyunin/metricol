// Package agent provides functionality to configure an agent with parameters
// that can be set via environment variables or command-line flags.
package agent

import (
	"flag"
	"fmt"

	"github.com/caarlos0/env/v6"
)

// All default settings.
const (
	defaultServerAddress  = "localhost:8080"
	defaultPollInterval   = 2
	defaultReportInterval = 10
)

// Config holds the configuration for the agent, including server address,
// polling interval, and reporting interval.
type Config struct {
	ServerAddress  string `env:"ADDRESS"`         // Address of the server to connect to
	PollInterval   int    `env:"POLL_INTERVAL"`   // Interval for polling metrics
	ReportInterval int    `env:"REPORT_INTERVAL"` // Interval for reporting metrics
}

// ParseConfig initializes the Config with default values,
// overrides them with environment variables if available,
// and finally allows command-line flags to set or override the configuration.
func ParseConfig() (*Config, error) {
	// Default settings for the agent configuration.
	cfg := Config{
		ServerAddress:  defaultServerAddress,
		PollInterval:   defaultPollInterval,
		ReportInterval: defaultReportInterval,
	}

	// Parse command-line arguments or set default settings if no args are provided.
	parseFlagsOrSetDefault(&cfg)

	// Attempt to parse values from environment variables; if unsuccessful, return the error.
	if err := env.Parse(&cfg); err != nil {
		return nil, fmt.Errorf("error parse env variables %w", err)
	}
	return &cfg, nil
}

// parseFlagsOrSetDefault attempts to populate the Config from command-line flags.
// If no flags are provided, it retains the default values set in the configuration.
func parseFlagsOrSetDefault(cfg *Config) {
	flag.IntVar(&cfg.PollInterval, "p", cfg.PollInterval, "Interval for collecting metrics")
	flag.IntVar(&cfg.ReportInterval, "r", cfg.ReportInterval, "Interval for sending metrics")
	flag.StringVar(&cfg.ServerAddress, "a", cfg.ServerAddress, "Address of the server")
	flag.Parse()
}
