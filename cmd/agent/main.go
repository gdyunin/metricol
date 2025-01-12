package main

import (
	"NewNewMetricol/internal/agent/agent"
	"NewNewMetricol/internal/agent/collect"
	"NewNewMetricol/internal/agent/config"
	"NewNewMetricol/internal/agent/send"
	"NewNewMetricol/pkg/convert"
	"NewNewMetricol/pkg/logging"
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	baseLogger, err := logging.Logger(logging.LevelINFO)
	if err != nil {
		log.Fatalf("Error occurred while getting base logger: %v", err)
	}
	defer func() { _ = baseLogger.Sync() }()

	appCfg, err := config.ParseConfig()
	if err != nil {
		log.Fatalf("Error occurred while parsing the application configuration: %v", err)
	}

	collector := collect.NewCollector(baseLogger.Named("collector"))
	sender := send.NewMetricsSender(appCfg.ServerAddress, baseLogger.Named("sender"))
	newAgent := agent.NewAgent(
		collector,
		sender,
		convert.IntegerToSeconds(appCfg.PollInterval),
		convert.IntegerToSeconds(appCfg.ReportInterval),
		baseLogger.Named("agent"),
	)

	setupGracefulShutdown(cancel, baseLogger.Named("graceful_shutdown"))

	newAgent.Start(ctx)
}

func setupGracefulShutdown(ctxCancel func(), logger *zap.SugaredLogger) {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		sig := <-signalChan
		logger.Infof("Received shutdown signal: %s. Initiating graceful shutdown.", sig)
		ctxCancel()
	}()
}
