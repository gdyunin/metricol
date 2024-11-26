package main

import (
	"flag"
	"github.com/caarlos0/env/v6"
)

// Default settings
const (
	defaultServerAddress = "localhost:8080"
)

type config struct {
	ServerAddress string `env:"ADDRESS"`
}

func appConfig() config {
	cfg := config{}

	// Parse command-line args
	flag.StringVar(&cfg.ServerAddress, "a", defaultServerAddress, "Адрес сервера")
	flag.Parse()

	// Parse values from env vars, if they exist replace config values
	// The error is ignored as it has no effect
	// A logger could be added in the future
	_ = env.Parse(&cfg)

	return cfg
}
