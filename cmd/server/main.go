package main

import (
	"github.com/gdyunin/metricol.git/internal/server/handlers"
	"github.com/gdyunin/metricol.git/internal/server/storage"
	"log"
	"net/http"
)

func main() {
	warehouse := storage.NewWarehouse()

	mux := http.NewServeMux()
	mux.Handle("/update/", http.StripPrefix("/update/", handlers.MetricPostHandler(warehouse)))

	log.Fatal(http.ListenAndServe(":8080", mux))
}
