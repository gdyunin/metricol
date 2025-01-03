package entities

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
// It returns true if both metrics have the same name and type; otherwise, it returns false.
func (m *Metric) Equal(compare *Metric) bool {
	return m.Name == compare.Name && m.Type == compare.Type
}

// MetricsRepository represents a repository for storing and retrieving metrics.
//
// Implementations of this interface should provide mechanisms to add a metric
// and retrieve all stored metrics as a collection.
type MetricsRepository interface {
	// Store stores a given metric in the repository.
	// The metric parameter is a pointer to the Metric instance to be added.
	Store(metric *Metric)

	// Metrics retrieves all metrics currently stored in the repository.
	// It returns a slice of pointers to Metric instances.
	Metrics() []*Metric
}

// RepositoryAbstractFactory provides an interface for creating a MetricsRepository instance.
type RepositoryAbstractFactory interface {
	// CreateMetricsRepository creates a MetricsRepository instance.
	CreateMetricsRepository() MetricsRepository
}

// MetricsInterface provides methods to interact with a metrics repository.
//
// This struct serves as an interface between the application logic and the underlying
// metrics repository implementation.
type MetricsInterface struct {
	repo MetricsRepository // Repository for storing and retrieving metrics.
}

// NewMetricsInterface creates a new instance of MetricsInterface.
//
// Parameters:
//   - repo: An implementation of MetricsRepository for managing metrics.
//
// Returns:
//   - A pointer to a newly created MetricsInterface instance.
func NewMetricsInterface(repo MetricsRepository) *MetricsInterface {
	return &MetricsInterface{
		repo: repo,
	}
}

// Store adds a metric to the repository.
//
// Parameters:
//   - metric: A pointer to the Metric instance to be stored.
//
// Returns:
//   - An error if storing the metric fails; otherwise, it returns nil.
func (mi *MetricsInterface) Store(metric *Metric) error {
	mi.repo.Store(metric)
	return nil
}

// Metrics retrieves all metrics from the repository.
//
// Returns:
//   - A slice of pointers to Metric instances currently stored in the repository.
func (mi *MetricsInterface) Metrics() []*Metric {
	return mi.repo.Metrics()
}
