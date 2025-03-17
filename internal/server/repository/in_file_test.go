package repository

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestNewInFileRepository(t *testing.T) {
	logger := zap.NewNop().Sugar()
	tests := []struct {
		name     string
		path     string
		filename string
		interval time.Duration
		restore  bool
	}{
		{name: "Default settings", path: "/tmp", filename: "metrics.json"},
		{name: "Auto flush enabled", path: "/tmp", filename: "metrics.json", interval: 1 * time.Second},
		{name: "Restore on build", path: "/tmp", filename: "metrics.json", restore: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewInFileRepository(logger, tt.path, tt.filename, tt.interval, tt.restore)
			assert.NotNil(t, repo, "Repository should not be nil")
		})
	}
}

func TestUpdateBatchInFile(t *testing.T) {
	logger := zap.NewNop().Sugar()
	repo := NewInFileRepository(logger, "/tmp", "test.json", 0, false)

	t.Run("Nil metrics should return error", func(t *testing.T) {
		err := repo.UpdateBatch(context.Background(), nil)
		assert.Error(t, err)
		assert.Equal(t, "metrics should be non-nil, but got nil", err.Error())
	})
}

func TestMustMakeFile(t *testing.T) {
	logger := zap.NewNop().Sugar()
	tempFile := "/tmp/test_metrics.json"
	defer func() { _ = os.Remove(tempFile) }()

	repo := NewInFileRepository(logger, "/tmp", "test_metrics.json", 0, false)
	assert.NotNil(t, repo)

	repo.mustMakeFile()
	_, err := os.Stat(tempFile)
	assert.NoError(t, err, "File should be created")
}

func TestRestore(t *testing.T) {
	logger := zap.NewNop().Sugar()
	tempFile := "/tmp/test_restore.json"
	defer func() { _ = os.Remove(tempFile) }()

	repo := NewInFileRepository(logger, "/tmp", "test_restore.json", 0, true)
	assert.NotNil(t, repo)

	err := repo.restore()
	assert.NoError(t, err, "Restore should not return an error even if file is empty")
}

func TestShutdown(t *testing.T) {
	logger := zap.NewNop().Sugar()
	repo := NewInFileRepository(logger, "/tmp", "test.json", 1*time.Second, false)
	assert.NotNil(t, repo)

	go func() {
		time.Sleep(2 * time.Second)
		repo.Shutdown()
	}()

	select {
	case <-repo.stopCh:
		assert.True(t, true, "Shutdown should send a stop signal")
	case <-time.After(3 * time.Second):
		assert.Fail(t, "Shutdown did not send stop signal in time")
	}
}
