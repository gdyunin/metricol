package retry

import "time"

func CalcByLinear(attempt int, k int, b int) time.Duration {
	return time.Duration(k*attempt+b) * time.Second
}
