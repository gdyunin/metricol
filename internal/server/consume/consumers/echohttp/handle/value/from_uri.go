package value

import (
	"errors"
	"net/http"

	"github.com/gdyunin/metricol.git/internal/server/adapters/consumers"
	"github.com/gdyunin/metricol.git/internal/server/consume/consumers/echohttp/parse"
	"github.com/gdyunin/metricol.git/internal/server/entities"
	"github.com/labstack/echo/v4"
)

// FromURI creates an HTTP handler function that processes a metric provided in the URI.
// The handler parses the metric, retrieves its stored value, and returns the result as a plain text response.
//
// Parameters:
//   - adp: An instance of EchoAdapter used to interact with the metrics repository.
//
// Returns:
//   - An Echo `HandlerFunc` that processes the URI request, validates and
//     retrieves the metric, and responds accordingly.
func FromURI(adp *consumers.EchoAdapter) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Parse the metric from the URI parameters.
		metric, err := parse.MetricFromURI(c)
		if err != nil {
			// Log the error and return a 400 Bad Request status with a descriptive message.
			return c.String(http.StatusBadRequest, "Failed to parse the metric from the URI parameters.")
		}

		// Attempt to retrieve the metric from the repository.
		stored, err := adp.PullMetric(metric)
		if errors.Is(err, entities.ErrMetricNotFound) {
			// Return a 404 Not Found status if the metric is not found.
			return c.String(http.StatusNotFound, "Metric not found in the repository.")
		} else if err != nil {
			// Log any other error and return a 500 Internal Server Error status.
			return c.String(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		}

		// Convert the stored metric value to a string.
		stringVal, err := stored.StringValue()
		if err != nil {
			// Log the error and return a 500 Internal Server Error status if the value cannot be converted.
			return c.String(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		}

		// Set the Content-Type header and return the metric value as plain text.
		c.Response().Header().Set("Content-Type", "text/plain")
		return c.String(http.StatusOK, stringVal)
	}
}
