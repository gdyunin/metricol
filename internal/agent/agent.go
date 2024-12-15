// Package agent provides functionality for sending metrics to a server.
package agent

import (
	"log"
	"math/rand"
	"runtime"
	"sync/atomic"
	"time"

	"github.com/gdyunin/metricol.git/internal/config/agent"
	"github.com/gdyunin/metricol.git/internal/metrics"
)

const (
	defaultMemStatsUpdateInterval = 1 // Default interval for updating memory statistics in seconds.
	resetErrorCountersIntervals   = 4 // Interval for resetting error counters in seconds.
	maxErrorsToInterrupt          = 3 // Maximum number of errors before interruption occurs.
)

// Fetcher defines the interface for fetching metrics.
type Fetcher interface {
	Fetch() error                            // Fetch retrieves metrics and returns any error encountered.
	Metrics() []metrics.Metric               // Metrics returns a slice of metrics collected by the fetcher.
	AddMetrics(newMetrics ...metrics.Metric) // AddMetrics allows adding new metrics to the fetcher.
}

// Sender defines the interface for sending metrics.
type Sender interface {
	Send() error               // Send transmits the metrics and returns any error encountered.
	RegisterObserver(Observer) // RegisterObserver adds an observer to be notified on events.
	RemoveObserver(Observer)   // RemoveObserver removes an observer from notifications.
}

// Agent is responsible for fetching and sending metrics.
type Agent struct {
	fetcher Fetcher // The component that fetches metrics.
	sender  Sender  // The component that sends metrics.
}

// errsCounter keeps track of the number of errors encountered.
type errsCounter struct {
	errsCounter uint8 // Count of errors encountered.
}

// runResetWithInterval starts a goroutine that resets the error counter after a specified interval.
func (c *errsCounter) runResetWithInterval(interval uint16) {
	go func() {
		time.Sleep(time.Duration(interval) * time.Second)
		c.errsCounter = 0 // Reset error counter to zero after the interval.
	}()
}

// SenderObserver observes actions and notifies when an event occurs.
type SenderObserver struct {
	actions []func() // List of actions to be executed upon notification.
}

// OnNotify executes all registered actions when an event occurs.
func (o *SenderObserver) OnNotify() {
	for _, a := range o.actions {
		a() // Execute each action in the list.
	}
}

// addAction registers a new action to be executed upon notification.
func (o *SenderObserver) addAction(a func()) {
	o.actions = append(o.actions, a)
}

// NewAgent creates a new Agent with the provided configuration and options.
func NewAgent(cfg *agent.Config, options ...func(*Agent)) *Agent {
	f := NewMetricsFetcher()
	s := NewMetricsSender(f, cfg.ServerAddress)
	a := &Agent{
		fetcher: f,
		sender:  s,
	}

	for _, o := range options {
		o(a)
	}

	return a
}

// DefaultAgent creates a new Agent with default configuration options.
func DefaultAgent(cfg *agent.Config) *Agent {
	return NewAgent(cfg, withDefaultMetrics())
}

// Polling starts polling for metrics at the specified interval.
func (a *Agent) Polling(interval uint16) {
	processStart(
		interval,
		a.fetcher.Fetch,             // Fetch metrics at the specified interval.
		"error fetching metric: %s", // Error message format for fetching errors.
		"fetcher accumulated too many errors and was stopped.", // Message when fetcher is stopped due to errors.
	)
}

// Reporting starts reporting metrics at the specified interval.
func (a *Agent) Reporting(interval uint16) {
	processStart(
		interval,
		a.sender.Send,              // Send metrics at the specified interval.
		"error sending metric: %s", // Error message format for sending errors.
		"sender accumulated too many errors and was stopped.", // Message when sender is stopped due to errors.
	)
}

// withDefaultMetrics returns a function that configures the provided Agent
// with default metrics related to memory statistics and polling counts.
//
// The returned function sets up a background goroutine that periodically
// reads memory statistics and updates metrics accordingly. It also registers
// a poll counter that tracks how many times certain actions have been performed.
func withDefaultMetrics() func(*Agent) {
	return func(a *Agent) {
		ms := &runtime.MemStats{} // Create a MemStats instance to hold memory statistics.

		// Start a goroutine to periodically read memory statistics.
		go func() {
			for {
				runtime.ReadMemStats(ms)
				time.Sleep(time.Duration(defaultMemStatsUpdateInterval) * time.Second)
			}
		}()

		var pollCounter int64 = 0                // Initialize a counter to track the number of polls.
		pollCounterResetter := &SenderObserver{} // Create an observer for resetting the poll counter.

		// Add an action to reset the poll counter to zero when notified.
		pollCounterResetter.addAction(func() { atomic.StoreInt64(&pollCounter, 0) })
		a.sender.RegisterObserver(pollCounterResetter) // Register the observer with the agent's sender.

		// Add various metrics related to memory statistics and polling count.
		a.fetcher.AddMetrics(
			metrics.NewGauge("Alloc", 0).SetFetcherAndReturn(func() float64 {
				return float64(ms.Alloc)
			}),
			metrics.NewGauge("BuckHashSys", 0).SetFetcherAndReturn(func() float64 {
				return float64(ms.BuckHashSys)
			}),
			metrics.NewGauge("Frees", 0).SetFetcherAndReturn(func() float64 {
				return float64(ms.Frees)
			}),
			metrics.NewGauge("GCCPUFraction", 0).SetFetcherAndReturn(func() float64 {
				return ms.GCCPUFraction
			}),
			metrics.NewGauge("GCSys", 0).SetFetcherAndReturn(func() float64 {
				return float64(ms.GCSys)
			}),
			metrics.NewGauge("HeapAlloc", 0).SetFetcherAndReturn(func() float64 {
				return float64(ms.HeapAlloc)
			}),
			metrics.NewGauge("HeapIdle", 0).SetFetcherAndReturn(func() float64 {
				return float64(ms.HeapIdle)
			}),
			metrics.NewGauge("HeapInuse", 0).SetFetcherAndReturn(func() float64 {
				return float64(ms.HeapInuse)
			}),
			metrics.NewGauge("HeapObjects", 0).SetFetcherAndReturn(func() float64 {
				return float64(ms.HeapObjects)
			}),
			metrics.NewGauge("HeapReleased", 0).SetFetcherAndReturn(func() float64 {
				return float64(ms.HeapReleased)
			}),
			metrics.NewGauge("HeapSys", 0).SetFetcherAndReturn(func() float64 {
				return float64(ms.HeapSys)
			}),
			metrics.NewGauge("LastGC", 0).SetFetcherAndReturn(func() float64 {
				return float64(ms.LastGC)
			}),
			metrics.NewGauge("Lookups", 0).SetFetcherAndReturn(func() float64 {
				return float64(ms.Lookups)
			}),
			metrics.NewGauge("MCacheInuse", 0).SetFetcherAndReturn(func() float64 {
				return float64(ms.MCacheInuse)
			}),
			metrics.NewGauge("MCacheSys", 0).SetFetcherAndReturn(func() float64 {
				return float64(ms.MCacheSys)
			}),
			metrics.NewGauge("MSpanInuse", 0).SetFetcherAndReturn(func() float64 {
				return float64(ms.MSpanInuse)
			}),
			metrics.NewGauge("MSpanSys", 0).SetFetcherAndReturn(func() float64 {
				return float64(ms.MSpanSys)
			}),
			metrics.NewGauge("Mallocs", 0).SetFetcherAndReturn(func() float64 {
				return float64(ms.Mallocs)
			}),
			metrics.NewGauge("NextGC", 0).SetFetcherAndReturn(func() float64 {
				return float64(ms.NextGC)
			}),
			metrics.NewGauge("NumForcedGC", 0).SetFetcherAndReturn(func() float64 {
				return float64(ms.NumForcedGC)
			}),
			metrics.NewGauge("NumGC", 0).SetFetcherAndReturn(func() float64 {
				return float64(ms.NumGC)
			}),
			metrics.NewGauge("OtherSys", 0).SetFetcherAndReturn(func() float64 {
				return float64(ms.OtherSys)
			}),
			metrics.NewGauge("PauseTotalNs", 0).SetFetcherAndReturn(func() float64 {
				return float64(ms.PauseTotalNs)
			}),
			metrics.NewGauge("StackInuse", 0).SetFetcherAndReturn(func() float64 {
				return float64(ms.StackInuse)
			}),
			metrics.NewGauge("StackSys", 0).SetFetcherAndReturn(func() float64 {
				return float64(ms.StackSys)
			}),
			metrics.NewGauge("Sys", 0).SetFetcherAndReturn(func() float64 {
				return float64(ms.Sys)
			}),
			metrics.NewGauge("TotalAlloc", 0).SetFetcherAndReturn(func() float64 {
				return float64(ms.TotalAlloc)
			}),
			metrics.NewGauge("RandomValue", 0).SetFetcherAndReturn(rand.Float64),
			metrics.NewCounter("PollCount", 0).SetFetcherAndReturn(func() int64 {
				atomic.AddInt64(&pollCounter, 1)
				return pollCounter
			}),
		)
	}
}

// processStart executes a given function at specified intervals,
// while tracking errors and handling interruptions based on error counts.
//
// Parameters:
// - interval: The time interval (in seconds) between function executions.
// - fn: The function to execute, which returns an error if it fails.
// - formatErrMsg: A format string for logging errors that occur during function execution.
// - interruptMsg: A message to log when the maximum error threshold is reached.
func processStart(interval uint16, fn func() error, formatErrMsg string, interruptMsg string) {
	errorsCounter := &errsCounter{}

	// Start a routine to reset the error counter at defined intervals.
	errorsCounter.runResetWithInterval(resetErrorCountersIntervals * interval)

	for {
		time.Sleep(time.Duration(interval) * time.Second) // Wait for the specified interval before executing the function.

		if err := fn(); err != nil { // Execute the provided function and check for errors.
			errorsCounter.errsCounter++ // Increment the error counter if an error occurred.
			log.Printf(formatErrMsg, err)
		}

		// Check if the number of errors has reached the maximum allowed before interruption.
		if errorsCounter.errsCounter >= maxErrorsToInterrupt {
			log.Println(interruptMsg)
			return // Exit the function if the maximum error count is exceeded.
		}
	}
}
