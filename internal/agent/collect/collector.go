package collect

import (
	"NewNewMetricol/internal/agent/internal/entity"
	"runtime"
	"sync"

	"go.uber.org/zap"
)

// Collector is responsible for collecting runtime memory statistics
// and exporting them as metrics along with metadata metrics.
type Collector struct {
	metadata           *Metadata          // Metadata for metrics.
	ms                 *runtime.MemStats  // Memory statistics from runtime.
	mu                 *sync.RWMutex      // Mutex for concurrent access.
	logger             *zap.SugaredLogger // Logger for logging events.
	resetMetaConfirmCh chan bool          // Channel for metadata reset confirmation.
}

// NewCollector initializes and returns a new instance of Collector.
//
// Parameters:
//   - logger: Logger instance for recording events.
//
// Returns:
//   - *Collector: A pointer to the newly created Collector instance.
func NewCollector(logger *zap.SugaredLogger) *Collector {
	logger.Info("Initializing Collector")
	return &Collector{
		metadata:           NewMetadata(),
		ms:                 &runtime.MemStats{},
		mu:                 &sync.RWMutex{},
		logger:             logger,
		resetMetaConfirmCh: make(chan bool),
	}
}

// Collect gathers runtime memory statistics and updates metadata.
func (c *Collector) Collect() {
	c.mu.Lock()
	defer c.mu.Unlock()

	runtime.ReadMemStats(c.ms)
	c.metadata.Update()
	c.logger.Info("Metrics collected.")
}

// Export combines runtime memory metrics and metadata metrics into a single
// metrics entity and starts a goroutine to handle metadata reset confirmation.
//
// Returns:
//   - *entity.Metrics: A pointer to the exported metrics.
//   - chan bool: A channel for metadata reset confirmation.
func (c *Collector) Export() (*entity.Metrics, chan bool) {
	c.mu.RLock()
	metrics := append(c.exportMemoryMetrics(), c.exportMetadataMetrics()...)
	c.mu.RUnlock()

	go c.waitResetConfirmation()

	c.logger.Infof("Exporting metrics: Count=%d, Metrics=[%s]", metrics.Length(), metrics.ToString())
	return &metrics, c.resetMetaConfirmCh
}

// exportMemoryMetrics converts runtime memory statistics into metrics.
//
// Returns:
//   - entity.Metrics: A collection of memory-related metrics.
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

// exportMetadataMetrics converts metadata information into metrics.
//
// Returns:
//   - entity.Metrics: A collection of metadata-related metrics.
func (c *Collector) exportMetadataMetrics() entity.Metrics {
	return entity.Metrics{
		{Name: "RandomValue", Type: entity.MetricTypeGauge, Value: c.metadata.LastPollSeed(), IsMetadata: true},
		{Name: "PollCount", Type: entity.MetricTypeCounter, Value: c.metadata.PollsCount(), IsMetadata: true},
	}
}

// waitResetConfirmation waits for a confirmation signal to reset metadata.
// Resets the metadata if the signal is received, otherwise logs cancellation.
func (c *Collector) waitResetConfirmation() {
	c.mu.RLock()
	defer c.mu.RUnlock()

	c.logger.Info("Waiting for metadata reset confirmation...")
	select {
	case reset := <-c.resetMetaConfirmCh:
		if reset {
			c.metadata.Reset()
			c.logger.Info("Metadata reset successfully.")
		} else {
			c.logger.Warn("Metadata reset was canceled.")
		}
	}
}
