package main

import (
	"log"
	"time"

	"github.com/gdyunin/metricol.git/internal/agent/collect"
	"github.com/gdyunin/metricol.git/internal/agent/collect/collectors/mscollector"
	"github.com/gdyunin/metricol.git/internal/agent/config"
	"github.com/gdyunin/metricol.git/internal/agent/entities"
	"github.com/gdyunin/metricol.git/internal/agent/orchestrate"
	"github.com/gdyunin/metricol.git/internal/agent/orchestrate/orchestrators/basic"
	"github.com/gdyunin/metricol.git/internal/agent/produce"
	"github.com/gdyunin/metricol.git/internal/agent/produce/produsers/rstclient"
	"github.com/gdyunin/metricol.git/internal/agent/repositories"
	"github.com/gdyunin/metricol.git/pkg/logger"
	"go.uber.org/zap"
)

type loggers struct {
	repository   *zap.SugaredLogger
	configParser *zap.SugaredLogger
	collector    *zap.SugaredLogger
	producer     *zap.SugaredLogger
	orchestrator *zap.SugaredLogger
}

func newLoggers(level string) *loggers {
	baseLogger, err := logger.Logger(level)
	if err != nil {
		log.Fatalf("failed to initialize logger: %w", err)
	}

	return &loggers{
		repository:   baseLogger.Named("repository"),
		configParser: baseLogger.Named("config_parser"),
		collector:    baseLogger.Named("collector"),
		producer:     baseLogger.Named("producer"),
		orchestrator: baseLogger.Named("orchestrator"),
	}
}

func setupRepository(logger *zap.SugaredLogger) entities.MetricsRepository {
	return repositories.NewInMemoryRepository(logger)
}

func setupCollector(collectInterval time.Duration, repo entities.MetricsRepository, logger *zap.SugaredLogger) collect.Collector {
	return mscollector.NewMemStatsCollector(
		collectInterval,
		repo,
		logger,
	)
}

func setupProducer(produceInterval time.Duration, serverAddress string, repo entities.MetricsRepository, logger *zap.SugaredLogger) produce.Producer {
	return rstclient.NewRestyClient(
		produceInterval,
		serverAddress,
		repo,
		logger,
	)
}

func setupOrchestrator(collector collect.Collector, producer produce.Producer, logger *zap.SugaredLogger) orchestrate.Orchestrator {
	o, err := basic.NewOrchestrator(
		collector,
		producer,
		logger,
		basic.WithSubscribeConsumer2Producer,
	)
	if err != nil {
		log.Fatalf("failed to initialize orchestrate: %v", err)
	}
	return o
}

func setupConfig(logger *zap.SugaredLogger) *config.Config {
	appCfg, err := config.ParseConfig(logger)
	if err != nil {
		log.Fatalf("failed to parse config: %v", err)
	}
	return appCfg
}
