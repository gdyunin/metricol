package common

import (
	"fmt"
	"time"
)

// Interrupter is a mechanism for tracking errors and triggering an interrupt signal
// when a specified error limit is exceeded within a given interval.
type Interrupter struct {
	C      chan bool    // Channel used to signal when the error limit is exceeded.
	t      *time.Ticker // Ticker for resetting the error counter periodically.
	errors uint8        // Current count of errors tracked by the interrupter.
	limit  uint8        // Maximum allowed errors before triggering an interrupt.
}

// NewInterrupter creates a new Interrupter instance.
// The interval specifies the duration between error counter resets,
// and the limit defines the maximum allowed errors before triggering an interrupt.
// Returns an error if the limit is less than or equal to zero.
func NewInterrupter(interval time.Duration, limit uint8) (*Interrupter, error) {
	if limit <= 0 {
		return nil, fmt.Errorf("the limit for errors must be a positive non-zero value")
	}

	i := &Interrupter{
		C:     make(chan bool, 1),
		t:     time.NewTicker(interval),
		limit: limit,
	}

	i.startResetter()

	return i, nil
}

// Stop stops the ticker, halting the periodic error counter resets.
func (i *Interrupter) Stop() {
	i.t.Stop()
}

// AddError increments the error counter.
// If the error count exceeds the defined limit, an interrupt signal is sent through the channel.
func (i *Interrupter) AddError() {
	i.errors++
	if !i.InLimit() {
		i.C <- true
	}
}

// InLimit checks whether the current error count is within the allowed limit.
// Returns true if the error count is below the limit, false otherwise.
func (i *Interrupter) InLimit() bool {
	return i.errors < i.limit
}

// startResetter starts a goroutine that resets the error counter periodically
// based on the interval defined by the ticker.
func (i *Interrupter) startResetter() {
	go func() {
		for range i.t.C {
			i.errors = 0
		}
	}()
}
