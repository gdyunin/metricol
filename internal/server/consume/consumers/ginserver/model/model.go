package model

import (
	"errors"
	"strconv"
)

// Metric represents a metric with its associated ID, type, and value fields.
type Metric struct {
	Delta *int64   `json:"delta,omitempty"` // Delta represents the change in the metric (used for counters).
	Value *float64 `json:"value,omitempty"` // Value represents the current value of the metric (used for gauges).
	ID    string   `json:"id" uri:"id"`     // ID uniquely name of the metric.
	MType string   `json:"type" uri:"type"` // MType specifies the type of the metric (e.g., "counter" or "gauge").
}

// StringValue returns the string representation of the metric's value.
// It ensures that either Delta or Value is set but not both.
//
// Returns:
//   - A string representation of the metric's value if either Delta or Value is set.
//   - An error if both Delta and Value are set or if neither field is set.
func (m *Metric) StringValue() (string, error) {
	// Check for invalid state: both Delta and Value are set.
	if m.Delta != nil && m.Value != nil {
		return "", errors.New("invalid metric state: both 'delta' and 'value' fields are set")
	}

	// If Delta is set, return its string representation.
	if m.Delta != nil {
		return strconv.FormatInt(*m.Delta, 10), nil
	}

	// If Value is set, return its string representation.
	if m.Value != nil {
		return strconv.FormatFloat(*m.Value, 'g', -1, 64), nil
	}

	// Return an error if neither Delta nor Value is set.
	return "", errors.New("invalid metric state: both 'delta' and 'value' fields are unset")
}
