package stategies

import (
	"fmt"
	"time"

	"github.com/gdyunin/metricol.git/internal/agent/internal/entity"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"
	"go.uber.org/zap"
)

// GopsStatsCollectStrategy is a collection strategy that gathers system memory and CPU metrics
// using the gopsutil library. It logs its operations via the provided zap.SugaredLogger.
type GopsStatsCollectStrategy struct {
	logger *zap.SugaredLogger
}

// GopsMemStatsCollectStrategy initializes and returns a new instance of GopsStatsCollectStrategy.
//
// Parameters:
//   - logger: Logger instance for recording events.
//
// Returns:
//   - *GopsStatsCollectStrategy: A pointer to the newly created GopsStatsCollectStrategy instance.
func GopsMemStatsCollectStrategy(logger *zap.SugaredLogger) *GopsStatsCollectStrategy {
	logger.Info("Initializing GopsStatsCollectStrategy")
	return &GopsStatsCollectStrategy{
		logger: logger,
	}
}

// Collect gathers memory and CPU metrics and returns them as a pointer to entity.Metrics.
// If any error occurs during the collection of metrics, it returns an error.
//
// Returns:
//   - *entity.Metrics: A pointer to the collected metrics.
//   - error: An error if the collection process fails; otherwise, nil.
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

// collectMemMetrics collects memory metrics using gopsutil's mem.VirtualMemory function.
// It returns the memory metrics as an entity.Metrics slice or an error if the collection fails.
//
// Returns:
//   - entity.Metrics: A slice of memory metrics.
//   - error: An error if the collection process fails; otherwise, nil.
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

// collectCPUMetrics collects CPU utilization metrics using gopsutil's cpu.Percent function
// over a one second interval. It returns the CPU metrics as an entity.Metrics slice or an error
// if the collection fails.
//
// Returns:
//   - entity.Metrics: A slice of CPU utilization metrics.
//   - error: An error if the collection process fails; otherwise, nil.
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
