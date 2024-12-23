package value

import (
	"errors"
	"net/http"

	"github.com/gdyunin/metricol.git/internal/server/adapter"
	"github.com/gdyunin/metricol.git/internal/server/consume/consumers/echoserver/parse"
	"github.com/gdyunin/metricol.git/internal/server/entity"
	"github.com/labstack/echo/v4"
)

func FromJSON(adp *adapter.EchoAdapter) func(echo.Context) error {
	return func(c echo.Context) error {
		metric, err := parse.MetricFromJSON(c)
		if err != nil {
			return c.String(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		}

		stored, err := adp.PullMetric(metric)
		if errors.Is(err, entity.ErrMetricNotFound) {
			return c.String(http.StatusNotFound, http.StatusText(http.StatusNotFound))
		}
		return c.JSON(http.StatusOK, stored)
	}
}
