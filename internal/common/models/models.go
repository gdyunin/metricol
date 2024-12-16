package models

// Metrics represents a data structure for storing metric information.
// It includes the metric`s ID, type, and its value or delta depending on the type.
type Metrics struct {
	ID    string   `json:"id"`              // ID of the metric.
	MType string   `json:"type"`            // Type of the metric, which can be either "gauge" or "counter".
	Delta *int64   `json:"delta,omitempty"` // The delta value for counter metrics, if applicable.
	Value *float64 `json:"value,omitempty"` // The value for gauge metrics, if applicable.
}
