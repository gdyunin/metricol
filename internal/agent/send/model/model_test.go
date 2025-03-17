package model

import (
	"testing"

	"github.com/gdyunin/metricol.git/internal/agent/internal/entity"
	"github.com/stretchr/testify/assert"
)

func TestNewFromEntityMetric(t *testing.T) {
	tests := []struct {
		name        string
		input       *entity.Metric
		expectError bool
		expected    *Metric
	}{
		{
			"Valid counter metric",
			&entity.Metric{Name: "test_counter", Type: "counter", Value: int64(42)},
			false,
			&Metric{ID: "test_counter", MType: "counter", Delta: int64Ptr(42)},
		},
		{
			"Valid gauge metric",
			&entity.Metric{Name: "test_gauge", Type: "gauge", Value: float64(3.14)},
			false,
			&Metric{ID: "test_gauge", MType: "gauge", Value: float64Ptr(3.14)},
		},
		{
			"Invalid counter metric type",
			&entity.Metric{Name: "invalid_counter", Type: "counter", Value: "string"},
			true,
			nil,
		},
		{
			"Invalid gauge metric type",
			&entity.Metric{Name: "invalid_gauge", Type: "gauge", Value: "string"},
			true,
			nil,
		},
		{
			"Unsupported metric type",
			&entity.Metric{Name: "unsupported", Type: "unknown", Value: 100},
			true,
			nil,
		},
		{
			"Nil entity metric",
			nil,
			true,
			nil,
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
