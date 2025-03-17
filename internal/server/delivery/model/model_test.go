package model

import (
	"testing"

	"github.com/gdyunin/metricol.git/internal/server/internal/entity"
	"github.com/stretchr/testify/assert"
)

func TestToEntityMetric(t *testing.T) {
	tests := []struct {
		name     string
		input    *Metric
		expected *entity.Metric
	}{
		{
			"Convert counter metric",
			&Metric{ID: "test_counter", MType: "counter", Delta: int64Ptr(5)},
			&entity.Metric{Name: "test_counter", Type: "counter", Value: int64(5)},
		},
		{
			"Convert gauge metric",
			&Metric{ID: "test_gauge", MType: "gauge", Value: float64Ptr(3.14)},
			&entity.Metric{Name: "test_gauge", Type: "gauge", Value: float64(3.14)},
		},
		{
			"Invalid metric type",
			&Metric{ID: "test_invalid", MType: "invalid"},
			&entity.Metric{Name: "test_invalid", Type: "invalid", Value: nil},
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
		name     string
		input    *entity.Metric
		expected *Metric
	}{
		{
			"Convert entity counter metric",
			&entity.Metric{Name: "test_counter", Type: "counter", Value: int64(10)},
			&Metric{ID: "test_counter", MType: "counter", Delta: int64Ptr(10)},
		},
		{
			"Convert entity gauge metric",
			&entity.Metric{Name: "test_gauge", Type: "gauge", Value: float64(2.71)},
			&Metric{ID: "test_gauge", MType: "gauge", Value: float64Ptr(2.71)},
		},
		{
			"Invalid entity metric type",
			&entity.Metric{Name: "test_invalid", Type: "invalid"},
			&Metric{ID: "test_invalid", MType: "invalid"},
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
