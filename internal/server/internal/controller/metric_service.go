package controller

import (
	"context"
	"errors"
	"fmt"

	"github.com/gdyunin/metricol.git/internal/server/internal/entity"
	"github.com/gdyunin/metricol.git/internal/server/repository"
	"github.com/gdyunin/metricol.git/pkg/convert"
)

var ErrNotFoundInRepository = errors.New("not found in repository")

// MetricService provides methods to manage and manipulate metrics.
type MetricService struct {
	repo repository.Repository // Repository for storing and retrieving metrics.
}

// NewMetricService creates a new instance of MetricService.
//
// Parameters:
//   - repo: Repository for handling metric data persistence.
//
// Returns:
//   - A new instance of MetricService.
func NewMetricService(repo repository.Repository) *MetricService {
	return &MetricService{repo: repo}
}

// PushMetric validates the given metric and stores it in the repository.
//
// Parameters:
//   - metric: The metric to validate and store. It must not be nil and should have a valid name, type, and value.
//
// Returns:
//   - The stored metric if successful.
//   - An error if the metric is invalid or the repository operation fails.
func (s *MetricService) PushMetric(metric *entity.Metric) (*entity.Metric, error) {
	if err := s.validate(metric); err != nil {
		return nil, fmt.Errorf("invalid metric: %w", err)
	}

	switch metric.Type {
	case entity.MetricTypeCounter:
		return s.handleCounter(metric)
	default:
		return s.saveMetric(metric)
	}
}

// Pull retrieves a metric by its type and name from the repository.
//
// Parameters:
//   - metricType: The type of the metric (e.g., "counter" or "gauge").
//   - name: The name of the metric to retrieve.
//
// Returns:
//   - The retrieved metric if found.
//   - Nil if the metric does not exist.
//   - An error if the repository operation fails.
func (s *MetricService) Pull(metricType, name string) (*entity.Metric, error) {
	exists, err := s.repo.IsExist(metricType, name)
	if err != nil {
		return nil, fmt.Errorf("repository check failed for type '%s', name '%s': %w", metricType, name, err)
	}
	if !exists {
		return nil, fmt.Errorf(
			"%w: metric with type=%s and name=%s not exist",
			ErrNotFoundInRepository,
			metricType,
			name,
		)
	}
	metric, err := s.repo.Find(metricType, name)
	if err != nil {
		return nil, fmt.Errorf("retrieval failed for type '%s', name '%s': %w", metricType, name, err)
	}
	return metric, nil
}

// PullAll retrieves all metrics from the repository.
//
// Returns:
//   - A collection of all metrics.
//   - An error if the repository operation fails.
func (s *MetricService) PullAll() (*entity.Metrics, error) {
	metrics, err := s.repo.All()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve all metrics: %w", err)
	}
	return metrics, nil
}

// CheckConnection verifies the connectivity to the repository.
//
// Parameters:
//   - ctx: The context used to control cancellation and timeout for the operation.
//
// Returns:
//   - error: An error indicating the connectivity issue, or nil if the connection is valid.
func (s *MetricService) CheckConnection(ctx context.Context) error {
	if err := s.repo.CheckConnection(ctx); err != nil {
		return fmt.Errorf("failed to check connection to the repository: %w", err)
	}
	return nil
}

// validate checks if the given metric is valid.
//
// Parameters:
//   - metric: The metric to validate. It must not be nil and should have a non-empty name, type, and value.
//
// Returns:
//   - An error if the metric is invalid; otherwise, nil.
func (s *MetricService) validate(metric *entity.Metric) error {
	if metric == nil {
		return errors.New("metric is nil")
	}
	if metric.Name == "" {
		return errors.New("metric name is missing")
	}
	if metric.Type == "" {
		return errors.New("metric type is missing")
	}
	if metric.Value == nil {
		return errors.New("metric value is missing")
	}
	return nil
}

// toInt64 validates and converts the value of the given metric to int64.
//
// Parameters:
//   - metric: The metric whose value needs conversion.
//
// Returns:
//   - The converted value as int64.
//   - An error if the conversion fails.
func (s *MetricService) toInt64(metric *entity.Metric) (int64, error) {
	value, err := convert.AnyToInt64(metric.Value)
	if err != nil {
		return 0, fmt.Errorf("conversion failed for '%s': %w", metric.Name, err)
	}
	return value, nil
}

// saveMetric stores the given metric in the repository.
//
// Parameters:
//   - metric: The metric to store.
//
// Returns:
//   - The stored metric.
//   - An error if the repository operation fails.
func (s *MetricService) saveMetric(metric *entity.Metric) (*entity.Metric, error) {
	if err := s.repo.Update(metric); err != nil {
		return nil, fmt.Errorf("storage failed for '%s': %w", metric.Name, err)
	}
	return metric, nil
}

// createCounter creates and stores a new counter metric in the repository.
//
// Parameters:
//   - metric: The metric to create.
//   - value: The initial value of the counter.
//
// Returns:
//   - The created counter metric.
//   - An error if the repository operation fails.
func (s *MetricService) createCounter(metric *entity.Metric, value int64) (*entity.Metric, error) {
	newMetric := &entity.Metric{
		Value: value,
		Name:  metric.Name,
		Type:  entity.MetricTypeCounter,
	}
	storedMetric, err := s.saveMetric(newMetric)
	if err != nil {
		return nil, fmt.Errorf("creation failed for counter '%s': %w", metric.Name, err)
	}
	return storedMetric, nil
}

// updateCounter increments and stores the value of an existing counter metric.
//
// Parameters:
//   - metric: The metric to update.
//   - value: The value to increment the counter by.
//
// Returns:
//   - The updated counter metric.
//   - An error if the repository operation fails.
func (s *MetricService) updateCounter(metric *entity.Metric, value int64) (*entity.Metric, error) {
	existingMetric, err := s.repo.Find(metric.Type, metric.Name)
	if err != nil {
		return nil, fmt.Errorf("retrieval failed for counter '%s': %w", metric.Name, err)
	}
	existingValue, err := convert.AnyToInt64(existingMetric.Value)
	if err != nil {
		return nil, fmt.Errorf("conversion failed for counter '%s': %w", metric.Name, err)
	}
	updatedMetric := &entity.Metric{
		Value: existingValue + value,
		Name:  metric.Name,
		Type:  entity.MetricTypeCounter,
	}
	storedMetric, err := s.saveMetric(updatedMetric)
	if err != nil {
		return nil, fmt.Errorf("update failed for counter '%s': %w", metric.Name, err)
	}
	return storedMetric, nil
}

// handleCounter processes and stores a counter metric in the repository.
// If the counter does not exist, it is created. Otherwise, it is updated.
//
// Parameters:
//   - metric: The counter metric to process.
//
// Returns:
//   - The processed counter metric.
//   - An error if any validation or repository operation fails.
func (s *MetricService) handleCounter(metric *entity.Metric) (*entity.Metric, error) {
	value, err := s.toInt64(metric)
	if err != nil {
		return nil, fmt.Errorf("validation failed for counter '%s': %w", metric.Name, err)
	}

	exists, err := s.repo.IsExist(metric.Type, metric.Name)
	if err != nil {
		return nil, fmt.Errorf("existence check failed for counter '%s': %w", metric.Name, err)
	}

	if !exists {
		return s.createCounter(metric, value)
	}

	return s.updateCounter(metric, value)
}
