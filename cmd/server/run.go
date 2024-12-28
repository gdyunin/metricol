package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gdyunin/metricol.git/internal/server/backup/managers/basic"
	"github.com/gdyunin/metricol.git/internal/server/config"
	"github.com/gdyunin/metricol.git/internal/server/consume"
	"github.com/gdyunin/metricol.git/internal/server/consume/consumers/echoserver"
	"github.com/gdyunin/metricol.git/internal/server/repositories"
	"github.com/gdyunin/metricol.git/pkg/logger"
)

// stopper is a function type used for managing cleanup tasks during shutdown.
type stopper func()

// shutdown holds a collection of cleanup functions to execute during a graceful shutdown.
type shutdown struct {
	fn []stopper // Slice of cleanup functions.
}

// executeAll runs all registered cleanup functions in the order they were added.
func (s *shutdown) executeAll() {
	for _, f := range s.fn {
		f() // Execute each cleanup function.
	}
}

// run initializes and starts the application, handling all core components and logic.
func run() error {
	// Initialize the logger with INFO level logging.
	baseInfoLogger, err := logger.Logger(logger.LevelINFO)
	if err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}

	// Parse the application configuration from environment variables or defaults.
	appCfg, err := config.ParseConfig()
	if err != nil {
		return fmt.Errorf("failed to parse config: %w", err)
	}

	// Create an in-memory repository for storing server data.
	repo := repositories.NewInMemoryRepository()

	// Initialize the EchoServer consumer with the server address and repository.
	consumer := echoserver.NewEchoServer(appCfg.ServerAddress, repo, baseInfoLogger)

	// Set up the backup system to save server data to a file.
	backupper := basic.NewBackupManager(
		appCfg.FileStoragePath, // Path to the backup file.
		"backup.txt",           // Backup file name.
		time.Duration(appCfg.StoreInterval)*time.Second, // Interval for saving backups.
		appCfg.Restore, // Whether to restore data on startup.
		repo,           // Repository to backup and restore.
	)

	// Restore data from the backup file, if enabled in the configuration.
	backupper.Restore()

	// StartAll the backup process in a separate goroutine.
	go backupper.StartBackup()

	// Set up graceful shutdown to ensure cleanup on termination.
	setupGracefulShutdown(&shutdown{fn: []stopper{
		backupper.StopBackup, // Add the StopBackup function to the shutdown tasks.
	}})

	// StartAll the consumer to process incoming data.
	if err = consume.Consumer(consumer).StartConsume(); err != nil {
		return fmt.Errorf("failed to start the consumption process: %w", err)
	}

	return nil
}

// setupGracefulShutdown configures signal handling for graceful application termination.
func setupGracefulShutdown(s *shutdown) {
	stopChan := make(chan os.Signal, 1)                    // Create a channel to receive OS signals.
	signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM) // Subscribe to interrupt and terminate signals.

	go func() {
		<-stopChan     // Wait for a termination signal.
		s.executeAll() // Execute all registered cleanup tasks.
		os.Exit(0)     // Exit the application with a success code.
	}()
}
