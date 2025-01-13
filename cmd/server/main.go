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

	server, shutdownActs := initComponentsWithShutdownActs(appCfg, logger)
	setupGracefulShutdown(mainCtxCancel, logger.Named(LoggerNameGracefulShutdown), shutdownActs...)

	server.Start(mainCtx)
}
