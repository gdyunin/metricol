// Package config provides functionality to configure an orchestrate with parameters
// that can be set via environment variables or command-line flags.
package config

import (
	"flag"
	"fmt"

	"github.com/caarlos0/env/v6"
	"go.uber.org/zap"
)

// All basic settings.
const (
	defaultServerAddress  = "localhost:8080"
	defaultPollInterval   = 2
	defaultReportInterval = 10
)

// Config holds the configuration for the orchestrate, including server address,
// polling interval, and reporting interval.
type Config struct {
	ServerAddress  string `env:"ADDRESS"`         // Address of the server to connect to
	PollInterval   int    `env:"POLL_INTERVAL"`   // Interval for polling metrics
	ReportInterval int    `env:"REPORT_INTERVAL"` // Interval for reporting metrics
}

// ParseConfig initializes the Config with basic values,
// overrides them with environment variables if available,
// and finally allows command-line flags to set or override the configuration.
// It returns an error if environment variable parsing fails.
func ParseConfig(logger *zap.SugaredLogger) (*Config, error) {
	// Default settings for the orchestrate configuration.
	cfg := Config{
		ServerAddress:  defaultServerAddress,
		PollInterval:   defaultPollInterval,
		ReportInterval: defaultReportInterval,
	}

	// Parse command-line arguments or set basic settings if no args are provided.
	parseFlagsOrSetDefault(&cfg)

	// Attempt to parse values from environment variables; if unsuccessful, return the error.
	if err := env.Parse(&cfg); err != nil {
		return nil, fmt.Errorf("failed to parse environment variables: %w", err)
	}

	logger.Infof("App config: %+v", cfg)
	return &cfg, nil
}

// parseFlagsOrSetDefault attempts to populate the Config from command-line flags.
// If no flags are provided, it retains the basic values set in the configuration.
func parseFlagsOrSetDefault(cfg *Config) {
	flag.IntVar(&cfg.PollInterval, "p", cfg.PollInterval, "Interval (in seconds) for collecting metrics")
	flag.IntVar(&cfg.ReportInterval, "r", cfg.ReportInterval, "Interval (in seconds) for sending metrics")
	flag.StringVar(&cfg.ServerAddress, "a", cfg.ServerAddress, "Address of the server to connect to")
	flag.Parse()
}
