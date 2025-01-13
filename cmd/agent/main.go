package main

func main() {
	mainCtx, mainCtxCancel := mainContext()
	defer mainCtxCancel()

	logger, loggerSync := loggerWithSyncFunc()
	defer loggerSync()

	setupGracefulShutdown(mainCtxCancel, logger.Named(LoggerNameGracefulShutdown))

	appCfg, err := loadConfig()
	if err != nil {
		logger.Fatalf("Error occurred while parsing the application configuration: %v", err)
	}

	metricsAgent := initComponents(appCfg, logger)

	metricsAgent.Start(mainCtx)
}
