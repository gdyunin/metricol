package update

import (
	"net/http"

	"github.com/gdyunin/metricol.git/internal/server/adapters/consumers"
	"github.com/gdyunin/metricol.git/internal/server/consume/consumers/echohttp/parse"
	"github.com/labstack/echo/v4"
)

// FromURI creates an HTTP handler function to update metrics from URI parameters.
// The handler parses the metric data from the URI, validates it, and updates the metric in the repository.
//
// Parameters:
//   - adp: An instance of EchoAdapter used to interact with the metrics repository.
//
// Returns:
//   - An Echo `HandlerFunc` that processes the request, validates and updates the metric, and responds with the status.
func FromURI(adp *consumers.EchoAdapter) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Parse the metric from URI parameters.
		metric, err := parse.MetricFromURI(c)
		if err != nil {
			// Return a 400 Bad Request status if parsing fails.
			return c.String(http.StatusBadRequest, "Invalid metric format or data in URI parameters.")
		}

		// Validate that the metric has a value or delta.
		if metric.Delta == nil && metric.Value == nil {
			// Return a 400 Bad Request status if both delta and value are empty.
			return c.String(http.StatusBadRequest, "Metric must contain either a delta or a value but both are missing.")
		}

		// Attempt to update the metric in the repository.
		_, err = adp.PushMetric(metric)
		if err != nil {
			// Return a 500 Internal Server Error status if the update fails.
			return c.String(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		}

		// Set the Content-Type header and return a success message.
		c.Response().Header().Set("Content-Type", "text/plain")
		return c.String(http.StatusOK, "Metric update successful.")
	}
}
