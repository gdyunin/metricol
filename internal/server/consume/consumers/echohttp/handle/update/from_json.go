package update

import (
	"net/http"

	"github.com/gdyunin/metricol.git/internal/server/adapters/consumers"
	"github.com/gdyunin/metricol.git/internal/server/consume/consumers/echohttp/model"
	"github.com/gdyunin/metricol.git/internal/server/consume/consumers/echohttp/parse"
	"github.com/gdyunin/metricol.git/internal/server/entities"
	"github.com/labstack/echo/v4"
)

// FromJSON creates an HTTP handler function that processes a metric sent in JSON format.
// The handler validates the metric, updates it in the repository, and returns the updated metric.
//
// Parameters:
//   - adp: An instance of EchoAdapter used to interact with the metrics repository.
//
// Returns:
//   - An Echo `HandlerFunc` that processes the JSON request, validates and updates the metric, and responds with the updated metric.
func FromJSON(adp *consumers.EchoAdapter) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Parse the metric from the JSON body.
		metric, err := parse.MetricFromJSON(c)
		if err != nil {
			// Log the error and return a 500 Internal Server Error status with a descriptive message.
			return c.String(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		}

		// Validate the parsed metric.
		if !isValidModel(metric) {
			// Return a 400 Bad Request status if the metric is invalid.
			return c.String(http.StatusBadRequest, "The metric data is invalid or incomplete.")
		}

		// Attempt to update the metric in the repository.
		updated, err := adp.PushMetric(metric)
		if err != nil {
			// Log the error and return a 500 Internal Server Error status if the update fails.
			return c.String(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		}

		// Set the Content-Type header and return the updated metric as JSON.
		c.Response().Header().Set("Content-Type", "application/json")
		return c.JSON(http.StatusOK, updated)
	}
}

// isValidModel validates the structure and data of the given metric.
//
// Parameters:
//   - metric: A pointer to the `model.Metric` to validate.
//
// Returns:
//   - `true` if the metric is valid, otherwise `false`.
func isValidModel(metric *model.Metric) bool {
	// Ensure the metric has an ID and type.
	if metric.ID == "" || metric.MType == "" {
		return false
	}

	// Validate the metric value based on its type.
	switch metric.MType {
	case entities.MetricTypeCounter:
		if metric.Delta == nil {
			return false
		}
	case entities.MetricTypeGauge:
		if metric.Value == nil {
			return false
		}
	default:
		// Return false for unsupported metric types.
		return false
	}

	return true
}
