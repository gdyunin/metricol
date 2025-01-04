package parse

import (
	"fmt"
	"strconv"

	"github.com/gdyunin/metricol.git/internal/server/consume/consumers/echohttp/model"
	"github.com/gdyunin/metricol.git/internal/server/entities"
	"github.com/labstack/echo/v4"
)

// MetricFromURI extracts a Metric object from URI parameters and request context.
//
// Parameters:
//   - c: The Echo context containing the request and URI parameters.
//
// Returns:
//   - A pointer to the extracted `model.Metric` object if successful.
//   - An error if there is an issue parsing the URI parameters or binding the metric.
//
// Errors:
//   - Returns an error if the metric cannot be parsed from the URI or
//     if the value cannot be converted to the appropriate type.
func MetricFromURI(c echo.Context) (*model.Metric, error) {
	m := model.Metric{}

	// Attempt to bind the metric from the URI context.
	if err := c.Bind(&m); err != nil {
		return nil, fmt.Errorf("failed to parse metric from URI: %w", err)
	}

	// Extract the value parameter from the URI.
	valueStr := c.Param("value")
	if valueStr != "" {
		switch m.MType {
		case entities.MetricTypeCounter:
			// Convert value string to int64 for counter metrics.
			delta, err := strconv.ParseInt(valueStr, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("failed to convert '%s' to int64 for counter metric: %w", valueStr, err)
			}
			m.Delta = &delta
		case entities.MetricTypeGauge:
			// Convert value string to float64 for gauge metrics.
			value, err := strconv.ParseFloat(valueStr, 64)
			if err != nil {
				return nil, fmt.Errorf("failed to convert '%s' to float64 for gauge metric: %w", valueStr, err)
			}
			m.Value = &value
		default:
			// Do nothing if the metric type is not recognized.
		}
	}

	return &m, nil
}
