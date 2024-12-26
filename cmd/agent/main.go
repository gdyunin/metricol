package main

import (
	"log"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("Application terminated due to a critical error: %v", err)
	}
}
