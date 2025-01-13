package convert

import (
	"fmt"
	"time"

	"golang.org/x/exp/constraints"
)

// IntegerToSeconds converts an integer representing seconds into a time.Duration.
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
//
// Parameters:
//   - number: A value of any type to be converted to int64.
//
// Returns:
//   - int64: The converted value if successful.
//   - error: An error if the conversion is not possible, including the type of the input value.
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
