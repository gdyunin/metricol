package repository

import (
	"NewNewMetricol/internal/server/internal/entity"
	"errors"
)

var (
	ErrNotFound = errors.New("metric not found")
)

type Repository interface {
	Update(metric *entity.Metric) error
	Find(metricType string, name string) (*entity.Metric, error)
	All() (*entity.Metrics, error)
}
