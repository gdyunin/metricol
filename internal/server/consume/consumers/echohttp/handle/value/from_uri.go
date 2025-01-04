package value

import (
	"errors"
	"net/http"

	"github.com/gdyunin/metricol.git/internal/server/adapters/consumers"
	"github.com/gdyunin/metricol.git/internal/server/consume/consumers/echohttp/parse"
	"github.com/gdyunin/metricol.git/internal/server/entities"
	"github.com/labstack/echo/v4"
)

func FromURI(adp *consumers.EchoAdapter) echo.HandlerFunc {
	return func(c echo.Context) error {
		metric, err := parse.MetricFromURI(c)
		if err != nil {
			return c.String(http.StatusBadRequest, http.StatusText(http.StatusBadRequest))
		}

		stored, err := adp.PullMetric(metric)
		if errors.Is(err, entities.ErrMetricNotFound) {
			return c.String(http.StatusNotFound, http.StatusText(http.StatusNotFound))
		}

		stringVal, err := stored.StringValue()
		if err != nil {
			return c.String(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		}

		c.Response().Header().Set("Content-Type", "text/plain")
		return c.String(http.StatusOK, stringVal)
	}
}
