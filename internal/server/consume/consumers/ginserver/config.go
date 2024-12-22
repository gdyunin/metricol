package ginserver

import (
	"flag"
	"fmt"

	"github.com/caarlos0/env/v6"
)

// Default server configuration settings.
const (
	defaultServerAddress = "localhost:8080" // Default server address if none is provided.
)

// GinServerConfig holds the configuration settings for the Gin server.
type GinServerConfig struct {
	// ServerAddress specifies the address where the server listens for incoming requests.
	ServerAddress string `env:"ADDRESS"`
}

// ParseConfig loads the server configuration by parsing command-line flags and environment variables.
// It initializes a GinServerConfig structure with default values, updates it based on the provided flags,
// and then applies any relevant environment variables.
//
// Returns:
// - *GinServerConfig: A pointer to the populated server configuration.
// - error: An error if any issues occur during environment variable parsing.
func ParseConfig() (*GinServerConfig, error) {
	// Initialize configuration with default values.
	cfg := GinServerConfig{
		ServerAddress: defaultServerAddress,
	}

	// Parse command-line flags or use defaults if no flags are provided.
	parseFlagsOrSetDefault(&cfg)

	// Parse environment variables and update the configuration.
	if err := env.Parse(&cfg); err != nil {
		return nil, fmt.Errorf("error parsing environment variables: %w", err)
	}

	return &cfg, nil
}

// parseFlagsOrSetDefault parses command-line flags and updates the configuration.
// If no flags are provided, the default values remain unchanged.
func parseFlagsOrSetDefault(cfg *GinServerConfig) {
	// Define a command-line flag for the server address, defaulting to the current configuration value.
	flag.StringVar(&cfg.ServerAddress, "a", cfg.ServerAddress, "Server address to bind and listen on.")
	// Parse all provided command-line flags.
	flag.Parse()
}
