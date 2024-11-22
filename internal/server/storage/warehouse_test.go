package storage

import (
	"errors"
	"github.com/gdyunin/metricol.git/internal/server/metrics"
	"github.com/gdyunin/metricol.git/internal/server/metrics/builder"
	"github.com/stretchr/testify/require"
	"strconv"
	"testing"
)

func TestNewWarehouse(t *testing.T) {
	tests := []struct {
		name string
		want *Warehouse
	}{
		{
			"build new warehouse",
			&Warehouse{
				metrics: func() map[metrics.MetricType]map[string]string {
					m := make(map[metrics.MetricType]map[string]string)
					m[metrics.MetricTypeGauge] = make(map[string]string)
					m[metrics.MetricTypeCounter] = make(map[string]string)
					return m
				}(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := NewWarehouse()
			require.Equal(t, tt.want, w)
		})
	}
}

func TestWarehouse_PushMetric(t *testing.T) {
	tests := []struct {
		name       string
		warehouse  *Warehouse
		pushMetric metrics.Metric
		wantMetric metrics.Metric
		wantErr    error
	}{
		{
			"push new gauge",
			func() *Warehouse {
				return NewWarehouse()
			}(),
			func() metrics.Metric {
				m, _ := builder.NewMetric(metrics.MetricTypeGauge)
				_ = m.SetName("test")
				_ = m.SetValue("5.3")
				return m
			}(),
			func() metrics.Metric {
				m, _ := builder.NewMetric(metrics.MetricTypeGauge)
				_ = m.SetName("test")
				_ = m.SetValue("5.3")
				return m
			}(),
			nil,
		},
		{
			"push repeat gauge",
			func() *Warehouse {
				w := NewWarehouse()
				_ = w.PushMetric(func() metrics.Metric {
					m, _ := builder.NewMetric(metrics.MetricTypeGauge)
					_ = m.SetName("test")
					_ = m.SetValue("5.3")
					return m
				}())
				return w
			}(),
			func() metrics.Metric {
				m, _ := builder.NewMetric(metrics.MetricTypeGauge)
				_ = m.SetName("test")
				_ = m.SetValue("5.8")
				return m
			}(),
			func() metrics.Metric {
				m, _ := builder.NewMetric(metrics.MetricTypeGauge)
				_ = m.SetName("test")
				_ = m.SetValue("5.8")
				return m
			}(),
			nil,
		},
		{
			"push new counter",
			func() *Warehouse {
				return NewWarehouse()
			}(),
			func() metrics.Metric {
				m, _ := builder.NewMetric(metrics.MetricTypeCounter)
				_ = m.SetName("test")
				_ = m.SetValue("5")
				return m
			}(),
			func() metrics.Metric {
				m, _ := builder.NewMetric(metrics.MetricTypeCounter)
				_ = m.SetName("test")
				_ = m.SetValue("5")
				return m
			}(),
			nil,
		},
		{
			"push repeat counter",
			func() *Warehouse {
				w := NewWarehouse()
				_ = w.PushMetric(func() metrics.Metric {
					m, _ := builder.NewMetric(metrics.MetricTypeCounter)
					_ = m.SetName("test")
					_ = m.SetValue("5")
					return m
				}())
				return w
			}(),
			func() metrics.Metric {
				m, _ := builder.NewMetric(metrics.MetricTypeCounter)
				_ = m.SetName("test")
				_ = m.SetValue("8")
				return m
			}(),
			func() metrics.Metric {
				m, _ := builder.NewMetric(metrics.MetricTypeCounter)
				_ = m.SetName("test")
				_ = m.SetValue("13")
				return m
			}(),
			nil,
		},
		{
			"push unknown metric",
			func() *Warehouse {
				return NewWarehouse()
			}(),
			func() metrics.Metric {
				m, _ := builder.NewMetric(metrics.MetricTypeOther)
				_ = m.SetName("test")
				_ = m.SetValue("other")
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
			require.EqualError(t, err, tt.wantErr.Error())

			if err == nil {
				require.Equal(t, tt.wantMetric.Value(), w.metrics[tt.pushMetric.Type()][tt.pushMetric.Name()])
			}
		})
	}
}

func TestWarehouse_init(t *testing.T) {
	tests := []struct {
		name string
		want *Warehouse
	}{
		{
			"init new warehouse",
			&Warehouse{
				metrics: func() map[metrics.MetricType]map[string]string {
					m := make(map[metrics.MetricType]map[string]string)
					m[metrics.MetricTypeGauge] = make(map[string]string)
					m[metrics.MetricTypeCounter] = make(map[string]string)
					return m
				}(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &Warehouse{}
			w.init()
			require.Equal(t, tt.want, w)
		})
	}
}

func TestWarehouse_pushCounter(t *testing.T) {
	type args struct {
		name  string
		value string
	}
	tests := []struct {
		name       string
		warehouse  *Warehouse
		args       args
		wantMetric metrics.Metric
		wantErr    error
	}{
		{
			"push new valid counter",
			func() *Warehouse {
				return NewWarehouse()
			}(),
			args{"test", "5"},
			func() metrics.Metric {
				m, _ := builder.NewMetric(metrics.MetricTypeCounter)
				_ = m.SetName("test")
				_ = m.SetValue("5")
				return m
			}(),
			nil,
		},
		{
			"push repeat valid counter",
			func() *Warehouse {
				w := NewWarehouse()
				_ = w.PushMetric(func() metrics.Metric {
					m, _ := builder.NewMetric(metrics.MetricTypeCounter)
					_ = m.SetName("test")
					_ = m.SetValue("5")
					return m
				}())
				return w
			}(),
			args{"test", "8"},
			func() metrics.Metric {
				m, _ := builder.NewMetric(metrics.MetricTypeCounter)
				_ = m.SetName("test")
				_ = m.SetValue("13")
				return m
			}(),
			nil,
		},
		{
			"push new invalid counter",
			func() *Warehouse {
				return NewWarehouse()
			}(),
			args{"test", "invalidInt"},
			nil,
			&strconv.NumError{},
		},
		{
			"push repeat invalid counter",
			func() *Warehouse {
				w := NewWarehouse()
				_ = w.PushMetric(func() metrics.Metric {
					m, _ := builder.NewMetric(metrics.MetricTypeCounter)
					_ = m.SetName("test")
					_ = m.SetValue("5")
					return m
				}())
				return w
			}(),
			args{"test", "invalidInt"},
			nil,
			&strconv.NumError{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			w := tt.warehouse

			err := w.pushCounter(tt.args.name, tt.args.value)
			if err != nil {
				require.EqualError(t, err, tt.wantErr.Error())
				return
			}

			require.Equal(t, tt.wantMetric.Value(), w.metrics[metrics.MetricTypeCounter][tt.args.name])

		})
	}
}

func TestWarehouse_pushGauge(t *testing.T) {
	type args struct {
		name  string
		value string
	}
	tests := []struct {
		name       string
		warehouse  *Warehouse
		args       args
		wantMetric metrics.Metric
		wantErr    error
	}{
		{
			"push new valid gauge",
			func() *Warehouse {
				return NewWarehouse()
			}(),
			args{"test", "5.3"},
			func() metrics.Metric {
				m, _ := builder.NewMetric(metrics.MetricTypeGauge)
				_ = m.SetName("test")
				_ = m.SetValue("5.3")
				return m
			}(),
			nil,
		},
		{
			"push repeat valid gauge",
			func() *Warehouse {
				w := NewWarehouse()
				_ = w.PushMetric(func() metrics.Metric {
					m, _ := builder.NewMetric(metrics.MetricTypeGauge)
					_ = m.SetName("test")
					_ = m.SetValue("5.3")
					return m
				}())
				return w
			}(),
			args{"test", "5.8"},
			func() metrics.Metric {
				m, _ := builder.NewMetric(metrics.MetricTypeGauge)
				_ = m.SetName("test")
				_ = m.SetValue("5.8")
				return m
			}(),
			nil,
		},
		{
			"push new invalid gauge",
			func() *Warehouse {
				return NewWarehouse()
			}(),
			args{"test", "invalidFloat"},
			nil,
			&strconv.NumError{},
		},
		{
			"push repeat invalid gauge",
			func() *Warehouse {
				w := NewWarehouse()
				_ = w.PushMetric(func() metrics.Metric {
					m, _ := builder.NewMetric(metrics.MetricTypeGauge)
					_ = m.SetName("test")
					_ = m.SetValue("5.3")
					return m
				}())
				return w
			}(),
			args{"test", "invalidFloat"},
			nil,
			&strconv.NumError{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := tt.warehouse

			err := w.pushGauge(tt.args.name, tt.args.value)
			if err != nil {
				require.EqualError(t, err, tt.wantErr.Error())
				return
			}

			require.Equal(t, tt.wantMetric.Value(), w.metrics[metrics.MetricTypeGauge][tt.args.name])

		})
	}
}
