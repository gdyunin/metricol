package fetch

import (
	"github.com/gdyunin/metricol.git/internal/metrics"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestMetricsFetcher_AddMetrics(t *testing.T) {
	tests := []struct {
		name       string
		newMetrics []metrics.Metric
	}{
		{
			"simple add metrics",
			[]metrics.Metric{
				metrics.NewCounter("counter1", 0),
				metrics.NewGauge("gauge1", 0.0),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fetcher := NewMetricsFetcher()
			fetcher.AddMetrics(tt.newMetrics...)

			require.Len(t, fetcher.metrics, len(tt.newMetrics))
			require.Equal(t, "counter1", fetcher.metrics[0].Name())
			require.Equal(t, "gauge1", fetcher.metrics[1].Name())
		})
	}
}

func TestMetricsFetcher_Fetch(t *testing.T) {
	tests := []struct {
		name        string
		testMetrics []metrics.Metric
		wantMetrics []metrics.Metric
	}{
		{
			"with one metric",
			[]metrics.Metric{
				metrics.NewCounter("test_counter", 0).SetFetcherAndReturn(func() int64 {
					return 42
				}),
			},
			[]metrics.Metric{
				metrics.NewCounter("test_counter", 42).SetFetcherAndReturn(func() int64 {
					return 42
				}),
			},
		},
		{
			"with many metrics",
			[]metrics.Metric{
				metrics.NewCounter("test_counter", 0).SetFetcherAndReturn(func() int64 {
					return 42
				}),
				metrics.NewGauge("test_gauge", 0).SetFetcherAndReturn(func() float64 {
					return 3.14
				}),
			},
			[]metrics.Metric{
				metrics.NewCounter("test_counter", 42),
				metrics.NewGauge("test_gauge", 3.14),
			},
		},
		{
			"without fetcher",
			[]metrics.Metric{
				metrics.NewCounter("test_counter", 15),
			},
			[]metrics.Metric{
				metrics.NewCounter("test_counter", 15),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fetcher := NewMetricsFetcher()
			fetcher.AddMetrics(tt.testMetrics...)
			fetcher.Fetch()

			var testValues []string
			for _, tm := range fetcher.metrics {
				testValues = append(testValues, tm.StringValue())
			}

			var wantValues []string
			for _, wm := range tt.wantMetrics {
				wantValues = append(wantValues, wm.StringValue())
			}

			require.Equal(t, wantValues, testValues)
		})
	}
}

func TestNewMetricsFetcher(t *testing.T) {
	tests := []struct {
		name string
		want *MetricsFetcher
	}{
		{"simple MetricsFetcher init", &MetricsFetcher{metrics: []metrics.Metric{}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, NewMetricsFetcher())
		})
	}
}

func TestMetricsFetcher_Metrics(t *testing.T) {
	tests := []struct {
		name        string
		newMetrics  []metrics.Metric
		wantMetrics []metrics.Metric
	}{
		{
			"simple test metrics",
			[]metrics.Metric{
				metrics.NewCounter("counter1", 0),
				metrics.NewGauge("gauge1", 0.0),
			},
			[]metrics.Metric{
				metrics.NewCounter("counter1", 0),
				metrics.NewGauge("gauge1", 0.0),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fetcher := NewMetricsFetcher()
			fetcher.AddMetrics(tt.newMetrics...)

			testMetrics := fetcher.Metrics()

			require.Len(t, testMetrics, len(tt.newMetrics))

			for i, m := range tt.wantMetrics {
				require.Equal(t, m.Name(), testMetrics[i].Name())
			}
		})
	}
}
