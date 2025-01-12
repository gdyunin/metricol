package collect

import (
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func TestNewCollector(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "Create new Collector"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := zaptest.NewLogger(t).Sugar()
			c := NewCollector(logger)

			require.NotNil(t, c)
			require.NotNil(t, c.metadata)
			require.NotNil(t, c.ms)
			require.NotNil(t, c.mu)
			require.NotNil(t, c.resetMetaConfirmCh)
			assert.Equal(t, logger, c.logger)
		})
	}
}

func TestCollector_Collect(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "Collect metrics and update metadata"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := zaptest.NewLogger(t).Sugar()
			c := NewCollector(logger)

			c.Collect()

			assert.Greater(t, c.metadata.PollsCount(), int64(0))
			assert.NotZero(t, c.metadata.LastPollSeed())
		})
	}
}

func TestCollector_Export(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "Export metrics and return channel"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := zaptest.NewLogger(t).Sugar()
			c := NewCollector(logger)

			metrics, resetCh := c.Export()

			require.NotNil(t, metrics)
			require.NotNil(t, resetCh)

			assert.Greater(t, metrics.Length(), 0)
		})
	}
}

func TestCollector_ExportMemoryMetrics(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "Export memory metrics from runtime.MemStats"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := zaptest.NewLogger(t).Sugar()
			c := NewCollector(logger)

			runtime.ReadMemStats(c.ms)
			memoryMetrics := c.exportMemoryMetrics()

			require.NotEmpty(t, memoryMetrics)
			for _, metric := range memoryMetrics {
				require.NotEmpty(t, metric.Name)
				require.NotNil(t, metric.Value)
			}
		})
	}
}

func TestCollector_ExportMetadataMetrics(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "Export metadata metrics"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := zaptest.NewLogger(t).Sugar()
			c := NewCollector(logger)

			c.metadata.Update()
			metadataMetrics := c.exportMetadataMetrics()

			require.NotEmpty(t, metadataMetrics)
			require.Len(t, metadataMetrics, 2)

			for _, metric := range metadataMetrics {
				assert.NotEmpty(t, metric.Name)
				assert.True(t, metric.IsMetadata)
				assert.NotNil(t, metric.Value)
			}
		})
	}
}

func TestCollector_WaitResetConfirmation(t *testing.T) {
	tests := []struct {
		name           string
		resetConfirmed bool
		expectedPolls  int64
		expectedSeed   float64
	}{
		{name: "Reset confirmed", resetConfirmed: true, expectedPolls: 0, expectedSeed: 0},
		{name: "Reset canceled", resetConfirmed: false, expectedPolls: 1, expectedSeed: -1}, // -1: any non-zero value
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := zaptest.NewLogger(t).Sugar()
			c := NewCollector(logger)

			c.metadata.Update()

			go func() {
				c.resetMetaConfirmCh <- tt.resetConfirmed
			}()

			c.waitResetConfirmation()

			if tt.resetConfirmed {
				assert.Equal(t, tt.expectedPolls, c.metadata.PollsCount())
				assert.Equal(t, tt.expectedSeed, c.metadata.LastPollSeed())
			} else {
				assert.Equal(t, tt.expectedPolls, c.metadata.PollsCount())
				assert.NotZero(t, c.metadata.LastPollSeed())
			}
		})
	}
}
