package update

import (
	"net/http"

	"github.com/gdyunin/metricol.git/internal/server/adapter"
	"github.com/gdyunin/metricol.git/internal/server/consume/consumers/echoserver/model"
	"github.com/gdyunin/metricol.git/internal/server/consume/consumers/echoserver/parse"
	"github.com/gdyunin/metricol.git/internal/server/entity"
	"github.com/gdyunin/metricol.git/pkg/logger"
	"github.com/labstack/echo/v4"
)

func FromJSON(adp *adapter.EchoAdapter) echo.HandlerFunc {
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
	case entity.MetricTypeCounter:
		if metric.Delta == nil {
			return false
		}
	case entity.MetricTypeGauge:
		if metric.Value == nil {
			return false
		}
	default:
		return false
	}

	return true
}
