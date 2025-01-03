package collectors

import (
	"fmt"

	"github.com/gdyunin/metricol.git/internal/agent/collect/collectors/memstat/model"
	"github.com/gdyunin/metricol.git/internal/agent/entities"
)

// MemStatsCollectorAdapter is an adapters that facilitates storing metrics
// from the memory statistics collectors into a metrics repository.
type MemStatsCollectorAdapter struct {
	metricInterface *entities.MetricsInterface // Interface for interacting with the metrics repository.
}

// NewMemStatsCollectorAdapter initializes and returns a new MemStatsCollectorAdapter.
// It sets up the adapters with a metrics interface backed by the provided repository.
//
// Parameters:
//   - repo: An instance of MetricsRepository used for storing metrics.
//
// Returns:
//   - A pointer to a fully initialized MemStatsCollectorAdapter.
func NewMemStatsCollectorAdapter(repo entities.MetricsRepository) *MemStatsCollectorAdapter {
	return &MemStatsCollectorAdapter{metricInterface: entities.NewMetricsInterface(repo)}
}

// Store saves a metric from the memory statistics collectors into the repository.
// It converts the incoming metric from the memory statistics collectors`s format
// to a format compatible with the repository before saving.
//
// Parameters:
//   - metric: A pointer to a Metric instance from the memory statistics collectors.
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

// m2em converts a metric from the memory statistics collectors format to an entities metric format.
// This function ensures the metric is compatible with the repository's storage requirements.
//
// Parameters:
//   - m: A pointer to a Metric instance from the memory statistics collectors.
//
// Returns:
//   - A pointer to a Metric instance formatted for storage in the repository.
func m2em(m *model.Metric) *entities.Metric {
	return &entities.Metric{
		Name:  m.Name,  // Name of the metric.
		Type:  m.Type,  // Type of the metric (e.g., gauge, counter).
		Value: m.Value, // Value of the metric.
	}
}
