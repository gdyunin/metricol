package utils

import (
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
