package main

import (
	"log"

	"github.com/gdyunin/metricol.git/pkg/logger"
)

func main() {
	// Define the logging level for the application.
	logLvl := logger.LevelINFO

	// Run the application with the specified logging level.
	// If an error occurs, log a critical error message and terminate the application.
	if err := run(logLvl); err != nil {
		log.Fatalf("A critical error occurred while running the application with logging level %s: %v.", logLvl, err)
	}

	// Also log a critical error message if the application was stopped without errors.
	log.Fatalf("The application with logging level %s completed without errors.", logLvl)
}
