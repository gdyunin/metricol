package handlers

import (
	"github.com/gdyunin/metricol.git/internal/server/memstorage"
	"github.com/gdyunin/metricol.git/internal/server/metrics"
	"github.com/gdyunin/metricol.git/internal/server/metrics/builder"
	"net/http"
)

func MetricPostHandler(memStorage memstorage.MemStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Разрешаем только POST
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		// Проверяем Content-Type
		if r.Header.Get("Content-Type") != "text/plain" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// Парсим УРЛ
		urlArgs := splitURL(r.URL.String(), 3)
		metricType, metricName, metricValue := urlArgs[0], urlArgs[1], urlArgs[2]

		// Обрабатываем ошибки
		if metricType == "" || metricValue == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if metricName == "" {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		metric, err := builder.NewMetric(metrics.MetricType(metricType))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		err = memStorage.PushMetric(metric)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
