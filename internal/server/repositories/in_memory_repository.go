package repositories

import (
	"fmt"
	"sync"

	"github.com/gdyunin/metricol.git/internal/common/patterns"
	"github.com/gdyunin/metricol.git/internal/server/entities"
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

// InMemoryRepository is an in-memory implementation of the MetricsRepository interface.
// It stores counter and gauge metrics in maps.
type InMemoryRepository struct {
	counters  map[string]int64   // Stores counter metrics.
	gauges    map[string]float64 // Stores gauge metrics.
	mu        *sync.RWMutex      // Provides thread-safe access to the metrics.
	observers map[patterns.Observer]struct{}
	logger    *zap.SugaredLogger
}

// NewInMemoryRepository creates and returns a new instance of InMemoryRepository.
// It initializes empty maps for counters and gauges.
func NewInMemoryRepository(logger *zap.SugaredLogger) *InMemoryRepository {
	return &InMemoryRepository{
		counters: make(map[string]int64),
		gauges:   make(map[string]float64),
		mu:       &sync.RWMutex{},
		logger:   logger,
	}
}

// Create adds a new metric to the repository based on its type.
// It returns an error if the metric type is unknown.
func (r *InMemoryRepository) Create(metric *entities.Metric) error {
	if err := r.Update(metric); err != nil {
		return fmt.Errorf("error create metric: %w", err)
	}
	return nil
}

// Read retrieves a metric from the repository based on the provided filter.
// It returns an error if the metric type is unknown or if the metric is not found.
func (r *InMemoryRepository) Read(filter *entities.Filter) (*entities.Metric, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	switch filter.Type {
	case entities.MetricTypeCounter:
		value, exists := r.counters[filter.Name]
		if !exists {
			return nil, fmt.Errorf("counter metric '%s' not found", filter.Name)
		}
		return &entities.Metric{
			Name:  filter.Name,
			Type:  filter.Type,
			Value: value,
		}, nil
	case entities.MetricTypeGauge:
		value, exists := r.gauges[filter.Name]
		if !exists {
			return nil, fmt.Errorf("gauge metric '%s' not found", filter.Name)
		}
		return &entities.Metric{
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
func (r *InMemoryRepository) Update(metric *entities.Metric) (err error) {
	r.mu.Lock()
	defer func() {
		r.mu.Unlock()
		if err == nil {
			r.NotifyObservers()
		}
	}()

	switch metric.Type {
	case entities.MetricTypeCounter:
		err = r.storeCounter(metric)
	case entities.MetricTypeGauge:
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
func (r *InMemoryRepository) IsExists(filter *entities.Filter) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	switch filter.Type {
	case entities.MetricTypeCounter:
		_, exists := r.counters[filter.Name]
		return exists, nil
	case entities.MetricTypeGauge:
		_, exists := r.gauges[filter.Name]
		return exists, nil
	default:
		return false, fmt.Errorf("unsupported metric type in existence check: %s", filter.Type)
	}
}

// All retrieves all metrics from the repository.
// It returns a slice of metrics and any error that may occur during the process.
func (r *InMemoryRepository) All() ([]*entities.Metric, error) {
	metrics := make([]*entities.Metric, 0, r.metricsCount())

	r.mu.RLock()
	defer r.mu.RUnlock()

	// Add counter metrics.
	for name, value := range r.counters {
		metrics = append(metrics, &entities.Metric{
			Name:  name,
			Type:  entities.MetricTypeCounter,
			Value: value,
		})
	}

	// Add gauge metrics.
	for name, value := range r.gauges {
		metrics = append(metrics, &entities.Metric{
			Name:  name,
			Type:  entities.MetricTypeGauge,
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

func (r *InMemoryRepository) RegisterObserver(observer patterns.Observer) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.observers[observer]; exists {
		return fmt.Errorf("observer %v is already registered", observer)
	}

	r.observers[observer] = struct{}{}
	return nil
}

func (r *InMemoryRepository) RemoveObserver(observer patterns.Observer) error {
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

func (r *InMemoryRepository) storeCounter(metric *entities.Metric) error {
	value, ok := metric.Value.(int64)
	if !ok {
		return fmt.Errorf("counter value must be int64 but got %T", metric.Value)
	}
	r.counters[metric.Name] = value
	return nil
}

func (r *InMemoryRepository) storeGauge(metric *entities.Metric) error {
	value, ok := metric.Value.(float64)
	if !ok {
		return fmt.Errorf("gauge value must be float64 but got %T", metric.Value)
	}
	r.gauges[metric.Name] = value
	return nil
}
