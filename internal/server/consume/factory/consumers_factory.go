package factory

import (
	"fmt"

	"github.com/gdyunin/metricol.git/internal/server/consume"
	"github.com/gdyunin/metricol.git/internal/server/consume/consumers/echohttp"
	"github.com/gdyunin/metricol.git/internal/server/entities"
	"go.uber.org/zap"
)

const (
	ConsumerTypeEchoServer = "echo http server"
)

func AbstractConsumerFactory(consumerType string, addr string, repo entities.MetricsRepository, logger *zap.SugaredLogger) (consume.ConsumerAbstractFactory, error) {
	switch consumerType {
	case ConsumerTypeEchoServer:
		return echohttp.NewEchoServerConsumerFactory(addr, repo, logger), nil
	default:
		return nil, fmt.Errorf("unsupported consumer type: '%s', please provide a valid producer type", consumerType)
	}
}
