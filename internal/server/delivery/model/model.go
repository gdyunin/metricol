package model

import (
	"NewNewMetricol/internal/server/internal/entity"
)

type Metric struct {
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
	ID    string   `param:"id" json:"id"`
	MType string   `param:"type" json:"type"`
}

func (m *Metric) ToEntityMetric() *entity.Metric {
	metric := entity.Metric{
		Name: m.ID,
		Type: m.MType,
	}

	switch m.MType {
	case entity.MetricTypeCounter:
		if m.Delta != nil {
			metric.Value = *m.Delta
		} else {
			metric.Value = nil
		}
	case entity.MetricTypeGauge:
		if m.Value != nil {
			metric.Value = *m.Value
		} else {
			metric.Value = nil
		}
	default:
		metric.Value = nil
	}

	return &metric
}

func FromEntityMetric(em *entity.Metric) *Metric {
	if em == nil {
		return nil
	}

	metric := Metric{
		ID:    em.Name,
		MType: em.Type,
	}

	switch em.Type {
	case entity.MetricTypeCounter:
		if value, ok := em.Value.(int64); ok {
			metric.Delta = &value
		}
	case entity.MetricTypeGauge:
		if value, ok := em.Value.(float64); ok {
			metric.Value = &value
		}
	}

	return &metric
}
