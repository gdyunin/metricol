package factory

import (
	"fmt"
	"time"

	"github.com/gdyunin/metricol.git/internal/agent/entities"
	"github.com/gdyunin/metricol.git/internal/agent/produce"
	"github.com/gdyunin/metricol.git/internal/agent/produce/produsers/rstclient"
	"go.uber.org/zap"
)

const (
	ProducerTypeRestyClient = "memory stats"
)

func AbstractProducerFactory(producerType string, interval time.Duration, serverAddress string, repo entities.MetricsRepository, logger *zap.SugaredLogger) (produce.ProducerAbstractFactory, error) {
	switch producerType {
	case ProducerTypeRestyClient:
		return rstclient.NewRestyClientProducerFactory(interval, serverAddress, repo, logger), nil
	default:
		return nil, fmt.Errorf("unsupported collector type: %s", producerType)
	}
}
