package storage

import "github.com/gdyunin/metricol.git/internal/server/metrics"

type Repository interface {
	PushMetric(metrics.Metric) error
	Metrics() map[metrics.MetricType]map[string]string
}
