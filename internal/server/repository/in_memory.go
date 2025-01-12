package repository

import (
	"NewNewMetricol/internal/server/internal/entity"
	"fmt"
	"sync"

	"go.uber.org/zap"
)

type InMemoryRepository struct {
	counters map[string]int64
	gauges   map[string]float64
	mu       *sync.RWMutex
	logger   *zap.SugaredLogger
}

func NewInMemoryRepository(logger *zap.SugaredLogger) *InMemoryRepository {
	return &InMemoryRepository{
		counters: make(map[string]int64),
		gauges:   make(map[string]float64),
		mu:       &sync.RWMutex{},
		logger:   logger,
	}
}

func (r *InMemoryRepository) Update(metric *entity.Metric) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	updateFunctions := map[string]func(string, any) error{
		entity.MetricTypeCounter: r.updateCounter,
		entity.MetricTypeGauge:   r.updateGauge,
	}

	updateFunc, ok := updateFunctions[metric.Type]
	if !ok {
		return fmt.Errorf("unknown metric type: %s", metric.Type)
	}

	if err := updateFunc(metric.Name, metric.Value); err != nil {
		return fmt.Errorf("error updating metric (%s): %w", metric.Type, err)
	}

	return nil
}

func (r *InMemoryRepository) updateCounter(name string, value any) error {
	convertedValue, ok := value.(int64)
	if !ok {
		return fmt.Errorf("counter value must be int64 but got %T", value)
	}

	r.counters[name] = convertedValue
	return nil
}

func (r *InMemoryRepository) updateGauge(name string, value any) error {
	convertedValue, ok := value.(float64)
	if !ok {
		return fmt.Errorf("gauge value must be float64 but got %T", value)
	}

	r.gauges[name] = convertedValue
	return nil
}

func (r *InMemoryRepository) Find(metricType string, name string) (found *entity.Metric, err error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	findFunctions := map[string]func(string) (*entity.Metric, error){
		entity.MetricTypeCounter: r.findCounter,
		entity.MetricTypeGauge:   r.findGauge,
	}

	findFunc, ok := findFunctions[metricType]
	if !ok {
		return nil, fmt.Errorf("unknown metric type: %s", metricType)
	}

	found, err = findFunc(name)
	if err != nil {
		return nil, fmt.Errorf("error find metric (%s): %w", name, err)
	}

	return
}

func (r *InMemoryRepository) findCounter(name string) (*entity.Metric, error) {
	value, ok := r.counters[name]
	if !ok {
		return nil, fmt.Errorf("%w: counter %s not exists in repository", ErrNotFound, name)
	}

	return &entity.Metric{
		Value: value,
		Name:  name,
		Type:  entity.MetricTypeCounter,
	}, nil
}

func (r *InMemoryRepository) findGauge(name string) (*entity.Metric, error) {
	value, ok := r.gauges[name]
	if !ok {
		return nil, fmt.Errorf("%w: counter %s not exists in repository", ErrNotFound, name)
	}

	return &entity.Metric{
		Value: value,
		Name:  name,
		Type:  entity.MetricTypeGauge,
	}, nil
}

func (r *InMemoryRepository) All() (*entity.Metrics, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	metrics := append(r.allCounters(), r.allGauges()...)

	return &metrics, nil
}

func (r *InMemoryRepository) allCounters() entity.Metrics {
	metrics := make(entity.Metrics, 0, len(r.counters))
	for name, value := range r.counters {
		metrics = append(metrics, &entity.Metric{
			Value: value,
			Name:  name,
			Type:  entity.MetricTypeCounter,
		})
	}
	return metrics
}

func (r *InMemoryRepository) allGauges() entity.Metrics {
	metrics := make(entity.Metrics, 0, len(r.gauges))
	for name, value := range r.gauges {
		metrics = append(metrics, &entity.Metric{
			Value: value,
			Name:  name,
			Type:  entity.MetricTypeGauge,
		})
	}
	return metrics
}
