package repositories

import (
	"testing"

	"github.com/gdyunin/metricol.git/internal/agent/entity"
	"github.com/stretchr/testify/assert"
)

/*
TestInMemoryRepository_Add tests the Add method of the InMemoryRepository.
It validates that metrics are correctly added, updated, or ignored when nil.
*/
func TestInMemoryRepository_Add(t *testing.T) {
	tests := []struct {
		name            string
		initialMetrics  []*entity.Metric
		newMetric       *entity.Metric
		expectedMetrics []*entity.Metric
	}{
		{
			name:           "Add a new metric to an empty repository",
			initialMetrics: []*entity.Metric{},
			newMetric: &entity.Metric{
				Name:  "metric1",
				Type:  "gauge",
				Value: 42.0,
			},
			expectedMetrics: []*entity.Metric{
				{
					Name:  "metric1",
					Type:  "gauge",
					Value: 42.0,
				},
			},
		},
		{
			name: "Update an existing metric's value",
			initialMetrics: []*entity.Metric{
				{
					Name:  "metric1",
					Type:  "gauge",
					Value: 10.0,
				},
			},
			newMetric: &entity.Metric{
				Name:  "metric1",
				Type:  "gauge",
				Value: 42.0,
			},
			expectedMetrics: []*entity.Metric{
				{
					Name:  "metric1",
					Type:  "gauge",
					Value: 42.0,
				},
			},
		},
		{
			name: "Add a metric to a repository with existing metrics",
			initialMetrics: []*entity.Metric{
				{
					Name:  "metric1",
					Type:  "gauge",
					Value: 10.0,
				},
			},
			newMetric: &entity.Metric{
				Name:  "metric2",
				Type:  "counter",
				Value: 1.0,
			},
			expectedMetrics: []*entity.Metric{
				{
					Name:  "metric1",
					Type:  "gauge",
					Value: 10.0,
				},
				{
					Name:  "metric2",
					Type:  "counter",
					Value: 1.0,
				},
			},
		},
		{
			name: "Ignore nil metric",
			initialMetrics: []*entity.Metric{
				{
					Name:  "metric1",
					Type:  "gauge",
					Value: 10.0,
				},
			},
			newMetric: nil,
			expectedMetrics: []*entity.Metric{
				{
					Name:  "metric1",
					Type:  "gauge",
					Value: 10.0,
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := NewInMemoryRepository()
			// Prepopulate the repository with initial metrics.
			for _, metric := range tc.initialMetrics {
				repo.Add(metric)
			}

			// Add the new metric.
			repo.Add(tc.newMetric)

			// Retrieve the metrics and validate.
			result := repo.Metrics()
			assert.Equal(t, tc.expectedMetrics, result)
		})
	}
}

/*
TestInMemoryRepository_Metrics tests the Metrics method of the InMemoryRepository.
It ensures that all metrics in the repository are retrieved correctly.
*/
func TestInMemoryRepository_Metrics(t *testing.T) {
	tests := []struct {
		name            string
		initialMetrics  []*entity.Metric
		expectedMetrics []*entity.Metric
	}{
		{
			name:            "Retrieve metrics from an empty repository",
			initialMetrics:  []*entity.Metric{},
			expectedMetrics: []*entity.Metric{},
		},
		{
			name: "Retrieve metrics from a repository with one metric",
			initialMetrics: []*entity.Metric{
				{
					Name:  "metric1",
					Type:  "gauge",
					Value: 42.0,
				},
			},
			expectedMetrics: []*entity.Metric{
				{
					Name:  "metric1",
					Type:  "gauge",
					Value: 42.0,
				},
			},
		},
		{
			name: "Retrieve metrics from a repository with multiple metrics",
			initialMetrics: []*entity.Metric{
				{
					Name:  "metric1",
					Type:  "gauge",
					Value: 42.0,
				},
				{
					Name:  "metric2",
					Type:  "counter",
					Value: 1.0,
				},
			},
			expectedMetrics: []*entity.Metric{
				{
					Name:  "metric1",
					Type:  "gauge",
					Value: 42.0,
				},
				{
					Name:  "metric2",
					Type:  "counter",
					Value: 1.0,
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := NewInMemoryRepository()
			// Prepopulate the repository with initial metrics.
			for _, metric := range tc.initialMetrics {
				repo.Add(metric)
			}

			// Retrieve the metrics and validate.
			result := repo.Metrics()
			assert.Equal(t, tc.expectedMetrics, result)
		})
	}
}
