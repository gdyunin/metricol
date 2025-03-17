package repository

import (
	"context"
	"testing"

	"github.com/gdyunin/metricol.git/internal/server/internal/entity"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestNewInMemoryRepository(t *testing.T) {
	logger := zap.NewNop().Sugar()
	repo := NewInMemoryRepository(logger)
	assert.NotNil(t, repo, "Repository should not be nil")
}

func TestUpdate(t *testing.T) {
	logger := zap.NewNop().Sugar()
	repo := NewInMemoryRepository(logger)
	ctx := context.Background()

	t.Run("Valid metric", func(t *testing.T) {
		metric := &entity.Metric{Name: "test", Type: "gauge", Value: 42.0}
		err := repo.Update(ctx, metric)
		assert.NoError(t, err)
	})

	t.Run("Nil metric", func(t *testing.T) {
		err := repo.Update(ctx, nil)
		assert.Error(t, err)
		assert.Equal(t, "metric should be non-nil, but got nil", err.Error())
	})
}

func TestUpdateBatchInMemory(t *testing.T) {
	logger := zap.NewNop().Sugar()
	repo := NewInMemoryRepository(logger)
	ctx := context.Background()

	t.Run("Valid batch", func(t *testing.T) {
		metrics := entity.Metrics{
			&entity.Metric{Name: "test1", Type: "gauge", Value: 1.0},
			&entity.Metric{Name: "test2", Type: "counter", Value: 2},
		}
		err := repo.UpdateBatch(ctx, &metrics)
		assert.NoError(t, err)
	})

	t.Run("Nil batch", func(t *testing.T) {
		err := repo.UpdateBatch(ctx, nil)
		assert.Error(t, err)
		assert.Equal(t, "metrics should be non-nil, but got nil", err.Error())
	})
}

func TestFind(t *testing.T) {
	logger := zap.NewNop().Sugar()
	repo := NewInMemoryRepository(logger)
	ctx := context.Background()

	metric := &entity.Metric{Name: "test", Type: "gauge", Value: 42.0}
	_ = repo.Update(ctx, metric)

	t.Run("Existing metric", func(t *testing.T) {
		result, err := repo.Find(ctx, "gauge", "test")
		assert.NoError(t, err)
		assert.Equal(t, metric, result)
	})

	t.Run("Non-existing metric", func(t *testing.T) {
		result, err := repo.Find(ctx, "gauge", "missing")
		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestAll(t *testing.T) {
	logger := zap.NewNop().Sugar()
	repo := NewInMemoryRepository(logger)
	ctx := context.Background()

	metrics := entity.Metrics{
		&entity.Metric{Name: "test1", Type: "gauge", Value: 1.0},
		&entity.Metric{Name: "test2", Type: "counter", Value: 2},
	}
	_ = repo.UpdateBatch(ctx, &metrics)

	result, err := repo.All(ctx)
	assert.NoError(t, err)
	assert.Equal(t, &metrics, result)
}

func TestCheckConnection(t *testing.T) {
	logger := zap.NewNop().Sugar()
	repo := NewInMemoryRepository(logger)
	ctx := context.Background()

	err := repo.CheckConnection(ctx)
	assert.NoError(t, err)
}
