package agent

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path"

	"github.com/gdyunin/metricol.git/internal/common/models"
	"github.com/gdyunin/metricol.git/internal/metrics"
	"github.com/go-resty/resty/v2"
)

// Observer interface represent followers of sender.
type Observer interface {
	OnNotify()
}

// MetricsSender implements the Sender interface, responsible for sending metrics.
type MetricsSender struct {
	observers      map[Observer]struct{} // List of unique Observer.
	metricsFetcher Fetcher               // Fetcher to retrieve metrics.
	client         *resty.Client         // HTTP-client for sending requests.
	serverAddress  string                // Address of the server to send metrics to.
}

// NewMetricsSender creates a new instance of MetricsSender with the provided fetcher and server address.
func NewMetricsSender(fetcher Fetcher, address string) *MetricsSender {
	return &MetricsSender{
		metricsFetcher: fetcher,
		serverAddress:  address,
		client:         resty.New(),
		observers:      make(map[Observer]struct{}),
	}
}

// Send sends all metrics to the configured server address.
// It returns an error if any metric fails to be sent.
func (m *MetricsSender) Send() error {
	var errString string
	for _, mm := range m.metricsFetcher.Metrics() {
		//var metricName, metricType string

		var metric models.Metrics

		// TODO: place it in a separate function, e.g. recognizeMetric()
		switch currentMetricStruct := mm.(type) {
		case *metrics.Counter:
			metric.ID = currentMetricStruct.Name
			metric.MType = metrics.MetricTypeCounter
			metric.Delta = &currentMetricStruct.Value
		case *metrics.Gauge:
			metric.ID = currentMetricStruct.Name
			metric.MType = metrics.MetricTypeGauge
			metric.Value = &currentMetricStruct.Value
		default:
			errString += fmt.Sprintf("error sending metric %v: failed conversion Metric to Struct\n", mm)
			continue // Skip to the next metric.
		}

		// TODO: place it in a separate function, e.g. sendUpdateRequest()
		u := url.URL{
			Scheme: "http",
			Path:   path.Join(m.serverAddress, "/update/"),
		}
		req := m.client.R()
		req.Method = http.MethodPost
		req.Header.Set("Content-Type", "application/json")
		req.URL = u.String()
		b, _ := json.Marshal(metric)
		req.Body = b

		if _, err := req.Send(); err != nil {
			errString += fmt.Sprintf("error sending metric %v: %v\n", mm, err)
		}
	}

	if errString != "" {
		return fmt.Errorf("one or more metrics were not sent to the server: %s", errString)
	}

	m.Notify() // Send notifications for observers.
	return nil
}

// RegisterObserver registration new observer.
func (m *MetricsSender) RegisterObserver(observer Observer) { // TODO: make tests for this
	m.observers[observer] = struct{}{}
}

// RemoveObserver remove observer from list of observers.
func (m *MetricsSender) RemoveObserver(observer Observer) { // TODO: make tests for this
	delete(m.observers, observer)
}

// Notify sending notifications for all observers.
func (m *MetricsSender) Notify() { // TODO: make tests for this
	for o := range m.observers {
		o.OnNotify()
	}
}
