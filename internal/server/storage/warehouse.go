package storage

import (
	"errors"
	"fmt"
	"github.com/gdyunin/metricol.git/internal/metrics"
	"strconv"
)

type Warehouse struct {
	counters map[string]int64
	gauges   map[string]float64
}

func NewWarehouse() *Warehouse {
	return &Warehouse{
		counters: make(map[string]int64),
		gauges:   make(map[string]float64),
	}
}

func (w *Warehouse) PushMetric(metric metrics.Metric) error {
	switch m := metric.(type) {
	case *metrics.Counter:
		w.counters[m.Name()] += m.Value()
	case *metrics.Gauge:
		w.gauges[m.Name()] = m.Value()
	default:
		return errors.New(ErrorUnknownMetricType)
	}
	return nil
}

func (w *Warehouse) GetMetric(name, metricType string) (string, error) {
	var value string
	switch metricType {
	case metrics.MetricTypeCounter:
		v, ok := w.counters[name]
		if !ok {
			return "", errors.New(ErrorUnknownMetricName)
		}
		value = strconv.FormatInt(v, 10)
	case metrics.MetricTypeGauge:
		v, ok := w.gauges[name]
		if !ok {
			return "", errors.New(ErrorUnknownMetricName)
		}
		value = fmt.Sprintf("%g", v)
	default:
		return "", errors.New(ErrorUnknownMetricType)
	}

	return value, nil
}

func (w *Warehouse) Metrics() map[string]map[string]string {
	allMetricsMap := make(map[string]map[string]string)

	allMetricsMap[metrics.MetricTypeCounter] = make(map[string]string)
	for name, value := range w.counters {
		allMetricsMap[metrics.MetricTypeCounter][name] = strconv.FormatInt(value, 10)
	}

	allMetricsMap[metrics.MetricTypeGauge] = make(map[string]string)
	for name, value := range w.gauges {
		allMetricsMap[metrics.MetricTypeGauge][name] = fmt.Sprintf("%g", value)
	}

	return allMetricsMap
}
