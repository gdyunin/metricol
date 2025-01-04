package value

import (
	"errors"
	"net/http"

	"github.com/gdyunin/metricol.git/internal/server/adapters/consumers"
	"github.com/gdyunin/metricol.git/internal/server/consume/consumers/echohttp/model"
	"github.com/gdyunin/metricol.git/internal/server/consume/consumers/echohttp/parse"
	"github.com/gdyunin/metricol.git/internal/server/entities"
	"github.com/labstack/echo/v4"
)

// FromJSON creates an HTTP handler function that processes a metric sent in JSON format.
// The handler validates the metric, attempts to retrieve it from the repository, and returns the result.
//
// Parameters:
//   - adp: An instance of EchoAdapter used to interact with the metrics repository.
//
// Returns:
//   - An Echo `HandlerFunc` that processes the JSON request, validates the metric, retrieves it, and responds accordingly.
func FromJSON(adp *consumers.EchoAdapter) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Parse the metric from the JSON request body.
		metric, err := parse.MetricFromJSON(c)
		if err != nil {
			// Log the error and return a 500 Internal Server Error status.
			return c.String(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		}

		// Validate the parsed metric model.
		if !isValidModel(metric) {
			// Return a 400 Bad Request status if the metric is invalid.
			return c.String(http.StatusBadRequest, "Invalid metric model: missing required fields.")
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

		// Set the Content-Type header and return the retrieved metric as JSON.
		c.Response().Header().Set("Content-Type", "application/json")
		return c.JSON(http.StatusOK, stored)
	}
}

// isValidModel validates the provided metric model.
//
// Parameters:
//   - metric: A pointer to the metric model to validate.
//
// Returns:
//   - True if the metric model is valid (i.e., it has both an ID and a type), false otherwise.
func isValidModel(metric *model.Metric) bool {
	return metric.ID != "" && metric.MType != ""
}
