package collect

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/gdyunin/metricol.git/internal/agent/internal/entity"
	"go.uber.org/zap"
)

// validStrategy returns a valid batch on its first call and an empty batch thereafter.
type validStrategy struct {
	mu     sync.Mutex
	called bool
}

func (v *validStrategy) Collect() (*entity.Metrics, error) {
	v.mu.Lock()
	defer v.mu.Unlock()
	if !v.called {
		v.called = true
		// Return a non-empty metrics batch.
		m := entity.Metrics{
			{
				Name:  "test",
				Type:  entity.MetricTypeGauge,
				Value: 3.14,
			},
		}
		return &m, nil
	}
	// Subsequent calls return an empty batch.
	empty := entity.Metrics{}
	return &empty, nil
}

// errorStrategy always returns an error.
type errorStrategy struct{}

func (e *errorStrategy) Collect() (*entity.Metrics, error) {
	return nil, errors.New("collect error")
}

// emptyStrategy always returns an empty (but non-nil) metrics batch.
type emptyStrategy struct{}

func (e *emptyStrategy) Collect() (*entity.Metrics, error) {
	empty := entity.Metrics{}
	return &empty, nil
}

func TestStreamCollector_StartStreaming(t *testing.T) {
	// Table-driven test cases.
	tests := []struct {
		name          string
		strategies    []Strategy
		interval      time.Duration // interval for ticker
		waitDuration  time.Duration // how long to let the collector run before cancellation
		expectedCount int           // expected number of valid batches sent on streamTo
	}{
		{
			name:          "valid strategy",
			strategies:    []Strategy{&validStrategy{}},
			interval:      100 * time.Millisecond,
			waitDuration:  150 * time.Millisecond,
			expectedCount: 1,
		},
		{
			name:          "error strategy",
			strategies:    []Strategy{&errorStrategy{}},
			interval:      100 * time.Millisecond,
			waitDuration:  150 * time.Millisecond,
			expectedCount: 0,
		},
		{
			name:          "empty strategy",
			strategies:    []Strategy{&emptyStrategy{}},
			interval:      100 * time.Millisecond,
			waitDuration:  150 * time.Millisecond,
			expectedCount: 0,
		},
		{
			name:         "mixed strategies",
			strategies:   []Strategy{&validStrategy{}, &errorStrategy{}},
			interval:     100 * time.Millisecond,
			waitDuration: 150 * time.Millisecond,
			// Only the valid strategy produces a valid batch.
			expectedCount: 1,
		},
		{
			name:         "multiple valid strategies",
			strategies:   []Strategy{&validStrategy{}, &validStrategy{}},
			interval:     100 * time.Millisecond,
			waitDuration: 150 * time.Millisecond,
			// Both strategies will produce one valid batch each.
			expectedCount: 2,
		},
	}

	// Use a no-op logger.
	logger := zap.NewNop().Sugar()

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Create a channel to receive metrics batches.
			streamTo := make(chan *entity.Metrics, 10)
			// Create the collector.
			collector := NewStreamCollector(streamTo, tc.interval, tc.strategies, logger)
			// Create a cancellable context.
			ctx, cancel := context.WithCancel(context.Background())
			// Start the collector in a separate goroutine.
			go collector.StartStreaming(ctx)

			// Allow some time for one ticker tick to occur.
			time.Sleep(tc.waitDuration)
			// Cancel the context to stop the collector.
			cancel()

			// Drain the streamTo channel. The collector is expected to close the channel.
			var count int
			for batch := range streamTo {
				// A valid batch is one whose Length is non-zero.
				if batch != nil && batch.Length() > 0 {
					count++
				}
			}

			if count != tc.expectedCount {
				t.Errorf("expected %d valid batch(es) but got %d", tc.expectedCount, count)
			}
		})
	}
}
