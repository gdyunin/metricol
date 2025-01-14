package updates

import (
	"net/http"

	"github.com/gdyunin/metricol.git/internal/server/delivery/model"
	"github.com/gdyunin/metricol.git/internal/server/internal/entity"
	"github.com/labstack/echo/v4"
)

// MetricsUpdater defines the interface for pushing metric updates.
type MetricsUpdater interface {
	PushMetric(*entity.Metric) (*entity.Metric, error)
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

		// TODO: Не самое оптимальное с точки зрения bigO и аллокаций. Нужно оптимизировать.
		for _, m := range models {
			if !isValidMetric(m) {
				return c.String(http.StatusBadRequest, "Invalid parameters provided in the request.")
			}
			metrics = append(metrics, m.ToEntityMetric())
		}

		updatedMetrics := entity.Metrics{}
		// [ДЛЯ РЕВЬЮ]: Да, это полный бред кидать по одной метрике, надо кидать пачкой. Не успел реализовать((.
		// TODO: Нужно у контроллера сделать ручку на обновление пачки метрик и дёргать ее сразу, без циклов.
		// TODO: она должна работать с контекстом и принимать *model.Metrics.
		for _, m := range metrics {
			updated, err := updater.PushMetric(m)
			if err != nil {
				return c.String(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
			}
			updatedMetrics = append(updatedMetrics, updated)
		}

		c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		return c.JSON(http.StatusOK, model.FromEntityMetrics(&updatedMetrics))
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
