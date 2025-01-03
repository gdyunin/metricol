package orchestrate

import (
	"fmt"

	"github.com/gdyunin/metricol.git/internal/agent/collect"
	"github.com/gdyunin/metricol.git/internal/agent/orchestrate/orchestrators/basic"
	"github.com/gdyunin/metricol.git/internal/agent/produce"
	"go.uber.org/zap"
)

const (
	OrchestratorTypeBasic = "basic"
)

func AbstractOrchestratorsFactory(orchestratorType string, collector collect.Collector, producer produce.Producer, logger *zap.SugaredLogger) (OrchestratorAbstractFactory, error) {
	switch orchestratorType {
	case OrchestratorTypeBasic:
		return basic.NewOrchestratorFactory(collector, producer, logger), nil
	default:
		return nil, fmt.Errorf("unsupported orchestrator type: %s", orchestratorType)
	}
}
