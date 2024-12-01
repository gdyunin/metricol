/*
Package metrics provides functionality for defining and managing various types of metrics,
including counters and gauges. It defines interfaces and structures to facilitate the
creation, updating, and string representation of these metrics.
*/
package metrics

import (
	"fmt"
	"strconv"
)

// Counter represents a metric that counts occurrences.
// It contains a fetcher function to retrieve the current count value.
type Counter struct {
	fetcher func() int64 // Function to fetch the current count value
	Name    string       // Name of the metric
	Value   int64        // Current value of the metric
}

// SetFetcher sets the fetcher function for the Counter.
func (c *Counter) SetFetcher(f func() int64) {
	c.fetcher = f
}

// SetFetcherAndReturn sets the fetcher function and returns the Counter instance.
// This allows for method chaining.
func (c *Counter) SetFetcherAndReturn(f func() int64) *Counter {
	c.SetFetcher(f)
	return c
}

// Update updates the current value of the Counter by calling the fetcher function.
// It returns an error if the fetcher is not set.
func (c *Counter) Update() error {
	if c.fetcher == nil {
		return fmt.Errorf("error updating metric %s: fetcher not set", c.Name)
	}
	c.Value = c.fetcher()
	return nil
}

// StringValue returns the string representation of the current value of the Counter.
func (c *Counter) StringValue() string {
	return strconv.FormatInt(c.Value, 10)
}

// newCounterFromStrings creates a new Counter instance from a name and a string representation of a value.
// It returns an error if the value cannot be parsed as an int64.
func newCounterFromStrings(name, value string) (Metric, error) {
	v, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("error creating counter from strings: %w", err)
	}

	return &Counter{
		Name:  name,
		Value: v,
	}, nil
}
