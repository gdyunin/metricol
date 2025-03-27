package repository

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/gdyunin/metricol.git/internal/server/internal/entity"

	"go.uber.org/zap"
)

// InMemoryRepository implements a thread-safe in-memory storage for metrics.
// Metrics are stored in a nested map organized by metric type and name.
type InMemoryRepository struct {
	storage map[string]map[string]any // storage maps metric type to a map of metric name to value.
	mu      *sync.RWMutex             // mu synchronizes access to the storage.
	logger  *zap.SugaredLogger        // logger is used for logging repository operations.
}

// NewInMemoryRepository creates a new instance of InMemoryRepository.
//
// Parameters:
//   - logger: A logger instance for recording repository operations.
//
// Returns:
//   - *InMemoryRepository: A pointer to the newly created InMemoryRepository.
func NewInMemoryRepository(logger *zap.SugaredLogger) *InMemoryRepository {
	return &InMemoryRepository{
		storage: make(map[string]map[string]any),
		mu:      &sync.RWMutex{},
		logger:  logger,
	}
}

// Update adds or updates a metric in the repository.
// It stores the metric value under its type and name.
//
// Parameters:
//   - ctx: The context for the operation.
//   - metric: A pointer to the Metric entity to update or add.
//
// Returns:
//   - error: An error if the metric is nil or the update fails.
func (r *InMemoryRepository) Update(_ context.Context, metric *entity.Metric) error {
	if metric == nil {
		return errors.New("metric should be non-nil, but got nil")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if r.storage[metric.Type] == nil {
		r.storage[metric.Type] = make(map[string]any)
	}

	r.storage[metric.Type][metric.Name] = metric.Value
	return nil
}

// UpdateBatch adds or updates a batch of metrics in the repository.
// It iterates through the metrics collection and calls Update for each metric.
//
// Parameters:
//   - ctx: The context for the operation.
//   - metrics: A pointer to the collection of Metrics to update.
//
// Returns:
//   - error: An error if any individual metric update fails.
func (r *InMemoryRepository) UpdateBatch(ctx context.Context, metrics *entity.Metrics) error {
	if metrics == nil {
		return errors.New("metrics should be non-nil, but got nil")
	}

	for _, m := range *metrics {
		if err := r.Update(ctx, m); err != nil {
			return fmt.Errorf("failed update one of metrics: %w", err)
		}
	}

	return nil
}

// Find retrieves a metric from the repository by its type and name.
// It returns the metric if it exists or an error if it is not found.
//
// Parameters:
//   - ctx: The context for the operation.
//   - metricType: The type of the metric.
//   - name: The name of the metric.
//
// Returns:
//   - *entity.Metric: A pointer to the retrieved Metric.
//   - error: An error if the metric does not exist.
func (r *InMemoryRepository) Find(_ context.Context, metricType string, name string) (*entity.Metric, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	value, exist := r.storage[metricType][name]
	if !exist {
		return nil, fmt.Errorf("%w: type=%s, name=%s", ErrNotFoundInRepo, metricType, name)
	}

	return &entity.Metric{
		Value: value,
		Name:  name,
		Type:  metricType,
	}, nil
}

// All retrieves all metrics stored in the repository.
// It compiles the metrics from the internal storage map into a Metrics slice.
//
// Parameters:
//   - ctx: The context for the operation.
//
// Returns:
//   - *entity.Metrics: A pointer to the collection of all metrics.
//   - error: An error if retrieval fails.
func (r *InMemoryRepository) All(_ context.Context) (*entity.Metrics, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	metrics := entity.Metrics{}
	for metricType, metricMap := range r.storage {
		for name, value := range metricMap {
			metrics = append(metrics, &entity.Metric{
				Value: value,
				Name:  name,
				Type:  metricType,
			})
		}
	}

	return &metrics, nil
}

// CheckConnection checks the connection status of the repository.
// Since the repository is in-memory, it always returns nil.
//
// Parameters:
//   - ctx: The context for the operation.
//
// Returns:
//   - error: Always nil.
func (r *InMemoryRepository) CheckConnection(_ context.Context) error {
	return nil
}
