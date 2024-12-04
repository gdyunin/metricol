package main

import (
	"log"
	"sync"

	"github.com/gdyunin/metricol.git/internal/agent"
	agentcfg "github.com/gdyunin/metricol.git/internal/config/agent"
)

// main is the entry point of the application.
// It initializes the agent configuration and starts the polling and reporting routines.
func main() {
	var workGroup sync.WaitGroup
	workGroup.Add(1) // If at least one goroutine ends, stop the application.

	// Parse the application configuration.
	appCfg, err := agentcfg.ParseConfig()
	if err != nil {
		log.Fatalf("Get configuration fail: %v", err)
	}

	// Create a new instance of the agent with the parsed configuration.
	a := agent.DefaultAgent(appCfg)

	// Start the polling in a separate goroutine.
	go func() {
		defer workGroup.Done()
		a.Polling(appCfg.PollInterval)
	}()

	// Start the reporting in a separate goroutine.
	go func() {
		defer workGroup.Done()
		a.Reporting(appCfg.ReportInterval)
	}()

	// Wait for at least one goroutine to complete.
	workGroup.Wait()
	// Application stopped, log it and exit,
	log.Fatal("Polling or sending was interrupted, the agent was stopped.")
}
