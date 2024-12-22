package entity

const (
	// MetricTypeCounter represents a metric type that counts occurrences over time.
	// Counters typically store integer values and only increase, except when explicitly reset.
	MetricTypeCounter = "counter"

	// MetricTypeGauge represents a metric type that measures a value at a specific point in time.
	// Gauges can increase or decrease and are commonly used to represent things like temperatures or current resource usage.
	MetricTypeGauge = "gauge"
)

// Metric represents a metric with a name, type, and value.
//
// Fields:
//   - Name: A descriptive identifier for the metric.
//   - Type: Specifies the metric type (e.g., "counter" or "gauge").
//   - Value: The actual data associated with the metric, whose type depends on the metric's type.
//     Counters typically use integers, while gauges may use floating-point numbers.
type Metric struct {
	// Name is the name for the metric.
	// It should be a descriptive and unique string that clearly defines the metric's purpose.
	Name string

	// Type specifies the type of the metric.
	// Valid types are defined as constants in this package, such as "counter" and "gauge".
	Type string

	// Value holds the data associated with the metric.
	// The data type depends on the metric's type:
	// - Counters typically use integers.
	// - Gauges often use floating-point numbers.
	Value any
}
