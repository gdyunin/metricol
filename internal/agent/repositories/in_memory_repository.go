package repositories

import (
	"sync"

	"github.com/gdyunin/metricol.git/internal/agent/entities"
	"go.uber.org/zap"
)

// InMemoryRepository is a thread-safe in-memory storage for metrics.
type InMemoryRepository struct {
	mu      *sync.RWMutex      // Mutex for concurrent access control.
	metrics []*entities.Metric // Slice to hold metrics.
	logger  *zap.SugaredLogger
}

// NewInMemoryRepository creates and returns a new instance of InMemoryRepository.
func NewInMemoryRepository(logger *zap.SugaredLogger) *InMemoryRepository {
	return &InMemoryRepository{
		metrics: make([]*entities.Metric, 0), // Initialize the metrics slice.
		mu:      &sync.RWMutex{},             // Initialize the mutex.
		logger:  logger,
	}
}

// Add adds a metric to the repository. If a metric with the same properties already exists, it updates its value.
func (mr *InMemoryRepository) Store(metric *entities.Metric) {
	if metric == nil {
		return
	}

	mr.mu.Lock()
	defer mr.mu.Unlock()

	for i := range mr.metrics {
		if mr.metrics[i].Equal(metric) {
			mr.metrics[i].Value = metric.Value
			return
		}
	}
	mr.metrics = append(mr.metrics, metric)
}

// Metrics retrieves all metrics from the repository.
func (mr *InMemoryRepository) Metrics() []*entities.Metric {
	mr.mu.RLock()
	defer mr.mu.RUnlock()
	return mr.metrics
}
