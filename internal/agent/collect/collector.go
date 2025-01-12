package collect

import (
	"NewNewMetricol/internal/agent/internal/entity"
	"runtime"
	"sync"

	"go.uber.org/zap"
)

// Collector is responsible for collecting runtime memory statistics
// and exporting them as a set of metrics along with metadata metrics.
type Collector struct {
	metadata *Metadata
	ms       *runtime.MemStats
	mu       *sync.RWMutex
	logger   *zap.SugaredLogger
	resetCh  chan bool
}

// NewCollector initializes a new Collector instance.
func NewCollector(logger *zap.SugaredLogger) *Collector {
	return &Collector{
		metadata: NewMetadata(),
		ms:       &runtime.MemStats{},
		mu:       &sync.RWMutex{},
		logger:   logger,
		resetCh:  make(chan bool),
	}
}

// Collect gathers the latest runtime memory statistics and updates the metadata.
func (c *Collector) Collect() {
	c.mu.Lock()
	defer c.mu.Unlock()

	runtime.ReadMemStats(c.ms)
	c.metadata.Update()
	c.logger.Info("Metrics collected.")
}

// Export gathers all collected metrics, including memory and metadata metrics,
// and resets the metadata state.
func (c *Collector) Export() (*entity.Metrics, chan bool) {
	c.mu.RLock()
	metrics := append(c.exportMemoryMetrics(), c.exportMetadataMetrics()...)
	c.mu.RUnlock()

	go c.awaitResetConfirmation()

	c.logger.Infof("Metrics exported. Count: %d | Metrics: [%s].", metrics.Length(), metrics.ToString())
	return &metrics, c.resetCh
}

// exportMemoryMetrics converts memory statistics into a slice of Metric entity.
func (c *Collector) exportMemoryMetrics() entity.Metrics {
	return entity.Metrics{
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
	}
}

// exportMetadataMetrics converts metadata into a slice of Metric entity.
func (c *Collector) exportMetadataMetrics() entity.Metrics {
	return entity.Metrics{
		{Name: "RandomValue", Type: entity.MetricTypeGauge, Value: c.metadata.LastPollSeed(), IsMetadata: true},
		{Name: "PollCount", Type: entity.MetricTypeCounter, Value: c.metadata.PollsCount(), IsMetadata: true},
	}
}

// awaitResetConfirmation waits for confirmation to reset metadata.
func (c *Collector) awaitResetConfirmation() {
	c.mu.RLock()
	defer c.mu.RUnlock()

	c.logger.Info("Wait confirmation for reset metadata...")
	select {
	case reset := <-c.resetCh:
		if reset {
			c.metadata.Reset()
			c.logger.Info("Metadata reset after confirmation.")
		} else {
			c.logger.Warn("Metadata reset canceled.")
		}
	}
}
