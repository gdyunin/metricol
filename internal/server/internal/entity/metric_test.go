package entity

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnmarshalJSON(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expectErr bool
	}{
		{name: "Valid counter", input: `{"name":"metric1","type":"counter","value":10}`},
		{name: "Valid gauge", input: `{"name":"metric2","type":"gauge","value":3.14}`},
		{
			name:      "Invalid counter value",
			input:     `{"name":"metric3","type":"counter","value":"invalid"}`,
			expectErr: true,
		},
		{name: "Invalid JSON", input: `{invalid json}`, expectErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var metric Metric
			err := json.Unmarshal([]byte(tt.input), &metric)
			if tt.expectErr {
				assert.Error(t, err, "Expected error for input %s", tt.input)
			} else {
				assert.NoError(t, err, "Did not expect error for input %s", tt.input)
			}
		})
	}
}

func TestMetrics_Functions(t *testing.T) {
	tests := []struct {
		name        string
		firstName   string
		metrics     Metrics
		expectedLen int
		firstNil    bool
	}{
		{
			name: "Multiple metrics", metrics: Metrics{
				&Metric{Name: "metric1", Type: "gauge", Value: 1.23},
				&Metric{Name: "metric2", Type: "counter", Value: 42},
			}, expectedLen: 2, firstName: "metric1"},
		{
			name: "Single metric", metrics: Metrics{
				&Metric{Name: "metric1", Type: "gauge", Value: 1.23},
			}, expectedLen: 1, firstName: "metric1"},
		{
			name: "Empty metrics", metrics: Metrics{}, firstNil: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expectedLen, tt.metrics.Length(), "Length mismatch")
			first := tt.metrics.First()
			if tt.firstNil {
				assert.Nil(t, first, "Expected first metric to be nil")
			} else {
				assert.NotNil(t, first, "First metric should not be nil")
				assert.Equal(t, tt.firstName, first.Name, "First metric name mismatch")
			}
		})
	}
}

func TestMetrics_String(t *testing.T) {
	tests := []struct {
		name     string
		expected string
		metrics  Metrics
	}{
		{
			name:     "Single metric",
			metrics:  Metrics{&Metric{Name: "metric1", Type: "gauge", Value: 1.23}},
			expected: "metric1",
		},
		{
			name:    "Empty metrics",
			metrics: Metrics{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Contains(t, tt.metrics.String(), tt.expected, "String representation mismatch")
		})
	}
}

func TestMergeDuplicates(t *testing.T) {
	tests := []struct {
		name     string
		metrics  Metrics
		expLen   int
		expValue int64
	}{
		{
			name: "Merging counter duplicates",
			metrics: Metrics{
				&Metric{Name: "metric1", Type: "counter", Value: 1},
				&Metric{Name: "metric1", Type: "counter", Value: 2},
				&Metric{Name: "metric2", Type: "gauge", Value: 5},
			}, expLen: 2, expValue: 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.metrics.MergeDuplicates()
			assert.Equal(t, tt.expLen, tt.metrics.Length(), "Unexpected length after merging")
			assert.Equal(
				t,
				tt.expValue,
				tt.metrics.First().Value.(int64), //nolint:errcheck,forcetypeassert // for tests
				"Unexpected value after merging",
			)
		})
	}
}
