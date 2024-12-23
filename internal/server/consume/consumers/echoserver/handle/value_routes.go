package handle

import (
	"net/http"

	"github.com/gdyunin/metricol.git/internal/server/adapter"
	"github.com/gdyunin/metricol.git/internal/server/consume/consumers/ginserver/parse"
	"github.com/labstack/echo/v4"
)

// ValueHandlerWithURIParams creates a handler function that processes metric retrieval
// using URI parameters.
//
// Parameters:
// - ctrl: A GinController instance used to manage metrics.
//
// Behavior:
// - Parses the metric from URI parameters using `MetricFromURI`.
// - Retrieves the metric from the controller using `PullMetric`.
// - Converts the metric value to a string using `StringValue`.
// - Returns an HTTP 400 (Bad Request) if the metric parsing fails.
// - Returns an HTTP 500 (Internal Server Error) if metric retrieval or value conversion fails.
// - Returns an HTTP 200 (OK) with the metric value as a string if successful.
func ValueHandlerWithURIParams(ctrl *adapter.GinController) func(echo.Context) error {
	return func(c echo.Context) error {
		// Parse the metric from URI parameters.
		m, err := parse.MetricFromURIe(c)
		if err != nil {
			// Respond with a clear error message if parsing fails.
			c.String(http.StatusBadRequest, "Failed to parse metric from URI parameters: invalid or incomplete data.")
			return nil
		}

		isExists, err := ctrl.IsExists(m)
		if err != nil {
			// Respond with a clear error message if metric retrieval fails.
			c.String(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
			return nil
		}
		if !isExists {
			// Respond with a clear error message if metric retrieval fails.
			c.String(http.StatusNotFound, http.StatusText(http.StatusNotFound))
			return nil
		}

		// Retrieve the metric from the controller.
		m, err = ctrl.PullMetric(m)
		if err != nil {
			// Respond with a clear error message if metric retrieval fails.
			c.String(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
			return nil
		}

		// Convert the metric value to a string.
		sValue, err := m.StringValue()
		if err != nil {
			// Respond with a clear error message if value conversion fails.
			c.String(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
			return nil
		}

		// Respond with the metric value.
		c.String(http.StatusOK, sValue)
		return nil
	}
}

// ValueHandlerWithJSONParams creates a handler function that processes metric retrieval
// using JSON parameters.
//
// Parameters:
// - ctrl: A GinController instance used to manage metrics.
//
// Behavior:
// - Parses the metric from the JSON body using `MetricFromJSON`.
// - Retrieves the metric from the controller using `PullMetric`.
// - Returns an HTTP 400 (Bad Request) if the metric parsing fails.
// - Returns an HTTP 500 (Internal Server Error) if metric retrieval fails.
// - Returns an HTTP 200 (OK) with the metric data in JSON format if successful.
func ValueHandlerWithJSONParams(ctrl *adapter.GinController) func(echo.Context) error {
	return func(c echo.Context) error {
		// Parse the metric from the JSON body.
		m, err := parse.MetricFromJSONe(c)
		if err != nil {
			// Respond with a clear error message if parsing fails.
			c.String(http.StatusBadRequest, "Failed to parse metric from JSON body: invalid or incomplete data.")
			return nil
		}

		isExists, err := ctrl.IsExists(m)
		if err != nil {
			// Respond with a clear error message if metric retrieval fails.
			c.String(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
			return nil
		}
		if !isExists {
			// Respond with a clear error message if metric retrieval fails.
			c.String(http.StatusNotFound, http.StatusText(http.StatusNotFound))
			return nil
		}

		// Retrieve the metric from the controller.
		m, err = ctrl.PullMetric(m)
		if err != nil {
			// Respond with a clear error message if metric retrieval fails.
			c.String(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
			return nil
		}

		// Respond with the metric data in JSON format.
		c.JSON(http.StatusOK, m)
		return nil
	}
}
