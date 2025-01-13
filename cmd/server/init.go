package main

import (
	"NewNewMetricol/internal/server/config"
	"NewNewMetricol/internal/server/delivery"
	"NewNewMetricol/internal/server/repository"
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
	// LoggerNameDelivery is the logger name for the delivery layer.
	LoggerNameDelivery = "delivery"
	// LoggerNameRepository is the logger name for the repository layer.
	LoggerNameRepository = "repository"
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

func initComponentsWithShutdownActs(cfg *config.Config, logger *zap.SugaredLogger) (*delivery.EchoServer, []func()) {
	shutdownActions := make([]func(), 0)

	repo, repoShutdownFunc := initRepo(cfg, logger.Named(LoggerNameRepository))
	shutdownActions = append(shutdownActions, repoShutdownFunc)

	echoDelivery := delivery.NewEchoServer(cfg.ServerAddress, repo, logger.Named(LoggerNameDelivery))

	return echoDelivery, shutdownActions
}

func initRepo(cfg *config.Config, logger *zap.SugaredLogger) (repository.Repository, func()) {
	r := repository.NewInFileRepository(
		logger,
		cfg.FileStoragePath,
		"backup.txt",
		convert.IntegerToSeconds(cfg.StoreInterval),
		cfg.Restore,
	)

	return r, r.Shutdown
}

// setupGracefulShutdown sets up a graceful shutdown mechanism for the application.
// It listens for system interrupt and termination signals (e.g., SIGTERM or SIGINT).
// When a signal is received, it cancels the provided context and waits for the graceful shutdown timeout before exiting.
//
// Parameters:
//   - ctxCancel: The cancel function associated with the application context.
//     This function will be called to signal the application to shut down.
func setupGracefulShutdown(ctxCancel context.CancelFunc, logger *zap.SugaredLogger, shutdownActions ...func()) {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-signalChan
		logger.Info("Received termination signal (SIGTERM or SIGINT). Initiating graceful shutdown...")
		ctxCancel() // Cancel the application context.

		for _, act := range shutdownActions {
			go func(fn func()) {
				fn()
			}(act)
		}

		logger.Infof(
			"Context canceled. Allowing %d seconds for cleanup operations before forced application exit...",
			GracefulShutdownTimeout/time.Second,
		)
		time.Sleep(GracefulShutdownTimeout) // Wait for a graceful shutdown.
		logger.Warn("Timeout reached. Forcing application to exit.")
		os.Exit(0) // Exit the application.
	}()
}
