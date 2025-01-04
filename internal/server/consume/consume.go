package consume

// Consumer represents an interface for a consumer that can process or consume data.
// It provides a method to initiate the consumption process.
type Consumer interface {
	// StartConsume begins the data consumption process.
	//
	// Returns:
	//   - An error if the consumption process fails to start or encounters an issue during execution.
	StartConsume() error
}

// ConsumerAbstractFactory defines an interface for creating instances of Consumer.
// Implementations of this interface should provide a method to create and return a Consumer instance.
type ConsumerAbstractFactory interface {
	// CreateConsumer creates and returns a new instance of a Consumer.
	//
	// Returns:
	//   - A Consumer instance that implements the data consumption process.
	CreateConsumer() Consumer
}
