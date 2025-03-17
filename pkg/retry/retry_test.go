package retry

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestLinearRetryIterator_Next(t *testing.T) {
	tests := []struct {
		name     string
		k        int
		b        int
		attempts int
		expected []time.Duration
	}{
		{
			name:     "Zero attempt",
			k:        2,
			b:        1,
			attempts: 3,
			expected: []time.Duration{0, 3 * time.Second, 5 * time.Second},
		},
		{
			name:     "Custom values",
			k:        3,
			b:        2,
			attempts: 4,
			expected: []time.Duration{0, 5 * time.Second, 8 * time.Second, 11 * time.Second},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			iterator := NewLinearRetryIterator(tt.k, tt.b)
			for i, expected := range tt.expected {
				actual := iterator.Next()
				assert.Equal(t, expected, actual, "Failed on attempt %d", i)
			}
		})
	}
}

func TestWithRetry(t *testing.T) {
	logger := zap.NewNop().Sugar()

	tests := []struct {
		name         string
		attempts     int
		errorAt      []int
		expectErr    bool
		expectErrMsg string
	}{
		{
			name:      "Success on first attempt",
			attempts:  3,
			errorAt:   make([]int, 0),
			expectErr: false,
		},
		{
			name:         "Fail all attempts",
			attempts:     3,
			errorAt:      []int{0, 1, 2},
			expectErr:    true,
			expectErrMsg: "operation failed after 3 attempts",
		},
		{
			name:      "Fail at some attempts",
			attempts:  3,
			errorAt:   []int{0, 1},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			currentAttempt := 0
			err := WithRetry(ctx, logger, "test_action", tt.attempts, func() error {
				defer func() { currentAttempt++ }()
				for _, errorAt := range tt.errorAt {
					if currentAttempt == errorAt {
						return errors.New("test error")
					}
				}
				return nil
			})

			if tt.expectErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectErrMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
