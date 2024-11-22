package fetch

import "github.com/gdyunin/metricol.git/internal/agent/metrics"

type Storage struct {
	metrics []metrics.Metric
}

func NewStorage() *Storage {
	return &Storage{
		metrics: []metrics.Metric{},
	}
}

func (s *Storage) Metrics() []metrics.Metric {
	return s.metrics
}

func (s *Storage) AddMetrics(metric ...metrics.Metric) {
	s.metrics = append(s.metrics, metric...)
}

func (s *Storage) UpdateMetrics() {
	for _, m := range s.metrics {
		m.UpdateValue()
	}
}
