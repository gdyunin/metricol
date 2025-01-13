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
type Metric struct {
	Value any    `json:"value"` // The value of the metric.
	Name  string `json:"name"`  // The name of the metric.
	Type  string `json:"type"`  // The type of the metric, e.g., "counter" or "gauge".
}

// UnmarshalJSON implements custom JSON unmarshalling for the Metric type.
func (m *Metric) UnmarshalJSON(data []byte) error {
	type MetricAlias Metric // Alias to avoid recursion during unmarshalling.
	aux := &struct {
		*MetricAlias
	}{
		MetricAlias: (*MetricAlias)(m),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return fmt.Errorf("unable to parse metric JSON: %w", err)
	}

	if m.Type == MetricTypeCounter {
		v, err := convert.AnyToInt64(m.Value)
		if err != nil {
			return fmt.Errorf("invalid value for counter metric \"%s\": expected integer, got %T", m.Name, m.Value)
		}
		m.Value = v
	}

	return nil
}

// Metrics represents a collection of metrics.
type Metrics []*Metric

// Length returns the number of metrics in the collection.
func (m *Metrics) Length() int {
	if m == nil {
		return 0
	}
	return len(*m)
}

// ToString returns a string representation of the metrics collection.
func (m *Metrics) ToString() string {
	if m == nil {
		return ""
	}

	strData := make([]string, 0, m.Length())
	for _, metric := range *m {
		if metric != nil {
			strData = append(strData, fmt.Sprintf("<Name=%s Type=%s Value=%v>", metric.Name, metric.Type, metric.Value))
		} else {
			strData = append(strData, "<nil>")
		}
	}

	return strings.Join(strData, ", ")
}
