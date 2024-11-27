package handlers

import (
	"net/http"

	"github.com/gdyunin/metricol.git/internal/metrics"
	"github.com/gdyunin/metricol.git/internal/server/storage"
	"github.com/go-chi/chi/v5"
)

// MetricGetHandler return a handler that get metric value by got metric type and name.
func MetricGetHandler(repository storage.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Set Content-Type.
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")

		// Parse params from URL.
		metricType := chi.URLParam(r, "metricType")
		metricName := chi.URLParam(r, "metricName")

		// Get metric.
		metricValue, err := repository.GetMetric(metricName, metricType)
		if err != nil {
			switch err.Error() {
			case metrics.ErrorUnknownMetricType:
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			case storage.ErrorUnknownMetricName:
				http.Error(w, err.Error(), http.StatusNotFound)
				return
			}
		}

		// The error is ignored as it has no effect.
		// A logger could be added in the future.
		_, _ = w.Write([]byte(metricValue))
	}
}
