// Package factory provides an abstract factory for creating producers.
// Producers are responsible for sending collected metrics to a target destination.
package factory

import (
	"fmt"
	"time"

	"github.com/gdyunin/metricol.git/internal/agent/entities"
	"github.com/gdyunin/metricol.git/internal/agent/produce"
	"github.com/gdyunin/metricol.git/internal/agent/produce/produsers/httpresty"
	"go.uber.org/zap"
)

const (
	// ProducerTypeRestyClient represents the type identifier for a producer that uses the Resty HTTP client.
	ProducerTypeRestyClient = "resty http client"
)

// AbstractProducerFactory creates an abstract producer factory based on the specified producer type.
// Parameters:
//   - producerType: The type of producer to create (e.g., "memory stats").
//   - interval: The interval at which metrics are produced.
//   - serverAddress: The address of the server where metrics will be sent.
//   - repo: The metrics repository containing the data to be produced.
//   - logger: Logger for logging activities and errors.
//
// Returns:
//   - A produce.ProducerAbstractFactory for the specified producer type.
//   - An error if the producer type is unsupported.
func AbstractProducerFactory(producerType string, interval time.Duration, serverAddress string, repo entities.MetricsRepository, logger *zap.SugaredLogger) (produce.ProducerAbstractFactory, error) {
	switch producerType {
	case ProducerTypeRestyClient:
		return httpresty.NewRestyClientProducerFactory(interval, serverAddress, repo, logger), nil
	default:
		return nil, fmt.Errorf("unsupported producer type: '%s', please provide a valid producer type", producerType)
	}
}
