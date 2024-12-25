package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gdyunin/metricol.git/internal/server/backup/backupers/basebackup"
	"github.com/gdyunin/metricol.git/internal/server/config"
	"github.com/gdyunin/metricol.git/internal/server/consume/consumers/echoserver"
	"github.com/gdyunin/metricol.git/pkg/logger"
	"github.com/gdyunin/metricol.git/pkg/server/repositories"
)

type stopper func()

type shutdown struct {
	fn []stopper
}

func (s *shutdown) executeAll() {
	for _, f := range s.fn {
		f()
	}
}

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

	backupper := basebackup.NewBaseBackupper(
		appCfg.FileStoragePath,
		"backup.txt",
		time.Duration(appCfg.StoreInterval)*time.Second,
		appCfg.Restore,
		repo,
	)

	fmt.Printf("Конфиг бэкапа: %v", backupper)

	backupper.Restore()
	go backupper.StartBackup()

	setupGracefulShutdown(&shutdown{fn: []stopper{
		backupper.StopBackup,
	}})

	if err = consumer.StartConsume(); err != nil {
		return fmt.Errorf("failed to start the consumption process: %w", err)
	}

	return nil
}

func setupGracefulShutdown(s *shutdown) {
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-stopChan
		s.executeAll()
		os.Exit(0)
	}()
}
