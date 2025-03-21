package repository

import (
	"context"
	"errors"

	"github.com/gdyunin/metricol.git/internal/server/internal/entity"
)

const (
	// Const defaultAttemptsDefaultCount is the default count of attempts for retry calls.
	defaultAttemptsDefaultCount = 4
)

var (
	// ErrNotFoundInRepo is returned when a requested metric is not found in the repository.
	ErrNotFoundInRepo = errors.New("not found in repository")
)

// Repository defines the interface for a metric storage repository.
// It provides methods for adding, updating, retrieving, and checking metrics.
type Repository interface {
	// Update adds or updates a metric in the repository.
	//
	// Parameters:
	//   - ctx: The context for the operation.
	//   - metric: A pointer to the Metric to be added or updated.
	//
	// Returns:
	//   - error: An error if the operation fails.
	Update(context.Context, *entity.Metric) error

	// UpdateBatch adds or updates a batch of metrics in the repository.
	//
	// Parameters:
	//   - ctx: The context for the operation.
	//   - metrics: A pointer to the Metrics to be added or updated.
	//
	// Returns:
	//   - error: An error if the operation fails.
	UpdateBatch(ctx context.Context, metrics *entity.Metrics) error

	// Find retrieves a metric from the repository by type and name.
	//
	// Parameters:
	//   - ctx: The context for the operation.
	//   - metricType: The type of the metric (e.g., counter, gauge).
	//   - metricName: The name of the metric.
	//
	// Returns:
	//   - *entity.Metric: A pointer to the Metric if found.
	//   - error: An error if the metric is not found or another issue occurs.
	Find(ctx context.Context, metricType string, metricName string) (*entity.Metric, error)

	// All retrieves all metrics from the repository.
	//
	// Parameters:
	//   - ctx: The context for the operation.
	//
	// Returns:
	//   - *entity.Metrics: A pointer to a collection of all metrics.
	//   - error: An error if the operation fails.
	All(context.Context) (*entity.Metrics, error)

	// CheckConnection verifies the repository's connection.
	//
	// Parameters:
	//   - ctx: The context used to manage the connection check lifecycle.
	//
	// Returns:
	//   - error: An error if the connection check fails.
	CheckConnection(context.Context) error
}
