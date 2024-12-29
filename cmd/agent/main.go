package main

import "github.com/gdyunin/metricol.git/pkg/logger"

func main() {
	appRunner := newRunner(logger.LevelINFO)
	appRunner.run()
}
