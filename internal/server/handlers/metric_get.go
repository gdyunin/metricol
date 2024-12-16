/*
Package handlers provides HTTP handler functions for managing metrics.

This package includes handlers for sending error responses,
displaying metrics, and interacting with a storage repository.
*/
package handlers

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/gdyunin/metricol.git/internal/common/models"
	"github.com/gdyunin/metricol.git/internal/metrics"
	"github.com/gdyunin/metricol.git/internal/server/storage"
	"github.com/go-chi/chi/v5"
)

// MetricGetFromURIHandler returns an HTTP handler function that retrieves a metric value
// based on the specified metric type and name from the provided repository.
func MetricGetFromURIHandler(repository storage.Repository) http.HandlerFunc {
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

		// Response with response code 200 OK.
		w.WriteHeader(http.StatusOK)
	}
}

func MetricGetFromBodyHandler(repository storage.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Set Content-Type for the response.
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		// Проверяем контент тайп
		if r.Header.Get("Content-Type") != "application/json" {
			BadRequest(w, r) // если не JSON, то сливаем
			return
		}

		metric := models.Metrics{} // сюда парсим
		dec := json.NewDecoder(r.Body)
		if err := dec.Decode(&metric); err != nil {
			InternalServerError(w, r)
			return
		}
		defer func() { _ = r.Body.Close() }()

		// Не пустое имя
		if metric.ID == "" {
			BadRequest(w, r)
			return
		}

		metricValue, _ := repository.GetMetric(metric.ID, metric.MType)
		var err error
		// Подходящий тип
		switch metric.MType {
		case metrics.MetricTypeCounter:
			mv, _ := strconv.Atoi(metricValue)
			mv1 := int64(mv)
			metric.Delta = &mv1
		case metrics.MetricTypeGauge:
			mv, _ := strconv.ParseFloat(metricValue, 64)
			metric.Value = &mv
		default:
			BadRequest(w, r)
			return
		}

		enc := json.NewEncoder(w)
		err = enc.Encode(metric)
		if err != nil {
			InternalServerError(w, r)
			return
		}

		// Response with response code 200 OK.
		w.WriteHeader(http.StatusOK)
	}
}
