// Package config provides functionality for parsing and managing the server's configuration.
// Configuration settings can be set via default values, command-line flags, or environment variables.
package config

import (
	"flag"
	"fmt"

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
	ServerAddress   string `env:"ADDRESS"`           // ServerAddress is the address on which the server listens.
	FileStoragePath string `env:"FILE_STORAGE_PATH"` // FileStoragePath is the file system path for storage.
	DatabaseDSN     string `env:"DATABASE_DSN"`      // DatabaseDSN is the Data Source Name for the database connection.
	SigningKey      string `env:"KEY"`               // SigningKey is used for checking request signatures.
	StoreInterval   int    `env:"STORE_INTERVAL"`    // StoreInterval is the interval (in seconds) for storing data.
	Restore         bool   `env:"RESTORE"`           // Restore indicates whether to restore previous state on startup.
	PprofFlag       bool   `env:"PPROF_SERVER_FLAG"` // PprofFlag toggles the use of pprof profiling.
	CryptoKey       string `env:"CRYPTO_KEY"`        // Path to private key file.
}

// ParseConfig initializes the Config with default values, overrides them with command-line flags if provided,
// and then allows environment variables to set or override the configuration.
// It returns a pointer to the Config structure or an error if environment parsing fails.
//
// Returns:
//   - *Config: A pointer to the populated Config structure.
//   - error: An error if parsing of environment variables fails.
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
		CryptoKey:      defaultCryptoKey,
	}

	// Populate the configuration from command-line flags.
	parseFlagsOrSetDefault(&cfg)

	// Parse environment variables and override configuration values if set.
	if err := env.Parse(&cfg); err != nil {
		return nil, fmt.Errorf("error parse env variables %w", err)
	}

	return &cfg, nil
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
