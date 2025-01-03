package model

// Metric represents a metric with its associated data.
// It contains the metric's identifier, type, and optional values for Delta and Value.
type Metric struct {
	Delta *int64   `json:"delta,omitempty"` // Delta is an optional field representing the metric's delta value.
	Value *float64 `json:"value,omitempty"` // Value is an optional field representing the metric's current value.
	ID    string   `json:"id" uri:"id"`     // ID is the unique name for the metric.
	MType string   `json:"type" uri:"type"` // MType indicates the type of the metric (e.g., gauge, counter).
}
