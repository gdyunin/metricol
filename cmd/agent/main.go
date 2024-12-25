package main

import (
	"log"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("Critical error encountered during application execution: %v", err)
	}
	log.Fatal("Application stopped")
}
