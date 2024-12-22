package repositories

import (
	"testing"

	"github.com/gdyunin/metricol.git/internal/server/entity"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

/*
TestInMemoryRepository_Create tests the Create method of the InMemoryRepository.
It validates that metrics of valid types are successfully created and that unsupported types return an error.
*/
func TestInMemoryRepository_Create(t *testing.T) {
	repo := NewInMemoryRepository()

	tests := []struct {
		name        string
		metric      *entity.Metric
		expectError bool
	}{
		{
			name: "Create counter metric",
			metric: &entity.Metric{
				Name:  "test_counter",
				Type:  entity.MetricTypeCounter,
				Value: int64(42),
			},
			expectError: false,
		},
		{
			name: "Create gauge metric",
			metric: &entity.Metric{
				Name:  "test_gauge",
				Type:  entity.MetricTypeGauge,
				Value: 3.14,
			},
			expectError: false,
		},
		{
			name: "Unsupported metric type",
			metric: &entity.Metric{
				Name:  "test_unknown",
				Type:  "unknown",
				Value: nil,
			},
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := repo.Create(tc.metric)
			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

/*
TestInMemoryRepository_Read tests the Read method of the InMemoryRepository.
It validates that metrics can be retrieved correctly and that errors are returned for non-existent
or unsupported metric types.
*/
func TestInMemoryRepository_Read(t *testing.T) {
	repo := NewInMemoryRepository()

	// Prepopulate the repository.
	require.NoError(t, repo.Create(&entity.Metric{
		Name:  "test_counter",
		Type:  entity.MetricTypeCounter,
		Value: int64(100),
	}))
	require.NoError(t, repo.Create(&entity.Metric{
		Name:  "test_gauge",
		Type:  entity.MetricTypeGauge,
		Value: 1.23,
	}))

	tests := []struct {
		name        string
		filter      *entity.Filter
		expectError bool
		expected    *entity.Metric
	}{
		{
			name: "Read existing counter metric",
			filter: &entity.Filter{
				Name: "test_counter",
				Type: entity.MetricTypeCounter,
			},
			expectError: false,
			expected: &entity.Metric{
				Name:  "test_counter",
				Type:  entity.MetricTypeCounter,
				Value: int64(100),
			},
		},
		{
			name: "Read existing gauge metric",
			filter: &entity.Filter{
				Name: "test_gauge",
				Type: entity.MetricTypeGauge,
			},
			expectError: false,
			expected: &entity.Metric{
				Name:  "test_gauge",
				Type:  entity.MetricTypeGauge,
				Value: 1.23,
			},
		},
		{
			name: "Read non-existent metric",
			filter: &entity.Filter{
				Name: "non_existent",
				Type: entity.MetricTypeCounter,
			},
			expectError: true,
		},
		{
			name: "Unsupported metric type",
			filter: &entity.Filter{
				Name: "test_counter",
				Type: "unknown",
			},
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			metric, err := repo.Read(tc.filter)
			if tc.expectError {
				require.Error(t, err)
				assert.Nil(t, metric)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expected, metric)
			}
		})
	}
}

/*
TestInMemoryRepository_Update tests the Update method of the InMemoryRepository.
It validates that existing metrics can be updated and that unsupported metric types return an error.
*/
func TestInMemoryRepository_Update(t *testing.T) {
	repo := NewInMemoryRepository()

	// Prepopulate the repository.
	require.NoError(t, repo.Create(&entity.Metric{
		Name:  "test_counter",
		Type:  entity.MetricTypeCounter,
		Value: int64(100),
	}))

	tests := []struct {
		name        string
		metric      *entity.Metric
		expectError bool
	}{
		{
			name: "Update existing counter metric",
			metric: &entity.Metric{
				Name:  "test_counter",
				Type:  entity.MetricTypeCounter,
				Value: int64(200),
			},
			expectError: false,
		},
		{
			name: "Unsupported metric type",
			metric: &entity.Metric{
				Name:  "test_counter",
				Type:  "unknown",
				Value: nil,
			},
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := repo.Update(tc.metric)
			if tc.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)

				// Verify the update.
				updatedMetric, _ := repo.Read(&entity.Filter{Name: tc.metric.Name, Type: tc.metric.Type})
				assert.Equal(t, tc.metric.Value, updatedMetric.Value)
			}
		})
	}
}

/*
TestInMemoryRepository_IsExists tests the IsExists method of the InMemoryRepository.
It verifies that existing metrics are correctly identified and unsupported metric types return an error.
*/
func TestInMemoryRepository_IsExists(t *testing.T) {
	repo := NewInMemoryRepository()

	// Prepopulate the repository.
	require.NoError(t, repo.Create(&entity.Metric{
		Name:  "test_counter",
		Type:  entity.MetricTypeCounter,
		Value: int64(100),
	}))

	tests := []struct {
		name        string
		filter      *entity.Filter
		expectError bool
		expectExist bool
	}{
		{
			name: "Check existing counter metric",
			filter: &entity.Filter{
				Name: "test_counter",
				Type: entity.MetricTypeCounter,
			},
			expectError: false,
			expectExist: true,
		},
		{
			name: "Check non-existent metric",
			filter: &entity.Filter{
				Name: "non_existent",
				Type: entity.MetricTypeCounter,
			},
			expectError: false,
			expectExist: false,
		},
		{
			name: "Unsupported metric type",
			filter: &entity.Filter{
				Name: "test_counter",
				Type: "unknown",
			},
			expectError: true,
			expectExist: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			exists, err := repo.IsExists(tc.filter)
			if tc.expectError {
				require.Error(t, err)
				assert.False(t, exists)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expectExist, exists)
			}
		})
	}
}
