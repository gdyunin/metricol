package controller

import (
	"context"
	"testing"

	"github.com/gdyunin/metricol.git/internal/server/internal/entity"
	"github.com/gdyunin/metricol.git/internal/server/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) Update(ctx context.Context, metric *entity.Metric) error {
	args := m.Called(ctx, metric)
	return args.Error(0) //nolint:wrapcheck // for tests
}

func (m *MockRepository) UpdateBatch(ctx context.Context, metrics *entity.Metrics) error {
	args := m.Called(ctx, metrics)
	return args.Error(0) //nolint:wrapcheck // for tests
}

func (m *MockRepository) Find(ctx context.Context, metricType, name string) (*entity.Metric, error) {
	args := m.Called(ctx, metricType, name)
	metric, ok := args.Get(0).(*entity.Metric)
	if !ok && args.Get(0) != nil {
		panic("unexpected type returned from mock")
	}
	return metric, args.Error(1) //nolint:wrapcheck // for tests
}

func (m *MockRepository) All(ctx context.Context) (*entity.Metrics, error) {
	args := m.Called(ctx)
	metrics, ok := args.Get(0).(*entity.Metrics)
	if !ok && args.Get(0) != nil {
		panic("unexpected type returned from mock")
	}
	return metrics, args.Error(1) //nolint:wrapcheck // for tests
}

func (m *MockRepository) CheckConnection(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0) //nolint:wrapcheck // for tests
}

func TestPushMetric(t *testing.T) {
	repo := new(MockRepository)
	service := NewMetricService(repo)
	ctx := context.Background()

	tests := []struct {
		metric    *entity.Metric
		name      string
		expectErr bool
	}{
		{name: "Push valid metric", metric: &entity.Metric{Name: "test", Type: "counter", Value: 10}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo.On("Find", mock.Anything, tt.metric.Type, tt.metric.Name).
				Return(nil, repository.ErrNotFoundInRepo)
			repo.On("UpdateBatch", mock.Anything, mock.Anything).Return(nil)
			_, err := service.PushMetric(ctx, tt.metric)
			assert.Equal(t, tt.expectErr, err != nil)
			repo.AssertExpectations(t)
		})
	}
}

func TestPull(t *testing.T) {
	repo := new(MockRepository)
	service := NewMetricService(repo)
	ctx := context.Background()

	tests := []struct {
		name       string
		metricType string
		metricName string
		expectErr  bool
	}{
		{name: "Pull existing metric", metricType: "counter", metricName: "test"},
		{name: "Pull non-existing metric", metricType: "counter", metricName: "unknown", expectErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.expectErr {
				repo.On("Find", mock.Anything, tt.metricType, tt.metricName).
					Return(nil, repository.ErrNotFoundInRepo)
			} else {
				repo.On(
					"Find",
					mock.Anything,
					tt.metricType,
					tt.metricName,
				).Return(&entity.Metric{Name: tt.metricName, Type: tt.metricType, Value: 10}, nil)
			}
			_, err := service.Pull(ctx, tt.metricType, tt.metricName)
			assert.Equal(t, tt.expectErr, err != nil)
		})
	}
}

func TestPullAll(t *testing.T) {
	repo := new(MockRepository)
	service := NewMetricService(repo)
	ctx := context.Background()

	tests := []struct {
		name      string
		expectErr bool
	}{
		{name: "Pull all metrics"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo.On("All", mock.Anything).
				Return(&entity.Metrics{{Name: "test", Type: "counter", Value: 10}}, nil)
			_, err := service.PullAll(ctx)
			assert.Equal(t, tt.expectErr, err != nil)
		})
	}
}

func TestCheckConnection(t *testing.T) {
	repo := new(MockRepository)
	service := NewMetricService(repo)
	ctx := context.Background()

	tests := []struct {
		name      string
		expectErr bool
	}{
		{name: "Check connection"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo.On("CheckConnection", mock.Anything).Return(nil)
			err := service.CheckConnection(ctx)
			assert.Equal(t, tt.expectErr, err != nil)
		})
	}
}

func TestValidate(t *testing.T) {
	service := NewMetricService(nil)
	tests := []struct {
		metric    *entity.Metric
		name      string
		expectErr bool
	}{
		{name: "Valid metric", metric: &entity.Metric{Name: "test", Type: "counter", Value: 10}},
		{name: "Nil metric", expectErr: true},
		{name: "Missing name", metric: &entity.Metric{Name: "", Type: "counter", Value: 10}, expectErr: true},
		{name: "Missing type", metric: &entity.Metric{Name: "test", Type: "", Value: 10}, expectErr: true},
		{
			name:      "Missing value",
			metric:    &entity.Metric{Name: "test", Type: "counter", Value: nil},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.validate(tt.metric)
			assert.Equal(t, tt.expectErr, err != nil)
		})
	}
}
