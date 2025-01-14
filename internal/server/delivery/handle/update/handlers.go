package update

import (
	"net/http"
	"strconv"

	"github.com/gdyunin/metricol.git/internal/server/delivery/model"
	"github.com/gdyunin/metricol.git/internal/server/internal/entity"

	"github.com/labstack/echo/v4"
)

// MetricsUpdater defines the interface for pushing metric updates.
type MetricsUpdater interface {
	PushMetric(*entity.Metric) (*entity.Metric, error)
}

// FromJSON handles metric updates from JSON payloads.
//
// Parameters:
//   - updater: An implementation of MetricsUpdater to process the metric update.
//
// Returns:
//   - An echo.HandlerFunc that processes JSON payloads to update metrics.
func FromJSON(updater MetricsUpdater) echo.HandlerFunc {
	return func(c echo.Context) error {
		m := model.Metric{}
		if err := c.Bind(&m); err != nil {
			return c.String(http.StatusBadRequest, "Invalid JSON payload provided.")
		}

		updated, err := updater.PushMetric(m.ToEntityMetric())
		if err != nil {
			return c.String(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		}

		c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		return c.JSON(http.StatusOK, model.FromEntityMetric(updated))
	}
}

// FromURI handles metric updates from URI parameters.
//
// Parameters:
//   - updater: An implementation of MetricsUpdater to process the metric update.
//
// Returns:
//   - An echo.HandlerFunc that processes URI parameters to update metrics.
func FromURI(updater MetricsUpdater) echo.HandlerFunc {
	return func(c echo.Context) error {
		m := model.Metric{}
		if err := c.Bind(&m); err != nil {
			return c.String(http.StatusBadRequest, "Invalid parameters provided in the request.")
		}

		valueStr := c.Param("value")
		if err := validateMetricValue(&m, valueStr); err != nil {
			return c.String(err.(*echo.HTTPError).Code, err.Error()) //nolint
		}

		_, err := updater.PushMetric(m.ToEntityMetric())
		if err != nil {
			return c.String(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		}

		c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		return c.String(http.StatusOK, "Metric update successful.")
	}
}

// validateMetricValue validates and converts the value based on the metric type.
//
// Parameters:
//   - m: A pointer to the Metric model.
//   - valueStr: The value string extracted from URI parameters.
//
// Returns:
//   - An error if the value is invalid or the metric type is unsupported.
func validateMetricValue(m *model.Metric, valueStr string) error {
	if valueStr == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Required 'value' parameter is missing.")
	}

	switch m.MType {
	case entity.MetricTypeCounter:
		delta, err := strconv.ParseInt(valueStr, 10, 64)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Provided counter value is invalid.")
		}
		m.Delta = &delta
	case entity.MetricTypeGauge:
		value, err := strconv.ParseFloat(valueStr, 64)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Provided gauge value is invalid.")
		}
		m.Value = &value
	default:
		return echo.NewHTTPError(http.StatusBadRequest, "Unsupported metric type.")
	}

	return nil
}
