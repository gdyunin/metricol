package server

import (
	"flag"
	"log"

	"github.com/caarlos0/env/v6"
)

type ServerConfig struct {
	ServerAddress string `env:"ADDRESS,notEmpty"`
}

func ParseServerConfig() ServerConfig {
	// Default settings.
	cfg := ServerConfig{
		ServerAddress: "localhost:8080",
	}

	// Parse values from env vars, if they exist config will be returned.
	if parseEnv(&cfg) {
		return cfg
	}

	// Parse command-line args or set default settings.
	parseFlagsOrSetDefault(&cfg)
	return cfg
}

// parseEnv try to fill got config from env and return true if success. Else return false and log error.
func parseEnv(cfg *ServerConfig) bool {
	err := env.Parse(cfg)
	if err != nil {
		log.Printf("error trying to get environment variables: %v\n, command line flags or defaults will be used", err)
		return false
	}
	return true
}

// parseFlagsOrSetDefault try to fill got config from cmd or set default.
func parseFlagsOrSetDefault(cfg *ServerConfig) {
	flag.StringVar(&cfg.ServerAddress, "a", cfg.ServerAddress, "Адрес сервера")
	flag.Parse()
}
