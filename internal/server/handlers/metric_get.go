/*
Package handlers provides HTTP handler functions for managing metrics.

This package includes handlers for sending error responses,
displaying metrics, and interacting with a storage repository.
*/
package handlers

import (
	"errors"
	"log"
	"net/http"

	"github.com/gdyunin/metricol.git/internal/server/storage"
	"github.com/go-chi/chi/v5"
)

// MetricGetHandler returns an HTTP handler function that retrieves a metric value
// based on the specified metric type and name from the provided repository.
func MetricGetHandler(repository storage.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Set Content-Type for the response.
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")

		// Parse metric type and name from URL parameters.
		metricType := chi.URLParam(r, "metricType")
		metricName := chi.URLParam(r, "metricName")

		// Retrieve the metric value from the repository.
		metricValue, err := repository.GetMetric(metricName, metricType)
		if err != nil {
			var statusCode int
			if errors.Is(err, storage.ErrUnknownMetricType) {
				statusCode = http.StatusBadRequest // Invalid metric type.
			} else {
				statusCode = http.StatusNotFound // Metric not found.
			}

			http.Error(w, http.StatusText(statusCode), statusCode)
			return
		}

		// Write the metric value to the response.
		_, err = w.Write([]byte(metricValue))
		if err != nil {
			log.Printf("error getting metric %s %s: %v", metricName, metricType, err)
		}
	}
}
