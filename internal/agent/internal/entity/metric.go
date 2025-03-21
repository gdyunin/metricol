// Package entity provides definitions for metrics and related utility functions.
// It defines types for representing individual metrics as well as collections of metrics,
// and includes helper methods for working with these collections.
package entity

import (
	"fmt"
	"strings"
)

const (
	// MetricTypeCounter represents a counter metric type.
	MetricTypeCounter = "counter"

	// MetricTypeGauge represents a gauge metric type.
	MetricTypeGauge = "gauge"
)

// Metric represents a single metric with its name, type, value, and a flag
// indicating whether it is metadata.
type Metric struct {
	Value      any    // Value holds the metric value.
	Name       string // Name is the identifier of the metric.
	Type       string // Type specifies the metric type, e.g., counter or gauge.
	IsMetadata bool   // IsMetadata indicates if the metric is metadata.
}

// Metrics is a collection of pointers to Metric.
type Metrics []*Metric

// Length returns the number of metrics in the collection.
//
// Returns:
//   - int: The count of metrics.
func (m *Metrics) Length() int {
	if m == nil {
		return 0
	}
	return len(*m)
}

// String converts the collection of metrics to a string representation.
// Each metric is formatted as <Name=..., Type=..., Value=...>.
//
// Returns:
//   - string: The string representation of the metrics collection.
func (m *Metrics) String() string {
	if m == nil {
		return ""
	}

	strData := make([]string, 0, m.Length())
	for _, metric := range *m {
		if metric != nil {
			strData = append(strData, fmt.Sprintf(
				"<Name=%s Type=%s Value=%v>",
				metric.Name,
				metric.Type,
				metric.Value,
			))
		} else {
			strData = append(strData, "<nil>")
		}
	}

	return strings.Join(strData, ", ")
}
