package main

import (
	"fmt"

	collectorFact "github.com/gdyunin/metricol.git/internal/agent/collect/factory"
	"github.com/gdyunin/metricol.git/internal/agent/config"
	"github.com/gdyunin/metricol.git/internal/agent/orchestrate"
	orchestratorFact "github.com/gdyunin/metricol.git/internal/agent/orchestrate/factory"
	producerFact "github.com/gdyunin/metricol.git/internal/agent/produce/factory"
	repositoryFact "github.com/gdyunin/metricol.git/internal/agent/repositories/factory"
	"github.com/gdyunin/metricol.git/internal/common/patterns"
	"github.com/gdyunin/metricol.git/internal/common/utils"
	"github.com/gdyunin/metricol.git/pkg/logger"
)

const (
	// LoggerNameConfigParser is the logger name used for configuration parsing.
	LoggerNameConfigParser = "config_parser"
	// LoggerNameRepository is the logger name used for repository operations.
	LoggerNameRepository = "repository"
	// LoggerNameCollector is the logger name used for metric collection.
	LoggerNameCollector = "collector"
	// LoggerNameProducer is the logger name used for producer operations.
	LoggerNameProducer = "producer"
	// LoggerNameOrchestrator is the logger name used for the orchestrator.
	LoggerNameOrchestrator = "orchestrator"
)

// run initializes the orchestrator and starts all its components.
func run(logLevel string) (err error) {
	orchestrator, err := makeOrchestrator(logLevel)
	if err != nil {
		err = fmt.Errorf("error occurred during orchestrator initialization: %w", err)
		return
	}

	err = orchestrator.StartAll()
	if err != nil {
		err = fmt.Errorf("error occurred while starting or running the orchestrator: %w", err)
		return
	}

	return
}

// makeOrchestrator initializes and returns an orchestrator configured with the required components.
func makeOrchestrator(logLevel string) (orchestrate.Orchestrator, error) {
	// Initialize the base logger.
	baseLogger, err := logger.Logger(logLevel)
	if err != nil {
		return nil, fmt.Errorf("error occurred while initializing the logger: %w", err)
	}

	// Parse the application configuration.
	appCfg, err := config.ParseConfig(baseLogger.Named(LoggerNameConfigParser))
	if err != nil {
		return nil, fmt.Errorf("error occurred while parsing the application configuration: %w", err)
	}

	// Initialize the repository factory and create the metrics repository.
	repositoryFactory, err := repositoryFact.AbstractRepositoriesFactory(
		repositoryFact.RepoTypeInMemory,
		baseLogger.Named(LoggerNameRepository),
	)
	if err != nil {
		return nil, fmt.Errorf("error occurred while initializing the repository: %w", err)
	}
	repository := repositoryFactory.CreateMetricsRepository()

	// Initialize the collector factory and create the collector.
	collectorFactory, err := collectorFact.AbstractCollectorFactory(
		collectorFact.CollectorTypeMemStats,
		utils.IntegerToSeconds(appCfg.PollInterval),
		repository,
		baseLogger.Named(LoggerNameCollector),
	)
	if err != nil {
		return nil, fmt.Errorf("error occurred while initializing the collector: %w", err)
	}
	collector := collectorFactory.CreateCollector()

	// Initialize the producer factory and create the producer.
	producerFactory, err := producerFact.AbstractProducerFactory(
		producerFact.ProducerTypeRestyClient,
		utils.IntegerToSeconds(appCfg.ReportInterval),
		appCfg.ServerAddress,
		repository,
		baseLogger.Named(LoggerNameProducer),
	)
	if err != nil {
		return nil, fmt.Errorf("error occurred while initializing the producer: %w", err)
	}
	producer := producerFactory.CreateProducer()

	// Link the collector to the producer using the observer pattern.
	err = patterns.Subscribe(collector, producer)
	if err != nil {
		return nil, fmt.Errorf("error occurred while linking the collector to the producer: %w", err)
	}

	// Initialize the orchestrator factory and create the orchestrator.
	orchestratorFactory, err := orchestratorFact.AbstractOrchestratorsFactory(
		orchestratorFact.OrchestratorTypeBasic,
		collector,
		producer,
		baseLogger.Named(LoggerNameOrchestrator),
	)
	if err != nil {
		return nil, fmt.Errorf("error occurred while initializing the orchestrator: %w", err)
	}

	return orchestratorFactory.CreateOrchestrator(), nil
}
