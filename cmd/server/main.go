package main

import (
	"github.com/gdyunin/metricol.git/internal/server/handlers"
	"github.com/gdyunin/metricol.git/internal/server/memstorage"
	"log"
	"net/http"
)

func main() {
	mux := http.NewServeMux()
	memStorage := memstorage.NewBaseMemStorage()

	mux.Handle("/update/", http.StripPrefix("/update/", handlers.MetricPostHandler(&memStorage)))

	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		log.Fatal(err)
	}
}
