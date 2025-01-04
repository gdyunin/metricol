package parse

import (
	"fmt"
	"strconv"

	"github.com/gdyunin/metricol.git/internal/server/consume/consumers/echohttp/model"
	"github.com/gdyunin/metricol.git/internal/server/entities"
	"github.com/labstack/echo/v4"
)

func MetricFromURI(c echo.Context) (*model.Metric, error) {
	m := model.Metric{}
	if err := c.Bind(&m); err != nil {
		return nil, fmt.Errorf("error when parse metric from URI: %w", err)
	}

	valueStr := c.Param("value")
	if valueStr != "" {
		switch m.MType {
		case entities.MetricTypeCounter:
			delta, err := strconv.ParseInt(valueStr, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("error when cast %s to int64: %w", valueStr, err)
			}
			m.Delta = &delta
		case entities.MetricTypeGauge:
			value, err := strconv.ParseFloat(valueStr, 64)
			if err != nil {
				return nil, fmt.Errorf("error when cast %s to float64: %w", valueStr, err)
			}
			m.Value = &value
		}
	}
	return &m, nil
}
