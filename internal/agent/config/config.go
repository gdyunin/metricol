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
	defaultConfigPath     = ""
)

// Config holds the configuration settings for the application.
// It contains the server address, signing key, intervals for polling and reporting metrics,
// a rate limit for HTTP requests, and a flag for enabling or disabling pprof profiling.
type Config struct {
	ServerAddress  string `env:"ADDRESS"         json:"server_address,omitempty"`
	SigningKey     string `env:"KEY"             json:"signing_key,omitempty"`
	CryptoKey      string `env:"CRYPTO_KEY"      json:"crypto_key,omitempty"`
	ConfigPath     string `env:"CONFIG"          json:"config_path,omitempty"`
	PollInterval   int    `env:"POLL_INTERVAL"   json:"poll_interval,omitempty"`
	ReportInterval int    `env:"REPORT_INTERVAL" json:"report_interval,omitempty"`
	RateLimit      int    `env:"RATE_LIMIT"      json:"rate_limit,omitempty"`
	PprofFlag      bool   `env:"PPROF_FLAG"      json:"pprof_flag,omitempty"`
}

// ParseConfig initializes a new Config instance with default values, then overrides these values
// using command-line flags and environment variables. The function first sets the defaults,
// then calls parseFlagsOrSetDefault to parse command-line flags, and finally uses the env package to
// parse environment variables. If parsing the environment variables fails, an error is returned.
//
// Returns:
//   - *Config: A pointer to the populated Config structure.
//   - error: An error if environment variable parsing fails; otherwise, nil.
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
		ConfigPath:     defaultConfigPath,
	}

	// Parse command-line arguments or set default settings if no arguments are provided.
	parseFlagsOrSetDefault(&cfg)

	// Attempt to parse values from environment variables; if unsuccessful, return the error.
	if err := env.Parse(&cfg); err != nil {
		return nil, fmt.Errorf("failed to parse environment variables: %w", err)
	}

	if cfg.ConfigPath != defaultConfigPath {
		if err := mergeConfigFile(&cfg); err != nil {
			return nil, fmt.Errorf("failed to merge configuration file: %w", err)
		}
	}

	return &cfg, nil
}

func mergeConfigFile(cfg *Config) error {
	data, err := os.ReadFile(cfg.ConfigPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	tempCfg := Config{}
	if err := json.Unmarshal(data, &tempCfg); err != nil {
		return fmt.Errorf("cannot parse JSON config %q: %w", cfg.ConfigPath, err)
	}

	// [ДЛЯ РЕВЬЮ] Если честно, то мне очень не нравится такая реализация и она точно будет работать неправильно,
	// если пользователь самостоятельно укажет в переменных окружения или флагах параметры, совпадающие с дефолтными,
	// Но я не смог придумать что-то лучше, чтобы при этом не нужно было полностью переписывать всю остальную логику.
	if cfg.ServerAddress == defaultServerAddress && tempCfg.ServerAddress != defaultServerAddress {
		cfg.ServerAddress = tempCfg.ServerAddress
	}
	if cfg.SigningKey == defaultSigningKey && tempCfg.SigningKey != defaultSigningKey {
		cfg.SigningKey = tempCfg.SigningKey
	}
	if cfg.CryptoKey == defaultCryptoKey && tempCfg.CryptoKey != defaultCryptoKey {
		cfg.CryptoKey = tempCfg.CryptoKey
	}
	if cfg.PollInterval == defaultPollInterval && tempCfg.PollInterval != defaultPollInterval {
		cfg.PollInterval = tempCfg.PollInterval
	}
	if cfg.ReportInterval == defaultReportInterval && tempCfg.ReportInterval != defaultReportInterval {
		cfg.ReportInterval = tempCfg.ReportInterval
	}
	if cfg.RateLimit == defaultRateLimit && tempCfg.RateLimit != defaultRateLimit {
		cfg.RateLimit = tempCfg.RateLimit
	}
	if !cfg.PprofFlag && tempCfg.PprofFlag {
		cfg.PprofFlag = tempCfg.PprofFlag
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
	flag.StringVar(&cfg.ConfigPath, "c", cfg.ConfigPath, "Path to config file.")
	flag.Parse()
}
