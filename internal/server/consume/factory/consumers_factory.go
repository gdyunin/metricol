package factory

import (
	"fmt"

	"github.com/gdyunin/metricol.git/internal/server/consume"
	"github.com/gdyunin/metricol.git/internal/server/consume/consumers/echohttp"
	"github.com/gdyunin/metricol.git/internal/server/entities"
	"go.uber.org/zap"
)

const (
	// ConsumerTypeEchoServer specifies the type for an HTTP echo server consumer.
	ConsumerTypeEchoServer = "echo http server"
)

// AbstractConsumerFactory creates and returns a factory for a specified consumer type.
//
// Parameters:
//   - consumerType: A string specifying the type of consumer to create (e.g., "echo http server").
//   - addr: The address where the consumer should listen or operate.
//   - repo: The metrics repository that the consumer interacts with.
//   - logger: A logger instance for capturing runtime information and errors.
//
// Returns:
//   - A ConsumerAbstractFactory for the specified consumer type.
//   - An error if the consumer type is unsupported or if any other issue occurs.
func AbstractConsumerFactory(consumerType string, addr string, repo entities.MetricsRepository, logger *zap.SugaredLogger) (consume.ConsumerAbstractFactory, error) {
	switch consumerType {
	case ConsumerTypeEchoServer:
		return echohttp.NewEchoServerConsumerFactory(addr, repo, logger), nil
	default:
		return nil, fmt.Errorf("unsupported consumer type: '%s', please provide a valid consumer type", consumerType)
	}
}
