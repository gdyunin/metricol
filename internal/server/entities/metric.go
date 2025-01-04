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
	// Gauges can increase or decrease and are commonly used to represent things current resource usage.
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
	// Value holds the data associated with the metric.
	// The data type depends on the metric's type:
	// - Counters typically use integers.
	// - Gauges often use floating-point numbers.
	Value any `json:"value"`

	// Name is the name for the metric.
	// It should be a descriptive and unique string that clearly defines the metric's purpose.
	Name string `json:"name"`

	// Type specifies the type of the metric.
	// Valid types are defined as constants in this package, such as "counter" and "gauge".
	Type string `json:"type"`
}

func (m *Metric) AfterJSONUnmarshalling() error {
	if m.Type == MetricTypeCounter {
		switch v := m.Value.(type) {
		case int64:
			m.Value = v
		case int:
			m.Value = int64(v)
		case float64:
			m.Value = int64(v) // Явное преобразование float64 -> int64
		default:
			return errors.New("invalid value type for counter")
		}
	}
	return nil
}

// Predefined error messages for common metric operations.
var (
	// ErrPushMetric indicates an error when pushing a metric to the repository.
	ErrPushMetric = errors.New("failed to push metric")

	// ErrPullMetric indicates an error when pulling a metric from the repository.
	ErrPullMetric = errors.New("failed to pull metric")

	// ErrAllMetricsInRepo indicates an error when retrieving all metrics from the repository.
	ErrAllMetricsInRepo = errors.New("failed to retrieve all metrics")

	ErrMetricNotFound = errors.New("metric not found in repository")
)

// MetricsInterface provides methods to manage metrics in a repository.
type MetricsInterface struct {
	repo MetricsRepository // The repository used for storing and managing metrics.
}

// NewMetricsInterface creates a new instance of `MetricsInterface` with the given repository.
// Returns a pointer to the newly created `MetricsInterface`.
func NewMetricsInterface(repo MetricsRepository) *MetricsInterface {
	return &MetricsInterface{repo: repo}
}

// PushMetric pushes a new metric to the repository or updates an existing one.
// Returns the updated metric and an error if the operation fails.
func (mi *MetricsInterface) PushMetric(metric *Metric) (*Metric, error) {
	if isValid := validateNewMetric(metric); !isValid {
		return nil, fmt.Errorf("%w: invalid metric data: %+v", ErrPushMetric, metric)
	}

	isExists, err := mi.repo.IsExists(&Filter{Name: metric.Name, Type: metric.Type})
	if err != nil {
		return nil, fmt.Errorf("%w: error checking metric existence: %w", ErrPushMetric, err)
	}

	if !isExists {
		// Create the metric if it does not exist.
		if err = mi.repo.Create(metric); err != nil {
			return nil, fmt.Errorf("%w: error creating new metric: %w", ErrPushMetric, err)
		}
		return metric, nil
	}

	// Update the existing metric.
	m, err := mi.updateExistsMetric(metric)
	if err != nil {
		return nil, fmt.Errorf("%w: error updating existing metric: %w", ErrPushMetric, err)
	}
	return m, nil
}

// PullMetric retrieves a metric from the repository by its name and type.
// Returns the metric and an error if the operation fails or if the metric does not exist.
func (mi *MetricsInterface) PullMetric(metric *Metric) (*Metric, error) {
	isExists, err := mi.IsExists(metric)
	if err != nil {
		return nil, fmt.Errorf("%w: error checking metric existence: %w", ErrPullMetric, err)
	}

	if !isExists {
		return nil, fmt.Errorf("%w: %w", ErrPullMetric, ErrMetricNotFound)
	}

	m, err := mi.repo.Read(&Filter{Name: metric.Name, Type: metric.Type})
	if err != nil {
		return nil, fmt.Errorf("%w: error reading metric from repository: %w", ErrPullMetric, err)
	}
	return m, nil
}

func (mi *MetricsInterface) IsExists(metric *Metric) (bool, error) {
	isExists, err := mi.repo.IsExists(&Filter{Name: metric.Name, Type: metric.Type})
	if err != nil {
		return false, fmt.Errorf("error checking metric existence in repo: %w", err)
	}

	return isExists, nil
}

// AllMetricsInRepo retrieves all metrics from the repository.
// Returns a slice of metrics and an error if the operation fails.
func (mi *MetricsInterface) AllMetricsInRepo() ([]*Metric, error) {
	m, err := mi.repo.All()
	if err != nil {
		return nil, fmt.Errorf("%w: error retrieving metrics: %w", ErrAllMetricsInRepo, err)
	}
	return m, nil
}

// validateNewMetric checks if the provided metric is valid.
// Returns true if the metric has a non-empty name, a valid type, and a value of the expected type.
//
// - MetricTypeCounter requires an `int64` value.
// - MetricTypeGauge requires a `float64` value.
func validateNewMetric(metric *Metric) bool {
	if metric.Name == "" || metric.Type == "" {
		return false
	}

	switch metric.Type {
	case MetricTypeCounter:
		_, ok := metric.Value.(int64) // Counter metrics must have an int64 value.
		return ok
	case MetricTypeGauge:
		_, ok := metric.Value.(float64) // Gauge metrics must have a float64 value.
		return ok
	default:
		return false // Unsupported metric types are considered invalid.
	}
}

// updateExistsMetric updates an existing metric in the repository based on its type.
// For Counter metrics, it increments the stored value by the new value.
// Returns the updated metric and an error if the operation fails.
func (mi *MetricsInterface) updateExistsMetric(metric *Metric) (*Metric, error) {
	if metric.Type == MetricTypeCounter {
		m, err := mi.repo.Read(&Filter{Name: metric.Name, Type: metric.Type})
		if err != nil {
			return nil, fmt.Errorf("error reading existing metric: %w", err)
		}

		newValue, _ := metric.Value.(int64)   // New value to add.
		storedValue, _ := m.Value.(int64)     // Current stored value.
		metric.Value = storedValue + newValue // Increment the value.
	}

	err := mi.repo.Update(metric)
	if err != nil {
		return nil, fmt.Errorf("error updating metric in repository: %w", err)
	}
	return metric, nil
}

// MetricsRepository defines the methods that any metric repository must implement.
//
// The interface provides a contract for operations on metrics, including creation, retrieval,
// updates, existence checks, and listing all stored metrics.
type MetricsRepository interface {
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

type RepositoryAbstractFactory interface {
	// CreateMetricsRepository creates a MetricsRepository instance.
	CreateMetricsRepository() MetricsRepository
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
