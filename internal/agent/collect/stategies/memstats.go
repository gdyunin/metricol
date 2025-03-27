package stategies

import (
	"runtime"
	"sync"

	"github.com/gdyunin/metricol.git/internal/agent/collect/metadata"
	"github.com/gdyunin/metricol.git/internal/agent/internal/entity"
	"go.uber.org/zap"
)

// MemStatsCollectStrategy is a collection strategy that gathers memory statistics from the Go runtime
// using runtime.ReadMemStats. It also collects metadata information to supplement the metrics.
// The strategy is safe for concurrent use.
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
//   - *MemStatsCollectStrategy: A pointer to the newly created MemStatsCollectStrategy instance.
func NewMemStatsCollectStrategy(logger *zap.SugaredLogger) *MemStatsCollectStrategy {
	logger.Info("Initializing MemStatsCollectStrategy")
	return &MemStatsCollectStrategy{
		metadata: metadata.NewMetadata(),
		ms:       &runtime.MemStats{},
		mu:       &sync.RWMutex{},
		logger:   logger,
	}
}

// Collect gathers memory statistics from the runtime and metadata information,
// converts them into metrics, and returns the metrics as a pointer to entity.Metrics.
// It returns an error if the collection process fails.
//
// Returns:
//   - *entity.Metrics: A pointer to the collected metrics.
//   - error: An error if the collection process fails; otherwise, nil.
func (m *MemStatsCollectStrategy) Collect() (*entity.Metrics, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	runtime.ReadMemStats(m.ms)
	m.metadata.Update()
	defer m.metadata.Reset()

	metrics := append(m.exportMemoryMetrics(), m.exportMetadataMetrics()...)
	return &metrics, nil
}

// exportMemoryMetrics converts runtime memory statistics into a slice of metrics.
// This function is used internally to generate memory-related metrics.
//
// Returns:
//   - entity.Metrics: A collection of memory-related metrics.
func (m *MemStatsCollectStrategy) exportMemoryMetrics() entity.Metrics {
	const metricsCount = 27
	metrics := make(entity.Metrics, 0, metricsCount)
	metrics = append(metrics,
		&entity.Metric{Name: "Alloc", Type: entity.MetricTypeGauge, Value: float64(m.ms.Alloc)},
		&entity.Metric{Name: "BuckHashSys", Type: entity.MetricTypeGauge, Value: float64(m.ms.BuckHashSys)},
		&entity.Metric{Name: "Frees", Type: entity.MetricTypeGauge, Value: float64(m.ms.Frees)},
		&entity.Metric{Name: "GCCPUFraction", Type: entity.MetricTypeGauge, Value: m.ms.GCCPUFraction},
		&entity.Metric{Name: "GCSys", Type: entity.MetricTypeGauge, Value: float64(m.ms.GCSys)},
		&entity.Metric{Name: "HeapAlloc", Type: entity.MetricTypeGauge, Value: float64(m.ms.HeapAlloc)},
		&entity.Metric{Name: "HeapIdle", Type: entity.MetricTypeGauge, Value: float64(m.ms.HeapIdle)},
		&entity.Metric{Name: "HeapInuse", Type: entity.MetricTypeGauge, Value: float64(m.ms.HeapInuse)},
		&entity.Metric{Name: "HeapObjects", Type: entity.MetricTypeGauge, Value: float64(m.ms.HeapObjects)},
		&entity.Metric{Name: "HeapReleased", Type: entity.MetricTypeGauge, Value: float64(m.ms.HeapReleased)},
		&entity.Metric{Name: "HeapSys", Type: entity.MetricTypeGauge, Value: float64(m.ms.HeapSys)},
		&entity.Metric{Name: "LastGC", Type: entity.MetricTypeGauge, Value: float64(m.ms.LastGC)},
		&entity.Metric{Name: "Lookups", Type: entity.MetricTypeGauge, Value: float64(m.ms.Lookups)},
		&entity.Metric{Name: "MCacheInuse", Type: entity.MetricTypeGauge, Value: float64(m.ms.MCacheInuse)},
		&entity.Metric{Name: "MCacheSys", Type: entity.MetricTypeGauge, Value: float64(m.ms.MCacheSys)},
		&entity.Metric{Name: "MSpanInuse", Type: entity.MetricTypeGauge, Value: float64(m.ms.MSpanInuse)},
		&entity.Metric{Name: "MSpanSys", Type: entity.MetricTypeGauge, Value: float64(m.ms.MSpanSys)},
		&entity.Metric{Name: "Mallocs", Type: entity.MetricTypeGauge, Value: float64(m.ms.Mallocs)},
		&entity.Metric{Name: "NextGC", Type: entity.MetricTypeGauge, Value: float64(m.ms.NextGC)},
		&entity.Metric{Name: "NumForcedGC", Type: entity.MetricTypeGauge, Value: float64(m.ms.NumForcedGC)},
		&entity.Metric{Name: "NumGC", Type: entity.MetricTypeGauge, Value: float64(m.ms.NumGC)},
		&entity.Metric{Name: "OtherSys", Type: entity.MetricTypeGauge, Value: float64(m.ms.OtherSys)},
		&entity.Metric{Name: "PauseTotalNs", Type: entity.MetricTypeGauge, Value: float64(m.ms.PauseTotalNs)},
		&entity.Metric{Name: "StackInuse", Type: entity.MetricTypeGauge, Value: float64(m.ms.StackInuse)},
		&entity.Metric{Name: "StackSys", Type: entity.MetricTypeGauge, Value: float64(m.ms.StackSys)},
		&entity.Metric{Name: "Sys", Type: entity.MetricTypeGauge, Value: float64(m.ms.Sys)},
		&entity.Metric{Name: "TotalAlloc", Type: entity.MetricTypeGauge, Value: float64(m.ms.TotalAlloc)},
	)
	return metrics
}

// exportMetadataMetrics converts metadata information into a slice of metrics.
// This function is used internally to generate metadata-related metrics.
//
// Returns:
//   - entity.Metrics: A collection of metadata-related metrics.
func (m *MemStatsCollectStrategy) exportMetadataMetrics() entity.Metrics {
	const metricsCount = 2
	metrics := make(entity.Metrics, 0, metricsCount)
	metrics = append(metrics,
		&entity.Metric{
			Name:       "RandomValue",
			Type:       entity.MetricTypeGauge,
			Value:      m.metadata.LastPollSeed(),
			IsMetadata: true,
		},
		&entity.Metric{
			Name:       "PollCount",
			Type:       entity.MetricTypeCounter,
			Value:      m.metadata.PollsCount(),
			IsMetadata: true,
		},
	)
	return metrics
}
