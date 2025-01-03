package produce

// Producer defines an interface for starting a production process.
//
// This interface can be implemented by any component that produces data or
// initiates a process requiring production.
type Producer interface {
	// StartProduce begins the production process.
	//
	// Returns:
	//   - An error if the production process fails to start or encounters any issues.
	StartProduce() error
}

// ProducerAbstractFactory defines an interface for creating instances of producers.
//
// Factories encapsulate the logic for constructing specific producer implementations.
type ProducerAbstractFactory interface {
	// CreateProducer creates and returns a new instance of a Producer.
	//
	// Returns:
	//   - A Producer instance configured according to the factory's implementation.
	CreateProducer() Producer
}
