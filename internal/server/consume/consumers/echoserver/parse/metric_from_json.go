package parse

import (
	"fmt"

	"github.com/gdyunin/metricol.git/internal/server/consume/consumers/echoserver/model"
	"github.com/labstack/echo/v4"
)

func MetricFromJSON(c echo.Context) (*model.Metric, error) {
	m := model.Metric{}
	if err := c.Bind(&m); err != nil {
		return nil, fmt.Errorf("error when parse metric from json body: %w", err)
	}
	return &m, nil
}
