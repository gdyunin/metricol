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
		{"Zero seconds", 0, 0},
		{"One second", 1, 1 * time.Second},
		{"Ten seconds", 10, 10 * time.Second},
		{"Negative value", -5, -5 * time.Second},
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
		name      string
		input     interface{}
		expected  int64
		expectErr bool
	}{
		{"Valid int64", int64(42), 42, false},
		{"Valid float64", float64(42.9), 42, false},
		{"Valid int", int(10), 10, false},
		{"Valid uint", uint(20), 20, false},
		{"Invalid string", "invalid", 0, true},
		{"Invalid struct", struct{}{}, 0, true},
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
