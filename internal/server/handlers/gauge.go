package handlers

import (
	"github.com/gdyunin/metricol.git/internal/server/memstorage"
	"github.com/gdyunin/metricol.git/internal/server/metrics"
	"net/http"
	"strconv"
	"strings"
)

func GaugeHandler(memStorage *memstorage.MemStorage) http.HandlerFunc {
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

		// Получаем название и значение
		separated := strings.SplitN(r.URL.String(), "/", 2)
		// Если нет имени метрики
		if len(separated) != 2 {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		name := separated[0]

		// Парсим значение
		value, err := strconv.ParseFloat(separated[1], 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		memStorage.SubmitMetric("gauge", metrics.NewGauge(name, value))
		_ = metrics.NewGauge(name, value)

		w.Header()
		w.WriteHeader(http.StatusOK)
	}
}
