package retry

import (
	"fmt"
	"time"
)

const DefaultLinearCoefficientScaling = 2

// CalcByLinear calculates the delay duration using a linear function.
//
// The function is based on the general linear equation `y = kx + b`, where:
//   - `y` is the resulting delay in seconds.
//   - `x` (or `attempt` in this context) is the current retry attempt number.
//   - `k` is the coefficient scaling the attempt number.
//   - `b` is the constant term in the linear function.
//
// Parameters:
//   - attempt: The current retry attempt number (equivalent to `x` in the formula).
//   - k: The coefficient applied to the attempt number.
//   - b: The constant term in the linear function.
//
// Returns:
//   - time.Duration: The calculated delay duration.
//
// TODO: Вот это надо реализовать через паттерн итератор, а не функцию.
// TODO: Это позволит инкапсулировать кол-во попыток и не заставлять потребителя думать, какой номер передать.
func CalcByLinear(attempt int, k int, b int) time.Duration {
	return time.Duration(k*attempt+b) * time.Second
}

// WithRetry executes a function with a specified number of retry attempts and linear backoff.
//
// Parameters:
//   - attempts: The total number of retry attempts.
//   - fn: The function to be executed, which should return an error if it fails.
//
// Returns:
//   - error: The last encountered error if all attempts fail, wrapped with context.
func WithRetry(attempts int, fn func() error) (err error) {
	// TODO: Начать принимать и работать с контекстом.
	for i := range attempts {
		if i > 0 {
			time.Sleep(CalcByLinear(i, DefaultLinearCoefficientScaling, 1))
		}

		err = fn()
		if err == nil {
			return nil
		}
		// TODO: Подумать, как логгировать каждую попытку.
	}
	return fmt.Errorf("operation failed after %d attempts: last error: %w", attempts, err)
}
