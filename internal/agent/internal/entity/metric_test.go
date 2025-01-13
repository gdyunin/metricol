package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMetrics_Length(t *testing.T) {
	tests := []struct {
		name     string
		metrics  Metrics
		expected int
	}{
		{
			name:     "Nil metrics",
			metrics:  nil,
			expected: 0,
		},
		{
			name:     "Empty metrics",
			metrics:  Metrics{},
			expected: 0,
		},
		{
			name: "Metrics with one element",
			metrics: Metrics{
				&Metric{Name: "test_metric", Type: MetricTypeCounter, Value: 123},
			},
			expected: 1,
		},
		{
			name: "Metrics with multiple elements",
			metrics: Metrics{
				&Metric{Name: "metric1", Type: MetricTypeGauge, Value: 1.23},
				&Metric{Name: "metric2", Type: MetricTypeCounter, Value: 456},
				&Metric{Name: "metric3", Type: MetricTypeGauge, Value: 7.89},
			},
			expected: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.metrics.Length()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMetrics_ToString(t *testing.T) {
	tests := []struct {
		name     string
		expected string
		metrics  Metrics
	}{
		{
			name:     "Nil metrics",
			metrics:  nil,
			expected: "",
		},
		{
			name:     "Empty metrics",
			metrics:  Metrics{},
			expected: "",
		},
		{
			name: "Metrics with one element",
			metrics: Metrics{
				&Metric{Name: "test_metric", Type: MetricTypeCounter, Value: 123},
			},
			expected: "<Name=test_metric Type=counter Value=123>",
		},
		{
			name: "Metrics with multiple elements",
			metrics: Metrics{
				&Metric{Name: "metric1", Type: MetricTypeGauge, Value: 1.23},
				&Metric{Name: "metric2", Type: MetricTypeCounter, Value: 456},
				&Metric{Name: "metric3", Type: MetricTypeGauge, Value: 7.89},
			},
			expected: "<Name=metric1 Type=gauge Value=1.23>, " +
				"<Name=metric2 Type=counter Value=456>, " +
				"<Name=metric3 Type=gauge Value=7.89>",
		},
		{
			name: "Metrics with nil elements",
			metrics: Metrics{
				&Metric{Name: "metric1", Type: MetricTypeGauge, Value: 1.23},
				nil,
				&Metric{Name: "metric3", Type: MetricTypeGauge, Value: 7.89},
			},
			expected: "<Name=metric1 Type=gauge Value=1.23>, <nil>, <Name=metric3 Type=gauge Value=7.89>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.metrics.ToString()
			assert.Equal(t, tt.expected, result)
		})
	}
}
