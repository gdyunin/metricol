package utils

import (
	"time"

	"golang.org/x/exp/constraints"
)

func IntegerToSeconds[T constraints.Integer](interval T) time.Duration {
	return time.Duration(interval) * time.Second
}
