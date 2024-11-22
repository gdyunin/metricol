package handlers

import (
	"github.com/gdyunin/metricol.git/internal/server/metrics"
	"github.com/gdyunin/metricol.git/internal/server/metrics/builder"
	"github.com/gdyunin/metricol.git/internal/server/storage"
	"net/http"
)

func MetricPostHandler(repository storage.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Allow only POST
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		// Check Content-Type
		if r.Header.Get("Content-Type") != "text/plain" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// Parse params from URL
		urlArgs := splitURI(r.URL.String(), 3)
		metricType, metricName, metricValue := urlArgs[0], urlArgs[1], urlArgs[2]

		// Check exists require params
		if metricType == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if metricName == "" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if metricValue == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

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
