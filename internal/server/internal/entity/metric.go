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

// First returns the first metric in the collection or nil if the collection is empty or nil.
func (m *Metrics) First() *Metric {
	if m == nil || len(*m) == 0 {
		return nil
	}
	return (*m)[0]
}

// String returns a string representation of the metrics collection.
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

// MergeDuplicates merges duplicate metrics in the collection by name and type, aggregating their values.
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
				repeatVal, _ := convert.AnyToInt64(existing.Value)
				existing.Value = existingVal + repeatVal
			} else {
				existing.Value = metric.Value // For gauge, simply replace the value.
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
