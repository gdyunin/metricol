package main

import (
	"flag"
	"github.com/caarlos0/env/v6"
)

const (
	defaultServerAddress = "localhost:8080"
)

type config struct {
	ServerAddress string `env:"ADDRESS"`
}

func appConfig() config {
	cfg := config{
		ServerAddress: defaultServerAddress,
	}

	flag.StringVar(&cfg.ServerAddress, "a", cfg.ServerAddress, "Адрес сервера")

	flag.Parse()
	env.Parse(&cfg)

	return cfg
}
