package repositories

import (
	"sync"

	"github.com/gdyunin/metricol.git/internal/agent/entity"
)

// InMemoryRepository is a thread-safe in-memory storage for metrics.
type InMemoryRepository struct {
	mu      *sync.RWMutex    // Mutex for concurrent access control.
	metrics []*entity.Metric // Slice to hold metrics.
}

// NewInMemoryRepository creates and returns a new instance of InMemoryRepository.
func NewInMemoryRepository() *InMemoryRepository {
	return &InMemoryRepository{
		metrics: make([]*entity.Metric, 0), // Initialize the metrics slice.
		mu:      &sync.RWMutex{},           // Initialize the mutex.
	}
}

// Add adds a metric to the repository. If a metric with the same properties already exists, it updates its value.
func (mr *InMemoryRepository) Add(metric *entity.Metric) {
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
func (mr *InMemoryRepository) Metrics() []*entity.Metric {
	mr.mu.RLock()
	defer mr.mu.RUnlock()
	return mr.metrics
}
