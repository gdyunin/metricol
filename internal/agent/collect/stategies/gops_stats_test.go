package stategies

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"go.uber.org/zap"

	"github.com/gdyunin/metricol.git/internal/agent/internal/entity"
)

// findGOPSMetric is a helper that returns a pointer to a metric with the given name.
func findGOPSMetric(metrics *entity.Metrics, name string) *entity.Metric {
	for _, m := range *metrics {
		if m.Name == name {
			return m
		}
	}
	return nil
}

func TestGopsStatsCollectStrategy_Collect_TableDriven(t *testing.T) {
	logger := zap.NewNop().Sugar()
	strat := GopsMemStatsCollectStrategy(logger)

	// Call Collect to obtain the metrics.
	metricsPtr, err := strat.Collect()
	if err != nil {
		t.Fatalf("Collect returned error: %v", err)
	}
	if metricsPtr == nil {
		t.Fatal("Collect returned nil metrics")
	}

	tests := []struct {
		name  string
		check func(metrics *entity.Metrics) error
	}{
		{
			name: "Memory metrics present",
			check: func(metrics *entity.Metrics) error {
				// Verify that memory metrics "TotalMemory" and "FreeMemory" exist and are of gauge type.
				keys := []string{"TotalMemory", "FreeMemory"}
				for _, key := range keys {
					m := findGOPSMetric(metrics, key)
					if m == nil {
						return fmt.Errorf("expected metric %q not found", key)
					}
					if m.Type != entity.MetricTypeGauge {
						return fmt.Errorf("expected metric %q to have type %q, got %q", key, entity.MetricTypeGauge, m.Type)
					}
				}
				return nil
			},
		},
		{
			name: "CPU metrics present",
			check: func(metrics *entity.Metrics) error {
				// Verify that at least one CPU utilization metric is present and of gauge type.
				var found bool
				for _, m := range *metrics {
					if strings.HasPrefix(m.Name, "CPUutilization") {
						found = true
						if m.Type != entity.MetricTypeGauge {
							return fmt.Errorf("expected CPU metric %q to have type %q, got %q", m.Name, entity.MetricTypeGauge, m.Type)
						}
					}
				}
				if !found {
					return fmt.Errorf("no CPU utilization metric found")
				}
				return nil
			},
		},
		{
			name: "Total metrics count at least 3",
			check: func(metrics *entity.Metrics) error {
				// The collected metrics should include memory metrics plus at least one CPU metric.
				if len(*metrics) < 3 {
					return fmt.Errorf("expected at least 3 metrics, got %d", len(*metrics))
				}
				return nil
			},
		},
		{
			name: "Collect can be called repeatedly",
			check: func(metrics *entity.Metrics) error {
				// Calling Collect a second time should also return valid metrics.
				// (Allow a brief delay if needed, as CPU metrics collection uses a 1-second interval.)
				time.Sleep(1100 * time.Millisecond)
				m2, err := strat.Collect()
				if err != nil {
					return fmt.Errorf("second call to Collect returned error: %v", err)
				}
				if m2 == nil || len(*m2) < 3 {
					return fmt.Errorf("second call to Collect returned insufficient metrics")
				}
				return nil
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if err := tc.check(metricsPtr); err != nil {
				t.Error(err)
			}
		})
	}
}
