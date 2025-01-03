package helpers

import (
	"errors"
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
		// Return an error if the limit is not a positive non-zero value.
		return nil, errors.New("the limit for errors must be a positive non-zero value")
	}

	// Initialize a new Interrupter with a ticker and error channel.
	i := &Interrupter{
		C:     make(chan bool, 1), // Buffered channel to avoid blocking.
		t:     time.NewTicker(interval),
		limit: limit,
	}

	// StartAll a goroutine to reset the error counter periodically.
	i.startResetter()

	return i, nil
}

// Stop stops the ticker, halting the periodic error counter resets.
func (i *Interrupter) Stop() {
	// Ensure the ticker is stopped to release resources.
	i.t.Stop()
}

// AddError increments the error counter.
// If the error count exceeds the defined limit, an interrupt signal is sent through the channel.
func (i *Interrupter) AddError() {
	// Increment the error counter.
	i.errors++
	// Send an interrupt signal if the error limit is exceeded.
	if !i.InLimit() {
		i.C <- true
	}
}

// InLimit checks whether the current error count is within the allowed limit.
// Returns true if the error count is below the limit, false otherwise.
func (i *Interrupter) InLimit() bool {
	// Compare the current error count with the defined limit.
	return i.errors < i.limit
}

// startResetter starts a goroutine that resets the error counter periodically
// based on the interval defined by the ticker.
func (i *Interrupter) startResetter() {
	go func() {
		for range i.t.C {
			// Reset the error counter to zero at each interval.
			i.errors = 0
		}
	}()
}
