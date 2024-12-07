package agent

import (
	"log"
	"math/rand"
	"runtime"
	"time"

	"github.com/gdyunin/metricol.git/internal/config/agent"
	"github.com/gdyunin/metricol.git/internal/metrics"
)

const (
	defaultMemStatsUpdateInterval = 1 // defines the default interval for memory statistics updates in seconds.
	resetErrorCountersIntervals   = 4 // defines after how many intervals is the error counter reset.
	maxErrorsToInterrupt          = 3 // defines maximum number of accumulated errors after which execution will abort.
)

// Fetcher defines an interface for fetching metrics.
type Fetcher interface {
	// Fetch retrieves the latest values for the metrics.
	Fetch() error
	// Metrics returns the list of metrics being managed.
	Metrics() []metrics.Metric
	// AddMetrics adds new metrics to the fetcher's collection.
	AddMetrics(newMetrics ...metrics.Metric)
}

// Sender defines an interface for sending metrics.
type Sender interface {
	// Send sends the metrics to the server and returns an error if the operation fails.
	Send() error
	RegisterObserver(Observer)
	RemoveObserver(Observer)
}

// Agent is responsible for fetching and sending metrics.
type Agent struct {
	fetcher Fetcher // The component that fetches metrics.
	sender  Sender  // The component that sends metrics.
}

type SenderObserver struct {
	actions []func()
}

func (o *SenderObserver) OnNotify() { // TODO: make tests for this
	for _, a := range o.actions {
		a()
	}
}
func (o *SenderObserver) addAction(a func()) { // TODO: make tests for this
	o.actions = append(o.actions, a)
}

// NewAgent creates a new Agent instance with the provided configuration.
func NewAgent(cfg *agent.Config) *Agent {
	f := NewMetricsFetcher()
	s := NewMetricsSender(f, cfg.ServerAddress)

	return &Agent{
		fetcher: f,
		sender:  s,
	}
}

// DefaultAgent creates a new Agent with default metrics set.
func DefaultAgent(cfg *agent.Config) *Agent {
	a := NewAgent(cfg)
	a.setDefaultMetrics() // Set default metrics for the agent.
	return a
}

// Polling periodically fetches metrics at the specified interval.
func (a *Agent) Polling(interval int) {
	errsCount := 0
	go func() {
		time.Sleep(resetErrorCountersIntervals * time.Duration(interval) * time.Second)
		errsCount = 0 // Reset errors counter.
	}()

	for {
		time.Sleep(time.Duration(interval) * time.Second)
		if err := a.fetcher.Fetch(); err != nil {
			errsCount++
			log.Printf("error fetching metric: %s", err)
		}
		if errsCount >= maxErrorsToInterrupt {
			log.Println("fetcher accumulated too many errors and was stopped.")
			break // Interrupt if errsCount reached the maxErrorsToInterrupt.
		}
	}
}

// Reporting periodically sends metrics at the specified interval.
func (a *Agent) Reporting(interval int) {
	errsCount := 0
	go func() {
		time.Sleep(resetErrorCountersIntervals * time.Duration(interval) * time.Second)
		errsCount = 0 // Reset errors counter.
	}()

	for {
		time.Sleep(time.Duration(interval) * time.Second)
		if err := a.sender.Send(); err != nil {
			errsCount++
			log.Printf("error sending metric: %s", err)
		}
		if errsCount >= maxErrorsToInterrupt {
			log.Println("sender accumulated too many errors and was stopped.")
			return // Interrupt if errsCount reached the maxErrorsToInterrupt.
		}
	}
}

// setDefaultMetrics initializes default memory metrics for the given fetcher.
func (a *Agent) setDefaultMetrics() {
	ms := &runtime.MemStats{}
	var pollCounter int64 = 0 // TODO: protect against data race (?)

	// Create and setup new observer and subscribe on sender then.
	// TODO: place it in a separate function, e.g. makeObserver()
	senderObserver := &SenderObserver{}
	senderObserver.addAction(func() {
		pollCounter = 0
	})
	a.sender.RegisterObserver(senderObserver)

	// Start the MemStats update in a separate goroutine with default update interval.
	go func() {
		for {
			runtime.ReadMemStats(ms)
			time.Sleep(time.Duration(defaultMemStatsUpdateInterval) * time.Second)
		}
	}()

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
			pollCounter++
			return pollCounter
		}),
	)
}
