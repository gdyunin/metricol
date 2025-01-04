package update

import (
	"net/http"

	"github.com/gdyunin/metricol.git/internal/server/adapters/consumers"
	"github.com/gdyunin/metricol.git/internal/server/consume/consumers/echohttp/model"
	"github.com/gdyunin/metricol.git/internal/server/consume/consumers/echohttp/parse"
	"github.com/gdyunin/metricol.git/internal/server/entities"
	"github.com/gdyunin/metricol.git/pkg/logger"
	"github.com/labstack/echo/v4"
)

func FromJSON(adp *consumers.EchoAdapter) echo.HandlerFunc {
	return func(c echo.Context) error {
		l, _ := logger.Logger("INFO")
		l.Info("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
		metric, err := parse.MetricFromJSON(c)
		if err != nil {
			return c.String(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		}

		if !isValidModel(metric) {
			return c.String(http.StatusBadRequest, http.StatusText(http.StatusBadRequest))
		}

		updated, err := adp.PushMetric(metric)
		if err != nil {
			return c.String(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		}

		c.Response().Header().Set("Content-Type", "application/json")
		return c.JSON(http.StatusOK, updated)
	}
}

func isValidModel(metric *model.Metric) bool {
	if metric.ID == "" || metric.MType == "" {
		return false
	}

	switch metric.MType {
	case entities.MetricTypeCounter:
		if metric.Delta == nil {
			return false
		}
	case entities.MetricTypeGauge:
		if metric.Value == nil {
			return false
		}
	default:
		return false
	}

	return true
}
