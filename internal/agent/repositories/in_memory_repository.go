package repositories

import (
	"sync"

	"github.com/gdyunin/metricol.git/internal/agent/entities"
	"go.uber.org/zap"
)

// InMemoryRepositoryFactory implements RepositoryAbstractFactory for InMemoryRepository.
type InMemoryRepositoryFactory struct {
	logger *zap.SugaredLogger // Logger for recording factory-related activities.
}

// NewInMemoryRepositoryFactory creates a new instance of InMemoryRepositoryFactory.
//
// Parameters:
//   - logger: Logger for recording factory-related activities.
//
// Returns:
//   - A pointer to an initialized InMemoryRepositoryFactory instance.
func NewInMemoryRepositoryFactory(logger *zap.SugaredLogger) *InMemoryRepositoryFactory {
	return &InMemoryRepositoryFactory{
		logger: logger,
	}
}

// CreateMetricsRepository creates and returns a new InMemoryRepository instance.
//
// Returns:
//   - An implementation of the MetricsRepository interface using in-memory storage.
func (f *InMemoryRepositoryFactory) CreateMetricsRepository() entities.MetricsRepository {
	f.logger.Info("Creating a new in-memory metrics repository.")
	return NewInMemoryRepository(f.logger)
}

// InMemoryRepository is a thread-safe in-memory storage for metrics.
type InMemoryRepository struct {
	logger  *zap.SugaredLogger // Logger for recording repository activities and errors.
	mu      *sync.RWMutex      // Mutex for concurrent access control.
	metrics []*entities.Metric // Slice to hold metrics.
}

// NewInMemoryRepository creates and returns a new instance of InMemoryRepository.
//
// Parameters:
//   - logger: Logger for recording repository-related activities.
//
// Returns:
//   - A pointer to an initialized InMemoryRepository instance.
func NewInMemoryRepository(logger *zap.SugaredLogger) *InMemoryRepository {
	return &InMemoryRepository{
		metrics: make([]*entities.Metric, 0), // Initialize the metrics slice.
		mu:      &sync.RWMutex{},             // Initialize the mutex.
		logger:  logger,
	}
}

// Store adds a metric to the repository.
//
// If a metric with the same name and type already exists in the repository,
// this method updates its value instead of adding a new entry.
//
// Parameters:
//   - metric: A pointer to the Metric instance to store.
func (mr *InMemoryRepository) Store(metric *entities.Metric) {
	if metric == nil {
		mr.logger.Warn("Attempted to store a nil metric. Operation skipped.")
		return
	}

	mr.mu.Lock()
	defer mr.mu.Unlock()

	for i := range mr.metrics {
		if mr.metrics[i].Equal(metric) {
			mr.logger.Infof("Updating existing metric: %s", metric.Name)
			mr.metrics[i].Value = metric.Value
			return
		}
	}

	mr.logger.Infof("Storing new metric: %s", metric.Name)
	mr.metrics = append(mr.metrics, metric)
}

// Metrics retrieves all metrics from the repository.
//
// Returns:
//   - A slice of pointers to Metric instances currently stored in the repository.
func (mr *InMemoryRepository) Metrics() []*entities.Metric {
	mr.mu.RLock()
	defer mr.mu.RUnlock()

	mr.logger.Info("Retrieving all metrics from the repository.")
	return mr.metrics
}
