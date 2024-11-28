package metrics

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewGauge(t *testing.T) {
	tests := []struct {
		name          string
		inputName     string
		inputValue    float64
		expectedName  string
		expectedValue float64
	}{
		{
			name:          "create simple gauge",
			inputName:     "cpu_usage",
			inputValue:    75.5,
			expectedName:  "cpu_usage",
			expectedValue: 75.5,
		},
		{
			name:          "create gauge with zero value",
			inputName:     "memory_usage",
			inputValue:    0.0,
			expectedName:  "memory_usage",
			expectedValue: 0.0,
		},
		{
			name:          "create gauge with negative value",
			inputName:     "temperature",
			inputValue:    -20.3,
			expectedName:  "temperature",
			expectedValue: -20.3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gauge := NewGauge(tt.inputName, tt.inputValue)
			require.Equal(t, tt.expectedName, gauge.Name())
			require.Equal(t, tt.expectedValue, gauge.Value())
			require.Equal(t, MetricTypeGauge, gauge.Type())
		})
	}
}

func TestNewCounter(t *testing.T) {
	tests := []struct {
		name          string
		inputName     string
		inputValue    int64
		expectedName  string
		expectedValue int64
	}{
		{
			name:          "create simple counter",
			inputName:     "requests",
			inputValue:    100,
			expectedName:  "requests",
			expectedValue: 100,
		},
		{
			name:          "create counter with zero value",
			inputName:     "errors",
			inputValue:    0,
			expectedName:  "errors",
			expectedValue: 0,
		},
		{
			name:          "create counter with negative value",
			inputName:     "negative_counter",
			inputValue:    -50,
			expectedName:  "negative_counter",
			expectedValue: -50,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			counter := NewCounter(tt.inputName, tt.inputValue)
			require.Equal(t, tt.expectedName, counter.Name())
			require.Equal(t, tt.expectedValue, counter.Value())
			require.Equal(t, MetricTypeCounter, counter.Type())
		})
	}
}

func TestNewFromStrings(t *testing.T) {
	tests := []struct {
		name           string
		inputName      string
		inputValue     string
		inputType      string
		expectedErr    error
		expectedMetric Metric
	}{
		{
			name:        "create gauge from valid strings",
			inputName:   "cpu_usage",
			inputValue:  "65.5",
			inputType:   MetricTypeGauge,
			expectedErr: nil,
			expectedMetric: &Gauge{
				name:  "cpu_usage",
				value: 65.5,
			},
		},
		{
			name:        "create counter from valid strings",
			inputName:   "requests",
			inputValue:  "150",
			inputType:   MetricTypeCounter,
			expectedErr: nil,
			expectedMetric: &Counter{
				name:  "requests",
				value: 150,
			},
		},
		{
			name:           "try create unknown",
			inputName:      "unknown_metric",
			inputValue:     "123",
			inputType:      "unknown",
			expectedErr:    errors.New(ErrorUnknownMetricType),
			expectedMetric: nil,
		},
		{
			name:           "create gauge from invalid(value) strings",
			inputName:      "invalid_gauge",
			inputValue:     "not_a_float",
			inputType:      MetricTypeGauge,
			expectedErr:    errors.New(ErrorParseMetricValue),
			expectedMetric: nil,
		},
		{
			name:           "create counter from invalid(value) strings",
			inputName:      "invalid_counter",
			inputValue:     "not_an_int",
			inputType:      MetricTypeCounter,
			expectedErr:    errors.New(ErrorParseMetricValue),
			expectedMetric: nil,
		},
		{
			name:           "try create gauge from empty(value) strings",
			inputName:      "empty_gauge",
			inputValue:     "",
			inputType:      MetricTypeGauge,
			expectedErr:    errors.New(ErrorParseMetricValue),
			expectedMetric: nil,
		},
		{
			name:           "try create counter from empty(value) strings",
			inputName:      "empty_counter",
			inputValue:     "",
			inputType:      MetricTypeCounter,
			expectedErr:    errors.New(ErrorParseMetricValue),
			expectedMetric: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metric, err := NewFromStrings(tt.inputName, tt.inputValue, tt.inputType)
			if err != nil {
				require.Equal(t, tt.expectedErr.Error(), err.Error())
				return
			}
			require.Equal(t, tt.expectedMetric.Name(), metric.Name())
			require.Equal(t, tt.expectedMetric.StringValue(), metric.StringValue())
			require.Equal(t, tt.expectedMetric.Type(), metric.Type())
		})
	}
}
