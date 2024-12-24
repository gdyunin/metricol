package main

import (
	"fmt"
	"log"
	"time"

	"github.com/gdyunin/metricol.git/internal/agent/agent"
	"github.com/gdyunin/metricol.git/internal/agent/collect/collectors/mscollector"
	"github.com/gdyunin/metricol.git/internal/agent/config"
	"github.com/gdyunin/metricol.git/internal/agent/produce/produsers/rstclient"
	"github.com/gdyunin/metricol.git/pkg/agent/repositories"
	"github.com/gdyunin/metricol.git/pkg/logger"
)

func run() error {
	appCfg, err := config.ParseConfig()
	if err != nil {
		log.Fatalf("Get configuration fail: %v", err)
	}

	baseLogger, err := logger.Logger(logger.LevelINFO)
	if err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}

	repo := repositories.NewInMemoryRepository()

	collector := mscollector.NewMemStatsCollector(
		time.Duration(appCfg.PollInterval)*time.Second,
		repo,
		baseLogger.Named("collector"),
	)

	producer := rstclient.NewRestyClient(
		time.Duration(appCfg.ReportInterval)*time.Second,
		appCfg.ServerAddress,
		repo,
		baseLogger.Named("producer"),
	)

	app, err := agent.NewAgent(
		collector,
		producer,
		baseLogger.Named("app"),
		agent.WithSubscribeConsumer2Producer,
	)
	if err != nil {
		return fmt.Errorf("failed to initialize agent: %w", err)
	}

	if err := app.Start(); err != nil {
		return fmt.Errorf("failed to start agent: %w", err)
	}

	return nil
}

func main() {
	if err := run(); err != nil {
		log.Fatalf("Critical error encountered during application execution: %v", err)
	}
}
