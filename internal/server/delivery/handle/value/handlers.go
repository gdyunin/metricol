package value

import (
	"NewNewMetricol/internal/server/delivery/model"
	"NewNewMetricol/internal/server/internal/entity"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
)

type MetricsPuller interface {
	Pull(metricType string, name string) (*entity.Metric, error)
}

func FromJSON(puller MetricsPuller) echo.HandlerFunc {
	return func(c echo.Context) error {
		m := model.Metric{}
		if err := c.Bind(&m); err != nil {
			return c.String(http.StatusBadRequest, http.StatusText(http.StatusBadRequest))
		}

		metric, err := puller.Pull(m.MType, m.ID)
		if err != nil {
			return c.String(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		}
		if metric == nil {
			return c.String(http.StatusNotFound, "Metric not found in the repository.")
		}

		c.Response().Header().Set("Content-Type", "application/json")
		return c.JSON(http.StatusOK, model.FromEntityMetric(metric))
	}
}

func FromURI(puller MetricsPuller) echo.HandlerFunc {
	return func(c echo.Context) error {
		m := model.Metric{}
		if err := c.Bind(&m); err != nil {
			return c.String(http.StatusBadRequest, http.StatusText(http.StatusBadRequest))
		}

		metric, err := puller.Pull(m.MType, m.ID)
		if err != nil {
			return c.String(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		}
		if metric == nil {
			return c.String(http.StatusNotFound, "Metric not found in the repository.")
		}

		c.Response().Header().Set("Content-Type", "text/plain")
		return c.String(http.StatusOK, fmt.Sprint(metric.Value))
	}
}
