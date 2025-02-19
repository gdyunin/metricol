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
)

// Config holds the configuration for the server, including the server address.
type Config struct {
	ServerAddress   string `env:"ADDRESS"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	DatabaseDSN     string `env:"DATABASE_DSN"`
	SigningKey      string `env:"KEY"`
	StoreInterval   int    `env:"STORE_INTERVAL"`
	Restore         bool   `env:"RESTORE"`
}

// ParseConfig initializes the Config with default values,
// overrides them with command-line flags if available,
// and allows environment variables to set or override the configuration.
func ParseConfig() (*Config, error) {
	// Default settings for the server configuration.
	cfg := Config{
		ServerAddress:   defaultServerAddress,
		StoreInterval:   defaultStoreInterval,
		FileStoragePath: defaultFileStoragePath,
		Restore:         defaultRestoreFlag,
		DatabaseDSN:     defaultDatabaseDSN,
		SigningKey:      defaultSigningKey,
	}

	// Parse command-line arguments or set default settings if no args are provided.
	parseFlagsOrSetDefault(&cfg)

	// Attempt to parse values from environment variables; if unsuccessful, return the error.
	if err := env.Parse(&cfg); err != nil {
		return nil, fmt.Errorf("error parse env variables %w", err)
	}

	return &cfg, nil
}

// parseFlagsOrSetDefault populates the Config from command-line flags
// or retains the default values set in the configuration.
func parseFlagsOrSetDefault(cfg *Config) {
	flag.StringVar(&cfg.ServerAddress, "a", cfg.ServerAddress, "Address of the server")
	flag.IntVar(
		&cfg.StoreInterval,
		"i",
		cfg.StoreInterval,
		"Interval for store to fs in sec, if = 0 sync store",
	)
	flag.StringVar(&cfg.FileStoragePath, "f", cfg.FileStoragePath, "file storage path")
	flag.BoolVar(&cfg.Restore, "r", cfg.Restore, "is restore need")
	flag.StringVar(&cfg.DatabaseDSN, "d", cfg.DatabaseDSN, "database dsn")
	flag.StringVar(&cfg.SigningKey, "k", cfg.SigningKey, "Signing key used for checking request signatures.")
	flag.Parse()
}
