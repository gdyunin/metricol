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

// Gauge represents a metric that can take on a range of values.
// It contains a fetcher function to retrieve the current value.
type Gauge struct {
	fetcher func() float64 // Function to fetch the current value
	Name    string         // Name of the metric
	Value   float64        // Current value of the metric
}

// StringName returns the name of the Gauge as a string.
func (g *Gauge) StringName() string {
	return g.Name
}

// StringValue returns the string representation of the current value of the Gauge.
func (g *Gauge) StringValue() string {
	return strconv.FormatFloat(g.Value, 'g', -1, 64)
}

// SetFetcher sets the fetcher function for the Gauge.
func (g *Gauge) SetFetcher(f func() float64) {
	g.fetcher = f
}

// SetFetcherAndReturn sets the fetcher function and returns the Gauge instance.
// This allows for method chaining.
func (g *Gauge) SetFetcherAndReturn(f func() float64) *Gauge {
	g.SetFetcher(f)
	return g
}

// Update updates the current value of the Gauge by calling the fetcher function.
// It returns an error if the fetcher is not set.
func (g *Gauge) Update() error {
	if g.fetcher == nil {
		return fmt.Errorf("error updating metric %s: fetcher not set", g.Name)
	}
	g.Value = g.fetcher() // Update Value using the fetcher function
	return nil
}

// newGaugeFromStrings creates a new Gauge instance from a name and a string representation of a value.
// It returns an error if the value cannot be parsed as a float64.
func newGaugeFromStrings(name, value string) (Metric, error) {
	v, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return nil, fmt.Errorf("error creating gauge from strings: %w", err)
	}

	return &Gauge{
		Name:  name,
		Value: v,
	}, nil
}
