package storage

import (
	"errors"
	"testing"

	"github.com/gdyunin/metricol.git/internal/metrics"
	"github.com/stretchr/testify/require"
)

func TestNewWarehouse(t *testing.T) {
	tests := []struct {
		name string
		want *Store
	}{
		{
			"simple create warehouse",
			&Store{
				counters: make(map[string]int64),
				gauges:   make(map[string]float64),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, NewStore())
		})
	}
}

func TestWarehouse_GetMetric(t *testing.T) {
	tests := []struct {
		name       string
		warehouse  *Store
		nameMetric string
		metricType string
		want       string
		wantErr    error
	}{
		{
			name: "get existing counter",
			warehouse: func() *Store {
				w := NewStore()
				_ = w.PushMetric(metrics.NewCounter("test_counter", 10))
				return w
			}(),
			nameMetric: "test_counter",
			metricType: metrics.MetricTypeCounter,
			want:       "10",
			wantErr:    nil,
		},
		{
			name: "get existing gauge",
			warehouse: func() *Store {
				w := NewStore()
				_ = w.PushMetric(metrics.NewGauge("test_gauge", 3.14))
				return w
			}(),
			nameMetric: "test_gauge",
			metricType: metrics.MetricTypeGauge,
			want:       "3.14",
			wantErr:    nil,
		},
		{
			name:       "get unknown metric name",
			warehouse:  NewStore(),
			nameMetric: "unknown_metric",
			metricType: metrics.MetricTypeCounter,
			want:       "",
			wantErr:    errors.New(ErrorUnknownMetricName),
		},
		{
			name: "get metric with unknown type",
			warehouse: func() *Store {
				w := NewStore()
				_ = w.PushMetric(metrics.NewCounter("test_counter", 10))
				return w
			}(),
			nameMetric: "test_counter",
			metricType: "unknown_type",
			want:       "",
			wantErr:    errors.New(ErrorUnknownMetricType),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, err := tt.warehouse.GetMetric(tt.nameMetric, tt.metricType)
			if err != nil {
				require.EqualError(t, err, tt.wantErr.Error())
			}

			require.Equal(t, tt.want, m)
		})
	}
}

func TestWarehouse_Metrics(t *testing.T) {
	tests := []struct {
		name      string
		warehouse *Store
		want      map[string]map[string]string
	}{
		{
			name:      "empty warehouse",
			warehouse: NewStore(),
			want: map[string]map[string]string{
				metrics.MetricTypeCounter: {},
				metrics.MetricTypeGauge:   {},
			},
		},
		{
			name: "only counters in warehouse",
			warehouse: func() *Store {
				w := NewStore()
				_ = w.PushMetric(metrics.NewCounter("counter1", 10))
				_ = w.PushMetric(metrics.NewCounter("counter2", 20))
				return w
			}(),
			want: map[string]map[string]string{
				metrics.MetricTypeCounter: {
					"counter1": "10",
					"counter2": "20",
				},
				metrics.MetricTypeGauge: {},
			},
		},
		{
			name: "only gauges in warehouse",
			warehouse: func() *Store {
				w := NewStore()
				_ = w.PushMetric(metrics.NewGauge("gauge1", 3.14))
				_ = w.PushMetric(metrics.NewGauge("gauge2", 42.42))
				return w
			}(),
			want: map[string]map[string]string{
				metrics.MetricTypeCounter: {},
				metrics.MetricTypeGauge: {
					"gauge1": "3.14",
					"gauge2": "42.42",
				},
			},
		},
		{
			name: "mixed metrics in warehouse",
			warehouse: func() *Store {
				w := NewStore()
				_ = w.PushMetric(metrics.NewCounter("counter1", 10))
				_ = w.PushMetric(metrics.NewGauge("gauge1", 3.14))
				_ = w.PushMetric(metrics.NewCounter("counter2", 20))
				return w
			}(),
			want: map[string]map[string]string{
				metrics.MetricTypeCounter: {
					"counter1": "10",
					"counter2": "20",
				},
				metrics.MetricTypeGauge: {
					"gauge1": "3.14",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.warehouse.Metrics()
			require.Equal(t, tt.want, got)
		})
	}
}

func TestWarehouse_PushMetric(t *testing.T) {
	tests := []struct {
		name       string
		warehouse  *Store
		pushMetric metrics.Metric
		wantMetric metrics.Metric
		wantErr    error
	}{
		{
			"push new gauge",
			NewStore(),
			func() metrics.Metric {
				m, _ := metrics.NewFromStrings("test_gauge0", "42.0", metrics.MetricTypeGauge)
				return m
			}(),
			func() metrics.Metric {
				m, _ := metrics.NewFromStrings("test_gauge0", "42.0", metrics.MetricTypeGauge)
				return m
			}(),
			nil,
		},
		{
			"push repeat gauge",
			func() *Store {
				w := NewStore()
				m, _ := metrics.NewFromStrings("test", "42.0", metrics.MetricTypeGauge)
				_ = w.PushMetric(m)
				return w
			}(),
			func() metrics.Metric {
				m, _ := metrics.NewFromStrings("test", "5.8", metrics.MetricTypeGauge)
				return m
			}(),
			func() metrics.Metric {
				m, _ := metrics.NewFromStrings("test", "5.8", metrics.MetricTypeGauge)
				return m
			}(),
			nil,
		},
		{
			"push new counter",
			NewStore(),
			func() metrics.Metric {
				m, _ := metrics.NewFromStrings("test_counter", "42", metrics.MetricTypeCounter)
				return m
			}(),
			func() metrics.Metric {
				m, _ := metrics.NewFromStrings("test_counter", "42", metrics.MetricTypeCounter)
				return m
			}(),
			nil,
		},
		{
			"push repeat counter",
			func() *Store {
				w := NewStore()
				m, _ := metrics.NewFromStrings("test_counter", "42", metrics.MetricTypeCounter)
				_ = w.PushMetric(m)
				return w
			}(),
			func() metrics.Metric {
				m, _ := metrics.NewFromStrings("test_counter", "42", metrics.MetricTypeCounter)
				return m
			}(),
			func() metrics.Metric {
				m, _ := metrics.NewFromStrings("test_counter", "84", metrics.MetricTypeCounter)
				return m
			}(),
			nil,
		},
		{
			"push unknown metric",
			NewStore(),
			func() metrics.Metric {
				m, _ := metrics.NewFromStrings("test_counter", "84", "some_unknown_metric_type")
				return m
			}(),
			nil,
			errors.New(ErrorUnknownMetricType),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := tt.warehouse

			err := w.PushMetric(tt.pushMetric)
			if err != nil {
				require.EqualError(t, err, tt.wantErr.Error())
				return
			}

			newValue, err := w.GetMetric(tt.pushMetric.Name(), tt.pushMetric.Type())
			require.NoError(t, err)
			require.Equal(t, tt.wantMetric.StringValue(), newValue)
		})
	}
}
