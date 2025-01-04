package factory

import (
	"fmt"
	"time"

	"github.com/gdyunin/metricol.git/internal/agent/collect"
	"github.com/gdyunin/metricol.git/internal/agent/collect/collectors/memstat"
	"github.com/gdyunin/metricol.git/internal/agent/entities"
	"go.uber.org/zap"
)

const (
	// CollectorTypeMemStats represents the type identifier for a memory statistics collector.
	CollectorTypeMemStats = "memory stats"
)

// AbstractCollectorFactory creates an abstract collector factory based on the specified collector type.
// Parameters:
//   - collectorType: The type of collector to create (e.g., "memory stats").
//   - interval: The time interval for data collection.
//   - repo: The metrics repository where collected data will be stored.
//   - logger: Logger for logging activities and errors.
//
// Returns:
//   - A collect.CollectorAbstractFactory for the specified collector type.
//   - An error if the collector type is unsupported.
func AbstractCollectorFactory(collectorType string, interval time.Duration, repo entities.MetricsRepository, logger *zap.SugaredLogger) (collect.CollectorAbstractFactory, error) {
	switch collectorType {
	case CollectorTypeMemStats:
		return memstat.NewMemStatsCollectorFactory(interval, repo, logger), nil
	default:
		return nil, fmt.Errorf("unsupported collector type: '%s', please provide a valid collector type", collectorType)
	}
}
