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
func (s *MetricService) PushMetric(ctx context.Context, metric *entity.Metric) (*entity.Metric, error) {
	batch := entity.Metrics{metric}
	updated, err := s.PushMetrics(ctx, &batch)
	if err != nil {
		return nil, fmt.Errorf("error while update: %w", err)
	}
	return updated.First(), nil
}

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

	if err := s.repo.UpdateBatch(pushCtx, &preparedMetricsBatch); err != nil {
		return nil, fmt.Errorf("failed store metrics batch: %w", err)
	}
	return &preparedMetricsBatch, nil
}

func (s *MetricService) prepareCounter(ctx context.Context, metric *entity.Metric) (*entity.Metric, error) {
	exist, err := s.repo.IsExist(ctx, metric.Type, metric.Name)
	if err != nil {
		return nil, fmt.Errorf("error occurred while check metric %s exist in repo: %w", metric.Name, err)
	}

	if !exist {
		return metric, nil
	}

	existingMetric, err := s.repo.Find(ctx, metric.Type, metric.Name)
	if err != nil {
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
//   - metricType: The type of the metric (e.g., "counter" or "gauge").
//   - name: The name of the metric to retrieve.
//
// Returns:
//   - The retrieved metric if found.
//   - Nil if the metric does not exist.
//   - An error if the repository operation fails.
func (s *MetricService) Pull(ctx context.Context, metricType, name string) (*entity.Metric, error) {
	pullCtx, cancel := context.WithTimeout(ctx, pullTimeout)
	defer cancel()

	// TODO: Оооочень спорно дробить на две операции: проверка существования и только потом получение.
	// TODO: Подумать и скорее всего убрать отсюда вызов isExist и генерировать какую-то "типизированную" ошибку
	// TODO: Сразу при попытке получения метрики.
	exists, err := s.repo.IsExist(ctx, metricType, name)
	if err != nil {
		return nil, fmt.Errorf("repository check failed for type '%s', name '%s': %w", metricType, name, err)
	}
	// TODO: Вот эту часть заменить "типизированной" ошибкой.
	if !exists {
		return nil, fmt.Errorf(
			"%w: metric with type=%s and name=%s not exist",
			ErrNotFoundInRepository,
			metricType,
			name,
		)
	}

	metric, err := s.repo.Find(pullCtx, metricType, name)
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
func (s *MetricService) PullAll(ctx context.Context) (*entity.Metrics, error) {
	pullCtx, cancel := context.WithTimeout(ctx, pullAllTimeout)
	defer cancel()

	metrics, err := s.repo.All(pullCtx)
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
