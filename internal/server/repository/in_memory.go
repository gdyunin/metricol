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
type InMemoryRepository struct {
	storage map[string]map[string]any // Storage for metrics organized by type and name.
	mu      *sync.RWMutex             // Mutex to synchronize access to the storage.
	logger  *zap.SugaredLogger        // Logger for logging repository operations.
}

// NewInMemoryRepository creates a new instance of InMemoryRepository.
//
// Parameters:
//   - logger: A logger instance for recording operations.
//
// Returns:
//   - A pointer to a new InMemoryRepository instance.
func NewInMemoryRepository(logger *zap.SugaredLogger) *InMemoryRepository {
	return &InMemoryRepository{
		storage: make(map[string]map[string]any),
		mu:      &sync.RWMutex{},
		logger:  logger,
	}
}

// Update updates or adds a metric in the repository.
//
// Parameters:
//   - metric: A pointer to the Metric entity to update or add.
//
// Returns:
//   - An error if the operation fails.
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

// IsExist checks if a metric exists in the repository.
//
// Parameters:
//   - metricType: The type of the metric.
//   - name: The name of the metric.
//
// Returns:
//   - A boolean indicating whether the metric exists.
//   - An error if the operation fails.
func (r *InMemoryRepository) IsExist(_ context.Context, metricType string, name string) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, exist := r.storage[metricType][name]
	return exist, nil
}

// Find retrieves a metric from the repository by type and name.
//
// Parameters:
//   - metricType: The type of the metric.
//   - name: The name of the metric.
//
// Returns:
//   - A pointer to the Metric entity if found.
//   - An error if the metric does not exist or another issue occurs.
func (r *InMemoryRepository) Find(_ context.Context, metricType string, name string) (*entity.Metric, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	value, exist := r.storage[metricType][name]
	if !exist {
		return nil, fmt.Errorf("metric not found: type=%s, name=%s", metricType, name)
	}

	return &entity.Metric{
		Value: value,
		Name:  name,
		Type:  metricType,
	}, nil
}

// All retrieves all metrics from the repository.
//
// Returns:
//   - A pointer to a Metrics slice containing all metrics.
//   - An error if the operation fails.
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
//
// Parameters:
//   - ctx: A context to allow cancellation or timeout control.
//
// Returns:
//   - An error if the connection check fails; otherwise, nil.
func (r *InMemoryRepository) CheckConnection(_ context.Context) error {
	return nil
}
