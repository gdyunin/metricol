package repository

import (
	"NewNewMetricol/internal/server/internal/entity"
	"fmt"
	"sync"

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
func (r *InMemoryRepository) Update(metric *entity.Metric) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.storage[metric.Type] == nil {
		r.storage[metric.Type] = make(map[string]any)
	}

	r.storage[metric.Type][metric.Name] = metric.Value
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
func (r *InMemoryRepository) IsExist(metricType string, name string) (bool, error) {
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
func (r *InMemoryRepository) Find(metricType string, name string) (*entity.Metric, error) {
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
func (r *InMemoryRepository) All() (*entity.Metrics, error) {
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
