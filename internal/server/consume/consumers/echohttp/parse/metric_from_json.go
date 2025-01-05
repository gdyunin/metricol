package parse

import (
	"fmt"

	"github.com/gdyunin/metricol.git/internal/server/consume/consumers/echohttp/model"
	"github.com/labstack/echo/v4"
)

// MetricFromJSON extracts a Metric object from the JSON body of an HTTP request.
//
// Parameters:
//   - c: The Echo context containing the request.
//
// Returns:
//   - A pointer to the extracted `model.Metric` object if successful.
//   - An error if the JSON body cannot be parsed or bound to the Metric object.
//
// Errors:
//   - Returns an error if there is an issue parsing the metric from the JSON body.
func MetricFromJSON(c echo.Context) (*model.Metric, error) {
	m := model.Metric{}

	// Attempt to bind the metric from the JSON body of the request.
	if err := c.Bind(&m); err != nil {
		return nil, fmt.Errorf("failed to parse metric from JSON body: %w", err)
	}

	return &m, nil
}
