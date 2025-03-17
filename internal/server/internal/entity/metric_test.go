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
		{"Valid counter", `{"name":"metric1","type":"counter","value":10}`, false},
		{"Valid gauge", `{"name":"metric2","type":"gauge","value":3.14}`, false},
		{"Invalid counter value", `{"name":"metric3","type":"counter","value":"invalid"}`, true},
		{"Invalid JSON", `{invalid json}`, true},
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
		metrics     Metrics
		expectedLen int
		firstName   string
		firstNil    bool
	}{
		{
			"Multiple metrics", Metrics{
				&Metric{Name: "metric1", Type: "gauge", Value: 1.23},
				&Metric{Name: "metric2", Type: "counter", Value: 42},
			}, 2, "metric1", false},
		{
			"Single metric", Metrics{
				&Metric{Name: "metric1", Type: "gauge", Value: 1.23},
			}, 1, "metric1", false},
		{
			"Empty metrics", Metrics{}, 0, "", true},
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
		metrics  Metrics
		expected string
	}{
		{
			"Single metric",
			Metrics{&Metric{Name: "metric1", Type: "gauge", Value: 1.23}},
			"metric1",
		},
		{
			"Empty metrics",
			Metrics{},
			"",
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
			"Merging counter duplicates",
			Metrics{
				&Metric{Name: "metric1", Type: "counter", Value: 1},
				&Metric{Name: "metric1", Type: "counter", Value: 2},
				&Metric{Name: "metric2", Type: "gauge", Value: 5},
			}, 2, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.metrics.MergeDuplicates()
			assert.Equal(t, tt.expLen, tt.metrics.Length(), "Unexpected length after merging")
			assert.Equal(t, tt.expValue, tt.metrics.First().Value.(int64), "Unexpected value after merging")
		})
	}
}
