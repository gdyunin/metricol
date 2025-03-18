package model

import (
	"testing"

	"github.com/gdyunin/metricol.git/internal/server/internal/entity"
	"github.com/stretchr/testify/assert"
)

func TestToEntityMetric(t *testing.T) {
	tests := []struct {
		input    *Metric
		expected *entity.Metric
		name     string
	}{
		{
			name:     "Convert counter metric",
			input:    &Metric{ID: "test_counter", MType: "counter", Delta: int64Ptr(5)},
			expected: &entity.Metric{Name: "test_counter", Type: "counter", Value: int64(5)},
		},
		{
			name:     "Convert gauge metric",
			input:    &Metric{ID: "test_gauge", MType: "gauge", Value: float64Ptr(3.14)},
			expected: &entity.Metric{Name: "test_gauge", Type: "gauge", Value: float64(3.14)},
		},
		{
			name:     "Invalid metric type",
			input:    &Metric{ID: "test_invalid", MType: "invalid"},
			expected: &entity.Metric{Name: "test_invalid", Type: "invalid", Value: nil},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.input.ToEntityMetric()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFromEntityMetric(t *testing.T) {
	tests := []struct {
		input    *entity.Metric
		expected *Metric
		name     string
	}{
		{
			name:     "Convert entity counter metric",
			input:    &entity.Metric{Name: "test_counter", Type: "counter", Value: int64(10)},
			expected: &Metric{ID: "test_counter", MType: "counter", Delta: int64Ptr(10)},
		},
		{
			name:     "Convert entity gauge metric",
			input:    &entity.Metric{Name: "test_gauge", Type: "gauge", Value: float64(2.71)},
			expected: &Metric{ID: "test_gauge", MType: "gauge", Value: float64Ptr(2.71)},
		},
		{
			name:     "Invalid entity metric type",
			input:    &entity.Metric{Name: "test_invalid", Type: "invalid"},
			expected: &Metric{ID: "test_invalid", MType: "invalid"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FromEntityMetric(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func int64Ptr(i int64) *int64 {
	return &i
}

func float64Ptr(f float64) *float64 {
	return &f
}
