package main

import (
	"fmt"
	"time"

	"github.com/gdyunin/metricol.git/internal/agent/agent"
	"github.com/gdyunin/metricol.git/internal/agent/collect/collectors/mscollector"
	"github.com/gdyunin/metricol.git/internal/agent/config"
	"github.com/gdyunin/metricol.git/internal/agent/produce/produsers/rstclient"
	"github.com/gdyunin/metricol.git/pkg/agent/repositories"
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

	// Initialize a memory statistics collector with the configured polling interval.
	collector := mscollector.NewMemStatsCollector(
		time.Duration(appCfg.PollInterval)*time.Second, // Convert poll interval to time.Duration.
		repo,                          // Repository to store collected metrics.
		baseLogger.Named("collector"), // Logger with a named scope for the collector.
	)

	// Initialize a REST client producer to send metrics to the server at the configured report interval.
	producer := rstclient.NewRestyClient(
		time.Duration(appCfg.ReportInterval)*time.Second, // Convert report interval to time.Duration.
		appCfg.ServerAddress,         // Server address for the REST client.
		repo,                         // Repository to fetch metrics for sending.
		baseLogger.Named("producer"), // Logger with a named scope for the producer.
	)

	// Create a new agent to manage the collector and producer.
	app, err := agent.NewAgent(
		collector,                            // The metrics collector component.
		producer,                             // The metrics producer component.
		baseLogger.Named("agent"),            // Logger with a named scope for the agent.
		agent.WithSubscribeConsumer2Producer, // Additional configuration option for the agent.
	)
	if err != nil {
		// Return an error if the agent initialization fails.
		return fmt.Errorf("failed to initialize agent: %w", err)
	}

	// Start the agent, which begins collecting and producing metrics.
	if err = app.Start(); err != nil {
		// Return an error if the agent fails to start.
		return fmt.Errorf("failed to start agent: %w", err)
	}

	// Return nil to indicate successful execution.
	return nil
}
