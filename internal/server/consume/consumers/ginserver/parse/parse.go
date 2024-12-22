package parse

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/gdyunin/metricol.git/internal/server/consume/consumers/ginserver/model"
	"github.com/gdyunin/metricol.git/internal/server/entity"

	"github.com/gin-gonic/gin"
)

// MetricFromURI parses a metric from the URI parameters provided in the HTTP request context.
// It binds the URI parameters to a Metric struct, validates that required fields (ID and MType) are present,
// and parses the value based on the metric type (if included).
//
// Parameters:
//   - c: The Gin HTTP context containing the request and its associated parameters.
//
// Returns:
//   - A pointer to a model.Metric struct representing the parsed metric.
//   - An error if parsing fails or required fields are missing.
func MetricFromURI(c *gin.Context) (*model.Metric, error) {
	var m model.Metric

	// Bind URI parameters to the Metric struct.
	if err := c.ShouldBindUri(&m); err != nil {
		return nil, fmt.Errorf("unable to bind URI parameters to metric: %w", err)
	}

	// Validate that the required fields (ID and MType) are present.
	if m.ID == "" {
		return nil, errors.New("missing required metric ID in URI")
	}
	if m.MType == "" {
		return nil, errors.New("missing required metric type in URI")
	}

	// Parse the metric value from the URI (if provided).
	stringValue := strings.TrimPrefix(c.Param("value"), "/")
	if stringValue != "" {
		switch m.MType {
		case entity.MetricTypeCounter:
			// Parse the value as an integer for counters (Delta).
			v, err := strconv.ParseInt(stringValue, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("invalid metric value '%s': expected integer for counter type: %w", stringValue, err)
			}
			m.Delta = &v
		case entity.MetricTypeGauge:
			// Parse the value as a float for gauges (Value).
			v, err := strconv.ParseFloat(stringValue, 64)
			if err != nil {
				return nil, fmt.Errorf("invalid metric value '%s': expected float for gauge type: %w", stringValue, err)
			}
			m.Value = &v
		default:
			return nil, fmt.Errorf("unknown metric type '%s' in URI: must be 'counter' or 'gauge'", m.MType)
		}
	}

	return &m, nil
}

// MetricFromJSON parses a metric from the JSON body of the HTTP request.
// It binds the JSON data to a Metric struct and validates its structure.
//
// Parameters:
//   - c: The Gin HTTP context containing the request and its associated JSON body.
//
// Returns:
//   - A pointer to a model.Metric struct representing the parsed metric.
//   - An error if the JSON body is invalid or cannot be parsed into a Metric struct.
func MetricFromJSON(c *gin.Context) (*model.Metric, error) {
	var m model.Metric

	// Bind the JSON body to the Metric struct.
	if err := c.ShouldBindJSON(&m); err != nil {
		return nil, fmt.Errorf("unable to bind JSON body to metric: %w", err)
	}

	return &m, nil
}
