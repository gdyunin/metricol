package main

import (
	"github.com/gdyunin/metricol.git/internal/agent/fetch"
	"github.com/gdyunin/metricol.git/internal/agent/send"
	"github.com/gdyunin/metricol.git/internal/metrics"
	"math/rand"
	"runtime"
	"time"
)

func main() {
	appCfg := appConfig()

	fetcher := fetch.NewMetricsFetcher()
	sender := send.NewMetricsSender(fetcher, appCfg.ServerAddress)

	ms := &runtime.MemStats{}
	fetcher.AddMetrics(
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
		metrics.NewGauge("RandomValue", 0).SetFetcherAndReturn(func() float64 {
			return rand.Float64()
		}),
		metrics.NewCounter("PollCount", 0).SetFetcherAndReturn(func() int64 {
			return 1
		}),
	)

	go func() {
		for {
			time.Sleep(time.Duration(appCfg.PollInterval) * time.Second)
			runtime.ReadMemStats(ms)
			fetcher.Fetch()
		}
	}()

	for {
		time.Sleep(time.Duration(appCfg.ReportInterval) * time.Second)
		sender.Send()
	}
}
