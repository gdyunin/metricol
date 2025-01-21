package config

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
	defaultSigningKey     = ""
)

type Config struct {
	ServerAddress  string `env:"ADDRESS"`         // Address of the server to connect to
	SigningKey     string `env:"KEY"`             // Key used for signing requests to the server.
	PollInterval   int    `env:"POLL_INTERVAL"`   // Interval for polling metrics
	ReportInterval int    `env:"REPORT_INTERVAL"` // Interval for reporting metrics
}

// ParseConfig initializes the Config with default values,
// overrides them with command-line flags if available,
// and finally allows environment variables to set or override the configuration.
// It returns an error if environment variable parsing fails.
func ParseConfig() (*Config, error) {
	// Default settings for the service configuration.
	cfg := Config{
		ServerAddress:  defaultServerAddress,
		PollInterval:   defaultPollInterval,
		ReportInterval: defaultReportInterval,
		SigningKey:     defaultSigningKey,
	}

	// Parse command-line arguments or set default settings if no args are provided.
	parseFlagsOrSetDefault(&cfg)

	// Attempt to parse values from environment variables; if unsuccessful, return the error.
	if err := env.Parse(&cfg); err != nil {
		return nil, fmt.Errorf("failed to parse environment variables: %w", err)
	}

	return &cfg, nil
}

// parseFlagsOrSetDefault attempts to populate the Config from command-line flags.
// If no flags are provided, it retains the default values set in the configuration.
func parseFlagsOrSetDefault(cfg *Config) {
	flag.IntVar(&cfg.PollInterval, "p", cfg.PollInterval, "Interval (in seconds) for collecting metrics")
	flag.IntVar(&cfg.ReportInterval, "r", cfg.ReportInterval, "Interval (in seconds) for sending metrics")
	flag.StringVar(&cfg.ServerAddress, "a", cfg.ServerAddress, "Address of the server to connect to")
	flag.StringVar(&cfg.SigningKey, "k", cfg.SigningKey, "Signing key used for creating request signatures.")
	flag.Parse()
}
