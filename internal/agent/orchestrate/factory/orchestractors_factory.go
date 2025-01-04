package factory

import (
	"fmt"

	"github.com/gdyunin/metricol.git/internal/agent/collect"
	"github.com/gdyunin/metricol.git/internal/agent/orchestrate"
	"github.com/gdyunin/metricol.git/internal/agent/orchestrate/orchestrators/basic"
	"github.com/gdyunin/metricol.git/internal/agent/produce"
	"go.uber.org/zap"
)

const (
	// OrchestratorTypeBasic represents the type identifier for the basic orchestrator.
	OrchestratorTypeBasic = "basic"
)

// AbstractOrchestratorsFactory creates an abstract factory for orchestrators based on the specified type.
// Parameters:
//   - orchestratorType: The type of orchestrator to create (e.g., "basic").
//   - collector: A Collector instance responsible for gathering metrics.
//   - producer: A Producer instance responsible for processing and sending metrics.
//   - logger: Logger for logging activities and errors.
//
// Returns:
//   - An instance of orchestrate.OrchestratorAbstractFactory for the specified orchestrator type.
//   - An error if the orchestrator type is unsupported.
func AbstractOrchestratorsFactory(orchestratorType string, collector collect.Collector, producer produce.Producer, logger *zap.SugaredLogger) (orchestrate.OrchestratorAbstractFactory, error) {
	switch orchestratorType {
	case OrchestratorTypeBasic:
		return basic.NewOrchestratorFactory(collector, producer, logger), nil
	default:
		return nil, fmt.Errorf("unsupported orchestrator type: '%s', please provide a valid orchestrator type", orchestratorType)
	}
}
