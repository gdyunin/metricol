package fetch

import (
	"fmt"
	"github.com/gdyunin/metricol.git/internal/metrics"
)

type MetricsFetcher struct {
	metrics []metrics.Metric
}

func NewMetricsFetcher() *MetricsFetcher {
	return &MetricsFetcher{
		metrics: []metrics.Metric{},
	}
}

func (m *MetricsFetcher) AddMetrics(newMetrics ...metrics.Metric) {
	m.metrics = append(m.metrics, newMetrics...)
}

func (m *MetricsFetcher) Fetch() {
	for _, mm := range m.metrics {
		if err := mm.Update(); err != nil {
			// A logger could be added in the future
			fmt.Printf("metric %s update fail: %s", mm.Name(), err.Error())
		}
	}
}

func (m *MetricsFetcher) Metrics() []metrics.Metric {
	return m.metrics
}
