package collect

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
)

func TestCollector_Collect(t *testing.T) {
	logger := zaptest.NewLogger(t).Sugar()
	collector := NewCollector(logger)

	collector.Collect()

	assert.NotNil(t, collector.ms, "MemStats should not be nil after collection")
	assert.True(t, collector.metadata.PollsCount() > 0, "PollsCount should increase after collection")
}

func TestCollector_Export(t *testing.T) {
	logger := zaptest.NewLogger(t).Sugar()
	collector := NewCollector(logger)
	collector.Collect()

	metrics, resetCh := collector.Export()

	assert.NotNil(t, metrics, "Exported metrics should not be nil")
	assert.Greater(t, len(*metrics), 0, "Exported metrics should contain elements")

	// Confirm reset
	go func() {
		resetCh <- true
	}()
	time.Sleep(100 * time.Millisecond) // Allow time for goroutine to process
	assert.True(t, collector.metadata.PollsCount() == 0, "Metadata should be reset after confirmation")
}

func TestCollector_ExportMemoryMetrics(t *testing.T) {
	logger := zaptest.NewLogger(t).Sugar()
	collector := NewCollector(logger)
	collector.Collect()

	memoryMetrics := collector.exportMemoryMetrics()

	assert.NotNil(t, memoryMetrics, "Memory metrics should not be nil")
	assert.Greater(t, len(memoryMetrics), 0, "Memory metrics should contain elements")
	for _, metric := range memoryMetrics {
		assert.NotEmpty(t, metric.Name, "Metric name should not be empty")
	}
}

func TestCollector_ExportMetadataMetrics(t *testing.T) {
	logger := zaptest.NewLogger(t).Sugar()
	collector := NewCollector(logger)
	collector.Collect()

	metadataMetrics := collector.exportMetadataMetrics()

	assert.NotNil(t, metadataMetrics, "Metadata metrics should not be nil")
	assert.Greater(t, len(metadataMetrics), 0, "Metadata metrics should contain elements")
	for _, metric := range metadataMetrics {
		assert.True(t, metric.IsMetadata, "Metric should be marked as metadata")
	}
}
