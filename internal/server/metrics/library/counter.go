package library

import (
	"errors"
	"fmt"
	"github.com/gdyunin/metricol.git/internal/server/metrics"
	"strconv"
	"strings"
)

type Counter struct {
	name       string
	value      int64
	metricType metrics.MetricType
}

func NewCounter() *Counter {
	return &Counter{
		metricType: metrics.MetricTypeCounter,
	}
}

func (c *Counter) ParseFromURLString(u string) error {
	separated := strings.SplitN(u, "/", 2)

	if len(separated) != 2 {
		return fmt.Errorf(metrics.ErrorParseMetricName)
	}
	c.name = separated[0]

	value, err := strconv.ParseInt(separated[1], 0, 64)
	if err != nil {
		return fmt.Errorf(metrics.ErrorParseMetricValue)
	}
	c.value = value

	return nil
}

func (c Counter) Name() string {
	return c.name
}

func (c *Counter) SetName(name string) {
	c.name = name
}

func (c *Counter) SetValue(val string) error {
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
	return c.metricType
}
