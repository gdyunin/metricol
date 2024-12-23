package parse

import (
	"strconv"

	"github.com/gdyunin/metricol.git/internal/server/consume/consumers/echoserver/model"
	"github.com/gdyunin/metricol.git/internal/server/entity"
	"github.com/labstack/echo/v4"
)

func MetricFromURI(c echo.Context) (*model.Metric, error) {
	m := model.Metric{}
	if err := c.Bind(&m); err != nil {
		return nil, err
	}

	valueStr := c.Param("value")
	if valueStr != "" {
		switch m.MType {
		case entity.MetricTypeCounter:
			delta, err := strconv.ParseInt(valueStr, 10, 64)
			if err != nil {
				return nil, err
			}
			m.Delta = &delta
		case entity.MetricTypeGauge:
			value, err := strconv.ParseFloat(valueStr, 64)
			if err != nil {
				return nil, err
			}
			m.Value = &value
		}
	}
	return &m, nil
}
