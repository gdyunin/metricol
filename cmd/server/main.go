package main

import (
	"github.com/gdyunin/metricol.git/internal/server/handlers"
	"github.com/gdyunin/metricol.git/internal/server/memstorage"
	"log"
	"net/http"
)

func main() {
	mux := http.NewServeMux()
	memStorage := memstorage.NewMemStorage()

	mux.Handle("/update/gauge/", http.StripPrefix("/update/gauge/", handlers.GaugeHandler(memStorage)))

	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		log.Fatal(err)
	}
}
