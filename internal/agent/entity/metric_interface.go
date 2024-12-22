package entity

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
//   - An error if storing the metric fails (currently always returns nil).
func (mi *MetricsInterface) Store(metric *Metric) error {
	mi.repo.Add(metric)
	return nil
}

// Metrics retrieves all metrics from the repository.
//
// Returns:
//   - A slice of pointers to Metric instances currently stored in the repository.
func (mi *MetricsInterface) Metrics() []*Metric {
	return mi.repo.Metrics()
}
