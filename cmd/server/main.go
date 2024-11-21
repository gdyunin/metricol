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

	mux.Handle("/update/gauge/", http.StripPrefix("/update/gauge/", handlers.GaugeHandler(&memStorage)))
	mux.Handle("/update/counter/", http.StripPrefix("/update/counter/", handlers.CounterHandler(&memStorage)))

	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		log.Fatal(err)
	}
}
