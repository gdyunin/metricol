package orchestrate

// Orchestrator represents an interface for managing and coordinating collectors and producers.
// It provides methods to start all orchestrated processes.
type Orchestrator interface {
	// StartAll starts all orchestrated processes managed by the orchestrator.
	// Returns:
	//   - An error if any of the processes fail to start.
	StartAll() error
}

// OrchestratorAbstractFactory defines an interface for creating instances of orchestrators.
// Factories encapsulate the creation logic for specific types of orchestrators.
type OrchestratorAbstractFactory interface {
	// CreateOrchestrator creates and returns a new instance of an Orchestrator.
	// Returns:
	//   - An Orchestrator instance configured as per the factory's implementation.
	CreateOrchestrator() Orchestrator
}
