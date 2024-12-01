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
	// Get ServerConfig.
	appCfg := server.ParseServerConfig()

	// Create structures.
	store := storage.NewStore()
	router := chi.NewRouter()

	// Setup router.
	setupRouter(router, store)

	// Start server.
	log.Fatal(http.ListenAndServe(appCfg.ServerAddress, router))
}

// setupRouter configure chi.Router with got storage.Repository.
func setupRouter(router chi.Router, store storage.Repository) {
	// Setup GET methods.
	router.Get("/", handlers.MainPageHandler(store))
	router.Get("/value/{metricType}/{metricName}", handlers.MetricGetHandler(store))

	// Setup POST methods
	router.Route("/update/", func(r chi.Router) {
		r.Post("/", handlers.BadRequest)
		r.Route("/{metricType}", func(r chi.Router) {
			r.Post("/", handlers.NotFound)
			r.Route("/{metricName}", func(r chi.Router) {
				r.Post("/", handlers.BadRequest)
				r.Post("/{metricValue}", handlers.MetricPostHandler(store))
			})
		})
	})
}
