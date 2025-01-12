package main

import (
	"NewNewMetricol/internal/server/config"
	"NewNewMetricol/internal/server/delivery"
	"NewNewMetricol/internal/server/repository"
	"NewNewMetricol/pkg/convert"

	"go.uber.org/zap"
)

func main() {
	appCfg, _ := config.ParseConfig()

	var repo repository.Repository
	r, err := repository.NewFileStorageRepository(zap.NewExample().Sugar(), appCfg.FileStoragePath, "backup.txt", convert.IntegerToSeconds(appCfg.StoreInterval), appCfg.Restore)
	if err != nil {
		repo = r
	} else {
		repo = repository.NewInMemoryRepository(zap.NewExample().Sugar())
	}

	es := delivery.NewEchoServer(appCfg.ServerAddress, repo, zap.NewExample().Sugar())
	es.Start()
}
