// Package send provides functionality to send metrics to a server.
package send

import (
	"fmt"
	"net/http"
	"net/url"
	"path"

	"github.com/gdyunin/metricol.git/internal/agent/fetch"
	"github.com/gdyunin/metricol.git/internal/metrics"
	"github.com/go-resty/resty/v2"
)

// Sender defines an interface for sending metrics.
type Sender interface {
	// Send sends the metrics to the server and returns an error if the operation fails.
	Send() error
}

// MetricsSender implements the Sender interface, responsible for sending metrics.
type MetricsSender struct {
	metricsFetcher fetch.Fetcher // Fetcher to retrieve metrics.
	client         *resty.Client // HTTP client for sending requests.
	serverAddress  string        // Address of the server to send metrics to.
}

// NewMetricsSender creates a new instance of MetricsSender with the provided fetcher and server address.
func NewMetricsSender(fetcher fetch.Fetcher, address string) *MetricsSender {
	return &MetricsSender{
		metricsFetcher: fetcher,
		serverAddress:  address,
		client:         resty.New(),
	}
}

// Send sends all metrics to the configured server address.
// It returns an error if any metric fails to be sent.
func (m *MetricsSender) Send() error {
	var errString string
	for _, mm := range m.metricsFetcher.Metrics() {
		var metricName, metricType string

		switch currentMetricStruct := mm.(type) {
		case *metrics.Counter:
			metricName = currentMetricStruct.Name
			metricType = metrics.MetricTypeCounter
		case *metrics.Gauge:
			metricName = currentMetricStruct.Name
			metricType = metrics.MetricTypeGauge // Fixed to use Gauge type
		default:
			errString += fmt.Sprintf("error sending metric %v: failed conversion Metric to Struct\n", mm)
			continue // Skip to the next metric
		}

		u := url.URL{
			Scheme: "http",
			Path:   path.Join(m.serverAddress, "/update/", metricType, metricName, mm.StringValue()),
		}

		req := m.client.R()
		req.Method = http.MethodPost
		req.Header.Set("Content-Type", "text/plain; charset=utf-8")
		req.URL = u.String()

		if _, err := req.Send(); err != nil {
			errString += fmt.Sprintf("error sending metric %v: %v\n", mm, err)
		}
	}

	if errString != "" {
		return fmt.Errorf("one or more metrics were not sent to the server: %s", errString)
	}
	return nil
}
