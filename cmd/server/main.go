package main

import (
	"github.com/gdyunin/metricol.git/internal/server/handlers"
	"github.com/gdyunin/metricol.git/internal/server/storage"
	"github.com/go-chi/chi/v5"
	"log"
	"net/http"
)

func main() {
	warehouse := storage.NewWarehouse()
	router := chi.NewRouter()

	router.Get("/", handlers.MainPageHandler(warehouse))
	router.Get("/value/{metricType}/{metricName}", handlers.MetricGetHandler(warehouse))

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

	log.Fatal(http.ListenAndServe(":8080", router))
}
