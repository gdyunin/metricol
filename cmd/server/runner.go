package main

import (
	"fmt"

	"github.com/gdyunin/metricol.git/internal/server/config"
	"github.com/gdyunin/metricol.git/internal/server/consume/consumers/echoserver"
	"github.com/gdyunin/metricol.git/pkg/logger"
	"github.com/gdyunin/metricol.git/pkg/server/repositories"
)

func run() error {
	baseInfoLogger, err := logger.Logger(logger.LevelINFO)
	if err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}

	appCfg, err := config.ParseConfig()
	if err != nil {
		return fmt.Errorf("failed to parse config: %w", err)
	}

	repo := repositories.NewInMemoryRepository()

	consumer := echoserver.NewEchoServer(appCfg.ServerAddress, repo, baseInfoLogger)

	if err = consumer.StartConsume(); err != nil {
		return fmt.Errorf("failed to start the consumption process: %w", err)
	}
	return nil
}