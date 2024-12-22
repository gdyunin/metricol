package entity

import (
	"errors"
	"fmt"
)

// Predefined error messages for common metric operations.
var (
	// ErrPushMetric indicates an error when pushing a metric to the repository.
	ErrPushMetric = errors.New("failed to push metric")

	// ErrPullMetric indicates an error when pulling a metric from the repository.
	ErrPullMetric = errors.New("failed to pull metric")

	// ErrAllMetricsInRepo indicates an error when retrieving all metrics from the repository.
	ErrAllMetricsInRepo = errors.New("failed to retrieve all metrics")
)

// MetricsInterface provides methods to manage metrics in a repository.
type MetricsInterface struct {
	repo MetricRepository // The repository used for storing and managing metrics.
}

// NewMetricsInterface creates a new instance of `MetricsInterface` with the given repository.
// Returns a pointer to the newly created `MetricsInterface`.
func NewMetricsInterface(repo MetricRepository) *MetricsInterface {
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
		return nil, fmt.Errorf("%w: metric not found", ErrPullMetric)
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
