package main

import (
	"fmt"
	"log"

	"github.com/gdyunin/metricol.git/internal/server/config"
	"github.com/gdyunin/metricol.git/internal/server/consume/consumers/ginserver"
	"github.com/gdyunin/metricol.git/pkg/logger"
	"github.com/gdyunin/metricol.git/pkg/server/repositories"
	"github.com/gin-gonic/gin"
)

// run initializes the in-memory repository and the Gin server consumer, then starts the consumption process.
// It returns detailed errors if any part of the initialization or execution fails.
func run() error {
	appCfg, err := config.ParseConfig()
	if err != nil {
		log.Fatalf("Get configuration fail: %v", err)
	}

	// Initialize the base logger with the INFO log level for application-wide logging.
	baseLogger, err := logger.Logger(logger.LevelINFO)
	if err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}

	// Create an in-memory repository to manage server-side data.
	repo := repositories.NewInMemoryRepository()

	// Initialize the Gin server consumer with a configuration parser and the repository.
	gin.SetMode(gin.ReleaseMode)
	consumer := ginserver.NewServer(appCfg.ServerAddress, repo, baseLogger)

	// Start the consumption process, handling any errors that occur during runtime.
	if err := consumer.StartConsume(); err != nil {
		return fmt.Errorf("failed to start the consumption process: %w", err)
	}

	return nil
}

// main is the application's entry point.
// It calls the run function to initialize and execute the application.
// Any critical errors are logged and terminate the application.
func main() {
	if err := run(); err != nil {
		log.Fatalf("Application encountered a critical error: %v", err)
	}
}
