package update

import (
	"net/http"

	"github.com/gdyunin/metricol.git/internal/server/adapter"
	"github.com/gdyunin/metricol.git/internal/server/consume/consumers/echoserver/parse"
	"github.com/labstack/echo/v4"
)

func FromJSON(adp *adapter.EchoAdapter) func(echo.Context) error {
	return func(c echo.Context) error {
		metric, err := parse.MetricFromJSON(c)
		if err != nil {
			return c.String(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		}

		if metric.Delta == nil && metric.Value == nil {
			return c.String(http.StatusBadRequest, "expected non-empty delta or value but got empty")
		}

		updated, err := adp.PushMetric(metric)
		if err != nil {
			return c.String(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		}
		return c.JSON(http.StatusOK, updated)
	}
}