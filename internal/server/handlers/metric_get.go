package handlers

import (
	"github.com/gdyunin/metricol.git/internal/server/storage"
	"net/http"
)

func MetricGetHandler(repository storage.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Response
		w.WriteHeader(http.StatusOK)
	}
}
