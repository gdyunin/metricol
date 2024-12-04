package main

import (
	"log"

	servercfg "github.com/gdyunin/metricol.git/internal/config/server"
	"github.com/gdyunin/metricol.git/internal/server"
)

func main() {
	// Parse the application configuration.
	appCfg, err := servercfg.ParseConfig()
	if err != nil {
		log.Fatalf("Get configuration fail: %v", err)
	}

	// Create a default server instance using the parsed configuration.
	s := server.DefaultServer(appCfg)

	// Start the server and log any fatal errors that occur.
	log.Fatal(s.Start())
}
