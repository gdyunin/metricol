package collect

import (
	"fmt"

	"github.com/gdyunin/metricol.git/internal/agent/collect/collectors/mscollector/model"
	"github.com/gdyunin/metricol.git/internal/agent/entity"
)

// MemStatsCollectorAdapter is an adapter that facilitates storing metrics
// from the memory statistics collector into a metrics repository.
type MemStatsCollectorAdapter struct {
	metricInterface *entity.MetricsInterface // Interface for interacting with the metrics repository.
}

// NewMemStatsCollectorAdapter initializes and returns a new MemStatsCollectorAdapter.
// It sets up the adapter with a metrics interface backed by the provided repository.
//
// Parameters:
//   - repo: An instance of MetricsRepository used for storing metrics.
//
// Returns:
//   - A pointer to a fully initialized MemStatsCollectorAdapter.
func NewMemStatsCollectorAdapter(repo entity.MetricsRepository) *MemStatsCollectorAdapter {
	return &MemStatsCollectorAdapter{metricInterface: entity.NewMetricsInterface(repo)}
}

// Store saves a metric from the memory statistics collector into the repository.
// It converts the incoming metric from the memory statistics collector's format
// to a format compatible with the repository before saving.
//
// Parameters:
//   - metric: A pointer to a Metric instance from the memory statistics collector.
//
// Returns:
//   - An error if the storage operation fails; otherwise, nil.
func (a *MemStatsCollectorAdapter) Store(metric *model.Metric) error {
	// Convert the metric to the repository-compatible format.
	em := m2em(metric)

	// Store the converted metric in the repository and handle potential errors.
	if err := a.metricInterface.Store(em); err != nil {
		return fmt.Errorf("failed to store metric '%s': %w", metric.Name, err)
	}

	return nil
}

// m2em converts a metric from the memory statistics collector format to an entity metric format.
// This function ensures the metric is compatible with the repository's storage requirements.
//
// Parameters:
//   - m: A pointer to a Metric instance from the memory statistics collector.
//
// Returns:
//   - A pointer to a Metric instance formatted for storage in the repository.
func m2em(m *model.Metric) *entity.Metric {
	return &entity.Metric{
		Name:  m.Name,  // Name of the metric.
		Type:  m.Type,  // Type of the metric (e.g., gauge, counter).
		Value: m.Value, // Value of the metric.
	}
}
