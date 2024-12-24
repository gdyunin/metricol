package update

import (
	"net/http"

	"github.com/gdyunin/metricol.git/internal/server/adapter"
	"github.com/gdyunin/metricol.git/internal/server/consume/consumers/echoserver/parse"
	"github.com/gdyunin/metricol.git/pkg/logger"
	"github.com/labstack/echo/v4"
)

func FromURI(adp *adapter.EchoAdapter) echo.HandlerFunc {
	return func(c echo.Context) error {
		l, _ := logger.Logger("INFO")
		l.Info("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
		metric, err := parse.MetricFromURI(c)
		if err != nil {
			return c.String(http.StatusBadRequest, http.StatusText(http.StatusBadRequest))
		}

		if metric.Delta == nil && metric.Value == nil {
			return c.String(http.StatusBadRequest, "expected non-empty delta or value but got empty")
		}

		_, err = adp.PushMetric(metric)
		if err != nil {
			return c.String(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		}

		c.Response().Header().Set("Content-Type", "text/plain")
		return c.String(http.StatusOK, "Metric update successful")
	}
}
