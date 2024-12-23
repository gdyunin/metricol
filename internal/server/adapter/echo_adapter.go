package adapter

import (
	"fmt"

	"github.com/gdyunin/metricol.git/internal/server/consume/consumers/echoserver/model"
	"github.com/gdyunin/metricol.git/internal/server/entity"
)

type EchoAdapter struct {
	mi *entity.MetricsInterface
}

func NewEchoAdapter(repository entity.MetricRepository) *EchoAdapter {
	ea := &EchoAdapter{mi: entity.NewMetricsInterface(repository)}

	return ea
}

func (ea *EchoAdapter) PushMetric(metric *model.Metric) (*model.Metric, error) {
	entityMetric := mtoem(metric)

	newEntityMetric, err := ea.mi.PushMetric(entityMetric)
	if err != nil {
		return nil, fmt.Errorf("error push metric %+v to repositories: %w", metric, err)
	}

	return emtom(newEntityMetric), nil
}

func (ea *EchoAdapter) PullMetric(metric *model.Metric) (*model.Metric, error) {
	entityMetric := mtoem(metric)

	newEntityMetric, err := ea.mi.PullMetric(entityMetric)
	if err != nil {
		return nil, fmt.Errorf("error pull metric %+v from repositories: %w", metric, err)
	}

	return emtom(newEntityMetric), nil
}

func (ea *EchoAdapter) PullAllMetrics() ([]*model.Metric, error) {
	ems, err := ea.mi.AllMetricsInRepo()
	if err != nil {
		return nil, fmt.Errorf("error pull metrics from repositories: %w", err)
	}

	m := make([]*model.Metric, 0, len(ems))
	for i, em := range ems {
		m[i] = emtom(em)
	}

	return m, nil
}

//func (ea *EchoAdapter) IsExists(metric *model.Metric) (bool, error) {
//	entityMetric := mtoem(metric)
//
//	isExists, err := ea.mi.IsExists(entityMetric)
//	if err != nil {
//		return false, fmt.Errorf("error check existing metric %+v from repositories: %w", metric, err)
//	}
//
//	return isExists, nil
//}

func mtoem(m *model.Metric) *entity.Metric {
	return &entity.Metric{
		Name:  m.ID,
		Type:  m.MType,
		Value: parseMValue(m),
	}

}

func parseMValue(m *model.Metric) (value any) {
	switch m.MType {
	case entity.MetricTypeCounter:
		if m.Delta != nil {
			value = any(*m.Delta)
		}
	case entity.MetricTypeGauge:
		if m.Value != nil {
			value = any(*m.Value)
		}
	}
	return
}

func emtom(em *entity.Metric) *model.Metric {
	m := &model.Metric{
		ID:    em.Name,
		MType: em.Type,
	}
	fillValueFields(m, em)
	return m

}

func fillValueFields(to *model.Metric, from *entity.Metric) {
	switch from.Type {
	case entity.MetricTypeCounter:
		if v, ok := from.Value.(int64); ok {
			to.Delta = &v
		}
	case entity.MetricTypeGauge:
		if v, ok := from.Value.(float64); ok {
			to.Value = &v
		}
	}
}
