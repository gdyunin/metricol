package basic

import (
	"errors"
	"fmt"
	"sync"

	"github.com/gdyunin/metricol.git/internal/agent/collect"
	"github.com/gdyunin/metricol.git/internal/agent/produce"
	"github.com/gdyunin/metricol.git/internal/common"

	"go.uber.org/zap"
)

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
	options ...func(*Orchestrator) error,
) (a *Orchestrator, err error) {
	// Initialize the Orchestrator with provided components.
	a = &Orchestrator{
		collector: collector,
		producer:  producer,
		log:       logger,
	}

	// Apply each provided functional option to configure the Orchestrator.
	for _, o := range options {
		if err = o(a); err != nil {
			return nil, fmt.Errorf("failed to apply option function of type %T: %w", o, err)
		}
	}
	return
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

// WithSubscribeConsumer2Producer subscribes the Collector to the Producer's events.
// It ensures the Collector implements the Observer interface and the Producer implements the ObserveSubject interface.
//
// Parameters:
//   - orchestrate: A pointer to the Orchestrator instance.
//
// Returns:
//   - An error if the subscription process fails.
func WithSubscribeConsumer2Producer(agent *Orchestrator) error {
	// Verify the collectors implements the Observer interface.
	observer, ok := agent.collector.(common.Observer)
	if !ok {
		return errors.New("collectors does not implement the Observer interface")
	}

	// Verify the producer implements the ObserveSubject interface.
	subject, ok := agent.producer.(common.ObserveSubject)
	if !ok {
		return errors.New("producer does not implement the ObserveSubject interface")
	}

	// Subscribe the collectors to the producer's events.
	if err := common.Subscribe(observer, subject); err != nil {
		return fmt.Errorf("failed to subscribe collectors to producer: %w", err)
	}

	return nil
}
