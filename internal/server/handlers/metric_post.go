package handlers

import (
	"github.com/gdyunin/metricol.git/internal/server/memstorage"
	"github.com/gdyunin/metricol.git/internal/server/metrics"
	"github.com/gdyunin/metricol.git/internal/server/metrics/builder"
	"net/http"
)

func MetricPostHandler(memStorage memstorage.MemStorage) http.HandlerFunc {
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
		urlArgs := splitURL(r.URL.String(), 3)
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
		metric.SetName(metricName)
		if err := metric.SetValue(metricValue); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// Push metric to storage
		err = memStorage.PushMetric(metric)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// Response
		w.WriteHeader(http.StatusOK)
	}
}
