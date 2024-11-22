package builder

import (
	"errors"
	"github.com/gdyunin/metricol.git/internal/server/metrics"
	"github.com/gdyunin/metricol.git/internal/server/metrics/library"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNewMetric(t *testing.T) {
	tests := []struct {
		name       string
		metricType metrics.MetricType
		want       metrics.Metric
		wantErr    error
	}{
		{
			"init new gauge",
			metrics.MetricTypeGauge,
			library.NewGauge(),
			nil,
		},
		{
			"init new counter",
			metrics.MetricTypeCounter,
			library.NewCounter(),
			nil,
		},
		{
			"init new other",
			metrics.MetricTypeOther,
			library.NewOther(),
			nil,
		},
		{
			"try init unknown metric",
			metrics.MetricType("test unknown"),
			nil,
			errors.New(metrics.ErrorUnknownMetricType),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewMetric(tt.metricType)

			if err != nil {
				require.EqualError(t, err, tt.wantErr.Error())
				return
			}

			require.Equal(t, got, tt.want)
		})
	}
}
