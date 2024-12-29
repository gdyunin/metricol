package mscollector

import (
	"errors"
	"fmt"
	"math/rand"
	"runtime"
	"sync"
	"time"

	"github.com/gdyunin/metricol.git/internal/agent/adapters/collectors"
	"github.com/gdyunin/metricol.git/internal/agent/collect/collectors/mscollector/model"
	"github.com/gdyunin/metricol.git/internal/agent/entities"
	"github.com/gdyunin/metricol.git/internal/common"

	"go.uber.org/zap"
)

const (
	// ResetErrorCountersIntervals defines how often the error counters are reset.
	ResetErrorCountersIntervals = 4
	// MaxErrorsToInterrupt sets the maximum allowable errors before collection is interrupted.
	MaxErrorsToInterrupt = 3
)

// MemStatsCollector collects memory statistics periodically and stores them as metrics.
type MemStatsCollector struct {
	adp         *collectors.MemStatsCollectorAdapter // Adapter to interface with the metrics repository.
	ms          *runtime.MemStats                    // Stores memory statistics collected from the runtime.
	ticker      *time.Ticker                         // Ticker to trigger periodic data collection.
	interrupter *common.Interrupter                  // Handles error limits and interrupt signals.
	mu          *sync.RWMutex                        // Synchronizes access to shared resources.
	log         *zap.SugaredLogger                   // Logger for recording process information.
	polls       int                                  // Tracks the number of polls performed.
	interval    time.Duration                        // Interval between data collection cycles.
	seed        float64                              // Random value of current metric poll.
}

// NewMemStatsCollector creates a new instance of MemStatsCollector with a specified collection interval.
// Parameters:
//   - interval: The duration between data collection cycles.
//   - repo: Repository to store collected metrics.
//   - logger: Logger instance for logging activities.
//
// Returns:
//   - A pointer to a MemStatsCollector instance.
func NewMemStatsCollector(
	interval time.Duration,
	repo entities.MetricsRepository,
	logger *zap.SugaredLogger,
) *MemStatsCollector {
	collector := MemStatsCollector{
		adp:      collectors.NewMemStatsCollectorAdapter(repo),
		ms:       &runtime.MemStats{},
		interval: interval,
		mu:       &sync.RWMutex{},
		log:      logger,
	}

	logger.Infof("Collector inited: %+v", collector)

	return &collector
}

// StartCollect begins periodic memory statistics collection.
// Returns:
//   - An error if the collection process is interrupted or initialization fails.
func (c *MemStatsCollector) StartCollect() error {
	c.ticker = time.NewTicker(c.interval)
	defer c.ticker.Stop() // Ensure ticker is stopped when the method exits.

	// Initialize an interrupter to manage error handling and stopping criteria.
	interrupter, err := common.NewInterrupter(c.interval*ResetErrorCountersIntervals, MaxErrorsToInterrupt)
	if err != nil {
		return fmt.Errorf("failed to initialize the interrupter for error handling: %w", err)
	}
	c.interrupter = interrupter
	defer c.interrupter.Stop() // Ensure interrupter is stopped when the method exits.

	c.log.Info("Starting memory statistics collection.")
	for {
		select {
		case <-c.ticker.C:
			c.update() // Collect the latest memory statistics.
			c.store()  // Store the collected metrics.
		case <-c.interrupter.C:
			// Stop the collection process if the interrupter signals an error threshold breach.
			return errors.New("memory statistics collection interrupted: maximum error limit reached")
		}
	}
}

// OnNotify handles notifications from producers and resets the poll counter.
func (c *MemStatsCollector) OnNotify() {
	c.log.Info("Received notification from producer. Resetting poll counter.")
	defer c.log.Info("Poll counter reset successfully.")

	c.mu.Lock()
	defer c.mu.Unlock()

	c.polls = 0 // Reset the poll counter to zero.
}

// update collects the latest memory statistics and updates the internal state.
func (c *MemStatsCollector) update() {
	c.log.Info("Updating memory statistics.")
	defer c.log.Info("Memory statistics update completed.")

	c.mu.Lock()
	defer c.mu.Unlock()

	// Read the latest memory statistics from the runtime package.
	runtime.ReadMemStats(c.ms)
	c.polls++               // Increment the poll counter.
	c.seed = rand.Float64() // Generate a new random seed value for testing purposes.
}

// store saves the collected metrics to the repository.
// Logs errors for any failed metric storage attempts and tracks them via the interrupter.
func (c *MemStatsCollector) store() {
	c.log.Info("Storing collected metrics.")

	metrics := c.metrics() // Generate the list of metrics to store.
	for _, m := range metrics {
		if !c.interrupter.InLimit() {
			// Abort storage if error limits are exceeded.
			c.log.Error("Aborting metrics storage due to exceeding interrupter error limits.")
			return
		}

		// Attempt to store the metric and handle errors.
		if err := c.adp.Store(m); err != nil {
			c.log.Errorf("Failed to store metric '%s': %v", m.Name, err)
			c.interrupter.AddError() // Increment the interrupter's error counter.
			continue
		}
	}
	c.log.Info("Metrics stored.")
}

// metrics generates a list of memory-related metrics to be stored in the repository.
func (c *MemStatsCollector) metrics() []*model.Metric {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Construct and return a slice of Metric objects based on memory statistics.
	return []*model.Metric{
		{Name: "Alloc", Type: entities.MetricTypeGauge, Value: float64(c.ms.Alloc)},
		{Name: "BuckHashSys", Type: entities.MetricTypeGauge, Value: float64(c.ms.BuckHashSys)},
		{Name: "Frees", Type: entities.MetricTypeGauge, Value: float64(c.ms.Frees)},
		{Name: "GCCPUFraction", Type: entities.MetricTypeGauge, Value: c.ms.GCCPUFraction},
		{Name: "GCSys", Type: entities.MetricTypeGauge, Value: float64(c.ms.GCSys)},
		{Name: "HeapAlloc", Type: entities.MetricTypeGauge, Value: float64(c.ms.HeapAlloc)},
		{Name: "HeapIdle", Type: entities.MetricTypeGauge, Value: float64(c.ms.HeapIdle)},
		{Name: "HeapInuse", Type: entities.MetricTypeGauge, Value: float64(c.ms.HeapInuse)},
		{Name: "HeapObjects", Type: entities.MetricTypeGauge, Value: float64(c.ms.HeapObjects)},
		{Name: "HeapReleased", Type: entities.MetricTypeGauge, Value: float64(c.ms.HeapReleased)},
		{Name: "HeapSys", Type: entities.MetricTypeGauge, Value: float64(c.ms.HeapSys)},
		{Name: "LastGC", Type: entities.MetricTypeGauge, Value: float64(c.ms.LastGC)},
		{Name: "Lookups", Type: entities.MetricTypeGauge, Value: float64(c.ms.Lookups)},
		{Name: "MCacheInuse", Type: entities.MetricTypeGauge, Value: float64(c.ms.MCacheInuse)},
		{Name: "MCacheSys", Type: entities.MetricTypeGauge, Value: float64(c.ms.MCacheSys)},
		{Name: "MSpanInuse", Type: entities.MetricTypeGauge, Value: float64(c.ms.MSpanInuse)},
		{Name: "MSpanSys", Type: entities.MetricTypeGauge, Value: float64(c.ms.MSpanSys)},
		{Name: "Mallocs", Type: entities.MetricTypeGauge, Value: float64(c.ms.Mallocs)},
		{Name: "NextGC", Type: entities.MetricTypeGauge, Value: float64(c.ms.NextGC)},
		{Name: "NumForcedGC", Type: entities.MetricTypeGauge, Value: float64(c.ms.NumForcedGC)},
		{Name: "NumGC", Type: entities.MetricTypeGauge, Value: float64(c.ms.NumGC)},
		{Name: "OtherSys", Type: entities.MetricTypeGauge, Value: float64(c.ms.OtherSys)},
		{Name: "PauseTotalNs", Type: entities.MetricTypeGauge, Value: float64(c.ms.PauseTotalNs)},
		{Name: "StackInuse", Type: entities.MetricTypeGauge, Value: float64(c.ms.StackInuse)},
		{Name: "StackSys", Type: entities.MetricTypeGauge, Value: float64(c.ms.StackSys)},
		{Name: "Sys", Type: entities.MetricTypeGauge, Value: float64(c.ms.Sys)},
		{Name: "TotalAlloc", Type: entities.MetricTypeGauge, Value: float64(c.ms.TotalAlloc)},
		{Name: "RandomValue", Type: entities.MetricTypeGauge, Value: c.seed},
		{Name: "PollCount", Type: entities.MetricTypeCounter, Value: int64(c.polls)},
	}
}
