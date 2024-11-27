package storage

import (
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
