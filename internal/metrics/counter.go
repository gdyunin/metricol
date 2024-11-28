package metrics

import (
	"errors"
	"strconv"
)

type Counter struct {
	fetcher func() int64
	name    string
	value   int64
}

func (c *Counter) Name() string {
	return c.name
}

func (c *Counter) Value() int64 {
	return c.value
}

func (c *Counter) Type() string {
	return MetricTypeCounter
}

func (c *Counter) SetFetcher(f func() int64) {
	c.fetcher = f
}

func (c *Counter) SetFetcherAndReturn(f func() int64) *Counter {
	c.SetFetcher(f)
	return c
}

func (c *Counter) Update() error {
	if c.fetcher == nil {
		return errors.New(ErrorFetcherNotSet)
	}
	c.value += c.fetcher()
	return nil
}

func (c *Counter) StringValue() string {
	return strconv.FormatInt(c.value, 10)
}

func newCounterFromStrings(name, value string) (Metric, error) {
	v, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return nil, errors.New(ErrorParseMetricValue)
	}

	return &Counter{
		name:  name,
		value: v,
	}, nil
}
