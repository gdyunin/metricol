package library

import (
	"github.com/gdyunin/metricol.git/internal/agent/metrics"
	"strconv"
)

type Gauge struct {
	name    string
	value   float64
	fetcher func() float64
}

func NewGauge(name string, fetcher func() float64) *Gauge {
	g := Gauge{
		name:    name,
		fetcher: fetcher,
	}
	g.UpdateValue()
	return &g
}

func (g Gauge) Type() metrics.MetricType {
	return metrics.MetricTypeGauge
}

func (g Gauge) Name() string {
	return g.name
}

func (g Gauge) Value() string {
	return strconv.FormatFloat(g.value, 'f', 12, 64)
}

func (g *Gauge) UpdateValue() {
	g.value = g.fetcher()
}
