package storage

import (
	"errors"
	"github.com/gdyunin/metricol.git/internal/server/metrics"
	"strconv"
)

type Warehouse struct {
	metrics map[metrics.MetricType]map[string]string
}

func NewWarehouse() *Warehouse {
	w := &Warehouse{}
	w.init()
	return w
}

func (w *Warehouse) PushMetric(metric metrics.Metric) error {
	mName := metric.Name()
	mValue := metric.Value()
	mType := metric.Type()

	switch mType {
	case metrics.MetricTypeGauge:
		return w.pushGauge(mName, mValue)
	case metrics.MetricTypeCounter:
		return w.pushCounter(mName, mValue)
	default:
		return errors.New(ErrorUnknownMetricType)
	}
}

func (w *Warehouse) init() {
	initMetrics := func() map[metrics.MetricType]map[string]string { return make(map[metrics.MetricType]map[string]string) }
	initMetricType := func() map[string]string { return make(map[string]string) }

	w.metrics = initMetrics()
	w.metrics[metrics.MetricTypeGauge] = initMetricType()
	w.metrics[metrics.MetricTypeCounter] = initMetricType()
}

func (w *Warehouse) pushGauge(name string, value string) error {
	metricType := metrics.MetricTypeGauge
	w.metrics[metricType][name] = value
	return nil
}

func (w *Warehouse) pushCounter(name string, value string) error {
	metricType := metrics.MetricTypeCounter

	curValue, ok := w.metrics[metricType][name]
	if !ok {
		w.metrics[metricType][name] = value
		return nil
	}

	cv, err := strconv.ParseInt(curValue, 10, 64)
	if err != nil {
		return err
	}

	v, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return err
	}

	newValue := strconv.FormatInt(cv+v, 10)

	w.metrics[metricType][name] = newValue
	return nil
}
