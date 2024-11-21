package handlers

import (
	"github.com/gdyunin/metricol.git/internal/server/memstorage"
	"github.com/gdyunin/metricol.git/internal/server/metrics"
	"github.com/gdyunin/metricol.git/internal/server/metrics/library"
	"net/http"
)

func CounterHandler(memStorage memstorage.MemStorage) http.HandlerFunc {
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

		counter := library.NewCounter()
		err := counter.ParseFromURLString(r.URL.String())
		if err != nil {
			switch err.Error() {
			case metrics.ErrorParseMetricName:
				w.WriteHeader(http.StatusNotFound)
				return
			case metrics.ErrorParseMetricValue:
				w.WriteHeader(http.StatusBadRequest)
				return
			default:
				w.WriteHeader(http.StatusBadRequest)
				return
			}
		}

		_ = memStorage.PushMetric(counter)

		w.WriteHeader(http.StatusOK)
	}
}
