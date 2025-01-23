package stategies

import (
	"fmt"
	"time"

	"github.com/gdyunin/metricol.git/internal/agent/internal/entity"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"
	"go.uber.org/zap"
)

type GopsStatsCollectStrategy struct {
	logger *zap.SugaredLogger
}

// GopsMemStatsCollectStrategy initializes and returns a new instance of GopsStatsCollectStrategy.
//
// Parameters:
//   - logger: Logger instance for recording events.
//
// Returns:
//   - *GopsStatsCollectStrategy: A pointer to the newly created Collector instance.
func GopsMemStatsCollectStrategy(logger *zap.SugaredLogger) *GopsStatsCollectStrategy {
	logger.Info("Initializing GopsStatsCollectStrategy")
	return &GopsStatsCollectStrategy{
		logger: logger,
	}
}

func (m *GopsStatsCollectStrategy) Collect() (*entity.Metrics, error) {
	var metrics entity.Metrics

	memory, err := m.collectMemMetrics()
	if err != nil {
		return nil, fmt.Errorf("failed collect mem: %w", err)
	}

	cpuUtilization, err := m.collectCPUMetrics()
	if err != nil {
		return nil, fmt.Errorf("failed collect cpu: %w", err)
	}

	metrics = append(metrics, memory...)
	metrics = append(metrics, cpuUtilization...)
	return &metrics, nil
}

func (m *GopsStatsCollectStrategy) collectMemMetrics() (entity.Metrics, error) {
	v, err := mem.VirtualMemory()
	if err != nil {
		return nil, fmt.Errorf("failed collect memory metrics: %w", err)
	}

	return entity.Metrics{
		&entity.Metric{
			Value:      float64(v.Total),
			Name:       "TotalMemory",
			Type:       entity.MetricTypeGauge,
			IsMetadata: false,
		},
		&entity.Metric{
			Value:      float64(v.Free),
			Name:       "FreeMemory",
			Type:       entity.MetricTypeGauge,
			IsMetadata: false,
		},
	}, nil
}

func (m *GopsStatsCollectStrategy) collectCPUMetrics() (entity.Metrics, error) {
	cpuPercentages, err := cpu.Percent(time.Second, true)
	if err != nil {
		return nil, fmt.Errorf("failed collect cpu metrics: %w", err)
	}

	var metrics entity.Metrics
	for i := range cpuPercentages {
		metrics = append(metrics, &entity.Metric{
			Value:      cpuPercentages[i],
			Name:       fmt.Sprint("CPUutilization", i+1),
			Type:       entity.MetricTypeGauge,
			IsMetadata: false,
		})
	}

	return metrics, nil
}
