package fetch

import (
	"github.com/gdyunin/metricol.git/internal/agent/metrics"
	"github.com/gdyunin/metricol.git/internal/agent/metrics/library"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNewStorage(t *testing.T) {
	tests := []struct {
		name string
		want *Storage
	}{
		{
			"create storage",
			&Storage{
				metrics: []metrics.Metric{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewStorage()
			require.Equal(t, tt.want, got)
		})
	}
}

func TestStorage_AddMetrics(t *testing.T) {
	testMetrics := []metrics.Metric{
		library.NewCounter("PollCount", func() int64 {
			return 1
		}),
	}
	type fields struct {
		metrics []metrics.Metric
	}
	type args struct {
		metric []metrics.Metric
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   fields
	}{
		{
			"add metrics",
			fields{
				metrics: []metrics.Metric{},
			},
			args{testMetrics},
			fields{testMetrics},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Storage{
				metrics: tt.fields.metrics,
			}
			s.AddMetrics(tt.args.metric...)
			require.Equal(t, tt.want.metrics, s.metrics)
		})
	}
}

func TestStorage_Metrics(t *testing.T) {
	testMetrics := []metrics.Metric{
		library.NewCounter("PollCount", func() int64 {
			return 1
		}),
	}
	type fields struct {
		metrics []metrics.Metric
	}
	tests := []struct {
		name   string
		fields fields
		want   []metrics.Metric
	}{
		{
			"get metrics",
			fields{
				testMetrics,
			},
			testMetrics,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Storage{
				metrics: tt.fields.metrics,
			}
			got := s.Metrics()
			require.Equal(t, tt.want, got)
		})
	}
}

func TestStorage_UpdateMetrics(t *testing.T) {
	testMetrics := []metrics.Metric{
		library.NewCounter("PollCount", func() int64 {
			return 1
		}),
	}
	type fields struct {
		metrics []metrics.Metric
	}
	tests := []struct {
		name   string
		fields fields
		want   fields
	}{
		{
			"update metrics",
			fields{
				testMetrics,
			},
			fields{
				testMetrics,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Storage{
				metrics: tt.fields.metrics,
			}
			s.UpdateMetrics()
			require.Equal(t, tt.want.metrics, s.metrics)
		})
	}
}
