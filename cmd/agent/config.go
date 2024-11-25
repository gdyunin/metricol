package main

import (
	"flag"
	"github.com/caarlos0/env/v6"
)

const (
	defaultPollInterval   = 2
	defaultReportInterval = 10
	defaultServerAddress  = "localhost:8080"
)

type config struct {
	PollInterval   int    `env:"POLL_INTERVAL"`
	ReportInterval int    `env:"REPORT_INTERVAL"`
	ServerAddress  string `env:"ADDRESS"`
}

func appConfig() config {
	cfg := config{
		PollInterval:   defaultPollInterval,
		ReportInterval: defaultReportInterval,
		ServerAddress:  defaultServerAddress,
	}

	flag.IntVar(&cfg.PollInterval, "p", cfg.PollInterval, "Интервал сбора метрик")
	flag.IntVar(&cfg.ReportInterval, "r", cfg.ReportInterval, "Интервал отправки метрик")
	flag.StringVar(&cfg.ServerAddress, "a", cfg.ServerAddress, "Адрес сервера")

	flag.Parse()
	env.Parse(&cfg)

	return cfg
}
