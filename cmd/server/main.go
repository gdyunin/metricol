package main

func main() {
	mainCtx, mainCtxCancel := mainContext()
	defer mainCtxCancel()

	logger, loggerSync := loggerWithSyncFunc()
	defer loggerSync()

	appCfg := loadConfig()

	server, shutdownActs := initComponentsWithShutdownActs(appCfg, logger)
	setupGracefulShutdown(mainCtxCancel, logger.Named(LoggerNameGracefulShutdown), shutdownActs...)

	server.Start(mainCtx)
}
