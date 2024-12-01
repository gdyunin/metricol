package main

import (
	"log"
	"net/http"

	"github.com/gdyunin/metricol.git/internal/config/server"
	"github.com/gdyunin/metricol.git/internal/server/handlers"
	"github.com/gdyunin/metricol.git/internal/server/storage"
	"github.com/go-chi/chi/v5"
)

func main() {
	// Get application configuration.
	appCfg := server.ParseConfig()

	// Create storage and router instances.
	store := storage.NewStore()
	router := chi.NewRouter()

	// Set up the router with appropriate handlers.
	setupRouter(router, store)

	// Start the HTTP server with the specified address and router.
	log.Fatal(http.ListenAndServe(appCfg.ServerAddress, router))
}

// setupRouter configures the chi.Router with the provided storage.Repository.
func setupRouter(router chi.Router, store storage.Repository) {
	// Setup GET methods for retrieving metrics.
	router.Get("/", handlers.MainPageHandler(store))
	router.Get("/value/{metricType}/{metricName}", handlers.MetricGetHandler(store))

	// Setup POST methods for updating metrics.
	router.Route("/update/", func(r chi.Router) {
		r.Post("/", handlers.BadRequest) // Metric type not passed.
		r.Route("/{metricType}", func(r chi.Router) {
			r.Post("/", handlers.NotFound) // Metric name not passed.
			r.Route("/{metricName}", func(r chi.Router) {
				r.Post("/", handlers.BadRequest)                            // Metric value not passed.
				r.Post("/{metricValue}", handlers.MetricPostHandler(store)) // Handle metric post query.
			})
		})
	})
}
