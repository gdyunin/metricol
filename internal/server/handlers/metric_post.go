package handlers

import (
	"github.com/gdyunin/metricol.git/internal/server/metrics"
	"github.com/gdyunin/metricol.git/internal/server/metrics/builder"
	"github.com/gdyunin/metricol.git/internal/server/storage"
	"github.com/go-chi/chi/v5"
	"net/http"
)

func MetricPostHandler(repository storage.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Check Content-Type
		if r.Header.Get("Content-Type") != "text/plain" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// Parse params from URL
		metricType := chi.URLParam(r, "metricType")
		metricName := chi.URLParam(r, "metricName")
		metricValue := chi.URLParam(r, "metricValue")

		// Create metric
		metric, err := builder.NewMetric(metrics.MetricType(metricType))
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// Fill metric
		if err := metric.SetName(metricName); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if err := metric.SetValue(metricValue); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// Push metric to storage
		err = repository.PushMetric(metric)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// Headers
		h := w.Header()
		h.Set("Content-Type", "text/plain")

		// Response
		w.WriteHeader(http.StatusOK)
	}
}
