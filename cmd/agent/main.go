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

// run sets up and starts the application by initializing its components:
// the repository, collector, producer, and agent.
// It returns detailed errors if any step fails during initialization or execution.
func run() error {
	appCfg, err := config.ParseConfig()
	if err != nil {
		log.Fatalf("Get configuration fail: %v", err)
	}

	// Initialize the base logger with the INFO level for consistent logging throughout the application.
	baseLogger, err := logger.Logger(logger.LevelINFO)
	if err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}

	// Create an in-memory repository for storing metrics and configuration data.
	repo := repositories.NewInMemoryRepository()

	// Initialize the memory statistics collector using the configuration parser and in-memory repository.
	collector := mscollector.NewMemStatsCollector(
		time.Duration(appCfg.PollInterval)*time.Second,
		repo,
		baseLogger.Named("collector"),
	)

	// Initialize the REST client producer using the configuration parser and in-memory repository.
	producer := rstclient.NewRestyClient(
		time.Duration(appCfg.ReportInterval)*time.Second,
		appCfg.ServerAddress,
		repo,
		baseLogger.Named("producer"),
	)

	// Create a new agent instance with the collector, producer, and logging setup.
	app, err := agent.NewAgent(
		collector,
		producer,
		baseLogger.Named("app"),
		agent.WithSubscribeConsumer2Producer,
	)
	if err != nil {
		return fmt.Errorf("failed to initialize agent: %w", err)
	}

	time.Sleep(3 * time.Second)

	// Start the agent and return any error encountered during its execution.
	if err := app.Start(); err != nil {
		return fmt.Errorf("failed to start agent: %w", err)
	}

	return nil
}

// main serves as the application's entry point.
// It calls the run function to start the application and logs any critical errors that prevent execution.
func main() {
	if err := run(); err != nil {
		log.Fatalf("Critical error encountered during application execution: %v", err)
	}
}
