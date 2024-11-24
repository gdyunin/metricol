package main

import (
	"github.com/gdyunin/metricol.git/internal/agent/fetch"
	"github.com/gdyunin/metricol.git/internal/agent/metrics/library"
	"github.com/gdyunin/metricol.git/internal/agent/send"
	"log"
	"math/rand"
	"runtime"
	"time"
)

func main() {
	parseFlags()
	pollPeriod := time.Duration(flagPollInterval) * time.Second
	reportInterval := time.Duration(flagReportInterval) * time.Second

	storage := fetch.NewStorage()
	sender := send.NewClient(storage, "localhost", 8080)

	ms := runtime.MemStats{}

	storage.AddMetrics(
		library.NewGauge("Alloc", func() float64 {
			runtime.ReadMemStats(&ms)
			return float64(ms.Alloc)
		}),
		library.NewGauge("BuckHashSys", func() float64 {
			runtime.ReadMemStats(&ms)
			return float64(ms.BuckHashSys)
		}),
		library.NewGauge("Frees", func() float64 {
			runtime.ReadMemStats(&ms)
			return float64(ms.Frees)
		}),
		library.NewGauge("GCCPUFraction", func() float64 {
			runtime.ReadMemStats(&ms)
			return ms.GCCPUFraction
		}),
		library.NewGauge("GCSys", func() float64 {
			runtime.ReadMemStats(&ms)
			return float64(ms.GCSys)
		}),
		library.NewGauge("HeapAlloc", func() float64 {
			runtime.ReadMemStats(&ms)
			return float64(ms.HeapAlloc)
		}),
		library.NewGauge("HeapIdle", func() float64 {
			runtime.ReadMemStats(&ms)
			return float64(ms.HeapIdle)
		}),
		library.NewGauge("HeapInuse", func() float64 {
			runtime.ReadMemStats(&ms)
			return float64(ms.HeapInuse)
		}),
		library.NewGauge("HeapObjects", func() float64 {
			runtime.ReadMemStats(&ms)
			return float64(ms.HeapObjects)
		}),
		library.NewGauge("HeapReleased", func() float64 {
			runtime.ReadMemStats(&ms)
			return float64(ms.HeapReleased)
		}),
		library.NewGauge("HeapSys", func() float64 {
			runtime.ReadMemStats(&ms)
			return float64(ms.HeapSys)
		}),
		library.NewGauge("LastGC", func() float64 {
			runtime.ReadMemStats(&ms)
			return float64(ms.LastGC)
		}),
		library.NewGauge("Lookups", func() float64 {
			runtime.ReadMemStats(&ms)
			return float64(ms.Lookups)
		}),
		library.NewGauge("MCacheInuse", func() float64 {
			runtime.ReadMemStats(&ms)
			return float64(ms.MCacheInuse)
		}),
		library.NewGauge("MCacheSys", func() float64 {
			runtime.ReadMemStats(&ms)
			return float64(ms.MCacheSys)
		}),
		library.NewGauge("MSpanInuse", func() float64 {
			runtime.ReadMemStats(&ms)
			return float64(ms.MSpanInuse)
		}),
		library.NewGauge("MSpanSys", func() float64 {
			runtime.ReadMemStats(&ms)
			return float64(ms.MSpanSys)
		}),
		library.NewGauge("Mallocs", func() float64 {
			runtime.ReadMemStats(&ms)
			return float64(ms.Mallocs)
		}),
		library.NewGauge("NextGC", func() float64 {
			runtime.ReadMemStats(&ms)
			return float64(ms.NextGC)
		}),
		library.NewGauge("NumForcedGC", func() float64 {
			runtime.ReadMemStats(&ms)
			return float64(ms.NumForcedGC)
		}),
		library.NewGauge("NumGC", func() float64 {
			runtime.ReadMemStats(&ms)
			return float64(ms.NumGC)
		}),
		library.NewGauge("OtherSys", func() float64 {
			runtime.ReadMemStats(&ms)
			return float64(ms.OtherSys)
		}),
		library.NewGauge("PauseTotalNs", func() float64 {
			runtime.ReadMemStats(&ms)
			return float64(ms.PauseTotalNs)
		}),
		library.NewGauge("StackInuse", func() float64 {
			runtime.ReadMemStats(&ms)
			return float64(ms.StackInuse)
		}),
		library.NewGauge("StackSys", func() float64 {
			runtime.ReadMemStats(&ms)
			return float64(ms.StackSys)
		}),
		library.NewGauge("Sys", func() float64 {
			runtime.ReadMemStats(&ms)
			return float64(ms.Sys)
		}),
		library.NewGauge("TotalAlloc", func() float64 {
			runtime.ReadMemStats(&ms)
			return float64(ms.TotalAlloc)
		}),
		library.NewGauge("RandomValue", func() float64 {
			return rand.Float64()
		}),
		library.NewCounter("PollCount", func() int64 {
			return 1
		}),
	)

	go func() {
		for {
			time.Sleep(pollPeriod)
			storage.UpdateMetrics()
		}
	}()

	for {
		time.Sleep(reportInterval)
		if err := sender.Send(); err != nil {
			log.Fatal(err)
		}
	}
}
