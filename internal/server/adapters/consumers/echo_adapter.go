package consumers

import (
	"fmt"

	"github.com/gdyunin/metricol.git/internal/server/consume/consumers/echohttp/model"
	"github.com/gdyunin/metricol.git/internal/server/entities"
)

// EchoAdapter is responsible for converting and handling metrics between
// the server model and the internal entities format, pushing and pulling them from the repository.
type EchoAdapter struct {
	mi *entities.MetricsInterface
}

// NewEchoAdapter creates and returns a new EchoAdapter instance using the provided repository.
func NewEchoAdapter(repository entities.MetricsRepository) *EchoAdapter {
	ea := &EchoAdapter{mi: entities.NewMetricsInterface(repository)}
	return ea
}

// PushMetric converts the model metric to an entities metric and pushes it to the repository.
// If an error occurs during the push, it returns a formatted error with the metric details.
func (ea *EchoAdapter) PushMetric(metric *model.Metric) (*model.Metric, error) {
	entityMetric := mtoem(metric)

	newEntityMetric, err := ea.mi.PushMetric(entityMetric)
	if err != nil {
		return nil, fmt.Errorf("failed to push metric %+v to repositories: %w", metric, err)
	}

	return emtom(newEntityMetric), nil
}

// PullMetric converts the model metric to an entities metric and pulls it from the repository.
// If an error occurs during the pull, it returns a formatted error with the metric details.
func (ea *EchoAdapter) PullMetric(metric *model.Metric) (*model.Metric, error) {
	entityMetric := mtoem(metric)

	newEntityMetric, err := ea.mi.PullMetric(entityMetric)
	if err != nil {
		return nil, fmt.Errorf("failed to pull metric %+v from repositories: %w", metric, err)
	}

	return emtom(newEntityMetric), nil
}

// PullAllMetrics retrieves all metrics from the repository and returns them as a slice of model metrics.
// If an error occurs while pulling the metrics, it returns a formatted error.
func (ea *EchoAdapter) PullAllMetrics() ([]*model.Metric, error) {
	ems, err := ea.mi.AllMetricsInRepo()
	if err != nil {
		return nil, fmt.Errorf("failed to pull all metrics from repositories: %w", err)
	}

	m := make([]*model.Metric, 0, len(ems))
	for _, em := range ems {
		m = append(m, emtom(em))
	}

	return m, nil
}

// mtoem converts a model metric to an entities metric.
func mtoem(m *model.Metric) *entities.Metric {
	return &entities.Metric{
		Name:  m.ID,
		Type:  m.MType,
		Value: parseMValue(m),
	}
}

// parseMValue converts the metric value based on its type.
func parseMValue(m *model.Metric) (value any) {
	switch m.MType {
	case entities.MetricTypeCounter:
		if m.Delta != nil {
			value = any(*m.Delta)
		}
	case entities.MetricTypeGauge:
		if m.Value != nil {
			value = any(*m.Value)
		}
	}
	return
}

// emtom converts an entities metric to a model metric.
func emtom(em *entities.Metric) *model.Metric {
	m := &model.Metric{
		ID:    em.Name,
		MType: em.Type,
	}
	fillValueFields(m, em)
	return m
}

// fillValueFields populates the value fields of a model metric based on the corresponding entities metric.
func fillValueFields(to *model.Metric, from *entities.Metric) {
	switch from.Type {
	case entities.MetricTypeCounter:
		if v, ok := from.Value.(int64); ok {
			to.Delta = &v
		}
	case entities.MetricTypeGauge:
		if v, ok := from.Value.(float64); ok {
			to.Value = &v
		}
	}
}
