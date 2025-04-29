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
	defaultConfigPath      = ""
)

// Config holds the configuration for the server, including its address,
// file storage settings, database connection details, and other operational flags.
// The configuration values can be provided via environment variables, command-line flags,
// or default settings defined in the package.
type Config struct {
	ServerAddress   string `env:"ADDRESS"           json:"server_address,omitempty"`
	FileStoragePath string `env:"FILE_STORAGE_PATH" json:"file_storage_path,omitempty"`
	DatabaseDSN     string `env:"DATABASE_DSN"      json:"database_dsn,omitempty"`
	SigningKey      string `env:"KEY"               json:"signing_key,omitempty"`
	CryptoKey       string `env:"CRYPTO_KEY"        json:"crypto_key,omitempty"`
	ConfigPath      string `env:"CONFIG"            json:"config_path,omitempty"`
	StoreInterval   int    `env:"STORE_INTERVAL"    json:"store_interval,omitempty"`
	Restore         bool   `env:"RESTORE"           json:"restore,omitempty"`
	PprofFlag       bool   `env:"PPROF_SERVER_FLAG" json:"pprof_flag,omitempty"`
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
		CryptoKey:       defaultCryptoKey,
		ConfigPath:      defaultConfigPath,
	}

	// Populate the configuration from command-line flags.
	parseFlagsOrSetDefault(&cfg)

	// Parse environment variables and override configuration values if set.
	if err := env.Parse(&cfg); err != nil {
		return nil, fmt.Errorf("error parse env variables %w", err)
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
	if cfg.FileStoragePath == defaultFileStoragePath && tempCfg.FileStoragePath != defaultFileStoragePath {
		cfg.FileStoragePath = tempCfg.FileStoragePath
	}
	if cfg.DatabaseDSN == defaultDatabaseDSN && tempCfg.DatabaseDSN != defaultDatabaseDSN {
		cfg.DatabaseDSN = tempCfg.DatabaseDSN
	}
	if cfg.SigningKey == defaultSigningKey && tempCfg.SigningKey != defaultSigningKey {
		cfg.SigningKey = tempCfg.SigningKey
	}
	if cfg.CryptoKey == defaultCryptoKey && tempCfg.CryptoKey != defaultCryptoKey {
		cfg.CryptoKey = tempCfg.CryptoKey
	}
	if cfg.StoreInterval == defaultStoreInterval && tempCfg.StoreInterval != defaultStoreInterval {
		cfg.StoreInterval = tempCfg.StoreInterval
	}
	if cfg.Restore && !tempCfg.Restore {
		cfg.Restore = tempCfg.Restore
	}
	if !cfg.PprofFlag && tempCfg.PprofFlag {
		cfg.PprofFlag = tempCfg.PprofFlag
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
	flag.StringVar(&cfg.ConfigPath, "c", cfg.ConfigPath, "Path to config file.")
	flag.Parse()
}
