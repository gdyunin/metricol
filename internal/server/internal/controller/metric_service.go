// Package controller provides functionality for managing and manipulating metrics
// in the server application. It defines the MetricService which validates, updates,
// retrieves, and persists metrics via an underlying repository.
package controller

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gdyunin/metricol.git/internal/server/internal/entity"
	"github.com/gdyunin/metricol.git/internal/server/repository"
	"github.com/gdyunin/metricol.git/pkg/convert"
)

const (
	pushTimeout    = 3 * time.Second
	pullTimeout    = 3 * time.Second
	pullAllTimeout = 3 * time.Second
)

var ErrNotFoundInRepository = errors.New("not found in repository")

// MetricService provides methods to manage and manipulate metrics.
// It interacts with a repository to validate, store, update, and retrieve metrics.
type MetricService struct {
	repo repository.Repository // repo is the repository for storing and retrieving metrics.
}

// NewMetricService creates and returns a new instance of MetricService.
//
// Parameters:
//   - repo: The repository that handles metric data persistence.
//
// Returns:
//   - *MetricService: A pointer to the newly created MetricService instance.
func NewMetricService(repo repository.Repository) *MetricService {
	return &MetricService{repo: repo}
}

// PushMetric validates the given metric and stores it in the repository.
// It wraps the metric into a batch and calls PushMetrics to process it.
//
// Parameters:
//   - ctx: The context for the operation, supporting cancellation and timeouts.
//   - metric: A pointer to the metric to be validated and stored. The metric must have a valid name, type, and value.
//
// Returns:
//   - *entity.Metric: A pointer to the stored metric if the operation is successful.
//   - error: An error if the metric is invalid or if the repository operation fails.
func (s *MetricService) PushMetric(ctx context.Context, metric *entity.Metric) (*entity.Metric, error) {
	batch := entity.Metrics{metric}
	updated, err := s.PushMetrics(ctx, &batch)
	if err != nil {
		return nil, fmt.Errorf("error while update: %w", err)
	}
	return updated.First(), nil
}

// PushMetrics validates and stores a batch of metrics in the repository.
// It iterates over each metric, validates it, prepares counter metrics,
// merges duplicate entries, and then updates the repository with the batch.
//
// Parameters:
//   - ctx: The context for the operation, supporting cancellation and timeouts.
//   - metrics: A pointer to a collection of metrics to be stored. It must not be nil.
//
// Returns:
//   - *entity.Metrics: A pointer to the updated collection of metrics after storage.
//   - error: An error if any metric fails validation, counter preparation, or if the repository update fails.
func (s *MetricService) PushMetrics(ctx context.Context, metrics *entity.Metrics) (*entity.Metrics, error) {
	if metrics == nil {
		return nil, errors.New("metrics batch is nil")
	}

	pushCtx, cancel := context.WithTimeout(ctx, pushTimeout)
	defer cancel()

	preparedMetricsBatch := make(entity.Metrics, 0, metrics.Length())
	for _, m := range *metrics {
		if err := s.validate(m); err != nil {
			return nil, fmt.Errorf("invalid metric: %w", err)
		}

		if m.Type != entity.MetricTypeCounter {
			preparedMetricsBatch = append(preparedMetricsBatch, m)
		}

		preparedMetric, err := s.prepareCounter(pushCtx, m)
		if err != nil {
			return nil, fmt.Errorf("failed prepare counter %s: %w", m.Name, err)
		}
		preparedMetricsBatch = append(preparedMetricsBatch, preparedMetric)
	}
	preparedMetricsBatch.MergeDuplicates()

	if err := s.repo.UpdateBatch(pushCtx, &preparedMetricsBatch); err != nil {
		return nil, fmt.Errorf("failed store metrics batch: %w", err)
	}
	return &preparedMetricsBatch, nil
}

// prepareCounter processes a counter metric by retrieving any existing value from the repository,
// converting the current and existing values to int64, and summing them.
// If the metric does not already exist, the original metric is returned.
//
// Parameters:
//   - ctx: The context for the operation.
//   - metric: A pointer to the counter metric to prepare.
//
// Returns:
//   - *entity.Metric: A pointer to the updated counter metric.
//   - error: An error if retrieval or conversion of metric values fails.
func (s *MetricService) prepareCounter(ctx context.Context, metric *entity.Metric) (*entity.Metric, error) {
	existingMetric, err := s.repo.Find(ctx, metric.Type, metric.Name)
	if err != nil {
		if errors.Is(err, repository.ErrNotFoundInRepo) {
			return metric, nil
		}
		return nil, fmt.Errorf("retrieval failed for counter '%s': %w", metric.Name, err)
	}

	existingValue, err := convert.AnyToInt64(existingMetric.Value)
	if err != nil {
		return nil, fmt.Errorf("conversion failed for counter '%s': %w", metric.Name, err)
	}

	newValue, err := convert.AnyToInt64(metric.Value)
	if err != nil {
		return nil, fmt.Errorf("conversion failed for counter '%s': %w", metric.Name, err)
	}

	updatedMetric := &entity.Metric{
		Value: existingValue + newValue,
		Name:  metric.Name,
		Type:  entity.MetricTypeCounter,
	}
	return updatedMetric, nil
}

// Pull retrieves a metric by its type and name from the repository.
//
// Parameters:
//   - ctx: The context for the operation, supporting cancellation and timeouts.
//   - metricType: The type of the metric (e.g., "counter" or "gauge").
//   - name: The name of the metric to retrieve.
//
// Returns:
//   - *entity.Metric: A pointer to the retrieved metric if found.
//   - error: An error if the metric is not found or if the repository operation fails.
func (s *MetricService) Pull(ctx context.Context, metricType, name string) (*entity.Metric, error) {
	pullCtx, cancel := context.WithTimeout(ctx, pullTimeout)
	defer cancel()

	metric, err := s.repo.Find(pullCtx, metricType, name)
	if err != nil {
		if errors.Is(err, repository.ErrNotFoundInRepo) {
			return nil, fmt.Errorf(
				"%w: metric with type=%s and name=%s not exist",
				ErrNotFoundInRepository,
				metricType,
				name,
			)
		}
		return nil, fmt.Errorf("retrieval failed for type '%s', name '%s': %w", metricType, name, err)
	}
	return metric, nil
}

// PullAll retrieves all metrics from the repository.
//
// Parameters:
//   - ctx: The context for the operation, supporting cancellation and timeouts.
//
// Returns:
//   - *entity.Metrics: A pointer to the collection of all metrics retrieved.
//   - error: An error if the repository operation fails.
func (s *MetricService) PullAll(ctx context.Context) (*entity.Metrics, error) {
	pullCtx, cancel := context.WithTimeout(ctx, pullAllTimeout)
	defer cancel()

	metrics, err := s.repo.All(pullCtx)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve all metrics: %w", err)
	}
	return metrics, nil
}

// CheckConnection verifies connectivity to the repository by invoking its connection check.
//
// Parameters:
//   - ctx: The context for the operation, supporting cancellation and timeouts.
//
// Returns:
//   - error: An error indicating a connectivity issue if the check fails; nil otherwise.
func (s *MetricService) CheckConnection(ctx context.Context) error {
	if err := s.repo.CheckConnection(ctx); err != nil {
		return fmt.Errorf("failed to check connection to the repository: %w", err)
	}
	return nil
}

// validate checks if the provided metric is valid.
// A valid metric must not be nil and must have a non-empty name, type, and a non-nil value.
//
// Parameters:
//   - metric: A pointer to the metric to validate.
//
// Returns:
//   - error: An error if the metric is invalid; nil if the metric passes validation.
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
