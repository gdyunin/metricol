package agent

import (
	"NewNewMetricol/internal/agent/internal/entity"
	"context"
	"time"

	"go.uber.org/zap"
)

// Collector defines the interface for collecting and exporting metrics.
type Collector interface {
	// Collect gathers metrics from the system or application.
	Collect()
	// Export returns a slice of collected metrics.
	Export() (*entity.Metrics, chan bool)
}

// Sender defines the interface for sending metrics to a remote server.
type Sender interface {
	// SendSingle sends a single metric to a remote server.
	// Parameters:
	//   - ctx: The context for managing timeout and cancellation.
	//   - *entity.Metric: The metric to be sent.
	// Returns:
	//   - error: An error if the operation fails, or nil if successful.
	SendSingle(context.Context, *entity.Metric) error
	SendBatch(context.Context, *entity.Metrics) error
}

// Agent is responsible for collecting and sending metrics at specified intervals.
type Agent struct {
	collector      Collector     // Component for collecting metrics.
	sender         Sender        // Component for sending metrics.
	pollInterval   time.Duration // Interval between metric collection.
	reportInterval time.Duration // Interval between sending metrics to the server.
	logger         *zap.SugaredLogger
}

// NewAgent creates and initializes a new Agent.
//
// Parameters:
//   - collector: An implementation of the Collector interface.
//   - sender: An implementation of the Sender interface.
//   - pollInterval: Duration between metric collection cycles.
//   - reportInterval: Duration between metric reporting cycles.
//
// Returns:
//   - *Agent: A pointer to the initialized Agent.
func NewAgent(collector Collector, sender Sender, pollInterval time.Duration, reportInterval time.Duration, logger *zap.SugaredLogger) *Agent {
	logger.Infof("Init agent with poll_interval=%d and report_interval=%d", pollInterval, reportInterval)
	return &Agent{
		collector:      collector,
		sender:         sender,
		pollInterval:   pollInterval,
		reportInterval: reportInterval,
		logger:         logger,
	}
}

// Start runs the Agent, collecting and sending metrics at configured intervals.
//
// This method blocks indefinitely, so if non-blocking behavior is required,
// it should be executed in a separate goroutine.
//
// Metrics collection is performed every `pollInterval`, and metrics are sent
// to the server every `reportInterval`. The operation can be stopped by canceling
// the provided context.
//
// Parameters:
//
//	ctx: The context used to manage the lifecycle of the agent. The context
//	     allows graceful termination when canceled or timed out.
func (a *Agent) Start(ctx context.Context) {
	a.logger.Info("agent started")
	// Set up a ticker for metric collection.
	collectTicker := time.NewTicker(a.pollInterval)
	defer collectTicker.Stop() // Ensure the ticker is stopped to release resources.

	// Set up a ticker for sending metrics.
	sendTicker := time.NewTicker(a.reportInterval)
	defer sendTicker.Stop() // Ensure the ticker is stopped to release resources.

	for {
		select {
		case <-ctx.Done():
			// Stop execution when the context is canceled or its deadline is exceeded.
			return
		case <-collectTicker.C:
			// Collect metrics at regular intervals defined by pollInterval.
			a.collect()
		case t := <-sendTicker.C:
			// Send metrics to the server at regular intervals defined by reportInterval.
			senderCtx, cancel := context.WithDeadline(ctx, t.Add(a.reportInterval))
			a.sendBySingle(senderCtx)
			cancel()
		}
	}
}

// collect triggers the collection of metrics using the Collector interface.
func (a *Agent) collect() {
	a.collector.Collect()
}

// sendBySingle sends all collected metrics using the Sender interface.
func (a *Agent) sendBySingle(ctx context.Context) {
	metrics, resetCh := a.collector.Export()

	reset := true
	defer func() { resetCh <- reset }()

	for _, m := range *metrics {
		if ctx.Err() != nil {
			// Context was canceled or deadline exceeded; stop sending.
			reset = false
			a.logger.Warn("Context canceled or deadline exceeded during sendBySingle")
			return
		}

		// Attempt to sendBySingle the metric.
		if err := a.sender.SendSingle(ctx, m); err != nil {
			reset = !m.IsMetadata
			a.logger.Warnf("Error sending metric %s: %v | metadata will be reset: %t", m.Name, err, reset)
		}
	}
}

func (a *Agent) sendByBatch(ctx context.Context) {
	metrics, resetCh := a.collector.Export()

	reset := true
	defer func() { resetCh <- reset }()

	if ctx.Err() != nil {
		// Context was canceled or deadline exceeded; stop sending.
		reset = false
		a.logger.Warn("Context canceled or deadline exceeded during sendBySingle")
		return
	}

	if err := a.sender.SendBatch(ctx, metrics); err != nil {
		reset = false
		a.logger.Warnf("Error sending batch of metrics. Count: %d | Metrics: [%s] | Error: %v | metadata will be reset: %t", metrics.Length(), metrics.ToString(), err, reset)
	}
}
