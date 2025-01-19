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
	deliveryWithShutdownActs.server.Start(mainCtx)
}
