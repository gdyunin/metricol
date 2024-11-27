package metrics

import "errors"

func NewCounter(name string, value int64) *Counter {
	return &Counter{
		name:  name,
		value: value,
	}
}

func NewGauge(name string, value float64) *Gauge {
	return &Gauge{
		name:  name,
		value: value,
	}
}

func NewFromStrings(name, value, metricType string) (Metric, error) {
	var createMetric func(string, string) (Metric, error)

	switch metricType {
	case MetricTypeGauge:
		createMetric = newGaugeFromStrings
	case MetricTypeCounter:
		createMetric = newCounterFromStrings
	default:
		return nil, errors.New(ErrorUnknownMetricType)
	}

	m, err := createMetric(name, value)
	if err != nil {
		return nil, err
	}
	return m, nil
}
