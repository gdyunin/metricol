package stategies

import (
	"runtime"
	"sync"

	"github.com/gdyunin/metricol.git/internal/agent/collect/metadata"
	"github.com/gdyunin/metricol.git/internal/agent/internal/entity"
	"go.uber.org/zap"
)

type MemStatsCollectStrategy struct {
	metadata *metadata.Metadata
	ms       *runtime.MemStats
	mu       *sync.RWMutex
	logger   *zap.SugaredLogger
}

// NewMemStatsCollectStrategy initializes and returns a new instance of MemStatsCollectStrategy.
//
// Parameters:
//   - logger: Logger instance for recording events.
//
// Returns:
//   - *MemStatsCollectStrategy: A pointer to the newly created Collector instance.
func NewMemStatsCollectStrategy(logger *zap.SugaredLogger) *MemStatsCollectStrategy {
	logger.Info("Initializing MemStatsCollectStrategy")
	return &MemStatsCollectStrategy{
		metadata: metadata.NewMetadata(),
		ms:       &runtime.MemStats{},
		mu:       &sync.RWMutex{},
		logger:   logger,
	}
}

func (m *MemStatsCollectStrategy) Collect() (*entity.Metrics, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	runtime.ReadMemStats(m.ms)
	m.metadata.Update()
	defer m.metadata.Reset()

	metrics := append(m.exportMemoryMetrics(), m.exportMetadataMetrics()...)
	return &metrics, nil
}

// exportMemoryMetrics converts runtime memory statistics into metrics.
//
// Returns:
//   - entity.Metrics: A collection of memory-related metrics.
func (m *MemStatsCollectStrategy) exportMemoryMetrics() entity.Metrics {
	return entity.Metrics{
		{Name: "Alloc", Type: entity.MetricTypeGauge, Value: float64(m.ms.Alloc)},
		{Name: "BuckHashSys", Type: entity.MetricTypeGauge, Value: float64(m.ms.BuckHashSys)},
		{Name: "Frees", Type: entity.MetricTypeGauge, Value: float64(m.ms.Frees)},
		{Name: "GCCPUFraction", Type: entity.MetricTypeGauge, Value: m.ms.GCCPUFraction},
		{Name: "GCSys", Type: entity.MetricTypeGauge, Value: float64(m.ms.GCSys)},
		{Name: "HeapAlloc", Type: entity.MetricTypeGauge, Value: float64(m.ms.HeapAlloc)},
		{Name: "HeapIdle", Type: entity.MetricTypeGauge, Value: float64(m.ms.HeapIdle)},
		{Name: "HeapInuse", Type: entity.MetricTypeGauge, Value: float64(m.ms.HeapInuse)},
		{Name: "HeapObjects", Type: entity.MetricTypeGauge, Value: float64(m.ms.HeapObjects)},
		{Name: "HeapReleased", Type: entity.MetricTypeGauge, Value: float64(m.ms.HeapReleased)},
		{Name: "HeapSys", Type: entity.MetricTypeGauge, Value: float64(m.ms.HeapSys)},
		{Name: "LastGC", Type: entity.MetricTypeGauge, Value: float64(m.ms.LastGC)},
		{Name: "Lookups", Type: entity.MetricTypeGauge, Value: float64(m.ms.Lookups)},
		{Name: "MCacheInuse", Type: entity.MetricTypeGauge, Value: float64(m.ms.MCacheInuse)},
		{Name: "MCacheSys", Type: entity.MetricTypeGauge, Value: float64(m.ms.MCacheSys)},
		{Name: "MSpanInuse", Type: entity.MetricTypeGauge, Value: float64(m.ms.MSpanInuse)},
		{Name: "MSpanSys", Type: entity.MetricTypeGauge, Value: float64(m.ms.MSpanSys)},
		{Name: "Mallocs", Type: entity.MetricTypeGauge, Value: float64(m.ms.Mallocs)},
		{Name: "NextGC", Type: entity.MetricTypeGauge, Value: float64(m.ms.NextGC)},
		{Name: "NumForcedGC", Type: entity.MetricTypeGauge, Value: float64(m.ms.NumForcedGC)},
		{Name: "NumGC", Type: entity.MetricTypeGauge, Value: float64(m.ms.NumGC)},
		{Name: "OtherSys", Type: entity.MetricTypeGauge, Value: float64(m.ms.OtherSys)},
		{Name: "PauseTotalNs", Type: entity.MetricTypeGauge, Value: float64(m.ms.PauseTotalNs)},
		{Name: "StackInuse", Type: entity.MetricTypeGauge, Value: float64(m.ms.StackInuse)},
		{Name: "StackSys", Type: entity.MetricTypeGauge, Value: float64(m.ms.StackSys)},
		{Name: "Sys", Type: entity.MetricTypeGauge, Value: float64(m.ms.Sys)},
		{Name: "TotalAlloc", Type: entity.MetricTypeGauge, Value: float64(m.ms.TotalAlloc)},
	}
}

// exportMetadataMetrics converts metadata information into metrics.
//
// Returns:
//   - entity.Metrics: A collection of metadata-related metrics.
func (m *MemStatsCollectStrategy) exportMetadataMetrics() entity.Metrics {
	return entity.Metrics{
		{
			Name:       "RandomValue",
			Type:       entity.MetricTypeGauge,
			Value:      m.metadata.LastPollSeed(),
			IsMetadata: true,
		},
		{Name: "PollCount", Type: entity.MetricTypeCounter, Value: m.metadata.PollsCount(), IsMetadata: true},
	}
}
