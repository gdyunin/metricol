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
//
// Parameters:
//   - repository: The metrics repository to interact with.
//
// Returns:
//   - A pointer to an initialized EchoAdapter instance.
func NewEchoAdapter(repository entities.MetricsRepository) *EchoAdapter {
	return &EchoAdapter{
		mi: entities.NewMetricsInterface(repository),
	}
}

// PushMetric converts a model metric to an entities metric and pushes it to the repository.
//
// Parameters:
//   - metric: A pointer to the model.Metric instance to be pushed.
//
// Returns:
//   - A pointer to the pushed model.Metric instance with updated data.
//   - An error if the metric could not be pushed to the repository.
func (ea *EchoAdapter) PushMetric(metric *model.Metric) (*model.Metric, error) {
	entityMetric := m2em(metric)

	newEntityMetric, err := ea.mi.PushMetric(entityMetric)
	if err != nil {
		return nil, fmt.Errorf("failed to push metric %+v to the repository: %w", metric, err)
	}

	return em2m(newEntityMetric), nil
}

// PullMetric converts a model metric to an entities metric and pulls it from the repository.
//
// Parameters:
//   - metric: A pointer to the model.Metric instance to be pulled.
//
// Returns:
//   - A pointer to the pulled model.Metric instance with retrieved data.
//   - An error if the metric could not be pulled from the repository.
func (ea *EchoAdapter) PullMetric(metric *model.Metric) (*model.Metric, error) {
	entityMetric := m2em(metric)

	newEntityMetric, err := ea.mi.PullMetric(entityMetric)
	if err != nil {
		return nil, fmt.Errorf("failed to pull metric %+v from the repository: %w", metric, err)
	}

	return em2m(newEntityMetric), nil
}

// PullAllMetrics retrieves all metrics from the repository and converts them to model metrics.
//
// Returns:
//   - A slice of pointers to model.Metric instances representing all retrieved metrics.
//   - An error if the metrics could not be retrieved from the repository.
func (ea *EchoAdapter) PullAllMetrics() ([]*model.Metric, error) {
	ems, err := ea.mi.AllMetricsInRepo()
	if err != nil {
		return nil, fmt.Errorf("failed to pull all metrics from the repository: %w", err)
	}

	metrics := make([]*model.Metric, 0, len(ems))
	for _, em := range ems {
		metrics = append(metrics, em2m(em))
	}

	return metrics, nil
}

// m2em converts a model metric to an entities metric.
//
// Parameters:
//   - m: A pointer to the model.Metric instance to convert.
//
// Returns:
//   - A pointer to the converted entities.Metric instance.
func m2em(m *model.Metric) *entities.Metric {
	return &entities.Metric{
		Name:  m.ID,
		Type:  m.MType,
		Value: parseMValue(m),
	}
}

// em2m converts an entities metric to a model metric.
//
// Parameters:
//   - em: A pointer to the entities.Metric instance to convert.
//
// Returns:
//   - A pointer to the converted model.Metric instance.
func em2m(em *entities.Metric) *model.Metric {
	metric := &model.Metric{
		ID:    em.Name,
		MType: em.Type,
	}
	fillValueFields(metric, em)
	return metric
}

// parseMValue converts the metric value from the model metric based on its type.
//
// Parameters:
//   - m: A pointer to the model.Metric instance containing the value to parse.
//
// Returns:
//   - The parsed value as an empty interface (any).
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

// fillValueFields populates the value fields of a model metric based on the corresponding entities metric.
//
// Parameters:
//   - to: A pointer to the model.Metric instance to populate.
//   - from: A pointer to the entities.Metric instance providing the value data.
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
