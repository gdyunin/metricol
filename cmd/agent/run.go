package main

import (
	"fmt"
	"time"

	"github.com/gdyunin/metricol.git/internal/agent/collect/collectors/mscollector"
	"github.com/gdyunin/metricol.git/internal/agent/config"
	"github.com/gdyunin/metricol.git/internal/agent/orchestrate/orchestrators/basic"
	"github.com/gdyunin/metricol.git/internal/agent/produce/produsers/rstclient"
	"github.com/gdyunin/metricol.git/internal/agent/repositories"
	"github.com/gdyunin/metricol.git/pkg/logger"
)

// run initializes and executes the application logic. It configures the logger, parses the configuration,
// sets up necessary components, and starts the application.
func run() error {
	// Initialize the logger with INFO level.
	baseLogger, err := logger.Logger(logger.LevelINFO)
	if err != nil {
		// Return an error if the logger initialization fails.
		return fmt.Errorf("failed to initialize logger: %w", err)
	}

	// Parse the application configuration from environment variables or defaults.
	appCfg, err := config.ParseConfig()
	if err != nil {
		// Return an error if the configuration parsing fails.
		return fmt.Errorf("failed to parse config: %w", err)
	}

	// Create an in-memory repository for storing metrics.
	repo := repositories.NewInMemoryRepository()

	// Initialize a memory statistics collectors with the configured polling interval.
	collector := mscollector.NewMemStatsCollector(
		time.Duration(appCfg.PollInterval)*time.Second, // Convert poll interval to time.Duration.
		repo,                           // Repository to store collected metrics.
		baseLogger.Named("collectors"), // Logger with a named scope for the collectors.
	)

	// Initialize a REST client producer to send metrics to the server at the configured report interval.
	producer := rstclient.NewRestyClient(
		time.Duration(appCfg.ReportInterval)*time.Second, // Convert report interval to time.Duration.
		appCfg.ServerAddress,         // Server address for the REST client.
		repo,                         // Repository to fetch metrics for sending.
		baseLogger.Named("producer"), // Logger with a named scope for the producer.
	)

	// Create a new orchestrate to manage the collectors and producer.
	app, err := basic.NewOrchestrator(
		collector,                            // The metrics collectors component.
		producer,                             // The metrics producer component.
		baseLogger.Named("orchestrate"),      // Logger with a named scope for the orchestrate.
		basic.WithSubscribeConsumer2Producer, // Additional configuration option for the orchestrate.
	)
	if err != nil {
		// Return an error if the orchestrate initialization fails.
		return fmt.Errorf("failed to initialize orchestrate: %w", err)
	}

	// StartAll the orchestrate, which begins collecting and producing metrics.
	if err = app.StartAll(); err != nil {
		// Return an error if the orchestrate fails to start.
		return fmt.Errorf("failed to start orchestrate: %w", err)
	}

	// Return nil to indicate successful execution.
	return nil
}
