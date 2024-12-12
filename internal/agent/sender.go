// Package agent provides functionality for sending metrics to a server.
package agent

import (
	"bytes"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"sync"

	"github.com/gdyunin/metricol.git/internal/metrics"
	"github.com/go-resty/resty/v2"
)

// Observer is an interface that defines a method to be called on notification.
type Observer interface {
	OnNotify()
}

// MetricsSender is responsible for sending metrics to a specified server address.
// It manages observers that need to be notified after sending metrics.
type MetricsSender struct {
	mu             *sync.Mutex           // Mutex to protect the observers map
	observers      map[Observer]struct{} // Registered observers
	metricsFetcher Fetcher               // Fetcher to retrieve metrics
	client         *resty.Client         // HTTP client for making requests
	serverAddress  string                // Address of the server to send metrics to
}

// NewMetricsSender creates a new instance of MetricsSender with the provided fetcher and server address.
func NewMetricsSender(fetcher Fetcher, address string) *MetricsSender {
	return &MetricsSender{
		metricsFetcher: fetcher,
		serverAddress:  address,
		client:         resty.New(),
		observers:      make(map[Observer]struct{}),
		mu:             &sync.Mutex{}, // Initialize the mutex
	}
}

// Send sends all fetched metrics to the server. It returns an error if any metric fails to send.
func (m *MetricsSender) Send() error {
	var buf bytes.Buffer
	for _, mm := range m.metricsFetcher.Metrics() {
		metricType, metricName, metricValue, ok := m.recognizeMetric(mm)
		if !ok {
			buf.WriteString(fmt.Sprintf("error sending metric %v: failed conversion Metric to Struct\n", mm))
			continue
		}

		req := m.makeRequest(metricType, metricName, metricValue)
		if _, err := req.Send(); err != nil {
			buf.WriteString(fmt.Sprintf("error sending metric %v: %v\n", mm, err))
		}
	}

	if buf.Len() != 0 {
		return fmt.Errorf("one or more metrics were not sent to the server: %s", buf.String())
	}

	m.Notify()
	return nil
}

// RegisterObserver adds an observer to the list of observers to be notified.
func (m *MetricsSender) RegisterObserver(observer Observer) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.observers[observer] = struct{}{}
}

// RemoveObserver removes an observer from the list of observers.
func (m *MetricsSender) RemoveObserver(observer Observer) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.observers, observer)
}

// Notify calls the OnNotify method on all registered observers.
func (m *MetricsSender) Notify() {
	for o := range m.observers {
		o.OnNotify()
	}
}

// makeRequest constructs an HTTP request for sending a metric to the server.
func (m *MetricsSender) makeRequest(mType, mName, mValue string) *resty.Request {
	u := url.URL{
		Scheme: "http",
		Path:   path.Join(m.serverAddress, "/update/", mType, mName, mValue),
	}

	req := m.client.R()
	req.Method = http.MethodPost
	req.Header.Set("Content-Type", "text/plain; charset=utf-8")
	req.URL = u.String()

	return req
}

// recognizeMetric identifies the type, name, and value of a metric.
// It returns these values along with a boolean indicating success or failure.
func (m *MetricsSender) recognizeMetric(mm metrics.Metric) (metricType, name, value string, ok bool) {
	name = mm.StringName()
	value = mm.StringValue()
	ok = true

	switch mm.(type) {
	case *metrics.Counter:
		metricType = metrics.MetricTypeCounter
	case *metrics.Gauge:
		metricType = metrics.MetricTypeGauge
	default:
		return "", "", "", false
	}

	return
}
