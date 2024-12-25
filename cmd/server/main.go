package main

import (
	"log"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("Application encountered a critical error: %v", err)
	}
	log.Fatal("Application stopped")
}
