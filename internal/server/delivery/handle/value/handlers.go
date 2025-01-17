package value

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gdyunin/metricol.git/internal/server/delivery/model"
	"github.com/gdyunin/metricol.git/internal/server/internal/controller"
	"github.com/gdyunin/metricol.git/internal/server/internal/entity"

	"github.com/labstack/echo/v4"
)

const MetricUpdateTimeout = 5 * time.Second

// MetricsPuller defines the interface for retrieving metrics.
type MetricsPuller interface {
	Pull(ctx context.Context, metricType string, name string) (*entity.Metric, error)
}

// FromJSON handles HTTP requests to fetch a metric's value using JSON payloads.
//
// Parameters:
//   - puller: An implementation of MetricsPuller to fetch metrics.
//
// Returns:
//   - An echo.HandlerFunc that processes JSON payloads to retrieve metric values.
func FromJSON(puller MetricsPuller) echo.HandlerFunc {
	return func(c echo.Context) error {
		m := model.Metric{}
		if err := c.Bind(&m); err != nil {
			return c.String(http.StatusBadRequest, http.StatusText(http.StatusBadRequest))
		}

		ctx, cancel := context.WithTimeout(c.Request().Context(), MetricUpdateTimeout)
		defer cancel()

		metric, err := pullMetric(ctx, puller, m)
		if err != nil {
			return c.String(err.(*echo.HTTPError).Code, err.Error()) //nolint
		}

		c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		return c.JSON(http.StatusOK, model.FromEntityMetric(metric))
	}
}

// FromURI handles HTTP requests to fetch a metric's value using URI parameters.
//
// Parameters:
//   - puller: An implementation of MetricsPuller to fetch metrics.
//
// Returns:
//   - An echo.HandlerFunc that processes URI parameters to retrieve metric values.
func FromURI(puller MetricsPuller) echo.HandlerFunc {
	return func(c echo.Context) error {
		m := model.Metric{}
		if err := c.Bind(&m); err != nil {
			return c.String(http.StatusBadRequest, http.StatusText(http.StatusBadRequest))
		}

		ctx, cancel := context.WithTimeout(c.Request().Context(), MetricUpdateTimeout)
		defer cancel()

		metric, err := pullMetric(ctx, puller, m)
		if err != nil {
			return c.String(err.(*echo.HTTPError).Code, err.Error()) //nolint
		}

		c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextPlain)
		return c.String(http.StatusOK, fmt.Sprint(metric.Value))
	}
}

// pullMetric is a helper function that retrieves a metric using the provided puller.
//
// Parameters:
//   - puller: The MetricsPuller instance used to fetch the metric.
//   - m: The Metric model containing the type and ID of the metric.
//
// Returns:
//   - The fetched metric if found.
//   - An error response if the metric is not found or if an error occurs during retrieval.
func pullMetric(ctx context.Context, puller MetricsPuller, m model.Metric) (*entity.Metric, error) {
	metric, err := puller.Pull(ctx, m.MType, m.ID)
	if err != nil {
		if errors.Is(err, controller.ErrNotFoundInRepository) {
			return nil, echo.NewHTTPError(http.StatusNotFound, "Metric not found in the repository.")
		}
		return nil, echo.NewHTTPError(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	}
	return metric, nil
}
