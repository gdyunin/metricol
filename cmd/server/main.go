package main

import (
	"NewNewMetricol/internal/server/config"
	"NewNewMetricol/internal/server/delivery"
	"NewNewMetricol/internal/server/repository"
	"NewNewMetricol/pkg/convert"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"
)

func main() {
	appCfg, _ := config.ParseConfig()

	var repo repository.Repository
	r, err := repository.NewFileStorageRepository(zap.NewExample().Sugar(), appCfg.FileStoragePath, "backup.txt", convert.IntegerToSeconds(appCfg.StoreInterval), appCfg.Restore)
	if err == nil {
		repo = r
		go func() {
			stopChan := make(chan os.Signal, 1)                    // Create a channel to receive OS signals.
			signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM) // Subscribe to interrupt and terminate signals.
			<-stopChan                                             // Wait for a termination signal.
			r.Shutdown()
			os.Exit(0)
		}()
	} else {
		repo = repository.NewInMemoryRepository(zap.NewExample().Sugar())
	}

	es := delivery.NewEchoServer(appCfg.ServerAddress, repo, zap.NewExample().Sugar())
	es.Start()
}
