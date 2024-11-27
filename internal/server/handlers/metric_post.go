package handlers

import (
	"net/http"

	"github.com/gdyunin/metricol.git/internal/metrics"
	"github.com/gdyunin/metricol.git/internal/server/storage"
	"github.com/go-chi/chi/v5"
)

// MetricPostHandler return a handler that push metric to repository.
func MetricPostHandler(repository storage.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Set Content-Type.
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")

		// Parse params from URL.
		metricType := chi.URLParam(r, "metricType")
		metricName := chi.URLParam(r, "metricName")
		metricValue := chi.URLParam(r, "metricValue")

		// Create metric.
		metric, err := metrics.NewFromStrings(metricName, metricValue, metricType)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Push metric to storage.
		err = repository.PushMetric(metric)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}
