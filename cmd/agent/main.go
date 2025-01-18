package main

func main() {
	mainCtx, mainCtxCancel := mainContext()
	defer mainCtxCancel()

	logger := baseLogger()
	defer func() { _ = logger.Sync() }()

	setupGracefulShutdown(mainCtxCancel, logger.Named(loggerNameGracefulShutdown))

	appCfg, err := loadConfig()
	if err != nil {
		logger.Fatalf("Error occurred while parsing the application configuration: %v", err)
	}

	metricsAgent := initComponents(appCfg, logger)

	metricsAgent.Start(mainCtx)
}
