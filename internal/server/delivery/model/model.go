package model

import (
	"github.com/gdyunin/metricol.git/internal/server/internal/entity"
)

// Metric represents the structure used for JSON serialization and deserialization.
// It contains fields for Counter and Gauge metric types and their respective values.
type Metric struct {
	// Delta integer value for Counter metrics. Optional and used only when MType is "counter".
	Delta *int64 `json:"delta,omitempty"`
	// Value floating-point value for Gauge metrics. Optional and used only when MType is "gauge".
	Value *float64 `json:"value,omitempty"`
	// ID unique identifier for the metric.
	ID string `param:"id" json:"id"`
	// MType type of the metric, such as "counter" or "gauge".
	MType string `param:"type" json:"type"`
}

// ToEntityMetric converts a Metric model to an entity.Metric.
//
// This method maps the ID and MType fields directly and assigns the appropriate value
// based on the MType. If MType is "counter", Delta is used. If MType is "gauge",
// Value is used. If neither field is set or MType is invalid, the resulting Value is nil.
//
// Returns:
//   - A pointer to an entity.Metric with values mapped from the Metric model.
func (m *Metric) ToEntityMetric() *entity.Metric {
	metric := entity.Metric{
		Name: m.ID,
		Type: m.MType,
	}

	switch {
	case m.MType == entity.MetricTypeCounter && m.Delta != nil:
		metric.Value = *m.Delta
	case m.MType == entity.MetricTypeGauge && m.Value != nil:
		metric.Value = *m.Value
	default:
		metric.Value = nil
	}

	return &metric
}

// FromEntityMetric converts an entity.Metric to a Metric model.
//
// This function translates an entity.Metric into a Metric model, mapping the Name and
// Type fields to ID and MType, respectively. It also converts the Value field based on
// the MType: for "counter", the Value is assigned to Delta; for "gauge", the Value is
// assigned to Value. If the input metric is nil, the function returns nil.
//
// Parameters:
//   - em: A pointer to an entity.Metric to convert. Can be nil.
//
// Returns:
//   - A pointer to a Metric model with values mapped from the entity.Metric.
//   - Nil if the input metric is nil.
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

// Metrics represents a slice of Metric pointers.
type Metrics []*Metric

// ToEntityMetrics converts a slice of Metric models to a slice of entity.Metric.
//
// This function iterates over each Metric in the slice, converts it to an entity.Metric
// using ToEntityMetric, and appends it to the resulting entity.Metrics slice.
//
// Returns:
//   - A pointer to an entity.Metrics slice with converted metrics.
func (m *Metrics) ToEntityMetrics() *entity.Metrics {
	eMetrics := entity.Metrics{}
	if m == nil {
		return &eMetrics
	}

	for _, model := range *m {
		if model == nil {
			continue
		}
		eMetrics = append(eMetrics, model.ToEntityMetric())
	}

	return &eMetrics
}

// FromEntityMetrics converts a slice of entity.Metric to a slice of Metric models.
//
// This function iterates over each entity.Metric in the slice, converts it to a Metric model
// using FromEntityMetric, and appends it to the resulting Metrics slice.
//
// Parameters:
//   - em: A pointer to an entity.Metrics slice to convert. Can be nil.
//
// Returns:
//   - A pointer to a Metrics slice with converted metrics.
func FromEntityMetrics(em *entity.Metrics) *Metrics {
	models := Metrics{}
	if em == nil {
		return &models
	}

	for _, metric := range *em {
		if metric == nil {
			continue
		}
		models = append(models, FromEntityMetric(metric))
	}

	return &models
}
