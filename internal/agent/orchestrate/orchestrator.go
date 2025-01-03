package orchestrate

type Orchestrator interface {
	StartAll() error
}

type OrchestratorAbstractFactory interface {
	CreateOrchestrator() Orchestrator
}
