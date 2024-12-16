/*
Package handlers provides HTTP handler functions for managing metrics.

This package includes handlers for sending error responses,
displaying metrics, and interacting with a storage repository.
*/
package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/gdyunin/metricol.git/internal/common/models"
	"github.com/gdyunin/metricol.git/internal/metrics"
	"github.com/gdyunin/metricol.git/internal/server/storage"
	"github.com/go-chi/chi/v5"
)

// MetricPostFromURIHandler returns an HTTP handler function that pushes a metric
// to the specified repository.
func MetricPostFromURIHandler(repository storage.Repository) http.HandlerFunc {
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

// MetricPostFromBodyHandler returns an HTTP handler function that pushes a metric
// to the specified repository.
func MetricPostFromBodyHandler(repository storage.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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

		var metricValue string
		var err error
		// Подходящий тип
		switch metric.MType {
		case metrics.MetricTypeCounter:
			metricValue = strconv.Itoa(int(*metric.Delta))
		case metrics.MetricTypeGauge:
			metricValue = strconv.FormatFloat(*metric.Value, 'g', -1, 64)
		default:
			BadRequest(w, r)
			return
		}

		// Create a new metric from the parsed parameters.
		m, err := metrics.NewFromStrings(metric.ID, metricValue, metric.MType)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		// Push the created metric to the storage repository.
		err = repository.PushMetric(m)
		if err != nil {
			log.Printf("error pushing metric to store: %v", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		newMetricValue, _ := repository.GetMetric(metric.ID, metric.MType)
		switch metric.MType {
		case metrics.MetricTypeCounter:
			mv, _ := strconv.Atoi(newMetricValue)
			mv1 := int64(mv)
			metric.Delta = &mv1
		case metrics.MetricTypeGauge:
			mv, _ := strconv.ParseFloat(newMetricValue, 64)
			metric.Value = &mv
		default:
			InternalServerError(w, r)
			return
		}

		enc := json.NewEncoder(w)
		err = enc.Encode(metric)
		if err != nil {
			InternalServerError(w, r)
			return
		}

		// Set Content-Type for the response.
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		// Response with response code 200 OK.
		w.WriteHeader(http.StatusOK)
	}
}
