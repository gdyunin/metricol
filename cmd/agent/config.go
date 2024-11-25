package main

import (
	"github.com/caarlos0/env/v6"
	"github.com/spf13/pflag"
	"time"
)

const (
	defaultPollInterval   = 2 * time.Second
	defaultReportInterval = 10 * time.Second
	defaultServerAddress  = "localhost:8080"
)

type config struct {
	pollInterval   time.Duration `env:"POLL_INTERVAL"`
	reportInterval time.Duration `env:"REPORT_INTERVAL"`
	serverAddress  string        `env:"SERVER_ADDRESS"`
}

func appConfig() config {
	cfg := config{
		pollInterval:   defaultPollInterval,
		reportInterval: defaultReportInterval,
		serverAddress:  defaultServerAddress,
	}

	pflag.DurationVar(&cfg.pollInterval, "p", cfg.pollInterval, "Интервал сбора метрик")
	pflag.DurationVar(&cfg.reportInterval, "r", cfg.reportInterval, "Интервал отправки метрик")
	pflag.StringVar(&cfg.serverAddress, "a", cfg.serverAddress, "Адрес сервера")

	pflag.Parse()
	env.Parse(&cfg)

	return cfg
}
