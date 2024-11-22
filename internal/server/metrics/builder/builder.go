package builder

import (
	"errors"
	"github.com/gdyunin/metricol.git/internal/server/metrics"
	"github.com/gdyunin/metricol.git/internal/server/metrics/library"
)

func NewMetric(metricType metrics.MetricType) (metrics.Metric, error) {
	switch metricType {
	case metrics.MetricTypeGauge:
		return library.NewGauge(), nil
	case metrics.MetricTypeCounter:
		return library.NewCounter(), nil
	case metrics.MetricTypeOther:
		return library.NewOther(), nil
	default:
		return nil, errors.New(metrics.ErrorUnknownMetricType)
	}
}
