package library

import (
	"github.com/gdyunin/metricol.git/internal/agent/metrics"
	"github.com/stretchr/testify/require"
	"strconv"
	"testing"
)

func TestCounter_Name(t *testing.T) {
	type fields struct {
		name    string
		value   int64
		fetcher func() int64
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
			g := Counter{
				name:    tt.fields.name,
				value:   tt.fields.value,
				fetcher: tt.fields.fetcher,
			}
			got := g.Name()
			require.Equal(t, tt.want, got)
		})
	}
}

func TestCounter_Type(t *testing.T) {
	type fields struct {
		name    string
		value   int64
		fetcher func() int64
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
			metrics.MetricTypeCounter,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := Counter{
				name:    tt.fields.name,
				value:   tt.fields.value,
				fetcher: tt.fields.fetcher,
			}
			got := g.Type()
			require.Equal(t, tt.want, got)
		})
	}
}

func TestCounter_UpdateValue(t *testing.T) {
	testFetcher := func() int64 { return 42 }
	type fields struct {
		name    string
		value   int64
		fetcher func() int64
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
				value:   42,
				fetcher: testFetcher,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &Counter{
				name:    tt.fields.name,
				value:   tt.fields.value,
				fetcher: tt.fields.fetcher,
			}
			g.UpdateValue()

			e := &Counter{
				name:    tt.want.name,
				value:   tt.want.value,
				fetcher: tt.want.fetcher,
			}
			require.EqualExportedValues(t, e, g)
		})
	}
}

func TestCounter_Value(t *testing.T) {
	type fields struct {
		name    string
		value   int64
		fetcher func() int64
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
				value:   42,
				fetcher: nil,
			},
			strconv.FormatInt(42, 10),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := Counter{
				name:    tt.fields.name,
				value:   tt.fields.value,
				fetcher: tt.fields.fetcher,
			}
			got := g.Value()
			require.Equal(t, tt.want, got)
		})
	}
}

func TestNewCounter(t *testing.T) {
	testFetcher := func() int64 { return 42 }
	type args struct {
		name    string
		fetcher func() int64
	}
	tests := []struct {
		name string
		args args
		want *Counter
	}{
		{
			"create new counter metric",
			args{
				name:    "test",
				fetcher: testFetcher,
			},
			&Counter{
				name:    "test",
				fetcher: testFetcher,
				value:   42,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewCounter(tt.args.name, tt.args.fetcher)
			require.EqualExportedValues(t, tt.want, got)
		})
	}
}
