package main

import (
	"log"
	"sync"

	"github.com/gdyunin/metricol.git/internal/agent"
	agentcfg "github.com/gdyunin/metricol.git/internal/config/agent"
)

func main() {
	var workGroup sync.WaitGroup
	workGroup.Add(1)

	// Retrieve the application configuration.
	appCfg := appConfig()
	// Create a default agent with the application configuration.
	a := agent.DefaultAgent(appCfg)

	// Start polling and reporting routines with the wait group.
	runRoutinesWithWG(
		&workGroup,
		func() {
			a.Polling(appCfg.PollInterval)
		},
		func() {
			a.Reporting(appCfg.ReportInterval)
		},
	)

	// Wait for all routines to finish before logging an error message.
	workGroup.Wait()
	log.Fatal("Polling or sending was interrupted, the agent was stopped.")
}

// appConfig retrieves and parses the application configuration.
// It logs a fatal error if the configuration cannot be parsed.
func appConfig() *agentcfg.Config {
	appCfg, err := agentcfg.ParseConfig()
	if err != nil {
		log.Fatalf("Get configuration fail: %v", err)
	}
	return appCfg
}

// runRoutinesWithWG starts multiple functions as goroutines and waits for them to complete.
// Each function is executed in its own goroutine, and the wait group is used to ensure
// that the main function waits for all routines to finish before proceeding.
func runRoutinesWithWG(wg *sync.WaitGroup, fn ...func()) {
	for _, f := range fn {
		go func(f func()) {
			defer wg.Done()
			f()
		}(f) // Pass f as an argument to avoid closure issues
	}
}
