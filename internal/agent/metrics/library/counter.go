package library

import (
	"github.com/gdyunin/metricol.git/internal/agent/metrics"
	"strconv"
)

type Counter struct {
	name    string
	value   int64
	fetcher func() int64
}

func NewCounter(name string, fetcher func() int64) *Counter {
	c := Counter{
		name:    name,
		fetcher: fetcher,
	}
	c.UpdateValue()
	return &c
}

func (c Counter) Type() metrics.MetricType {
	return metrics.MetricTypeCounter
}

func (c Counter) Name() string {
	return c.name
}

func (c Counter) Value() string {
	return strconv.FormatInt(c.value, 10)
}

func (c *Counter) UpdateValue() {
	c.value = c.fetcher()
}
