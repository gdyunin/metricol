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
// It stores counter and gauge metrics in maps and provides thread-safe access.
type InMemoryRepository struct {
	counters  map[string]int64               // Stores counter metrics.
	gauges    map[string]float64             // Stores gauge metrics.
	mu        *sync.RWMutex                  // Mutex for thread-safe access to the metrics.
	observers map[patterns.Observer]struct{} // Observers notified on updates.
	logger    *zap.SugaredLogger             // Logger for repository activities.
}

// NewInMemoryRepository creates and returns a new instance of InMemoryRepository.
// It initializes empty maps for counters and gauges.
//
// Parameters:
//   - logger: Logger for repository activities.
//
// Returns:
//   - A pointer to a new InMemoryRepository instance.
func NewInMemoryRepository(logger *zap.SugaredLogger) *InMemoryRepository {
	return &InMemoryRepository{
		counters:  make(map[string]int64),
		gauges:    make(map[string]float64),
		mu:        &sync.RWMutex{},
		observers: make(map[patterns.Observer]struct{}),
		logger:    logger,
	}
}

// Create adds a new metric to the repository based on its type.
//
// Parameters:
//   - metric: The metric to be added.
//
// Returns:
//   - An error if the metric cannot be added or updated.
func (r *InMemoryRepository) Create(metric *entities.Metric) error {
	if err := r.Update(metric); err != nil {
		return fmt.Errorf("failed to create metric: %w", err)
	}
	return nil
}

// Read retrieves a metric from the repository based on the provided filter.
//
// Parameters:
//   - filter: The filter specifying the metric name and type.
//
// Returns:
//   - The requested metric.
//   - An error if the metric is not found or if the type is unsupported.
func (r *InMemoryRepository) Read(filter *entities.Filter) (*entities.Metric, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	switch filter.Type {
	case entities.MetricTypeCounter:
		value, exists := r.counters[filter.Name]
		if !exists {
			return nil, fmt.Errorf("counter metric '%s' not found", filter.Name)
		}
		return &entities.Metric{Name: filter.Name, Type: filter.Type, Value: value}, nil
	case entities.MetricTypeGauge:
		value, exists := r.gauges[filter.Name]
		if !exists {
			return nil, fmt.Errorf("gauge metric '%s' not found", filter.Name)
		}
		return &entities.Metric{Name: filter.Name, Type: filter.Type, Value: value}, nil
	default:
		return nil, fmt.Errorf("unsupported metric type: %s", filter.Type)
	}
}

// Update modifies an existing metric in the repository based on its type.
//
// Parameters:
//   - metric: The metric to be updated.
//
// Returns:
//   - An error if the metric type is unsupported or the update fails.
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
		err = fmt.Errorf("failed to update metric: %w", err)
	}

	return
}

// IsExists checks if a metric with the specified filter exists in the repository.
//
// Parameters:
//   - filter: The filter specifying the metric name and type.
//
// Returns:
//   - A boolean indicating whether the metric exists.
//   - An error if the filter type is unsupported.
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
//
// Returns:
//   - A slice of all metrics in the repository.
//   - An error if retrieval fails.
func (r *InMemoryRepository) All() ([]*entities.Metric, error) {
	metrics := make([]*entities.Metric, 0, r.metricsCount())

	r.mu.RLock()
	defer r.mu.RUnlock()

	for name, value := range r.counters {
		metrics = append(metrics, &entities.Metric{Name: name, Type: entities.MetricTypeCounter, Value: value})
	}

	for name, value := range r.gauges {
		metrics = append(metrics, &entities.Metric{Name: name, Type: entities.MetricTypeGauge, Value: value})
	}

	return metrics, nil
}

// metricsCount returns the total number of metrics in the repository.
func (r *InMemoryRepository) metricsCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return len(r.counters) + len(r.gauges)
}

// RegisterObserver registers an observer to be notified of metric updates.
//
// Parameters:
//   - observer: The observer to register.
//
// Returns:
//   - An error if the observer is already registered.
func (r *InMemoryRepository) RegisterObserver(observer patterns.Observer) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.observers[observer]; exists {
		return fmt.Errorf("observer %v is already registered", observer)
	}

	r.observers[observer] = struct{}{}
	return nil
}

// RemoveObserver removes a previously registered observer.
//
// Parameters:
//   - observer: The observer to remove.
//
// Returns:
//   - An error if the observer is not registered.
func (r *InMemoryRepository) RemoveObserver(observer patterns.Observer) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.observers[observer]; exists {
		delete(r.observers, observer)
		return nil
	}

	return fmt.Errorf("observer %v is not registered", observer)
}

// NotifyObservers notifies all registered observers about metric updates.
func (r *InMemoryRepository) NotifyObservers() {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for o := range r.observers {
		o.OnNotify()
	}
}

// storeCounter stores or updates a counter metric in the repository.
//
// Parameters:
//   - metric: The counter metric to store.
//
// Returns:
//   - An error if the metric value is not an int64.
func (r *InMemoryRepository) storeCounter(metric *entities.Metric) error {
	value, ok := metric.Value.(int64)
	if !ok {
		return fmt.Errorf("counter value must be int64 but got %T", metric.Value)
	}
	r.counters[metric.Name] = value
	return nil
}

// storeGauge stores or updates a gauge metric in the repository.
//
// Parameters:
//   - metric: The gauge metric to store.
//
// Returns:
//   - An error if the metric value is not a float64.
func (r *InMemoryRepository) storeGauge(metric *entities.Metric) error {
	value, ok := metric.Value.(float64)
	if !ok {
		return fmt.Errorf("gauge value must be float64 but got %T", metric.Value)
	}
	r.gauges[metric.Name] = value
	return nil
}
