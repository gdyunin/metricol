// Package convert provides utility functions for numeric conversions.
// It includes functions to convert integer values representing seconds into time.Duration
// and to convert values of various numeric types to int64. The functions are designed to be simple,
// ensuring that the project has a straightforward mechanism for numeric conversions.
package convert

import (
	"fmt"
	"time"

	"golang.org/x/exp/constraints"
)

// IntegerToSeconds converts an integer representing seconds into a time.Duration.
// It multiplies the provided integer by time.Second.
//
// Parameters:
//   - seconds: An integer value representing seconds.
//
// Returns:
//   - time.Duration: The equivalent duration in seconds.
func IntegerToSeconds[T constraints.Integer](seconds T) time.Duration {
	return time.Duration(seconds) * time.Second
}

// AnyToInt64 converts various numeric types to int64.
// It supports int64, float64, int, and uint types. If the conversion is unsupported,
// it returns an error indicating the input value's type.
//
// Parameters:
//   - number: A value of any type to be converted to int64.
//
// Returns:
//   - int64: The converted value if successful.
//   - error: An error if the conversion is not possible.
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
		return 0, fmt.Errorf("type %T is unsupported for conversion to int64", v)
	}
}
