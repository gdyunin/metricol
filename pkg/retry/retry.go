package retry

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
)

const DefaultLinearCoefficientScaling = 2

// LinearRetryIterator provides an iterator to calculate delay durations
// for retry attempts using a linear function.
//
// The linear function is based on the general equation `y = kx + b`, where:
//   - `y` is the resulting delay in seconds.
//   - `x` (or `attempt` in this context) is the current retry attempt number.
//   - `k` is the coefficient scaling the attempt number.
//   - `b` is the constant term in the linear function.
//
// This iterator maintains the current retry attempt state and calculates
// the delay dynamically with each call to `Next()`.
//
// Fields:
//   - attempt: The current retry attempt number (equivalent to `x` in the formula).
//   - k: The coefficient applied to the attempt number.
//   - b: The constant term in the linear function.
type LinearRetryIterator struct {
	attempt int // Tracks the current retry attempt number.
	k       int // Scaling coefficient for the linear function.
	b       int // Constant term in the linear function.
}

// NewLinearRetryIterator creates and initializes a new LinearRetryIterator.
//
// Parameters:
//   - k: The coefficient scaling the attempt number.
//   - b: The constant term in the linear function.
//
// Returns:
//   - *LinearRetryIterator: A pointer to the initialized LinearRetryIterator.
func NewLinearRetryIterator(k int, b int) *LinearRetryIterator {
	return &LinearRetryIterator{k: k, b: b}
}

// Next calculates the delay duration for the current retry attempt
// and increments the attempt counter.
//
// The delay is computed using the linear equation `y = kx + b`, where:
//   - `y` is the delay duration in seconds.
//   - `x` is the current retry attempt number.
//
// Returns:
//   - time.Duration: The calculated delay duration for the current attempt.
//     If this is the first call (attempt == 0), the delay is 0 seconds.
func (i *LinearRetryIterator) Next() time.Duration {
	if i.attempt > 0 {
		interval := time.Duration(i.k*i.attempt+i.b) * time.Second
		i.attempt++
		return interval
	}
	i.attempt++
	return 0
}

// SetCurrentAttempt sets the current retry attempt to a specified value.
//
// Parameters:
//   - currentAttempt: The retry attempt number to set as the current attempt.
func (i *LinearRetryIterator) SetCurrentAttempt(currentAttempt int) {
	i.attempt = currentAttempt
}

// WithRetry executes a function with a specified number of retry attempts and linear backoff.
//
// Parameters:
//   - attempts: The total number of retry attempts.
//   - fn: The function to be executed, which should return an error if it fails.
//
// Returns:
//   - error: The last encountered error if all attempts fail, wrapped with context.
func WithRetry(
	ctx context.Context,
	logger *zap.SugaredLogger,
	actMsg string,
	attempts int,
	fn func() error,
) (err error) {
	intervalIterator := NewLinearRetryIterator(DefaultLinearCoefficientScaling, -1)
	for i := range attempts {
		time.Sleep(intervalIterator.Next())

		select {
		case <-ctx.Done():
			return fmt.Errorf("the retring process was interrupted at attempt %d: the context has expired", i)
		default:
			if err = fn(); err == nil {
				return nil
			}
			logger.Infof(
				"Attempt %d for action <%s> ended in error, move on to the next attempt...",
				i,
				actMsg,
			)
		}
	}
	return fmt.Errorf("operation failed after %d attempts: last error: %w", attempts, err)
}
