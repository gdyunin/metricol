package convert

import (
	"fmt"
	"time"

	"golang.org/x/exp/constraints"
)

func IntegerToSeconds[T constraints.Integer](seconds T) time.Duration {
	return time.Duration(seconds) * time.Second
}

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
