package model

import (
	"NewNewMetricol/internal/agent/internal/entity"
	"errors"
	"fmt"
)

// Metric represents a single metric, including its type, ID, and value.
type Metric struct {
	Delta *int64   `json:"delta,omitempty"` // Delta value for counter metrics.
	Value *float64 `json:"value,omitempty"` // Value for gauge metrics.
	ID    string   `json:"id" uri:"id"`     // Unique identifier of the metric.
	MType string   `json:"type" uri:"type"` // Type of the metric (e.g., counter or gauge).
}

// NewFromEntityMetric converts an entity.Metric to a model.Metric.
//
// Parameters:
//   - entityMetric: The source entity.Metric to convert.
//
// Returns:
//   - *Metric: A pointer to the converted Metric.
//   - error: An error if the conversion fails.
func NewFromEntityMetric(entityMetric *entity.Metric) (*Metric, error) {
	if entityMetric == nil {
		return nil, errors.New("entityMetric is nil; cannot perform conversion")
	}

	metric := &Metric{
		ID:    entityMetric.Name,
		MType: entityMetric.Type,
	}

	switch entityMetric.Type {
	case entity.MetricTypeCounter:
		if v, ok := entityMetric.Value.(int64); ok {
			metric.Delta = &v
		} else {
			return nil, fmt.Errorf(
				"unexpected value type for counter metric '%s': got %T, expected int64",
				entityMetric.Name,
				entityMetric.Value,
			)
		}
	case entity.MetricTypeGauge:
		if v, ok := entityMetric.Value.(float64); ok {
			metric.Value = &v
		} else {
			return nil, fmt.Errorf(
				"unexpected value type for gauge metric '%s': got %T, expected float64",
				entityMetric.Name,
				entityMetric.Value,
			)
		}
	default:
		return nil, fmt.Errorf(
			"unsupported metric type '%s' for metric '%s'",
			entityMetric.Type,
			entityMetric.Name,
		)
	}

	return metric, nil
}

// Metrics is a collection of Metric pointers.
type Metrics []*Metric

// NewFromEntityMetrics converts a collection of entity.Metrics to model.Metrics.
//
// Parameters:
//   - entityMetrics: The source entity.Metrics to convert.
//
// Returns:
//   - *Metrics: A pointer to the converted Metrics collection.
//   - error: An error if any metric conversion fails.
func NewFromEntityMetrics(entityMetrics *entity.Metrics) (*Metrics, error) {
	if entityMetrics == nil {
		return &Metrics{}, nil
	}

	metrics := make(Metrics, 0, entityMetrics.Length())
	for _, m := range *entityMetrics {
		metric, err := NewFromEntityMetric(m)
		if err != nil {
			return nil, fmt.Errorf("failed to convert metric '%s': %w", m.Name, err)
		}
		metrics = append(metrics, metric)
	}

	return &metrics, nil
}
