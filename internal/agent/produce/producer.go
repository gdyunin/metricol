package produce

// Producer defines an interface for starting a production process.
//
// This interface can be implemented by any component that produces data or
// initiates a process requiring production.
type Producer interface {
	// StartProduce begins the production process.
	//
	// Returns:
	//   - An error if the production process fails to start or encounters issues.
	StartProduce() error
}

type ProducerAbstractFactory interface {
	CreateProducer() Producer
}
