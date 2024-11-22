package library

import (
	"errors"
	"github.com/gdyunin/metricol.git/internal/server/metrics"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestCounter_Name(t *testing.T) {
	tests := []struct {
		name    string
		counter Counter
		want    string
	}{
		{
			"get simple name",
			Counter{
				name:  "simple_name",
				value: 0,
			},
			"simple_name",
		},
		{
			"get long name",
			Counter{
				name:  "loooooooooooooooooooooooooooooooooooooooooooooooong_name",
				value: 0,
			},
			"loooooooooooooooooooooooooooooooooooooooooooooooong_name",
		},
		{
			"get name with numbers",
			Counter{
				name:  "123_some_567",
				value: 0,
			},
			"123_some_567",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.counter.Name()
			require.Equal(t, tt.want, got)
		})
	}
}

func TestCounter_SetName(t *testing.T) {
	tests := []struct {
		name    string
		counter *Counter
		newName string
		want    string
		wantErr error
	}{
		{
			"set simple name",
			&Counter{},
			"simple_name",
			"simple_name",
			nil,
		},
		{
			"set long name",
			&Counter{},
			"loooooooooooooooooooooooooooooooooooooooooooooooong_name",
			"loooooooooooooooooooooooooooooooooooooooooooooooong_name",
			nil,
		},
		{
			"set name with numbers",
			&Counter{},
			"123_some_567",
			"123_some_567",
			nil,
		},
		{
			"set empty name",
			&Counter{},
			"",
			"",
			errors.New(metrics.ErrorEmptyName),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := tt.counter

			err := c.SetName(tt.newName)
			if err != nil {
				require.EqualError(t, tt.wantErr, err.Error())
				return
			}

			require.Equal(t, tt.want, c.name)
		})
	}
}

func TestCounter_SetValue(t *testing.T) {

	tests := []struct {
		name     string
		counter  *Counter
		newValue string
		want     int64
		wantErr  error
	}{
		{
			"set valid int",
			&Counter{},
			"5",
			5,
			nil,
		},
		{
			"set invalid int",
			&Counter{},
			"invalidInt",
			0,
			errors.New(metrics.ErrorParseMetricValue),
		},
		{
			"set empty value",
			&Counter{},
			"",
			0,
			errors.New(metrics.ErrorEmptyValue),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := tt.counter

			err := c.SetValue(tt.newValue)
			if err != nil {
				require.EqualError(t, err, tt.wantErr.Error())
				return
			}

			require.Equal(t, tt.want, c.value)
		})
	}
}

func TestCounter_Type(t *testing.T) {
	tests := []struct {
		name    string
		counter Counter
		want    metrics.MetricType
	}{
		{
			"get counter type",
			Counter{},
			metrics.MetricTypeCounter,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.counter.Type()
			require.Equal(t, tt.want, got)
		})
	}
}

func TestCounter_Value(t *testing.T) {
	tests := []struct {
		name    string
		counter Counter
		want    string
	}{
		{
			"get counter simple value",
			Counter{
				name:  "simple_name",
				value: 0,
			},
			"0",
		},
		{
			"get counter negative value",
			Counter{
				name:  "simple_name",
				value: -5,
			},
			"-5",
		},
		{
			"get counter big value",
			Counter{
				name:  "simple_name",
				value: 99999999999999999,
			},
			"99999999999999999",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := tt.counter

			v := c.Value()
			require.Equal(t, tt.want, v)
		})
	}
}

func TestNewCounter(t *testing.T) {
	tests := []struct {
		name string
		want *Counter
	}{
		{
			"init new counter",
			&Counter{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewCounter()
			require.Equal(t, tt.want, c)
		})
	}
}
