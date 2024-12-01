// Package server provides functionality to configure a server with parameters
// that can be set via environment variables or command-line flags.
package server

import (
	"flag"
	"log"

	"github.com/caarlos0/env/v6"
)

// All default settings.
const (
	defaultServerAddress = "localhost:8080"
)

// Config holds the configuration for the server, including the server address.
type Config struct {
	ServerAddress string `env:"ADDRESS,notEmpty"` // Server address to connect to
}

// ParseConfig initializes the Config with default values,
// overrides them with environment variables if available,
// and allows command-line flags to set or override the configuration.
func ParseConfig() Config {
	// Default settings for the server configuration.
	cfg := Config{
		ServerAddress: defaultServerAddress,
	}

	// Attempt to parse values from environment variables; if successful, return the config.
	if parseEnv(&cfg) {
		return cfg
	}

	// Parse command-line arguments or set default settings if no args are provided.
	parseFlagsOrSetDefault(&cfg)
	return cfg
}

// parseEnv populates the Config from environment variables.
// It returns true if successful, otherwise logs an error and returns false.
func parseEnv(cfg *Config) bool {
	err := env.Parse(cfg)
	if err != nil {
		log.Printf("error trying to get environment variables: %v\ncommand line flags or defaults will be used", err)
		return false
	}
	return true
}

// parseFlagsOrSetDefault populates the Config from command-line flags
// or retains the default values set in the configuration.
func parseFlagsOrSetDefault(cfg *Config) {
	flag.StringVar(&cfg.ServerAddress, "a", cfg.ServerAddress, "Address of the server")
	flag.Parse()
}
