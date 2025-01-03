package model

// Metric represents a single metric with its name, type, and value.
type Metric struct {
	Value any    // Value of the metric, which can be of any type depending on the metric.
	Name  string // Name of the metric, describing what it measures.
	Type  string // Type of the metric, e.g. "gauge" or "counter".
}
