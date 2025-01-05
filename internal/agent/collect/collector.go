package collect

// Collector represents an interface for a metric collector.
// A collector is responsible for starting a data collection process.
type Collector interface {
	// StartCollect begins the collection of metrics.
	// Returns:
	//   - An error if the collection process fails to start or encounters an issue.
	StartCollect() error
}

// CollectorAbstractFactory defines an interface for creating instances of collectors.
// Factories encapsulate the creation logic for specific types of collectors.
type CollectorAbstractFactory interface {
	// CreateCollector creates and returns a new instance of a Collector.
	// Returns:
	//   - A Collector instance configured as per the factory`s implementation.
	CreateCollector() Collector
}
