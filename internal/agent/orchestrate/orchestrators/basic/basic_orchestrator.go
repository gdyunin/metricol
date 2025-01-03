package basic

import (
	"fmt"
	"sync"

	"github.com/gdyunin/metricol.git/internal/agent/collect"
	"github.com/gdyunin/metricol.git/internal/agent/orchestrate"
	"github.com/gdyunin/metricol.git/internal/agent/produce"
	"go.uber.org/zap"
)

type OrchestratorFactory struct {
	collector collect.Collector
	producer  produce.Producer
	logger    *zap.SugaredLogger
}

func NewOrchestratorFactory(collector collect.Collector, producer produce.Producer, logger *zap.SugaredLogger) *OrchestratorFactory {
	return &OrchestratorFactory{
		collector: collector,
		producer:  producer,
		logger:    logger,
	}
}

func (f *OrchestratorFactory) CreateOrchestrator() orchestrate.Orchestrator {
	return NewOrchestrator(f.collector, f.producer, f.logger)
}

// Orchestrator is the main struct responsible for managing data collection and production.
// It coordinates the Collector and Producer and ensures proper execution flow.
type Orchestrator struct {
	collector collect.Collector  // Component responsible for collecting data.
	producer  produce.Producer   // Component responsible for producing data.
	log       *zap.SugaredLogger // Logger for capturing runtime information and errors.
}

// NewOrchestrator initializes a new Orchestrator instance with the provided Collector, Producer, and logger.
// Optionally, it applies additional configurations through functional options.
//
// Parameters:
//   - collectors: An implementation of the Collector interface.
//   - producer: An implementation of the Producer interface.
//   - logger: An instance of zap.SugaredLogger for logging purposes.
//   - options: Optional functional configurations for the Orchestrator.
//
// Returns:
//   - A pointer to the initialized Orchestrator instance.
//   - An error if applying any of the options fails.
func NewOrchestrator(
	collector collect.Collector,
	producer produce.Producer,
	logger *zap.SugaredLogger,
) *Orchestrator {
	// Initialize the Orchestrator with provided components.
	return &Orchestrator{
		collector: collector,
		producer:  producer,
		log:       logger,
	}
}

// StartAll begins the data collection and production processes in parallel.
// If either process encounters an error, it stops the orchestrate and returns the error.
//
// Returns:
//   - An error indicating which executor (collectors or producer) stopped the orchestrate.
func (a *Orchestrator) StartAll() error {
	var workGroup sync.WaitGroup
	workGroup.Add(1) // If one of the task (producer or consumer) was done, the apllication will be stopped.

	var interruptedExecutor string // Tracks which executor caused an interruption.
	var err error                  // Holds the error encountered by an executor.

	// StartAll the data collection process.
	go func() {
		defer func() {
			interruptedExecutor = "collectors" // Mark the collectors as the interrupted executor if it fails.
			workGroup.Done()                   // Signal the WaitGroup that this task is complete.
		}()

		if collectorErr := a.collector.StartCollect(); collectorErr != nil {
			err = collectorErr // Capture the error encountered by the collectors.
		}
	}()

	// StartAll the data production process.
	go func() {
		defer func() {
			interruptedExecutor = "producer" // Mark the producer as the interrupted executor if it fails.
			workGroup.Done()                 // Signal the WaitGroup that this task is complete.
		}()

		if producerErr := a.producer.StartProduce(); producerErr != nil {
			err = producerErr // Capture the error encountered by the producer.
		}
	}()

	workGroup.Wait()

	// Return an error if either executor failed.
	if err != nil {
		return fmt.Errorf("orchestrate was stopped: %s process encountered an error: %w", interruptedExecutor, err)
	}
	return nil
}
