package usecase

import (
	"NewNewMetricol/internal/server/internal/entity"
	"NewNewMetricol/internal/server/repository"
	"NewNewMetricol/pkg/convert"
	"errors"
	"fmt"
)

type MetricService struct {
	repo repository.Repository
}

func NewMetricService(repo repository.Repository) *MetricService {
	return &MetricService{repo: repo}
}

func (s *MetricService) PushMetric(metric *entity.Metric) (*entity.Metric, error) {
	if err := s.isValidMetric(metric); err != nil {
		return nil, fmt.Errorf("metric invalid: %w", err)
	}

	switch metric.Type {
	case entity.MetricTypeCounter:
		return s.pushCounter(metric)
	default:
		return s.pushDefault(metric)
	}
}

func (s *MetricService) Pull(metricType string, name string) (*entity.Metric, error) {
	exist, err := s.repo.IsExist(metricType, name)
	if err != nil {
		return nil, fmt.Errorf("failed check is exist in repo %w", err)
	}
	if !exist {
		return nil, fmt.Errorf("metric not exist in repo %w", err)
	}

	m, err := s.repo.Find(metricType, name)
	if err != nil {
		return nil, fmt.Errorf("error while pull from repo: %w", err)
	}

	return m, nil
}

func (s *MetricService) PullAll() (*entity.Metrics, error) {
	return s.repo.All()
}

func (s *MetricService) isValidMetric(metric *entity.Metric) error {
	if metric == nil {
		return errors.New("metric dont be nil")
	}

	if metric.Name == "" {
		return errors.New("metric name empty")
	}

	if metric.Type == "" {
		return errors.New("metric type empty")
	}

	if metric.Value == nil {
		return errors.New("metric value empty")
	}

	return nil
}

func (s *MetricService) pushCounter(metric *entity.Metric) (*entity.Metric, error) {
	var currentVal int64
	var newVal int64

	exist, err := s.repo.IsExist(metric.Type, metric.Name)
	if err != nil {
		return nil, fmt.Errorf("failed check is exist in repo %w", err)
	}

	if !exist {
		return s.pushDefault(&entity.Metric{
			Value: currentVal + newVal,
			Name:  metric.Name,
			Type:  entity.MetricTypeCounter,
		})
	}

	current, err := s.repo.Find(metric.Type, metric.Name)
	if err != nil {
		return nil, fmt.Errorf("failed push counter: %w", err)
	}

	if current != nil {
		currentVal, err = convert.AnyToInt64(current.Value)
		if err != nil {
			return nil, fmt.Errorf("failed while get current state of counter: %w", err)
		}
	}

	newVal, err = convert.AnyToInt64(metric.Value)
	if err != nil {
		return nil, fmt.Errorf("failed while get value counter: %w", err)
	}

	return s.pushDefault(&entity.Metric{
		Value: currentVal + newVal,
		Name:  metric.Name,
		Type:  entity.MetricTypeCounter,
	})
}

func (s *MetricService) pushDefault(metric *entity.Metric) (*entity.Metric, error) {
	if err := s.repo.Update(metric); err != nil {
		return nil, fmt.Errorf("failed push metric to repository: %w", err)
	}
	return metric, nil
}
