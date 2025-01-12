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
	"time"

	"go.uber.org/zap"
)

const (
	// LoggerNameCollector is the logger name for the metrics collector.
	LoggerNameCollector = "collector"
	// LoggerNameSender is the logger name for the metrics sender.
	LoggerNameSender = "sender"
	// LoggerNameAgent is the logger name for the main agent.
	LoggerNameAgent = "agent"
	// LoggerNameGracefulShutdown is the logger name for the graceful shutdown events.
	LoggerNameGracefulShutdown = "graceful_shutdown"
	// GracefulShutdownTimeout is the time to wait for ongoing tasks to complete during shutdown.
	GracefulShutdownTimeout = 5 * time.Second
)

// mainContext initializes the main application context with a cancel function.
func mainContext() (context.Context, context.CancelFunc) {
	return context.WithCancel(context.Background())
}

// loggerWithSyncFunc initializes and returns a SugaredLogger instance
// along with a function to flush buffered logs.
func loggerWithSyncFunc() (*zap.SugaredLogger, func()) {
	l := logging.Logger(logging.LevelINFO)
	syncFunc := func() { _ = l.Sync() }
	return l, syncFunc
}

// loadConfig parses the configuration file and returns a Config instance.
// Logs a fatal error if parsing fails.
func loadConfig() *config.Config {
	cfg, err := config.ParseConfig()
	if err != nil {
		log.Fatalf("Error occurred while parsing the application configuration: %v", err)
	}
	return cfg
}

// initComponents initializes the main components of the agent, including the
// metrics collector, metrics sender, and the agent itself.
func initComponents(cfg *config.Config, logger *zap.SugaredLogger) *agent.Agent {
	collector := collect.NewCollector(logger.Named(LoggerNameCollector))
	sender := send.NewMetricsSender(cfg.ServerAddress, logger.Named(LoggerNameSender))

	return agent.NewAgent(
		collector,
		sender,
		convert.IntegerToSeconds(cfg.PollInterval),
		convert.IntegerToSeconds(cfg.ReportInterval),
		logger.Named(LoggerNameAgent),
	)
}

// setupGracefulShutdown sets up a graceful shutdown mechanism for the application.
// It listens for system interrupt and termination signals (e.g., SIGTERM or SIGINT).
// When a signal is received, it cancels the provided context and waits for the graceful shutdown timeout before exiting.
//
// Parameters:
//   - ctxCancel: The cancel function associated with the application context.
//     This function will be called to signal the application to shut down.
func setupGracefulShutdown(ctxCancel context.CancelFunc, logger *zap.SugaredLogger) {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-signalChan
		logger.Info("Received termination signal (SIGTERM or SIGINT). Initiating graceful shutdown...")
		ctxCancel() // Cancel the application context.
		logger.Infof(
			"Context canceled. Allowing %d seconds for cleanup operations before forced application exit...",
			GracefulShutdownTimeout/time.Second,
		)
		time.Sleep(GracefulShutdownTimeout) // Wait for a graceful shutdown.
		logger.Warn("Timeout reached. Forcing application to exit.")
		os.Exit(0) // Exit the application.
	}()
}
