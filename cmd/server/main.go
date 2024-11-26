package main

import (
	"github.com/gdyunin/metricol.git/internal/server/handlers"
	"github.com/gdyunin/metricol.git/internal/server/storage"
	"github.com/go-chi/chi/v5"
	"log"
	"net/http"
)

func main() {
	// Get config
	appCfg := appConfig()

	// Create structures
	warehouse := storage.NewWarehouse()
	router := chi.NewRouter()

	// Setup GET methods
	router.Get("/", handlers.MainPageHandler(warehouse))
	router.Get("/value/{metricType}/{metricName}", handlers.MetricGetHandler(warehouse))

	// Setup POST methods
	router.Route("/update/", func(r chi.Router) {
		r.Post("/", handlers.BadRequest)
		r.Route("/{metricType}", func(r chi.Router) {
			r.Post("/", handlers.NotFound)
			r.Route("/{metricName}", func(r chi.Router) {
				r.Post("/", handlers.BadRequest)
				r.Post("/{metricValue}", handlers.MetricPostHandler(warehouse))
			})
		})
	})

	// Start server
	log.Fatal(http.ListenAndServe(appCfg.ServerAddress, router))
}
