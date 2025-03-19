package stategies

import (
	"sync"
	"testing"

	"github.com/gdyunin/metricol.git/internal/agent/internal/entity"
	"go.uber.org/zap"
)

// findMemStatsMetric is a helper that returns the pointer to a metric by name.
func findMemStatsMetric(metrics *entity.Metrics, name string) *entity.Metric {
	for _, m := range *metrics {
		if m.Name == name {
			return m
		}
	}
	return nil
}

// TestMemStatsCollectStrategy_Collect verifies that a single call to Collect
// returns the expected set of metrics with valid metadata values.
func TestMemStatsCollectStrategy_Collect(t *testing.T) {
	logger := zap.NewNop().Sugar()
	strategy := NewMemStatsCollectStrategy(logger)

	metricsPtr, err := strategy.Collect()
	if err != nil {
		t.Fatalf("Collect returned error: %v", err)
	}
	if metricsPtr == nil {
		t.Fatalf("Collect returned nil metrics")
	}

	metrics := *metricsPtr

	// Expecting 27 memory metrics + 2 metadata metrics = 29 metrics.
	const expectedCount = 29
	if len(metrics) != expectedCount {
		t.Errorf("expected %d metrics but got %d", expectedCount, len(metrics))
	}

	// Check for a memory metric, e.g. "Alloc".
	allocMetric := findMemStatsMetric(metricsPtr, "Alloc")
	if allocMetric == nil {
		t.Errorf("memory metric 'Alloc' not found")
	} else {
		// Expecting a gauge metric with a numeric value.
		if allocMetric.Type != entity.MetricTypeGauge {
			t.Errorf("expected 'Alloc' metric type %s but got %s", entity.MetricTypeGauge, allocMetric.Type)
		}
		// The value should be a float64.
		if _, ok := allocMetric.Value.(float64); !ok {
			t.Errorf("expected 'Alloc' metric value to be float64, got %T", allocMetric.Value)
		}
	}

	// Check metadata metrics.
	pollCountMetric := findMemStatsMetric(metricsPtr, "PollCount")
	if pollCountMetric == nil {
		t.Errorf("metadata metric 'PollCount' not found")
	} else {
		// Since NewMetadata initializes pollsCount to 0 and Update increments it,
		// the first call to Collect should yield PollCount == 1.
		if v, ok := pollCountMetric.Value.(int64); !ok {
			t.Errorf("expected 'PollCount' value to be int64, got %T", pollCountMetric.Value)
		} else if v != 1 {
			t.Errorf("expected 'PollCount' to be 1, got %d", v)
		}
	}

	randomValueMetric := findMemStatsMetric(metricsPtr, "RandomValue")
	if randomValueMetric == nil {
		t.Errorf("metadata metric 'RandomValue' not found")
	} else {
		if v, ok := randomValueMetric.Value.(float64); !ok {
			t.Errorf("expected 'RandomValue' value to be float64, got %T", randomValueMetric.Value)
		} else if v == 0 {
			t.Errorf("expected 'RandomValue' to be non-zero, got %v", v)
		}
	}
}

// TestMemStatsCollectStrategy_Concurrent calls Collect concurrently from several
// goroutines to verify that the strategy is safe for concurrent use.
func TestMemStatsCollectStrategy_Concurrent(t *testing.T) {
	logger := zap.NewNop().Sugar()
	strategy := NewMemStatsCollectStrategy(logger)

	const goroutines = 10
	var wg sync.WaitGroup
	results := make(chan *entity.Metrics, goroutines)

	for range goroutines {
		wg.Add(1)
		go func() {
			defer wg.Done()
			metricsPtr, err := strategy.Collect()
			if err != nil {
				t.Errorf("Collect returned error: %v", err)
				return
			}
			results <- metricsPtr
		}()
	}

	wg.Wait()
	close(results)

	for metricsPtr := range results {
		if metricsPtr == nil {
			t.Error("Collect returned nil metrics in concurrent call")
			continue
		}
		metrics := *metricsPtr
		if len(metrics) != 29 {
			t.Errorf("expected %d metrics but got %d in concurrent call", 29, len(metrics))
		}
		// Check one metadata metric.
		pollCountMetric := findMemStatsMetric(metricsPtr, "PollCount")
		if pollCountMetric == nil {
			t.Error("metadata metric 'PollCount' not found in concurrent call")
		} else {
			if v, ok := pollCountMetric.Value.(int64); !ok || v != 1 {
				t.Errorf(
					"expected 'PollCount' to be 1 in concurrent call, got %v (type %T)",
					pollCountMetric.Value,
					pollCountMetric.Value,
				)
			}
		}
	}
}
