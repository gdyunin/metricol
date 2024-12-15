package common

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestRunRoutinesWithWG tests the RunRoutinesWithWG function to ensure that it correctly
// starts multiple goroutines, executes the provided functions, and properly synchronizes
// their completion using the provided WaitGroup. The test cases cover scenarios with
// zero, one, and multiple functions, as well as functions that include delays to simulate
// concurrent execution.
func TestRunRoutinesWithWG(t *testing.T) {
	// Define the test cases using a table-driven approach.
	tests := []struct {
		name           string   // Description of the test case.
		functions      []func() // Slice of functions to be executed concurrently.
		initialWGCount int      // Initial count to add to the WaitGroup.
		expectedCalls  int      // Expected number of function calls.
	}{
		{
			name:           "No functions passed.",
			functions:      []func(){},
			initialWGCount: 0,
			expectedCalls:  0,
		},
		{
			name: "Single function passed.",
			functions: []func(){
				func() {},
			},
			initialWGCount: 1,
			expectedCalls:  1,
		},
		{
			name: "Multiple functions passed.",
			functions: []func(){
				func() {},
				func() {},
				func() {},
			},
			initialWGCount: 3,
			expectedCalls:  3,
		},
		{
			name: "Functions with delays to simulate concurrency.",
			functions: []func(){
				func() { time.Sleep(100 * time.Millisecond) },
				func() { time.Sleep(200 * time.Millisecond) },
				func() { time.Sleep(150 * time.Millisecond) },
			},
			initialWGCount: 3,
			expectedCalls:  3,
		},
	}

	// Iterate over each test case.
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Initialize a WaitGroup and set its counter.
			var wg sync.WaitGroup
			wg.Add(tc.initialWGCount)

			// Initialize a counter to track the number of function calls.
			var callCounter int
			var counterMutex sync.Mutex

			// Wrap each function to increment the callCounter.
			wrappedFunctions := make([]func(), len(tc.functions))
			for i, f := range tc.functions {
				wrappedFunctions[i] = func(fn func()) func() {
					return func() {
						fn() // Execute the original function.
						// Safely increment the call counter.
						counterMutex.Lock()
						defer counterMutex.Unlock()
						callCounter++
					}
				}(f)
			}

			// Execute RunRoutinesWithWG with the wrapped functions.
			RunRoutinesWithWG(&wg, wrappedFunctions...)

			// Wait for all goroutines to complete.
			wg.Wait()

			// Assert that the expected number of functions were called.
			assert.Equal(t, tc.expectedCalls, callCounter, "Number of function calls should match the expected value")
		})
	}
}
