package model

import (
	"testing"

	"github.com/gdyunin/metricol.git/internal/agent/internal/entity"
	"github.com/stretchr/testify/assert"
)

func TestNewFromEntityMetric(t *testing.T) {
	tests := []struct {
		input       *entity.Metric
		expected    *Metric
		name        string
		expectError bool
	}{
		{
			name:     "Valid counter metric",
			input:    &entity.Metric{Name: "test_counter", Type: "counter", Value: int64(42)},
			expected: &Metric{ID: "test_counter", MType: "counter", Delta: int64Ptr(42)},
		},
		{
			name:     "Valid gauge metric",
			input:    &entity.Metric{Name: "test_gauge", Type: "gauge", Value: float64(3.14)},
			expected: &Metric{ID: "test_gauge", MType: "gauge", Value: float64Ptr(3.14)},
		},
		{
			name:        "Invalid counter metric type",
			input:       &entity.Metric{Name: "invalid_counter", Type: "counter", Value: "string"},
			expectError: true,
		},
		{
			name:        "Invalid gauge metric type",
			input:       &entity.Metric{Name: "invalid_gauge", Type: "gauge", Value: "string"},
			expectError: true,
		},
		{
			name:        "Unsupported metric type",
			input:       &entity.Metric{Name: "unsupported", Type: "unknown", Value: 100},
			expectError: true,
		},
		{
			name:        "Nil entity metric",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := NewFromEntityMetric(tt.input)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func int64Ptr(i int64) *int64 {
	return &i
}

func float64Ptr(f float64) *float64 {
	return &f
}
