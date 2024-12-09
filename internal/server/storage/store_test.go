package storage

import (
	"fmt"
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
			"simple create store",
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
		store      *Store
		nameMetric string
		metricType string
		want       string
		wantErr    error
	}{
		{
			name: "get existing counter",
			store: func() *Store {
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
			store: func() *Store {
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
			store:      NewStore(),
			nameMetric: "unknown_metric",
			metricType: metrics.MetricTypeCounter,
			want:       "",
			wantErr:    fmt.Errorf("error get metric %s %s: unknown metric name", "unknown_metric", metrics.MetricTypeCounter),
		},
		{
			name: "get metric with unknown type",
			store: func() *Store {
				w := NewStore()
				_ = w.PushMetric(metrics.NewCounter("test_counter", 10))
				return w
			}(),
			nameMetric: "test_counter",
			metricType: "unknown_type",
			want:       "",
			wantErr:    fmt.Errorf("error get metric %s %s: %w", "test_counter", "unknown_type", ErrUnknownMetricType),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, err := tt.store.GetMetric(tt.nameMetric, tt.metricType)
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
			name:      "empty store",
			warehouse: NewStore(),
			want: map[string]map[string]string{
				metrics.MetricTypeCounter: {},
				metrics.MetricTypeGauge:   {},
			},
		},
		{
			name: "only counters in store",
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
			name: "only gauges in store",
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
			name: "mixed metrics in store",
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
		store      *Store
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
			fmt.Errorf("error push metric %v: %w", func() metrics.Metric {
				m, _ := metrics.NewFromStrings("test_counter", "84", "some_unknown_metric_type")
				return m
			}(), ErrUnknownMetricType),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := tt.store

			err := s.PushMetric(tt.pushMetric)
			if err != nil {
				require.EqualError(t, err, tt.wantErr.Error())
				return
			}

			var (
				metricName string
				metricType string
			)

			switch m := tt.pushMetric.(type) {
			case *metrics.Counter:
				metricName = m.Name
				metricType = metrics.MetricTypeCounter
			case *metrics.Gauge:
				metricName = m.Name
				metricType = metrics.MetricTypeGauge
			default:
				require.Fail(t, "Metric isn`t counter or gauge!")
			}

			newValue, err := s.GetMetric(metricName, metricType)
			require.NoError(t, err)
			require.Equal(t, tt.wantMetric.StringValue(), newValue)
		})
	}
}
