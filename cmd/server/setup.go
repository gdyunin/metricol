package main

import (
	"log"
	"time"

	"github.com/gdyunin/metricol.git/internal/server/backup"
	"github.com/gdyunin/metricol.git/internal/server/backup/managers/basic"
	"github.com/gdyunin/metricol.git/internal/server/config"
	"github.com/gdyunin/metricol.git/internal/server/consume"
	"github.com/gdyunin/metricol.git/internal/server/consume/consumers/echoserver"
	"github.com/gdyunin/metricol.git/internal/server/entities"
	"github.com/gdyunin/metricol.git/internal/server/repositories"
	"github.com/gdyunin/metricol.git/pkg/logger"
	"go.uber.org/zap"
)

type loggers struct {
	repository   *zap.SugaredLogger
	configParser *zap.SugaredLogger
	consumer     *zap.SugaredLogger
}

func newLoggers(level string) *loggers {
	baseLogger, err := logger.Logger(level)
	if err != nil {
		log.Fatalf("failed to initialize logger: %v", err)
	}

	return &loggers{
		repository:   baseLogger.Named("repository"),
		configParser: baseLogger.Named("config_parser"),
		consumer:     baseLogger.Named("consumer"),
	}
}

func setupRepository(logger *zap.SugaredLogger) entities.MetricRepository {
	return repositories.NewInMemoryRepository(logger)
}

func setupConsumer(serverAddress string, repo entities.MetricRepository, logger *zap.SugaredLogger) consume.Consumer {
	return echoserver.NewEchoServer(serverAddress, repo, logger)
}

func setupBackupManager(path string, interval time.Duration, needRestore bool, repo entities.MetricRepository) backup.BackupManager {
	return basic.NewBackupManager(
		path,
		"backup.txt",
		interval,
		needRestore,
		repo,
	)
}

func setupConfig(logger *zap.SugaredLogger) *config.Config {
	appCfg, err := config.ParseConfig(logger)
	if err != nil {
		log.Fatalf("failed to parse config: %v", err)
	}
	return appCfg
}
