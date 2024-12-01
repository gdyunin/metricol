package fetch

import (
	"log"

	"github.com/gdyunin/metricol.git/internal/metrics"
)

type Fetcher interface {
	Fetch()
	Metrics() []metrics.Metric
	AddMetrics(newMetrics ...metrics.Metric)
}

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
			log.Printf("metric %s update fail: %v", mm.Name, err)
		}
	}
}

func (m *MetricsFetcher) Metrics() []metrics.Metric {
	return m.metrics
}
