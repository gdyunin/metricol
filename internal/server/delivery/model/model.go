// Package model defines the data structures and conversion functions used to map
// between the internal entity representation of a metric and the model representation
// used for JSON serialization and deserialization. This package supports both counter
// and gauge metric types.
package model

import (
	"github.com/gdyunin/metricol.git/internal/server/internal/entity"
	"github.com/gdyunin/metricol.git/pkg/convert"
)

// Metric represents the structure used for JSON serialization and deserialization of metrics.
// It includes optional fields for Counter and Gauge metrics. For counter metrics, the Delta field
// is used, whereas for gauge metrics, the Value field is used.
// The ID field corresponds to the unique identifier of the metric, and MType indicates the metric type.
type Metric struct {
	// Delta holds the integer value for counter metrics.
	// It is optional and is only used when MType is "counter".
	Delta *int64 `json:"delta,omitempty"`
	// Value holds the floating-point value for gauge metrics.
	// It is optional and is only used when MType is "gauge".
	Value *float64 `json:"value,omitempty"`
	// ID is the unique identifier for the metric.
	ID string `json:"id"              param:"id"`
	// MType represents the type of the metric, such as "counter" or "gauge".
	MType string `json:"type"            param:"type"`
}

// ToEntityMetric converts a Metric model to an entity.Metric.
// It maps the ID and MType fields directly and assigns the appropriate value based on the metric type.
// If MType is "counter" and Delta is non-nil, Delta is used; if MType is "gauge" and Value is non-nil,
// Value is used; otherwise, the Value field of the resulting entity.Metric is set to nil.
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
// It maps the Name and Type fields to ID and MType respectively, and converts the Value field
// based on the metric type: for "counter", it converts the value to an integer (Delta),
// and for "gauge", it assigns the value to Value.
// If the input entity.Metric is nil, the function returns nil.
//
// Parameters:
//   - em: A pointer to an entity.Metric to convert. Can be nil.
//
// Returns:
//   - A pointer to a Metric model with values mapped from the entity.Metric, or nil if the input is nil.
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
		if value, err := convert.AnyToInt64(em.Value); err == nil {
			metric.Delta = &value
		}
	case entity.MetricTypeGauge:
		if value, ok := em.Value.(float64); ok {
			metric.Value = &value
		}
	}

	return &metric
}

// Metrics represents a slice of pointers to Metric models.
type Metrics []*Metric

// ToEntityMetrics converts a slice of Metric models to a slice of entity.Metric.
// It iterates over each Metric in the slice, converts it to an entity.Metric using ToEntityMetric,
// and appends the result to an entity.Metrics collection.
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
// It iterates over each entity.Metric in the input collection, converts it to a Metric model
// using FromEntityMetric, and appends the result to a Metrics slice.
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
