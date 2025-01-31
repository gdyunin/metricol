package agent

import (
	"context"
	"sync"
	"time"

	"github.com/gdyunin/metricol.git/internal/agent/collect"
	"github.com/gdyunin/metricol.git/internal/agent/collect/stategies"
	"github.com/gdyunin/metricol.git/internal/agent/internal/entity"
	"github.com/gdyunin/metricol.git/internal/agent/send"

	"go.uber.org/zap"
)

const sendQueueSizeCoefficient = 5

// Collector defines an interface for collecting and exporting metrics.
type Collector interface {
	// Collect gathers metrics from the system or application.
	Collect()
	// Export returns the collected metrics along with a reset channel.
	Export() (*entity.Metrics, chan bool)
}

// Sender defines an interface for sending metrics to a remote server.
type Sender interface {
	// SendBatch sends a batch of metrics to the server.
	SendBatch(context.Context, *entity.Metrics) error
}

// Agent manages the collection and sending of metrics at specified intervals.
type Agent struct {
	logger         *zap.SugaredLogger
	sendQueue      chan *entity.Metrics
	serverAddress  string
	signKey        string
	pollInterval   time.Duration
	reportInterval time.Duration
	maxSendRate    int
}

// NewAgent creates and initializes a new Agent.
//
// Parameters:
//   - collector: Implementation of the Collector interface.
//   - sender: Implementation of the Sender interface.
//   - pollInterval: Interval for collecting metrics.
//   - reportInterval: Interval for sending metrics to the server.
//   - logger: Logger instance for logging events.
//
// Returns:
//   - *Agent: A pointer to the initialized Agent.
func NewAgent(
	pollInterval time.Duration,
	reportInterval time.Duration,
	logger *zap.SugaredLogger,
	maxSendRate int,
	serverAddress string,
	signKey string,
) *Agent {
	logger.Infof(
		"Initializing Agent: pollInterval=%ds, reportInterval=%ds",
		pollInterval/time.Second,
		reportInterval/time.Second,
	)
	return &Agent{
		pollInterval:   pollInterval,
		reportInterval: reportInterval,
		logger:         logger,
		sendQueue:      make(chan *entity.Metrics, maxSendRate*sendQueueSizeCoefficient),
		maxSendRate:    maxSendRate,
		serverAddress:  serverAddress,
		signKey:        signKey,
	}
}

// Start begins the operation of the Agent.
//
// This method runs indefinitely, managing the collection and sending
// of metrics based on the configured intervals. It can be stopped by
// canceling the provided context.
//
// Parameters:
//   - ctx: Context for managing the lifecycle of the Agent.
func (a *Agent) Start(ctx context.Context) {
	a.logger.Infof(
		"Agent started: pollInterval=%ds, reportInterval=%ds",
		a.pollInterval/time.Second,
		a.reportInterval/time.Second,
	)

	collectStrategies := []collect.Strategy{
		stategies.NewMemStatsCollectStrategy(a.logger.Named("mem_strategy")),
		stategies.GopsMemStatsCollectStrategy(a.logger.Named("gops_strategy")),
	}
	streamCollector := collect.NewStreamCollector(
		a.sendQueue,
		a.pollInterval,
		collectStrategies,
		a.logger.Named("collector"),
	)

	streamSenderLogger := a.logger.Named("stream_sender")
	streamSender := send.NewStreamSender(
		a.sendQueue,
		a.reportInterval,
		a.maxSendRate,
		a.serverAddress,
		a.signKey,
		streamSenderLogger,
	)

	workers := []func(context.Context){
		streamCollector.StartStreaming,
		streamSender.StartStreaming,
	}

	var wg sync.WaitGroup
	for _, worker := range workers {
		wg.Add(1)
		go func(w func(context.Context)) {
			defer wg.Done()
			w(ctx)
		}(worker)
	}

	wg.Wait()
}
