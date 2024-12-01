package metrics

import (
	"fmt"
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
			name:          "Create simple gauge",
			inputName:     "cpu_usage",
			inputValue:    75.5,
			expectedName:  "cpu_usage",
			expectedValue: 75.5,
		},
		{
			name:          "Create gauge with zero value",
			inputName:     "memory_usage",
			inputValue:    0.0,
			expectedName:  "memory_usage",
			expectedValue: 0.0,
		},
		{
			name:          "Create gauge with negative value",
			inputName:     "temperature",
			inputValue:    -20.3,
			expectedName:  "temperature",
			expectedValue: -20.3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gauge := NewGauge(tt.inputName, tt.inputValue)
			require.Equal(t, tt.expectedName, gauge.Name)
			require.Equal(t, tt.expectedValue, gauge.Value)
			require.IsType(t, &Gauge{}, gauge)
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
			name:          "Create simple counter",
			inputName:     "requests",
			inputValue:    100,
			expectedName:  "requests",
			expectedValue: 100,
		},
		{
			name:          "Create counter with zero value",
			inputName:     "errors",
			inputValue:    0,
			expectedName:  "errors",
			expectedValue: 0,
		},
		{
			name:          "Create counter with negative value",
			inputName:     "negative_counter",
			inputValue:    -50,
			expectedName:  "negative_counter",
			expectedValue: -50,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			counter := NewCounter(tt.inputName, tt.inputValue)
			require.Equal(t, tt.expectedName, counter.Name)
			require.Equal(t, tt.expectedValue, counter.Value)
			require.IsType(t, &Counter{}, counter)
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
			name:        "Create gauge from valid strings",
			inputName:   "cpu_usage",
			inputValue:  "65.5",
			inputType:   MetricTypeGauge,
			expectedErr: nil,
			expectedMetric: &Gauge{
				Name:  "cpu_usage",
				Value: 65.5,
			},
		},
		{
			name:        "Create counter from valid strings",
			inputName:   "requests",
			inputValue:  "150",
			inputType:   MetricTypeCounter,
			expectedErr: nil,
			expectedMetric: &Counter{
				Name:  "requests",
				Value: 150,
			},
		},
		{
			name:           "Try create unknown",
			inputName:      "unknown_metric",
			inputValue:     "123",
			inputType:      "unknown",
			expectedErr:    fmt.Errorf("unknown metric type: %s", "unknown"),
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

			switch m := metric.(type) {
			case *Counter:
				require.Equal(t, tt.inputName, m.Name)
			case *Gauge:
				require.Equal(t, tt.inputName, m.Name)
			default:
				require.Fail(t, "Metric isn`t counter or gauge!")
			}
			require.Equal(t, tt.expectedMetric.StringValue(), metric.StringValue())
			require.IsType(t, tt.expectedMetric, metric)
		})
	}
}
