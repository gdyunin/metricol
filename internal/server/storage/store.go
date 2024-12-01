package storage

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/gdyunin/metricol.git/internal/metrics"
)

const (
	ErrorUnknownMetricType = "error unknown metric type"
	ErrorUnknownMetricName = "error unknown metric name"
)

type Repository interface {
	PushMetric(metrics.Metric) error
	GetMetric(string, string) (string, error)
	Metrics() map[string]map[string]string
}

type Store struct {
	counters map[string]int64
	gauges   map[string]float64
}

func NewStore() *Store {
	return &Store{
		counters: make(map[string]int64),
		gauges:   make(map[string]float64),
	}
}

func (s *Store) PushMetric(metric metrics.Metric) error {
	switch m := metric.(type) {
	case *metrics.Counter:
		s.counters[m.Name] += m.Value
	case *metrics.Gauge:
		s.gauges[m.Name] = m.Value
	default:
		return errors.New(ErrorUnknownMetricType)
	}
	return nil
}

func (s *Store) GetMetric(name, metricType string) (string, error) {
	var value string
	switch metricType {
	case metrics.MetricTypeCounter:
		v, ok := s.counters[name]
		if !ok {
			return "", errors.New(ErrorUnknownMetricName)
		}
		value = strconv.FormatInt(v, 10)
	case metrics.MetricTypeGauge:
		v, ok := s.gauges[name]
		if !ok {
			return "", errors.New(ErrorUnknownMetricName)
		}
		value = fmt.Sprintf("%g", v)
	default:
		return "", errors.New(ErrorUnknownMetricType)
	}

	return value, nil
}

func (s *Store) Metrics() map[string]map[string]string {
	allMetricsMap := make(map[string]map[string]string)

	allMetricsMap[metrics.MetricTypeCounter] = make(map[string]string)
	for name, value := range s.counters {
		allMetricsMap[metrics.MetricTypeCounter][name] = strconv.FormatInt(value, 10)
	}

	allMetricsMap[metrics.MetricTypeGauge] = make(map[string]string)
	for name, value := range s.gauges {
		allMetricsMap[metrics.MetricTypeGauge][name] = fmt.Sprintf("%g", value)
	}

	return allMetricsMap
}
