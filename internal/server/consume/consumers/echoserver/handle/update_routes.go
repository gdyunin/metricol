package handle

import (
	"net/http"

	"github.com/gdyunin/metricol.git/internal/server/adapter"
	"github.com/gdyunin/metricol.git/internal/server/consume/consumers/ginserver/model"
	"github.com/gdyunin/metricol.git/internal/server/consume/consumers/ginserver/parse"
	"github.com/labstack/echo/v4"
)

// UpdateHandlerWithURIParams creates a handler function that processes metric updates
// using URI parameters.
//
// Parameters:
// - ctrl: A GinController instance used to manage metrics.
//
// Behavior:
// - Parses the metric from the URI parameters using `MetricFromURI`.
// - Validates that the metric contains either a `Value` or a `Delta` field.
// - Pushes the metric to the controller for storage or processing.
// - Returns an HTTP 400 (Bad Request) if the metric is invalid or missing required fields.
// - Returns an HTTP 500 (Internal Server Error) if pushing the metric fails.
// - Returns an HTTP 200 (OK) if the metric is successfully pushed.
func UpdateHandlerWithURIParams(ctrl *adapter.GinController) func(echo.Context) error {
	return func(c echo.Context) error {
		// Parse the metric from URI parameters.
		m, err := parse.MetricFromURIe(c)
		if err != nil || !isFillValueFields(m) {
			// Respond with a clear error message if the metric is invalid.
			c.String(http.StatusBadRequest, "Invalid metric data: ensure all required fields are properly set.")
			return nil
		}

		// Push the metric to the controller.
		_, err = ctrl.PushMetric(m)
		if err != nil {
			// Respond with a clear error message if the metric cannot be pushed.
			c.String(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
			return nil
		}

		// Respond with a success message.
		c.String(http.StatusOK, "Metric successfully saved.")
		return nil
	}
}

// UpdateHandlerWithJSONParams creates a handler function that processes metric updates
// using JSON parameters.
//
// Parameters:
// - ctrl: A GinController instance used to manage metrics.
//
// Behavior:
// - Parses the metric from the JSON body using `MetricFromJSON`.
// - Validates that the metric contains either a `Value` or a `Delta` field.
// - Pushes the metric to the controller for storage or processing.
// - Returns an HTTP 400 (Bad Request) if the metric is invalid or missing required fields.
// - Returns an HTTP 500 (Internal Server Error) if pushing the metric fails.
// - Returns an HTTP 200 (OK) along with the updated metric in JSON format if successful.
func UpdateHandlerWithJSONParams(ctrl *adapter.GinController) func(echo.Context) error {
	return func(c echo.Context) error {
		// Parse the metric from the JSON body.
		m, err := parse.MetricFromJSONe(c)
		if err != nil || !isFillValueFields(m) {
			// Respond with a clear error message if the metric is invalid.
			c.String(http.StatusBadRequest, "Invalid metric data: ensure all required fields are properly set.")
			return nil
		}

		// Push the metric to the controller.
		m, err = ctrl.PushMetric(m)
		if err != nil {
			// Respond with a clear error message if the metric cannot be pushed.
			c.String(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
			return nil
		}

		// Respond with the updated metric in JSON format.
		c.JSON(http.StatusOK, m)
		return nil
	}
}

// isFillValueFields checks if the metric contains either a value or a delta.
//
// Parameters:
// - m: A pointer to the Metric struct to validate.
//
// Returns:
// - true if the metric contains a non-nil `Value` or `Delta` field.
// - false otherwise.
func isFillValueFields(m *model.Metric) bool {
	return m.Value != nil || m.Delta != nil
}
