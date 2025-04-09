package main

import (
	"sync"

	"github.com/labstack/gommon/log"
)

func main() {
	printAppInfo()

	mainCtx, mainCtxCancel := mainContext()
	defer mainCtxCancel()

	logger := baseLogger()
	defer func() {
		if err := logger.Sync(); err != nil {
			log.Errorf("Zap logger sync error: %v", err)
		}
	}()

	setupGracefulShutdown(mainCtx, logger.Named(loggerNameGracefulShutdown))

	appCfg, err := loadConfig()
	if err != nil {
		logger.Fatalf("Error occurred while parsing the application configuration: %v", err)
	}

	metricsAgent := initAgent(appCfg, logger)

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		metricsAgent.Start(mainCtx)
	}()

	if appCfg.PprofFlag {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err = startProf(mainCtx, ":34658"); err != nil {
				logger.Fatalf("Profiling server error: %v", err)
			}
		}()
	}

	wg.Wait()
}
