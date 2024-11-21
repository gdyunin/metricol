package memstorage

import (
	"errors"
	"github.com/gdyunin/metricol.git/internal/server/metrics"
	"strconv"
)

type MemStorage interface {
	PushMetric(metrics.Metric) error
	PullMetric()
}

type BaseMemStorage struct {
	storage map[metrics.MetricType]map[string]string
}

func NewBaseMemStorage() BaseMemStorage {
	return BaseMemStorage{
		make(map[metrics.MetricType]map[string]string),
	}
}

func (s *BaseMemStorage) PushMetric(metric metrics.Metric) error {
	switch metric.Type() {
	case metrics.MetricTypeGauge:
		s.pushGauge(metric)
		return nil
	case metrics.MetricTypeCounter:
		s.pushCounter(metric)
		return nil
	default:
		return errors.New(ErrorUnknownMetricType)
	}
}

func (s *BaseMemStorage) PullMetric() {
	//TODO implement me
	//panic("implement me")
}

func (s *BaseMemStorage) pushGauge(metric metrics.Metric) {
	_, ok := s.storage[metrics.MetricTypeGauge]
	if !ok {
		s.storage[metrics.MetricTypeGauge] = make(map[string]string)
	}
	s.storage[metrics.MetricTypeGauge][metric.Name()] = metric.Value()
}

func (s *BaseMemStorage) pushCounter(metric metrics.Metric) {
	m, ok := s.storage[metrics.MetricTypeCounter]
	if !ok {
		s.storage[metrics.MetricTypeCounter] = make(map[string]string)
		s.storage[metrics.MetricTypeCounter][metric.Name()] = metric.Value()
		return
	}

	v, _ := strconv.ParseInt(m[metric.Name()], 10, 64)
	nv, _ := strconv.ParseInt(metric.Value(), 10, 64)

	v += nv

	m[metric.Name()] = strconv.FormatInt(v, 10)
}
