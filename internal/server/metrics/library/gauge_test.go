package library

import (
	"errors"
	"github.com/gdyunin/metricol.git/internal/server/metrics"
	"github.com/stretchr/testify/require"
	"strconv"
	"testing"
)

func TestGauge_Name(t *testing.T) {
	tests := []struct {
		name  string
		gauge Gauge
		want  string
	}{
		{
			"get simple name",
			Gauge{
				name:  "simple_name",
				value: 0,
			},
			"simple_name",
		},
		{
			"get long name",
			Gauge{
				name:  "loooooooooooooooooooooooooooooooooooooooooooooooong_name",
				value: 0,
			},
			"loooooooooooooooooooooooooooooooooooooooooooooooong_name",
		},
		{
			"get name with numbers",
			Gauge{
				name:  "123_some_567",
				value: 0,
			},
			"123_some_567",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.gauge.Name()
			require.Equal(t, tt.want, got)
		})
	}
}

func TestGauge_SetName(t *testing.T) {
	tests := []struct {
		name    string
		gauge   *Gauge
		newName string
		want    string
		wantErr error
	}{
		{
			"set simple name",
			&Gauge{},
			"simple_name",
			"simple_name",
			nil,
		},
		{
			"set long name",
			&Gauge{},
			"loooooooooooooooooooooooooooooooooooooooooooooooong_name",
			"loooooooooooooooooooooooooooooooooooooooooooooooong_name",
			nil,
		},
		{
			"set name with numbers",
			&Gauge{},
			"123_some_567",
			"123_some_567",
			nil,
		},
		{
			"set empty name",
			&Gauge{},
			"",
			"",
			errors.New(metrics.ErrorEmptyName),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := tt.gauge

			err := g.SetName(tt.newName)
			if err != nil {
				require.EqualError(t, tt.wantErr, err.Error())
				return
			}

			require.Equal(t, tt.want, g.name)
		})
	}
}

func TestGauge_SetValue(t *testing.T) {

	tests := []struct {
		name     string
		gauge    *Gauge
		newValue string
		want     float64
		wantErr  error
	}{
		{
			"set valid Float",
			&Gauge{},
			"5.3",
			5.3,
			nil,
		},
		{
			"set invalid Float",
			&Gauge{},
			"invalidFloat",
			0,
			errors.New(metrics.ErrorParseMetricValue),
		},
		{
			"set empty value",
			&Gauge{},
			"",
			0,
			errors.New(metrics.ErrorEmptyValue),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := tt.gauge

			err := g.SetValue(tt.newValue)
			if err != nil {
				require.EqualError(t, err, tt.wantErr.Error())
				return
			}

			require.Equal(t, tt.want, g.value)
		})
	}
}

func TestGauge_Type(t *testing.T) {
	tests := []struct {
		name  string
		gauge Gauge
		want  metrics.MetricType
	}{
		{
			"get gauge type",
			Gauge{},
			metrics.MetricTypeGauge,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.gauge.Type()
			require.Equal(t, tt.want, got)
		})
	}
}

func TestGauge_Value(t *testing.T) {
	tests := []struct {
		name  string
		gauge Gauge
		want  string
	}{
		{
			"get gauge simple value",
			Gauge{
				name:  "simple_name",
				value: 5.3,
			},
			strconv.FormatFloat(5.3, 'f', 6, 64),
		},
		{
			"get gauge negative value",
			Gauge{
				name:  "simple_name",
				value: -5.3,
			},
			strconv.FormatFloat(-5.3, 'f', 6, 64),
		},
		{
			"get gauge big value",
			Gauge{
				name:  "simple_name",
				value: 9999999.3548762,
			},
			strconv.FormatFloat(9999999.3548762, 'f', 6, 64),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := tt.gauge

			v := g.Value()
			require.Equal(t, tt.want, v)
		})
	}
}

func TestNewGauge(t *testing.T) {
	tests := []struct {
		name string
		want *Gauge
	}{
		{
			"init new gauge",
			&Gauge{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewGauge()
			require.Equal(t, tt.want, g)
		})
	}
}
