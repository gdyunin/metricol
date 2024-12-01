package fetch

import (
	"fmt"
	"testing"

	"github.com/gdyunin/metricol.git/internal/metrics"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMetricsFetcher_AddMetrics tests the AddMetrics method of the MetricsFetcher.
func TestMetricsFetcher_AddMetrics(t *testing.T) {
	tests := []struct {
		name     string           // Name of the test case
		initial  []metrics.Metric // Initial metrics in the fetcher
		toAdd    []metrics.Metric // Metrics to be added
		expected int              // Expected number of metrics after addition
	}{
		{
			name:    "Add single metric to empty fetcher",
			initial: []metrics.Metric{},
			toAdd: []metrics.Metric{
				metrics.NewCounter("test_counter", 42),
			},
			expected: 1,
		},
		{
			name: "Add multiple metrics to existing fetcher",
			initial: []metrics.Metric{
				metrics.NewCounter("test_counter", 4242),
			},
			toAdd: []metrics.Metric{
				metrics.NewCounter("test_counter", 42),
				metrics.NewGauge("test_gauge", 3.14),
			},
			expected: 3,
		},
		{
			name: "Add no metrics",
			initial: []metrics.Metric{
				metrics.NewCounter("test_counter", 4242),
			},
			toAdd:    []metrics.Metric{},
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fetcher := &MetricsFetcher{
				metrics: tt.initial,
			}
			fetcher.AddMetrics(tt.toAdd...)
			require.Equal(t, tt.expected, len(fetcher.metrics))
		})
	}
}

// TestMetricsFetcher_Metrics tests the Metrics method of the MetricsFetcher.
func TestMetricsFetcher_Metrics(t *testing.T) {
	tests := []struct {
		name     string           // Name of the test case
		initial  []metrics.Metric // Initial metrics in the fetcher
		expected []metrics.Metric // Expected metrics to be retrieved
	}{
		{
			name:     "Retrieve metrics from empty fetcher",
			initial:  []metrics.Metric{},
			expected: []metrics.Metric{},
		},
		{
			name: "Retrieve multiple metrics",
			initial: []metrics.Metric{
				metrics.NewCounter("test_counter", 42),
				metrics.NewGauge("test_gauge", 3.14),
			},
			expected: []metrics.Metric{
				metrics.NewCounter("test_counter", 42),
				metrics.NewGauge("test_gauge", 3.14),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fetcher := &MetricsFetcher{
				metrics: tt.initial,
			}
			result := fetcher.Metrics()
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestMetricsFetcher_Fetch tests the Fetch method of the MetricsFetcher.
func TestMetricsFetcher_Fetch(t *testing.T) {
	tests := []struct {
		name          string           // Name of the test case
		initial       []metrics.Metric // Initial metrics in the fetcher
		updateErrors  []error          // Errors expected during metric updates (not used in this test)
		expectedError error            // Expected error after fetching metrics
	}{
		{
			name:          "No metrics",
			initial:       []metrics.Metric{},
			expectedError: nil,
		},
		{
			name: "All metrics update successful",
			initial: []metrics.Metric{
				metrics.NewCounter("test_counter", 42).SetFetcherAndReturn(func() int64 {
					return 42
				}),
				metrics.NewGauge("test_gauge", 3.14).SetFetcherAndReturn(func() float64 {
					return 3.14
				}),
			},
			expectedError: nil,
		},
		{
			name: "One metric update fail",
			initial: []metrics.Metric{
				metrics.NewCounter("test_counter", 42).SetFetcherAndReturn(func() int64 {
					return 42
				}),
				metrics.NewGauge("test_gauge", 3.14), // This gauge does not have a fetcher set
			},
			expectedError: fmt.Errorf(
				"one or more metrics were not fetch: metric %v update fail: error updating metric test_gauge: fetcher not set\n",
				metrics.NewGauge("test_gauge", 3.14),
			),
		},
		{
			name: "Several metrics update fail",
			initial: []metrics.Metric{
				metrics.NewCounter("test_counter", 42), // This counter does not have a fetcher set
				metrics.NewCounter("test_counter2", 42).SetFetcherAndReturn(func() int64 {
					return 42
				}),
				metrics.NewGauge("test_gauge", 3.14), // This gauge does not have a fetcher set
			},
			expectedError: fmt.Errorf(
				"one or more metrics were not fetch: "+
					"metric %v update fail: error updating metric test_counter: fetcher not set\n"+
					"metric %v update fail: error updating metric test_gauge: fetcher not set\n",
				metrics.NewCounter("test_counter", 42),
				metrics.NewGauge("test_gauge", 3.14),
			),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fetcher := &MetricsFetcher{
				metrics: tt.initial,
			}

			err := fetcher.Fetch()

			if err != nil {
				require.Error(t, tt.expectedError)
				require.EqualError(t, err, tt.expectedError.Error())
				return
			}

			require.NoError(t, tt.expectedError) // Ensure no error is expected if err is nil
		})
	}
}
