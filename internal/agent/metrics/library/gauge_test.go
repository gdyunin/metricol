package library

import (
	"github.com/gdyunin/metricol.git/internal/agent/metrics"
	"github.com/stretchr/testify/require"
	"strconv"
	"testing"
)

func TestGauge_Name(t *testing.T) {
	type fields struct {
		name    string
		value   float64
		fetcher func() float64
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			"get simple name",
			fields{
				name:    "simple name",
				value:   0,
				fetcher: nil,
			},
			"simple name",
		},
		{
			"get long name",
			fields{
				name:    "loooooooooooooooooooooooooooooooooooooooooong name",
				value:   0,
				fetcher: nil,
			},
			"loooooooooooooooooooooooooooooooooooooooooong name",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := Gauge{
				name:    tt.fields.name,
				value:   tt.fields.value,
				fetcher: tt.fields.fetcher,
			}
			got := g.Name()
			require.Equal(t, tt.want, got)
		})
	}
}

func TestGauge_Type(t *testing.T) {
	type fields struct {
		name    string
		value   float64
		fetcher func() float64
	}
	tests := []struct {
		name   string
		fields fields
		want   metrics.MetricType
	}{
		{
			"get type",
			fields{
				name:    "name",
				value:   0,
				fetcher: nil,
			},
			metrics.MetricTypeGauge,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := Gauge{
				name:    tt.fields.name,
				value:   tt.fields.value,
				fetcher: tt.fields.fetcher,
			}
			got := g.Type()
			require.Equal(t, tt.want, got)
		})
	}
}

func TestGauge_UpdateValue(t *testing.T) {
	testFetcher := func() float64 { return 4.2 }
	type fields struct {
		name    string
		value   float64
		fetcher func() float64
	}
	tests := []struct {
		name   string
		fields fields
		want   fields
	}{
		{
			"update value",
			fields{
				name:    "test",
				value:   0,
				fetcher: testFetcher,
			},
			fields{
				name:    "test",
				value:   4.2,
				fetcher: testFetcher,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &Gauge{
				name:    tt.fields.name,
				value:   tt.fields.value,
				fetcher: tt.fields.fetcher,
			}
			g.UpdateValue()

			e := &Gauge{
				name:    tt.want.name,
				value:   tt.want.value,
				fetcher: tt.want.fetcher,
			}
			require.EqualExportedValues(t, e, g)
		})
	}
}

func TestGauge_Value(t *testing.T) {
	type fields struct {
		name    string
		value   float64
		fetcher func() float64
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			"get type",
			fields{
				name:    "name",
				value:   4.2,
				fetcher: nil,
			},
			strconv.FormatFloat(4.2, 'f', 6, 64),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := Gauge{
				name:    tt.fields.name,
				value:   tt.fields.value,
				fetcher: tt.fields.fetcher,
			}
			got := g.Value()
			require.Equal(t, tt.want, got)
		})
	}
}

func TestNewGauge(t *testing.T) {
	testFetcher := func() float64 { return 4.2 }
	type args struct {
		name    string
		fetcher func() float64
	}
	tests := []struct {
		name string
		args args
		want *Gauge
	}{
		{
			"create new gauge metric",
			args{
				name:    "test",
				fetcher: testFetcher,
			},
			&Gauge{
				name:    "test",
				fetcher: testFetcher,
				value:   4.2,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewGauge(tt.args.name, tt.args.fetcher)
			require.EqualExportedValues(t, tt.want, got)
		})
	}
}
