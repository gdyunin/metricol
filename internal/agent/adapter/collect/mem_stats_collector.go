package collect

import (
	"fmt"

	"github.com/gdyunin/metricol.git/internal/agent/collect/collectors/mscollector/model"
	"github.com/gdyunin/metricol.git/internal/agent/entity"
)

// MemStatsCollectorAdapter is an adapter that provides methods to store metrics
// from the memory statistics collector into a metrics repository.
type MemStatsCollectorAdapter struct {
	metricInterface *entity.MetricsInterface
}

// NewMemStatsCollectorAdapter creates a new instance of MemStatsCollectorAdapter.
// It initializes the adapter with a metrics interface using the provided repository.
//
// Parameters:
//   - repo: An instance of MetricsRepository used for storing metrics.
//
// Returns:
//   - A pointer to an initialized MemStatsCollectorAdapter.
func NewMemStatsCollectorAdapter(repo entity.MetricsRepository) *MemStatsCollectorAdapter {
	return &MemStatsCollectorAdapter{metricInterface: entity.NewMetricsInterface(repo)}
}

// Store saves the provided metric into the repository.
// It converts the incoming metric from the memory statistics collector's format
// to the entity format before storing it.
//
// Parameters:
//   - metric: A pointer to a Metric instance from the memory statistics collector.
//
// Returns:
//   - An error, if the storage operation fails. Otherwise, nil.
func (a *MemStatsCollectorAdapter) Store(metric *model.Metric) error {
	em := m2em(metric)

	// Store the converted metric entity into the repository.
	if err := a.metricInterface.Store(em); err != nil {
		return fmt.Errorf("failed to store metric '%s': %w", metric.Name, err)
	}

	return nil
}

// m2em converts a memory statistics collector metric into an entity metric.
// This transformation ensures compatibility with the storage layer.
//
// Parameters:
//   - m: A pointer to a Metric instance from the memory statistics collector.
//
// Returns:
//   - A pointer to an entity Metric instance.
func m2em(m *model.Metric) *entity.Metric {
	return &entity.Metric{
		Name:  m.Name,
		Type:  m.Type,
		Value: m.Value,
	}
}
