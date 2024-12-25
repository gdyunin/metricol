package repositories

import (
	"fmt"
	"sync"

	"github.com/gdyunin/metricol.git/internal/common"
	"github.com/gdyunin/metricol.git/internal/server/entity"
)

// InMemoryRepository is an in-memory implementation of the MetricRepository interface.
// It stores counter and gauge metrics in maps.
type InMemoryRepository struct {
	counters  map[string]int64   // Stores counter metrics.
	gauges    map[string]float64 // Stores gauge metrics.
	mu        *sync.RWMutex      // Provides thread-safe access to the metrics.
	observers map[common.Observer]struct{}
}

// NewInMemoryRepository creates and returns a new instance of InMemoryRepository.
// It initializes empty maps for counters and gauges.
func NewInMemoryRepository() *InMemoryRepository {
	return &InMemoryRepository{
		counters: make(map[string]int64),
		gauges:   make(map[string]float64),
		mu:       &sync.RWMutex{},
	}
}

// Create adds a new metric to the repository based on its type.
// It returns an error if the metric type is unknown.
func (r *InMemoryRepository) Create(metric *entity.Metric) error {
	if err := r.Update(metric); err != nil {
		return fmt.Errorf("error create metric: %w", err)
	}
	return nil
}

// Read retrieves a metric from the repository based on the provided filter.
// It returns an error if the metric type is unknown or if the metric is not found.
func (r *InMemoryRepository) Read(filter *entity.Filter) (*entity.Metric, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	switch filter.Type {
	case entity.MetricTypeCounter:
		value, exists := r.counters[filter.Name]
		if !exists {
			return nil, fmt.Errorf("counter metric '%s' not found", filter.Name)
		}
		return &entity.Metric{
			Name:  filter.Name,
			Type:  filter.Type,
			Value: value,
		}, nil
	case entity.MetricTypeGauge:
		value, exists := r.gauges[filter.Name]
		if !exists {
			return nil, fmt.Errorf("gauge metric '%s' not found", filter.Name)
		}
		return &entity.Metric{
			Name:  filter.Name,
			Type:  filter.Type,
			Value: value,
		}, nil
	default:
		return nil, fmt.Errorf("unsupported metric type: %s", filter.Type)
	}
}

// Update modifies an existing metric in the repository based on its type.
// It returns an error if the metric type is unknown.
func (r *InMemoryRepository) Update(metric *entity.Metric) (err error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	switch metric.Type {
	case entity.MetricTypeCounter:
		err = r.storeCounter(metric)
	case entity.MetricTypeGauge:
		err = r.storeGauge(metric)
	default:
		err = fmt.Errorf("unsupported metric type: %s", metric.Type)
	}

	if err != nil {
		err = fmt.Errorf("error update metric: %w", err)
	}
	return
}

// IsExists checks if a metric with the specified filter exists in the repository.
// It returns true if the metric exists, otherwise false. It also returns an error if the filter type is unknown.
func (r *InMemoryRepository) IsExists(filter *entity.Filter) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	switch filter.Type {
	case entity.MetricTypeCounter:
		_, exists := r.counters[filter.Name]
		return exists, nil
	case entity.MetricTypeGauge:
		_, exists := r.gauges[filter.Name]
		return exists, nil
	default:
		return false, fmt.Errorf("unsupported metric type in existence check: %s", filter.Type)
	}
}

// All retrieves all metrics from the repository.
// It returns a slice of metrics and any error that may occur during the process.
func (r *InMemoryRepository) All() ([]*entity.Metric, error) {
	metrics := make([]*entity.Metric, 0, r.metricsCount())

	r.mu.RLock()
	defer r.mu.RUnlock()

	// Add counter metrics.
	for name, value := range r.counters {
		metrics = append(metrics, &entity.Metric{
			Name:  name,
			Type:  entity.MetricTypeCounter,
			Value: value,
		})
	}

	// Add gauge metrics.
	for name, value := range r.gauges {
		metrics = append(metrics, &entity.Metric{
			Name:  name,
			Type:  entity.MetricTypeGauge,
			Value: value,
		})
	}

	return metrics, nil
}

// metricsCount returns the total number of metrics (both counters and gauges) in the repository.
func (r *InMemoryRepository) metricsCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return len(r.counters) + len(r.gauges)
}

func (r *InMemoryRepository) RegisterObserver(observer common.Observer) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.observers[observer]; exists {
		return fmt.Errorf("observer %v is already registered", observer)
	}

	r.observers[observer] = struct{}{}
	return nil
}

func (r *InMemoryRepository) RemoveObserver(observer common.Observer) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.observers[observer]; exists {
		delete(r.observers, observer)
		return nil
	}

	return fmt.Errorf("observer %v is not registered", observer)
}

func (r *InMemoryRepository) NotifyObservers() {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for o := range r.observers {
		o.OnNotify()
	}
}

func (r *InMemoryRepository) storeCounter(metric *entity.Metric) error {
	value, ok := metric.Value.(int64)
	if !ok {
		return fmt.Errorf("counter value must be int64 but got %T", metric.Value)
	}
	r.counters[metric.Name] = value
	return nil
}

func (r *InMemoryRepository) storeGauge(metric *entity.Metric) error {
	value, ok := metric.Value.(float64)
	if !ok {
		return fmt.Errorf("gauge value must be float64 but got %T", metric.Value)
	}
	r.gauges[metric.Name] = value
	return nil
}
