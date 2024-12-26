package collect

// Collector represents a data collection interface.
// This interface defines the methods that a data collector must implement.
// It serves as an abstraction for starting data collection processes.
type Collector interface {
	// StartCollect initiates the data collection process.
	//
	// Returns:
	//   - An error if the data collection process fails to start or encounters issues.
	StartCollect() error
}
