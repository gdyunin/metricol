package main

import (
	"log"

	servercfg "github.com/gdyunin/metricol.git/internal/config/server"
	"github.com/gdyunin/metricol.git/internal/server"
)

func main() {
	// Retrieve the application configuration.
	appCfg := appConfig()
	// Create a default server with the application configuration.
	s := server.DefaultServer(appCfg)
	// Start the server and log any fatal errors that occur.
	log.Fatal(s.Start())
}

// appConfig retrieves and parses the application configuration.
// It logs a fatal error if the configuration cannot be parsed.
// Returns a pointer to the server configuration.
func appConfig() *servercfg.Config {
	appCfg, err := servercfg.ParseConfig()
	if err != nil {
		log.Fatalf("Get configuration fail: %v", err)
	}
	return appCfg
}
