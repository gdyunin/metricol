package consume

// Consumer represents an interface for a consumer that can process or consume data.
// It defines a method to start the consumption process.
type Consumer interface {
	// StartConsume begins the data consumption process.
	// It returns an error if the consumption process fails to start or encounters an issue.
	StartConsume() error
}

type ConsumerAbstractFactory interface {
	CreateConsumer() Consumer
}
