// Package agent provides functionality for sending metrics to a server.
package agent

import (
	"bytes"
	"fmt"

	"github.com/gdyunin/metricol.git/internal/metrics"
)

// MetricsFetcher implements the Fetcher interface, managing a collection of metrics.
type MetricsFetcher struct {
	metrics []metrics.Metric // Collection of metrics to be fetched.
}

// NewMetricsFetcher creates a new instance of MetricsFetcher with an empty metric collection.
func NewMetricsFetcher() *MetricsFetcher {
	return &MetricsFetcher{
		metrics: []metrics.Metric{},
	}
}

// AddMetrics appends new metrics to the existing collection of metrics.
func (m *MetricsFetcher) AddMetrics(newMetrics ...metrics.Metric) {
	m.metrics = append(m.metrics, newMetrics...)
}

// Fetch updates all metrics in the collection.
// It returns an error if any metric fails to update.
func (m *MetricsFetcher) Fetch() error {
	var buf bytes.Buffer // Buffer to collect error messages
	for _, mm := range m.metrics {
		if err := mm.Update(); err != nil {
			buf.WriteString(fmt.Sprintf("metric %v update fail: %v\n", mm, err))
			continue // Skip to the next metric
		}
	}

	if buf.Len() != 0 {
		return fmt.Errorf("one or more metrics were not fetched: %s", buf.String())
	}
	return nil
}

// Metrics returns the list of metrics managed by the fetcher.
func (m *MetricsFetcher) Metrics() []metrics.Metric {
	return m.metrics
}
