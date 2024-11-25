package main

import (
	"flag"
	"github.com/caarlos0/env/v6"
)

const (
	defaultServerAddress = "localhost:8080"
)

type config struct {
	serverAddress string `env:"ADDRESS,require"`
}

func appConfig() config {
	cfg := config{
		serverAddress: defaultServerAddress,
	}

	flag.StringVar(&cfg.serverAddress, "a", cfg.serverAddress, "Адрес сервера")

	flag.Parse()
	env.Parse(&cfg)

	return cfg
}
