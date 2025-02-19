package updates

import (
	"context"
	"net/http"
	"time"

	"github.com/gdyunin/metricol.git/internal/server/delivery/model"
	"github.com/gdyunin/metricol.git/internal/server/internal/entity"
	"github.com/labstack/echo/v4"
)

const metricUpdateTimeout = 5 * time.Second

// MetricsUpdater defines the interface for pushing metric updates.
type MetricsUpdater interface {
	PushMetrics(context.Context, *entity.Metrics) (*entity.Metrics, error)
}

// FromJSON handles incoming JSON requests to update metrics.
// It validates the input, processes each metric, and returns the updated metrics in JSON format.
//
// Parameters:
//   - updater: An implementation of the MetricsUpdater interface used to process the metrics.
//
// Returns:
//   - An echo.HandlerFunc to handle the HTTP request and response cycle.
func FromJSON(updater MetricsUpdater) echo.HandlerFunc {
	return func(c echo.Context) error {
		models := model.Metrics{}
		metrics := entity.Metrics{}
		if err := c.Bind(&models); err != nil {
			return c.String(http.StatusBadRequest, "Invalid parameters provided in the request.")
		}

		for _, m := range models {
			if !isValidMetric(m) {
				return c.String(http.StatusBadRequest, "Invalid parameters provided in the request.")
			}
			metrics = append(metrics, m.ToEntityMetric())
		}

		ctx, cancel := context.WithTimeout(c.Request().Context(), metricUpdateTimeout)
		defer cancel()

		updatedMetrics, err := updater.PushMetrics(ctx, &metrics)
		if err != nil {
			return c.String(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		}

		c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		return c.JSON(http.StatusOK, model.FromEntityMetrics(updatedMetrics))
	}
}

// isValidMetric validates the structure of a metric.
// A valid metric must have a non-empty ID and type, and at least one non-nil value field.
//
// Parameters:
//   - m: A pointer to a model.Metric object to be validated.
//
// Returns:
//   - A boolean indicating whether the metric is valid.
func isValidMetric(m *model.Metric) bool {
	return m.ID != "" && m.MType != "" && (m.Delta != nil || m.Value != nil)
}
