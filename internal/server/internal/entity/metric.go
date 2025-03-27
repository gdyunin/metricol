// Package entity defines the data structures and helper functions for representing and
// manipulating metrics. Metrics are identified by a name, type, and value. This package
// provides custom JSON unmarshalling for metrics, along with utility methods to work with
// collections of metrics.
package entity

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gdyunin/metricol.git/pkg/convert"
)

const (
	// MetricTypeCounter defines the metric type for counters.
	MetricTypeCounter = "counter"
	// MetricTypeGauge defines the metric type for gauges.
	MetricTypeGauge = "gauge"
)

// Metric represents a single metric with a name, type, and value.
// It is used to encapsulate the measurement data.
type Metric struct {
	Value any    `json:"value"` // Value holds the metric's value.
	Name  string `json:"name"`  // Name is the identifier of the metric.
	Type  string `json:"type"`  // Type specifies the metric's category, e.g., "counter" or "gauge".
}

// UnmarshalJSON implements custom JSON unmarshalling for the Metric type.
// It parses the JSON data into a Metric and performs type conversion for counter metrics.
// If the metric is of type "counter", it converts the value to int64.
//
// Parameters:
//   - data: A byte slice containing the JSON representation of a Metric.
//
// Returns:
//   - error: An error if the JSON data cannot be parsed or if the type conversion fails.
func (m *Metric) UnmarshalJSON(data []byte) error {
	// Define an alias to avoid recursive call to UnmarshalJSON.
	type MetricAlias Metric
	aux := &struct {
		*MetricAlias
	}{
		MetricAlias: (*MetricAlias)(m),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return fmt.Errorf("unable to parse metric JSON: %w", err)
	}

	// For counter metrics, ensure the value is converted to int64.
	if m.Type == MetricTypeCounter {
		v, err := convert.AnyToInt64(m.Value)
		if err != nil {
			return fmt.Errorf(
				"invalid value for counter metric \"%s\": expected integer, got %T",
				m.Name,
				m.Value,
			)
		}
		m.Value = v
	}

	return nil
}

// Metrics represents a collection of Metric pointers.
type Metrics []*Metric

// Length returns the number of metrics in the collection.
//
// Returns:
//   - int: The count of metrics in the collection.
func (m *Metrics) Length() int {
	if m == nil {
		return 0
	}
	return len(*m)
}

// First returns the first metric in the collection.
// If the collection is empty or nil, it returns nil.
//
// Returns:
//   - *Metric: A pointer to the first Metric in the collection or nil if empty.
func (m *Metrics) First() *Metric {
	if m == nil || len(*m) == 0 {
		return nil
	}
	return (*m)[0]
}

// String returns a string representation of the metrics collection.
// Each metric is formatted as "<Name=... Type=... Value=...>" and concatenated with commas.
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

// MergeDuplicates merges duplicate metrics in the collection.
// Two metrics are considered duplicates if they share the same name and type.
// For counter metrics, their values are summed; for gauge metrics, the latest value replaces the previous one.
// The merged collection replaces the original one.
//
// Returns:
//   - This function does not return a value; it modifies the receiver in place.
func (m *Metrics) MergeDuplicates() {
	if m == nil || len(*m) == 0 {
		return
	}

	merged := make(map[string]*Metric)

	for _, metric := range *m {
		if metric == nil {
			continue
		}

		key := metric.Name + "|" + metric.Type
		if existing, found := merged[key]; found {
			if metric.Type == MetricTypeCounter {
				existingVal, _ := convert.AnyToInt64(existing.Value)
				repeatVal, _ := convert.AnyToInt64(metric.Value)
				existing.Value = existingVal + repeatVal
			} else {
				// For gauge metrics, replace with the latest value.
				existing.Value = metric.Value
			}
		} else {
			merged[key] = metric
		}
	}

	result := make(Metrics, 0, len(merged))
	for _, metric := range merged {
		result = append(result, metric)
	}

	*m = result
}
