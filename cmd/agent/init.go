package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gdyunin/metricol.git/internal/agent/agent"
	"github.com/gdyunin/metricol.git/internal/agent/config"
	"github.com/gdyunin/metricol.git/pkg/convert"
	"github.com/gdyunin/metricol.git/pkg/logging"

	"go.uber.org/zap"
)

const (
	// LoggerNameAgent is the logger name for the main agent.
	loggerNameAgent = "agent"
	// LoggerNameGracefulShutdown is the logger name for the graceful shutdown events.
	loggerNameGracefulShutdown = "graceful_shutdown"
	// GracefulShutdownTimeout is the time to wait for ongoing tasks to complete during shutdown.
	gracefulShutdownTimeout = 5 * time.Second
)

// mainContext initializes the main application context with a cancel function.
func mainContext() (context.Context, context.CancelFunc) {
	return signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
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

// initAgent initializes the agent, including the
// metrics collectors, metrics senders.
func initAgent(cfg *config.Config, logger *zap.SugaredLogger) *agent.Agent {
	return agent.NewAgent(
		convert.IntegerToSeconds(cfg.PollInterval),
		convert.IntegerToSeconds(cfg.ReportInterval),
		logger.Named(loggerNameAgent),
		cfg.RateLimit,
		cfg.ServerAddress,
		cfg.SigningKey,
	)
}

// setupGracefulShutdown establishes a mechanism to gracefully shut down the
// application. This function reacts to the cancellation of the provided
// context, which can be triggered by external components handling system
// interrupt and termination signals such as SIGTERM or SIGINT. It allows a
// configured timeout for cleanup operations before forcefully terminating the
// application.
//
// Parameters:
//   - ctx: The context representing the application's lifecycle. Cancellation
//     of this context initiates the shutdown process.
//     This function will be called to signal the application to shut down.
func setupGracefulShutdown(ctx context.Context, logger *zap.SugaredLogger) {
	go func() {
		<-ctx.Done()
		logger.Infof(
			"Context canceled. Allowing %d seconds for cleanup operations before forced application exit...",
			gracefulShutdownTimeout/time.Second,
		)
		time.Sleep(gracefulShutdownTimeout) // Wait for a graceful shutdown.
		logger.Warn("Timeout reached. Forcing application to exit.")
		os.Exit(1) // Exit the application.
	}()
}
