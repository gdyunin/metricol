package main

import (
	"sync"

	"github.com/labstack/gommon/log"
)

func main() {
	mainCtx, mainCtxCancel := mainContext()
	defer mainCtxCancel()

	logger := baseLogger()
	defer func() {
		if err := logger.Sync(); err != nil {
			log.Errorf("Zap logger sync error: %v", err)
		}
	}()

	appCfg, err := loadConfig()
	if err != nil {
		logger.Fatalf("Error occurred while parsing the application configuration: %v", err)
	}

	deliveryWithShutdownActs, err := initComponentsWithShutdownActs(appCfg, logger)
	if err != nil {
		logger.Fatalf("Error occurred while initialize the application components: %v", err)
	}

	setupGracefulShutdown(
		mainCtxCancel,
		logger.Named(loggerNameGracefulShutdown),
		deliveryWithShutdownActs.shutdownActions...,
	)

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		deliveryWithShutdownActs.server.Start(mainCtx)
	}()

	if appCfg.PprofFlag {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err = startProf(mainCtx, ":34659"); err != nil {
				logger.Fatalf("Profiling server error: %v", err)
			}
		}()
	}

	wg.Wait()
}
