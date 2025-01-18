package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gdyunin/metricol.git/internal/agent/agent"
	"github.com/gdyunin/metricol.git/internal/agent/collect"
	"github.com/gdyunin/metricol.git/internal/agent/config"
	"github.com/gdyunin/metricol.git/internal/agent/send"
	"github.com/gdyunin/metricol.git/pkg/convert"
	"github.com/gdyunin/metricol.git/pkg/logging"

	"go.uber.org/zap"
)

const (
	// LoggerNameCollector is the logger name for the metrics collector.
	loggerNameCollector = "collector"
	// LoggerNameSender is the logger name for the metrics sender.
	loggerNameSender = "sender"
	// LoggerNameAgent is the logger name for the main agent.
	loggerNameAgent = "agent"
	// LoggerNameGracefulShutdown is the logger name for the graceful shutdown events.
	loggerNameGracefulShutdown = "graceful_shutdown"
	// GracefulShutdownTimeout is the time to wait for ongoing tasks to complete during shutdown.
	gracefulShutdownTimeout = 5 * time.Second
)

// mainContext initializes the main application context with a cancel function.
func mainContext() (context.Context, context.CancelFunc) {
	return context.WithCancel(context.Background())
}

// baseLogger initializes and returns a SugaredLogger instance
// along with a function to flush buffered logs.
func baseLogger() *zap.SugaredLogger {
	return logging.Logger(logging.LevelINFO)
}

// loadConfig parses the application's configuration file.
//
// Returns:
//   - *config.Config: The parsed configuration.
//   - error: An error if parsing fails.
func loadConfig() (*config.Config, error) {
	cfg, err := config.ParseConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}
	return cfg, nil
}

// initComponents initializes the main components of the agent, including the
// metrics collector, metrics sender, and the agent itself.
func initComponents(cfg *config.Config, logger *zap.SugaredLogger) *agent.Agent {
	collector := collect.NewCollector(logger.Named(loggerNameCollector))
	sender := send.NewMetricsSender(cfg.ServerAddress, logger.Named(loggerNameSender))

	return agent.NewAgent(
		collector,
		sender,
		convert.IntegerToSeconds(cfg.PollInterval),
		convert.IntegerToSeconds(cfg.ReportInterval),
		logger.Named(loggerNameAgent),
	)
}

// setupGracefulShutdown sets up a graceful shutdown mechanism for the
// application. It listens for system interrupt and termination signals (e.g.,
// SIGTERM or SIGINT). When a signal is received, it cancels the provided context
// and waits for the graceful shutdown timeout before exiting.
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
			gracefulShutdownTimeout/time.Second,
		)
		time.Sleep(gracefulShutdownTimeout) // Wait for a graceful shutdown.
		logger.Warn("Timeout reached. Forcing application to exit.")
		os.Exit(0) // Exit the application.
	}()
}
