// Package config provides functionality for parsing and managing the server's configuration.
// Configuration settings can be set via default values, command-line flags, or environment variables.
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
	defaultServerAddress   = "localhost:8080"
	defaultStoreInterval   = 300
	defaultFileStoragePath = ""
	defaultRestoreFlag     = true
	defaultDatabaseDSN     = ""
	defaultSigningKey      = ""
	defaultPprofFlag       = false
	defaultCryptoKey       = ""
)

// Config holds the configuration for the server, including its address,
// file storage settings, database connection details, and other operational flags.
// The configuration values can be provided via environment variables, command-line flags,
// or default settings defined in the package.
type Config struct {
	ServerAddress   string `env:"ADDRESS"           json:"server_address"`
	FileStoragePath string `env:"FILE_STORAGE_PATH" json:"file_storage_path"`
	DatabaseDSN     string `env:"DATABASE_DSN"      json:"database_dsn"`
	SigningKey      string `env:"KEY"               json:"signing_key"`
	CryptoKey       string `env:"CRYPTO_KEY"        json:"crypto_key"`
	StoreInterval   int    `env:"STORE_INTERVAL"    json:"store_interval"`
	Restore         bool   `env:"RESTORE"           json:"restore"`
	PprofFlag       bool   `env:"PPROF_SERVER_FLAG" json:"pprof_flag"`
}

// ParseConfig initializes the Config with default values, overrides them with command-line flags if provided,
// and then allows environment variables to set or override the configuration.
// It returns a pointer to the Config structure or an error if environment parsing fails.
//
// Returns:
//   - *Config: A pointer to the populated Config structure.
//   - error: An error if parsing of environment variables fails.
//
// ParseConfig initializes and returns a Config struct by parsing configuration
// values from multiple sources in the following order of precedence:
// 1. Default settings defined in the code.
// 2. Configuration file, if available.
// 3. Command-line flags, which override the previous values.
// 4. Environment variables, which override all previous values.
//
// If any error occurs during parsing the configuration file or environment
// variables, it returns an error.
//
// Returns:
//   - A pointer to the populated Config struct.
//   - An error if parsing fails at any stage.
//
// ParseConfig initializes and returns a Config object by combining default values,
// configuration file settings, command-line flags, and environment variables.
//
// The function performs the following steps:
//  1. Initializes a Config object with default values.
//  2. Attempts to parse a configuration file to override default values.
//     If parsing the file fails, an error is returned.
//  3. Parses command-line flags to override existing configuration values or set defaults.
//  4. Parses environment variables to override configuration values if they are set.
//     If parsing environment variables fails, an error is returned.
//
// Returns:
// - A pointer to the populated Config object.
// - An error if any step in the configuration process fails.
func ParseConfig() (*Config, error) {
	// Default settings for the server configuration.
	cfg := Config{
		ServerAddress:   defaultServerAddress,
		StoreInterval:   defaultStoreInterval,
		FileStoragePath: defaultFileStoragePath,
		Restore:         defaultRestoreFlag,
		DatabaseDSN:     defaultDatabaseDSN,
		SigningKey:      defaultSigningKey,
		PprofFlag:       defaultPprofFlag,
		CryptoKey:       defaultCryptoKey,
	}

	if err := parseConfigFile(&cfg); err != nil {
		return nil, fmt.Errorf("error parsing config file: %w", err)
	}

	// Populate the configuration from command-line flags.
	parseFlagsOrSetDefault(&cfg)

	// Parse environment variables and override configuration values if set.
	if err := env.Parse(&cfg); err != nil {
		return nil, fmt.Errorf("error parse env variables %w", err)
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

// parseFlagsOrSetDefault populates the Config from command-line flags,
// retaining default values if flags are not provided.
// This function updates the provided Config pointer with flag values.
func parseFlagsOrSetDefault(cfg *Config) {
	flag.StringVar(&cfg.ServerAddress, "a", cfg.ServerAddress, "Address of the server")
	flag.IntVar(
		&cfg.StoreInterval,
		"i",
		cfg.StoreInterval,
		"Interval for store to fs in sec, if = 0 sync store",
	)
	flag.StringVar(&cfg.FileStoragePath, "f", cfg.FileStoragePath, "File storage path")
	flag.BoolVar(&cfg.Restore, "r", cfg.Restore, "Indicates whether restore is needed")
	flag.StringVar(&cfg.DatabaseDSN, "d", cfg.DatabaseDSN, "Database DSN")
	flag.StringVar(&cfg.SigningKey, "k", cfg.SigningKey, "Signing key for checking request signatures.")
	flag.BoolVar(&cfg.PprofFlag, "pf", cfg.PprofFlag, "Enable or disable profiling with pprof")
	flag.StringVar(&cfg.CryptoKey, "crypto-key", cfg.CryptoKey, "Path to private key file.")
	flag.Parse()
}
