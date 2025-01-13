package general

import (
	"fmt"
	"net/http"

	"github.com/gdyunin/metricol.git/internal/server/internal/entity"

	"github.com/labstack/echo/v4"
)

// tr represents a table row with a metric name and value.
type tr struct {
	Name  string // Name of the metric.
	Value string // Value of the metric as a string.
}

// PullerAll defines an interface for retrieving all metrics.
type PullerAll interface {
	PullAll() (*entity.Metrics, error) // PullAll retrieves all metrics from the repository or other storage.
}

// MainPage returns an HTTP handler function that renders the main page with metrics.
//
// Parameters:
//   - puller: An implementation of the PullerAll interface for fetching all metrics.
//
// Returns:
//   - An echo.HandlerFunc that handles HTTP requests for the main page.
func MainPage(puller PullerAll) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Attempt to fetch all metrics. If an error occurs or the result is nil, respond with 500 Internal Server Error.
		allMetrics, err := puller.PullAll()
		if err != nil || allMetrics == nil {
			return c.String(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		}

		// Initialize a slice to store table rows, pre-allocated to the number of metrics for efficiency.
		table := make([]*tr, 0, allMetrics.Length())

		// Iterate through all metrics to transform them into table rows.
		for _, metric := range *allMetrics {
			// Extract the name and value of the metric. Use fmt.Sprint to safely convert the value to a string.
			name := metric.Name
			value := fmt.Sprint(metric.Value)

			// Append a new row to the table with the metric's name and value.
			table = append(table, &tr{
				Name:  name,
				Value: value,
			})
		}

		return c.Render(http.StatusOK, "main_page.html", table)
	}
}
