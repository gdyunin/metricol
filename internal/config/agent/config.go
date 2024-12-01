package agent

import (
	"flag"
	"log"

	"github.com/caarlos0/env/v6"
)

type AgentConfig struct {
	ServerAddress  string `env:"ADDRESS,notEmpty"`
	PollInterval   int    `env:"POLL_INTERVAL,notEmpty"`
	ReportInterval int    `env:"REPORT_INTERVAL,notEmpty"`
}

func ParseAgentConfig() AgentConfig {
	// Default settings.
	cfg := AgentConfig{
		ServerAddress:  "localhost:8080",
		PollInterval:   2,
		ReportInterval: 10,
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
func parseEnv(cfg *AgentConfig) bool {
	err := env.Parse(cfg)
	if err != nil {
		log.Printf("error trying to get environment variables: %v\n, command line flags will be used", err)
		return false
	}
	return true
}

// parseFlagsOrSetDefault try to fill got config from cmd or set default.
func parseFlagsOrSetDefault(cfg *AgentConfig) {
	flag.IntVar(&cfg.PollInterval, "p", cfg.PollInterval, "Интервал сбора метрик")
	flag.IntVar(&cfg.ReportInterval, "r", cfg.ReportInterval, "Интервал отправки метрик")
	flag.StringVar(&cfg.ServerAddress, "a", cfg.ServerAddress, "Адрес сервера")
	flag.Parse()
}
