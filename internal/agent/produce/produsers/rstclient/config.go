package rstclient

import (
	"flag"
	"fmt"

	"github.com/caarlos0/env/v6"
)

// Default configuration settings.
const (
	defaultServerAddress  = "localhost:8080" // Default server address.
	defaultReportInterval = 10               // Default interval for reporting metrics in seconds.
)

// RestyClientConfig holds the configuration for the Resty client.
type RestyClientConfig struct {
	ServerAddress  string `env:"ADDRESS"`         // Address of the server to connect to.
	ReportInterval int    `env:"REPORT_INTERVAL"` // Interval for reporting metrics in seconds.
}

// ParseConfig initializes the configuration with default values,
// overrides them with environment variables if available,
// and finally allows command-line flags to further set or override the configuration values.
func ParseConfig() (*RestyClientConfig, error) {
	// Initialize configuration with default values.
	cfg := RestyClientConfig{
		ServerAddress:  defaultServerAddress,
		ReportInterval: defaultReportInterval,
	}

	// Parse command-line flags to potentially override default values.
	parseFlagsOrSetDefault(&cfg)

	// Attempt to override with environment variables.
	if err := env.Parse(&cfg); err != nil {
		return nil, fmt.Errorf("failed to parse environment variables: %w", err)
	}

	return &cfg, nil
}

// parseFlagsOrSetDefault populates the configuration values from command-line flags,
// or retains the default values if no flags are provided.
func parseFlagsOrSetDefault(cfg *RestyClientConfig) {
	flag.IntVar(&cfg.ReportInterval, "r", cfg.ReportInterval, "Interval for sending metrics in seconds.")
	flag.StringVar(&cfg.ServerAddress, "a", cfg.ServerAddress, "Address of the server to connect to.")
	flag.Parse()
}
