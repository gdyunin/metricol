package general

import (
	"net/http"

	"github.com/gdyunin/metricol.git/internal/server/adapters/consumers"
	"github.com/labstack/echo/v4"
)

// tr represents a table row containing a metric's name and its string value.
type tr struct {
	Name  string // The name or ID of the metric.
	Value string // The string representation of the metric's value.
}

// MainPage creates an HTTP handler function that serves the main page of the application.
// The handler retrieves all metrics from the adapter, processes them, and renders them
// in a template.
//
// Parameters:
//   - adp: An instance of EchoAdapter used to interact with metrics.
//
// Returns:
//   - An Echo `HandlerFunc` that renders the main page or responds with an appropriate
//     HTTP status code and message in case of an error.
func MainPage(adp *consumers.EchoAdapter) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Retrieve all metrics from the adapter.
		allMetrics, err := adp.PullAllMetrics()
		if err != nil {
			// Log an error and return a 500 Internal Server Error status.
			return c.String(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		}

		// Process metrics and prepare a table for rendering.
		table := make([]*tr, 0, len(allMetrics))
		for _, metric := range allMetrics {
			name := metric.ID
			value, err := metric.StringValue()
			if err != nil {
				// Handle cases where the metric's value cannot be converted to a string.
				value = "<invalid metric value>"
			}
			table = append(table, &tr{
				Name:  name,
				Value: value,
			})
		}

		// Render the main page template with the metrics table.
		return c.Render(http.StatusOK, "main_page.html", table)
	}
}
