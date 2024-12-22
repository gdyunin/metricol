package agent

import (
	"fmt"
	"sync"

	"github.com/gdyunin/metricol.git/internal/agent/collect"
	"github.com/gdyunin/metricol.git/internal/agent/common"
	"github.com/gdyunin/metricol.git/internal/agent/produce"

	"go.uber.org/zap"
)

// Agent is the main struct responsible for managing data collection and production.
// It coordinates the Collector and Producer and ensures proper execution flow.
type Agent struct {
	collector collect.Collector
	producer  produce.Producer
	log       *zap.SugaredLogger
}

// NewAgent initializes a new Agent instance with the provided Collector, Producer, and logger.
// Optionally, it applies additional configurations through functional options.
//
// Parameters:
//   - collector: An implementation of the Collector interface.
//   - producer: An implementation of the Producer interface.
//   - logger: An instance of zap.SugaredLogger for logging purposes.
//   - options: Optional functional configurations for the Agent.
//
// Returns:
//   - A pointer to the initialized Agent instance.
//   - An error if applying any of the options fails.
func NewAgent(collector collect.Collector, producer produce.Producer, logger *zap.SugaredLogger, options ...func(*Agent) error) (a *Agent, err error) {
	a = &Agent{
		collector: collector,
		producer:  producer,
		log:       logger,
	}

	for _, o := range options {
		if err = o(a); err != nil {
			return nil, fmt.Errorf("failed to apply option function of type %T: %w", o, err)
		}
	}
	return
}

// Start begins the data collection and production processes in parallel.
// If either process encounters an error, it stops the agent and returns the error.
//
// Returns:
//   - An error indicating which executor (collector or producer) stopped the agent.
func (a *Agent) Start() error {
	var workGroup sync.WaitGroup
	workGroup.Add(1)

	var interruptedExecutor string
	var err error

	// Start the data collection process.
	go func() {
		defer func() {
			interruptedExecutor = "collector"
			workGroup.Done()
		}()

		if collectorErr := a.collector.StartCollect(); collectorErr != nil {
			a.log.Errorf("Collector encountered an error: %v", collectorErr)
			err = collectorErr
		}
	}()

	// Start the data production process.
	go func() {
		defer func() {
			interruptedExecutor = "producer"
			workGroup.Done()
		}()

		if producerErr := a.producer.StartProduce(); producerErr != nil {
			a.log.Errorf("Producer encountered an error: %v", producerErr)
			err = producerErr
		}
	}()

	// Wait for both processes to complete.
	workGroup.Wait()

	// Return an error if either executor failed.
	if err != nil {
		return fmt.Errorf("agent was stopped: %s process encountered an error: %w", interruptedExecutor, err)
	}
	return nil
}

// WithSubscribeConsumer2Producer subscribes the Collector to the Producer's events.
// It ensures the Collector implements the Observer interface and the Producer implements the ObserveSubject interface.
//
// Parameters:
//   - agent: A pointer to the Agent instance.
//
// Returns:
//   - An error if the subscription process fails.
func WithSubscribeConsumer2Producer(agent *Agent) error {
	observer, ok := agent.collector.(common.Observer)
	if !ok {
		return fmt.Errorf("collector does not implement the Observer interface")
	}

	subject, ok := agent.producer.(common.ObserveSubject)
	if !ok {
		return fmt.Errorf("producer does not implement the ObserveSubject interface")
	}

	if err := common.Subscribe(observer, subject); err != nil {
		return fmt.Errorf("failed to subscribe collector to producer: %w", err)
	}

	return nil
}
