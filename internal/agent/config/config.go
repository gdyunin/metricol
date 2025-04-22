// Package config provides functionality for parsing and handling configuration settings for the application.
// The configuration can be provided via default settings, command-line flags, or environment variables.
// This package defines a Config structure that holds settings such as server address, polling intervals,
// reporting intervals, signing keys, rate limits, and profiling flags.
package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/caarlos0/env/v6"
)

// All default settings.
const (
	defaultServerAddress  = "localhost:8080"
	defaultPollInterval   = 2
	defaultReportInterval = 10
	defaultSigningKey     = ""
	defaultRateLimit      = 3
	defaultPprofFlag      = false
	defaultCryptoKey      = ""
)

// Config holds the configuration settings for the application.
// It contains the server address, signing key, intervals for polling and reporting metrics,
// a rate limit for HTTP requests, and a flag for enabling or disabling pprof profiling.
type Config struct {
	ServerAddress  string `env:"ADDRESS"         json:"server_address"`
	SigningKey     string `env:"KEY"             json:"signing_key"`
	CryptoKey      string `env:"CRYPTO_KEY"      json:"crypto_key"`
	PollInterval   int    `env:"POLL_INTERVAL"   json:"poll_interval"`
	ReportInterval int    `env:"REPORT_INTERVAL" json:"report_interval"`
	RateLimit      int    `env:"RATE_LIMIT"      json:"rate_limit"`
	PprofFlag      bool   `env:"PPROF_FLAG"      json:"pprof_flag"`
}

// ParseConfig initializes a new Config instance with default values, then overrides these values
// using command-line flags and environment variables. The function first sets the defaults,
// then calls parseFlagsOrSetDefault to parse command-line flags, and finally uses the env package to
// parse environment variables. If parsing the environment variables fails, an error is returned.
//
// Returns:
//   - *Config: A pointer to the populated Config structure.
//   - error: An error if environment variable parsing fails; otherwise, nil.
//
// ParseConfig initializes and returns a Config object by parsing configuration
// from multiple sources in the following order of precedence:
// 1. Default settings are applied.
// 2. Configuration file values are parsed and override defaults.
// 3. Command-line arguments are parsed and override previous values.
// 4. Environment variables are parsed and override previous values.
//
// If any error occurs during the parsing of the configuration file or
// environment variables, it returns an error.
//
// Returns:
// - A pointer to the populated Config object.
// - An error if parsing fails at any stage.
func ParseConfig() (*Config, error) {
	// Default settings for the service configuration.
	cfg := Config{
		ServerAddress:  defaultServerAddress,
		PollInterval:   defaultPollInterval,
		ReportInterval: defaultReportInterval,
		SigningKey:     defaultSigningKey,
		RateLimit:      defaultRateLimit,
		PprofFlag:      defaultPprofFlag,
		CryptoKey:      defaultCryptoKey,
	}

	if err := parseConfigFile(&cfg); err != nil {
		return nil, fmt.Errorf("error parsing config file: %w", err)
	}

	// Parse command-line arguments or set default settings if no arguments are provided.
	parseFlagsOrSetDefault(&cfg)

	// Attempt to parse values from environment variables; if unsuccessful, return the error.
	if err := env.Parse(&cfg); err != nil {
		return nil, fmt.Errorf("failed to parse environment variables: %w", err)
	}

	return &cfg, nil
}

// parseConfigFile parses the configuration for the application.
// It first attempts to read the configuration file path from the command-line
// flag "-c". If the flag is not provided, it falls back to the "CONFIG"
// environment variable. If a configuration file path is found, it reads the
// file and unmarshals its JSON content into the provided Config struct.
//
// Parameters:
//   - cfg: A pointer to a Config struct where the parsed configuration will be stored.
//
// Returns:
//   - An error if there is an issue with parsing command-line flags, reading the
//     configuration file, or unmarshaling the JSON content. Returns nil if the
//     configuration is successfully parsed.
func parseConfigFile(cfg *Config) error {
	var configPath string

	fs := flag.NewFlagSet("config", flag.ContinueOnError)
	fs.StringVar(&configPath, "c", "", "Path to JSON config file")
	if err := fs.Parse(os.Args[1:]); err != nil {
		return fmt.Errorf("error parsing command-line flags: %w", err)
	}

	if configPath == "" {
		configPath = os.Getenv("CONFIG")
	}

	if configPath != "" {
		data, err := os.ReadFile(configPath)
		if err != nil {
			return fmt.Errorf("failed to read config file: %w", err)
		}

		if err := json.Unmarshal(data, cfg); err != nil {
			return fmt.Errorf("cannot parse JSON config %q: %w", configPath, err)
		}
	}
	return nil
}

// parseFlagsOrSetDefault populates the Config structure with values provided as command-line flags.
// If a flag is not provided, the default value remains unchanged. This function updates the provided
// Config pointer with the flag values.
//
// Parameters:
//   - cfg: A pointer to the Config structure to be populated with flag values.
func parseFlagsOrSetDefault(cfg *Config) {
	flag.IntVar(&cfg.PollInterval, "p", cfg.PollInterval, "Interval (in seconds) for collecting metrics.")
	flag.IntVar(&cfg.ReportInterval, "r", cfg.ReportInterval, "Interval (in seconds) for sending metrics.")
	flag.StringVar(&cfg.ServerAddress, "a", cfg.ServerAddress, "Address of the server to connect to.")
	flag.StringVar(&cfg.SigningKey, "k", cfg.SigningKey, "Signing key used for creating request signatures.")
	flag.IntVar(&cfg.RateLimit, "l", cfg.RateLimit, "Maximum rate for sending HTTP requests per interval.")
	flag.BoolVar(&cfg.PprofFlag, "pf", cfg.PprofFlag, "Enable or disable profiling with pprof.")
	flag.StringVar(&cfg.CryptoKey, "crypto-key", cfg.CryptoKey, "Path to public key file.")
	flag.Parse()
}
