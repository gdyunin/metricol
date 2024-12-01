// Package agent provides functionality to configure an agent with parameters
// that can be set via environment variables or command-line flags.
package agent

import (
	"flag"
	"log"

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
	ServerAddress  string `env:"ADDRESS,notEmpty"`         // Address of the server to connect to
	PollInterval   int    `env:"POLL_INTERVAL,notEmpty"`   // Interval for polling metrics
	ReportInterval int    `env:"REPORT_INTERVAL,notEmpty"` // Interval for reporting metrics
}

// ParseConfig initializes the Config with default values,
// overrides them with environment variables if available,
// and finally allows command-line flags to set or override the configuration.
func ParseConfig() Config {
	// Default settings for the agent configuration.
	cfg := Config{
		ServerAddress:  defaultServerAddress,
		PollInterval:   defaultPollInterval,
		ReportInterval: defaultReportInterval,
	}

	// Attempt to parse values from environment variables; if successful, return the config.
	if parseEnv(&cfg) {
		return cfg
	}

	// Parse command-line arguments or set default settings if no args are provided.
	parseFlagsOrSetDefault(&cfg)
	return cfg
}

// parseEnv attempts to populate the Config from environment variables.
// It returns true if successful, otherwise logs an error and returns false.
func parseEnv(cfg *Config) bool {
	err := env.Parse(cfg)
	if err != nil {
		log.Printf("error trying to get environment variables: %v\ncommand line flags will be used", err)
		return false
	}
	return true
}

// parseFlagsOrSetDefault attempts to populate the Config from command-line flags.
// If no flags are provided, it retains the default values set in the configuration.
func parseFlagsOrSetDefault(cfg *Config) {
	flag.IntVar(&cfg.PollInterval, "p", cfg.PollInterval, "Interval for collecting metrics")
	flag.IntVar(&cfg.ReportInterval, "r", cfg.ReportInterval, "Interval for sending metrics")
	flag.StringVar(&cfg.ServerAddress, "a", cfg.ServerAddress, "Address of the server")
	flag.Parse()
}
