package convert

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestIntegerToSeconds(t *testing.T) {
	tests := []struct {
		name     string
		input    int
		expected time.Duration
	}{
		{name: "Zero seconds"},
		{name: "One second", input: 1, expected: 1 * time.Second},
		{name: "Ten seconds", input: 10, expected: 10 * time.Second},
		{name: "Negative value", input: -5, expected: -5 * time.Second},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := IntegerToSeconds(tt.input)
			assert.Equal(t, tt.expected, actual, "Failed for input %d", tt.input)
		})
	}
}

func TestAnyToInt64(t *testing.T) {
	tests := []struct {
		input     interface{}
		name      string
		expected  int64
		expectErr bool
	}{
		{name: "Valid int64", input: int64(42), expected: 42},
		{name: "Valid float64", input: float64(42.9), expected: 42},
		{name: "Valid int", input: int(10), expected: 10},
		{name: "Valid uint", input: uint(20), expected: 20},
		{name: "Invalid string", input: "invalid", expectErr: true},
		{name: "Invalid struct", input: struct{}{}, expectErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := AnyToInt64(tt.input)

			if tt.expectErr {
				assert.Error(t, err, "Expected error for input %v", tt.input)
			} else {
				assert.NoError(t, err, "Did not expect error for input %v", tt.input)
				assert.Equal(t, tt.expected, actual, "Failed for input %v", tt.input)
			}
		})
	}
}
