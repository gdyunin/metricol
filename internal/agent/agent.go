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
	defaultMemStatsUpdateInterval = 1
	resetErrorCountersIntervals   = 4
	maxErrorsToInterrupt          = 3
)

type Fetcher interface {
	Fetch() error
	Metrics() []metrics.Metric
	AddMetrics(newMetrics ...metrics.Metric)
}

type Sender interface {
	Send() error
	RegisterObserver(Observer)
	RemoveObserver(Observer)
}

type Agent struct {
	fetcher Fetcher // The component that fetches metrics.
	sender  Sender  // The component that sends metrics.
}

type errsCounter struct {
	errsCounter uint8
}

func (c *errsCounter) runResetWithInterval(interval uint16) {
	go func() {
		time.Sleep(time.Duration(interval) * time.Second)
		c.errsCounter = 0
	}()
}

type SenderObserver struct {
	actions []func()
}

func (o *SenderObserver) OnNotify() {
	for _, a := range o.actions {
		a()
	}
}

func (o *SenderObserver) addAction(a func()) {
	o.actions = append(o.actions, a)
}

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

func DefaultAgent(cfg *agent.Config) *Agent {
	return NewAgent(cfg, withDefaultMetrics())
}

func (a *Agent) Polling(interval uint16) {
	processStart(
		interval,
		a.fetcher.Fetch,
		"error fetching metric: %s",
		"fetcher accumulated too many errors and was stopped.",
	)
}

func (a *Agent) Reporting(interval uint16) {
	processStart(
		interval,
		a.sender.Send,
		"error sending metric: %s",
		"sender accumulated too many errors and was stopped.",
	)
}

func withDefaultMetrics() func(*Agent) {
	return func(a *Agent) {
		ms := &runtime.MemStats{}
		go func() {
			for {
				runtime.ReadMemStats(ms)
				time.Sleep(time.Duration(defaultMemStatsUpdateInterval) * time.Second)
			}
		}()

		var pollCounter int64 = 0
		pollCounterResetter := &SenderObserver{}
		pollCounterResetter.addAction(func() { atomic.StoreInt64(&pollCounter, 0) })
		a.sender.RegisterObserver(pollCounterResetter)

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

func processStart(interval uint16, fn func() error, formatErrMsg string, interruptMsg string) {
	errorsCounter := &errsCounter{}
	errorsCounter.runResetWithInterval(resetErrorCountersIntervals * interval)

	for {
		time.Sleep(time.Duration(interval) * time.Second)
		if err := fn(); err != nil {
			errorsCounter.errsCounter++
			log.Printf(formatErrMsg, err)
		}
		if errorsCounter.errsCounter >= maxErrorsToInterrupt {
			log.Println(interruptMsg)
			return
		}
	}
}
