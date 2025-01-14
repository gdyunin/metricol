package agent

import (
	"context"
	"time"

	"github.com/gdyunin/metricol.git/internal/agent/internal/entity"

	"go.uber.org/zap"
)

// Collector defines an interface for collecting and exporting metrics.
type Collector interface {
	// Collect gathers metrics from the system or application.
	Collect()
	// Export returns the collected metrics along with a reset channel.
	Export() (*entity.Metrics, chan bool)
}

// Sender defines an interface for sending metrics to a remote server.
type Sender interface {
	// SendSingle sends a single metric to the server.
	SendSingle(context.Context, *entity.Metric) error
	// SendBatch sends a batch of metrics to the server.
	SendBatch(context.Context, *entity.Metrics) error
}

// Agent manages the collection and sending of metrics at specified intervals.
type Agent struct {
	collector      Collector
	sender         Sender
	logger         *zap.SugaredLogger
	pollInterval   time.Duration
	reportInterval time.Duration
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
	collector Collector,
	sender Sender,
	pollInterval time.Duration,
	reportInterval time.Duration,
	logger *zap.SugaredLogger,
) *Agent {
	logger.Infof(
		"Initializing Agent: pollInterval=%ds, reportInterval=%ds",
		pollInterval/time.Second,
		reportInterval/time.Second,
	)
	return &Agent{
		collector:      collector,
		sender:         sender,
		pollInterval:   pollInterval,
		reportInterval: reportInterval,
		logger:         logger,
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

	collectTicker := time.NewTicker(a.pollInterval)
	defer collectTicker.Stop()

	sendTicker := time.NewTicker(a.reportInterval)
	defer sendTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			a.logger.Info("Context canceled: stopping agent")
			return
		case <-collectTicker.C:
			a.collect()
		case t := <-sendTicker.C:
			// [ДЛЯ РЕВЬЮ]: Дедлайн контекста произойдет примерно в то же время, когда снова тикнет sendTicker.
			senderCtx, cancel := context.WithDeadline(ctx, t.Add(a.reportInterval))
			a.sendByBatch(senderCtx)
			cancel()
		}
	}
}

// collect triggers the collection of metrics using the Collector interface.
// TODO: Подумать, нужен ли вообще. Используется только в 1 месте и просто проксирует другой метод без доп логики.
func (a *Agent) collect() {
	a.collector.Collect()
}

// sendBySingle sends collected metrics one by one to the server.
//
// Parameters:
//   - ctx: Context for managing the lifecycle of the operation.
func (a *Agent) sendBySingle(ctx context.Context) {
	metrics, resetCh := a.collector.Export()
	reset := true
	defer func() { resetCh <- reset }()

	if metrics.Length() == 0 {
		a.logger.Info("No metrics to send")
		return
	}

	a.logger.Infof("Preparing to send %d metrics individually", metrics.Length())

	for _, m := range *metrics {
		if ctx.Err() != nil {
			reset = false
			a.logger.Warn("Context canceled or deadline exceeded during sendBySingle, stopping")
			return
		}

		if err := a.sender.SendSingle(ctx, m); err != nil {
			reset = !m.IsMetadata
			a.logger.Warnf("Failed to send metric: name=%s, error=%v, metadataReset=%t", m.Name, err, reset)
		} else {
			a.logger.Infof("Successfully sent metric: name=%s", m.Name)
		}
	}
}

// sendByBatch sends collected metrics in a batch to the server.
//
// Parameters:
//   - ctx: Context for managing the lifecycle of the operation.
func (a *Agent) sendByBatch(ctx context.Context) {
	metrics, resetCh := a.collector.Export()
	reset := true
	defer func() { resetCh <- reset }()

	if metrics.Length() == 0 {
		a.logger.Info("No metrics to send in batch")
		return
	}

	a.logger.Infof("Preparing to send %d metrics in batch", metrics.Length())

	if ctx.Err() != nil {
		reset = false
		a.logger.Warn("Context canceled or deadline exceeded during sendByBatch, stopping")
		return
	}

	if err := a.sender.SendBatch(ctx, metrics); err != nil {
		reset = false
		a.logger.Warnf(
			"Failed to send metrics batch: count=%d, error=%v, metadataReset=%t",
			metrics.Length(),
			err,
			reset,
		)
	} else {
		a.logger.Infof("Successfully sent batch of metrics: count=%d", metrics.Length())
	}
}
