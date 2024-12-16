/*
Package handlers provides HTTP handler functions for managing metrics.

This package includes handlers for sending error responses,
displaying metrics, and interacting with a storage repository.
*/
package handlers

import (
	"log"
	"net/http"

	"github.com/gdyunin/metricol.git/internal/metrics"
	"github.com/gdyunin/metricol.git/internal/server/storage"
	"github.com/go-chi/chi/v5"
)

// MetricPostHandler returns an HTTP handler function that pushes a metric
// to the specified repository.
func MetricPostHandler(repository storage.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Set Content-Type for the response.
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")

		// Parse parameters from the URL.
		metricType := chi.URLParam(r, "metricType")
		metricName := chi.URLParam(r, "metricName")
		metricValue := chi.URLParam(r, "metricValue")

		// Create a new metric from the parsed parameters.
		metric, err := metrics.NewFromStrings(metricName, metricValue, metricType)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		// Push the created metric to the storage repository.
		err = repository.PushMetric(metric)
		if err != nil {
			log.Printf("error pushing metric to store: %v", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		// Response with response code 200 OK.
		w.WriteHeader(http.StatusOK)
	}
}
