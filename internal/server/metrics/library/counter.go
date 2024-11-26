package library

import (
	"errors"
	"github.com/gdyunin/metricol.git/internal/server/metrics"
	"strconv"
)

type Counter struct {
	name  string
	value int64
}

func NewCounter() *Counter {
	return &Counter{}
}

func (c Counter) Name() string {
	return c.name
}

func (c *Counter) SetName(name string) error {
	if name == "" {
		return errors.New(metrics.ErrorEmptyName)
	}
	c.name = name
	return nil
}

func (c *Counter) SetValue(val string) error {
	if val == "" {
		return errors.New(metrics.ErrorEmptyValue)
	}

	v, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return errors.New(metrics.ErrorParseMetricValue)
	}

	c.value = v
	return nil
}

func (c Counter) Value() string {
	return strconv.FormatInt(c.value, 10)
}

func (c Counter) Type() metrics.MetricType {
	return metrics.MetricTypeCounter
}
