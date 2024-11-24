package handlers

import (
	"github.com/gdyunin/metricol.git/internal/server/metrics"
	"github.com/gdyunin/metricol.git/internal/server/storage"
	"github.com/go-chi/chi/v5"
	"net/http"
)

func MetricGetHandler(repository storage.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Check Content-Type
		if r.Header.Get("Content-Type") != "text/plain" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// Parse params from URL
		metricType := chi.URLParam(r, "metricType")
		metricName := chi.URLParam(r, "metricName")

		metricValue, ok := repository.Metrics()[metrics.MetricType(metricType)][metricName]
		if !ok {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		_, _ = w.Write([]byte(metricValue))

		// Headers
		h := w.Header()
		h.Set("Content-Type", "text/plain")
	}
}
