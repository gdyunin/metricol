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
	PollInterval   time.Duration `env:"POLL_INTERVAL"`
	ReportInterval time.Duration `env:"REPORT_INTERVAL"`
	ServerAddress  string        `env:"ADDRESS"`
}

func appConfig() config {
	cfg := config{
		PollInterval:   defaultPollInterval,
		ReportInterval: defaultReportInterval,
		ServerAddress:  defaultServerAddress,
	}

	flag.DurationVar(&cfg.PollInterval, "p", cfg.PollInterval, "Интервал сбора метрик")
	flag.DurationVar(&cfg.ReportInterval, "r", cfg.ReportInterval, "Интервал отправки метрик")
	flag.StringVar(&cfg.ServerAddress, "a", cfg.ServerAddress, "Адрес сервера")

	flag.Parse()
	env.Parse(&cfg)

	return cfg
}
