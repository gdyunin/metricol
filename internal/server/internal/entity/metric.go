package entity

import (
	"NewNewMetricol/pkg/convert"
	"encoding/json"
	"fmt"
	"strings"
)

const (
	MetricTypeCounter = "counter"
	MetricTypeGauge   = "gauge"
)

type Metric struct {
	Value any    `json:"value"`
	Name  string `json:"name"`
	Type  string `json:"type"`
}

func (m *Metric) UnmarshalJSON(data []byte) error {
	type MetricAlias Metric

	aux := &struct {
		*MetricAlias
	}{
		MetricAlias: (*MetricAlias)(m),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return fmt.Errorf("failed while unmarshalling metric: %w", err)
	}

	if m.Type == MetricTypeCounter {
		v, err := convert.AnyToInt64(m.Value)
		if err != nil {
			return fmt.Errorf("invalid type for value in counter metric: expected number, got %T", m.Value)
		}
		m.Value = v
	}

	return nil
}

type Metrics []*Metric

func (m *Metrics) Length() int {
	if m == nil {
		return 0
	}
	return len(*m)
}

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
