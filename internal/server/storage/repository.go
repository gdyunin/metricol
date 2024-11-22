package storage

import "github.com/gdyunin/metricol.git/internal/server/metrics"

type Repository interface {
	PushMetric(metrics.Metric) error
}
