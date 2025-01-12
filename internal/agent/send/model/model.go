package model

import (
	"NewNewMetricol/internal/agent/internal/entity"
	"errors"
	"fmt"
)

// Metric represents a metric in the JSON format used for communication.
//
// Fields:
//   - Delta: Pointer to the value of a counter metric, if applicable.
//   - Value: Pointer to the value of a gauge metric, if applicable.
//   - ID: Identifier of the metric.
//   - MType: Type of the metric (e.g., "counter", "gauge").
type Metric struct {
	Delta *int64   `json:"delta,omitempty"` // Counter value
	Value *float64 `json:"value,omitempty"` // Gauge value
	ID    string   `json:"id" uri:"id"`     // Metric identifier
	MType string   `json:"type" uri:"type"` // Metric type
}

// NewFromEntityMetric converts an `entity.Metric` to a `model.Metric`.
//
// Parameters:
//   - entityMetric: A pointer to an `entity.Metric` structure to be converted.
//
// Returns:
//   - *Metric: A pointer to the converted `Metric` structure.
//   - error: An error if the input is invalid or the conversion fails.
//
// Behavior:
//   - Converts a `MetricTypeCounter` to a `Metric` with the `Delta` field populated.
//   - Converts a `MetricTypeGauge` to a `Metric` with the `Value` field populated.
//   - Returns an error if the input is `nil`, the type is unsupported, or the value is invalid.
func NewFromEntityMetric(entityMetric *entity.Metric) (*Metric, error) {
	// Check for nil input to prevent dereferencing null pointers.
	if entityMetric == nil {
		return nil, errors.New("entityMetric is nil")
	}

	// Initialize the Metric structure.
	metric := &Metric{
		ID:    entityMetric.Name,
		MType: entityMetric.Type,
	}

	// Handle metric conversion based on type.
	switch entityMetric.Type {
	case entity.MetricTypeCounter:
		// Validate and assign the counter value.
		if v, ok := entityMetric.Value.(int64); ok {
			metric.Delta = &v
		} else {
			return nil, fmt.Errorf("invalid value type for counter metric: %T", entityMetric.Value)
		}
	case entity.MetricTypeGauge:
		// Validate and assign the gauge value.
		if v, ok := entityMetric.Value.(float64); ok {
			metric.Value = &v
		} else {
			return nil, fmt.Errorf("invalid value type for gauge metric: %T", entityMetric.Value)
		}
	default:
		// Return an error for unsupported metric types.
		return nil, fmt.Errorf("unsupported metric type: %s", entityMetric.Type)
	}

	// Return the successfully converted Metric.
	return metric, nil
}

type Metrics []*Metric

func NewFromEntityMetrics(entityMetrics *entity.Metrics) (*Metrics, error) {
	if entityMetrics == nil {
		return &Metrics{}, nil
	}

	metrics := make(Metrics, 0, entityMetrics.Length())
	for _, m := range *entityMetrics {
		metric, err := NewFromEntityMetric(m)
		if err != nil {
			return nil, fmt.Errorf("failed to conver metric to model: %w", err)
		}
		metrics = append(metrics, metric)
	}

	return &metrics, nil
}
