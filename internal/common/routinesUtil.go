package common

import "sync"

// RunRoutinesWithWG starts multiple functions as goroutines and waits for them to complete.
// It takes a wait group and a variadic list of functions to be executed concurrently.
// Each function is executed in its own goroutine, and the wait group is used to ensure
// that the main function waits for all routines to finish before proceeding.
//
// Parameters:
//
//	wg: A pointer to a sync.WaitGroup that is used to synchronize the completion of the goroutines.
//	fn: A variadic list of functions to be executed concurrently.
func RunRoutinesWithWG(wg *sync.WaitGroup, fn ...func()) {
	for _, f := range fn {
		go func() {
			defer wg.Done() // Decrement the wait group counter when the goroutine completes.
			f()             // Execute the function.
		}()
	}
}
