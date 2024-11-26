package storage

import (
	"github.com/gdyunin/metricol.git/internal/server/metrics"
)

const (
	ErrorUnknownMetricType = "error unknown metric type"
)

type Repository interface {
	PushMetric(metrics.Metric) error
	Metrics() map[metrics.MetricType]map[string]string
}
