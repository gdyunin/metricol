package produce

import (
	"github.com/gdyunin/metricol.git/internal/agent/entity"
	"github.com/gdyunin/metricol.git/internal/agent/produce/produsers/rstclient/model"

	"go.uber.org/zap"
)

var logger *zap.SugaredLogger

// RestyClientAdapter serves as an adapter for interacting with the metrics repository.
// It converts metrics from the internal entity format to the Resty Client's model format.
type RestyClientAdapter struct {
	metricInterface *entity.MetricsInterface
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
func NewRestyClientAdapter(repo entity.MetricsRepository, log *zap.SugaredLogger) *RestyClientAdapter {
	logger = log
	return &RestyClientAdapter{metricInterface: entity.NewMetricsInterface(repo)}
}

// Metrics retrieves all stored metrics from the repository and converts them into
// the format required by the Resty Client.
//
// Returns:
//   - A slice of pointers to Metric instances formatted for the Resty Client.
func (r *RestyClientAdapter) Metrics() []*model.Metric {
	// Retrieve all entity metrics from the repository.
	emAll := r.metricInterface.Metrics()

	// Convert each entity metric into a Resty Client metric.
	mAll := make([]*model.Metric, 0, len(emAll))
	for _, em := range emAll {
		mAll = append(mAll, em2m(em))
	}

	return mAll
}

// em2m converts an entity metric into a Resty Client metric model.
// It handles different metric types (e.g., counter and gauge) and ensures proper value mapping.
//
// Parameters:
//   - em: A pointer to an entity Metric instance.
//
// Returns:
//   - A pointer to a Resty Client Metric instance.
func em2m(em *entity.Metric) *model.Metric {
	m := &model.Metric{
		ID:    em.Name,
		MType: em.Type,
	}

	// Map values based on metric type.
	switch em.Type {
	case entity.MetricTypeCounter:
		if v, ok := em.Value.(int64); ok {
			m.Delta = &v
		} else {
			logger.Errorf("metric skipped: failed to convert counter metric '%s': expected int64 but got %T", em.Name, em.Value)
		}
	case entity.MetricTypeGauge:
		if v, ok := em.Value.(float64); ok {
			m.Value = &v
		} else {
			logger.Errorf("metric skipped: failed to convert gauge metric '%s': expected float64 but got %T", em.Name, em.Value)
		}
	default:
		logger.Warnf("metric skipped: unknown metric type for metric '%s': %s", em.Name, em.Type)
	}

	return m
}
