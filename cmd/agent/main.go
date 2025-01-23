package main

import "github.com/labstack/gommon/log"

func main() {
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

	metricsAgent.Start(mainCtx)
}
