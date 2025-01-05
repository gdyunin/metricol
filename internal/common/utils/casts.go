package utils

import (
	"fmt"
	"time"

	"golang.org/x/exp/constraints"
)

// IntegerToSeconds converts an integer value to a time.Duration representing seconds.
//
// Parameters:
//   - interval: An integer value representing the interval in seconds.
//
// Returns:
//   - A time.Duration representing the input interval in seconds.
func IntegerToSeconds[T constraints.Integer](interval T) time.Duration {
	return time.Duration(interval) * time.Second
}

// AnyToInt64 converts an any value contains number to int64.
//
// Parameters:
//   - number: An any number value to be converted to int64.
//
// Returns:
//   - value after casting to int64.
func AnyToInt64[T any](number T) (int64, error) {
	switch v := any(number).(type) {
	case int64:
		return v, nil
	case float64:
		return int64(v), nil
	case int:
		return int64(v), nil
	case uint:
		return int64(v), nil
	default:
		return 0, fmt.Errorf("cannot convert type %T to int64", v)
	}
}
