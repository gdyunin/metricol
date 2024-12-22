package adapter

import (
	"fmt"

	"github.com/gdyunin/metricol.git/internal/server/consume/consumers/ginserver/model"
	"github.com/gdyunin/metricol.git/internal/server/entity"
)

type GinController struct {
	metricInterface *entity.MetricsInterface
}

func NewGinController(repo entity.MetricRepository) *GinController {
	return &GinController{
		metricInterface: entity.NewMetricsInterface(repo),
	}
}

func (gc *GinController) PushMetric(metric *model.Metric) (*model.Metric, error) {
	em := m2em(metric)

	em, err := gc.metricInterface.PushMetric(em)
	if err != nil {
		return nil, fmt.Errorf("error push metric %+v to repositories: %w", metric, err)
	}

	return em2m(em), nil
}

func (gc *GinController) PullMetric(metric *model.Metric) (*model.Metric, error) {
	em := m2em(metric)

	em, err := gc.metricInterface.PullMetric(em)
	if err != nil {
		return nil, fmt.Errorf("error pull metric %+v from repositories: %w", metric, err)
	}

	return em2m(em), nil
}

func (gc *GinController) PullAllMetrics() ([]*model.Metric, error) {
	ems, err := gc.metricInterface.AllMetricsInRepo()
	if err != nil {
		return nil, fmt.Errorf("error pull metrics from repositories: %w", err)
	}

	m := make([]*model.Metric, 0, len(ems))
	for _, em := range ems {
		m = append(m, em2m(em))
	}
	return m, nil
}

func m2em(m *model.Metric) (em *entity.Metric) {
	em = &entity.Metric{
		Name: m.ID,
		Type: m.MType,
	}

	switch m.MType {
	case entity.MetricTypeCounter:
		if m.Delta != nil {
			em.Value = any(*m.Delta)
		} else {
			em.Value = nil
		}
	case entity.MetricTypeGauge:
		if m.Value != nil {
			em.Value = any(*m.Value)
		} else {
			em.Value = nil
		}
	}

	return
}

func em2m(em *entity.Metric) (m *model.Metric) {
	m = &model.Metric{
		ID:    em.Name,
		MType: em.Type,
	}

	switch em.Type {
	case entity.MetricTypeCounter:
		v, _ := em.Value.(int64)
		m.Delta = &v
	case entity.MetricTypeGauge:
		v, _ := em.Value.(float64)
		m.Value = &v
	}

	return
}
