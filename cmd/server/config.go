package main

import (
	"flag"
)

const (
	defaultServerAddress = "localhost:8080"
)

type config struct {
	serverAddress string `env:"ADDRESS"`
}

func appConfig() config {
	cfg := config{
		serverAddress: defaultServerAddress,
	}

	flag.StringVar(&cfg.serverAddress, "a", cfg.serverAddress, "Адрес сервера")

	flag.Parse()
	//env.Parse(&cfg)

	return cfg
}
