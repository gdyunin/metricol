/*
Package metrics provides functionality for defining and managing various types of metrics,
including counters and gauges. It defines interfaces and structures to facilitate the
creation, updating, and string representation of these metrics.
*/
package metrics

import (
	"fmt"
)

// NewCounter creates a new Counter metric with the specified name and initial value.
// It returns a pointer to the newly created Counter.
func NewCounter(name string, value int64) *Counter {
	return &Counter{
		Name:  name,
		Value: value,
	}
}

// NewGauge creates a new Gauge metric with the specified name and initial value.
// It returns a pointer to the newly created Gauge.
func NewGauge(name string, value float64) *Gauge {
	return &Gauge{
		Name:  name,
		Value: value,
	}
}

// NewFromStrings creates a Metric based on the provided name, value, and metric type.
// It returns an error if the metric type is unknown or if there is an error creating the metric.
func NewFromStrings(name, value, metricType string) (Metric, error) {
	var createMetric func(string, string) (Metric, error)

	// Determine the appropriate function to create a metric based on its type.
	switch metricType {
	case MetricTypeGauge:
		createMetric = newGaugeFromStrings
	case MetricTypeCounter:
		createMetric = newCounterFromStrings
	default:
		return nil, fmt.Errorf("unknown metric type: %s", metricType)
	}

	m, err := createMetric(name, value)
	if err != nil {
		return nil, fmt.Errorf("error creating metric from strings: %w", err)
	}
	return m, nil
}
