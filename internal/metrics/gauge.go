package metrics

import (
	"errors"
	"fmt"
	"strconv"
)

type Gauge struct {
	fetcher func() float64
	Name    string
	Value   float64
}

func (g *Gauge) StringValue() string {
	return fmt.Sprintf("%g", g.Value)
}

func (g *Gauge) SetFetcher(f func() float64) {
	g.fetcher = f
}

func (g *Gauge) SetFetcherAndReturn(f func() float64) *Gauge {
	g.SetFetcher(f)
	return g
}

func (g *Gauge) Update() error {
	if g.fetcher == nil {
		return errors.New(ErrorFetcherNotSet)
	}
	g.Value = g.fetcher()
	return nil
}

func newGaugeFromStrings(name, value string) (Metric, error) {
	v, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return nil, errors.New(ErrorParseMetricValue)
	}

	return &Gauge{
		Name:  name,
		Value: v,
	}, nil
}
