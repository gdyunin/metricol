package entities

import (
	"errors"
	"fmt"
)

const (
	// MetricTypeCounter represents a metric type that counts occurrences over time.
	// Counters typically store integer values and only increase, except when explicitly reset.
	MetricTypeCounter = "counter"

	// MetricTypeGauge represents a metric type that measures a value at a specific point in time.
	// Gauges can increase or decrease and are commonly used to represent current resource usage.
	MetricTypeGauge = "gauge"
)

// Predefined error messages for common metric operations.
var (
	// ErrPushMetric indicates an error when pushing a metric to the repository.
	ErrPushMetric = errors.New("failed to push the metric to the repository")

	// ErrPullMetric indicates an error when pulling a metric from the repository.
	ErrPullMetric = errors.New("failed to pull the metric from the repository")

	// ErrAllMetricsInRepo indicates an error when retrieving all metrics from the repository.
	ErrAllMetricsInRepo = errors.New("failed to retrieve all metrics from the repository")

	// ErrMetricNotFound indicates an error when a metric is not found in the repository.
	ErrMetricNotFound = errors.New("the requested metric was not found in the repository")
)

// Metric represents a metric with a name, type, and value.
type Metric struct {
	Value any    `json:"value"` // The data associated with the metric.
	Name  string `json:"name"`  // The name of the metric.
	Type  string `json:"type"`  // The type of the metric (e.g., "counter" or "gauge").
}

// AfterJSONUnmarshalling processes the metric after unmarshalling from JSON.
// Ensures the Value field matches the expected type based on the metric type.
//
// Returns:
//   - An error if the metric's value does not match its type.
func (m *Metric) AfterJSONUnmarshalling() error {
	if m.Type == MetricTypeCounter {
		switch v := m.Value.(type) {
		case int64:
			m.Value = v
		case int:
			m.Value = int64(v)
		case float64:
			m.Value = int64(v)
		default:
			return errors.New("invalid value type for counter metric")
		}
	}
	return nil
}

// Filter represents the criteria used to filter metrics in the repository.
type Filter struct {
	Name string // Name of the metric to filter by.
	Type string // Type of the metric to filter by.
}

// MetricsRepository defines the methods that any metric repository must implement.
type MetricsRepository interface {
	// Create adds a new metric to the repository.
	Create(metric *Metric) error

	// Read retrieves a metric from the repository based on the provided filter.
	Read(filter *Filter) (*Metric, error)

	// Update modifies an existing metric in the repository.
	Update(metric *Metric) error

	// IsExists checks whether a metric exists in the repository based on the provided filter.
	IsExists(filter *Filter) (bool, error)

	// All retrieves all metrics stored in the repository.
	All() ([]*Metric, error)
}

// RepositoryAbstractFactory defines a factory for creating MetricsRepository instances.
type RepositoryAbstractFactory interface {
	// CreateMetricsRepository creates a new MetricsRepository instance.
	CreateMetricsRepository() MetricsRepository
}

// MetricsInterface provides methods to manage metrics in a repository.
type MetricsInterface struct {
	repo MetricsRepository
}

// NewMetricsInterface creates a new instance of MetricsInterface with the given repository.
//
// Parameters:
//   - repo: The repository used for storing and managing metrics.
//
// Returns:
//   - A pointer to a newly created MetricsInterface.
func NewMetricsInterface(repo MetricsRepository) *MetricsInterface {
	return &MetricsInterface{repo: repo}
}

// PushMetric pushes a new metric to the repository or updates an existing one.
//
// Parameters:
//   - metric: A pointer to the metric to be added or updated.
//
// Returns:
//   - The updated or newly created metric.
//   - An error if the operation fails.
func (mi *MetricsInterface) PushMetric(metric *Metric) (*Metric, error) {
	if isValid := validateNewMetric(metric); !isValid {
		return nil, fmt.Errorf("%w: invalid metric data: %+v", ErrPushMetric, metric)
	}

	isExists, err := mi.repo.IsExists(&Filter{Name: metric.Name, Type: metric.Type})
	if err != nil {
		return nil, fmt.Errorf("%w: failed to check metric existence: %w", ErrPushMetric, err)
	}

	if !isExists {
		if err = mi.repo.Create(metric); err != nil {
			return nil, fmt.Errorf("%w: failed to create a new metric: %w", ErrPushMetric, err)
		}
		return metric, nil
	}

	m, err := mi.updateExistsMetric(metric)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to update the existing metric: %w", ErrPushMetric, err)
	}
	return m, nil
}

// PullMetric retrieves a metric from the repository by its name and type.
//
// Parameters:
//   - metric: A pointer to the metric to be retrieved.
//
// Returns:
//   - The requested metric.
//   - An error if the metric does not exist or the operation fails.
func (mi *MetricsInterface) PullMetric(metric *Metric) (*Metric, error) {
	isExists, err := mi.IsExists(metric)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to check metric existence: %w", ErrPullMetric, err)
	}

	if !isExists {
		return nil, fmt.Errorf("%w: %w", ErrPullMetric, ErrMetricNotFound)
	}

	m, err := mi.repo.Read(&Filter{Name: metric.Name, Type: metric.Type})
	if err != nil {
		return nil, fmt.Errorf("%w: failed to read the metric from the repository: %w", ErrPullMetric, err)
	}
	return m, nil
}

// IsExists checks if a metric exists in the repository.
//
// Parameters:
//   - metric: A pointer to the metric to check.
//
// Returns:
//   - A boolean indicating whether the metric exists.
//   - An error if the operation fails.
func (mi *MetricsInterface) IsExists(metric *Metric) (bool, error) {
	isExists, err := mi.repo.IsExists(&Filter{Name: metric.Name, Type: metric.Type})
	if err != nil {
		return false, fmt.Errorf("failed to check metric existence in the repository: %w", err)
	}

	return isExists, nil
}

// AllMetricsInRepo retrieves all metrics from the repository.
//
// Returns:
//   - A slice of pointers to all metrics in the repository.
//   - An error if the operation fails.
func (mi *MetricsInterface) AllMetricsInRepo() ([]*Metric, error) {
	m, err := mi.repo.All()
	if err != nil {
		return nil, fmt.Errorf("%w: failed to retrieve metrics: %w", ErrAllMetricsInRepo, err)
	}
	return m, nil
}

// validateNewMetric checks if the provided metric is valid.
//
// Parameters:
//   - metric: A pointer to the metric to validate.
//
// Returns:
//   - A boolean indicating whether the metric is valid.
func validateNewMetric(metric *Metric) bool {
	if metric.Name == "" || metric.Type == "" {
		return false
	}

	switch metric.Type {
	case MetricTypeCounter:
		_, ok := metric.Value.(int64)
		return ok
	case MetricTypeGauge:
		_, ok := metric.Value.(float64)
		return ok
	default:
		return false
	}
}

// updateExistsMetric updates an existing metric in the repository based on its type.
//
// Parameters:
//   - metric: A pointer to the metric to update.
//
// Returns:
//   - The updated metric.
//   - An error if the update operation fails.
func (mi *MetricsInterface) updateExistsMetric(metric *Metric) (*Metric, error) {
	if metric.Type == MetricTypeCounter {
		m, err := mi.repo.Read(&Filter{Name: metric.Name, Type: metric.Type})
		if err != nil {
			return nil, fmt.Errorf("failed to read the existing metric: %w", err)
		}

		newValue, _ := metric.Value.(int64)
		storedValue, _ := m.Value.(int64)
		metric.Value = storedValue + newValue
	}

	err := mi.repo.Update(metric)
	if err != nil {
		return nil, fmt.Errorf("failed to update the metric in the repository: %w", err)
	}
	return metric, nil
}
