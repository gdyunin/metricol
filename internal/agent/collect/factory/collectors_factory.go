package factory

import (
	"fmt"
	"time"

	"github.com/gdyunin/metricol.git/internal/agent/collect"
	"github.com/gdyunin/metricol.git/internal/agent/collect/collectors/mscollector"
	"github.com/gdyunin/metricol.git/internal/agent/entities"
	"go.uber.org/zap"
)

const (
	CollectorTypeMemStats = "memory stats"
)

func AbstractCollectorFactory(collectorType string, interval time.Duration, repo entities.MetricsRepository, logger *zap.SugaredLogger) (collect.CollectorAbstractFactory, error) {
	switch collectorType {
	case CollectorTypeMemStats:
		return mscollector.NewMemStatsCollectorFactory(interval, repo, logger), nil
	default:
		return nil, fmt.Errorf("unsupported collector type: %s", collectorType)
	}
}
