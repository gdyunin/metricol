package repository

import (
	"context"

	"github.com/gdyunin/metricol.git/internal/server/internal/entity"
)

const (
	defaultAttemptsDefaultCount = 4 // Default count of attempts for retry calls.
)

// Repository defines the interface for a metric storage repository.
// It provides methods for managing and querying metrics.
type Repository interface {
	// Update adds or updates a metric in the repository.
	//
	// Parameters:
	//   - metric: Pointer to the Metric to be added or updated.
	//
	// Returns:
	//   - An error if the operation fails.
	Update(context.Context, *entity.Metric) error

	// UpdateBatch adds or updates a metrics batch in the repository.
	//
	// Parameters:
	//   - metrics: Pointer to the Metrics to be added or updated.
	//
	// Returns:
	//   - An error if the operation fails.
	UpdateBatch(ctx context.Context, metrics *entity.Metrics) error

	// IsExist checks if a metric exists in the repository.
	//
	// Parameters:
	//   - metricType: The type of the metric (e.g., counter, gauge).
	//   - metricName: The name of the metric.
	//
	// Returns:
	//   - A boolean indicating whether the metric exists.
	//   - An error if the operation fails.
	IsExist(ctx context.Context, metricType string, metricName string) (bool, error)

	// Find retrieves a metric from the repository by type and name.
	//
	// Parameters:
	//   - metricType: The type of the metric (e.g., counter, gauge).
	//   - metricName: The name of the metric.
	//
	// Returns:
	//   - A pointer to the Metric entity if found.
	//   - An error if the metric is not found or another issue occurs.
	Find(ctx context.Context, metricType string, metricName string) (*entity.Metric, error)

	// All retrieves all metrics from the repository.
	//
	// Returns:
	//   - A pointer to a Metrics slice containing all metrics.
	//   - An error if the operation fails.
	All(context.Context) (*entity.Metrics, error)

	// CheckConnection verifies the repository's connection.
	//
	// Parameters:
	//   - ctx: The context used to manage the connection check lifecycle.
	//
	// Returns:
	//   - An error if the connection check fails.
	CheckConnection(context.Context) error
}
