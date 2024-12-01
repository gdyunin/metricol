package metrics

import (
	"errors"
	"strconv"
)

type Counter struct {
	fetcher func() int64
	Name    string
	Value   int64
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
	c.Value = c.fetcher()
	return nil
}

func (c *Counter) StringValue() string {
	return strconv.FormatInt(c.Value, 10)
}

func newCounterFromStrings(name, value string) (Metric, error) {
	v, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return nil, errors.New(ErrorParseMetricValue)
	}

	return &Counter{
		Name:  name,
		Value: v,
	}, nil
}
