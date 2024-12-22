package handle

import (
	"net/http"

	"github.com/gdyunin/metricol.git/internal/server/adapter"

	"github.com/gin-gonic/gin"
)

// tr represents a simplified table row for body with a name and its corresponding string value.
type tr struct {
	Name  string // The name or ID of the metric.
	Value string // The string representation of the metric's value.
}

// MainPageHandler creates a handler function that retrieves all metrics from the controller,
// converts their values to strings, and renders them in the "main_page.html" template.
//
// Parameters:
// - ctrl: A GinController instance used to interact with the metrics repository.
//
// Behavior:
// - Retrieves all metrics using the `PullAllMetrics` method from the controller.
// - Converts each metric's value to a string, using a placeholder for invalid values.
// - Responds with an HTTP 500 (Internal Server Error) if the metrics cannot be retrieved.
// - If successful, returns an HTTP 200 (OK) response with the rendered HTML page.
func MainPageHandler(ctrl *adapter.GinController) func(*gin.Context) {
	return func(c *gin.Context) {
		// Attempt to retrieve all metrics from the controller.
		am, err := ctrl.PullAllMetrics()
		if err != nil {
			// Return a 500 error with a clear explanation.
			c.String(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
			return
		}

		// Prepare a table to hold the metric name and its string value.
		table := make([]*tr, 0, len(am))
		for _, metric := range am {
			n := metric.ID
			v, err := metric.StringValue()
			if err != nil {
				// Use a placeholder if the metric's value is invalid.
				v = "<invalid metric value>"
			}

			// Add the metric to the table.
			table = append(table, &tr{
				Name:  n,
				Value: v,
			})
		}

		// Render the metrics in the "main_page.html" template and return an HTTP 200 response.
		c.HTML(http.StatusOK, "main_page.html", table)
	}
}
