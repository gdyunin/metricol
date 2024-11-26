package library

import (
	"errors"
	"fmt"
	"github.com/gdyunin/metricol.git/internal/server/metrics"
	"strconv"
)

type Gauge struct {
	name  string
	value float64
}

func NewGauge() *Gauge {
	return &Gauge{}
}

func (g *Gauge) SetName(name string) error {
	if name == "" {
		return errors.New(metrics.ErrorEmptyName)
	}
	g.name = name
	return nil
}

func (g *Gauge) SetValue(val string) error {
	if val == "" {
		return errors.New(metrics.ErrorEmptyValue)
	}

	v, err := strconv.ParseFloat(val, 64)
	if err != nil {
		return errors.New(metrics.ErrorParseMetricValue)
	}

	g.value = v
	return nil
}

func (g Gauge) Name() string {
	return g.name
}

func (g Gauge) Value() string {
	return fmt.Sprintf("%g", g.value)
}

func (g Gauge) Type() metrics.MetricType {
	return metrics.MetricTypeGauge
}
