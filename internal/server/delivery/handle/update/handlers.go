package update

import (
	"NewNewMetricol/internal/server/delivery/model"
	"NewNewMetricol/internal/server/internal/entity"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

type MetricsUpdater interface {
	PushMetric(*entity.Metric) (*entity.Metric, error)
}

func FromJSON(updater MetricsUpdater) echo.HandlerFunc {
	return func(c echo.Context) error {
		m := model.Metric{}
		if err := c.Bind(&m); err != nil {
			return c.String(http.StatusBadRequest, http.StatusText(http.StatusBadRequest))
		}

		updated, err := updater.PushMetric(m.ToEntityMetric())
		if err != nil {
			return c.String(http.StatusBadRequest, http.StatusText(http.StatusBadRequest))
		}

		c.Response().Header().Set("Content-Type", "application/json")
		return c.JSON(http.StatusOK, model.FromEntityMetric(updated))
	}
}

func FromURI(updater MetricsUpdater) echo.HandlerFunc {
	return func(c echo.Context) error {
		m := model.Metric{}
		if err := c.Bind(&m); err != nil {
			return c.String(http.StatusBadRequest, http.StatusText(http.StatusBadRequest))
		}

		valueStr := c.Param("value")
		if valueStr == "" {
			return c.String(http.StatusBadRequest, http.StatusText(http.StatusBadRequest))
		}

		switch m.MType {
		case entity.MetricTypeCounter:
			delta, err := strconv.ParseInt(valueStr, 10, 64)
			if err != nil {
				return c.String(http.StatusBadRequest, http.StatusText(http.StatusBadRequest))
			}
			m.Delta = &delta
		case entity.MetricTypeGauge:
			value, err := strconv.ParseFloat(valueStr, 64)
			if err != nil {
				return c.String(http.StatusBadRequest, http.StatusText(http.StatusBadRequest))
			}
			m.Value = &value
		default:
			return c.String(http.StatusBadRequest, http.StatusText(http.StatusBadRequest))
		}

		_, err := updater.PushMetric(m.ToEntityMetric())
		if err != nil {
			return c.String(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		}

		c.Response().Header().Set("Content-Type", "application/json")
		return c.String(http.StatusOK, "Metric update successful.")
	}
}
