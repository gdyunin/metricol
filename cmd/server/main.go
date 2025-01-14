package main

func main() {
	mainCtx, mainCtxCancel := mainContext()
	defer mainCtxCancel()

	logger, loggerSync := loggerWithSyncFunc()
	defer loggerSync()

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
		logger.Named(LoggerNameGracefulShutdown),
		deliveryWithShutdownActs.shutdownActions...,
	)
	deliveryWithShutdownActs.server.Start(mainCtx)
}
