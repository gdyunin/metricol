package main

import (
	"flag"
	"github.com/caarlos0/env/v6"
	"time"
)

const (
	defaultPollInterval   = 2 * time.Second
	defaultReportInterval = 10 * time.Second
	defaultServerAddress  = "localhost:8080"
)

type config struct {
	pollInterval   time.Duration `env:"POLL_INTERVAL,require"`
	reportInterval time.Duration `env:"REPORT_INTERVAL,require"`
	serverAddress  string        `env:"ADDRESS,require"`
}

func appConfig() config {
	cfg := config{
		pollInterval:   defaultPollInterval,
		reportInterval: defaultReportInterval,
		serverAddress:  defaultServerAddress,
	}

	flag.DurationVar(&cfg.pollInterval, "p", cfg.pollInterval, "Интервал сбора метрик")
	flag.DurationVar(&cfg.reportInterval, "r", cfg.reportInterval, "Интервал отправки метрик")
	flag.StringVar(&cfg.serverAddress, "a", cfg.serverAddress, "Адрес сервера")

	flag.Parse()
	env.Parse(&cfg)

	return cfg
}
