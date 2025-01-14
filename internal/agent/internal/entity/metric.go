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
	Value      any
	Name       string
	Type       string
	IsMetadata bool
}

// Metrics is a collection of Metric pointers.
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
