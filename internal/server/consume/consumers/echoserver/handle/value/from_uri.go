package value

import (
	"errors"
	"net/http"

	"github.com/gdyunin/metricol.git/internal/server/adapter"
	"github.com/gdyunin/metricol.git/internal/server/consume/consumers/echoserver/parse"
	"github.com/gdyunin/metricol.git/internal/server/entity"
	"github.com/labstack/echo/v4"
)

func FromURI(adp *adapter.EchoAdapter) func(echo.Context) error {
	return func(c echo.Context) error {
		metric, err := parse.MetricFromURI(c)
		if err != nil {
			return c.String(http.StatusBadRequest, http.StatusText(http.StatusBadRequest))
		}

		stored, err := adp.PullMetric(metric)
		if errors.Is(err, entity.ErrMetricNotFound) {
			return c.String(http.StatusNotFound, http.StatusText(http.StatusNotFound))
		}

		stringVal, err := stored.StringValue()
		if err != nil {
			return c.String(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		}
		return c.String(http.StatusOK, stringVal)
	}
}
