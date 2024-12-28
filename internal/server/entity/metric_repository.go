package entity

// MetricRepository defines the methods that any metric repository must implement.
//
// The interface provides a contract for operations on metrics, including creation, retrieval,
// updates, existence checks, and listing all stored metrics.
type MetricRepository interface {
	// Create adds a new metric to the repository.
	// Returns an error if the operation fails, such as due to duplicate entries or storage issues.
	Create(metric *Metric) error

	// Read retrieves a metric from the repository based on the provided filter.
	// If no matching metric is found or if an error occurs during retrieval, it returns an error.
	Read(filter *Filter) (*Metric, error)

	// Update modifies an existing metric in the repository.
	// Returns an error if the metric does not exist or if the update operation fails.
	Update(metric *Metric) error

	// IsExists checks whether a metric exists in the repository based on the provided filter.
	// Returns a boolean indicating the existence of the metric and an error if the operation fails.
	IsExists(filter *Filter) (bool, error)

	// All retrieves all metrics stored in the repository.
	// Returns a slice of metrics and an error if the operation fails.
	All() ([]*Metric, error)
}

// Filter represents the criteria used to filter metrics in the repository.
//
// Fields:
// - Name: The name of the metric to filter by. If empty, no filtering is applied by name.
// - Type: The type of the metric to filter by (e.g., "counter" or "gauge"). If empty, no filtering is applied by type.
type Filter struct {
	Name string // Name of the metric to filter by.
	Type string // Type of the metric to filter by.
}
