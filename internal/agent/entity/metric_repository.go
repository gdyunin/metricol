package entity

// MetricsRepository represents a repository for storing and retrieving metrics.
//
// Implementations of this interface should provide mechanisms to add a metric
// and retrieve all stored metrics as a collection.
type MetricsRepository interface {
	// Add stores a given metric in the repository.
	// The metric parameter is a pointer to the Metric instance to be added.
	Add(metric *Metric)

	// Metrics retrieves all metrics currently stored in the repository.
	// It returns a slice of pointers to Metric instances.
	Metrics() []*Metric
}
