package main

import (
	"github.com/caarlos0/env/v6"
	"github.com/spf13/pflag"
)

const (
	defaultServerAddress = "localhost:8080"
)

type config struct {
	serverAddress string `env:"SERVER_ADDRESS"`
}

func appConfig() config {
	cfg := config{
		serverAddress: defaultServerAddress,
	}

	pflag.StringVar(&cfg.serverAddress, "a", cfg.serverAddress, "Адрес сервера")

	pflag.Parse()
	env.Parse(&cfg)

	return cfg
}
