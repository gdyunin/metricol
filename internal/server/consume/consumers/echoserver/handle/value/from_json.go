package value

import (
	"errors"
	"net/http"

	"github.com/gdyunin/metricol.git/internal/server/adapter"
	"github.com/gdyunin/metricol.git/internal/server/consume/consumers/echoserver/model"
	"github.com/gdyunin/metricol.git/internal/server/consume/consumers/echoserver/parse"
	"github.com/gdyunin/metricol.git/internal/server/entity"
	"github.com/labstack/echo/v4"
)

func FromJSON(adp *adapter.EchoAdapter) echo.HandlerFunc {
	return func(c echo.Context) error {
		metric, err := parse.MetricFromJSON(c)
		if err != nil {
			return c.String(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		}

		if !isValidModel(metric) {
			return c.String(http.StatusBadRequest, http.StatusText(http.StatusBadRequest))
		}

		stored, err := adp.PullMetric(metric)
		if errors.Is(err, entity.ErrMetricNotFound) {
			return c.String(http.StatusNotFound, http.StatusText(http.StatusNotFound))
		}

		c.Response().Header().Set("Content-Type", "application/json")
		return c.JSON(http.StatusOK, stored)
	}
}

func isValidModel(metric *model.Metric) bool {
	if metric.ID == "" || metric.MType == "" {
		return false
	}
	return true
}
