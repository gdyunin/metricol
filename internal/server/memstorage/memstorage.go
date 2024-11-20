package memstorage

import (
	"fmt"
	"github.com/gdyunin/metricol.git/internal/server/metrics"
)

type MetricType string

type MemStorage struct {
	storage map[MetricType][]metrics.BaseMetric
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		make(map[MetricType][]metrics.BaseMetric),
	}
}

func (m *MemStorage) SubmitMetric(metricType MetricType, metric metrics.BaseMetric) {
	_, err := m.searchMetric(metricType, metric.MetricName())
	if err != nil {
		m.CreateMetric(metricType, metric)
	}
}

func (m *MemStorage) searchMetric(metricType MetricType, metricName string) (*metrics.BaseMetric, error) {
	metricsIn, ok := m.storage[metricType]
	if !ok {
		return nil, fmt.Errorf("нет типа %q", metricType)
	}

	for _, metric := range metricsIn {
		if metric.MetricName() == metricName {
			return &metric, nil
		}
	}
	return nil, fmt.Errorf("нет метрики %q", metricName)
}

func (m *MemStorage) CreateMetric(metricType MetricType, metric metrics.BaseMetric) {
	m.storage[metricType] = append(m.storage[metricType], metric)
}
