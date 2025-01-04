package main

import (
	"fmt"

	"github.com/gdyunin/metricol.git/internal/common/helpers"
	"github.com/gdyunin/metricol.git/internal/common/utils"
	"github.com/gdyunin/metricol.git/internal/server/backup"
	backupManagerFact "github.com/gdyunin/metricol.git/internal/server/backup/factory"
	"github.com/gdyunin/metricol.git/internal/server/config"
	"github.com/gdyunin/metricol.git/internal/server/consume"
	consumerFact "github.com/gdyunin/metricol.git/internal/server/consume/factory"
	"github.com/gdyunin/metricol.git/internal/server/entities"
	repositoryFact "github.com/gdyunin/metricol.git/internal/server/repositories/factory"
	"github.com/gdyunin/metricol.git/pkg/logger"
	"go.uber.org/zap"
)

const (
	// LoggerNameConfigParser is the logger name used for configuration parsing.
	LoggerNameConfigParser = "config_parser"
	// LoggerNameRepository is the logger name used for repository operations.
	LoggerNameRepository = "repository"
	// LoggerNameConsumer is the logger name used for consumer operations.
	LoggerNameConsumer = "consumer"
)

// Components aggregates all application components for streamlined initialization and management.
type Components struct {
	baseLogger      *zap.SugaredLogger
	repository      entities.MetricsRepository
	consumer        consume.Consumer
	backupManager   backup.Manager
	shutdownManager *helpers.ShutdownManager
}

// run initializes the application components and starts the main application loop.
//
// Parameters:
//   - logLevel: Logging level for the application.
//
// Returns:
//   - An error if the application fails to initialize or run.
func run(logLevel string) (err error) {
	appComponents, err := prepareComponents(logLevel)
	if err != nil {
		return fmt.Errorf("error occurred while initializing application components: %w", err)
	}

	// Restore metrics from backup if configured.
	appComponents.backupManager.Restore()

	// Start periodic backups in a separate goroutine.
	go appComponents.backupManager.Start()

	// Ensure the backup manager stops gracefully on application shutdown.
	appComponents.shutdownManager.Add(appComponents.backupManager.Stop)
	helpers.SetupGracefulShutdown(appComponents.shutdownManager)

	// Start the consumer to handle incoming metrics.
	err = appComponents.consumer.StartConsume()
	if err != nil {
		return fmt.Errorf("failed to start or run the consumer: %w", err)
	}

	return nil
}

// prepareComponents initializes all required application components.
//
// Parameters:
//   - logLevel: Logging level for the application.
//
// Returns:
//   - A pointer to a Components struct containing all initialized components.
//   - An error if any component fails to initialize.
func prepareComponents(logLevel string) (*Components, error) {
	baseLogger, err := logger.Logger(logLevel)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize the logger: %w", err)
	}

	appCfg, err := config.ParseConfig(baseLogger.Named(LoggerNameConfigParser))
	if err != nil {
		return nil, fmt.Errorf("failed to parse application configuration: %w", err)
	}

	repositoryFactory, err := repositoryFact.AbstractRepositoriesFactory(repositoryFact.RepoTypeInMemory, baseLogger.Named(LoggerNameRepository))
	if err != nil {
		return nil, fmt.Errorf("failed to initialize the repository: %w", err)
	}
	repository := repositoryFactory.CreateMetricsRepository()

	consumerFactory, err := consumerFact.AbstractConsumerFactory(consumerFact.ConsumerTypeEchoServer, appCfg.ServerAddress, repository, baseLogger.Named(LoggerNameConsumer))
	if err != nil {
		return nil, fmt.Errorf("failed to initialize the consumer: %w", err)
	}
	consumer := consumerFactory.CreateConsumer()

	backupManagerFactory, err := backupManagerFact.AbstractManagerFactory(
		backupManagerFact.ManagerTypeBasic,
		appCfg.FileStoragePath,
		"backup.txt",
		utils.IntegerToSeconds(appCfg.StoreInterval),
		appCfg.Restore,
		repository,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize the backup manager: %w", err)
	}
	backupManager := backupManagerFactory.CreateManager()

	shutdownManager := helpers.NewShutdownManager()

	return &Components{
		baseLogger:      baseLogger,
		repository:      repository,
		consumer:        consumer,
		backupManager:   backupManager,
		shutdownManager: shutdownManager,
	}, nil
}
