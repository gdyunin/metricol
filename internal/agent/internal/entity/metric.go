package entity

import (
	"fmt"
	"strings"
)

const (
	MetricTypeCounter = "counter"
	MetricTypeGauge   = "gauge"
)

type Metric struct {
	Name       string
	Type       string
	Value      any
	IsMetadata bool
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
