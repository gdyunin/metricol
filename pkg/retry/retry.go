package retry

import (
	"fmt"
	"time"
)

func CalcByLinear(attempt int, k int, b int) time.Duration {
	return time.Duration(k*attempt+b) * time.Second
}

func WithRetry(attempts int, fn func() error) (err error) {
	for i := 0; i < attempts; i++ {
		if i > 0 {
			time.Sleep(CalcByLinear(i, 2, 1))
		}

		err = fn()
		if err == nil {
			return
		}
	}
	return fmt.Errorf("all retry attempts failed: %w", err)
}
