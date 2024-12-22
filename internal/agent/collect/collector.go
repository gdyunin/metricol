package collect

// Collector represents a data collection interface.
// Implementations of this interface are responsible for starting the data collection process.
type Collector interface {
	// StartCollect initiates the data collection process.
	//
	// Returns:
	//   - An error if the data collection process fails.
	StartCollect() error
}
