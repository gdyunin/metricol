// Package agent provides functionalities for collecting and sending metrics.
// It defines interfaces and implementations to collect system or application metrics
// and send them to a remote server at specified intervals. The package supports concurrent
// operations and structured logging. All time intervals are specified as time.Duration.
// This package is intended for use in monitoring and metric collection applications.
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

// sendQueueSizeCoefficient is used to calculate the size of the send queue based on the maximum send rate.
const sendQueueSizeCoefficient = 5

// Collector defines an interface for collecting and exporting metrics.
// Implementations of Collector should gather metrics from the system or application and provide
// a mechanism to export the collected metrics along with a reset channel.
type Collector interface {
	// Collect gathers metrics from the system or application.
	// This function does not take any input parameters and does not return any value.
	Collect()
	// Export returns the collected metrics along with a reset channel.
	//
	// Returns:
	//   - *entity.Metrics: Pointer to the collected metrics data.
	//   - chan bool: A channel used to signal a reset of the metrics data.
	Export() (*entity.Metrics, chan bool)
}

// Sender defines an interface for sending metrics to a remote server.
// Implementations of Sender should send a batch of metrics and handle any errors that occur.
type Sender interface {
	// SendBatch sends a batch of metrics to the server.
	//
	// Parameters:
	//   - ctx: Context for managing cancellation and deadlines (context.Context).
	//   - metrics: Pointer to the metrics data to be sent (*entity.Metrics).
	//
	// Returns:
	//   - error: An error value if sending fails; otherwise, nil.
	SendBatch(context.Context, *entity.Metrics) error
}

// Agent manages the collection and sending of metrics at specified intervals.
// It utilizes a Collector to gather metrics and a Sender to transmit the collected metrics
// to a remote server.
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
//   - pollInterval: Interval for collecting metrics (time.Duration).
//   - reportInterval: Interval for sending metrics to the server (time.Duration).
//   - logger: Logger instance for logging events (*zap.SugaredLogger).
//   - maxSendRate: Maximum number of metric batches that can be sent per report interval (int).
//   - serverAddress: Address of the remote server to which metrics are sent (string).
//   - signKey: Signing key used for authentication when sending metrics (string).
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
// This method runs indefinitely, managing the collection and sending of metrics based on the configured intervals.
// It launches separate goroutines for collecting and sending metrics,
// and can be stopped by canceling the provided context.
//
// Parameters:
//   - ctx: Context for managing the lifecycle of the Agent (context.Context).
//
// Returns:
//   - This function does not return any value.
func (a *Agent) Start(ctx context.Context) {
	a.logger.Infof(
		"Agent started: pollInterval=%ds, reportInterval=%ds",
		a.pollInterval/time.Second,
		a.reportInterval/time.Second,
	)

	// Initialize collection strategies for gathering metrics.
	collectStrategies := []collect.Strategy{
		stategies.NewMemStatsCollectStrategy(a.logger.Named("mem_strategy")),
		stategies.GopsMemStatsCollectStrategy(a.logger.Named("gops_strategy")),
	}

	// Create a new stream collector that gathers metrics and sends them to the sendQueue.
	streamCollector := collect.NewStreamCollector(
		a.sendQueue,
		a.pollInterval,
		collectStrategies,
		a.logger.Named("collector"),
	)

	// Create a new stream sender that sends metrics from the sendQueue to the remote server.
	streamSenderLogger := a.logger.Named("stream_sender")
	streamSender := send.NewStreamSender(
		a.sendQueue,
		a.reportInterval,
		a.maxSendRate,
		a.serverAddress,
		a.signKey,
		streamSenderLogger,
	)

	// Define workers for collection and sending.
	workers := []func(context.Context){
		streamCollector.StartStreaming,
		streamSender.StartStreaming,
	}

	var wg sync.WaitGroup
	// Start each worker in its own goroutine.
	for _, worker := range workers {
		wg.Add(1)
		go func(w func(context.Context)) {
			defer wg.Done()
			w(ctx)
		}(worker)
	}

	// Wait for all workers to finish.
	wg.Wait()
}
