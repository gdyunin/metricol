package common

import "sync"

// RunRoutinesWithWG starts multiple functions as goroutines and waits for them to complete.
// Each function is executed in its own goroutine, and the wait group is used to ensure
// that the main function waits for all routines to finish before proceeding.
func RunRoutinesWithWG(wg *sync.WaitGroup, fn ...func()) {
	for _, f := range fn {
		go func(f func()) {
			defer wg.Done()
			f()
		}(f) // Pass f as an argument to avoid closure issues
	}
}
