package main

import (
	"NewNewMetricol/internal/server/config"
	"NewNewMetricol/internal/server/delivery"
	"NewNewMetricol/internal/server/repository"

	"go.uber.org/zap"
)

func main() {
	appCfg, _ := config.ParseConfig()

	es := delivery.NewEchoServer(appCfg.ServerAddress, repository.NewInMemoryRepository(zap.NewExample().Sugar()), zap.NewExample().Sugar())
	es.Start()
}
