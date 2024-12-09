/*
Package metrics provides functionality for defining and managing various types of metrics,
including counters and gauges. It defines interfaces and structures to facilitate the
creation, updating, and string representation of these metrics.
*/
package metrics

const (
	// MetricTypeGauge represents a gauge metric type.
	MetricTypeGauge = "gauge"

	// MetricTypeCounter represents a counter metric type.
	MetricTypeCounter = "counter"
)

// Metric is an interface that defines the behavior of a metric.
// It requires methods for obtaining a string representation of the metric value
// and updating the metric`s value.
type Metric interface {
	// StringValue returns the current value of the metric as a string.
	StringValue() string

	// Update refreshes the metric`s value and returns an error if the update fails.
	Update() error
}
