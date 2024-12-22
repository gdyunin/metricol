package mscollector

import (
	"errors"
	"fmt"
	"math/rand"
	"runtime"
	"sync"
	"time"

	"github.com/gdyunin/metricol.git/internal/agent/adapter/collect"
	"github.com/gdyunin/metricol.git/internal/agent/collect/collectors/mscollector/model"
	"github.com/gdyunin/metricol.git/internal/agent/common"
	"github.com/gdyunin/metricol.git/internal/agent/entity"

	"go.uber.org/zap"
)

const (
	resetErrorCountersIntervals = 4 // Interval multiplier for resetting error counters.
	maxErrorsToInterrupt        = 3 // Maximum errors allowed before interrupting collection.
)

// MemStatsCollector collects memory statistics periodically and stores them as metrics.
type MemStatsCollector struct {
	adp         *collect.MemStatsCollectorAdapter
	ms          *runtime.MemStats
	ticker      *time.Ticker
	interrupter *common.Interrupter
	mu          *sync.RWMutex
	log         *zap.SugaredLogger
	polls       int
	interval    time.Duration
	seed        float64
}

// NewMemStatsCollector creates a new instance of MemStatsCollector with a specified collection interval.
func NewMemStatsCollector(
	interval time.Duration,
	repo entity.MetricsRepository,
	logger *zap.SugaredLogger,
) *MemStatsCollector {
	return &MemStatsCollector{
		adp:      collect.NewMemStatsCollectorAdapter(repo),
		ms:       &runtime.MemStats{},
		interval: interval,
		mu:       &sync.RWMutex{},
		log:      logger,
	}
}

// StartCollect begins periodic memory statistics collection.
// Returns an error if the collection process is interrupted or initialization fails.
func (c *MemStatsCollector) StartCollect() error {
	c.ticker = time.NewTicker(c.interval)
	defer c.ticker.Stop()

	interrupter, err := common.NewInterrupter(c.interval*resetErrorCountersIntervals, maxErrorsToInterrupt)
	if err != nil {
		return fmt.Errorf("failed to initialize the interrupter for error handling: %w", err)
	}
	c.interrupter = interrupter
	defer c.interrupter.Stop()

	c.log.Info("Starting memory statistics collection.")
	for {
		select {
		case <-c.ticker.C:
			c.update()
			c.store()
		case <-c.interrupter.C:
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

	c.polls = 0
}

// update collects the latest memory statistics and updates the internal state.
func (c *MemStatsCollector) update() {
	c.log.Info("Updating memory statistics.")
	defer c.log.Info("Memory statistics update completed.")

	c.mu.Lock()
	defer c.mu.Unlock()

	runtime.ReadMemStats(c.ms)
	c.polls++
	c.seed = rand.Float64()
}

// store saves the collected metrics to the repository.
// Logs errors for any failed metric storage attempts and tracks them via the interrupter.
func (c *MemStatsCollector) store() {
	c.log.Info("Storing collected metrics.")

	metrics := c.metrics()
	for _, m := range metrics {
		if !c.interrupter.InLimit() {
			c.log.Error("Aborting metrics storage due to exceeding interrupter error limits.")
			return
		}

		if err := c.adp.Store(m); err != nil {
			c.log.Errorf("Failed to store metric '%s': %v", m.Name, err)
			c.interrupter.AddError()
			continue
		}
	}
	c.log.Info("Metrics stored.")
}

// metrics generates a list of memory-related metrics to be stored in the repository.
func (c *MemStatsCollector) metrics() []*model.Metric {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return []*model.Metric{
		{Name: "Alloc", Type: entity.MetricTypeGauge, Value: float64(c.ms.Alloc)},
		{Name: "BuckHashSys", Type: entity.MetricTypeGauge, Value: float64(c.ms.BuckHashSys)},
		{Name: "Frees", Type: entity.MetricTypeGauge, Value: float64(c.ms.Frees)},
		{Name: "GCCPUFraction", Type: entity.MetricTypeGauge, Value: c.ms.GCCPUFraction},
		{Name: "GCSys", Type: entity.MetricTypeGauge, Value: float64(c.ms.GCSys)},
		{Name: "HeapAlloc", Type: entity.MetricTypeGauge, Value: float64(c.ms.HeapAlloc)},
		{Name: "HeapIdle", Type: entity.MetricTypeGauge, Value: float64(c.ms.HeapIdle)},
		{Name: "HeapInuse", Type: entity.MetricTypeGauge, Value: float64(c.ms.HeapInuse)},
		{Name: "HeapObjects", Type: entity.MetricTypeGauge, Value: float64(c.ms.HeapObjects)},
		{Name: "HeapReleased", Type: entity.MetricTypeGauge, Value: float64(c.ms.HeapReleased)},
		{Name: "HeapSys", Type: entity.MetricTypeGauge, Value: float64(c.ms.HeapSys)},
		{Name: "LastGC", Type: entity.MetricTypeGauge, Value: float64(c.ms.LastGC)},
		{Name: "Lookups", Type: entity.MetricTypeGauge, Value: float64(c.ms.Lookups)},
		{Name: "MCacheInuse", Type: entity.MetricTypeGauge, Value: float64(c.ms.MCacheInuse)},
		{Name: "MCacheSys", Type: entity.MetricTypeGauge, Value: float64(c.ms.MCacheSys)},
		{Name: "MSpanInuse", Type: entity.MetricTypeGauge, Value: float64(c.ms.MSpanInuse)},
		{Name: "MSpanSys", Type: entity.MetricTypeGauge, Value: float64(c.ms.MSpanSys)},
		{Name: "Mallocs", Type: entity.MetricTypeGauge, Value: float64(c.ms.Mallocs)},
		{Name: "NextGC", Type: entity.MetricTypeGauge, Value: float64(c.ms.NextGC)},
		{Name: "NumForcedGC", Type: entity.MetricTypeGauge, Value: float64(c.ms.NumForcedGC)},
		{Name: "NumGC", Type: entity.MetricTypeGauge, Value: float64(c.ms.NumGC)},
		{Name: "OtherSys", Type: entity.MetricTypeGauge, Value: float64(c.ms.OtherSys)},
		{Name: "PauseTotalNs", Type: entity.MetricTypeGauge, Value: float64(c.ms.PauseTotalNs)},
		{Name: "StackInuse", Type: entity.MetricTypeGauge, Value: float64(c.ms.StackInuse)},
		{Name: "StackSys", Type: entity.MetricTypeGauge, Value: float64(c.ms.StackSys)},
		{Name: "Sys", Type: entity.MetricTypeGauge, Value: float64(c.ms.Sys)},
		{Name: "TotalAlloc", Type: entity.MetricTypeGauge, Value: float64(c.ms.TotalAlloc)},
		{Name: "RandomValue", Type: entity.MetricTypeGauge, Value: c.seed},
		{Name: "PollCount", Type: entity.MetricTypeCounter, Value: int64(c.polls)},
	}
}
