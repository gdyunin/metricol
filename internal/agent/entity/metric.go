package entity

const (
	// MetricTypeCounter represents a metric type that counts occurrences over time.
	MetricTypeCounter = "counter"
	// MetricTypeGauge represents a metric type that measures a value at a specific point in time.
	MetricTypeGauge = "gauge"
)

// Metric represents a metric with a name, type, and associated value.
// The Value field is an interface, allowing it to store values of any type.
type Metric struct {
	Value any    // Value of the metric, which can hold any data type.
	Name  string // Name of the metric.
	Type  string // Type of the metric (e.g., "counter" or "gauge").
}

// Equal compares the current Metric instance with another Metric.
// It returns true if both metrics have the same name and type, otherwise false.
func (m *Metric) Equal(compare *Metric) bool {
	return m.Name == compare.Name && m.Type == compare.Type
}
