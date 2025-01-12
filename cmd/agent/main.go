package main

func main() {
	mainCtx, mainCtxCancel := mainContext()
	defer mainCtxCancel()

	logger, loggerSync := loggerWithSyncFunc()
	defer loggerSync()

	setupGracefulShutdown(mainCtxCancel, logger.Named(LoggerNameGracefulShutdown))

	appCfg := loadConfig()
	metricsAgent := initComponents(appCfg, logger)

	metricsAgent.Start(mainCtx)
}
