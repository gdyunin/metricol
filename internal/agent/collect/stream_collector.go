// Package collect provides interfaces and implementations for collecting metrics.
// It defines strategies to gather metrics from various sources and stream them through
// a channel for further processing. The package supports concurrent operations and is
// designed for efficient metrics collection.
package collect

import (
	"context"
	"sync"
	"time"

	"github.com/gdyunin/metricol.git/internal/agent/internal/entity"
	"go.uber.org/zap"
)

// Strategy defines an interface for a metric collection strategy.
// Implementations of Strategy should provide a Collect method that gathers metrics
// and returns them along with an error if one occurs.
type Strategy interface {
	// Collect gathers metrics based on the implemented strategy.
	//
	// Returns:
	//   - *entity.Metrics: Pointer to the collected metrics data.
	//   - error: An error if the collection fails; otherwise, nil.
	Collect() (*entity.Metrics, error)
}

// StreamCollector periodically collects metrics from multiple strategies and streams
// them to a specified channel. It is designed to work concurrently using a ticker.
type StreamCollector struct {
	streamTo          chan *entity.Metrics
	logger            *zap.SugaredLogger
	collectStrategies []Strategy
	interval          time.Duration
}

// NewStreamCollector creates and initializes a new StreamCollector instance.
//
// Parameters:
//   - streamTo: Channel to which collected metrics will be sent (chan *entity.Metrics).
//   - interval: Time duration between successive metric collections (time.Duration).
//   - collectStrategies: Slice of strategies to be used for collecting metrics ([]Strategy).
//   - logger: Logger instance for logging events (*zap.SugaredLogger).
//
// Returns:
//   - *StreamCollector: A pointer to the newly created StreamCollector.
func NewStreamCollector(
	streamTo chan *entity.Metrics,
	interval time.Duration,
	collectStrategies []Strategy,
	logger *zap.SugaredLogger,
) *StreamCollector {
	return &StreamCollector{
		streamTo:          streamTo,
		interval:          interval,
		collectStrategies: collectStrategies,
		logger:            logger,
	}
}

// StartStreaming begins the process of periodically collecting metrics using the defined strategies.
// The function runs indefinitely until the provided context is canceled. Metrics collection is performed
// concurrently and each successful collection is sent to the streamTo channel.
//
// Parameters:
//   - ctx: Context for managing the lifecycle of the metric streaming (context.Context).
//
// Returns:
//   - This function does not return any value; it exits when the context is canceled.
func (sc *StreamCollector) StartStreaming(ctx context.Context) {
	ticker := time.NewTicker(sc.interval)
	defer ticker.Stop()

	var wg sync.WaitGroup
	defer func() {
		wg.Wait()
		close(sc.streamTo)
	}()

	for {
		select {
		case <-ctx.Done():
			sc.logger.Info("Context canceled: stopping stream.")
			return
		case <-ticker.C:
			for _, strategy := range sc.collectStrategies {
				// For review: Ideally, this should be done via a worker pool or semaphore.
				// However, given the limited number of strategies, this limitation is acceptable
				// at the current stage of the project.
				wg.Add(1)
				go func(s Strategy) {
					defer wg.Done()

					collected, err := s.Collect()
					if err != nil {
						sc.logger.Errorf("Collect failed with %T and error: %v", sc.collectStrategies, err)
						return
					}

					if collected == nil || collected.Length() == 0 {
						sc.logger.Errorf("Received empty batch from %T and skipping.", sc.collectStrategies)
						return
					}

					sc.streamTo <- collected
				}(strategy)
			}
		}
	}
}
