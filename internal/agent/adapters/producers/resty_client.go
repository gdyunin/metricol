package producers

import (
	"github.com/gdyunin/metricol.git/internal/agent/entities"
	"github.com/gdyunin/metricol.git/internal/agent/produce/produsers/rstclient/model"

	"go.uber.org/zap"
)

// RestyClientAdapter serves as an adapters for interacting with the metrics repository.
// It converts metrics from the internal entities format to the Resty Client's model format.
type RestyClientAdapter struct {
	metricInterface *entities.MetricsInterface // Interface for interacting with the metrics repository.
	logger          *zap.SugaredLogger         // Logger for capturing runtime information and errors.
}

// NewRestyClientAdapter initializes a new RestyClientAdapter with the provided metrics repository and logger.
// It sets up the internal metrics interface for accessing stored metrics.
//
// Parameters:
//   - repo: An instance of MetricsRepository to manage stored metrics.
//   - log: An instance of zap.SugaredLogger for logging purposes.
//
// Returns:
//   - A pointer to an initialized RestyClientAdapter.
func NewRestyClientAdapter(repo entities.MetricsRepository, log *zap.SugaredLogger) *RestyClientAdapter {
	return &RestyClientAdapter{metricInterface: entities.NewMetricsInterface(repo), logger: log}
}

// Metrics retrieves all stored metrics from the repository and converts them into
// the format required by the Resty Client.
//
// Returns:
//   - A slice of pointers to Metric instances formatted for the Resty Client.
func (r *RestyClientAdapter) Metrics() []*model.Metric {
	// Retrieve all entities metrics from the repository.
	emAll := r.metricInterface.Metrics()

	// Prepare a slice to hold the converted metrics.
	mAll := make([]*model.Metric, 0, len(emAll))

	// Iterate over the entities metrics and convert each one.
	for _, em := range emAll {
		mAll = append(mAll, r.em2m(em))
	}

	return mAll
}

// em2m converts an entities metric into a Resty Client metric model.
// It handles different metric types (e.g., counter and gauge) and ensures proper value mapping.
//
// Parameters:
//   - em: A pointer to an entities Metric instance.
//
// Returns:
//   - A pointer to a Resty Client Metric instance.
func (r *RestyClientAdapter) em2m(em *entities.Metric) *model.Metric {
	// Initialize a new Resty Client metric with basic fields.
	m := &model.Metric{
		ID:    em.Name, // Set the metric ID to the entities metric's name.
		MType: em.Type, // Set the metric type (e.g., counter, gauge).
	}

	// Map metric values based on the metric type.
	switch em.Type {
	case entities.MetricTypeCounter:
		// Attempt to convert and set the value for counter metrics.
		if v, ok := em.Value.(int64); ok {
			m.Delta = &v // Set Delta field for counter metrics.
		} else {
			// Log an error if the value type is incorrect.
			r.logger.Errorf(
				"metric skipped: failed to convert counter metric '%s':"+
					" expected int64 but got %T", em.Name, em.Value,
			)
		}
	case entities.MetricTypeGauge:
		// Attempt to convert and set the value for gauge metrics.
		if v, ok := em.Value.(float64); ok {
			m.Value = &v // Set Value field for gauge metrics.
		} else {
			// Log an error if the value type is incorrect.
			r.logger.Errorf("metric skipped: failed to convert gauge metric '%s':"+
				" expected float64 but got %T", em.Name, em.Value,
			)
		}
	default:
		// Log a warning for unknown metric types.
		r.logger.Warnf("metric skipped: unknown metric type for metric '%s': %s", em.Name, em.Type)
	}

	return m // Return the converted metric.
}
