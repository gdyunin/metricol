package library

import (
	"fmt"
	"github.com/gdyunin/metricol.git/internal/agent/metrics"
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
	return fmt.Sprintf("%g", g.value)
}

func (g *Gauge) UpdateValue() {
	g.value = g.fetcher()
}
