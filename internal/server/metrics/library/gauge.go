package library

import (
	"fmt"
	"github.com/gdyunin/metricol.git/internal/server/metrics"
	"strconv"
	"strings"
)

type Gauge struct {
	name       string
	value      float64
	metricType metrics.MetricType
}

func NewGauge() *Gauge {
	return &Gauge{
		metricType: metrics.MetricTypeGauge,
	}
}

func (g *Gauge) ParseFromURLString(u string) error {
	separated := strings.SplitN(u, "/", 2)

	if len(separated) != 2 {
		return fmt.Errorf(metrics.ErrorParseMetricName)
	}
	g.name = separated[0]

	value, err := strconv.ParseFloat(separated[1], 64)
	if err != nil {
		return fmt.Errorf(metrics.ErrorParseMetricValue)
	}
	g.value = value

	return nil
}

func (g Gauge) Name() string {
	return g.name
}

func (g Gauge) Value() string {
	return strconv.FormatFloat(g.value, 'f', 12, 64)
}

func (g Gauge) Type() metrics.MetricType {
	return g.metricType
}
